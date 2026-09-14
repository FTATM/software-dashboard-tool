package app

import (
	"context"
	"os"
	"strconv"
)

type App interface {
	Run(ctx context.Context) error
	Close(ctx context.Context)
}

func GetEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func GetEnvIntOrDefault(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}
