package app

import "os"

var defaultCategories = []string{
    "General",
    "Announcements",
    "Help",
    "Off Topic",
}

func (app *App) initDB() error {
    if _, err := app.db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
        return err
    }
    if _, err := app.db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
        return err
    }

    schema, err := os.ReadFile("schema.sql")
    if err != nil {
        return err
    }
    if _, err := app.db.Exec(string(schema)); err != nil {
        return err
    }
    return app.seedCategories()
}

func (app *App) seedCategories() error {
    var count int
    err := app.db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
    if err != nil {
        return err
    }
    if count > 0 {
        return nil
    }
    for _, name := range defaultCategories {
        if _, err := app.db.Exec("INSERT INTO categories (name) VALUES (?)", name); err != nil {
            return err
        }
    }
    return nil
}
