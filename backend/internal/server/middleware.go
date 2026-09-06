package server

import (
	internalmiddleware "github.com/freel/backend/internal/middleware"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) setupMiddleware() {
	// 1. Trace requests end-to-end by generating/propagating correlation IDs immediately.
	s.router.Use(internalmiddleware.TraceMiddleware)
	
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Basic CORS setup
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.cfg.FrontendURL, s.cfg.FrontendProdURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Correlation-ID"},
		ExposedHeaders:   []string{"Link", "X-Correlation-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

