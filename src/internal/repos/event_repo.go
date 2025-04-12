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

func (r *EventRepo) UpdateEvent(event *models.Event) error {
	if err := r.DB.Save(event).Error; err != nil {
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
