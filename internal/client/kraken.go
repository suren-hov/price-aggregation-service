package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type Kraken struct {
	httpClient *http.Client
}

func NewKraken(timeout time.Duration) *Kraken {
	return &Kraken{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (k *Kraken) Name() string {
	return "kraken"
}

func (k *Kraken) Fetch(ctx context.Context) (float64, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.kraken.com/0/public/Ticker?pair=XBTUSD",
		nil,
	)
	if err != nil {
		return 0, err
	}

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("non-200 response")
	}

	var parsed struct {
		Result map[string]struct {
			C []string `json:"c"` // last trade price
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, err
	}

	for _, v := range parsed.Result {
		if len(v.C) == 0 {
			return 0, errors.New("no price data")
		}
		return strconv.ParseFloat(v.C[0], 64)
	}

	return 0, errors.New("unexpected kraken response")
}