package handlers

import (
	"encoding/json"
	"net/http"
	"scti/config"
	"scti/internal/models"
	"scti/internal/services"
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

	acess_token, refresh, err := h.AuthService.Login(user.Email, user.Password, r)
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

	acess_token, refresh, err := h.AuthService.Login(user.Email, user.Password, r)
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
	user := u.GetUserFromContext(r.Context())
	if user == nil {
		u.Send(w, "User not found in context", nil, http.StatusUnauthorized)
		return
	}

	refreshHeader := r.Header.Get("Refresh")
	refreshTokenString := strings.TrimPrefix(refreshHeader, "Bearer ")

	err := h.AuthService.Logout(user.ID, refreshTokenString)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusUnauthorized)
		return
	}

	u.Send(w, "Logged out", nil, http.StatusOK)
}

func (h *AuthHandler) GetRefreshTokens(w http.ResponseWriter, r *http.Request) {
	user := u.GetUserFromContext(r.Context())
	if user == nil {
		u.Send(w, "User not found in context", nil, http.StatusBadRequest)
		return
	}

	refreshTokens, err := h.AuthService.GetRefreshTokens(user.ID)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "", refreshTokens, http.StatusOK)
}

func (h *AuthHandler) RevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	user := u.GetUserFromContext(r.Context())
	if user == nil {
		u.Send(w, "User not found in context", nil, http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Token string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		u.Send(w, "Error decoding JSON", nil, http.StatusBadRequest)
		return
	}

	if requestBody.Token == "" {
		u.Send(w, "Token is required", nil, http.StatusBadRequest)
		return
	}

	err := h.AuthService.RevokeRefreshToken(user.ID, requestBody.Token)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "Refresh token revoked", nil, http.StatusOK)
}

func (h *AuthHandler) VerifyAccount(w http.ResponseWriter, r *http.Request) {
	userClaims := u.GetUserFromContext(r.Context())
	if userClaims == nil {
		u.Send(w, "User not found in context", nil, http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		u.Send(w, "Error decoding JSON", nil, http.StatusBadRequest)
		return
	}

	if requestBody.Token == "" {
		u.Send(w, "Token is required", nil, http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.AuthRepo.FindUserByID(userClaims.ID)
	if err != nil {
		u.Send(w, "couldn't find user", nil, http.StatusBadRequest)
	}

	err = h.AuthService.VerifyUser(user, requestBody.Token)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "Account verified", nil, http.StatusOK)
}

func (h *AuthHandler) VerifyJWT(w http.ResponseWriter, r *http.Request) {
	var secretKey string = config.GetJWTSecret()
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		u.Send(w, "Authorization header is required", nil, http.StatusUnauthorized)
		return
	}

	refreshHeader := r.Header.Get("Refresh")
	if refreshHeader == "" {
		u.Send(w, "Authorization header is required", nil, http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		u.Send(w, "Authorization header format must be Bearer {token}", nil, http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(refreshHeader, "Bearer ") {
		u.Send(w, "Refresh header format must be Bearer {token}", nil, http.StatusUnauthorized)
		return
	}

	accessTokenString := strings.TrimPrefix(authHeader, "Bearer ")
	refreshTokenString := strings.TrimPrefix(refreshHeader, "Bearer ")

	accessToken, err := jwt.ParseWithClaims(accessTokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			u.Send(w, "Invalid signing method", nil, http.StatusUnauthorized)
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		u.Send(w, "Invalid access token:"+err.Error(), nil, http.StatusUnauthorized)
		return
	}

	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			u.Send(w, "Invalid signing method", nil, http.StatusUnauthorized)
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		u.Send(w, "Invalid refresh token:"+err.Error(), nil, http.StatusUnauthorized)
		return
	}

	if !accessToken.Valid || !refreshToken.Valid {
		u.Send(w, "Token is not valid or has expired", nil, http.StatusUnauthorized)
		return
	}

	u.Send(w, "Success", nil, http.StatusOK)
}
