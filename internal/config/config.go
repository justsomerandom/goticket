package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL, HTTPAddress string
	ShutdownTimeout          time.Duration
}

func Load() (Config, error) {
	c := Config{DatabaseURL: os.Getenv("DATABASE_URL"), HTTPAddress: os.Getenv("HTTP_ADDR"), ShutdownTimeout: 10 * time.Second}
	if c.HTTPAddress == "" {
		c.HTTPAddress = ":8080"
	}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	return c, nil
}
