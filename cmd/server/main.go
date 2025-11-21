package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	logger, err := initLoger()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	var db *sql.DB
	if cfg.DatabaseDsn != "" {
		db, err = initDB(cfg.DatabaseDsn)
		if err != nil {
			logger.Fatal("failed to initialize database", zap.Error(err))
		}
		defer db.Close()
	}

	storage := storage.NewMemory()

	linkRepo := repository.NewLinkRepository(storage)
	linkService := service.NewLinkService(linkRepo, cfg.BaseShortURL) // Do I need to pass BaseShortURL here or in repo?
	linkHandler := handler.NewLinkHandler(linkService)

	pingHandler := handler.NewPingHandler(db)

	r := router.NewRouter(pingHandler, linkHandler).SetupRoutes(logger)

	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		panic(err)
	}
}

// Where to collect all init functions?
func initLoger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err) // Do I need panic with no logger?
	}
	defer logger.Sync()
	return logger, err
}

func initDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}