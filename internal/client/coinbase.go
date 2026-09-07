package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const coinbaseURL = "https://api.coinbase.com/v2/prices/BTC-USD/spot"

type Coinbase struct {
	httpClient *http.Client
	baseURL    string
}

func NewCoinbase(timeout time.Duration) *Coinbase {
	return &Coinbase{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: coinbaseURL,
	}
}

// NewCoinbaseWithURL is like NewCoinbase but overrides the endpoint, for tests.
func NewCoinbaseWithURL(timeout time.Duration, url string) *Coinbase {
	c := NewCoinbase(timeout)
	c.baseURL = url
	return c
}

func (c *Coinbase) Name() string {
	return "coinbase"
}

type coinbaseResponse struct {
	Data struct {
		Amount   string `json:"amount"`
		Base     string `json:"base"`
		Currency string `json:"currency"`
	} `json:"data"`
}

func (c *Coinbase) Fetch(ctx context.Context) (float64, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("coinbase: create request: %w", err)
	}

	req.Header.Set("User-Agent", "btc-aggregator/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("coinbase: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("coinbase: unexpected status %d", resp.StatusCode)
	}

	var parsed coinbaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, fmt.Errorf("coinbase: decode failed: %w", err)
	}

	if parsed.Data.Amount == "" {
		return 0, errors.New("coinbase: empty price")
	}

	price, err := strconv.ParseFloat(parsed.Data.Amount, 64)
	if err != nil {
		return 0, fmt.Errorf("coinbase: invalid price format: %w", err)
	}

	return price, nil
}
