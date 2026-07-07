package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	JWTSecret     string
	AuthURL       string
	RestURL       string
	GraphqlURL    string
	RealtimeURL   string
	StorageURL    string
	FunctionsURL  string
	AiURL         string
	RateLimitMax  int
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", ":8000"),
		RedisAddr:     getEnv("REDIS_ADDR", "novabase-redis:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "redis_secure_pass_123"),
		JWTSecret:     getEnv("JWT_SECRET", "novabase_super_secret_jwt_key_98765"),
		AuthURL:       getEnv("AUTH_SERVICE_URL", "http://auth:8081"),
		RestURL:       getEnv("REST_SERVICE_URL", "http://rest:8082"),
		GraphqlURL:    getEnv("GRAPHQL_SERVICE_URL", "http://graphql:8087"),
		RealtimeURL:   getEnv("REALTIME_SERVICE_URL", "http://realtime:8083"),
		StorageURL:    getEnv("STORAGE_SERVICE_URL", "http://storage:8084"),
		FunctionsURL:  getEnv("FUNCTIONS_SERVICE_URL", "http://functions:8085"),
		AiURL:         getEnv("AI_SERVICE_URL", "http://ai:8086"),
		RateLimitMax:  getEnvInt("RATE_LIMIT_MAX", 100),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if valStr, exists := os.LookupEnv(key); exists {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return defaultValue
}
