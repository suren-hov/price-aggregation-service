package model

import "time"

type Price struct {
	Value       float64   `json:"price"`
	Currency    string    `json:"currency"`
	SourcesUsed int       `json:"sources_used"`
	LastUpdated time.Time `json:"last_updated"`
	Stale       bool      `json:"stale"`
}

// IsStale reports whether the price should be considered stale: either the
// poller explicitly marked it so (e.g. every source failed on the last
// cycle), it has never been populated, or it hasn't been refreshed within
// maxAge (e.g. the poller hung or stopped without updating Stale).
func (p Price) IsStale(maxAge time.Duration, now time.Time) bool {
	if p.Stale || p.LastUpdated.IsZero() {
		return true
	}
	return now.Sub(p.LastUpdated) > maxAge
}
