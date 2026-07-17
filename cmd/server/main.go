package main

import (
	"log"

	"github.com/impossibleclone/fserver/internal/auth"
	"github.com/impossibleclone/fserver/internal/config"
	"github.com/impossibleclone/fserver/internal/server"
)

func main() {
	cfg := config.DefaultConfig()
	
	authenticator := auth.NewMemoryAuth()
	if err := authenticator.AddUser(cfg.Username, cfg.Password); err != nil {
		log.Fatalf("Failed to add default user: %v", err)
	}
	
	srv := server.NewFileServer(cfg, authenticator)
	log.Printf("Starting server on port %s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
