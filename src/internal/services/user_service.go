package services

import (
	"errors"
	"scti/internal/models"
	repos "scti/internal/repositories"

	"github.com/google/uuid"
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

func (s *UserService) GetUserInfoFromID(userID string) (*models.UserInfo, error) {
	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	info := models.UserInfo{
		Name:         user.Name,
		LastName:     user.LastName,
		Email:        user.Email,
		IsUenf:       user.IsUenf,
		UenfSemester: user.UenfSemester,
	}

	return &info, nil
}

func (s *UserService) GetUserInfoFromIDBatch(id_array []string) ([]models.UserInfo, error) {
	var result []models.UserInfo
	for _, id := range id_array {
		if _, err := uuid.Parse(id); err != nil {
			// Malformed UUID
			result = append(result, models.UserInfo{
				Name:         "MALFORMED USER",
				LastName:     "MALFORMED USER",
				Email:        "MALFORMED USER",
				IsUenf:       false,
				UenfSemester: -1,
			})
			continue
		}

		user, err := s.UserRepo.GetUserByID(id)
		if err != nil {
			// Could not find user, treat as malformed
			result = append(result, models.UserInfo{
				Name:         "MISSING USER",
				LastName:     "MISSING USER",
				Email:        "MISSING USER",
				IsUenf:       false,
				UenfSemester: -1,
			})
			continue
		}

		info := models.UserInfo{
			Name:         user.Name,
			LastName:     user.LastName,
			Email:        user.Email,
			IsUenf:       user.IsUenf,
			UenfSemester: user.UenfSemester,
		}
		result = append(result, info)
	}

	return result, nil
}
