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
