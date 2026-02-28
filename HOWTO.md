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
├─ aggregator/           # Aggregation logic (average/median)
├─ client/               # Exchange clients: Kraken, Coinbase, CryptoCompare
├─ config/               # Environment config loader
└─ store/                # Thread-safe storage of last known price
```

* **Exchange Clients:** Each client implements a `Fetch(ctx) (float64, error)` method.
* **Aggregator:** Computes average price from available sources. Falls back if one or more sources fail.
* **Store:** Holds last known price safely with mutex.
* **HTTP API:** Exposes `/price`, `/health`, `/metrics`.
* **Concurrency:** Pollers run concurrently with proper context cancellation and timeouts.

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

# Optional API keys (for private endpoints)
COINBASE_KEY=
COINBASE_SECRET=
COINBASE_PASSPHRASE=
KRAKEN_KEY=
KRAKEN_SECRET=
```

The service loads these variables via `internal/config` package.

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

> Ensure Docker DNS works correctly for external APIs (see HOWTO note below).

---

## HTTP Endpoints

| Endpoint   | Description                                 | Response Example                                                                                                      |
| ---------- | ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `/price`   | Current aggregated BTC price                | `{ "price": 66565.29, "currency": "USD", "sources_used": 3, "last_updated": "2026-02-28T21:25:38Z", "stale": false }` |
| `/health`  | Service health based on source availability | `200 OK` if ≥1 source healthy, `503` if all failing                                                                   |
| `/metrics` | Prometheus metrics                          | `fetch_success_total`, `fetch_failure_total`, `current_price`, `source_status`                                        |

---

## Testing

Unit tests cover:

* Aggregation logic (average, median)
* Retry behavior
* Failure handling for individual clients
* Mocked API responses

Run tests with:

```bash
go test ./...
```

Coverage target: **60–70% meaningful coverage**.

---

## Logging & Observability

Structured logging (using `slog`/`zap`) includes:

* Source (exchange)
* Fetch latency
* Errors
* Retry attempts

Example logs:

```json
{"time":"2026-02-28T21:25:38Z","level":"INFO","msg":"fetch success","source":"kraken","latency":531909661}
{"time":"2026-02-28T21:25:38Z","level":"WARN","msg":"retrying fetch","source":"cryptocompare","attempt":2,"error":"request failed"}
```

---

## Graceful Shutdown

* Pollers listen for `SIGINT` / `SIGTERM`
* In-flight requests complete
* HTTP server stops with timeout
* Store closes safely

---

## Future Improvements

1. **Circuit Breaker** per source for heavy failure protection.
2. **Rate Limiting** on API calls to avoid hitting exchange limits.
3. **Caching** to reduce unnecessary API calls.
4. **Docker Compose** for multi-service deployments (Prometheus, Grafana).
5. **Benchmark Tests** for aggregation and polling latency.
6. **Optional Authenticated Endpoints** (balance, orders) using stored API keys.

---

**Authors / Maintainers**: Suren Hovhannisyan
