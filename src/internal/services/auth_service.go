package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"scti/config"
	"scti/internal/models"
	repos "scti/internal/repositories"
	"scti/internal/utilities"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
		return errors.New("all fields are required")
	}

	email = strings.TrimSpace(strings.ToLower(email))

	// Regex to check email
	if !utilities.IsValidEmail(email) {
		return errors.New("invalid email format")
	}

	exists, _ := s.AuthRepo.FindUserByEmail(email)
	if exists != nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userID := uuid.New().String()
	user := &models.User{
		ID:         userID,
		Name:       name,
		LastName:   last_name,
		Email:      email,
		IsVerified: false,
		UserPass: models.UserPass{
			ID:       userID,
			Password: string(hashedPassword),
		},
	}

	if err := s.AuthRepo.CreateUser(user); err != nil {
		return err
	}

	verificationNumber := utilities.GenerateVerificationCode()

	if err := s.AuthRepo.CreateUserVerification(user.ID, verificationNumber); err != nil {
		return err
	}

	err = s.SendVerificationEmail(user, verificationNumber)
	if err != nil {
		return err
	}

	return nil
}

type verificationEmailData struct {
	UserName         string
	VerificationCode string
	SupportEmail     string
}

var templateFuncs = template.FuncMap{
	"substr": func(s string, i, j int) string {
		if i >= len(s) {
			return ""
		}
		if j > len(s) {
			j = len(s)
		}
		return s[i:j]
	},
}

func (s *AuthService) SendVerificationEmail(user *models.User, verificationNumber int) error {
	from := config.GetSystemEmail()
	password := config.GetSystemEmailPass()

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	templatePath := filepath.Join("templates", "verification_email.html")

	file, err := os.Open(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open email template: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read email template: %v", err)
	}

	tmpl, err := template.New("emailTemplate").Funcs(templateFuncs).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	verificationCode := fmt.Sprintf("%06d", verificationNumber)

	data := verificationEmailData{
		UserName:         user.Name + " " + user.LastName,
		VerificationCode: verificationCode,
		SupportEmail:     config.GetSystemEmail(),
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	subject := "Verificação de Conta"

	message := []byte(fmt.Sprintf("Subject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
		subject, body.String()))

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{user.Email}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (s *AuthService) VerifyUser(user *models.User, token string) error {
	if user.IsVerified {
		return errors.New("user is already verified")
	}

	storedToken, err := s.AuthRepo.GetUserVerification(user.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("no verification token found")
		}
		return err
	}

	tokenInt, err := strconv.Atoi(token)
	if err != nil {
		return err
	}

	if storedToken != tokenInt {
		return errors.New("invalid verification token")
	}

	user.IsVerified = true
	err = s.AuthRepo.UpdateUser(user)
	if err != nil {
		return err
	}

	err = s.AuthRepo.DeleteUserVerification(user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) Login(email, password string, r *http.Request) (string, string, error) {
	if email == "" || password == "" {
		return "", "", errors.New("all fields are required")
	}

	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.AuthRepo.FindUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	if user == nil {
		return "", "", errors.New("user with specified email not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPass.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid password")
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

func (s *AuthService) MakeJSONAdminMap(userID string) (string, error) {
	statuses, err := s.AuthRepo.GetAllAdminStatusFromUser(userID)
	if err != nil {
		return "", err
	}

	if statuses == nil {
		return "", errors.New("user has no admin status")
	}

	adminMap := make(map[string]map[string]string)
	for _, status := range statuses {
		if _, ok := adminMap[status.EventSlug]; !ok {
			adminMap[status.EventSlug] = make(map[string]string)
		}
		adminMap[status.EventSlug][string(status.AdminType)] = status.EventID
	}

	jsonString, err := json.Marshal(adminMap)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

func (s *AuthService) GenerateAcessToken(user *models.User) (string, error) {
	adminMap, err := s.MakeJSONAdminMap(user.ID)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":           user.ID,
		"name":         user.Name,
		"last_name":    user.LastName,
		"email":        user.Email,
		"admin_status": adminMap,
		"is_verified":  user.IsVerified,
		"exp":          time.Now().Add(5 * time.Minute).Unix(),
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
		return nil, errors.New("refresh token not found")
	}
	return token, nil
}

func (s *AuthService) GeneratePasswordResetToken(userID string) (string, error) {
	claims := &models.PasswordResetClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:          userID,
		IsPasswordReset: true,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.JWTSecret))
}

func (s *AuthService) SendPasswordResetEmail(user *models.User, resetToken string) error {
	from := config.GetSystemEmail()
	password := config.GetSystemEmailPass()

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	resetLink := fmt.Sprintf("http://%s/change-password?token=%s", config.GetSiteURL(), resetToken)

	templatePath := filepath.Join("templates", "password_reset_email.html")
	file, err := os.Open(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open email template: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read email template: %v", err)
	}

	tmpl, err := template.New("resetTemplate").Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	data := struct {
		UserName     string
		ResetLink    string
		SupportEmail string
	}{
		UserName:     user.Name + " " + user.LastName,
		ResetLink:    resetLink,
		SupportEmail: config.GetSystemEmail(),
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	subject := "Redefinição de Senha"
	message := []byte(fmt.Sprintf(
		"Subject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
		subject, body.String()))

	auth := smtp.PlainAuth("", from, password, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{user.Email}, message)
}

func (s *AuthService) InitiatePasswordReset(email string) error {
	user, err := s.AuthRepo.FindUserByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	resetToken, err := s.GeneratePasswordResetToken(user.ID)
	if err != nil {
		return err
	}

	return s.SendPasswordResetEmail(user, resetToken)
}

func (s *AuthService) ChangePassword(userID string, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.AuthRepo.UpdateUserPassword(userID, string(hashedPassword))
}
