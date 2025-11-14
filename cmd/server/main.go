package main

import (
	"net/http"
	"go.uber.org/zap"


	"short-linker/internal/config"
	"short-linker/internal/handler"
	"short-linker/internal/repository"
	"short-linker/internal/router"
	"short-linker/internal/service"
	"short-linker/internal/storage"
)

func main() {
	cfg := config.GetConfig()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err) // Do I need panic with no logger?
	}
	defer logger.Sync()

	storage := storage.NewMemory()

	linkRepo := repository.NewLinkRepository(storage)
	linkService := service.NewLinkService(linkRepo, cfg.BaseShortURL) // Do I need to pass BaseShortURL here or in repo?
	linkHandler := handler.NewLinkHandler(linkService)

	r := router.NewRouter(linkHandler).SetupRoutes(logger)

	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		panic(err)
	}
}
