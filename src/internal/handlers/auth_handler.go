package handlers

import (
	"encoding/json"
	"net/http"
	"scti/config"
	"scti/internal/models"
	"scti/internal/services"
	"scti/internal/utilities"
	u "scti/internal/utilities"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Erro ao ler JSON", http.StatusBadRequest)
		return
	}

	err := h.AuthService.Register(user.Email, user.Password, user.Name, user.LastName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	acess_token, refresh, err := h.AuthService.Login(user.Email, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	u.Send(w, "", map[string]string{
		"access_token":  acess_token,
		"refresh_token": refresh,
	}, http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var user models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Erro ao ler JSON", http.StatusBadRequest)
		return
	}

	acess_token, refresh, err := h.AuthService.Login(user.Email, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	u.Send(w, "", map[string]string{
		"access_token":  acess_token,
		"refresh_token": refresh,
	}, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user := utilities.GetUserFromContext(r.Context())
	if user == nil {
		u.Send(w, "mw-error: User not found in context", nil, http.StatusUnauthorized)
		return
	}

	refreshHeader := r.Header.Get("Refresh")
	refreshTokenString := strings.TrimPrefix(refreshHeader, "Bearer ")

	err := h.AuthService.Logout(user.ID, refreshTokenString)
	if err != nil {
		u.Send(w, "mw-error: "+err.Error(), nil, http.StatusUnauthorized)
		return
	}

	u.Send(w, "Logged out", nil, http.StatusOK)
}

func (h *AuthHandler) VerifyJWT(w http.ResponseWriter, r *http.Request) {
	var secretKey string = config.GetJWTSecret()
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		u.Send(w, "mw-error: Authorization header is required", nil, http.StatusUnauthorized)
		return
	}

	refreshHeader := r.Header.Get("Refresh")
	if refreshHeader == "" {
		u.Send(w, "mw-error: Authorization header is required", nil, http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		u.Send(w, "mw-error: Authorization header format must be Bearer {token}", nil, http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(refreshHeader, "Bearer ") {
		u.Send(w, "mw-error: Refresh header format must be Bearer {token}", nil, http.StatusUnauthorized)
		return
	}

	accessTokenString := strings.TrimPrefix(authHeader, "Bearer ")
	refreshTokenString := strings.TrimPrefix(refreshHeader, "Bearer ")

	accessToken, err := jwt.ParseWithClaims(accessTokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			u.Send(w, "mw-error: Invalid signing method", nil, http.StatusUnauthorized)
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		u.Send(w, "mw-error: Invalid access token:"+err.Error(), nil, http.StatusUnauthorized)
		return
	}

	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			u.Send(w, "mw-error: Invalid signing method", nil, http.StatusUnauthorized)
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		u.Send(w, "mw-error: Invalid refresh token:"+err.Error(), nil, http.StatusUnauthorized)
		return
	}

	if !accessToken.Valid || !refreshToken.Valid {
		u.Send(w, "mw-error: Token is not valid or has expired", nil, http.StatusUnauthorized)
		return
	}

	u.Send(w, "Success", nil, http.StatusOK)
}
