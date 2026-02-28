package poller

import (
	"context"
	"time"
)

func (p *Poller) fetchWithRetry(
	ctx context.Context,
	sourceName string,
	fetchFn func(ctx context.Context) (float64, error),
) (float64, error) {

	var lastErr error

	for attempt := 0; attempt <= p.RetryConfig.MaxRetries; attempt++ {

		// Check if context already cancelled
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}

		price, err := fetchFn(ctx)
		if err == nil {
			return price, nil
		}

		lastErr = err

		p.logger.Warn("retrying fetch",
			"source", sourceName,
			"attempt", attempt,
			"error", err,
		)

		// Exponential backoff
		backoff := p.RetryConfig.BaseDelay * (1 << attempt)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	return 0, lastErr
}