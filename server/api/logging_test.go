package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLoggerRecordsRequestResponseAndRequestID(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := requestLogger(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/example?kind=test", nil)
	request.Header.Set(requestIDHeader, "browser-request-1")
	recording := httptest.NewRecorder()
	handler.ServeHTTP(recording, request)

	if recording.Code != http.StatusOK || recording.Body.String() != "hello" {
		t.Fatalf("response = status %d body %q", recording.Code, recording.Body.String())
	}
	if recording.Header().Get(requestIDHeader) != "browser-request-1" {
		t.Fatalf("response request ID = %q", recording.Header().Get(requestIDHeader))
	}

	messages := logMessages(t, output.String())
	if got := messages[0]["msg"]; got != "http request started" {
		t.Fatalf("first log message = %v", got)
	}
	if got := messages[1]["msg"]; got != "http response started" {
		t.Fatalf("second log message = %v", got)
	}
	completed := messages[2]
	if completed["msg"] != "http request completed" || completed["request_id"] != "browser-request-1" || completed["path"] != "/api/example" {
		t.Fatalf("completion log = %#v", completed)
	}
	if completed["status"] != float64(http.StatusOK) || completed["bytes"] != float64(len("hello")) {
		t.Fatalf("completion result = %#v", completed)
	}
	if duration, ok := completed["duration_ms"].(float64); !ok || duration < 0 {
		t.Fatalf("completion duration = %#v", completed["duration_ms"])
	}
}

func TestRequestLoggerGeneratesRequestID(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := requestLogger(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recording := httptest.NewRecorder()
	handler.ServeHTTP(recording, httptest.NewRequest(http.MethodDelete, "/api/example", nil))

	requestID := recording.Header().Get(requestIDHeader)
	if len(requestID) != 32 || strings.Trim(requestID, "0123456789abcdef") != "" {
		t.Fatalf("generated request ID = %q", requestID)
	}
}

func logMessages(t *testing.T, output string) []map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	messages := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var message map[string]any
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			t.Fatalf("decode log %q: %v", line, err)
		}
		messages = append(messages, message)
	}
	return messages
}
