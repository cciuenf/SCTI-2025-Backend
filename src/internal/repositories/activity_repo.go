package repos

import (
	"scti/internal/models"
)

func (r *EventRepo) CreateEventActivity(activity *models.Activity) error {
	if err := r.DB.Create(activity).Error; err != nil {
		return err
	}
	return nil
}
