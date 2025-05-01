package repos

import (
	"errors"
	"scti/internal/models"
	"time"

	"gorm.io/gorm"
)

type ActivityRepo struct {
	DB *gorm.DB
}

func NewActivityRepo(db *gorm.DB) *ActivityRepo {
	return &ActivityRepo{DB: db}
}

func (r *ActivityRepo) CreateActivity(activity *models.Activity) error {
	return r.DB.Create(activity).Error
}

func (r *ActivityRepo) GetActivityByID(id string) (*models.Activity, error) {
	var activity models.Activity
	if err := r.DB.Where("id = ? AND is_hidden = ?", id, false).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepo) GetActivityByStandaloneSlug(slug string) (*models.Activity, error) {
	var activity models.Activity
	if err := r.DB.Where("standalone_slug = ? AND is_standalone = ? AND is_hidden = ?", slug, true, false).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepo) GetAllActivitiesFromEvent(eventID string) ([]models.Activity, error) {
	var activities []models.Activity
	if err := r.DB.Where("event_id = ? AND is_hidden = ?", eventID, false).Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *ActivityRepo) UpdateActivity(activity *models.Activity) error {
	return r.DB.Save(activity).Error
}

func (r *ActivityRepo) DeleteActivity(id string) error {
	return r.DB.Where("id = ?", id).Delete(&models.Activity{}).Error
}

func (r *ActivityRepo) RegisterUserToActivity(registration *models.ActivityRegistration) error {
	var count int64
	err := r.DB.Model(&models.ActivityRegistration{}).
		Where("activity_id = ? AND user_id = ?", registration.ActivityID, registration.UserID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("user already registered to this activity")
	}

	return r.DB.Create(registration).Error
}

func (r *ActivityRepo) UnregisterUserFromActivity(activityID, userID string) error {
	return r.DB.Where("activity_id = ? AND user_id = ?", activityID, userID).
		Unscoped().
		Delete(&models.ActivityRegistration{}).Error
}

func (r *ActivityRepo) IsUserRegisteredToActivity(activityID, userID string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.ActivityRegistration{}).
		Where("activity_id = ? AND user_id = ?", activityID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ActivityRepo) SetUserAttendance(activityID, userID string, attended bool) error {
	var registration models.ActivityRegistration
	err := r.DB.Where("activity_id = ? AND user_id = ?", activityID, userID).
		First(&registration).Error

	if err != nil {
		return err
	}

	if attended {
		now := time.Now()
		registration.AttendedAt = &now
	} else {
		registration.AttendedAt = nil
	}

	return r.DB.Save(&registration).Error
}

func (r *ActivityRepo) GetActivityCapacity(activityID string) (int, int, error) {
	var activity models.Activity
	if err := r.DB.First(&activity, "id = ?", activityID).Error; err != nil {
		return 0, 0, err
	}

	var count int64
	if err := r.DB.Model(&models.ActivityRegistration{}).
		Where("activity_id = ?", activityID).
		Count(&count).Error; err != nil {
		return 0, 0, err
	}

	return int(count), activity.MaxCapacity, nil
}

func (r *ActivityRepo) IsEventBlocked(eventID string) (bool, error) {
	var event models.Event
	if err := r.DB.Select("is_blocked").Where("id = ?", eventID).First(&event).Error; err != nil {
		return false, err
	}
	return event.IsBlocked, nil
}

func (r *ActivityRepo) IsActivityBlocked(activityID string) (bool, error) {
	var activity models.Activity
	if err := r.DB.Select("is_blocked").Where("id = ?", activityID).First(&activity).Error; err != nil {
		return false, err
	}
	return activity.IsBlocked, nil
}

func (r *ActivityRepo) HasUserEventRegistration(userID, eventID string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.EventRegistration{}).
		Where("user_id = ? AND event_id = ?", userID, eventID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ActivityRepo) GetEventByActivityID(activityID string) (*models.Event, error) {
	var activity models.Activity
	if err := r.DB.Select("event_id").Where("id = ?", activityID).First(&activity).Error; err != nil {
		return nil, err
	}

	if activity.EventID == nil {
		return nil, errors.New("activity does not belong to any event")
	}

	var event models.Event
	if err := r.DB.Where("id = ?", *activity.EventID).First(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *ActivityRepo) GetUserByID(userID string) (models.User, error) {
	var user models.User
	err := r.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *ActivityRepo) GetEventBySlug(slug string) (*models.Event, error) {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *ActivityRepo) GetUserAdminStatusBySlug(userID string, slug string) (*models.AdminStatus, error) {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return nil, err
	}

	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_id = ?", userID, event.ID).First(&adminStatus).Error; err != nil {
		return nil, err
	}

	return &adminStatus, nil
}

func (r *ActivityRepo) IsUserRegisteredToEvent(userID string, slug string) (bool, error) {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return false, err
	}

	var count int64
	err := r.DB.Model(&models.EventRegistration{}).
		Where("user_id = ? AND event_id = ?", userID, event.ID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ActivityRepo) GetActivityAttendees(activityID string) ([]models.ActivityRegistration, error) {
  var attendees []models.ActivityRegistration

  err := r.DB.Where("activity_id = ?", activityID).Find(&attendees).Error
  if err != nil {
    return nil, err
  }

	return attendees, nil
}
