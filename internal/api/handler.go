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
}

func New(store *store.Store, staleThreshold time.Duration) *Handler {
	return &Handler{store: store, staleThreshold: staleThreshold}
}

func (h *Handler) Price(w http.ResponseWriter, r *http.Request) {
	price := h.store.Get()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
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
