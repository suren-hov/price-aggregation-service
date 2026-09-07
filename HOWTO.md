# BTC Price Aggregation Service

This service fetches BTC/USD prices from multiple public APIs (Kraken, Coinbase, CryptoCompare), aggregates them, and exposes an internal HTTP API and Prometheus metrics. It is designed to be resilient, observable, and production-ready.

---

## Table of Contents

1. [Architecture](#architecture)
2. [Configuration](#configuration)
3. [Running the Service](#running-the-service)
4. [HTTP Endpoints](#http-endpoints)
5. [Testing](#testing)
6. [Logging & Observability](#logging--observability)
7. [Future Improvements](#future-improvements)

---

## Architecture

The service follows a modular, interface-driven design:

```
main.go        # Entrypoint
internal/
├─ aggregator/           # Aggregation logic (average)
├─ api/                  # HTTP handlers for /price and /health
├─ client/               # Exchange clients: Kraken, Coinbase, CryptoCompare
├─ config/               # Environment config loader
├─ metrics/              # Prometheus metric definitions
├─ model/                # Shared Price type + staleness rules
├─ poller/               # Polling loop, retry/backoff, metric recording
└─ store/                # Thread-safe storage of last known price
```

* **Exchange Clients:** Each client implements a `Fetch(ctx) (float64, error)` method against its exchange's public (unauthenticated) endpoint. Endpoints are overridable per client (`NewXWithURL`) for testing.
* **Aggregator:** Computes the average price from whichever sources succeeded this cycle. A source failing doesn't block aggregation as long as at least one succeeds.
* **Store:** Holds the last known price safely behind a mutex.
* **HTTP API:** Exposes `/price`, `/health`, `/metrics`.
* **Concurrency:** Each poll cycle fans out one goroutine per exchange, with its own retry/backoff, bounded by a per-cycle context timeout.

---

## Configuration

All settings are configurable via **environment variables**. Defaults are used if variables are missing.

```env
# HTTP server
PORT=8080

# Polling
POLL_INTERVAL=10s
REQUEST_TIMEOUT=5s
MAX_RETRIES=3
BASE_RETRY_DELAY=200ms

# How old the last price can get before /health returns 503, even if the
# poller never explicitly marked it stale (e.g. it hung). Defaults to
# 3x POLL_INTERVAL if unset.
STALE_THRESHOLD=30s
```

The service loads these variables via the `internal/config` package. All exchange clients call public, unauthenticated endpoints — there are no API key settings.

---

## Running the Service

### 1. Local

```bash
# Install dependencies
go mod download

# Load .env
export $(cat .env | xargs)

# Run server
go run .
```

### 2. Docker

```bash
docker build -t btc-service .
docker run --env-file .env -p 8080:8080 btc-service
```

> `--network=host` is only needed as a workaround on hosts where the default
> Docker bridge network can't resolve external DNS (some corporate/VPN
> setups). Try the plain commands above first.

---

## HTTP Endpoints

| Endpoint   | Description                                 | Response Example                                                                                                      |
| ---------- | ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `/price`   | Current aggregated BTC price                | `{ "price": 66565.29, "currency": "USD", "sources_used": 3, "last_updated": "2026-02-28T21:25:38Z", "stale": false }` |
| `/health`  | Service health based on price freshness     | `200 OK` if the last price is fresh; `503` if all sources just failed, nothing has been fetched yet, or the price is older than `STALE_THRESHOLD` |
| `/metrics` | Prometheus metrics                          | `fetch_success_total`, `fetch_failure_total`, `current_price`, `source_status`                                        |

---

## Testing

Unit tests cover:

* Aggregation logic (average)
* Retry/backoff behavior, including context cancellation mid-retry
* Poller behavior: all-healthy, partial failure, all-failed (stale marking), aggregator errors
* Exchange clients against mocked HTTP responses (success, non-200, malformed JSON, invalid/zero price)
* Config defaults and env var overrides
* HTTP handlers (`/price`, `/health` under fresh/stale/never-updated/aged-out conditions)
* Store concurrency (via `-race`)

Run tests with:

```bash
go test ./... -race -cover
```

Actual coverage (excluding `main.go`, which is just wiring): `aggregator` ~89%, `api` 100%, `client` ~88%, `config` 100%, `model` 100%, `poller` ~86%, `store` 100%.

---

## Logging & Observability

Structured JSON logging (Go's standard `log/slog`) includes:

* Source (exchange)
* Fetch latency
* Errors
* Retry attempts

Prometheus metrics (`/metrics`) are updated on every poll cycle:

* `fetch_success_total{source}` / `fetch_failure_total{source}` — per-exchange counters
* `source_status{source}` — 1 if the exchange's last fetch succeeded, 0 otherwise
* `current_price` — the last successfully aggregated price

Example logs:

```json
{"time":"2026-02-28T21:25:38Z","level":"INFO","msg":"fetch success","source":"kraken","latency":531909661}
{"time":"2026-02-28T21:25:38Z","level":"WARN","msg":"retrying fetch","source":"cryptocompare","attempt":2,"error":"request failed"}
```

---

## Graceful Shutdown

* `SIGINT` / `SIGTERM` cancel the root context, which stops the poller loop and unblocks the HTTP server.
* The HTTP server is given 5s (`server.Shutdown`) to let in-flight requests complete before exiting.
* The store is a plain in-memory struct — there's nothing to close.

---

## Future Improvements

1. **Circuit Breaker** per source for heavy failure protection.
2. **Rate Limiting** on API calls to avoid hitting exchange limits.
3. **Caching** to reduce unnecessary API calls.
4. **Docker Compose** for multi-service deployments (Prometheus, Grafana).
5. **Benchmark Tests** for aggregation and polling latency.
6. **Median/weighted aggregator** as an alternative to the current simple average.

---

**Authors / Maintainers**: Suren Hovhannisyan
