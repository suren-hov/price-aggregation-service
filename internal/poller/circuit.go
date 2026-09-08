package poller

import (
	"sync"
	"time"
)

// circuitFailureThreshold is how many consecutive cycle failures for a
// source trip its circuit open.
const circuitFailureThreshold = 3

// circuitCooldown is how long a tripped circuit stays open before a single
// fetch is allowed through again to test recovery (half-open).
const circuitCooldown = 30 * time.Second

// circuitBreaker tracks per-source failures across poll cycles. A source
// that fails circuitFailureThreshold cycles in a row is "opened": further
// cycles skip fetching it entirely (no network call, no retries) until
// circuitCooldown elapses, instead of paying full retry latency against an
// exchange that's already known to be down or rate-limiting.
type circuitBreaker struct {
	mu    sync.Mutex
	now   func() time.Time
	state map[string]*circuitState
}

type circuitState struct {
	consecutiveFailures int
	openUntil           time.Time
}

func newCircuitBreaker() *circuitBreaker {
	return &circuitBreaker{
		now:   time.Now,
		state: make(map[string]*circuitState),
	}
}

// Allow reports whether source may be fetched this cycle.
func (c *circuitBreaker) Allow(source string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.state[source]
	if !ok || s.openUntil.IsZero() {
		return true
	}
	return !c.now().Before(s.openUntil)
}

// RecordResult updates source's circuit based on whether its fetch attempt
// succeeded. It must only be called for sources that were actually
// attempted (i.e. Allow returned true), not for skipped ones. Returns
// whether the circuit is open after this update.
func (c *circuitBreaker) RecordResult(source string, success bool) (open bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.state[source]
	if !ok {
		s = &circuitState{}
		c.state[source] = s
	}

	if success {
		s.consecutiveFailures = 0
		s.openUntil = time.Time{}
		return false
	}

	s.consecutiveFailures++
	if s.consecutiveFailures >= circuitFailureThreshold {
		s.openUntil = c.now().Add(circuitCooldown)
		return true
	}
	return false
}
