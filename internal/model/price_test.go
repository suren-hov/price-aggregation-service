package model

import (
	"testing"
	"time"
)

func TestIsStale(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		price Price
		want  bool
	}{
		{
			name:  "never updated",
			price: Price{},
			want:  true,
		},
		{
			name:  "explicitly marked stale",
			price: Price{Stale: true, LastUpdated: now},
			want:  true,
		},
		{
			name:  "fresh",
			price: Price{Stale: false, LastUpdated: now.Add(-1 * time.Second)},
			want:  false,
		},
		{
			name:  "older than max age",
			price: Price{Stale: false, LastUpdated: now.Add(-time.Hour)},
			want:  true,
		},
		{
			name:  "exactly at max age boundary is not stale",
			price: Price{Stale: false, LastUpdated: now.Add(-30 * time.Second)},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.price.IsStale(30*time.Second, now)
			if got != tt.want {
				t.Errorf("IsStale() = %v, want %v", got, tt.want)
			}
		})
	}
}
