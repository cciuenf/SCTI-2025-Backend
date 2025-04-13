package repos

import (
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

func (r *EventRepo) RegisterToEvent(userID string, slug string) error {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return err
	}

	var user models.User
	if err := r.DB.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	if err := r.DB.Model(&event).Association("Atendees").Append(&user); err != nil {
		return err
	}

	return nil
}

func (r *EventRepo) UnregisterToEvent(userID string, slug string) error {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return err
	}

	var user models.User
	if err := r.DB.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	if err := r.DB.Model(&event).Association("Atendees").Delete(&user); err != nil {
		return err
	}

	return nil
}

func (r *EventRepo) GetEventAtendeesBySlug(slug string) (*[]models.EventUser, error) {
	var eventUsers []models.EventUser
	if err := r.DB.Where("event_slug = ?", slug).Find(&eventUsers).Error; err != nil {
		return nil, err
	}

	return &eventUsers, nil
}
