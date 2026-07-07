package config

import (
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	JWTSecret     string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", ":8081"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://strata_admin:strata_secure_pass_123@strata-postgres:5432/strata?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "strata-redis:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "redis_secure_pass_123"),
		JWTSecret:     getEnv("JWT_SECRET", "strata_super_secret_jwt_key_98765"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
