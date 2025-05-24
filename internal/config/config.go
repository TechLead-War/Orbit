package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		ServerPort: "8080",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBName:     "orbit",
	}
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := DefaultConfig()

	if port := getEnvOrDefault("PORT", cfg.ServerPort); port != "" {
		cfg.ServerPort = port
	}
	if host := getEnvOrDefault("DB_HOST", cfg.DBHost); host != "" {
		cfg.DBHost = host
	}
	if port := getEnvOrDefault("DB_PORT", cfg.DBPort); port != "" {
		cfg.DBPort = port
	}
	if user := getEnvOrDefault("DB_USER", cfg.DBUser); user != "" {
		cfg.DBUser = user
	}
	if password := getEnvOrDefault("DB_PASSWORD", cfg.DBPassword); password != "" {
		cfg.DBPassword = password
	}
	if name := getEnvOrDefault("DB_NAME", cfg.DBName); name != "" {
		cfg.DBName = name
	}

	return cfg
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
