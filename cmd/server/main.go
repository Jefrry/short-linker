package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	"short-linker/internal/handler"
	"short-linker/internal/storage"
)

func main() {
	r := chi.NewRouter()

	storage := storage.NewMemory()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handler.MainPage(w, r, storage)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		handler.RedirectPage(w, r, storage, id)
	})

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
