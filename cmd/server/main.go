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

	envs := config.ParseEnvs()
	flags := config.ParseFlags()

	storage := storage.NewMemory()

	currentAdress := flags.Address
	if envs.Address != "" {
		currentAdress = envs.Address
	}
	currentBaseURL := flags.BaseShortURL
	if envs.BaseShortURL != "" {
		currentBaseURL = envs.BaseShortURL
	}

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		handler.MainPage(w, r, storage, currentBaseURL)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		handler.RedirectPage(w, r, storage, id)
	})

	err := http.ListenAndServe(currentAdress, r)
	if err != nil {
		panic(err)
	}
}
