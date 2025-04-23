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

// Register godoc
// @Summary      Register new user and send a verification email
// @Description  Register a new user in the system, generates a verification code that is stored
// @Description  in the database for 24 hours and sent in a verification email to the user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body docs.UserRegisterRequest true "User registration info"
// @Success      201  {object}  docs.StandardResponse
// @Failure      400  {object}  docs.StandardResponse
// @Failure      500  {object}  docs.StandardResponse
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		u.SendError(w, []string{"error reading json:" + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	err := h.AuthService.Register(user.Email, user.Password, user.Name, user.LastName)
	if err != nil {
		u.SendError(w, []string{"error registering user:" + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	acess_token, refresh, err := h.AuthService.Login(user.Email, user.Password, r)
	if err != nil {
		u.SendError(w, []string{"error trying to login:" + err.Error()}, "auth-stack", http.StatusUnauthorized)
		return
	}

	u.SendSuccess(w, map[string]string{
		"access_token":  acess_token,
		"refresh_token": refresh,
	}, "", http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var user models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		u.SendError(w, []string{"error reading json:" + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	acess_token, refresh, err := h.AuthService.Login(user.Email, user.Password, r)
	if err != nil {
		u.SendError(w, []string{"error trying to login:" + err.Error()}, "auth-stack", http.StatusUnauthorized)
		return
	}

	u.SendSuccess(w, map[string]string{
		"access_token":  acess_token,
		"refresh_token": refresh,
	}, "", http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user := u.GetUserFromContext(r.Context())

	refreshHeader := r.Header.Get("Refresh")
	refreshTokenString := strings.TrimPrefix(refreshHeader, "Bearer ")

	err := h.AuthService.Logout(user.ID, refreshTokenString)
	if err != nil {
		u.SendError(w, []string{"error trying to logout:" + err.Error()}, "auth-stack", http.StatusUnauthorized)
		return
	}

	u.SendSuccess(w, nil, "logged out", http.StatusOK)
}

func (h *AuthHandler) GetRefreshTokens(w http.ResponseWriter, r *http.Request) {
	user := u.GetUserFromContext(r.Context())

	refreshTokens, err := h.AuthService.GetRefreshTokens(user.ID)
	if err != nil {
		u.SendError(w, []string{"error getting refresh tokens:" + err.Error()}, "auth-stack", http.StatusUnauthorized)
		return
	}

	u.SendSuccess(w, refreshTokens, "", http.StatusOK)
}

func (h *AuthHandler) RevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	user := u.GetUserFromContext(r.Context())

	var requestBody struct {
		Token string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		u.SendError(w, []string{"error reading json:" + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	if requestBody.Token == "" {
		u.SendError(w, []string{"refresh token to be revoked is required:"}, "auth-stack", http.StatusBadRequest)
		return
	}

	err := h.AuthService.RevokeRefreshToken(user.ID, requestBody.Token)
	if err != nil {
		u.SendError(w, []string{"error revoking token: " + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, nil, "refresh token revoked successfully", http.StatusOK)
}

func (h *AuthHandler) VerifyAccount(w http.ResponseWriter, r *http.Request) {
	userClaims := u.GetUserFromContext(r.Context())

	var requestBody struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		u.SendError(w, []string{"error reading json: " + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	if requestBody.Token == "" {
		u.SendError(w, []string{"verification token is required"}, "auth-stack", http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.AuthRepo.FindUserByID(userClaims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	err = h.AuthService.VerifyUser(user, requestBody.Token)
	if err != nil {
		u.SendError(w, []string{"error verifying user: " + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, nil, "Account verified", http.StatusOK)
}

func (h *AuthHandler) VerifyJWT(w http.ResponseWriter, r *http.Request) {
	var secretKey string = config.GetJWTSecret()
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		u.SendError(w, []string{"\"Authorization\" header is required"}, "auth-stack", http.StatusBadRequest)
		return
	}

	refreshHeader := r.Header.Get("Refresh")
	if refreshHeader == "" {
		u.SendError(w, []string{"\"Refresh\" header is required"}, "auth-stack", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		u.SendError(w, []string{"authorization header format must be \"Bearer {token}\""}, "auth-stack", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(refreshHeader, "Bearer ") {
		u.SendError(w, []string{"refresh header format must be \"Bearer {token}\""}, "auth-stack", http.StatusBadRequest)
		return
	}

	accessTokenString := strings.TrimPrefix(authHeader, "Bearer ")
	refreshTokenString := strings.TrimPrefix(refreshHeader, "Bearer ")

	_, err := jwt.ParseWithClaims(accessTokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		u.SendError(w, []string{"error parsing access token: " + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		u.SendError(w, []string{"error parsing refresh token: " + err.Error()}, "auth-stack", http.StatusBadRequest)
		return
	}

	if !refreshToken.Valid {
		u.SendError(w, []string{"invalid refresh token"}, "auth-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, nil, "Success", http.StatusOK)
}
