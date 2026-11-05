package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jrozner/weby/wrappers"
)

func TestLoggerLogsAndPropagatesPanic(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	panicValue := "handler panic"
	handler := Logger(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(panicValue)
	}))

	w := wrappers.NewResponseWrapper(httptest.NewRecorder())
	r := httptest.NewRequest(http.MethodGet, "/panic", nil)
	got := recoveredPanic(func() { handler.ServeHTTP(w, r) })

	if got != panicValue {
		t.Fatalf("recovered panic = %v, want %q", got, panicValue)
	}
	entry := logs.String()
	if !strings.Contains(entry, `level=ERROR msg="http request"`) {
		t.Errorf("panic log missing error level and message: %s", entry)
	}
	if !strings.Contains(entry, `method=GET path=/panic`) {
		t.Errorf("panic log missing request fields: %s", entry)
	}
	if !strings.Contains(entry, `outcome=panicked`) {
		t.Errorf("panic log missing outcome: %s", entry)
	}
	if strings.Contains(entry, `status=200`) {
		t.Errorf("uncommitted panic log contains a false 200 status: %s", entry)
	}
}

func TestLoggerIncludesCommittedStatusWhenHandlerPanics(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		panic("handler panic")
	}))

	w := wrappers.NewResponseWrapper(httptest.NewRecorder())
	r := httptest.NewRequest(http.MethodPost, "/jobs", nil)
	_ = recoveredPanic(func() { handler.ServeHTTP(w, r) })

	entry := logs.String()
	if !strings.Contains(entry, `outcome=panicked status=202`) {
		t.Errorf("panic log missing committed status: %s", entry)
	}
}

func TestLoggerLogsCompletedRequest(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	w := wrappers.NewResponseWrapper(httptest.NewRecorder())
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(w, r)

	entry := logs.String()
	if !strings.Contains(entry, `level=INFO msg="http request"`) {
		t.Errorf("request log missing info level and message: %s", entry)
	}
	if !strings.Contains(entry, `method=GET path=/health`) {
		t.Errorf("request log missing request fields: %s", entry)
	}
	if !strings.Contains(entry, `status=204`) {
		t.Errorf("request log missing status: %s", entry)
	}
	if strings.Contains(entry, `outcome=panicked`) {
		t.Errorf("completed request logged as panicked: %s", entry)
	}
}

func recoveredPanic(fn func()) (recovered any) {
	defer func() {
		recovered = recover()
	}()
	fn()
	return nil
}
