package poller

import (
	"context"
	"log/slog"
	"time"

	"price-aggregation-service/internal/aggregator"
	"price-aggregation-service/internal/client"
	"price-aggregation-service/internal/store"
	"price-aggregation-service/internal/model"
)

type Poller struct {
	sources    []client.PriceSource
	aggregator aggregator.Aggregator
	store      *store.Store
	interval   time.Duration
	logger     *slog.Logger
	RetryConfig RetryConfig
}

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
}

type fetchResult struct {
	source string
	price  float64
	err    error
}

func New(
	sources []client.PriceSource,
	agg aggregator.Aggregator,
	st *store.Store,
	interval time.Duration,
	logger *slog.Logger,
) *Poller {
	return &Poller{
		sources:    sources,
		aggregator: agg,
		store:      st,
		interval:   interval,
		logger:     logger,
	}
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("poller stopped")
			return
		case <-ticker.C:
			p.pollOnce(ctx)
		}
	}	
}

func (p *Poller) pollOnce(parentCtx context.Context) {
	// Global timeout for whole polling cycle
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	resultsCh := make(chan fetchResult, len(p.sources))

	for _, src := range p.sources {
		go func(s client.PriceSource) {
			start := time.Now()

			price, err := p.fetchWithRetry(
				ctx,
				s.Name(),
				s.Fetch,
			)

			latency := time.Since(start)

			if err != nil {
				p.logger.Error("fetch failed",
					"source", s.Name(),
					"latency", latency,
					"error", err,
				)
			} else {
				p.logger.Info("fetch success",
					"source", s.Name(),
					"latency", latency,
				)
			}

			resultsCh <- fetchResult{
				source: s.Name(),
				price:  price,
				err:    err,
			}
		}(src)
	}

	var validPrices []float64
	healthySources := 0

	for i := 0; i < len(p.sources); i++ {
		res := <-resultsCh

		if res.err == nil {
			validPrices = append(validPrices, res.price)
			healthySources++
		}
	}

	current := p.store.Get()

	if healthySources == 0 {
		// All failed → mark stale
		current.Stale = true
		p.store.Update(current)
		return
	}

	aggregated, err := p.aggregator.Aggregate(validPrices)
	if err != nil {
		p.logger.Error("aggregation failed", "error", err)
		return
	}

	p.store.Update(model.Price{
		Value:       aggregated,
		Currency:    "USD",
		SourcesUsed: healthySources,
		LastUpdated: time.Now().UTC(),
		Stale:       false,
	})
}