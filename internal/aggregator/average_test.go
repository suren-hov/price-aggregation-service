package aggregator

import "testing"

func TestAverage(t *testing.T) {
	agg := NewAverage()

	result, err := agg.Aggregate([]float64{10, 20, 30})
	if err != nil {
		t.Fatal(err)
	}

	if result != 20 {
		t.Fatalf("expected 20 got %f", result)
	}
}