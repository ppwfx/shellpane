package communication

import (
	"net/http"
	"strings"

	"github.com/ppwfx/shellpane/internal/business"
)

type BasicAuthConfig struct {
	Username string
	Password string
}

func WithUserIDMiddleware(handler http.Handler, userIDHeader string, defaultUserID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerUserID := r.Header.Get(userIDHeader)

		var userID string
		switch {
		case headerUserID != "":
			userID = headerUserID
		case defaultUserID != "":
			userID = defaultUserID
		default:
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := business.WithUserID(r.Context(), userID)

		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	}
}

func WithBasicAuthMiddleware(next http.Handler, config BasicAuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if u != config.Username || p != config.Password {
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func CorsMiddleware(next http.Handler, corsOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
