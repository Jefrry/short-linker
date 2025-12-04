package main

import (
	"log"
	"net/http"
	"go.uber.org/zap"

	"short-linker/internal/app"
	"short-linker/internal/config"
)

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

	if err := http.ListenAndServe(cfg.Address, a.Router); err != nil {
		a.Logger.Fatal("server error", zap.Error(err))
	}
}