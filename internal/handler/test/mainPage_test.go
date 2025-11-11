package handler_test


import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"short-linker/internal/handler"
	"short-linker/internal/storage"
)

func TestMainPage(t *testing.T) {
	type reqData struct {
		method      string
		contentType string
		body        string
	}
	type respData struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name         string
		reqData      reqData
		respData     respData
		needTestBody bool
	}{
		{
			name: "Success request",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "text/plain",
				body:        "http://example.com",
			},
			respData: respData{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
			},
			needTestBody: false,
		},
		{
			name: "Response body check",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "text/plain",
				body:        "http://example.com",
			},
			respData: respData{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
			},
			needTestBody: true,
		},
		{
			name: "Invalid method",
			reqData: reqData{
				method:      http.MethodGet,
				contentType: "text/plain",
				body:        "http://example.com",
			},
			respData: respData{
				statusCode:  http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
			},
			needTestBody: false,
		},
		{
			name: "Invalid content type",
			reqData: reqData{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        "http://example.com",
			},
			respData: respData{
				statusCode:  http.StatusUnsupportedMediaType,
				contentType: "text/plain; charset=utf-8",
			},
			needTestBody: false,
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
			needTestBody: false,
		},
	}
	host := "http://localhost/"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Creating a mock storage is an ideal realization,
			// but our storage is a simple memory map, so it is ok for now.
			store := storage.NewMemory()

			req := httptest.NewRequest(tt.reqData.method, host, strings.NewReader(tt.reqData.body))
			req.Header.Set("Content-Type", tt.reqData.contentType)

			w := httptest.NewRecorder()

			handler.MainPage(w, req, store)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode, "status code should match")
			assert.Equal(t, tt.respData.contentType, resp.Header.Get("Content-Type"), "content type should match")

			if tt.needTestBody {
				body := w.Body.String()

				id := strings.TrimPrefix(body, host)
				_, exists := store.Get(id)
				assert.True(t, exists, "short link ID should exist in storage")

				assert.Equal(t, host+id, body, "response body should match expected short link")
			}
		})
	}
}
