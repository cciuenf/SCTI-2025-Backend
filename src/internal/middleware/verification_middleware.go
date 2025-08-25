package middleware

import (
	"net/http"
	u "scti/internal/utilities"
)

func IsVerifiedMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := u.GetUserFromContext(r.Context())
			if user == nil {
				u.SendError(w, []string{"user context not found"}, "is-verified-middleware", http.StatusUnauthorized)
				return
			}

			if !user.IsVerified {
				u.SendError(w, []string{"account verification required"}, "is-verified-middleware", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
