package main

import (
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/server"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	s, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	err = s.Start()
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
