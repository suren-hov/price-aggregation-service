package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBitstamp_Fetch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"last":"65123.45","bid":"65123.44","ask":"65123.45"}`))
	}))
	defer srv.Close()

	b := NewBitstampWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	price, err := b.Fetch(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 65123.45 {
		t.Fatalf("expected 65123.45, got %f", price)
	}
	if b.Name() != "bitstamp" {
		t.Fatalf("expected name bitstamp, got %s", b.Name())
	}
}

func TestBitstamp_Fetch_EmptyLast(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"last":""}`))
	}))
	defer srv.Close()

	b := NewBitstampWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	if _, err := b.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for empty last price")
	}
}

func TestBitstamp_Fetch_InvalidPriceFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"last":"not-a-number"}`))
	}))
	defer srv.Close()

	b := NewBitstampWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	if _, err := b.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for invalid price format")
	}
}

func TestBitstamp_Fetch_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	b := NewBitstampWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	if _, err := b.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestBitstamp_Fetch_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	b := NewBitstampWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	if _, err := b.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}
