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

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type UserRegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
	Name     string `json:"name" example:"John"`
	LastName string `json:"lastName" example:"Doe"`
}

// Register godoc
// @Summary      Register new user and send a verification email
// @Description  Register a new user in the system, generates a verification code that is stored
// @Description  in the database for 24 hours and sent in a verification email to the user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body UserRegisterRequest true "User registration info"
// @Success      201  {object}  NoMessageSuccessResponse{data=AuthTokensResponse}
// @Failure      400  {object}  AuthStandardErrorResponse
// @Failure      401  {object}  AuthStandardErrorResponse
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
	}, "", http.StatusCreated)
}

type UserLoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

// Login godoc
// @Summary      Logs in the user
// @Description  Logging successfully creates a refresh token in the database so the user can
// @Description  invalidate specific session from any other session\n
// @Description  Returns both an Access Token of 5 minutes duration and a Refresh Token of 2 days duration
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body UserLoginRequest true "User login info"
// @Success      200  {object}  NoMessageSuccessResponse{data=AuthTokensResponse}
// @Failure      400  {object}  AuthStandardErrorResponse
// @Failure      401  {object}  AuthStandardErrorResponse
// @Router       /login [post]
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

// Logout godoc
// @Summary      Logs out the user
// @Description  Invalidates the refresh token used in the request in the database, effectively logging out the user from the current session
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      401  {object}  AuthStandardErrorResponse
// @Router       /logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user := u.GetUserFromContext(r.Context())

	refreshHeader := r.Header.Get("Refresh")
	refreshTokenString := strings.TrimPrefix(refreshHeader, "Bearer ")

	err := h.AuthService.Logout(user.ID, refreshTokenString)
	if err != nil {
		u.SendError(w, []string{"error trying to logout:" + err.Error()}, "auth-stack", http.StatusUnauthorized)
		return
	}

	u.SendSuccess(w, nil, "logged out successfully", http.StatusOK)
}

// GetRefreshTokens godoc
// @Summary      Get user's refresh tokens
// @Description  Returns all refresh tokens associated with the user's account
// @Tags         auth
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.RefreshToken}
// @Failure      401  {object}  AuthStandardErrorResponse
// @Router       /refresh-tokens [get]
func (h *AuthHandler) GetRefreshTokens(w http.ResponseWriter, r *http.Request) {
	user := u.GetUserFromContext(r.Context())

	refreshTokens, err := h.AuthService.GetRefreshTokens(user.ID)
	if err != nil {
		u.SendError(w, []string{"error getting refresh tokens:" + err.Error()}, "auth-stack", http.StatusUnauthorized)
		return
	}

	u.SendSuccess(w, refreshTokens, "", http.StatusOK)
}

type RevokeTokenRequest struct {
	Token string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RevokeRefreshToken godoc
// @Summary      Revoke a refresh token
// @Description  Invalidates a specific refresh token for the authenticated user
// @Description  Can't be passed the same refresh token the user is using to access the route
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        request body RevokeTokenRequest true "Refresh token to revoke"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  AuthStandardErrorResponse
// @Failure      401  {object}  AuthStandardErrorResponse
// @Router       /revoke-refresh-token [post]
func (h *AuthHandler) RevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	user := u.GetUserFromContext(r.Context())

	var requestBody RevokeTokenRequest
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

type VerifyAccountRequest struct {
	Token string `json:"token" example:"123456"`
}

// VerifyAccount godoc
// @Summary      Verify user account with token
// @Description  Validates the verification token sent to user's email and marks the account as verified
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        request body VerifyAccountRequest true "Verification token from email"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  AuthStandardErrorResponse
// @Failure      401  {object}  AuthStandardErrorResponse
// @Router       /verify-account [post]
func (h *AuthHandler) VerifyAccount(w http.ResponseWriter, r *http.Request) {
	userClaims := u.GetUserFromContext(r.Context())

	var requestBody VerifyAccountRequest
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

	u.SendSuccess(w, nil, "account verified", http.StatusOK)
}

// VerifyJWT godoc
// @Summary      Verify JWT tokens
// @Description  Validates both access token and refresh token signatures
// @Tags         auth
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  AuthStandardErrorResponse
// @Router       /verify-tokens [post]
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

	u.SendSuccess(w, nil, "success", http.StatusOK)
}
