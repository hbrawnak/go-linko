package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/hbrawnak/go-linko/internal/handlers"
	"net/http"
)

// SetupRoutes creates and configures the router with all routes
func SetupRoutes(handler *handlers.AppHandler) http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	mux.Get("/", handler.HandleMain)
	mux.Post("/shorten", handler.HandleShorten)
	mux.Get("/{code}", handler.HandleRedirect)

	return mux
}
