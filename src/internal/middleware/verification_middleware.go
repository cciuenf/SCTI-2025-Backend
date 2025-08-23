package middleware

import (
	"net/http"
	"scti/internal/models"
	u "scti/internal/utilities"
)

func IsVerifiedMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := r.Context().Value(models.UserContextValue)
			if userCtx == nil {
				u.SendError(w, []string{"user context not found"}, "is-verified-middleware", http.StatusUnauthorized)
				return
			}

			user, ok := userCtx.(*models.UserClaims)
			if !ok {
				u.SendError(w, []string{"invalid user context"}, "is-verified-middleware", http.StatusUnauthorized)
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
