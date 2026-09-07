package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCoinbase_Fetch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"amount":"67890.12","base":"BTC","currency":"USD"}}`))
	}))
	defer srv.Close()

	c := NewCoinbaseWithURL(time.Second, srv.URL)
	price, err := c.Fetch(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 67890.12 {
		t.Fatalf("expected 67890.12, got %f", price)
	}
	if c.Name() != "coinbase" {
		t.Fatalf("expected name coinbase, got %s", c.Name())
	}
}

func TestCoinbase_Fetch_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewCoinbaseWithURL(time.Second, srv.URL)
	if _, err := c.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestCoinbase_Fetch_EmptyAmount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"amount":"","base":"BTC","currency":"USD"}}`))
	}))
	defer srv.Close()

	c := NewCoinbaseWithURL(time.Second, srv.URL)
	if _, err := c.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for empty amount")
	}
}

func TestCoinbase_Fetch_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	c := NewCoinbaseWithURL(time.Second, srv.URL)
	if _, err := c.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestCoinbase_Fetch_InvalidPriceFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"amount":"not-a-number","base":"BTC","currency":"USD"}}`))
	}))
	defer srv.Close()

	c := NewCoinbaseWithURL(time.Second, srv.URL)
	if _, err := c.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for invalid price format")
	}
}
