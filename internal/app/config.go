package app

type Config struct {
    Port         string
    DBPath       string
    CookieSecure bool
}

func LoadConfig() Config {
    return Config{
        Port:         getenv("PORT", "8080"),
        DBPath:       getenv("DB_PATH", "./data/forum.db"),
        CookieSecure: getenv("COOKIE_SECURE", "") == "1",
    }
}

func (c Config) Addr() string {
    return ":" + c.Port
}
