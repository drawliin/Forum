package app

import (
	"errors"
	"forum/internal/config"
	"forum/internal/db"
	"forum/internal/handlers"
	"forum/internal/templates"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func New(cfg *config.Config) error {
	if err := db.InitDB(cfg.DBPath); err != nil {
		return err
	}
	if err := templates.InitTemplates(); err != nil {
		return err
	}

	return nil
}

func Serve(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handlers.RecoverPanic(handlers.SetupRoutes()),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
