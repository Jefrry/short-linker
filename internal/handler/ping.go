package handler

import (
	"database/sql"
	"net/http"

	"short-linker/internal/utils"
)

type PingHandler struct {
	db    *sql.DB
	utils utils.HandlerUtils
}

func NewPingHandler(db *sql.DB, u utils.HandlerUtils) *PingHandler {
	return &PingHandler{
		db:    db,
		utils: u,
	}
}

// I know that handler should not use db directly, but for ping it's ok?
func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	h.utils.WritePlain(w, http.StatusOK, "pong")
}
