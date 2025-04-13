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
	Description string    `gorm:"not null"`
	Location    string    `gorm:"not null"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`
	Redes       string    `gorm:"not null"`

	// Decidir como funciona o pagamento
	// Price    string    `gorm:"not null"`

	Atendees []User `gorm:"many2many:event_users;constraint:OnDelete:CASCADE"`
}

type EventUser struct {
	gorm.Model
	EventID   string `gorm:"type:varchar(36);primaryKey"`
	EventSlug string `gorm:"type:varchar(100);primaryKey"`
	UserID    string `gorm:"type:varchar(36);primaryKey"`
	HasPaid   bool   `gorm:"default:false"`
	PaidAt    time.Time
	Amount    float64
}

func (EventUser) TableName() string {
	return "event_users"
}
