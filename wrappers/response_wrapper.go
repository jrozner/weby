package wrappers

import (
	"net/http"
)

// ResponseWrapper records the final HTTP response status while forwarding
// headers and body writes to an underlying http.ResponseWriter.
//
// Handlers must continue to call Header, WriteHeader, and Write on the wrapper
// so it can observe the response. To use advanced response operations, pass the
// wrapper itself to http.NewResponseController. Do not call Unwrap and then
// write through the returned writer, because those writes bypass status
// tracking.
type ResponseWrapper struct {
	status      int
	wroteHeader bool
	http.ResponseWriter
}

func (w *ResponseWrapper) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)

	// Informational responses do not commit the final response status. A 101
	// response is final because the connection switches protocols immediately.
	if status >= 100 && status <= 199 && status != http.StatusSwitchingProtocols {
		return
	}
	if !w.wroteHeader {
		w.status = status
		w.wroteHeader = true
	}
}

// Write captures the implicit 200 OK that net/http writes when a handler
// calls Write without first calling WriteHeader. Without this, Status()
// returns 0 for handlers that don't set a status explicitly, which makes
// middleware that reports on the status (e.g. request loggers) misreport
// every successful response as 0.
func (w *ResponseWrapper) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.status = http.StatusOK
		w.wroteHeader = true
	}
	return w.ResponseWriter.Write(b)
}

func (w *ResponseWrapper) Status() int {
	if !w.wroteHeader {
		return http.StatusOK
	}
	return w.status
}

// Committed reports whether the wrapper has observed a final response status.
// Writes made directly through the unwrapped ResponseWriter bypass tracking and
// cannot be reported here.
func (w *ResponseWrapper) Committed() bool {
	return w.wroteHeader
}

// FlushError flushes the underlying ResponseWriter through
// http.ResponseController and records the implicit 200 OK committed by a
// successful flush. ResponseController prefers this method over Unwrap, which
// keeps status tracking accurate if WriteHeader is called after a flush.
func (w *ResponseWrapper) FlushError() error {
	if err := http.NewResponseController(w.ResponseWriter).Flush(); err != nil {
		return err
	}
	if !w.wroteHeader {
		w.status = http.StatusOK
		w.wroteHeader = true
	}
	return nil
}

// Unwrap returns the underlying ResponseWriter so http.ResponseController can
// access optional capabilities such as flushing, hijacking, deadlines, and
// full-duplex operation. Callers should pass the wrapper directly to
// http.NewResponseController rather than calling Unwrap themselves. Normal
// response writes made through the unwrapped writer cannot be observed by the
// wrapper and will make Status inaccurate.
func (w *ResponseWrapper) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		ResponseWriter: w,
	}
}
