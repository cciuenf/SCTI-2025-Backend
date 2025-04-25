package services

import (
	"errors"
	"scti/internal/models"
	"strings"

	"github.com/google/uuid"
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

	activity.ID = uuid.New().String()
	activity.EventSlug = slug
	activity.EventID = event.ID
	activity.IsStandalone = false
	activity.StandaloneSlug = ""

	return s.EventRepo.CreateEventActivity(activity)
}

// TODO: Implement check with blocked_by, and check with has_fee
func (s *EventService) RegisterUserToActivity(user models.User, activityID string) error {
	activity, err := s.EventRepo.GetActivityByID(activityID)
	if err != nil {
		return err
	}

	if activity.IsStandalone {
		err := s.EventRepo.RegisterUserToActivity(user, activity, nil)
		if err != nil {
			return err
		}
		return nil
	}

	status, err := s.EventRepo.IsUserRegisteredToEvent(user.ID, activity.EventSlug)
	if err != nil {
		return err
	}
	if !status {
		return errors.New("user is not registered to the event of the activity")
	}

	event, err := s.EventRepo.GetEventBySlug(activity.EventSlug)
	if err != nil {
		return err
	}

	return s.EventRepo.RegisterUserToActivity(user, activity, &event)
}
