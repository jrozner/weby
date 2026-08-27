package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/sessions"
)

/*
This implementation relies on a few security assumptions listed below. It passes all tokens in the request/response
headers and leaves it up to the client to sort out how to handle that. No changes are required in the handlers on the
go side to implement this, everything is done in the middleware. This implementation doesn't rotate the actual token
for the duration of the session but does change the token value that is sent to the client by creating a one time pad
and XORing the static value in the session, then sending that value. The OTP is pre-pended to the XORed value so that
the server can XOR it back for comparison.

token = <one time pad> <one time pad ^ token>

This XOR operation is not intended as a form of encryption but a protection against the TLS BREACH attack. It is simply
here to create variation in an otherwise static value without changing the underlying value. This avoids creating a
situation where the client and server have desynchronized and requests fail. This also means in the event of token
capture it is possible to recover the original token or perform replay attacks. These replay attacks are limited to the
duration of the session, however. As an alternative we can implement a token that is not bound to the session such as
https://medium.com/@jrozner/wiping-out-csrf-ded97ae7e83f.

The design of this is similar to the Rails and Django implementations. More information about the Rails implementation
can be found https://medium.com/rubyinside/a-deep-dive-into-csrf-protection-in-rails-19fa0a42c0ef.

security assumptions:
- all communication happens over TLS
- sessions are opaque to the user (encrypted or stored server side), if not the token will be leaked
- sessions are tamper resistant (authenticated encryption, key remains secret, or stored server side)
- all state changing requests never use GET or HEAD requests
*/

var ErrLengthMismatch = errors.New("xor: length mismatch")

const csrfTokenLength = 32

func xor(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, ErrLengthMismatch
	}

	res := make([]byte, len(a))

	for i := range len(a) {
		res[i] = a[i] ^ b[i]
	}

	return res, nil
}

var safeMethods = []string{http.MethodGet, http.MethodHead, http.MethodTrace}

func CSRF(cookieName string, store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var (
				session             *sessions.Session
				err                 error
				token               []byte
				ok                  bool
				otp                 []byte
				maskedToken         []byte
				tokenEncoded        string
				requestTokenEncoded string
				requestToken        []byte
				unmaskedToken       []byte
			)

			session, err = store.Get(r, cookieName)
			if err != nil {
				slog.ErrorContext(r.Context(), "unable to get session", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// do we have a csrf token? if not, add one
			if token, ok = session.Values["CSRFToken"].([]byte); !ok {
				token = make([]byte, csrfTokenLength)
				_, err = rand.Read(token)
				if err != nil {
					slog.ErrorContext(r.Context(), "unable to generate CSRF token", "error", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				session.Values["CSRFToken"] = token
				err = session.Save(r, w)
				if err != nil {
					slog.ErrorContext(r.Context(), "unable to save session", "error", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}

			otp = make([]byte, csrfTokenLength)
			_, err = rand.Read(otp)
			if err != nil {
				slog.ErrorContext(r.Context(), "unable to generate csrf otp", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			maskedToken, err = xor(token, otp)
			if err != nil {
				// this should never happen, it's a bug
				slog.ErrorContext(r.Context(), "failed to mask token", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			tokenEncoded = base64.StdEncoding.EncodeToString(append(otp, maskedToken...))
			w.Header().Set("X-CSRF-Token", tokenEncoded)

			// if the request uses a safe verb we don't need to check
			if slices.Contains(safeMethods, r.Method) {
				goto end
			}

			// above this should only be fallible from an implementation bug. Below failures can be caused by clients
			// behaving incorrectly, badly, or malicious requests

			requestTokenEncoded = r.Header.Get("X-CSRF-Token")
			requestToken, err = base64.StdEncoding.DecodeString(requestTokenEncoded)
			if err != nil {
				// incorrectly formed token
				slog.ErrorContext(r.Context(), "malformed csrf token", "error", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if len(requestToken) != (csrfTokenLength * 2) {
				slog.ErrorContext(r.Context(), "csrf token wrong size", "expected", csrfTokenLength*2, "actual", len(requestToken))
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			unmaskedToken, err = xor(requestToken[:csrfTokenLength], requestToken[csrfTokenLength:])
			if err != nil {
				// this shouldn't be possible because we've already confirmed the size. The only error xor() can return
				// is the wrong size
				slog.ErrorContext(r.Context(), "failed to unmask token", "error", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if subtle.ConstantTimeCompare(token, unmaskedToken) != 1 {
				// token doesn't match
				slog.WarnContext(r.Context(), "csrf token mismatch", "expected", requestToken[csrfTokenLength:], "actual", token)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

		end:
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
