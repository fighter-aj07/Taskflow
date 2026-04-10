package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	APIPort    string
	JWTSecret  string
	BcryptCost int
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func Load() Config {
	cost, err := strconv.Atoi(getEnv("BCRYPT_COST", "12"))
	if err != nil || cost < 10 {
		cost = 12
	}

	return Config{
		DBHost:     getEnv("POSTGRES_HOST", "localhost"),
		DBPort:     getEnv("POSTGRES_PORT", "5432"),
		DBUser:     getEnv("POSTGRES_USER", "taskflow"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "taskflow_secret"),
		DBName:     getEnv("POSTGRES_DB", "taskflow"),
		APIPort:    getEnv("API_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
		BcryptCost: cost,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
