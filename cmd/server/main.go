package main

//go:generate swag init -g cmd/server/main.go --parseInternal -d ../../ -o ../../docs

import (
	"log"
	"net/http"

	"short-linker/internal/app"
	"short-linker/internal/config"
	"short-linker/internal/logger"
)

// @title Short Linker API
// @version 1.0
// @description URL shortener service with user authentication

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token for authentication (format: Bearer <token>)

func main() {
	cfg := config.GetConfig()

	a, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	defer func() {
		if a.DB != nil {
			_ = a.DB.Close()
		}
		_ = a.Logger.Sync()
	}()

	a.Logger.Info("app is running", logger.String("address", cfg.Address))

	if err := http.ListenAndServe(cfg.Address, a.Router); err != nil {
		a.Logger.Fatal("server error", logger.Error(err))
	}
}
