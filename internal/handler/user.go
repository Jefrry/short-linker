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

// TODO: add validation
func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var data model.SignupPayload
	if !h.utils.ReadJSON(w, r, &data) {
		return
	}

	if data.Name == "" || data.Email == "" || data.Password == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	res, err := h.service.Signup(r.Context(), data)
	if err != nil {
		http.Error(w, "Failed to signup user", http.StatusInternalServerError)
		h.logger.Error("Signup error", logger.Error(err))
		return
	}

	h.utils.WriteJSON(w, http.StatusCreated, res)
}

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

	token, err := h.service.Signin(r.Context(), data.Email, data.Password)
	if err != nil {
		http.Error(w, "Failed to signin user", http.StatusInternalServerError)
		h.logger.Error("Signin error", logger.Error(err))
		return
	}

	h.utils.SetSessionCookie(w, token)
	h.utils.WritePlain(w, http.StatusOK, token)
}

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

func (h *UserHandler) Signout(w http.ResponseWriter, r *http.Request) {
	h.utils.RemoveSessionCookie(w)
	h.utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "logged out",
	})
}

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
