package server

import (
	"net/http"

	internalmiddleware "github.com/freel/backend/internal/middleware"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) setupMiddleware() {
	// 1. Trace requests end-to-end by generating/propagating correlation IDs immediately.
	s.router.Use(internalmiddleware.TraceMiddleware)
	
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// 2. Production Security Headers
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	})

	// Basic CORS setup
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			s.cfg.FrontendURL,
			s.cfg.FrontendProdURL,
			s.cfg.SportalURL,
			s.cfg.SportalProdURL,
			"http://127.0.0.1:5173",
			"http://localhost:5173",
			"http://127.0.0.1:5174",
			"http://localhost:5174",
			"https://app.logisticshq.in",
			"https://sportal.logisticshq.in",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Correlation-ID", "X-Test-Org-ID"},
		ExposedHeaders:   []string{"Link", "X-Correlation-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

