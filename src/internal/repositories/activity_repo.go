package repos

import (
	"scti/internal/models"
	"time"
)

func (r *EventRepo) CreateActivity(activity *models.Activity) error {
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

func (r *EventRepo) GetActivityRegistrationByID(activityID string, userID string) (*models.ActivityRegistration, error) {
	var registration models.ActivityRegistration
	err := r.DB.Where("user_id = ? AND activity_id = ?", userID, activityID).First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

// TODO: Make this a transaction when implement the activity token so if any fails the user doesn't lose the token
func (r *EventRepo) RegisterUserToStandaloneActivity(user models.User, activity models.Activity) error {
	registration := models.ActivityRegistration{
		UserID:              user.ID,
		ActivityID:          activity.ID,
		RegisteredAt:        time.Now(),
		RegisteredFromEvent: false,
		EventID:             "",
		EventSlug:           "",
	}
	return r.DB.Create(&registration).Error
}

// TODO: Make this a transaction when implement the activity token so if any fails the user doesn't lose the token
func (r *EventRepo) RegisterUserToActivityFromEvent(user models.User, activity models.Activity, event models.Event) error {
	registration := models.ActivityRegistration{
		UserID:              user.ID,
		ActivityID:          activity.ID,
		RegisteredAt:        time.Now(),
		RegisteredFromEvent: true,
		EventID:             event.ID,
		EventSlug:           event.Slug,
	}
	return r.DB.Create(&registration).Error
}

func (r *EventRepo) UnregisterUserFromActivity(user models.User, activity models.Activity) error {
	return r.DB.
		Unscoped().
		Where("user_id = ? AND activity_id = ?", user.ID, activity.ID).
		Delete(&models.ActivityRegistration{}).Error
}

func (r *EventRepo) GetAllEventActivities(slug string) ([]models.Activity, error) {
	var activities []models.Activity
	err := r.DB.Where("event_slug = ?", slug).Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, err
}
