package repos

import (
	"errors"
	"scti/internal/models"
	"slices"

	"gorm.io/gorm"
)

type EventRepo struct {
	DB *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{DB: db}
}

func (r *EventRepo) CreateEvent(event *models.Event) error {
	return r.DB.Create(event).Error
}

func (r *EventRepo) GetEventBySlug(slug string) (*models.Event, error) {
	var event models.Event
	if err := r.DB.Where("slug = ? AND is_hidden = ?", slug, false).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepo) GetAllEvents() ([]models.Event, error) {
	var events []models.Event
	if err := r.DB.Where("is_hidden = ?", false).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepo) GetAllPublicEvents() ([]models.Event, error) {
	var events []models.Event
	if err := r.DB.Where("is_hidden = ? AND is_public = ?", false, true).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepo) UpdateEvent(event *models.Event) error {
	return r.DB.Save(event).Error
}

func (r *EventRepo) DeleteEvent(slug string) error {
	return r.DB.Where("slug = ?", slug).Delete(&models.Event{}).Error
}

func (r *EventRepo) CreateEventRegistration(registration *models.EventRegistration) error {
	return r.DB.Create(registration).Error
}

func (r *EventRepo) DeleteEventRegistration(userID string, eventID string) error {
	return r.DB.Where("user_id = ? AND event_id = ?", userID, eventID).
		Unscoped().
		Delete(&models.EventRegistration{}).Error
}

func (r *EventRepo) IsUserRegisteredToEvent(userID string, slug string) (bool, error) {
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

func (r *EventRepo) GetUserAdminStatusBySlug(userID string, slug string) (*models.AdminStatus, error) {
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

func (r *EventRepo) PromoteUserOfEventBySlug(userID string, slug string) error {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return err
	}

	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_id = ?", userID, event.ID).First(&adminStatus).Error; err != nil {
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

func (r *EventRepo) DemoteUserOfEventBySlug(userID string, slug string) error {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return err
	}

	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_id = ?", userID, event.ID).First(&adminStatus).Error; err != nil {
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

func (r *EventRepo) MakeAdminOfEventBySlug(userID string, slug string) error {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return err
	}

	adminStatus := models.AdminStatus{
		UserID:    userID,
		EventID:   event.ID,
		AdminType: models.AdminTypeNormal,
	}

	if err := r.DB.Create(&adminStatus).Error; err != nil {
		return err
	}

	return nil
}

func (r *EventRepo) RemoveAdminOfEventBySlug(userID string, slug string) error {
	var event models.Event
	if err := r.DB.Where("slug = ?", slug).First(&event).Error; err != nil {
		return err
	}

	var adminStatus models.AdminStatus
	if err := r.DB.Where("user_id = ? AND event_id = ?", userID, event.ID).First(&adminStatus).Error; err != nil {
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

func (r *EventRepo) GetEventsCreatedByUser(userID string) ([]models.Event, error) {
	var events []models.Event
	err := r.DB.Where("created_by = ?", userID).Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepo) GetUserEvents(userID string) ([]models.Event, error) {
	var registrations []models.EventRegistration
	err := r.DB.Where("user_id = ?", userID).Find(&registrations).Error
	if err != nil {
		return nil, err
	}

	var eventIDs []string
	for _, registration := range registrations {
		eventIDs = append(eventIDs, registration.EventID)
	}

	var events []models.Event
	err = r.DB.Where("id IN ?", eventIDs).Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepo) GetEventBoughtProductsIDs(eventID string) ([]string, error) {
	var products []models.Product
	if err := r.DB.Where("event_id = ?", eventID).Find(&products).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if len(products) == 0 {
		return nil, nil
	}

	var productIDs []string
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	var purchases []models.Purchase
	if err := r.DB.Where("product_id IN ?", productIDs).Find(&purchases).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if len(purchases) == 0 {
		return nil, nil
	}

	var purchasedProductsIDs []string
	var verifiedProductsIDs []string
	for _, purchase := range purchases {
		if slices.Contains(verifiedProductsIDs, purchase.ProductID) {
			continue
		}
		verifiedProductsIDs = append(verifiedProductsIDs, purchase.ProductID)
		purchasedProductsIDs = append(purchasedProductsIDs, purchase.ProductID)
	}

	return purchasedProductsIDs, nil
}

func (r *EventRepo) GetAllActivitiesFromEvent(eventID string) ([]models.Activity, error) {
	var activities []models.Activity
	if err := r.DB.Where("event_id = ? AND is_hidden = ?", eventID, false).Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *EventRepo) RegisterUserToActivity(registration *models.ActivityRegistration) error {
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
