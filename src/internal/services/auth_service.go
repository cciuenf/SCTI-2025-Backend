package services

import (
	"errors"
	"scti/internal/models"
	"scti/internal/repos"
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

func (s *AuthService) Register(email, password, name string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		UserID:   uuid.New().String(),
		Nome:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	return s.UserRepo.Create(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.UserRepo.FindByEmail(email) // Chamando o método FindByEmail
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.UserID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(s.JWTSecret))
}
