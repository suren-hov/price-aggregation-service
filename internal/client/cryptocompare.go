package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const cryptoCompareURL = "https://min-api.cryptocompare.com/data/price?fsym=BTC&tsyms=USD"

type CryptoCompare struct {
	httpClient *http.Client
	baseURL    string
}

func NewCryptoCompare(httpClient *http.Client) *CryptoCompare {
	return &CryptoCompare{
		httpClient: httpClient,
		baseURL:    cryptoCompareURL,
	}
}

// NewCryptoCompareWithURL is like NewCryptoCompare but overrides the endpoint, for tests.
func NewCryptoCompareWithURL(httpClient *http.Client, url string) *CryptoCompare {
	c := NewCryptoCompare(httpClient)
	c.baseURL = url
	return c
}

func (c *CryptoCompare) Name() string {
	return "cryptocompare"
}

type cryptoCompareResponse struct {
	USD float64 `json:"USD"`
}

func (c *CryptoCompare) Fetch(ctx context.Context) (float64, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("cryptocompare: create request: %w", err)
	}

	req.Header.Set("User-Agent", "btc-aggregator/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cryptocompare: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("cryptocompare: unexpected status %d", resp.StatusCode)
	}

	var parsed cryptoCompareResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, fmt.Errorf("cryptocompare: decode failed: %w", err)
	}

	if parsed.USD == 0 {
		return 0, errors.New("cryptocompare: invalid price")
	}

	return parsed.USD, nil
}
