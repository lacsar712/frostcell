package ingest_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lacsar712/frostcell/internal/ingest"
)

func TestHMACEmptySecretNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("empty HMAC secret must not panic: %v", r)
		}
	}()

	_, err := ingest.Sign("", []byte(`{}`))
	if err != ingest.ErrEmptySecret {
		t.Fatalf("Sign empty secret: got %v want ErrEmptySecret", err)
	}

	h := ingest.NewHandler("", &stubProc{})
	req := httptest.NewRequest(http.MethodPost, "/v1/probes/sample", bytes.NewReader([]byte(`{}`)))
	req.Header.Set(ingest.HeaderName(), "deadbeef")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("empty")) {
		t.Fatalf("response should mention empty secret, got %s", w.Body.String())
	}
}
