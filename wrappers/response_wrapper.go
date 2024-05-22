package wrappers

import (
	"bytes"
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

func (w *ResponseWrapper) Status() int {
	return w.status
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		status:         0,
		ResponseWriter: w,
	}
}

type BufferedResponseWrapper struct {
	status int
	body   *bytes.Buffer
	http.ResponseWriter
}

func (w *BufferedResponseWrapper) WriteHeader(status int) {
	w.status = status
}

func (w *BufferedResponseWrapper) Write(body []byte) (int, error) {
	return w.body.Write(body)
}

func (w *BufferedResponseWrapper) Close() error {
	w.ResponseWriter.WriteHeader(w.status)
	_, err := w.ResponseWriter.Write(w.body.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (w *BufferedResponseWrapper) Status() int {
	return w.status
}

func NewBufferedResponseWrapper(w http.ResponseWriter) *BufferedResponseWrapper {
	return &BufferedResponseWrapper{
		status:         0,
		body:           bytes.NewBuffer(make([]byte, 0, 4096)),
		ResponseWriter: w,
	}
}
