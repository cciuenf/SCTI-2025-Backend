package services

import (
	"errors"
	"scti/internal/models"
	"scti/internal/repos"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo  *repos.UserRepo
	JWTSecret string
}

func NewAuthService(repo *repos.UserRepo, secret string) *AuthService {
	return &AuthService{
		UserRepo:  repo,
		JWTSecret: secret,
	}
}

func (s *AuthService) Register(email, password, name string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	// Checar se já existe usuário
	existing, _ := s.UserRepo.FindByEmail(email)
	if existing != nil {
		return "", "", errors.New("usuario já existe")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	user := &models.User{
		UserID:       uuid.New().String(),
		Name:         name,
		Email:        email,
		Password:     string(hashedPassword),
		RefreshToken: uuid.New().String(), // cria refresh inicial
	}

	if err := s.UserRepo.Create(user); err != nil {
		return "", "", err
	}

	accessToken, err := s.generateJWT(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, user.RefreshToken, nil
}

func (s *AuthService) Login(email, password string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil || user == nil {
		return "", "", errors.New("email ou senha inválidos")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("email ou senha inválidos")
	}

	accessToken, err := s.generateJWT(user)
	if err != nil {
		return "", "", err
	}

	newRefresh := uuid.New().String()
	if err := s.UserRepo.UpdateRefreshToken(user.UserID, newRefresh); err != nil {
		return "", "", err
	}

	return accessToken, newRefresh, nil
}

func (s *AuthService) generateJWT(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.UserID,
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	})
	return token.SignedString([]byte(s.JWTSecret))
}
