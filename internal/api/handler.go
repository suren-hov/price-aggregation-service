package api

import (
	"encoding/json"
	"net/http"

	"price-aggregation-service/internal/store"
)

type Handler struct {
	store *store.Store
}

func New(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Price(w http.ResponseWriter, r *http.Request) {
	price := h.store.Get()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	price := h.store.Get()

	if price.Stale {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}