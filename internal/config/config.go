package config

import (
	"os"
	"strings"
)

type Config struct {
	Port         string
	DBPath       string
	CookieSecure bool
}

var config *Config = &Config{
	Port:         getenv("PORT", "8080"),
	DBPath:       getenv("DB_PATH", "./data/forum.db"),
	CookieSecure: getenv("COOKIE_SECURE", "") == "1",
}

func GetConfig() *Config {
	return config
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
