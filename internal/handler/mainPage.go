package handler

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	
	"short-linker/internal/storage"
	"short-linker/pkg"
)


func MainPage(w http.ResponseWriter, r *http.Request, store *storage.Memory) {
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


	var id string
	for {
		id, err = pkg.RandomStringDefault()
		if err != nil {
			http.Error(w, "Failed to generate short link", http.StatusInternalServerError)
			return
		}

		if !store.Exists(id) {
			err = store.Set(id, string(body))
			if err != nil {
				http.Error(w, "Failed to store link", http.StatusInternalServerError)
				return
			}
			break
		}
	}

	shortLink := fmt.Sprintf("http://%s/%s", r.Host, id)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortLink))
}