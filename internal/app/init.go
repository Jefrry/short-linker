package app

import (
	"database/sql"
	"fmt"
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

type App struct {
	Cfg    *config.Config
	Logger *zap.Logger
	DB     *sql.DB
	Router http.Handler
}

func NewApp(cfg *config.Config) (*App, error) {
	logger, err := initLogger()
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	var db *sql.DB
	if cfg.DatabaseDsn != "" {
		db, err = initDB(cfg.DatabaseDsn)
		if err != nil {
			logger.Error("failed to initialize database", zap.Error(err))
			return nil, fmt.Errorf("init db: %w", err)
		}

		if err = initMigration(cfg.DatabaseDsn); err != nil {
			logger.Error("failed to run database migrations", zap.Error(err))
			return nil, fmt.Errorf("init migrations: %w", err)
		}
	}

	memStorage := storage.NewMemory()

	linkRepo := repository.NewLinkRepository(memStorage, db)
	userRepo := repository.NewUserRepository(db)

	linkService := service.NewLinkService(linkRepo, cfg.BaseShortURL)
	userService := service.NewUserService(service.NewTokenService(cfg.JWTSecret), userRepo, linkRepo)

	pingHandler := handler.NewPingHandler(db)
	linkHandler := handler.NewLinkHandler(logger, linkService)
	userHandler := handler.NewUserHandler(logger, userService)

	r := router.NewRouter(cfg.JWTSecret, pingHandler, linkHandler, userHandler).SetupRoutes(logger)

	return &App{
		Cfg:    cfg,
		Logger: logger,
		DB:     db,
		Router: r,
	}, nil
}


func initLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func initDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func initMigration(dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
