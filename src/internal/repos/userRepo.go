package repos

import (
	"fmt"
	"scti/internal/models"

	"gorm.io/gorm"
)

type UserRepo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (r *UserRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("usuario nao encontrado: %v", err)
	}
	return &user, nil
}

func (r *UserRepo) Create(user *models.User) error {
	err := r.DB.Create(user).Error
	if err != nil {
		return fmt.Errorf("Nao pode criar o usuario")
	}
	return nil
}

func (r *UserRepo) GetAll() (users *[]models.User, err error) {
	err = r.DB.Select("name").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) UpdateRefreshToken(userID, refreshToken string) error {
	return r.DB.Model(&models.User{}).
		Where("user_id = ?", userID).
		Update("refresh_token", refreshToken).Error
}

func (r *UserRepo) CreateRefreshToken(userID, refreshToken string) (error, int64) {
	token := models.RefreshTokens{
		UserID:        userID,
		TokenStr: refreshToken,
	}

	err := r.DB.Create(&token).Error

	if err != nil {
		return err, -1
	}

	return err, token.ID
}
