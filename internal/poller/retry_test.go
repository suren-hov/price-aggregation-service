package poller

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func newTestPoller(t *testing.T, retries int, baseDelay time.Duration) *Poller {
	t.Helper()
	p := New(nil, nil, nil, time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
	p.RetryConfig = RetryConfig{MaxRetries: retries, BaseDelay: baseDelay}
	return p
}

func TestFetchWithRetry_SucceedsFirstTry(t *testing.T) {
	p := newTestPoller(t, 3, time.Millisecond)
	calls := 0

	price, err := p.fetchWithRetry(context.Background(), "test", func(ctx context.Context) (float64, error) {
		calls++
		return 123.45, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 123.45 {
		t.Fatalf("expected 123.45, got %f", price)
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call, got %d", calls)
	}
}

func TestFetchWithRetry_SucceedsAfterFailures(t *testing.T) {
	p := newTestPoller(t, 3, time.Millisecond)
	calls := 0

	price, err := p.fetchWithRetry(context.Background(), "test", func(ctx context.Context) (float64, error) {
		calls++
		if calls < 3 {
			return 0, errors.New("transient failure")
		}
		return 99, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 99 {
		t.Fatalf("expected 99, got %f", price)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestFetchWithRetry_ExhaustsRetriesAndReturnsLastError(t *testing.T) {
	p := newTestPoller(t, 2, time.Millisecond)
	calls := 0
	wantErr := errors.New("permanent failure")

	_, err := p.fetchWithRetry(context.Background(), "test", func(ctx context.Context) (float64, error) {
		calls++
		return 0, wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected last error to be returned, got %v", err)
	}
	// MaxRetries=2 means the initial attempt plus 2 retries = 3 calls.
	if calls != 3 {
		t.Fatalf("expected 3 calls (1 initial + 2 retries), got %d", calls)
	}
}

func TestFetchWithRetry_StopsOnContextCancellation(t *testing.T) {
	p := newTestPoller(t, 5, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := p.fetchWithRetry(ctx, "test", func(ctx context.Context) (float64, error) {
		calls++
		return 0, errors.New("always fails")
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls >= 6 {
		t.Fatalf("expected retries to stop early due to cancellation, got %d calls", calls)
	}
}
