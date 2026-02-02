package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"short-linker/internal/handler"
	"short-linker/internal/logger"
	"short-linker/internal/middleware"
)

type Router struct {
	router    *chi.Mux
	JWTSecret string

	pingHandler *handler.PingHandler
	linkHandler *handler.LinkHandler
	userHandler *handler.UserHandler
}

func NewRouter(JWTSecret string, pingHandler *handler.PingHandler, linkHandler *handler.LinkHandler, userHandler *handler.UserHandler) *Router {
	return &Router{
		JWTSecret:   JWTSecret,
		pingHandler: pingHandler,
		linkHandler: linkHandler,
		userHandler: userHandler,
	}
}

func (r *Router) SetupRoutes(l logger.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.LoggerMiddleware(l))
	router.Use(chiMiddleware.Timeout(3 * time.Second))
	// I know, this is an overengineered, but I wanted to learn and practice it
	router.Use(middleware.GzipMiddleware)

	r.router = router

	r.baseRoutes()
	r.linkRoutes()
	r.userRoutes()

	return router
}
