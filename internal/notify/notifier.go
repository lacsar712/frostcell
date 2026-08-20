package notify

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

var (
	ErrCircuitOpen = errors.New("notify circuit open")
	ErrNoURL       = errors.New("notify URL not configured")
)

// FailureKind classifies outbound delivery errors.
type FailureKind int

const (
	FailureUnknown FailureKind = iota
	FailureTimeout
	FailureHTTP
	FailureCancelled
)

func (k FailureKind) String() string {
	switch k {
	case FailureTimeout:
		return "timeout"
	case FailureHTTP:
		return "http"
	case FailureCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// SendResult captures delivery outcome.
type SendResult struct {
	Kind      FailureKind
	Err       error
	Status    int
	Duration  time.Duration
	Success   bool
}

// Sender delivers alarm events to external channels.
type Sender interface {
	Send(ctx context.Context, event model.AlarmEvent) SendResult
}

// Service wraps sender with circuit breaker semantics.
type Service struct {
	mu      sync.Mutex
	sender  Sender
	circuit *CircuitBreaker
	url     string
}

// NewService creates a notification service.
func NewService(url string, sender Sender, threshold int, cooldown time.Duration) *Service {
	return &Service{
		sender:  sender,
		circuit: NewCircuitBreaker(threshold, cooldown),
		url:     url,
	}
}

// Notify dispatches events respecting context cancellation.
func (s *Service) Notify(ctx context.Context, events []model.AlarmEvent) []SendResult {
	if s.url == "" || s.sender == nil {
		return nil
	}

	results := make([]SendResult, 0, len(events))
	for _, ev := range events {
		if ctx.Err() != nil {
			results = append(results, SendResult{
				Kind:    FailureCancelled,
				Err:     ctx.Err(),
				Success: false,
			})
			return results
		}

		res := s.sendOne(ctx, ev)
		results = append(results, res)
	}
	return results
}

func (s *Service) sendOne(ctx context.Context, ev model.AlarmEvent) SendResult {
	s.mu.Lock()
	if s.circuit.IsOpen() {
		s.mu.Unlock()
		return SendResult{Kind: FailureUnknown, Err: ErrCircuitOpen, Success: false}
	}
	s.mu.Unlock()

	start := time.Now()
	result := s.sender.Send(ctx, ev)
	result.Duration = time.Since(start)

	s.mu.Lock()
	defer s.mu.Unlock()
	if !result.Success {
		s.circuit.RecordFailure()
	}
	return result
}

// CircuitState exposes breaker state for health endpoints.
func (s *Service) CircuitState() CircuitState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.circuit.State()
}
