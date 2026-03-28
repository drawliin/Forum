package main

import (
	"log"

	"forum/internal/app"
	"forum/internal/config"
)

func main() {
	cfg := config.GetConfig()

	err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	addr := ":" + cfg.Port
	log.Printf("forum listening on http://localhost%s", addr)
	if err := app.Serve(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
