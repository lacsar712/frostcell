package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// HTTPSender posts JSON payloads to a webhook URL.
type HTTPSender struct {
	URL        string
	Client     *http.Client
	Timeout    time.Duration
}

// NewHTTPSender creates an HTTP notification sender.
func NewHTTPSender(url string, timeout time.Duration) *HTTPSender {
	return &HTTPSender{
		URL:     url,
		Timeout: timeout,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Send delivers one alarm event via HTTP POST.
func (h *HTTPSender) Send(ctx context.Context, event model.AlarmEvent) SendResult {
	if h.URL == "" {
		return SendResult{Kind: FailureUnknown, Err: ErrNoURL, Success: false}
	}

	payload := model.NotifyPayload{
		Event:     event.Type,
		CellID:    event.CellID,
		State:     event.State,
		Timestamp: event.Timestamp,
		MeanTempC: event.MeanTempC,
		MaxTempC:  event.MaxTempC,
		OverRatio: event.OverRatio,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return SendResult{Kind: FailureUnknown, Err: err, Success: false}
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, h.URL, bytes.NewReader(body))
	if err != nil {
		return SendResult{Kind: FailureUnknown, Err: err, Success: false}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "frostcell/1.0")

	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: h.Timeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		kind := FailureUnknown
		if ctx.Err() != nil {
			kind = FailureCancelled
		} else if isTimeout(err) {
			kind = FailureTimeout
		}
		return SendResult{Kind: kind, Err: err, Success: false}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return SendResult{Status: resp.StatusCode, Success: true}
	}
	return SendResult{
		Kind:    FailureHTTP,
		Status:  resp.StatusCode,
		Success: false,
		Err:     errHTTPStatus(resp.StatusCode),
	}
}

func isTimeout(err error) bool {
	type timeout interface{ Timeout() bool }
	if te, ok := err.(timeout); ok {
		return te.Timeout()
	}
	return false
}

type httpStatusError int

func (e httpStatusError) Error() string {
	return "notify http status " + itoa(int(e))
}

func errHTTPStatus(code int) error {
	return httpStatusError(code)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	d := []byte{}
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}
