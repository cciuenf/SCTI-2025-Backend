package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"scti/config"
	"scti/internal/models"
	"scti/internal/utilities"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var secretKey string = config.GetJWTSecret()
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utilities.Send(w, "mw-error: Authorization header is required", nil, http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			utilities.Send(w, "mw-error: Authorization header format must be Bearer {token}", nil, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				utilities.Send(w, "mw-error: Invalid signing method", nil, http.StatusUnauthorized)
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			utilities.Send(w, "mw-error: Invalid token:"+err.Error(), nil, http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			utilities.Send(w, "mw-error: Token is not valid or has expired", nil, http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*models.UserClaims); ok {
			// Expiry check
			if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
				utilities.Send(w, "mw-error: Token has expired", nil, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			utilities.Send(w, "mw-error: Invalid token claims", nil, http.StatusUnauthorized)
		}
	})
}

