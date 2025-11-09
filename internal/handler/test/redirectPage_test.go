package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"short-linker/internal/handler"
	"short-linker/internal/storage"
)

func TestRedirectPage(t *testing.T) {
	type reqData struct {
		method      string
	}
	type respData struct {
		statusCode  int
	}

	tests := []struct {
		name         string
		reqData      reqData
		respData     respData
	}{
		{
			name: "Success request",
			reqData: reqData{
				method:      http.MethodGet,
			},
			respData: respData{
				statusCode:  http.StatusTemporaryRedirect,
			},
		},
		{
			name: "Invalid method",
			reqData: reqData{
				method:      http.MethodPost, // Do I need to test other methods?
			},
			respData: respData{
				statusCode:  http.StatusBadRequest,  // TODO: implement status code tests after fixing them in handler
			},
		},
	}
	randomID := "abc123"
	host := "http://localhost/"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemory()
			store.Set(randomID, host+randomID)
			
			req := httptest.NewRequest(tt.reqData.method, "/"+randomID, nil)

			w := httptest.NewRecorder()

			handler.RedirectPage(w, req, store, randomID)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.respData.statusCode, resp.StatusCode, "status code should match")
		})
	}
}
