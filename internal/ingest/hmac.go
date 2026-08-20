package ingest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var (
	ErrMissingSignature = errors.New("missing X-Frostcell-Signature header")
	ErrInvalidSignature = errors.New("invalid HMAC signature")
	ErrEmptySecret      = errors.New("HMAC secret is empty")
)

const signatureHeader = "X-Frostcell-Signature"

// Sign computes HMAC-SHA256 hex digest of body using secret.
func Sign(secret string, body []byte) (string, error) {
	if secret == "" {
		return "", ErrEmptySecret
	}
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write(body); err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// Verify checks the signature header against body and secret.
func Verify(secret string, body []byte, header string) error {
	if header == "" {
		return ErrMissingSignature
	}
	expected, err := Sign(secret, body)
	if err != nil {
		return err
	}
	provided := strings.TrimSpace(header)
	if !hmac.Equal([]byte(expected), []byte(provided)) {
		return ErrInvalidSignature
	}
	return nil
}

// HeaderName returns the HTTP header used for HMAC verification.
func HeaderName() string {
	return signatureHeader
}
