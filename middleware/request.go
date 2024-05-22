package middleware

import (
	"bytes"
	"io"
	"net/http"
)

func UpdateRequestHeader(header, value string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set(header, value)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func RemoveRequestHeader(header string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del(header)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AppendRequestBody(str string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				// TODO: handle read error
			}

			body = append(body, []byte(str)...)
			bodyBuffer := bytes.NewReader(body)
			_, err = bodyBuffer.Seek(0, 0)
			if err != nil {
				// TODO: handle seek error
			}

			// this is probably not the correct way to handle chunked requests
			r.ContentLength = bodyBuffer.Size()

			err = r.Body.Close()
			if err != nil {
				// TODO: handle Close error
			}
			r.Body = io.NopCloser(bodyBuffer)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func PrependRequest(str string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			body := []byte(str)
			bodyBuffer := bytes.NewBuffer(body)

			_, err := r.Body.Read(body)
			if err != nil {
				// TODO: handle read error
			}

			err = r.Body.Close()
			if err != nil {
				// TODO: handle Close error
			}
			r.Body = io.NopCloser(bodyBuffer)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
