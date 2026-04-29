package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host        string
	Port        int
	DevMode     bool
	DatabaseURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Host:        getEnv("HOST", "localhost"),
		Port:        getEnvAsInt("PORT", 8000),
		DevMode:     getEnv("DEV_MODE", "false") == "true",
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/otterboard?sslmode=disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid value for %s: %s, using default: %d\n", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}
