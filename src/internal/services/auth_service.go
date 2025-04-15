package services

import (
	"errors"
	"net/http"
	"scti/internal/models"
	"scti/internal/repos"
	"scti/internal/utilities"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	AuthRepo  *repos.AuthRepo
	JWTSecret string
}

func NewAuthService(repo *repos.AuthRepo, secret string) *AuthService {
	return &AuthService{
		AuthRepo:  repo,
		JWTSecret: secret,
	}
}

func (s *AuthService) Register(email, password, name, last_name string) error {
	if email == "" || password == "" || name == "" || last_name == "" {
		return errors.New("AUTH: All fields are required")
	}

	// Regex to check email
	if !utilities.IsValidEmail(email) {
		return errors.New("AUTH: Invalid email format")
	}

	email = strings.TrimSpace(strings.ToLower(email))

	exists, _ := s.AuthRepo.FindUserByEmail(email)
	if exists != nil {
		return errors.New("AUTH: User already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		ID:       uuid.New().String(),
		Name:     name,
		LastName: last_name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.AuthRepo.CreateUser(user); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) Login(email, password string, r *http.Request) (string, string, error) {
	if email == "" || password == "" {
		return "", "", errors.New("AUTH: All fields are required")
	}

	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.AuthRepo.FindUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	if user == nil {
		return "", "", errors.New("AUTH: User with specified email not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("AUTH: Invalid password")
	}

	accessToken, err := s.GenerateAcessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.GenerateRefreshToken(user.ID, r)
	if err != nil {
		return "", "", err
	}

	if err := s.AuthRepo.CreateRefreshToken(user.ID, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Logout(ID, refreshTokenString string) error {
	err := s.AuthRepo.DeleteRefreshToken(ID, refreshTokenString)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) GetRefreshTokens(userID string) ([]models.RefreshToken, error) {
	tokens, err := s.AuthRepo.GetRefreshTokens(userID)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *AuthService) RevokeRefreshToken(userID, tokenStr string) error {
	err := s.AuthRepo.DeleteRefreshToken(userID, tokenStr)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) GenerateAcessToken(user *models.User) (string, error) {
	var adminType string
	if user.IsAdmin {
		adminType = "admin"
	}

	if user.IsMasterAdmin {
		adminType = "master_admin"
	}

	if user.IsMasterUser {
		adminType = "master_user"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":          user.ID,
		"name":        user.Name,
		"last_name":   user.LastName,
		"email":       user.Email,
		"event":       user.Event,
		"admin_type":  adminType,
		"is_verified": user.IsVerified,
		"exp":         time.Now().Add(5 * time.Minute).Unix(),
	})
	return token.SignedString([]byte(s.JWTSecret))
}

func (s *AuthService) GenerateRefreshToken(userID string, r *http.Request) (string, error) {
	userAgent := r.UserAgent()
	ipAddress := r.RemoteAddr
	// Se o server estiver atrás de um proxy, use o seguinte:
	// ipAddress = r.Header.Get("X-Forwarded-For")
	deviceInfo := utilities.ParseUserAgent(userAgent)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":          userID,
		"user_agent":  userAgent,
		"device_info": deviceInfo,
		"ip_address":  ipAddress,
		"last_used":   time.Now(),
		"exp":         time.Now().Add(2 * 24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(s.JWTSecret))
}

func (s *AuthService) FindRefreshToken(userID, tokenStr string) (*models.RefreshToken, error) {
	token := s.AuthRepo.FindRefreshToken(userID, tokenStr)
	if token == nil {
		return nil, errors.New("AUTH: Refresh token not found")
	}
	return token, nil
}
