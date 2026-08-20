package notify

import (
	"sync"
	"time"
)

// CircuitState represents breaker open/closed status.
type CircuitState struct {
	Open            bool
	Failures        int
	Threshold       int
	OpenedAt        time.Time
	CooldownExpires time.Time
}

// CircuitBreaker opens after consecutive failures and cools down before retry.
type CircuitBreaker struct {
	mu        sync.Mutex
	threshold int
	cooldown  time.Duration
	failures  int
	openUntil time.Time
}

// NewCircuitBreaker creates a breaker with failure threshold and cooldown.
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	if threshold < 1 {
		threshold = 1
	}
	return &CircuitBreaker{threshold: threshold, cooldown: cooldown}
}

// IsOpen reports whether requests should be short-circuited.
func (c *CircuitBreaker) IsOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.openUntil.IsZero() {
		return false
	}
	if time.Now().After(c.openUntil) {
		c.openUntil = time.Time{}
		c.failures = 0
		return false
	}
	return true
}

// RecordFailure increments failures and opens when threshold reached.
func (c *CircuitBreaker) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	if c.failures >= c.threshold {
		c.openUntil = time.Now().Add(c.cooldown)
	}
}

// RecordSuccess clears failure count and closes the breaker.
func (c *CircuitBreaker) RecordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures = 0
	c.openUntil = time.Time{}
}

// State returns a snapshot of breaker internals.
func (c *CircuitBreaker) State() CircuitState {
	c.mu.Lock()
	defer c.mu.Unlock()
	st := CircuitState{
		Failures:  c.failures,
		Threshold: c.threshold,
	}
	if !c.openUntil.IsZero() && time.Now().Before(c.openUntil) {
		st.Open = true
		st.OpenedAt = c.openUntil.Add(-c.cooldown)
		st.CooldownExpires = c.openUntil
	}
	return st
}
