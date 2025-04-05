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

func (s *AuthService) Register(email, password, name, last_name string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	existing, _ := s.UserRepo.FindByEmail(email)
	if existing != nil {
		return errors.New("usuario já existe")
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

	if err := s.UserRepo.Create(user); err != nil {
		return err
	}
	return nil
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

	accessToken, err := s.generateAcessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	if err, _ := s.UserRepo.CreateRefreshToken(user.ID, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) generateAcessToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":        user.ID,
		"name":      user.Name,
		"last_name": user.LastName,
		"event":     user.Event,
		"is_paid":   user.IsPaid,
		"exp":       time.Now().Add(5 * time.Minute).Unix(),
	})
	return token.SignedString([]byte(s.JWTSecret))
}

func (s *AuthService) generateRefreshToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"token_id": -1,
		"exp":      time.Now().Add(2 * 24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(s.JWTSecret))
}
