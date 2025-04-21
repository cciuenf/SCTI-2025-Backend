package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	ID          string    `gorm:"type:varchar(36);primaryKey;"`
	Slug        string    `gorm:"type:varchar(100);primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`

	Atendees []User `gorm:"many2many:event_users;constraint:OnDelete:CASCADE" json:"atendees"`
	//Implement safely
	//Admins   []User `gorm:"many2many:event_admins;constraint:OnDelete:CASCADE" json:"admins"`
}

type EventUser struct {
	gorm.Model
	EventID   string    `gorm:"type:varchar(36);primaryKey" json:"event_id"`
	EventSlug string    `gorm:"type:varchar(100);primaryKey" json:"event_slug"`
	UserID    string    `gorm:"type:varchar(36);primaryKey" json:"user_id"`
	HasPaid   bool      `gorm:"default:false" json:"has_paid"`
	PaidAt    time.Time `gorm:"default:NULL" json:"paid_at"`
	Amount    float64   `gorm:"default:0" json:"amount"`
}

func (EventUser) TableName() string {
	return "event_users"
}
