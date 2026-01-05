package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port        string
	LogLevel    string
	Environment string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Cache
	CacheHost string
	CachePort int

	// CORS
	CORSOrigins string

	// ODK Central
	ODKBaseURL     string
	ODKEmail       string
	ODKPassword    string
	ODKProjectID   int
	ODKFormID      string
	ODKFeedFormID  string

	// Storage
	PhotoStoragePath string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("API_PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "senyar"),
		DBPassword:  getEnv("DB_PASSWORD", "senyar123"),
		DBName:      getEnv("DB_NAME", "senyar"),
		CacheHost:   getEnv("CACHE_HOST", "localhost"),
		CachePort:   getEnvInt("CACHE_PORT", 6379),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000"),
		// ODK Central
		ODKBaseURL:    getEnv("ODK_BASE_URL", "https://data.dayawarga.com"),
		ODKEmail:      getEnv("ODK_EMAIL", ""),
		ODKPassword:   getEnv("ODK_PASSWORD", ""),
		ODKProjectID:  getEnvInt("ODK_PROJECT_ID", 3),
		ODKFormID:        getEnv("ODK_FORM_ID", "form_posko_v1"),
		ODKFeedFormID:    getEnv("ODK_FEED_FORM_ID", "form_feed_v1"),
		PhotoStoragePath: getEnv("PHOTO_STORAGE_PATH", "./storage/photos"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
