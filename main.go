package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"price-aggregation-service/internal/api"
	"price-aggregation-service/internal/aggregator"
	"price-aggregation-service/internal/client"
	"price-aggregation-service/internal/config"
	"price-aggregation-service/internal/metrics"
	"price-aggregation-service/internal/poller"
	"price-aggregation-service/internal/store"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metrics.Register()

	st := store.New()
	agg := aggregator.NewAverage()

	httpClient := http.Client{
		Timeout: cfg.RequestTimeout,
	}
	
	sources := []client.PriceSource{
		client.NewCoinbase(&httpClient),
		client.NewKraken(&httpClient),
		client.NewCryptoCompare(&httpClient),
	}

	pl := poller.New(
		sources,
		agg,
		st,
		cfg.PollInterval,
		logger,
	)

	pl.RetryConfig = poller.RetryConfig{
		MaxRetries: cfg.MaxRetries,
		BaseDelay:  cfg.BaseRetryDelay,
	}

	go pl.Start(rootCtx)

	handler := api.New(st)

	mux := http.NewServeMux()
	mux.HandleFunc("/price", handler.Price)
	mux.HandleFunc("/health", handler.Health)
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		logger.Info("server started", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	<-rootCtx.Done()

	logger.Info("shutting down server")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctxShutdown)
}