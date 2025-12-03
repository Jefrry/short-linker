package router

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"

	"short-linker/internal/handler"
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

func (r *Router) SetupRoutes(logger *zap.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.LoggerMiddleware(logger))
	// I know, this is an overengineered, but I wanted to learn and practice it
	router.Use(middleware.GzipMiddleware)

	r.router = router

	r.baseRoutes()
	r.linkRoutes()
	r.userRoutes()

	return router
}
