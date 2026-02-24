package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"short-linker/internal/handler"
	"short-linker/internal/logger"
	"short-linker/internal/model"
	"short-linker/internal/utils"
)

// Maybe I should move it to service
type mockLinkService struct {
	createResult  model.LinkItem
	getResult     model.LinkItem
	batchResult   []model.LinkBatchResponse
	metricsResult []model.VisitsByDate

	createErr  error
	getErr     error
	batchErr   error
	metricsErr error
}

func (m *mockLinkService) CreateShortLink(ctx context.Context, originalURL string, userID int64) (model.LinkItem, error) {
	return m.createResult, m.createErr
}

func (m *mockLinkService) GetOriginalURL(ctx context.Context, id string) (model.LinkItem, error) {
	return m.getResult, m.getErr
}

func (m *mockLinkService) CreateShortLinkBatch(ctx context.Context, items []model.LinkBatchPayload, userID int64) ([]model.LinkBatchResponse, error) {
	return m.batchResult, m.batchErr
}

func (m *mockLinkService) RecordVisit(ctx context.Context, visit model.Visit) {
	// no-op
}

func (m *mockLinkService) GetLinkMetrics(ctx context.Context, linkID string, userID int64, from, to time.Time) ([]model.VisitsByDate, error) {
	return m.metricsResult, m.metricsErr
}

func TestCreateShortLink(t *testing.T) {
	type reqData struct {
		method      string
		contentType string
		body        string
	}
	type respData struct {
		statusCode  int
		contentType string
		body        string
	}

	host := "http://localhost:8080/"
	tests := []struct {
		name          string
		reqData       reqData
		respData      respData
		serviceResult model.LinkItem
		serviceErr    error
		needTestBody  bool
	}{
		{
			name: "Success request",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        "http://example.com",
			},
			respData: respData{
				body:        "abc",
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
			serviceResult: model.LinkItem{ID: "abc", OriginalURL: "http://example.com", ShortURL: host + "abc"},
			needTestBody:  false,
		},
		{
			name: "Response body check",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        "http://example.com",
			},
			respData: respData{
				body:        "abc",
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
			serviceResult: model.LinkItem{ID: "abc", OriginalURL: "http://example.com", ShortURL: host + "abc"},
			needTestBody:  true,
		},
		{
			name: "Invalid content type",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "text/plain",
				body:        "http://example.com",
			},
			respData: respData{
				body:        "",
				statusCode:  http.StatusUnsupportedMediaType,
				contentType: "text/plain; charset=utf-8",
			},
			serviceResult: model.LinkItem{},
			needTestBody:  false,
		},
		{
			name: "Empty body",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        "",
			},
			respData: respData{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
			serviceResult: model.LinkItem{},
			needTestBody:  false,
		},
		{
			name: "Service error",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        "http://example.com",
			},
			respData: respData{
				statusCode:  http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
			},
			serviceErr:    errors.New("something wrong with service"),
			serviceResult: model.LinkItem{},
			needTestBody:  false,
		},
	}

	logger := logger.NewLogger(zap.NewNop())
	handlerUtils := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := &mockLinkService{createResult: tt.serviceResult, createErr: tt.serviceErr}
			handler := handler.NewLinkHandler(logger, linkService, handlerUtils)

			var bodyReader io.Reader
			if tt.reqData.body != "" {
				payload := model.LinkPayload{URL: tt.reqData.body}
				b, _ := json.Marshal(payload)
				bodyReader = bytes.NewReader(b)
			} else {
				bodyReader = nil
			}

			req := httptest.NewRequest(tt.reqData.method, host, bodyReader)
			req.Header.Set("Content-Type", tt.reqData.contentType)

			w := httptest.NewRecorder()

			req = req.WithContext(context.WithValue(req.Context(), model.JWTUserIDKey, int64(0)))

			handler.CreateShortLink(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode, "status code should match")
			assert.Equal(t, tt.respData.contentType, resp.Header.Get("Content-Type"), "content type should match")

			if tt.needTestBody {
				body := w.Body.String()

				var res model.LinkItem
				err := json.Unmarshal([]byte(body), &res)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v; body=%q", err, body)
				}

				assert.Equal(t, tt.respData.body, res.ID, "response body should match expected ID")
			}
		})
	}
}

func TestRedirectPage(t *testing.T) {
	type reqData struct {
		method string
		id     string
	}
	type respData struct {
		statusCode int
	}

	randomID := "abc"
	host := "http://localhost/"
	tests := []struct {
		name          string
		reqData       reqData
		respData      respData
		serviceResult model.LinkItem
		serviceErr    error
	}{
		{
			name: "Success request",
			reqData: reqData{
				method: http.MethodGet,
				id:     randomID,
			},
			respData: respData{
				statusCode: http.StatusTemporaryRedirect,
			},
			serviceResult: model.LinkItem{
				ID:          randomID,
				OriginalURL: host + randomID,
				Deleted:     false,
				UserID:      1, // Set UserID to trigger RecordVisit goroutine (though it's a mock)
			},
		},
		{
			name: "ID not found",
			reqData: reqData{
				method: http.MethodGet,
				id:     "missing",
			},
			respData: respData{
				statusCode: http.StatusNotFound,
			},
			serviceResult: model.LinkItem{},
			serviceErr:    errors.New("link not found"),
		},
		{
			name: "ID deleted",
			reqData: reqData{
				method: http.MethodGet,
				id:     randomID,
			},
			respData: respData{
				statusCode: http.StatusGone,
			},
			serviceResult: model.LinkItem{
				ID:          randomID,
				OriginalURL: host + randomID,
				Deleted:     true,
			},
		},
	}

	logger := logger.NewLogger(zap.NewNop())
	handlerUtils := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := &mockLinkService{getResult: tt.serviceResult, getErr: tt.serviceErr}
			handler := handler.NewLinkHandler(logger, linkService, handlerUtils)

			req := httptest.NewRequest(tt.reqData.method, "/"+tt.reqData.id, nil)
			w := httptest.NewRecorder()

			handler.RedirectPage(w, req, tt.reqData.id)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode, "status code should match")
		})
	}
}

func TestCreateShortLinkBatch(t *testing.T) {
	type respData struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name          string
		body          []model.LinkBatchPayload
		respData      respData
		serviceResult []model.LinkBatchResponse
		serviceErr    error
	}{
		{
			name: "Success request",
			body: []model.LinkBatchPayload{
				{CorrelationID: "1", URL: "http://example.com/1"},
				{CorrelationID: "2", URL: "http://example.com/2"},
			},
			respData: respData{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
			serviceResult: []model.LinkBatchResponse{
				{CorrelationID: "1", ShortURL: "http://localhost:8080/abc"},
				{CorrelationID: "2", ShortURL: "http://localhost:8080/def"},
			},
		},
		{
			name: "Service error",
			body: []model.LinkBatchPayload{
				{CorrelationID: "1", URL: "http://example.com/1"},
			},
			respData: respData{
				statusCode:  http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
			},
			serviceErr: errors.New("batch error"),
		},
	}

	logger := logger.NewLogger(zap.NewNop())
	handlerUtils := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := &mockLinkService{batchResult: tt.serviceResult, batchErr: tt.serviceErr}
			handler := handler.NewLinkHandler(logger, linkService, handlerUtils)

			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), model.JWTUserIDKey, int64(0)))

			w := httptest.NewRecorder()
			handler.CreateShortLinkBatch(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode)
			assert.Equal(t, tt.respData.contentType, resp.Header.Get("Content-Type"))

			if tt.respData.statusCode == http.StatusCreated {
				var result []model.LinkBatchResponse
				err := json.NewDecoder(resp.Body).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, tt.serviceResult, result)
			}
		})
	}
}

func TestCreateShortLinkPlain(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		body           string
		serviceResult  model.LinkItem
		serviceErr     error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success request",
			contentType:    "text/plain",
			body:           "http://example.com",
			serviceResult:  model.LinkItem{ID: "abc", OriginalURL: "http://example.com", ShortURL: "http://localhost:8080/abc"},
			expectedStatus: http.StatusCreated,
			expectedBody:   "abc",
		},
		{
			name:           "Invalid content type",
			contentType:    "application/json",
			body:           "http://example.com",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "Empty body",
			contentType:    "text/plain",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Service error",
			contentType:    "text/plain",
			body:           "http://example.com",
			serviceErr:     errors.New("plain error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	logger := logger.NewLogger(zap.NewNop())
	handlerUtils := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := &mockLinkService{createResult: tt.serviceResult, createErr: tt.serviceErr}
			handler := handler.NewLinkHandler(logger, linkService, handlerUtils)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			handler.CreateShortLinkPlain(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, tt.expectedBody, string(body))
			}
		})
	}
}
