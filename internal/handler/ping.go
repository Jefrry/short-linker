package handler

import (
	"database/sql"
	"net/http"
)

type PingHandler struct{
	db *sql.DB
}

func NewPingHandler(db *sql.DB) *PingHandler {
	return &PingHandler{
		db: db,
	}
}

// I know that handler should not use db directly, but for ping it's ok?
func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request)  {
	err := h.db.Ping()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}