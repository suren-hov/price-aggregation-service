package aggregator

import "errors"

type Aggregator interface {
	Aggregate(prices []float64) (float64, error)
}

type AverageAggregator struct{}

func NewAverage() *AverageAggregator {
	return &AverageAggregator{}
}

func (a *AverageAggregator) Aggregate(prices []float64) (float64, error) {
	if len(prices) == 0 {
		return 0, errors.New("no valid prices")
	}

	var sum float64
	for _, p := range prices {
		sum += p
	}

	return sum / float64(len(prices)), nil
}