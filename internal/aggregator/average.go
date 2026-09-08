package aggregator

import (
	"errors"
	"math"
	"sort"
)

// PricePoint pairs a fetched price with the source it came from, so an
// Aggregator can report which sources actually contributed to its result
// (e.g. after excluding outliers) rather than just a bare number.
type PricePoint struct {
	Source string
	Price  float64
}

// Result is what an Aggregator produces: the aggregated value, plus which
// sources were used or excluded, so callers can report accurate counts and
// log/alert on exclusions instead of silently averaging over bad data.
type Result struct {
	Value           float64
	UsedSources     []string
	ExcludedSources []string
}

type Aggregator interface {
	Aggregate(prices []PricePoint) (Result, error)
}

// maxDeviationFromMedian is how far a single source's price may differ from
// the median of all sources before it's treated as an outlier (a parsing
// bug, a stale/wrong feed, etc.) and excluded from the average. Major
// exchanges' BTC/USD prices normally track within a fraction of a percent
// of each other, so 5% is generous enough to tolerate real market
// volatility while still catching a genuinely broken source.
const maxDeviationFromMedian = 0.05

type AverageAggregator struct{}

func NewAverage() *AverageAggregator {
	return &AverageAggregator{}
}

func (a *AverageAggregator) Aggregate(prices []PricePoint) (Result, error) {
	if len(prices) == 0 {
		return Result{}, errors.New("no valid prices")
	}

	used := prices
	var excluded []PricePoint

	// Outlier detection needs a majority to compare against; with fewer
	// than 3 points there's no way to tell which of two disagreeing prices
	// is the bad one, so every point is used as-is.
	if len(prices) >= 3 {
		median := medianPrice(prices)
		used, excluded = partitionByDeviation(prices, median, maxDeviationFromMedian)

		// Safety net: filtering should never empty the set given the median
		// is computed from these same points, but never discard all data
		// over a filtering edge case.
		if len(used) == 0 {
			used, excluded = prices, nil
		}
	}

	var sum float64
	for _, p := range used {
		sum += p.Price
	}
	avg := sum / float64(len(used))
	avg = math.Round(avg*100) / 100

	return Result{
		Value:           avg,
		UsedSources:     sourceNames(used),
		ExcludedSources: sourceNames(excluded),
	}, nil
}

func medianPrice(prices []PricePoint) float64 {
	values := make([]float64, len(prices))
	for i, p := range prices {
		values[i] = p.Price
	}
	sort.Float64s(values)

	n := len(values)
	if n%2 == 1 {
		return values[n/2]
	}
	return (values[n/2-1] + values[n/2]) / 2
}

func partitionByDeviation(prices []PricePoint, median, maxDeviation float64) (kept, excluded []PricePoint) {
	for _, p := range prices {
		if median == 0 || math.Abs(p.Price-median)/median <= maxDeviation {
			kept = append(kept, p)
		} else {
			excluded = append(excluded, p)
		}
	}
	return kept, excluded
}

func sourceNames(prices []PricePoint) []string {
	if len(prices) == 0 {
		return nil
	}
	names := make([]string, len(prices))
	for i, p := range prices {
		names[i] = p.Source
	}
	return names
}
