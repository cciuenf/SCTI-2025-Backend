package services

import (
	"fmt"
	"scti/internal/models"
	"scti/internal/repos"
)

type UserService struct {
	Repo *repos.UserRepo
}

func NewUserService(r *repos.UserRepo) *UserService {
	return &UserService{Repo: r}
}

func (s *UserService) GetAll() (*[]models.User, error) {
	return s.Repo.GetAll()
}

func (s *UserService) Create(user *models.User) (*models.User, error) {
	if user.Name == "Joao" {
		return nil, fmt.Errorf("Usuarios nao podem se chamar Joao")
	}
	err := s.Repo.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
