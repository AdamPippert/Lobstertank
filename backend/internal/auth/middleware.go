package auth

import (
	"log/slog"
	"net/http"
)

// Middleware returns an HTTP middleware that enforces authentication using
// the given provider.
func Middleware(provider Provider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, err := provider.Authenticate(r.Context(), r)
			if err != nil {
				slog.Warn("authentication failed",
					"path", r.URL.Path,
					"remote", r.RemoteAddr,
					"error", err,
				)
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := ContextWithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
