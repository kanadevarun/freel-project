package server

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Basic CORS setup
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.cfg.FrontendURL, s.cfg.FrontendProdURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}
