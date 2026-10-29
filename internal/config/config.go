package config

import (
	"fmt"
	"os"
	"strings"

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

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	serverPort := strings.TrimSpace(os.Getenv("SERVER_PORT"))
	if serverPort == "" {
		serverPort = "8080"
	}

	environment := strings.TrimSpace(os.Getenv("APP_ENV"))
	if environment == "" {
		environment = "development"
	}

	logLevel := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "info"
	}

	logFormat := strings.TrimSpace(os.Getenv("LOG_FORMAT"))
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
