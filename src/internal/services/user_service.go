package services

import (
	"errors"
	"scti/internal/models"
	repos "scti/internal/repositories"
)

type UserService struct {
	UserRepo *repos.UserRepo
}

func NewUserService(userRepo *repos.UserRepo) *UserService {
	return &UserService{UserRepo: userRepo}
}

func (s *UserService) CreateEventCreator(user *models.User, email string) (*models.User, error) {
	if !user.IsSuperUser {
		return nil, errors.New("user is not a super user")
	}

	creator, err := s.UserRepo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if creator == nil {
		return nil, errors.New("user not found")
	}

	if creator.IsEventCreator {
		return nil, errors.New("user is already an event creator")
	}

	creator.IsEventCreator = true

	return s.UserRepo.UpdateUser(creator)
}
