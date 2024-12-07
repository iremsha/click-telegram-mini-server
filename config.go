package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/caarlos0/env"
)

type Config struct {
	MongoDSN      string `env:"MONGO_DSN"`
	MongoDatabaseName   string `env:"MONGO_DB_NAME"`
	TelegramToken string `env:"TELEGRAM_TOKEN"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, relying on environment variables.")
	}

	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}

	return config, nil
}
