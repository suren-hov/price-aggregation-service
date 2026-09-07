package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

const bitstampURL = "https://www.bitstamp.net/api/v2/ticker/btcusd/"

type Bitstamp struct {
	httpClient *http.Client
	baseURL    string
}

func NewBitstamp(httpClient *http.Client) *Bitstamp {
	return &Bitstamp{
		httpClient: httpClient,
		baseURL:    bitstampURL,
	}
}

// NewBitstampWithURL is like NewBitstamp but overrides the endpoint, for tests.
func NewBitstampWithURL(httpClient *http.Client, url string) *Bitstamp {
	b := NewBitstamp(httpClient)
	b.baseURL = url
	return b
}

func (b *Bitstamp) Name() string {
	return "bitstamp"
}

type bitstampResponse struct {
	Last string `json:"last"`
}

func (b *Bitstamp) Fetch(ctx context.Context) (float64, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		b.baseURL,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("bitstamp: create request: %w", err)
	}

	req.Header.Set("User-Agent", "btc-aggregator/1.0")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("bitstamp: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bitstamp: unexpected status %d", resp.StatusCode)
	}

	var parsed bitstampResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, fmt.Errorf("bitstamp: decode failed: %w", err)
	}

	if parsed.Last == "" {
		return 0, errors.New("bitstamp: empty price")
	}

	price, err := strconv.ParseFloat(parsed.Last, 64)
	if err != nil {
		return 0, fmt.Errorf("bitstamp: invalid price format: %w", err)
	}

	return price, nil
}
