package middleware

import (
	"context"
	"net/http"
	"strings"

	"scti/config"
	"scti/internal/models"
	"scti/internal/services"
	"scti/internal/utilities"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secretKey := config.GetJWTSecret()

			accessHeader := r.Header.Get("Authorization")
			if accessHeader == "" {
				utilities.Send(w, "mw-error: Authorization header is required", nil, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(accessHeader, "Bearer ") {
				utilities.Send(w, "mw-error: Authorization header format must be Bearer {token}", nil, http.StatusUnauthorized)
				return
			}
			accessTokenString := strings.TrimPrefix(accessHeader, "Bearer ")

			refreshHeader := r.Header.Get("Refresh")
			if refreshHeader == "" {
				utilities.Send(w, "mw-error: Refresh token required for token renewal", nil, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(refreshHeader, "Bearer ") {
				utilities.Send(w, "mw-error: Refresh header format must be Bearer {token}", nil, http.StatusUnauthorized)
				return
			}
			refreshTokenString := strings.TrimPrefix(refreshHeader, "Bearer ")

			accessToken, accessTokenErr := jwt.ParseWithClaims(accessTokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secretKey), nil
			})

			refreshToken, refreshErr := jwt.ParseWithClaims(refreshTokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secretKey), nil
			})

			if accessTokenErr != nil && !strings.Contains(accessTokenErr.Error(), "token is expired") {
				utilities.Send(w, "mw-error: Invalid access token: "+accessTokenErr.Error(), nil, http.StatusUnauthorized)
				return
			}

			var accessClaims *models.UserClaims
			var ok bool

			if accessToken != nil {
				accessClaims, ok = accessToken.Claims.(*models.UserClaims)
				if !ok {
					utilities.Send(w, "mw-error: Invalid token claims", nil, http.StatusUnauthorized)
					return
				}
			}

			if refreshErr != nil {
				utilities.Send(w, "mw-error: Invalid refresh token: "+refreshErr.Error(), nil, http.StatusUnauthorized)
				return
			}

			if !refreshToken.Valid {
				utilities.Send(w, "mw-error: Refresh token is expired or invalid", nil, http.StatusUnauthorized)
				return
			}

			refreshClaims, ok := refreshToken.Claims.(*jwt.MapClaims)
			if !ok {
				utilities.Send(w, "mw-error: Invalid token claims", nil, http.StatusUnauthorized)
				return
			}

			userID, ok := (*refreshClaims)["id"].(string)
			if !ok {
				utilities.Send(w, "mw-error: Invalid user ID in refresh token", nil, http.StatusUnauthorized)
				return
			}

			storedToken, err := authService.FindRefreshToken(userID, refreshTokenString)
			if err != nil || storedToken == nil {
				utilities.Send(w, "mw-error: Refresh token not found or revoked", nil, http.StatusUnauthorized)
				return
			}

			if accessToken != nil && accessToken.Valid {
				ctx := context.WithValue(r.Context(), "user", accessClaims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			user, err := authService.AuthRepo.FindUserByID(userID)
			if err != nil {
				utilities.Send(w, "mw-error: User not found", nil, http.StatusUnauthorized)
				return
			}

			newAccessToken, err := authService.GenerateAcessToken(user)
			if err != nil {
				utilities.Send(w, "mw-error: Failed to generate new access token", nil, http.StatusInternalServerError)
				return
			}

			newRefreshToken, err := authService.GenerateRefreshToken(user.ID)
			if err != nil {
				utilities.Send(w, "mw-error: Failed to generate new refresh token", nil, http.StatusInternalServerError)
				return
			}

			if err := authService.AuthRepo.UpdateRefreshToken(user.ID, refreshTokenString, newRefreshToken); err != nil {
				utilities.Send(w, "mw-error: Failed to update refresh token", nil, http.StatusInternalServerError)
				return
			}

			newAccessJWT, _ := jwt.ParseWithClaims(newAccessToken, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})

			newAccessClaims, _ := newAccessJWT.Claims.(*models.UserClaims)

			w.Header().Set("X-New-Access-Token", newAccessToken)
			w.Header().Set("X-New-Refresh-Token", newRefreshToken)

			ctx := context.WithValue(r.Context(), "user", newAccessClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
