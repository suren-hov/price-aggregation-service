package config

import (
	"log"
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
	pollInterval := getPositiveDuration("POLL_INTERVAL", 10*time.Second)

	return &Config{
		Port:           getEnv("PORT", "8080"),
		PollInterval:   pollInterval,
		RequestTimeout: getPositiveDuration("REQUEST_TIMEOUT", 5*time.Second),
		MaxRetries:     getNonNegativeInt("MAX_RETRIES", 3),
		BaseRetryDelay: getPositiveDuration("BASE_RETRY_DELAY", 200*time.Millisecond),
		// A price is considered stale once it's older than 3 poll cycles,
		// which tolerates a couple of missed/slow cycles before /health flips.
		StaleThreshold: getPositiveDuration("STALE_THRESHOLD", 3*pollInterval),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getNonNegativeInt returns def if the env var is unset, fails to parse, or
// is negative - silently accepting a malformed value here would surface as
// a confusing runtime bug far from its cause (e.g. MAX_RETRIES=-1 disabling
// every fetch attempt entirely).
func getNonNegativeInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	i, err := strconv.Atoi(v)
	if err != nil || i < 0 {
		log.Printf("config: invalid %s=%q, using default %d", key, v, def)
		return def
	}
	return i
}

// getPositiveDuration returns def if the env var is unset, fails to parse,
// or is zero/negative. Without this, a malformed duration silently becomes
// 0 (time.ParseDuration's zero value on error) and time.NewTicker(0) - used
// for POLL_INTERVAL - panics, taking down the whole process over a typo in
// an env var.
func getPositiveDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		log.Printf("config: invalid %s=%q, using default %s", key, v, def)
		return def
	}
	return d
}
