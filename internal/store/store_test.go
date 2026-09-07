package store

import (
	"sync"
	"testing"
	"time"

	"price-aggregation-service/internal/model"
)

func TestStore_GetReturnsZeroValueInitially(t *testing.T) {
	s := New()
	got := s.Get()
	if got != (model.Price{}) {
		t.Fatalf("expected zero value, got %+v", got)
	}
}

func TestStore_UpdateThenGet(t *testing.T) {
	s := New()
	want := model.Price{Value: 123, Currency: "USD", SourcesUsed: 3, LastUpdated: time.Now()}
	s.Update(want)

	if got := s.Get(); got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	s := New()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(v float64) {
			defer wg.Done()
			s.Update(model.Price{Value: v})
		}(float64(i))
		go func() {
			defer wg.Done()
			_ = s.Get()
		}()
	}

	wg.Wait()
}
