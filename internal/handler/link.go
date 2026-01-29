package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"io"
	"mime"
	"net/http"
	"time"

	"short-linker/internal/middleware"
	"short-linker/internal/model"
	"short-linker/internal/service"
)

type LinkHandler struct {
	service service.LinkService
	logger  *zap.Logger
}

func NewLinkHandler(logger *zap.Logger, service service.LinkService) *LinkHandler {
	return &LinkHandler{
		service: service,
		logger:  logger,
	}
}

func (h *LinkHandler) CreateShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct != "application/json" {
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

	var data model.LinkPayload
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userID, _ := middleware.GetUserID(r.Context())

	shortLink, err := h.service.CreateShortLink(ctx, data.URL, userID)
	if err != nil && errors.Is(err, model.ErrOriginalURLExists) {
		_ = h.writeJSONResponse(w, http.StatusConflict, shortLink)
		return
	}
	if err != nil {
		http.Error(w, "Failed to create short link", http.StatusInternalServerError)
		h.logger.Error("CreateShortLink error", zap.Error(err))
		return
	}
	_ = h.writeJSONResponse(w, http.StatusCreated, shortLink)
}

// TODO: add tests
func (h *LinkHandler) CreateShortLinkBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct != "application/json" {
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

	var data []model.LinkBatchPayload
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userID, _ := middleware.GetUserID(r.Context())

	var res []model.LinkBatchResponse
	res, err = h.service.CreateShortLinkBatch(ctx, data, userID)
	if err != nil {
		http.Error(w, "Failed to create short link batch", http.StatusInternalServerError)
		return
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		h.logger.Error("CreateShortLinkBatch error", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resBytes)
}

// Deprecated: use CreateShortLink with Content-Type: application/json instead
func (h *LinkHandler) CreateShortLinkPlain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	shortLink, err := h.service.CreateShortLink(ctx, string(body), 0)
	if err != nil && errors.Is(err, model.ErrOriginalURLExists) {
		h.writePlainResponse(w, http.StatusConflict, shortLink)
		return
	}
	if err != nil {
		http.Error(w, "Failed to create short link", http.StatusInternalServerError)
		h.logger.Error("CreateShortLinkPlain error", zap.Error(err))
		return
	}
	h.writePlainResponse(w, http.StatusCreated, shortLink)
}

func (h *LinkHandler) RedirectPage(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	originalURL, deleted, err := h.service.GetOriginalURL(ctx, id)
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

func (h *LinkHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, shortURL string) error {
	res := model.LinkResponse{
		ShortURL: shortURL,
	}
	shortLinkBytes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(shortLinkBytes)
	return nil
}

func (h *LinkHandler) writePlainResponse(w http.ResponseWriter, statusCode int, shortURL string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(shortURL))
}
