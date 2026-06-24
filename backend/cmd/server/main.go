package main

import (
	"log"

	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	authService := auth.NewService(cfg)
	srv := server.NewServer(cfg, authService)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
