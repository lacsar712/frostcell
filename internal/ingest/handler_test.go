package ingest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/ingest"
	"github.com/lacsar712/frostcell/internal/model"
)

type stubProc struct {
	last model.ProbeSample
}

func (s *stubProc) ProcessSample(sample model.ProbeSample, _ time.Time) (model.ProcessingResult, error) {
	s.last = sample
	return model.ProcessingResult{CellID: sample.CellID, Alarm: model.AlarmRecord{CellID: sample.CellID, State: model.StateNormal}}, nil
}

func TestHMACVerify(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"cellID":"c1"}`)
	sig, err := ingest.Sign(secret, body)
	if err != nil {
		t.Fatal(err)
	}
	if err := ingest.Verify(secret, body, sig); err != nil {
		t.Fatal(err)
	}
	if err := ingest.Verify(secret, body, "bad"); err == nil {
		t.Fatal("expected invalid sig error")
	}
}

func TestIngestHandler(t *testing.T) {
	proc := &stubProc{}
	h := ingest.NewHandler("secret", proc)
	h.Clock = func() time.Time { return time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC) }

	sample := model.ProbeSample{CellID: "c1", ProbeID: "p1", TempC: -17, TS: h.Clock()}
	body, _ := json.Marshal(sample)
	sig, _ := ingest.Sign("secret", body)

	req := httptest.NewRequest(http.MethodPost, "/v1/probes/sample", bytes.NewReader(body))
	req.Header.Set(ingest.HeaderName(), sig)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if proc.last.CellID != "c1" {
		t.Fatalf("proc last %+v", proc.last)
	}
}

func TestIngestBadSignature(t *testing.T) {
	h := ingest.NewHandler("secret", &stubProc{})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", w.Code)
	}
}
