package config

import (
	"os"
)

type Config struct {
	Host      string
	Port      string
	DBPath    string
	AuthToken string
	CorsOrigin string
}

func Load() *Config {
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "autodidact.db"
	}

	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		authToken = "autodidact-local-key"
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}

	return &Config{
		Host:       host,
		Port:       port,
		DBPath:     dbPath,
		AuthToken:  authToken,
		CorsOrigin: corsOrigin,
	}
}
