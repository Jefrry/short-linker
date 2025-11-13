package main

import (
	"net/http"

	"short-linker/internal/config"
	"short-linker/internal/handler"
	"short-linker/internal/repository"
	"short-linker/internal/router"
	"short-linker/internal/service"
	"short-linker/internal/storage"
)

func main() {
	cfg := config.GetConfig()

	storage := storage.NewMemory()

	linkRepo := repository.NewLinkRepository(storage)
	linkService := service.NewLinkService(linkRepo, cfg.BaseShortURL) // Do I need to pass BaseShortURL here or in repo?
	linkHandler := handler.NewLinkHandler(linkService)

	r := router.NewRouter(linkHandler).SetupRoutes()

	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		panic(err)
	}
}
