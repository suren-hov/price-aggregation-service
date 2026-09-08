package poller

import (
	"testing"
	"time"
)

func TestCircuitBreaker_AllowsUntilThresholdReached(t *testing.T) {
	c := newCircuitBreaker()

	for i := 0; i < circuitFailureThreshold-1; i++ {
		if !c.Allow("x") {
			t.Fatalf("expected circuit to allow attempt %d", i)
		}
		if opened := c.RecordResult("x", false); opened {
			t.Fatalf("expected circuit to stay closed before reaching the threshold (attempt %d)", i)
		}
	}

	if !c.Allow("x") {
		t.Fatal("expected circuit to still allow the threshold-th attempt")
	}
}

func TestCircuitBreaker_OpensAtThresholdAndBlocksUntilCooldown(t *testing.T) {
	now := time.Now()
	c := newCircuitBreaker()
	c.now = func() time.Time { return now }

	var opened bool
	for i := 0; i < circuitFailureThreshold; i++ {
		opened = c.RecordResult("x", false)
	}
	if !opened {
		t.Fatal("expected circuit to open after reaching the failure threshold")
	}

	if c.Allow("x") {
		t.Fatal("expected circuit to block fetches immediately after opening")
	}

	now = now.Add(circuitCooldown - time.Second)
	if c.Allow("x") {
		t.Fatal("expected circuit to still be open just before cooldown elapses")
	}

	now = now.Add(2 * time.Second)
	if !c.Allow("x") {
		t.Fatal("expected circuit to allow a test fetch once cooldown has elapsed")
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	c := newCircuitBreaker()

	c.RecordResult("x", false)
	c.RecordResult("x", false)
	c.RecordResult("x", true) // reset

	var opened bool
	for i := 0; i < circuitFailureThreshold-1; i++ {
		opened = c.RecordResult("x", false)
	}
	if opened {
		t.Fatal("expected circuit to require a fresh run of failures after a success reset it")
	}
}

func TestCircuitBreaker_ReopensIfHalfOpenAttemptFails(t *testing.T) {
	now := time.Now()
	c := newCircuitBreaker()
	c.now = func() time.Time { return now }

	for i := 0; i < circuitFailureThreshold; i++ {
		c.RecordResult("x", false)
	}
	now = now.Add(circuitCooldown + time.Second)
	if !c.Allow("x") {
		t.Fatal("expected the half-open test attempt to be allowed")
	}

	if opened := c.RecordResult("x", false); !opened {
		t.Fatal("expected the circuit to reopen after the half-open attempt also failed")
	}
	if c.Allow("x") {
		t.Fatal("expected the circuit to block again immediately after reopening")
	}
}

func TestCircuitBreaker_TracksSourcesIndependently(t *testing.T) {
	c := newCircuitBreaker()

	for i := 0; i < circuitFailureThreshold; i++ {
		c.RecordResult("a", false)
	}

	if c.Allow("a") {
		t.Fatal("expected source 'a' to be blocked")
	}
	if !c.Allow("b") {
		t.Fatal("expected source 'b' to be unaffected by 'a' failing")
	}
}
