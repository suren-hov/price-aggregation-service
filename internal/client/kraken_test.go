package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestKraken_Fetch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":[],"result":{"XXBTZUSD":{"c":["54321.99","0.1"]}}}`))
	}))
	defer srv.Close()

	k := NewKrakenWithURL(time.Second, srv.URL)
	price, err := k.Fetch(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 54321.99 {
		t.Fatalf("expected 54321.99, got %f", price)
	}
	if k.Name() != "kraken" {
		t.Fatalf("expected name kraken, got %s", k.Name())
	}
}

func TestKraken_Fetch_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":["EQuery:Unknown asset pair"],"result":{}}`))
	}))
	defer srv.Close()

	k := NewKrakenWithURL(time.Second, srv.URL)
	if _, err := k.Fetch(t.Context()); err == nil {
		t.Fatal("expected error when Kraken reports an API error")
	}
}

func TestKraken_Fetch_EmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":[],"result":{}}`))
	}))
	defer srv.Close()

	k := NewKrakenWithURL(time.Second, srv.URL)
	if _, err := k.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for empty result")
	}
}

func TestKraken_Fetch_MissingPriceData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":[],"result":{"XXBTZUSD":{"c":[]}}}`))
	}))
	defer srv.Close()

	k := NewKrakenWithURL(time.Second, srv.URL)
	if _, err := k.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for missing price data")
	}
}

func TestKraken_Fetch_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	k := NewKrakenWithURL(time.Second, srv.URL)
	if _, err := k.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}
