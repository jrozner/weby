package wrappers

import (
	"net/http"
)

type ResponseWrapper struct {
	status int
	http.ResponseWriter
}

func (w *ResponseWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(w.status)
}

// Write captures the implicit 200 OK that net/http writes when a handler
// calls Write without first calling WriteHeader. Without this, Status()
// returns 0 for handlers that don't set a status explicitly, which makes
// middleware that reports on the status (e.g. request loggers) misreport
// every successful response as 0.
func (w *ResponseWrapper) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func (w *ResponseWrapper) Status() int {
	return w.status
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		status:         0,
		ResponseWriter: w,
	}
}
