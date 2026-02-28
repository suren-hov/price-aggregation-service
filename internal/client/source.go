package client

import "context"

type PriceSource interface {
	Name() string
	Fetch(ctx context.Context) (float64, error)
}