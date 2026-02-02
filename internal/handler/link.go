package handler

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"short-linker/internal/logger"
	"short-linker/internal/middleware"
	"short-linker/internal/model"
	"short-linker/internal/service"
	"short-linker/internal/utils"
)

type LinkHandler struct {
	service service.LinkService
	logger  logger.Logger
	utils   utils.HandlerUtils
}

func NewLinkHandler(l logger.Logger, service service.LinkService, u utils.HandlerUtils) *LinkHandler {
	return &LinkHandler{
		service: service,
		logger:  l,
		utils:   u,
	}
}

func (h *LinkHandler) CreateShortLink(w http.ResponseWriter, r *http.Request) {
	var data model.LinkPayload
	if !h.utils.ReadJSON(w, r, &data) {
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	shortLink, err := h.service.CreateShortLink(r.Context(), data.URL, userID)
	if err != nil && errors.Is(err, model.ErrOriginalURLExists) {
		h.utils.WriteJSON(w, http.StatusConflict, model.LinkResponse{ShortURL: shortLink})
		return
	}
	if err != nil {
		http.Error(w, "Failed to create short link", http.StatusInternalServerError)
		h.logger.Error("CreateShortLink error", logger.Error(err))
		return
	}
	h.utils.WriteJSON(w, http.StatusCreated, model.LinkResponse{ShortURL: shortLink})
}

// TODO: add tests
func (h *LinkHandler) CreateShortLinkBatch(w http.ResponseWriter, r *http.Request) {
	var data []model.LinkBatchPayload
	if !h.utils.ReadJSON(w, r, &data) {
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	res, err := h.service.CreateShortLinkBatch(r.Context(), data, userID)
	if err != nil {
		http.Error(w, "Failed to create short link batch", http.StatusInternalServerError)
		return
	}

	h.utils.WriteJSON(w, http.StatusCreated, res)
}

// Deprecated: use CreateShortLink with Content-Type: application/json instead
func (h *LinkHandler) CreateShortLinkPlain(w http.ResponseWriter, r *http.Request) {
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct != "text/plain" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	shortLink, err := h.service.CreateShortLink(r.Context(), string(body), 0)
	if err != nil && errors.Is(err, model.ErrOriginalURLExists) {
		h.utils.WritePlain(w, http.StatusConflict, shortLink)
		return
	}
	if err != nil {
		http.Error(w, "Failed to create short link", http.StatusInternalServerError)
		h.logger.Error("CreateShortLinkPlain error", logger.Error(err))
		return
	}
	h.utils.WritePlain(w, http.StatusCreated, shortLink)
}

func (h *LinkHandler) RedirectPage(w http.ResponseWriter, r *http.Request, id string) {
	originalURL, deleted, err := h.service.GetOriginalURL(r.Context(), id)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	if deleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
