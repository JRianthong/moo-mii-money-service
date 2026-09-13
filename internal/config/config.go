package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	DatabaseURL            string
	LineChannelSecret      string
	LineChannelAccessToken string
	DefaultCurrency        string
}

func Load() (Config, error) {
	loadDotEnv()

	cfg := Config{
		Port:                   getenv("PORT", "8080"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		LineChannelSecret:      os.Getenv("LINE_CHANNEL_SECRET"),
		LineChannelAccessToken: os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
		DefaultCurrency:        getenv("DEFAULT_CURRENCY", "THB"),
	}
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	missing := make([]string, 0, 3)
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.LineChannelSecret == "" {
		missing = append(missing, "LINE_CHANNEL_SECRET")
	}
	if c.LineChannelAccessToken == "" {
		missing = append(missing, "LINE_CHANNEL_ACCESS_TOKEN")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}
	return nil
}

func loadDotEnv() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("load .env: %v", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
