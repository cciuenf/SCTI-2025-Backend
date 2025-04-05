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

func (s *UserService) Me(user *models.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("Usuario esta nulo")
	}
	return fmt.Sprintf(" Ola %v seu salario de %v e muito bom", user.Nome, user.Salario), nil
}

func (s *UserService) GetAll() (*[]models.User, error) {
	return s.Repo.GetAll()
}

func (s *UserService) Create(user *models.User) (*models.User, error) {
	if user.Nome == "Joao" {
		return nil, fmt.Errorf("Usuarios nao podem se chamar Joao")
	}
	err := s.Repo.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
