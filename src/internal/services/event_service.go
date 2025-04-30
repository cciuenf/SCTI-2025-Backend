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

func (s *EventService) CreateEvent(user models.User, body models.CreateEventRequest) (*models.Event, error) {
	if !user.IsMasterUser || !user.IsSuperUser {
		return nil, errors.New("only master users can create events")
	}

	var event models.Event
	event.ID = uuid.New().String()
	event.CreatedBy = user.ID

	if body.Slug == "" {
		return nil, errors.New("event slug can't be empty")
	}

	if body.EndDate.Before(body.StartDate) {
		return nil, errors.New("event end can't be before event start")
	}

	event.Name = body.Name
	event.Slug = strings.ToLower(body.Slug)
	event.Description = body.Description
	event.Location = body.Location
	event.StartDate = body.StartDate
	event.EndDate = body.EndDate
	event.IsPublic = true
	event.IsHidden = body.IsHidden
	event.IsBlocked = body.IsBlocked

	err := s.EventRepo.CreateEvent(&event)
	return &event, err
}

func (s *EventService) GetEvent(slug string) (*models.Event, error) {
	return s.EventRepo.GetEventBySlug(slug)
}

func (s *EventService) GetAllEvents() ([]models.Event, error) {
	return s.EventRepo.GetAllEvents()
}

func (s *EventService) UpdateEvent(user models.User, slug string, newData *models.UpdateEventRequest) (*models.Event, error) {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return nil, err
	}

	if !user.IsSuperUser {
		if event.CreatedBy != user.ID {
			return nil, errors.New("event can only be edited by its creator")
		}
	}

	if newData.Slug == "" {
		return nil, errors.New("event slug can't be empty")
	}

	if newData.EndDate.Before(newData.StartDate) {
		return nil, errors.New("event end can't be before event start")
	}

	event.Name = newData.Name
	event.Slug = strings.ToLower(newData.Slug)
	event.Description = newData.Description
	event.Location = newData.Location
	event.StartDate = newData.StartDate
	event.EndDate = newData.EndDate
	event.IsHidden = newData.IsHidden
	event.IsBlocked = newData.IsBlocked

	err = s.EventRepo.UpdateEvent(event)
	return event, err
}

// TODO: Prohibit deletion if any product from the event was bought or any user attended any activity
// This includes if the only bought product is a ticket from outside the event for a standalone activity
func (s *EventService) DeleteEvent(user models.User, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	if !user.IsSuperUser {
		if event.CreatedBy != user.ID {
			return errors.New("only the event creator can delete it")
		}
	}

	// TODO: Prohibit deletion if any product from the event was bought

	// TODO: Prohibit deletion if any user attended any activity from the event

	return s.EventRepo.DeleteEvent(slug)
}

func (s *EventService) RegisterUserToEvent(user models.User, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	if event.IsBlocked {
		return errors.New("event is blocked and not accepting registrations")
	}

	isRegistered, err := s.EventRepo.IsUserRegisteredToEvent(user.ID, slug)
	if err != nil {
		return err
	}
	if isRegistered {
		return errors.New("user already registered to this event")
	}

	registration := models.EventRegistration{
		EventID:      event.ID,
		UserID:       user.ID,
		RegisteredAt: time.Now(),
	}

	return s.EventRepo.CreateEventRegistration(&registration)
}

// TODO: Prohibit unregistration if the user paid for any product from the event
// Also if the user attended any activity from the event
func (s *EventService) UnregisterUserFromEvent(user models.User, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	isRegistered, err := s.EventRepo.IsUserRegisteredToEvent(user.ID, slug)
	if err != nil {
		return err
	}
	if !isRegistered {
		return errors.New("user is not registered to this event")
	}

	// TODO: Prohibit unregistration if the user paid for any product from the event

	// TODO: Prohibit unregistration if the user attended any activity from the event

	return s.EventRepo.DeleteEventRegistration(user.ID, event.ID)
}

func (s *EventService) IsUserRegisteredToEvent(user models.User, slug string) (bool, error) {
	return s.EventRepo.IsUserRegisteredToEvent(user.ID, slug)
}

func (s *EventService) IsAdminTypeOf(user models.User, adminType models.AdminType, slug string) (bool, error) {
	adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(user.ID, slug)
	if err != nil {
		return false, err
	}

	return adminStatus.AdminType == adminType, nil
}

func (s *EventService) PromoteUserOfEventBySlug(requester models.User, email string, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	targetUser, err := s.EventRepo.GetUserByEmail(email)
	if err != nil {
		return errors.New("target user not found: " + err.Error())
	}

	if targetUser.ID == requester.ID {
		return errors.New("users cannot promote themselves")
	}

	isCreator := event.CreatedBy == targetUser.ID
	if isCreator || targetUser.IsSuperUser {
		return errors.New("cannot promote event creator or super user")
	}

	isRegistered, err := s.EventRepo.IsUserRegisteredToEvent(targetUser.ID, slug)
	if err != nil {
		return err
	}
	if !isRegistered {
		return errors.New("user must be registered to the event to be promoted")
	}

	if requester.IsSuperUser || event.CreatedBy == requester.ID {
		adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(targetUser.ID, slug)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if adminStatus == nil {
			return s.EventRepo.MakeAdminOfEventBySlug(targetUser.ID, slug)
		}

		if adminStatus.AdminType == models.AdminTypeNormal {
			return s.EventRepo.PromoteUserOfEventBySlug(targetUser.ID, slug)
		}

		return errors.New("user is already a master admin")
	}

	isMasterAdmin, err := s.IsAdminTypeOf(requester, models.AdminTypeMaster, slug)
	if err != nil {
		return err
	}

	if isMasterAdmin {
		adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(targetUser.ID, slug)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if adminStatus != nil {
			return errors.New("master admins can only promote to normal admin, user already has admin status")
		}

		return s.EventRepo.MakeAdminOfEventBySlug(targetUser.ID, slug)
	}

	return errors.New("only super users, event creators, or master admins can promote users")
}

func (s *EventService) DemoteUserOfEventBySlug(requester models.User, email string, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	targetUser, err := s.EventRepo.GetUserByEmail(email)
	if err != nil {
		return errors.New("target user not found: " + err.Error())
	}

	if targetUser.ID == requester.ID {
		return errors.New("users cannot demote themselves")
	}

	isCreator := event.CreatedBy == targetUser.ID
	if isCreator || targetUser.IsSuperUser {
		return errors.New("cannot demote event creator or super user")
	}

	adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(targetUser.ID, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user is not an admin of this event")
		}
		return err
	}

	if requester.IsSuperUser || event.CreatedBy == requester.ID {
		if adminStatus.AdminType == models.AdminTypeNormal {
			return s.EventRepo.RemoveAdminOfEventBySlug(targetUser.ID, slug)
		}

		return s.EventRepo.DemoteUserOfEventBySlug(targetUser.ID, slug)
	}

	isMasterAdmin, err := s.IsAdminTypeOf(requester, models.AdminTypeMaster, slug)
	if err != nil {
		return err
	}

	if isMasterAdmin {
		if adminStatus.AdminType == models.AdminTypeMaster {
			return errors.New("master admins cannot demote other master admins")
		}

		return s.EventRepo.RemoveAdminOfEventBySlug(targetUser.ID, slug)
	}

	return errors.New("only super users, event creators, or master admins can demote users")
}

func (s *EventService) GetAllPublicEvents() ([]models.Event, error) {
	events, err := s.EventRepo.GetAllEvents()
	if err != nil {
		return nil, err
	}

	publicEvents := make([]models.Event, 0)
	for _, event := range events {
		if event.IsPublic {
			publicEvents = append(publicEvents, event)
		}
	}

	return publicEvents, nil
}

func (s *EventService) GetUserByID(userID string) (models.User, error) {
	return s.EventRepo.GetUserByID(userID)
}
