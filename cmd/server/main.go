package main

import (
	"net/http"
	"github.com/go-chi/chi/v5"

	"short-linker/internal/config"
	"short-linker/internal/handler"
	"short-linker/internal/storage"
)

func main() {
	r := chi.NewRouter()
	flags := config.ParseFlags()

	storage := storage.NewMemory()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handler.MainPage(w, r, storage, flags.BaseShortURL)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		handler.RedirectPage(w, r, storage, id)
	})

	err := http.ListenAndServe(flags.Address, r)
	if err != nil {
		panic(err)
	}
}
