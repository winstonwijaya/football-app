package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port string
	DB   DBConfig
	JWT  JWTConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

func Load() Config {
	return Config{
		Port: getEnv("APP_PORT", "8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "football"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: loadJWTConfig(),
	}
}

func loadJWTConfig() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
		log.Println("WARNING: JWT_SECRET not set, using an insecure default. Set JWT_SECRET in any non-local environment.")
	}

	tokenDuration, err := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", "60"))
	if err != nil || tokenDuration <= 0 {
		tokenDuration = 60
	}

	return JWTConfig{
		Secret: secret,
		TTL:    time.Duration(tokenDuration) * time.Minute,
	}
}

// DSN builds the connection string GORM's postgres driver expects.
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
