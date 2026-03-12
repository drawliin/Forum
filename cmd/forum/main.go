package main

import (
	"log"

	"forum/internal/app"
)

func main() {
	cfg := app.LoadConfig()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	addr := cfg.Addr()
	log.Printf("forum listening on http://localhost%s", addr)
	if err := application.Serve(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
