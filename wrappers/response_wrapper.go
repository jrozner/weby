package wrappers

import (
	"net/http"
)

type ResponseWrapper struct {
	status      int
	wroteHeader bool
	http.ResponseWriter
}

func (w *ResponseWrapper) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)

	// Informational responses do not commit the final response status.
	if status >= 100 && status <= 199 {
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

// Unwrap returns the underlying ResponseWriter so http.ResponseController can
// access optional capabilities such as flushing, hijacking, deadlines, and
// full-duplex operation.
func (w *ResponseWrapper) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		ResponseWriter: w,
	}
}
