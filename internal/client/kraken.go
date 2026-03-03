package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type Kraken struct {
	httpClient *http.Client
}

func NewKraken(httpClient *http.Client) *Kraken {
	return &Kraken{
		httpClient: httpClient,
	}
}

func (k *Kraken) Name() string {
	return "kraken"
}

type krakenTickerResponse struct {
	Error  []string                       `json:"error"`
	Result map[string]krakenTickerPayload `json:"result"`
}

type krakenTickerPayload struct {
	C []string `json:"c"` // last trade price
}

func (k *Kraken) Fetch(ctx context.Context) (float64, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.kraken.com/0/public/Ticker?pair=XBTUSD",
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("kraken: create request: %w", err)
	}

	// Some exchanges behave better with UA set
	req.Header.Set("User-Agent", "btc-aggregator/1.0")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("kraken: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("kraken: unexpected status %d", resp.StatusCode)
	}

	var parsed krakenTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, fmt.Errorf("kraken: decode failed: %w", err)
	}

	// Kraken-level error check
	if len(parsed.Error) > 0 {
		return 0, fmt.Errorf("kraken API error: %v", parsed.Error)
	}

	if len(parsed.Result) == 0 {
		return 0, errors.New("kraken: empty result")
	}

	for _, v := range parsed.Result {
		if len(v.C) == 0 {
			return 0, errors.New("kraken: missing price data")
		}

		price, err := strconv.ParseFloat(v.C[0], 64)
		if err != nil {
			return 0, fmt.Errorf("kraken: invalid price format: %w", err)
		}

		return price, nil
	}

	return 0, errors.New("kraken: unexpected response format")
}