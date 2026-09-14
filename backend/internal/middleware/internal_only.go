package middleware

import (
	"log/slog"
	"net/http"
)

func InternalOnly(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Now it just compares a local string in memory—zero locks, zero OS calls!
			if secret == "" || r.Header.Get("X-Internal-Secret") != secret {
				slog.Warn("Blocked unauthorized attempt to access internal Gateway route",
					slog.String("ip", r.RemoteAddr),
					slog.String("path", r.URL.Path),
				)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
