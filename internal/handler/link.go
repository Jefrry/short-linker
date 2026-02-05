package handler

import (
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

// CreateShortLink godoc
// @Summary Create a short link
// @Description Create a new short link from a long URL
// @Tags links
// @Accept json
// @Produce json
// @Param link body model.LinkPayload true "Link payload"
// @Success 201 {object} model.LinkResponse "Short link created"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /api/shorten [post]
func (h *LinkHandler) CreateShortLink(w http.ResponseWriter, r *http.Request) {
	var data model.LinkPayload
	if !h.utils.ReadJSON(w, r, &data) {
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	shortLink, err := h.service.CreateShortLink(r.Context(), data.URL, userID)
	if err != nil {
		http.Error(w, "Failed to create short link", http.StatusInternalServerError)
		h.logger.Error("CreateShortLink error", logger.Error(err))
		return
	}
	h.utils.WriteJSON(w, http.StatusCreated, model.LinkResponse{ShortURL: shortLink})
}

// CreateShortLinkBatch godoc
// @Summary Create multiple short links
// @Description Create short links for a batch of URLs
// @Tags links
// @Accept json
// @Produce json
// @Param links body []model.LinkBatchPayload true "Batch of links"
// @Success 201 {array} model.LinkBatchResponse "Short links created"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /api/shorten/batch [post]
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

// CreateShortLinkPlain godoc
// @Summary Create a short link (plain text)
// @Description Create a new short link from a plain text URL body
// @Tags links
// @Accept plain
// @Produce plain
// @Param url body string true "Original URL"
// @Success 201 {string} string "Short URL"
// @Failure 400 {string} string "Bad request"
// @Failure 415 {string} string "Unsupported media type"
// @Failure 500 {string} string "Internal server error"
// @Deprecated
// @Router / [post]
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
	if err != nil {
		http.Error(w, "Failed to create short link", http.StatusInternalServerError)
		h.logger.Error("CreateShortLinkPlain error", logger.Error(err))
		return
	}
	h.utils.WritePlain(w, http.StatusCreated, shortLink)
}

// RedirectPage godoc
// @Summary Redirect to original URL
// @Description Redirect to the original URL by short link ID
// @Tags links
// @Param id path string true "Short link ID"
// @Success 307 "Temporary redirect to original URL"
// @Failure 404 {string} string "Link not found"
// @Failure 410 {string} string "Link has been deleted"
// @Router /{id} [get]
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
