package main

import (
	"database/sql"
	"log"
	"net/http"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

		err = initMigration(cfg.DatabaseDsn)
		if err != nil {
			logger.Fatal("failed to run database migrations", zap.Error(err))
		}
	}

	storage := storage.NewMemory()

	tokenService := service.NewTokenService(cfg.JWTSecret)

	pingHandler := handler.NewPingHandler(db)

	linkRepo := repository.NewLinkRepository(storage, db)
	linkService := service.NewLinkService(linkRepo, cfg.BaseShortURL) // Do I need to pass BaseShortURL here or in repo?
	linkHandler := handler.NewLinkHandler(linkService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, tokenService)
	userHandler := handler.NewUserHandler(userService)

	r := router.NewRouter(pingHandler, linkHandler, userHandler).SetupRoutes(logger)

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

func initDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initMigration(dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
