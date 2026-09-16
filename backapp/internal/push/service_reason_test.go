package push

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadServiceReason(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"Apple authentication failure", `{"reason":"BadJwtToken"}`, "BadJwtToken"},
		{"Apple key mismatch", `{"reason":"VapidPkHashMismatch"}`, "VapidPkHashMismatch"},
		{"Apple expired subscription", `{"reason":"Unregistered"}`, "Unregistered"},
		{"FCM authentication failure", `{"error":{"status":"UNAUTHENTICATED","message":"sensitive-response"}}`, "UNAUTHENTICATED"},
		{"unrelated fields are omitted", `{"reason":"BadJwtToken","endpoint":"https://web.push.apple.com/sensitive-token","auth":"secret"}`, "BadJwtToken"},
		{"unknown code is omitted", `{"reason":"sensitive-token"}`, "unknown"},
		{"URL is omitted", `{"reason":"https://web.push.apple.com/sensitive-token"}`, "unknown"},
		{"line breaks are omitted", `{"reason":"BadJwtToken\nforged-entry"}`, "unknown"},
		{"provider text is omitted", "sensitive-response", "unknown"},
		{"malformed JSON", `{"reason":"BadJwtToken"`, "unknown"},
		{"empty body", "", "unknown"},
		{"null body", "null", "unknown"},
		{"wrong field type", `{"reason":123}`, "unknown"},
		{"oversized body", `{"reason":"BadJwtToken","message":"` + strings.Repeat("x", int(maxErrorBodyBytes)) + `"}`, "response_too_large"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := readServiceReason(strings.NewReader(test.body))
			if err != nil || got != test.want {
				t.Fatalf("readServiceReason() = %q, %v; want %q, nil", got, err, test.want)
			}
		})
	}
}

func TestReadServiceReasonBoundsResponseRead(t *testing.T) {
	reader := strings.NewReader(strings.Repeat("x", int(maxErrorBodyBytes)*2))
	initialLength := reader.Len()
	if reason, err := readServiceReason(reader); reason != "response_too_large" || err != nil {
		t.Fatalf("unexpected result: %q, %v", reason, err)
	}
	if read := initialLength - reader.Len(); int64(read) != maxErrorBodyBytes+1 {
		t.Fatalf("read %d bytes, want %d", read, maxErrorBodyBytes+1)
	}
}

func TestReadServiceReasonPreservesReadError(t *testing.T) {
	reason, err := readServiceReason(failingResponseReader{})
	if reason != "response_read_failed" || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("unexpected result: %q, %v", reason, err)
	}
}

type failingResponseReader struct{}

func (failingResponseReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
