package config

import (
	"os"
)

type Config struct {
	Port        string
	BoltDBPath  string
	ServiceName string
	Environment string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("BOLT_DB_PATH")
	if dbPath == "" {
		dbPath = "/var/data/mercury-dasha.db"
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "mercury-dasha"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	return &Config{
		Port:        port,
		BoltDBPath:  dbPath,
		ServiceName: serviceName,
		Environment: env,
	}
}
