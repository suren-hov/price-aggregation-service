package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	PollInterval   time.Duration
	RequestTimeout time.Duration
	MaxRetries     int
	BaseRetryDelay time.Duration
	StaleThreshold time.Duration
}

func Load() *Config {
	pollInterval := getDuration("POLL_INTERVAL", 10*time.Second)

	return &Config{
		Port:           getEnv("PORT", "8080"),
		PollInterval:   pollInterval,
		RequestTimeout: getDuration("REQUEST_TIMEOUT", 5*time.Second),
		MaxRetries:     getInt("MAX_RETRIES", 3),
		BaseRetryDelay: getDuration("BASE_RETRY_DELAY", 200*time.Millisecond),
		// A price is considered stale once it's older than 3 poll cycles,
		// which tolerates a couple of missed/slow cycles before /health flips.
		StaleThreshold: getDuration("STALE_THRESHOLD", 3*pollInterval),
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
