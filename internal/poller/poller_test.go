package poller

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"price-aggregation-service/internal/aggregator"
	"price-aggregation-service/internal/client"
	"price-aggregation-service/internal/model"
	"price-aggregation-service/internal/store"
)

type fakeSource struct {
	name      string
	price     float64
	err       error
	callCount atomic.Int32
}

func (f *fakeSource) Name() string { return f.name }

func (f *fakeSource) Fetch(ctx context.Context) (float64, error) {
	f.callCount.Add(1)
	if f.err != nil {
		return 0, f.err
	}
	return f.price, nil
}

type erroringAggregator struct{}

func (erroringAggregator) Aggregate(prices []aggregator.PricePoint) (aggregator.Result, error) {
	return aggregator.Result{}, errors.New("aggregation always fails")
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestPollOnce_AllSourcesHealthy(t *testing.T) {
	st := store.New()
	p := New(
		[]client.PriceSource{
			&fakeSource{name: "a", price: 100},
			&fakeSource{name: "b", price: 200},
		},
		aggregator.NewAverage(),
		st,
		time.Second,
		testLogger(),
	)
	p.RetryConfig = RetryConfig{MaxRetries: 0, BaseDelay: time.Millisecond}

	p.pollOnce(context.Background())

	got := st.Get()
	if got.Stale {
		t.Fatalf("expected not stale, got stale")
	}
	if got.SourcesUsed != 2 {
		t.Fatalf("expected 2 sources used, got %d", got.SourcesUsed)
	}
	if got.Value != 150 {
		t.Fatalf("expected average 150, got %f", got.Value)
	}
}

func TestPollOnce_PartialFailure(t *testing.T) {
	st := store.New()
	p := New(
		[]client.PriceSource{
			&fakeSource{name: "a", price: 100},
			&fakeSource{name: "b", err: errors.New("boom")},
		},
		aggregator.NewAverage(),
		st,
		time.Second,
		testLogger(),
	)
	p.RetryConfig = RetryConfig{MaxRetries: 0, BaseDelay: time.Millisecond}

	p.pollOnce(context.Background())

	got := st.Get()
	if got.Stale {
		t.Fatalf("expected not stale when at least one source succeeds")
	}
	if got.SourcesUsed != 1 {
		t.Fatalf("expected 1 source used, got %d", got.SourcesUsed)
	}
	if got.Value != 100 {
		t.Fatalf("expected 100, got %f", got.Value)
	}
}

func TestPollOnce_AllSourcesFail_MarksStaleWithoutClobberingLastGoodPrice(t *testing.T) {
	st := store.New()
	st.Update(model.Price{Value: 500, SourcesUsed: 2, LastUpdated: time.Now(), Stale: false})

	p := New(
		[]client.PriceSource{
			&fakeSource{name: "a", err: errors.New("boom")},
			&fakeSource{name: "b", err: errors.New("boom")},
		},
		aggregator.NewAverage(),
		st,
		time.Second,
		testLogger(),
	)
	p.RetryConfig = RetryConfig{MaxRetries: 0, BaseDelay: time.Millisecond}

	p.pollOnce(context.Background())

	got := st.Get()
	if !got.Stale {
		t.Fatalf("expected stale when all sources fail")
	}
	if got.Value != 500 {
		t.Fatalf("expected last known price 500 preserved, got %f", got.Value)
	}
}

func TestPollOnce_OutlierSourceExcluded_SourcesUsedReflectsIt(t *testing.T) {
	st := store.New()
	p := New(
		[]client.PriceSource{
			&fakeSource{name: "a", price: 50000},
			&fakeSource{name: "b", price: 50100},
			&fakeSource{name: "c", price: 5000000}, // wildly off - a broken feed
		},
		aggregator.NewAverage(),
		st,
		time.Second,
		testLogger(),
	)
	p.RetryConfig = RetryConfig{MaxRetries: 0, BaseDelay: time.Millisecond}

	p.pollOnce(context.Background())

	got := st.Get()
	if got.SourcesUsed != 2 {
		t.Fatalf("expected 2 sources used after excluding the outlier, got %d", got.SourcesUsed)
	}
	if got.Value != 50050 {
		t.Fatalf("expected average of the 2 agreeing sources (50050), got %f", got.Value)
	}
}

func TestPollOnce_CircuitOpensAndSkipsFurtherFetches(t *testing.T) {
	st := store.New()
	failing := &fakeSource{name: "bad", err: errors.New("boom")}
	healthy := &fakeSource{name: "good", price: 100}

	p := New(
		[]client.PriceSource{failing, healthy},
		aggregator.NewAverage(),
		st,
		time.Second,
		testLogger(),
	)
	p.RetryConfig = RetryConfig{MaxRetries: 0, BaseDelay: time.Millisecond}

	for i := 0; i < circuitFailureThreshold; i++ {
		p.pollOnce(context.Background())
	}
	if got := failing.callCount.Load(); got != circuitFailureThreshold {
		t.Fatalf("expected %d fetch attempts before the circuit opens, got %d", circuitFailureThreshold, got)
	}

	// One more cycle: the circuit should now be open, so Fetch is not
	// called again even though pollOnce runs.
	p.pollOnce(context.Background())
	if got := failing.callCount.Load(); got != circuitFailureThreshold {
		t.Fatalf("expected no additional fetch attempts while the circuit is open, got %d calls", got)
	}

	// The healthy source is unaffected and the store still reflects it.
	got := st.Get()
	if got.Value != 100 {
		t.Fatalf("expected the healthy source's price 100, got %f", got.Value)
	}
}

func TestPollOnce_AggregatorError_DoesNotUpdateStore(t *testing.T) {
	st := store.New()
	st.Update(model.Price{Value: 42, LastUpdated: time.Now(), Stale: false})

	p := New(
		[]client.PriceSource{
			&fakeSource{name: "a", price: 100},
		},
		erroringAggregator{},
		st,
		time.Second,
		testLogger(),
	)
	p.RetryConfig = RetryConfig{MaxRetries: 0, BaseDelay: time.Millisecond}

	p.pollOnce(context.Background())

	got := st.Get()
	if got.Value != 42 {
		t.Fatalf("expected store to remain unchanged on aggregator error, got %f", got.Value)
	}
}
