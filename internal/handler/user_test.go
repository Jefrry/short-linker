package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"short-linker/internal/handler"
	"short-linker/internal/logger"
	"short-linker/internal/model"
	"short-linker/internal/utils"
)

type mockUserService struct {
	signupResult model.User
	signupToken  string
	signupErr    error

	signinResult string
	signinErr    error

	profileResult model.User
	profileErr    error

	getLinksResult []model.LinkItem
	getLinksErr    error

	deleteLinksErr error
}

func (m *mockUserService) Signup(ctx context.Context, data model.SignupPayload) (model.User, string, error) {
	return m.signupResult, m.signupToken, m.signupErr
}

func (m *mockUserService) Signin(ctx context.Context, email, password string) (string, error) {
	return m.signinResult, m.signinErr
}

func (m *mockUserService) GetProfile(ctx context.Context, userID int64) (model.User, error) {
	return m.profileResult, m.profileErr
}

func (m *mockUserService) GetLinks(ctx context.Context, userID int64) ([]model.LinkItem, error) {
	return m.getLinksResult, m.getLinksErr
}

func (m *mockUserService) DeleteLinks(ctx context.Context, links []string, userID int64) error {
	return m.deleteLinksErr
}

func TestSignup(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		serviceResult  model.User
		serviceToken   string
		serviceErr     error
		expectedStatus int
	}{
		{
			name: "Success",
			body: model.SignupPayload{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			serviceResult:  model.User{ID: 1, Name: "Test User", Email: "test@example.com"},
			serviceToken:   "valid-token",
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing Fields",
			body: model.SignupPayload{
				Name: "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Email",
			body: model.SignupPayload{
				Name:     "Test User",
				Email:    "invalid-email",
				Password: "Password123!",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Password",
			body: model.SignupPayload{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			body: model.SignupPayload{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			serviceErr:     errors.New("signup failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	l := logger.NewLogger(zap.NewNop())
	u := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{signupResult: tt.serviceResult, signupToken: tt.serviceToken, signupErr: tt.serviceErr}
			h := handler.NewUserHandler(l, svc, u)

			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/user/signup", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Signup(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var res model.User
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.Equal(t, tt.serviceResult.Email, res.Email)
			}
		})
	}
}

func TestSignin(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		serviceResult  string
		serviceErr     error
		expectedStatus int
	}{
		{
			name: "Success",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			serviceResult:  "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing Fields",
			body: map[string]string{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Password",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			serviceErr:     errors.New("signin failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	l := logger.NewLogger(zap.NewNop())
	u := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{signinResult: tt.serviceResult, signinErr: tt.serviceErr}
			h := handler.NewUserHandler(l, svc, u)

			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/user/signin", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Signin(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.serviceResult, w.Body.String())
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		serviceResult  model.User
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           "Success",
			userID:         1,
			serviceResult:  model.User{ID: 1, Name: "Test User", Email: "test@example.com"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Service Error",
			userID:         1,
			serviceErr:     errors.New("profile failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	l := logger.NewLogger(zap.NewNop())
	u := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{profileResult: tt.serviceResult, profileErr: tt.serviceErr}
			h := handler.NewUserHandler(l, svc, u)

			req := httptest.NewRequest(http.MethodGet, "/api/user/profile", nil)
			req = req.WithContext(context.WithValue(req.Context(), model.JWTUserIDKey, tt.userID))
			w := httptest.NewRecorder()

			h.GetProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var res model.User
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.Equal(t, tt.serviceResult.Email, res.Email)
			}
		})
	}
}

func TestSignout(t *testing.T) {
	l := logger.NewLogger(zap.NewNop())
	u := utils.NewHandlerUtils()
	svc := &mockUserService{}
	h := handler.NewUserHandler(l, svc, u)

	req := httptest.NewRequest(http.MethodPost, "/api/user/signout", nil)
	w := httptest.NewRecorder()

	h.Signout(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var res map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, "logged out", res["message"])
}

func TestGetLinks(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		serviceResult  []model.LinkItem
		serviceErr     error
		expectedStatus int
	}{
		{
			name:   "Success",
			userID: 1,
			serviceResult: []model.LinkItem{
				{OriginalURL: "http://example.com/1", ID: "1"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No Content",
			userID:         1,
			serviceResult:  []model.LinkItem{},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Service Error",
			userID:         1,
			serviceErr:     errors.New("links failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	l := logger.NewLogger(zap.NewNop())
	u := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{getLinksResult: tt.serviceResult, getLinksErr: tt.serviceErr}
			h := handler.NewUserHandler(l, svc, u)

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			req = req.WithContext(context.WithValue(req.Context(), model.JWTUserIDKey, tt.userID))
			w := httptest.NewRecorder()

			h.GetLinks(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var res []model.LinkItem
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.Len(t, res, len(tt.serviceResult))
			}
		})
	}
}

func TestDeleteLinks(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		expectedStatus int
	}{
		{
			name:           "Success",
			body:           []string{"id1", "id2"},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Empty Body",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			body:           "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	l := logger.NewLogger(zap.NewNop())
	u := utils.NewHandlerUtils()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			h := handler.NewUserHandler(l, svc, u)

			var b []byte
			if tt.body != nil {
				b, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(b))
			req = req.WithContext(context.WithValue(req.Context(), model.JWTUserIDKey, int64(1)))
			w := httptest.NewRecorder()

			h.DeleteLinks(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
