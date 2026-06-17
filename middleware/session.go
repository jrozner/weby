package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
)

func Session(cookieName string, store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, cookieName)
			if err != nil {
				slog.WarnContext(r.Context(), "session cookie invalid", "error", err)
			}

			// Get will return the session if it exists, a new session if it doesn't exist, and a new session with an
			// error if it can't be decoded. However, in the event that the cookie name is invalid it will return an
			// error without a new session (nil). We need to do a nil check to avoid a nil pointer deref on the IsNew
			// check.
			if session == nil {
				slog.ErrorContext(r.Context(), "returned a nil session, bailing out")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if session.IsNew {
				err = session.Save(r, w)
				if err != nil {
					slog.ErrorContext(r.Context(), "unable to save session")
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
