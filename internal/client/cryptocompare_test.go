package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCryptoCompare_Fetch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"USD":65000.5}`))
	}))
	defer srv.Close()

	c := NewCryptoCompareWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	price, err := c.Fetch(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 65000.5 {
		t.Fatalf("expected 65000.5, got %f", price)
	}
	if c.Name() != "cryptocompare" {
		t.Fatalf("expected name cryptocompare, got %s", c.Name())
	}
}

func TestCryptoCompare_Fetch_ZeroPriceTreatedAsInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"USD":0}`))
	}))
	defer srv.Close()

	c := NewCryptoCompareWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	if _, err := c.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for zero price")
	}
}

func TestCryptoCompare_Fetch_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := NewCryptoCompareWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	if _, err := c.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestCryptoCompare_Fetch_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{not json`))
	}))
	defer srv.Close()

	c := NewCryptoCompareWithURL(&http.Client{Timeout: time.Second}, srv.URL)
	if _, err := c.Fetch(t.Context()); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}
