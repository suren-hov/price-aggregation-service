package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"price-aggregation-service/internal/model"
	"price-aggregation-service/internal/store"
)

func TestHealth_NeverUpdated(t *testing.T) {
	st := store.New()
	h := New(st, time.Minute, 10*time.Second)

	rec := httptest.NewRecorder()
	h.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 before any price is set, got %d", rec.Code)
	}
}

func TestHealth_Fresh(t *testing.T) {
	st := store.New()
	st.Update(model.Price{Value: 100, LastUpdated: time.Now(), Stale: false})
	h := New(st, time.Minute, 10*time.Second)

	rec := httptest.NewRecorder()
	h.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for fresh price, got %d", rec.Code)
	}
}

func TestHealth_ExplicitlyStale(t *testing.T) {
	st := store.New()
	st.Update(model.Price{Value: 100, LastUpdated: time.Now(), Stale: true})
	h := New(st, time.Minute, 10*time.Second)

	rec := httptest.NewRecorder()
	h.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for explicitly stale price, got %d", rec.Code)
	}
}

func TestHealth_AgedOutPastThreshold(t *testing.T) {
	st := store.New()
	st.Update(model.Price{Value: 100, LastUpdated: time.Now().Add(-time.Hour), Stale: false})
	h := New(st, time.Minute, 10*time.Second)

	rec := httptest.NewRecorder()
	h.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 once the price is older than the stale threshold, got %d", rec.Code)
	}
}

func TestPrice_ReturnsStoredValue(t *testing.T) {
	st := store.New()
	want := model.Price{Value: 42.5, Currency: "USD", SourcesUsed: 2, LastUpdated: time.Now()}
	st.Update(want)
	h := New(st, time.Minute, 10*time.Second)

	rec := httptest.NewRecorder()
	h.Price(rec, httptest.NewRequest(http.MethodGet, "/price", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}
}

func TestConfig_ReturnsPollIntervalInSeconds(t *testing.T) {
	st := store.New()
	h := New(st, time.Minute, 15*time.Second)

	rec := httptest.NewRecorder()
	h.Config(rec, httptest.NewRequest(http.MethodGet, "/config", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}

	var got configResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if got.PollIntervalSeconds != 15 {
		t.Fatalf("expected poll_interval_seconds 15, got %v", got.PollIntervalSeconds)
	}
}
