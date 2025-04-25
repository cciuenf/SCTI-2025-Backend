package repos

import (
	"scti/internal/models"
	"time"
)

func (r *EventRepo) CreateEventActivity(activity *models.Activity) error {
	if err := r.DB.Create(activity).Error; err != nil {
		return err
	}
	return nil
}

func (r *EventRepo) GetActivityByID(id string) (models.Activity, error) {
	var activity models.Activity
	err := r.DB.Where("id = ?", id).First(&activity).Error
	return activity, err
}

// TODO: Make this a transaction when implement the activity token so if any fails the user doesn't lose the token
func (r *EventRepo) RegisterUserToActivity(user models.User, activity models.Activity, event *models.Event) error {
	registration := models.ActivityRegistration{
		UserID:       user.ID,
		ActivityID:   activity.ID,
		RegisteredAt: time.Now(),
		EventID:      "",
		EventSlug:    "",
	}
	if event != nil {
		registration.EventID = event.ID
		registration.EventSlug = event.Slug
	}
	return r.DB.Create(&registration).Error
}
