package main

import (
	"net/http"
	"short-linker/internal/handler"
	"short-linker/internal/storage"
)

func main() {
	mux := http.NewServeMux()

	storage := storage.NewMemory()

	mux.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[1:]
		if id == "" {
			handler.MainPage(w, r, storage)
			return
		}

		handler.RedirectPage(w, r, storage, id)
	})

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}