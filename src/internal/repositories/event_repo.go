package repos

import (
	"errors"
	"scti/internal/models"

	"gorm.io/gorm"
)

type EventRepo struct {
	DB *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{DB: db}
}

func (r *EventRepo) CreateEvent(event *models.Event) error {
	if err := r.DB.Create(event).Error; err != nil {
		return err
	}
	return nil
}

func (r *EventRepo) GetEventByID(eventID string) (models.Event, error) {
	var event models.Event
	if err := r.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (r *EventRepo) GetAllEvents() ([]models.Event, error) {
	var events []models.Event
	if err := r.DB.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepo) GetEventBySlug(slug string) (models.Event, error) {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (r *EventRepo) GetEventBySlugWithActivities(slug string) (models.Event, error) {
	var event models.Event
	if err := r.DB.Preload("Activities").Where("slug = ?", slug).First(&event).Error; err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (r *EventRepo) ExistsEventBySlug(slug string) (bool, error) {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *EventRepo) ExistsEventByID(eventID string) (bool, error) {
	var event models.Event
	if err := r.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *EventRepo) UpdateEvent(event *models.Event) error {
	if err := r.DB.Save(event).Error; err != nil {
		return err
	}
	return nil
}

func (r *EventRepo) DeleteEventBySlug(slug string) error {
	if err := r.DB.Where("slug = ?", slug).Delete(&models.Event{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *EventRepo) GetUserByID(userID string) (models.User, error) {
	var user models.User
	if err := r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *EventRepo) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *EventRepo) ExistsUserByID(userID string) (bool, error) {
	var user models.User
	if err := r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *EventRepo) GetUserEventRegistration(user models.User, event models.Event) (models.EventUser, error) {
	var eventUser models.EventUser
	err := r.DB.Where(
		"user_id = ? AND event_id = ? AND event_slug = ?",
		user.ID,
		event.ID,
		event.Slug,
	).First(&eventUser).Error
	return eventUser, err
}

func (r *EventRepo) RegisterToEvent(user models.User, event models.Event) error {
	eventUser := models.EventUser{
		UserID:    user.ID,
		EventID:   event.ID,
		EventSlug: event.Slug,
	}
	return r.DB.Create(&eventUser).Error
}

func (r *EventRepo) UnregisterToEvent(registration models.EventUser) error {
	return r.DB.Where(
		"event_id = ? AND event_slug = ? AND user_id = ?",
		registration.EventID,
		registration.EventSlug,
		registration.UserID,
	).Delete(&models.EventUser{}).Error
}

func (r *EventRepo) GetEventAttendeesBySlug(slug string) (*[]models.User, error) {
	var event models.Event
	err := r.DB.Preload("Attendees").
		Where("slug = ?", slug).
		First(&event).Error
	if err != nil {
		return nil, err
	}

	return &event.Attendees, nil
}

func (r *EventRepo) GetAtendeeByIDAndSlug(userID, slug string) (*models.EventUser, error) {
	var eventUser *models.EventUser
	if err := r.DB.Where("user_id = ? and event_slug = ?", userID, slug).First(&eventUser).Error; err != nil {
		return nil, err
	}
	return eventUser, nil
}

func (r *EventRepo) IsUserRegistered(userID string, slug string) (bool, error) {
	var eventUser models.EventUser
	if err := r.DB.Where("user_id = ? AND event_slug = ?", userID, slug).First(&eventUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *EventRepo) GetUserAdminStatusBySlug(userID string, slug string) (*models.AdminStatus, error) {
	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_slug = ?", userID, slug).First(&adminStatus).Error; err != nil {
		return nil, err
	}
	return &adminStatus, nil
}

func (r *EventRepo) MakeAdminOfEventBySlug(userID string, slug string) error {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return err
	}

	adminStatus := models.AdminStatus{
		UserID:    userID,
		EventID:   event.ID,
		EventSlug: event.Slug,
		AdminType: models.AdminTypeNormal,
	}

	if err := r.DB.Create(&adminStatus).Error; err != nil {
		return err
	}

	return nil
}

func (r *EventRepo) PromoteUserOfEventBySlug(userID string, slug string) error {
	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_slug = ?", userID, slug).First(&adminStatus).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("a user that is not an admin can't be promoted")
		}
		return err
	}

	if adminStatus.AdminType == models.AdminTypeMaster {
		return errors.New("user is already a master admin")
	}

	adminStatus.AdminType = models.AdminTypeMaster

	if err := r.DB.Save(&adminStatus).Error; err != nil {
		return err
	}

	return nil
}

func (r *EventRepo) DemoteUserOfEventBySlug(userID, slug string) error {
	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_slug = ?", userID, slug).First(&adminStatus).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("a user that is not an admin can't be demoted")
		}
		return err
	}

	if adminStatus.AdminType == models.AdminTypeNormal {
		return errors.New("user is already a normal admin")
	}

	adminStatus.AdminType = models.AdminTypeNormal

	if err := r.DB.Save(&adminStatus).Error; err != nil {
		return err
	}

	return nil
}

func (r *EventRepo) RemoveAdminOfEventBySlug(userID string, slug string) error {
	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_slug = ?", userID, slug).First(&adminStatus).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user already not an admin")
		}
		return err
	}

	if err := r.DB.Delete(&adminStatus).Error; err != nil {
		return err
	}

	return nil
}
