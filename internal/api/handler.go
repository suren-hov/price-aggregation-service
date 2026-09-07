package api

import (
	"encoding/json"
	"net/http"
	"time"

	"price-aggregation-service/internal/store"
)

type Handler struct {
	store          *store.Store
	staleThreshold time.Duration
	pollInterval   time.Duration
}

func New(store *store.Store, staleThreshold, pollInterval time.Duration) *Handler {
	return &Handler{store: store, staleThreshold: staleThreshold, pollInterval: pollInterval}
}

func (h *Handler) Price(w http.ResponseWriter, r *http.Request) {
	price := h.store.Get()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}

type configResponse struct {
	PollIntervalSeconds float64 `json:"poll_interval_seconds"`
}

// Config exposes the settings a client needs to interpret the API correctly
// (e.g. how often a new price can be expected), so the frontend doesn't have
// to duplicate backend defaults.
func (h *Handler) Config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configResponse{
		PollIntervalSeconds: h.pollInterval.Seconds(),
	})
}

// Health reports 503 when the last price is explicitly marked stale, has
// never been populated (nothing fetched yet), or is older than the
// configured staleness threshold (the poller stopped making progress
// without ever flipping the Stale flag).
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	price := h.store.Get()

	if price.IsStale(h.staleThreshold, time.Now()) {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
