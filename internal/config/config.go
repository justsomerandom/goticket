package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL, HTTPAddress          string
	ShutdownTimeout, JobLeaseDuration time.Duration
}

func Load() (Config, error) {
	c := Config{DatabaseURL: os.Getenv("DATABASE_URL"), HTTPAddress: os.Getenv("HTTP_ADDR"), ShutdownTimeout: 10 * time.Second, JobLeaseDuration: 5 * time.Minute}
	if c.HTTPAddress == "" {
		c.HTTPAddress = ":8080"
	}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	if raw := os.Getenv("JOB_LEASE_DURATION"); raw != "" {
		lease, err := time.ParseDuration(raw)
		if err != nil || lease <= 0 {
			return c, fmt.Errorf("JOB_LEASE_DURATION must be a positive duration")
		}
		c.JobLeaseDuration = lease
	}
	return c, nil
}
