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

	activity.ID = uuid.New().String()
	activity.EventSlug = slug
	activity.EventID = event.ID

	if activity.Type == models.ActivityPalestra {
		activity.IsMandatory = true
	}

	if activity.EndTime.Before(activity.StartTime) {
		return errors.New("activity end can't be before its start")
	}

	if activity.StartTime.Before(event.StartDate) {
		return errors.New("activity start can't be before its event start")
	}

	activity.IsStandalone = false
	activity.StandaloneSlug = ""

	return s.EventRepo.CreateEventActivity(activity)
}
