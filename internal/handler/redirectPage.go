package handler

import (
	"net/http"
	
	"short-linker/internal/storage"
)

func RedirectPage(w http.ResponseWriter, r *http.Request, store *storage.Memory, id string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	url, exists := store.Get(id)
	if !exists {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}