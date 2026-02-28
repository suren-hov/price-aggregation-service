package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	PollInterval    time.Duration
	RequestTimeout  time.Duration
	MaxRetries      int
	BaseRetryDelay  time.Duration
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		PollInterval:   getDuration("POLL_INTERVAL", 10*time.Second),
		RequestTimeout: getDuration("REQUEST_TIMEOUT", 3*time.Second),
		MaxRetries:     getInt("MAX_RETRIES", 2),
		BaseRetryDelay: getDuration("BASE_RETRY_DELAY", 200*time.Millisecond),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		i, _ := strconv.Atoi(v)
		return i
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, _ := time.ParseDuration(v)
		return d
	}
	return def
}