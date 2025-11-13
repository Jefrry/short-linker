package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"short-linker/internal/handler"
)

// Maybe I should move it to service
type mockLinkService struct {
	createResult string
	getResult    string

	createErr error
	getErr    error
}

func (m *mockLinkService) CreateShortLink(originalURL string) (string, error) {
	return m.createResult, m.createErr
}

func (m *mockLinkService) GetOriginalURL(id string) (string, error) {
	return m.getResult, m.getErr
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
				contentType: "text/plain",
				body:        "http://example.com",
			},
			respData: respData{
				body:        host + "abc",
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
			},
			serviceResult: host + "abc",
			needTestBody:  false,
		},
		{
			name: "Response body check",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "text/plain",
				body:        "http://example.com",
			},
			respData: respData{
				body:        "",
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
			},
			serviceResult: "",
			needTestBody:  true,
		},
		{
			name: "Invalid method",
			reqData: reqData{
				method:      http.MethodGet,
				contentType: "text/plain",
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
				contentType: "application/json",
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
				contentType: "text/plain",
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
				contentType: "text/plain",
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := &mockLinkService{createResult: tt.serviceResult, createErr: tt.serviceErr}
			handler := handler.NewLinkHandler(linkService)

			req := httptest.NewRequest(tt.reqData.method, host, strings.NewReader(tt.reqData.body))
			req.Header.Set("Content-Type", tt.reqData.contentType)

			w := httptest.NewRecorder()

			handler.CreateShortLink(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode, "status code should match")
			assert.Equal(t, tt.respData.contentType, resp.Header.Get("Content-Type"), "content type should match")

			if tt.needTestBody {
				body := w.Body.String()
				assert.Equal(t, body, tt.respData.body, "response body should match expected short link")
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
		serviceResult string
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
			serviceResult: host + randomID,
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
			serviceResult: "",
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
			serviceResult: "",
			serviceErr:    errors.New("link not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := &mockLinkService{getResult: tt.serviceResult, getErr: tt.serviceErr}
			handler := handler.NewLinkHandler(linkService)

			req := httptest.NewRequest(tt.reqData.method, "/" + tt.reqData.id, nil)
			w := httptest.NewRecorder()

			handler.RedirectPage(w, req, randomID)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode, "status code should match")
		})
	}
}
