package aggregator

import (
	"reflect"
	"sort"
	"testing"
)

func points(sourcesToPrices map[string]float64) []PricePoint {
	names := make([]string, 0, len(sourcesToPrices))
	for name := range sourcesToPrices {
		names = append(names, name)
	}
	sort.Strings(names)

	pts := make([]PricePoint, len(names))
	for i, name := range names {
		pts[i] = PricePoint{Source: name, Price: sourcesToPrices[name]}
	}
	return pts
}

func sortedCopy(s []string) []string {
	c := append([]string(nil), s...)
	sort.Strings(c)
	return c
}

func TestAverage(t *testing.T) {
	agg := NewAverage()

	// Clustered, realistic values - 10/20/30 (the original test data) are
	// >5% apart and would now be treated as mutual outliers.
	result, err := agg.Aggregate(points(map[string]float64{"a": 19.5, "b": 20, "c": 20.5}))
	if err != nil {
		t.Fatal(err)
	}

	if result.Value != 20 {
		t.Fatalf("expected 20 got %f", result.Value)
	}
	if len(result.ExcludedSources) != 0 {
		t.Fatalf("expected no exclusions among agreeing prices, got %v", result.ExcludedSources)
	}
}

func TestAggregate_NoPrices(t *testing.T) {
	agg := NewAverage()

	if _, err := agg.Aggregate(nil); err == nil {
		t.Fatal("expected an error for no prices")
	}
}

func TestAggregate_SinglePrice_NoOutlierFiltering(t *testing.T) {
	agg := NewAverage()

	// With only one source, there's nothing to compare it against.
	result, err := agg.Aggregate(points(map[string]float64{"a": 999999}))
	if err != nil {
		t.Fatal(err)
	}
	if result.Value != 999999 {
		t.Fatalf("expected the lone price to be used as-is, got %f", result.Value)
	}
}

func TestAggregate_TwoPrices_NoOutlierFiltering(t *testing.T) {
	agg := NewAverage()

	// With only two sources there's no majority to trust, so both are used
	// even though they disagree wildly.
	result, err := agg.Aggregate(points(map[string]float64{"a": 100, "b": 1000}))
	if err != nil {
		t.Fatal(err)
	}
	if result.Value != 550 {
		t.Fatalf("expected average of both prices (550), got %f", result.Value)
	}
	if len(result.ExcludedSources) != 0 {
		t.Fatalf("expected no exclusions with only two sources, got %v", result.ExcludedSources)
	}
}

func TestAggregate_ExcludesOutlierAmongThreeOrMore(t *testing.T) {
	agg := NewAverage()

	// a and b agree closely; c is wildly off (a broken feed).
	result, err := agg.Aggregate(points(map[string]float64{
		"a": 50000,
		"b": 50100,
		"c": 5000000,
	}))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(sortedCopy(result.ExcludedSources), []string{"c"}) {
		t.Fatalf("expected only 'c' excluded as an outlier, got %v", result.ExcludedSources)
	}
	if !reflect.DeepEqual(sortedCopy(result.UsedSources), []string{"a", "b"}) {
		t.Fatalf("expected 'a' and 'b' used, got %v", result.UsedSources)
	}
	if result.Value != 50050 {
		t.Fatalf("expected average of a and b (50050), got %f", result.Value)
	}
}

func TestAggregate_NoExclusionsWithinTolerance(t *testing.T) {
	agg := NewAverage()

	// All three within a fraction of a percent of each other - normal
	// market variance across exchanges, not an outlier.
	result, err := agg.Aggregate(points(map[string]float64{
		"a": 50000,
		"b": 50050,
		"c": 49980,
	}))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.ExcludedSources) != 0 {
		t.Fatalf("expected no exclusions for prices within tolerance, got %v", result.ExcludedSources)
	}
	if len(result.UsedSources) != 3 {
		t.Fatalf("expected all 3 sources used, got %v", result.UsedSources)
	}
}
