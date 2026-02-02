package utils

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"
)

type HandlerUtils interface {
	ReadJSON(w http.ResponseWriter, r *http.Request, data any) bool
	WriteJSON(w http.ResponseWriter, statusCode int, data any)
	WritePlain(w http.ResponseWriter, statusCode int, text string)
	SetSessionCookie(w http.ResponseWriter, token string)
	RemoveSessionCookie(w http.ResponseWriter)
}

type handlerUtils struct{}

func NewHandlerUtils() HandlerUtils {
	return &handlerUtils{}
}

func (u *handlerUtils) ReadJSON(w http.ResponseWriter, r *http.Request, data any) bool {
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct != "application/json" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return false
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return false
	}

	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return false
	}

	err = json.Unmarshal(body, data)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return false
	}

	return true
}

func (u *handlerUtils) WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func (u *handlerUtils) WritePlain(w http.ResponseWriter, statusCode int, text string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(text))
}

func (u *handlerUtils) SetSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // CAUTION: change to true in production
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
		MaxAge:   24 * 60 * 60,
	}
	http.SetCookie(w, cookie)
}

func (u *handlerUtils) RemoveSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // CAUTION: change to true in production
		SameSite: http.SameSiteLaxMode,
	})
}
