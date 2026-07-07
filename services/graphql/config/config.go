package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	NATSURL     string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", ":8087"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://novabase_admin:novabase_secure_pass_123@novabase-postgres:5432/novabase?sslmode=disable"),
		NATSURL:     getEnv("NATS_URL", "nats://nats:4222"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
