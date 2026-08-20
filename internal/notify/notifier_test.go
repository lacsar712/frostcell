package notify_test

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/notify"
)

type stubSender struct {
_calls int
}

func (s *stubSender) Send(ctx context.Context, _ model.AlarmEvent) notify.SendResult {
	if ctx.Err() != nil {
		return notify.SendResult{Kind: notify.FailureCancelled, Err: ctx.Err()}
	}
	s._calls++
	return notify.SendResult{Success: true}
}

func TestCircuitBreaker(t *testing.T) {
	cb := notify.NewCircuitBreaker(2, time.Second)
	cb.RecordFailure()
	if cb.IsOpen() {
		t.Fatal("should stay closed after one failure")
	}
	cb.RecordFailure()
	if !cb.IsOpen() {
		t.Fatal("should open after threshold")
	}
	cb.RecordSuccess()
	if cb.IsOpen() {
		t.Fatal("success should close breaker")
	}
}

func TestNotifyRespectsContext(t *testing.T) {
	sender := &stubSender{}
	svc := notify.NewService("http://example.com", sender, 5, time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := svc.Notify(ctx, []model.AlarmEvent{{Type: model.EventAlarmRaised, CellID: "c1"}})
	if len(results) != 1 || results[0].Kind != notify.FailureCancelled {
		t.Fatalf("results %+v", results)
	}
}
