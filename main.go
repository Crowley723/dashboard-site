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

	if err := server.Start(cfg); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
