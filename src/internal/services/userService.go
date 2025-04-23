package services

import (
	"scti/internal/models"
	repos "scti/internal/repositories"
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
