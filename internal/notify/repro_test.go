package notify_test

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/notify"
)

func TestHTTPSenderHonorsCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		accepted <- c
		_, _ = io.Copy(io.Discard, c)
		_ = c.Close()
	}()

	sender := notify.NewHTTPSender("http://"+ln.Addr().String(), 15*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan notify.SendResult, 1)
	go func() {
		done <- sender.Send(ctx, model.AlarmEvent{Type: model.EventAlarmRaised, CellID: "c1"})
	}()
	select {
	case <-accepted:
	case <-time.After(2 * time.Second):
		t.Fatal("outbound request never started")
	}
	cancel()
	select {
	case res := <-done:
		if res.Success {
			t.Fatalf("cancelled send must not succeed: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handle ignored cancel and kept the outbound request alive")
	}
}

func TestNotifyServicePropagatesContext(t *testing.T) {
	var saw context.Context
	sender := &ctxCaptureSender{fn: func(ctx context.Context, _ model.AlarmEvent) notify.SendResult {
		saw = ctx
		return notify.SendResult{Success: true}
	}}
	svc := notify.NewService("http://example.com", sender, 5, time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = svc.Notify(ctx, []model.AlarmEvent{{Type: model.EventAlarmRaised, CellID: "c1"}})
	if saw == nil {
		t.Fatal("sender was not called")
	}
	if saw != ctx {
		t.Fatal("Notify must pass the caller context through to Sender (not context.Background)")
	}
	select {
	case <-saw.Done():
		t.Fatal("caller context should still be active")
	default:
	}
	cancel()
	select {
	case <-saw.Done():
	case <-time.After(time.Second):
		t.Fatal("propagated context must cancel with the caller")
	}
}

func TestSuccessClearsCircuit(t *testing.T) {
	cb := notify.NewCircuitBreaker(2, time.Minute)
	cb.RecordFailure()
	if cb.IsOpen() {
		t.Fatal("one failure must not open")
	}
	cb.RecordSuccess()
	cb.RecordFailure()
	if cb.IsOpen() {
		t.Fatal("success must clear failure count so a single later failure does not open the breaker")
	}

	failer := &seqSender{results: []notify.SendResult{
		{Success: false, Kind: notify.FailureHTTP},
		{Success: true},
		{Success: false, Kind: notify.FailureHTTP},
	}}
	svc := notify.NewService("http://example.com", failer, 2, time.Minute)
	ctx := context.Background()
	_ = svc.Notify(ctx, []model.AlarmEvent{{CellID: "c1"}})
	_ = svc.Notify(ctx, []model.AlarmEvent{{CellID: "c1"}})
	_ = svc.Notify(ctx, []model.AlarmEvent{{CellID: "c1"}})
	st := svc.CircuitState()
	if st.Open {
		t.Fatal("after failure+success+failure the circuit must stay closed because success clears the count")
	}
}

type seqSender struct {
	results []notify.SendResult
	i       int
}

func (s *seqSender) Send(ctx context.Context, _ model.AlarmEvent) notify.SendResult {
	if ctx.Err() != nil {
		return notify.SendResult{Kind: notify.FailureCancelled, Err: ctx.Err()}
	}
	if s.i >= len(s.results) {
		return notify.SendResult{Success: true}
	}
	r := s.results[s.i]
	s.i++
	return r
}

type ctxCaptureSender struct {
	fn func(context.Context, model.AlarmEvent) notify.SendResult
}

func (s *ctxCaptureSender) Send(ctx context.Context, ev model.AlarmEvent) notify.SendResult {
	return s.fn(ctx, ev)
}
