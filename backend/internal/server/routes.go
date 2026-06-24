package server

import (
	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/health"
	"github.com/go-chi/chi/v5"
)

func (s *Server) setupRoutes() {
	// Health
	s.router.Get("/health", health.HealthCheck)

	// Auth Handlers
	authHandler := auth.NewHandler(s.authService)

	s.router.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/verify-email", authHandler.VerifyEmail)
		r.Post("/login", authHandler.Login)
		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.Post("/reset-password", authHandler.ResetPassword)
		r.Get("/me", authHandler.GetMe)
	})
}
