package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"short-linker/internal/logger"
	"short-linker/internal/middleware"
	"short-linker/internal/model"
	"short-linker/internal/service"
	"short-linker/internal/utils"
	"short-linker/pkg"
)

type UserHandler struct {
	service service.UserService
	logger  logger.Logger
	utils   utils.HandlerUtils
}

func NewUserHandler(l logger.Logger, service service.UserService, u utils.HandlerUtils) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  l,
		utils:   u,
	}
}

// Signup godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.SignupPayload true "User signup data"
// @Success 201 {object} model.User "User created"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /api/user/signup [post]
func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var data model.SignupPayload
	if !h.utils.ReadJSON(w, r, &data) {
		return
	}

	if data.Name == "" || data.Email == "" || data.Password == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	if !pkg.ValidateEmail(data.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}
	if ok, msg := pkg.ValidatePassword(data.Password); !ok {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	res, token, err := h.service.Signup(r.Context(), data)
	if err != nil {
		http.Error(w, "Failed to signup user", http.StatusInternalServerError)
		h.logger.Error("Signup error", logger.Error(err))
		return
	}

	h.utils.SetSessionCookie(w, token)
	h.utils.WriteJSON(w, http.StatusCreated, res)
}

// Signin godoc
// @Summary Sign in a user
// @Description Authenticate user and return JWT token
// @Tags users
// @Accept json
// @Produce plain
// @Param credentials body object{email=string,password=string} true "User credentials"
// @Success 204 "No content"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /api/user/signin [post]
func (h *UserHandler) Signin(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !h.utils.ReadJSON(w, r, &data) {
		return
	}

	if data.Email == "" || data.Password == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	if !pkg.ValidateEmail(data.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}
	if ok, msg := pkg.ValidatePassword(data.Password); !ok {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	token, err := h.service.Signin(r.Context(), data.Email, data.Password)
	if err != nil {
		http.Error(w, "Failed to signin user", http.StatusInternalServerError)
		h.logger.Error("Signin error", logger.Error(err))
		return
	}

	h.utils.SetSessionCookie(w, token)
	w.WriteHeader(http.StatusNoContent)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the authenticated user's profile
// @Tags users
// @Produce json
// @Success 200 {object} model.User "User profile"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /api/user/profile [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	user, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		h.logger.Error("GetProfile error", logger.Error(err))
		return
	}

	h.utils.WriteJSON(w, http.StatusOK, user)
}

// Signout godoc
// @Summary Sign out a user
// @Description Log out the current user and clear session
// @Tags users
// @Produce json
// @Success 200 {object} object{message=string} "Logged out"
// @Failure 401 {string} string "Unauthorized"
// @Security BearerAuth
// @Router /api/user/signout [post]
func (h *UserHandler) Signout(w http.ResponseWriter, r *http.Request) {
	h.utils.RemoveSessionCookie(w)
	h.utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "logged out",
	})
}

// GetLinks godoc
// @Summary Get user's links
// @Description Get all short links created by the authenticated user
// @Tags users
// @Produce json
// @Success 200 {array} model.LinkItem "List of links"
// @Success 204 "No content"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Security BearerAuth
// @Router /api/user/urls [get]
func (h *UserHandler) GetLinks(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	links, err := h.service.GetLinks(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user links", http.StatusInternalServerError)
		h.logger.Error("GetLinks error", logger.Error(err))
		return
	}

	statusCode := http.StatusOK
	if len(links) == 0 {
		statusCode = http.StatusNoContent
	}

	h.utils.WriteJSON(w, statusCode, links)
}

// DeleteLinks godoc
// @Summary Delete user's links
// @Description Delete multiple short links by their IDs (async operation)
// @Tags users
// @Accept json
// @Param ids body []string true "Array of short link IDs to delete"
// @Success 202 "Accepted - deletion in progress"
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Security BearerAuth
// @Router /api/user/urls [delete]
func (h *UserHandler) DeleteLinks(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	var data []string
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	go h.service.DeleteLinks(context.Background(), data, userID)

	w.WriteHeader(http.StatusAccepted)
}
