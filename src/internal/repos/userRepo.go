package repos

import (
	"scti/internal/models"
	"gorm.io/gorm"
	"fmt"
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
	err = r.DB.Select("nome").Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

