package app

import (
	"forum/internal/config"
	"forum/internal/db"
	"forum/internal/handlers"
	"forum/internal/templates"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// New prepares the shared parts of the app before requests start coming in.
func New(cfg *config.Config) error {
	if err := db.InitDB(cfg.DBPath); err != nil {
		return err
	}
	if err := templates.InitTemplates(); err != nil {
		return err
	}

	return nil
}

// Serve builds the HTTP server and starts listening on the given address.
func Serve(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handlers.SetupRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
