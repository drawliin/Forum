package config

import (
	"os"
	"strings"
)

// Config keeps the small set of settings used by the app at startup.
type Config struct {
	Port         string
	DBPath       string
	CookieSecure bool
}

var config *Config = &Config{
	Port:         getenv("PORT", "8080"),
	DBPath:       ResolvePath(getenv("DB_PATH", "./data/forum.db")),
	CookieSecure: getenv("COOKIE_SECURE", "") == "1",
}

// GetConfig returns the app config that was built from the environment.
func GetConfig() *Config {
	return config
}

// getenv reads an env var and falls back if it is empty.
func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
