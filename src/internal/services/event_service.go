package services

import (
	"errors"
	"scti/internal/models"
	repos "scti/internal/repositories"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventService struct {
	EventRepo *repos.EventRepo
}

func NewEventService(repo *repos.EventRepo) *EventService {
	return &EventService{
		EventRepo: repo,
	}
}

func (s *EventService) CreateEvent(event *models.Event) error {
	event.ID = uuid.New().String()
	event.Slug = strings.ToLower(event.Slug)
	return s.EventRepo.CreateEvent(event)
}

func (s *EventService) GetEventBySlug(Slug string) (models.Event, error) {
	slug := strings.ToLower(Slug)
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (s *EventService) GetEventBySlugWithActivities(Slug string) (models.Event, error) {
	slug := strings.ToLower(Slug)
	event, err := s.EventRepo.GetEventBySlugWithActivities(slug)
	if err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (s *EventService) GetAllEvents() ([]models.Event, error) {
	events, err := s.EventRepo.GetAllEvents()
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *EventService) UpdateEvent(NewEvent *models.Event) (*models.Event, error) {
	StoredEvent, err := s.EventRepo.GetEventByID(NewEvent.ID)
	if err != nil {
		return nil, err
	}

	if NewEvent.Name != "" {
		StoredEvent.Name = NewEvent.Name
	}

	if NewEvent.Slug != "" {
		slug := strings.ToLower(NewEvent.Slug)
		StoredEvent.Slug = slug
	}

	if NewEvent.Description != "" {
		StoredEvent.Description = NewEvent.Description
	}

	if NewEvent.Location != "" {
		StoredEvent.Location = NewEvent.Location
	}

	if !NewEvent.StartDate.IsZero() {
		StoredEvent.StartDate = NewEvent.StartDate
	}

	if !NewEvent.EndDate.IsZero() {
		StoredEvent.EndDate = NewEvent.EndDate
	}

	StoredEvent.UpdatedAt = time.Now()

	if err := s.EventRepo.UpdateEvent(&StoredEvent); err != nil {
		return nil, err
	}
	return &StoredEvent, nil
}

func (s *EventService) UpdateEventBySlug(Slug string, NewEvent *models.Event) (*models.Event, error) {
	slug := strings.ToLower(Slug)
	StoredEvent, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return nil, err
	}

	if NewEvent.Name != "" {
		StoredEvent.Name = NewEvent.Name
	}

	if NewEvent.Description != "" {
		StoredEvent.Description = NewEvent.Description
	}

	if NewEvent.Slug != "" {
		slug := strings.ToLower(NewEvent.Slug)
		StoredEvent.Slug = slug
	}

	if NewEvent.Location != "" {
		StoredEvent.Location = NewEvent.Location
	}

	if !NewEvent.StartDate.IsZero() {
		StoredEvent.StartDate = NewEvent.StartDate
	}

	if !NewEvent.EndDate.IsZero() {
		StoredEvent.EndDate = NewEvent.EndDate
	}

	StoredEvent.UpdatedAt = time.Now()

	if err := s.EventRepo.UpdateEvent(&StoredEvent); err != nil {
		return nil, err
	}
	return &StoredEvent, nil
}

func (s *EventService) DeleteEventBySlug(Slug string) error {
	slug := strings.ToLower(Slug)
	exists, err := s.EventRepo.ExistsEventBySlug(slug)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("event not found")
	}

	if err := s.EventRepo.DeleteEventBySlug(slug); err != nil {
		return err
	}
	return nil
}

func (s *EventService) GetUserByID(userID string) (models.User, error) {
	user, err := s.EventRepo.GetUserByID(userID)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (s *EventService) RegisterToEvent(userID string, Slug string) error {
	slug := strings.ToLower(Slug)

	user, err := s.EventRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	_, err = s.EventRepo.GetUserEventRegistration(user, event)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err := s.EventRepo.RegisterToEvent(user, event); err != nil {
		return err
	}
	return nil
}

func (s *EventService) UnregisterToEvent(userID string, Slug string) error {
	slug := strings.ToLower(Slug)

	user, err := s.EventRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	event, err := s.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	registration, err := s.EventRepo.GetUserEventRegistration(user, event)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user already not registered")
		}
		return err
	}

	if registration.HasPaid {
		return errors.New("user can't unregister after payment")
	}

	if err := s.EventRepo.UnregisterToEvent(registration); err != nil {
		return err
	}
	return nil
}

func (s *EventService) GetEventAttendeesBySlug(slug string) (*[]models.User, error) {
	slug = strings.ToLower(slug)
	attendees, err := s.EventRepo.GetEventAttendeesBySlug(slug)
	if err != nil {
		return nil, err
	}
	return attendees, nil
}

func (s *EventService) IsUserRegistered(userID string, slug string) (bool, error) {
	slug = strings.ToLower(slug)
	registered, err := s.EventRepo.IsUserRegistered(userID, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return registered, nil
}

func (s *EventService) IsAdminOf(userID string, slug string) (bool, *models.AdminStatus, error) {
	slug = strings.ToLower(slug)
	adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(userID, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	} else if adminStatus == nil {
		return false, nil, nil
	}
	return true, adminStatus, nil
}

func (s *EventService) IsAdminTypeOf(user models.User, adminType models.AdminType, slug string) (bool, error) {
	// Master users are considered to be every type of admin
	if user.IsMasterUser {
		return true, nil
	}

	slug = strings.ToLower(slug)
	adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(user.ID, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	if adminStatus != nil && adminStatus.AdminType == adminType {
		return true, nil
	}

	return false, nil
}

func (s *EventService) PromoteUserOfEventBySlug(email, requesterID, slug string) error {
	requester, err := s.EventRepo.GetUserByID(requesterID)
	if err != nil {
		return err
	}
	user, err := s.EventRepo.GetUserByEmail(email)
	if err != nil {
		return err
	}

	if requester.ID == user.ID {
		return errors.New("user cannot alter his own admin status")
	}

	if user.IsMasterUser {
		return errors.New("can't change a master user admin status")
	}

	slug = strings.ToLower(slug)

	status, err := s.IsUserRegistered(user.ID, slug)
	if err != nil || !status {
		if !status {
			return errors.New("user not registered to event")
		}
		return err
	}

	var isRequesterAdmin bool
	var requesterAdminStatus *models.AdminStatus
	if !requester.IsMasterUser {
		if requesterAdminStatus, err = s.EventRepo.GetUserAdminStatusBySlug(requester.ID, slug); err != nil || requesterAdminStatus == nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("requester is not an admin")
			} else if err != nil {
				return err
			} else if requesterAdminStatus == nil {
				return errors.New("requester is not an admin")
			}
		} else {
			isRequesterAdmin = true
		}
	} else {
		isRequesterAdmin = true
	}

	var isUserAdmin bool
	userAdminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(user.ID, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			isUserAdmin = false
		} else {
			return err
		}
	} else {
		isUserAdmin = true
	}

	if requester.IsMasterUser {
		if isUserAdmin && userAdminStatus.AdminType == models.AdminTypeMaster {
			return errors.New("user is already a master admin of this event")
		} else if isUserAdmin && userAdminStatus.AdminType == models.AdminTypeNormal {
			err = s.EventRepo.PromoteUserOfEventBySlug(user.ID, slug)
			if err != nil {
				return err
			}
			return nil
		} else if !isUserAdmin {
			err = s.EventRepo.MakeAdminOfEventBySlug(user.ID, slug)
			if err != nil {
				return err
			}
			return nil
		}
	}

	if isRequesterAdmin && requesterAdminStatus.AdminType == models.AdminTypeMaster {
		if isUserAdmin && userAdminStatus.AdminType == models.AdminTypeMaster {
			return errors.New("master admins can't change another master admin status")
		} else if isUserAdmin && userAdminStatus.AdminType == models.AdminTypeNormal {
			return errors.New("master admins can't promote admins to master admins")
		} else if !isUserAdmin {
			err = s.EventRepo.MakeAdminOfEventBySlug(user.ID, slug)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("requester is not a master admin or master user")
}

func (s *EventService) DemoteUserOfEventBySlug(email, requesterID, slug string) error {
	requester, err := s.EventRepo.GetUserByID(requesterID)
	if err != nil {
		return err
	}
	user, err := s.EventRepo.GetUserByEmail(email)
	if err != nil {
		return err
	}

	if requester.ID == user.ID {
		return errors.New("user cannot alter his own admin status")
	}

	if user.IsMasterUser {
		return errors.New("can't change a master user admin status")
	}

	slug = strings.ToLower(slug)

	var isRequesterAdmin bool
	var requesterAdminStatus *models.AdminStatus
	if !requester.IsMasterUser {
		if requesterAdminStatus, err = s.EventRepo.GetUserAdminStatusBySlug(requester.ID, slug); err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("requester is not an admin")
			} else {
				return err
			}
		} else {
			isRequesterAdmin = true
		}
	} else {
		isRequesterAdmin = true
	}

	var isUserAdmin bool
	userAdminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(user.ID, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			isUserAdmin = false
		} else {
			return err
		}
	} else {
		isUserAdmin = true
	}

	if requester.IsMasterUser {
		if isUserAdmin && userAdminStatus.AdminType == models.AdminTypeMaster {
			s.EventRepo.DemoteUserOfEventBySlug(user.ID, slug)
			return nil
		} else if isUserAdmin && userAdminStatus.AdminType == models.AdminTypeNormal {
			s.EventRepo.RemoveAdminOfEventBySlug(user.ID, slug)
			return nil
		} else if !isUserAdmin {
			return errors.New("user is not an admin of this event")
		}
	}

	if isRequesterAdmin && requesterAdminStatus.AdminType == models.AdminTypeMaster {
		if isUserAdmin && userAdminStatus.AdminType == models.AdminTypeMaster {
			return errors.New("master admins can't change another master admin status")
		} else if isUserAdmin && userAdminStatus.AdminType == models.AdminTypeNormal {
			s.EventRepo.RemoveAdminOfEventBySlug(user.ID, slug)
			return nil
		} else if !isUserAdmin {
			return errors.New("user is not an admin of this event")
		}
	}

	return errors.New("requester is not a master admin or master user")
}
