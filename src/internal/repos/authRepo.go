package repos

import (
	"errors"
	"log"
	"scti/config"
	"scti/internal/models"

	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthRepo struct {
	DB *gorm.DB
}

func NewAuthRepo(db *gorm.DB) *AuthRepo {
	return &AuthRepo{DB: db}
}

func (r *AuthRepo) CreateUser(user *models.User) error {
	err := r.DB.Create(user).Error
	if err != nil {
		return errors.New("AUTH-REPO: Error creating user: " + err.Error())
	}
	return nil
}

func (r *AuthRepo) CreateMasterUser() {
	var existingUser models.User
	err := r.DB.Where("email = ?", config.GetSystemEmail()).First(&existingUser).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.GetMasterUserPass()), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Could not hash password for MasterUser")
	}

	userID := uuid.New().String()
	MasterUser := &models.User{
		ID:         userID,
		Name:       "Master",
		LastName:   "User",
		Email:      config.GetSystemEmail(),
		IsVerified: true,
		UserPass: models.UserPass{
			ID:       userID,
			Password: string(hashedPassword),
		},
		IsMasterUser: true,
	}

	err = r.DB.Create(MasterUser).Error
	if err != nil {
		log.Fatal("Could not create MasterUser")
	}
}

func (r *AuthRepo) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.
		Preload("UserPass").
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("AUTH-REPO: User not found: %v", err)
	}
	return &user, nil
}

func (r *AuthRepo) FindUserByID(id string) (*models.User, error) {
	var user models.User
	err := r.DB.
		Preload("UserPass").
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("AUTH-REPO: User not found: %v", err)
	}
	return &user, nil
}

func (r *AuthRepo) CreateRefreshToken(userID, refreshToken string) error {
	token := models.RefreshToken{
		UserID:   userID,
		TokenStr: refreshToken,
	}

	err := r.DB.Create(&token).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepo) GetRefreshTokens(userID string) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := r.DB.Where("user_id = ?", userID).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *AuthRepo) UpdateRefreshToken(userID, oldToken, newToken string) error {
	return r.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND token_str = ?", userID, oldToken).
		Update("token_str", newToken).Error
}

func (r *AuthRepo) FindRefreshToken(userID, tokenStr string) *models.RefreshToken {
	var token models.RefreshToken
	err := r.DB.
		Where("user_id = ? AND token_str = ?", userID, tokenStr).
		First(&token).Error

	if err != nil {
		return nil
	}

	return &token
}

func (r *AuthRepo) DeleteRefreshToken(userID, tokenStr string) error {
	return r.DB.
		Where("user_id = ? AND token_str = ?", userID, tokenStr).
		Delete(&models.RefreshToken{}).Error
}

func (r *AuthRepo) GetAllAdminStatusFromUser(userID string) ([]models.AdminStatus, error) {
	var adminStatuses []models.AdminStatus
	err := r.DB.Where("user_id = ?", userID).Find(&adminStatuses).Error
	if err != nil {
		return nil, err
	}
	return adminStatuses, nil
}
