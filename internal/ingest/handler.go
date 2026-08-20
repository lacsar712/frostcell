package ingest

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// Processor handles validated samples.
type Processor interface {
	ProcessSample(sample model.ProbeSample, now time.Time) (model.ProcessingResult, error)
}

// Handler serves POST /v1/probes/sample with HMAC verification.
type Handler struct {
	Secret    string
	Processor Processor
	Clock     func() time.Time
}

// NewHandler creates an ingest handler.
func NewHandler(secret string, proc Processor) *Handler {
	return &Handler{
		Secret:    secret,
		Processor: proc,
		Clock:     time.Now,
	}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if h.Secret == "" {
		http.Error(w, ErrEmptySecret.Error(), http.StatusUnauthorized)
		return
	}

	if err := Verify(h.Secret, body, r.Header.Get(HeaderName())); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var sample model.ProbeSample
	if err := json.Unmarshal(body, &sample); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	now := h.Clock()
	if err := sample.Validate(now); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.Processor.ProcessSample(sample, now)
	if err != nil {
		if err == model.ErrCellNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseBody{
		OK:       true,
		CellID:   result.CellID,
		Alarm:    result.Alarm,
		Snapshot: result.Snapshot,
		Events:   result.Events,
	})
}

type responseBody struct {
	OK       bool                  `json:"ok"`
	CellID   string                `json:"cellID"`
	Alarm    model.AlarmRecord     `json:"alarm"`
	Snapshot model.WindowSnapshot  `json:"snapshot"`
	Events   []model.AlarmEvent    `json:"events"`
}
