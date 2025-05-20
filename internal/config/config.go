package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Server configuration
	ServerPort string

	// LeetCode API configuration (for future use)
	LeetCodeAPIKey string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "orbit_db"),

		// Server configuration
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// LeetCode API configuration
		LeetCodeAPIKey: getEnv("LEETCODE_API_KEY", ""),
	}
}

// getEnv reads an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
