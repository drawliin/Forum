package main

import (
	"fmt"
	"log"

	"forum/internal/app"
	"forum/internal/config"
)

func main() {
	cfg := config.GetConfig()
	fmt.Println(cfg)

	err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	addr := ":" + cfg.Port
	log.Printf("forum listening on http://localhost%s", addr)
	if err := app.Serve(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
