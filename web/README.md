# BTC Price Aggregator — Web Dashboard

A small React + TypeScript + Tailwind dashboard for the [price-aggregation-service](../HOWTO.md) Go API. Polls `/price` and `/health` every 5 seconds and shows the current aggregated BTC price, source count, staleness, and API health.

## Running

Start the Go backend first (from the repo root):

```bash
go run .
```

Then, in this directory:

```bash
npm install
npm run dev
```

Open the printed local URL (default `http://localhost:5173`). The Vite dev server proxies `/price`, `/health`, and `/metrics` to `http://localhost:8080`, so no CORS setup is needed and no API base URL needs configuring.

## Building for production

```bash
npm run build
```

Outputs a static bundle to `dist/`. Serve it behind the same reverse proxy/origin as the Go API (or update the proxy target in `vite.config.ts` if the API lives elsewhere).
