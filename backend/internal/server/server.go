package server

import (
	"log"
	"net/http"

	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/config"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	cfg         *config.Config
	router      *chi.Mux
	authService *auth.Service
}

func NewServer(cfg *config.Config, authService *auth.Service) *Server {
	s := &Server{
		cfg:         cfg,
		router:      chi.NewRouter(),
		authService: authService,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.cfg.Port)
	return http.ListenAndServe(":"+s.cfg.Port, s.router)
}
