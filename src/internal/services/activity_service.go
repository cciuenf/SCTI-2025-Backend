package services

import (
	"errors"
	"scti/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *EventService) CreateEventActivity(activity *models.Activity, Slug string) error {
	slug := strings.ToLower(Slug)
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	if activity.Type != models.ActivityMiniCurso &&
		activity.Type != models.ActivityPalestra &&
		activity.Type != models.ActivityVisitaTecnica {
		return errors.New("unsupported activity type, should be either of {\"mini-curso\", \"palestra\", \"visita-tecnica\"}")
	}

	if activity.MaxCapacity < 0 {
		return errors.New("max capacity can't be negative")
	}

	if activity.Name == "" {
		return errors.New("activity name can't be empty")
	}

	if activity.Type == models.ActivityPalestra {
		activity.IsMandatory = true
	}

	if activity.EndTime.Before(activity.StartTime) {
		return errors.New("activity end can't be before its start")
	}

	if activity.StartTime.Before(event.StartDate) {
		return errors.New("activity start can't be before its event start")
	}

	if activity.StartTime.After(event.EndDate) {
		return errors.New("activity start can't be afetr event end")
	}

	if activity.EndTime.After(event.EndDate) {
		return errors.New("activity end can't be after event end")
	}

	if activity.StartTime.Before(time.Now()) {
		return errors.New("activities can't be created in the past")
	}

	if activity.HasFee && activity.IsMandatory {
		return errors.New("activity can't be mandatory and have a fee")
	}

	activity.ID = uuid.New().String()
	activity.EventSlug = &slug
	activity.EventID = &event.ID

	if activity.IsStandalone && activity.StandaloneSlug == "" {
		return errors.New("an activity thats created as standalone needs a standalone_slug")
	}

	return s.EventRepo.CreateActivity(activity)
}

func (s *EventService) CreateEventlessActivity(activity *models.Activity) error {
	if activity.Type != models.ActivityMiniCurso &&
		activity.Type != models.ActivityPalestra &&
		activity.Type != models.ActivityVisitaTecnica {
		return errors.New("unsupported activity type, should be either of {\"mini-curso\", \"palestra\", \"visita-tecnica\"}")
	}

	if activity.MaxCapacity < 0 {
		return errors.New("max capacity can't be negative")
	}

	if activity.Name == "" {
		return errors.New("activity name can't be empty")
	}

	if activity.StartTime.Before(time.Now()) {
		return errors.New("activities can't be created in the past")
	}

	if activity.EndTime.Before(activity.StartTime) {
		return errors.New("activity end can't be before its start")
	}

	if activity.StandaloneSlug == "" {
		return errors.New("a fully standalone activity needs a standalone_slug")
	}

	if activity.HasFee {
		return errors.New("a fully standalone activity can't have a fee")
	}

	if activity.IsMandatory {
		return errors.New("a fully standalone activity can't be mandatory")
	}

	activity.ID = uuid.New().String()
	activity.EventSlug = nil
	activity.EventID = nil
	activity.IsStandalone = true

	return s.EventRepo.CreateActivity(activity)
}

// TODO: Implement check with blocked_by, and check with has_fee
func (s *EventService) RegisterUserToStandaloneActivity(user models.User, activityID string) error {
	activity, err := s.EventRepo.GetActivityByID(activityID)
	if err != nil {
		return err
	}

	if !activity.IsStandalone {
		return errors.New("this isn't a standalone activity")
	}

	return s.EventRepo.RegisterUserToStandaloneActivity(user, activity)
}

// TODO: Implement check with blocked_by, and check with has_fee if registering through event
func (s *EventService) RegisterUserToActivityFromEvent(user models.User, activityID string) error {
	activity, err := s.EventRepo.GetActivityByID(activityID)
	if err != nil {
		return err
	}

	if activity.EventID == nil || activity.EventSlug == nil {
		return errors.New("this is a fully standalone activity, use the correct endpoint")
	}

	status, err := s.EventRepo.IsUserRegisteredToEvent(user.ID, *activity.EventSlug)
	if err != nil {
		return err
	}
	if !status {
		return errors.New("user is not registered to the event of the activity")
	}

	event, err := s.EventRepo.GetEventBySlug(*activity.EventSlug)
	if err != nil {
		return err
	}

	return s.EventRepo.RegisterUserToActivityFromEvent(user, activity, event)
}

// TODO: don't let paid users from a standalone activity that registered to it outside an event unregister
func (s *EventService) UnregisterUserFromActivity(user models.User, activityID string) error {
	activity, err := s.EventRepo.GetActivityByID(activityID)
	if err != nil {
		return err
	}

	registration, err := s.EventRepo.GetActivityRegistrationByID(activity.ID, user.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user is already unregistered from activity")
		}
		return err
	}

	if registration.HasAttended {
		return errors.New("can't unregister from an activity that user has attended")
	}

	if registration.RegisteredFromEvent {
		status, err := s.EventRepo.IsUserRegisteredToEvent(user.ID, *activity.EventSlug)
		if err != nil {
			return err
		}
		if !status {
			return errors.New("can't unregister from activity if not attending the event")
		}
	}

	return s.EventRepo.UnregisterUserFromActivity(user, activity)
}

func (s *EventService) GetAllEventActivities(Slug string) ([]models.Activity, error) {
	slug := strings.ToLower(Slug)
	return s.EventRepo.GetAllEventActivities(slug)
}
