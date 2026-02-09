package app

import (
    "database/sql"
    "errors"
    "html/template"
    "net/http"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

type App struct {
    db           *sql.DB
    templates    map[string]*template.Template
    cookieSecure bool
}

func New(cfg Config) (*App, error) {
    db, err := sql.Open("sqlite3", cfg.DBPath)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(1)

    application := &App{
        db:           db,
        cookieSecure: cfg.CookieSecure,
    }

    if err := application.initDB(); err != nil {
        return nil, err
    }
    if err := application.loadTemplates(); err != nil {
        return nil, err
    }

    return application, nil
}

func (app *App) Serve(addr string) error {
    srv := &http.Server{
        Addr:         addr,
        Handler:      app.recoverPanic(app.routes()),
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        return err
    }
    return nil
}
