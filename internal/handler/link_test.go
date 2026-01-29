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

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"short-linker/internal/handler"
	"short-linker/internal/model"
)

// Maybe I should move it to service
type mockLinkService struct {
	createResult string
	getResult    string
	deleted      bool

	createErr error
	getErr    error
}

func (m *mockLinkService) CreateShortLink(ctx context.Context, originalURL string, userID int64) (string, error) {
	return m.createResult, m.createErr
}

func (m *mockLinkService) GetOriginalURL(ctx context.Context, id string) (string, bool, error) {
	return m.getResult, m.deleted, m.getErr
}

func (m *mockLinkService) CreateShortLinkBatch(ctx context.Context, items []model.LinkBatchPayload, userID int64) ([]model.LinkBatchResponse, error) {
	return nil, nil
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
		serviceResult string
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
				body:        host + "abc",
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
			serviceResult: host + "abc",
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
				body:        "",
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
			serviceResult: "",
			needTestBody:  true,
		},
		{
			name: "Invalid method",
			reqData: reqData{
				method:      http.MethodGet,
				contentType: "application/json",
				body:        "http://example.com",
			},
			respData: respData{
				body:        "",
				statusCode:  http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
			},
			serviceResult: "",
			needTestBody:  false,
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
			serviceResult: "",
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
			serviceResult: "",
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
			serviceResult: "",
			needTestBody:  false,
		},
	}

	logger := zap.NewNop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := &mockLinkService{createResult: tt.serviceResult, createErr: tt.serviceErr}
			handler := handler.NewLinkHandler(logger, linkService)

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

			req = req.WithContext(context.WithValue(req.Context(), model.JWTUserIDKey, 0))

			handler.CreateShortLink(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode, "status code should match")
			assert.Equal(t, tt.respData.contentType, resp.Header.Get("Content-Type"), "content type should match")

			if tt.needTestBody {
				body := w.Body.String()

				var resp model.LinkResponse
				err := json.Unmarshal([]byte(body), &resp)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v; body=%q", err, body)
				}

				assert.Equal(t, tt.respData.body, resp.ShortURL, "response body should match expected short link")
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
		name           string
		reqData        reqData
		respData       respData
		serviceResult  string
		serviceDeleted bool
		serviceErr     error
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
			serviceResult:  host + randomID,
			serviceDeleted: false,
		},
		{
			name: "Invalid method",
			reqData: reqData{
				method: http.MethodPost, // Do I need to test other methods?
				id:     randomID,
			},
			respData: respData{
				statusCode: http.StatusMethodNotAllowed,
			},
			serviceResult:  "",
			serviceDeleted: false,
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
			serviceResult:  "",
			serviceDeleted: false,
			serviceErr:     errors.New("link not found"),
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
			serviceResult:  host + randomID,
			serviceDeleted: true,
		},
	}

	logger := zap.NewNop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := &mockLinkService{getResult: tt.serviceResult, deleted: tt.serviceDeleted, getErr: tt.serviceErr}
			handler := handler.NewLinkHandler(logger, linkService)

			req := httptest.NewRequest(tt.reqData.method, "/"+tt.reqData.id, nil)
			w := httptest.NewRecorder()

			handler.RedirectPage(w, req, tt.reqData.id)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode, "status code should match")
		})
	}
}
