package config

import (
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"PORT", "POLL_INTERVAL", "REQUEST_TIMEOUT",
		"MAX_RETRIES", "BASE_RETRY_DELAY", "STALE_THRESHOLD",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.PollInterval != 10*time.Second {
		t.Errorf("expected default poll interval 10s, got %s", cfg.PollInterval)
	}
	if cfg.RequestTimeout != 5*time.Second {
		t.Errorf("expected default request timeout 5s, got %s", cfg.RequestTimeout)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", cfg.MaxRetries)
	}
	if cfg.BaseRetryDelay != 200*time.Millisecond {
		t.Errorf("expected default base retry delay 200ms, got %s", cfg.BaseRetryDelay)
	}
	if cfg.StaleThreshold != 30*time.Second {
		t.Errorf("expected default stale threshold to be 3x poll interval (30s), got %s", cfg.StaleThreshold)
	}
}

func TestLoad_Overrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("POLL_INTERVAL", "5s")
	t.Setenv("REQUEST_TIMEOUT", "2s")
	t.Setenv("MAX_RETRIES", "7")
	t.Setenv("BASE_RETRY_DELAY", "50ms")
	t.Setenv("STALE_THRESHOLD", "1m")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Errorf("expected poll interval 5s, got %s", cfg.PollInterval)
	}
	if cfg.RequestTimeout != 2*time.Second {
		t.Errorf("expected request timeout 2s, got %s", cfg.RequestTimeout)
	}
	if cfg.MaxRetries != 7 {
		t.Errorf("expected max retries 7, got %d", cfg.MaxRetries)
	}
	if cfg.BaseRetryDelay != 50*time.Millisecond {
		t.Errorf("expected base retry delay 50ms, got %s", cfg.BaseRetryDelay)
	}
	if cfg.StaleThreshold != time.Minute {
		t.Errorf("expected stale threshold 1m, got %s", cfg.StaleThreshold)
	}
}

func TestLoad_StaleThresholdDefaultsRelativeToPollInterval(t *testing.T) {
	clearEnv(t)
	t.Setenv("POLL_INTERVAL", "20s")

	cfg := Load()

	if cfg.StaleThreshold != 60*time.Second {
		t.Errorf("expected stale threshold to scale with poll interval (60s), got %s", cfg.StaleThreshold)
	}
}

func TestLoad_MalformedDurationFallsBackToDefault(t *testing.T) {
	clearEnv(t)
	t.Setenv("POLL_INTERVAL", "not-a-duration")

	cfg := Load()

	if cfg.PollInterval != 10*time.Second {
		t.Errorf("expected malformed POLL_INTERVAL to fall back to the 10s default, got %s", cfg.PollInterval)
	}
}

func TestLoad_ZeroOrNegativeDurationFallsBackToDefault(t *testing.T) {
	for _, v := range []string{"0s", "-5s"} {
		t.Run(v, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("REQUEST_TIMEOUT", v)

			cfg := Load()

			// A zero or negative duration would either disable the HTTP
			// client's timeout entirely or panic time.NewTicker elsewhere,
			// so both must fall back to the default rather than pass through.
			if cfg.RequestTimeout != 5*time.Second {
				t.Errorf("expected REQUEST_TIMEOUT=%s to fall back to the 5s default, got %s", v, cfg.RequestTimeout)
			}
		})
	}
}

func TestLoad_MalformedIntFallsBackToDefault(t *testing.T) {
	clearEnv(t)
	t.Setenv("MAX_RETRIES", "not-a-number")

	cfg := Load()

	if cfg.MaxRetries != 3 {
		t.Errorf("expected malformed MAX_RETRIES to fall back to the default 3, got %d", cfg.MaxRetries)
	}
}

func TestLoad_NegativeIntFallsBackToDefault(t *testing.T) {
	clearEnv(t)
	t.Setenv("MAX_RETRIES", "-1")

	cfg := Load()

	// A negative MaxRetries would make the poller's retry loop
	// (`for attempt := 0; attempt <= MaxRetries; attempt++`) never execute
	// at all, silently skipping every fetch attempt.
	if cfg.MaxRetries != 3 {
		t.Errorf("expected negative MAX_RETRIES to fall back to the default 3, got %d", cfg.MaxRetries)
	}
}
