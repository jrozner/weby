package wrappers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusCapturedFromWriteHeader(t *testing.T) {
	w := NewResponseWrapper(httptest.NewRecorder())
	w.WriteHeader(http.StatusTeapot)
	if got := w.Status(); got != http.StatusTeapot {
		t.Errorf("Status() = %d, want %d", got, http.StatusTeapot)
	}
}

func TestStatusCapturedFromImplicit200(t *testing.T) {
	w := NewResponseWrapper(httptest.NewRecorder())
	if _, err := w.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	if got := w.Status(); got != http.StatusOK {
		t.Errorf("Status() = %d, want %d (implicit 200 from Write)", got, http.StatusOK)
	}
}

func TestExplicitHeaderWinsOverWrite(t *testing.T) {
	w := NewResponseWrapper(httptest.NewRecorder())
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	if got := w.Status(); got != http.StatusCreated {
		t.Errorf("Status() = %d, want %d (explicit status must not be overwritten by Write)", got, http.StatusCreated)
	}
}

func TestFirstFinalHeaderWins(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := NewResponseWrapper(recorder)
	w.WriteHeader(http.StatusCreated)
	w.WriteHeader(http.StatusTeapot)

	if got := w.Status(); got != http.StatusCreated {
		t.Errorf("Status() = %d, want %d", got, http.StatusCreated)
	}
	if got := recorder.Code; got != http.StatusCreated {
		t.Errorf("committed status = %d, want %d", got, http.StatusCreated)
	}
}

func TestUntouchedResponseDefaultsToOK(t *testing.T) {
	w := NewResponseWrapper(httptest.NewRecorder())

	if got := w.Status(); got != http.StatusOK {
		t.Errorf("Status() = %d, want %d", got, http.StatusOK)
	}
}

func TestInformationalHeaderDoesNotSetFinalStatus(t *testing.T) {
	w := NewResponseWrapper(&recordingResponseWriter{})
	w.WriteHeader(http.StatusEarlyHints)
	w.WriteHeader(http.StatusNoContent)

	if got := w.Status(); got != http.StatusNoContent {
		t.Errorf("Status() = %d, want %d", got, http.StatusNoContent)
	}
}

func TestSwitchingProtocolsSetsFinalStatus(t *testing.T) {
	w := NewResponseWrapper(&recordingResponseWriter{})
	w.WriteHeader(http.StatusSwitchingProtocols)
	w.WriteHeader(http.StatusInternalServerError)

	if got := w.Status(); got != http.StatusSwitchingProtocols {
		t.Errorf("Status() = %d, want %d", got, http.StatusSwitchingProtocols)
	}
}

func TestResponseControllerCanFlushUnderlyingWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := NewResponseWrapper(recorder)

	if err := http.NewResponseController(w).Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if !recorder.Flushed {
		t.Fatal("underlying ResponseWriter was not flushed")
	}
}

func TestResponseControllerFlushCommitsTrackedOK(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := NewResponseWrapper(recorder)

	if err := http.NewResponseController(w).Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	w.WriteHeader(http.StatusInternalServerError)

	if got := w.Status(); got != http.StatusOK {
		t.Errorf("Status() after flush and WriteHeader = %d, want %d", got, http.StatusOK)
	}
	if got := recorder.Code; got != http.StatusOK {
		t.Errorf("committed status = %d, want %d", got, http.StatusOK)
	}
}

func TestUnwrapReturnsUnderlyingWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := NewResponseWrapper(recorder)

	if got := w.Unwrap(); got != recorder {
		t.Fatalf("Unwrap() = %T %p, want %T %p", got, got, recorder, recorder)
	}
}

type recordingResponseWriter struct {
	header http.Header
	codes  []int
}

func (w *recordingResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *recordingResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w *recordingResponseWriter) WriteHeader(status int) {
	w.codes = append(w.codes, status)
}
