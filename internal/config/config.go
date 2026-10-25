package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
	Environment string
	LogLevel    string
	LogFormat   string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "" {
		if environment == "production" {
			logFormat = "json"
		} else {
			logFormat = "text"
		}
	}

	return &Config{
		DatabaseURL: databaseURL,
		ServerPort:  serverPort,
		Environment: environment,
		LogLevel:    logLevel,
		LogFormat:   logFormat,
	}, nil
}
