package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type Coinbase struct {
	httpClient *http.Client
}

func NewCoinbase(timeout time.Duration) *Coinbase {
	return &Coinbase{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Coinbase) Name() string {
	return "coinbase"
}

func (c *Coinbase) Fetch(ctx context.Context) (float64, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.coinbase.com/v2/prices/BTC-USD/spot",
		nil,
	)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("non-200 response")
	}

	var parsed struct {
		Data struct {
			Amount string `json:"amount"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, err
	}

	return strconv.ParseFloat(parsed.Data.Amount, 64)
}