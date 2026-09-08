package poller

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"price-aggregation-service/internal/aggregator"
	"price-aggregation-service/internal/client"
	"price-aggregation-service/internal/metrics"
	"price-aggregation-service/internal/model"
	"price-aggregation-service/internal/store"
)

var errCircuitOpen = errors.New("circuit open: skipping fetch")

type Poller struct {
	sources     []client.PriceSource
	aggregator  aggregator.Aggregator
	store       *store.Store
	interval    time.Duration
	logger      *slog.Logger
	circuits    *circuitBreaker
	RetryConfig RetryConfig
}

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
}

type fetchResult struct {
	source  string
	price   float64
	err     error
	skipped bool
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
		circuits:   newCircuitBreaker(),
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
		name := src.Name()

		if !p.circuits.Allow(name) {
			p.logger.Warn("skipping fetch: circuit open", "source", name)
			metrics.CircuitOpen.WithLabelValues(name).Set(1)
			metrics.SourceStatus.WithLabelValues(name).Set(0)
			resultsCh <- fetchResult{source: name, err: errCircuitOpen, skipped: true}
			continue
		}

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
				metrics.FetchFailure.WithLabelValues(s.Name()).Inc()
				metrics.SourceStatus.WithLabelValues(s.Name()).Set(0)
			} else {
				p.logger.Info("fetch success",
					"source", s.Name(),
					"latency", latency,
				)
				metrics.FetchSuccess.WithLabelValues(s.Name()).Inc()
				metrics.SourceStatus.WithLabelValues(s.Name()).Set(1)
			}

			resultsCh <- fetchResult{
				source: s.Name(),
				price:  price,
				err:    err,
			}
		}(src)
	}

	var validPrices []aggregator.PricePoint

	for i := 0; i < len(p.sources); i++ {
		res := <-resultsCh

		if !res.skipped {
			if opened := p.circuits.RecordResult(res.source, res.err == nil); opened {
				p.logger.Warn("circuit opened after repeated failures", "source", res.source)
				metrics.CircuitOpen.WithLabelValues(res.source).Set(1)
			} else {
				metrics.CircuitOpen.WithLabelValues(res.source).Set(0)
			}
		}

		if res.err == nil {
			validPrices = append(validPrices, aggregator.PricePoint{Source: res.source, Price: res.price})
		}
	}

	current := p.store.Get()

	if len(validPrices) == 0 {
		// All failed → mark stale
		current.Stale = true
		p.store.Update(current)
		return
	}

	result, err := p.aggregator.Aggregate(validPrices)
	if err != nil {
		p.logger.Error("aggregation failed", "error", err)
		return
	}

	if len(result.ExcludedSources) > 0 {
		p.logger.Warn("aggregator excluded outlier sources",
			"excluded", result.ExcludedSources,
			"used", result.UsedSources,
		)
		for _, name := range result.ExcludedSources {
			metrics.AggregatorExcluded.WithLabelValues(name).Inc()
		}
	}

	metrics.CurrentPrice.Set(result.Value)

	p.store.Update(model.Price{
		Value:       result.Value,
		Currency:    "USD",
		SourcesUsed: len(result.UsedSources),
		LastUpdated: time.Now().UTC(),
		Stale:       false,
	})
}
