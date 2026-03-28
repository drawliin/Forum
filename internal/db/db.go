package db

import (
	"database/sql"
	"os"
)

var defaultCategories = []string{
	"General",
	"Announcements",
	"Help",
	"Off Topic",
}

var Database *sql.DB

func InitDB(dbPath string) error {
	var err error
	Database, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	Database.SetMaxOpenConns(1)

	if _, err := Database.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}
	if _, err := Database.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return err
	}

	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	if _, err := Database.Exec(string(schema)); err != nil {
		return err
	}
	return seedCategories()
}

func seedCategories() error {
	var count int
	err := Database.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, name := range defaultCategories {
		if _, err := Database.Exec("INSERT INTO categories (name) VALUES (?)", name); err != nil {
			return err
		}
	}
	return nil
}
