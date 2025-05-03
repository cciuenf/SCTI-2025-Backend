package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID          string    `gorm:"type:varchar(36);primaryKey"`
	Slug        string    `gorm:"type:varchar(100);unique;not null"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`

	// Admission control
	IsPublic bool `gorm:"default:true" json:"is_public"` // Whether the event is visible to non-registered users

	// Visibility and blocking
	IsHidden  bool `gorm:"default:false" json:"is_hidden"`  // Whether the event is hidden from search/listings
	IsBlocked bool `gorm:"default:false" json:"is_blocked"` // Whether the event is blocked from interactions

	MaxTokensPerUser int `gorm:"default:0" json:"max_tokens_per_user"` // Maximum number of tokens a user can have for this event

	// Relationships
	Activities []Activity `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE" json:"activities"`
	Products   []Product  `gorm:"many2many:event_products;constraint:OnDelete:CASCADE" json:"products"`
	Attendees  []User     `gorm:"many2many:event_registrations;constraint:OnDelete:CASCADE" json:"attendees"`

	CreatedBy string         `gorm:"type:varchar(36)" json:"created_by"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type EventRegistration struct {
	EventID string `gorm:"type:varchar(36);primaryKey" json:"event_id"`
	UserID  string `gorm:"type:varchar(36);primaryKey" json:"user_id"`

	RegisteredAt time.Time `gorm:"autoCreateTime" json:"registered_at"`
	// CheckedInAt  *time.Time `json:"checked_in_at"` // Time of check-in if applicable, null if not checked in yet

	// Product used for registration, I.E a ticket
	ProductID *string `gorm:"type:varchar(36)" json:"product_id"` // Which product granted access

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Event) TableName() string {
	return "events"
}

func (EventRegistration) TableName() string {
	return "event_registrations"
}

// ------------------ Request and Response Models ------------------ //

type CreateEventRequest struct {
	Slug        string    `json:"slug" example:"gws"`
	Name        string    `json:"name" example:"Go Workshop"`
	Description string    `json:"description" example:"Learn Go programming"`
	StartDate   time.Time `json:"start_date" example:"2025-05-01T14:00:00Z"`
	EndDate     time.Time `json:"end_date" example:"2025-05-01T17:00:00Z"`
	Location    string    `json:"location" example:"Room 101"`

	MaxTokensPerUser int `json:"max_tokens_per_user" example:"1"`

	IsHidden  bool `json:"is_hidden" example:"true"`
	IsBlocked bool `json:"is_blocked" example:"false"`
}

type UpdateEventRequest struct {
	Slug        string    `json:"slug" example:"uws"`
	Name        string    `json:"name" example:"Updated Workshop"`
	Description string    `json:"description" example:"Updated workshop description"`
	Location    string    `json:"location" example:"Room 202"`
	StartDate   time.Time `json:"start_date" example:"2030-11-11T00:00:00Z"`
	EndDate     time.Time `json:"end_date" example:"2030-11-11T23:59:59Z"`

	MaxTokensPerUser int `json:"max_tokens_per_user" example:"1"`

	IsHidden  bool `json:"is_hidden" example:"true"`
	IsBlocked bool `json:"is_blocked" example:"false"`
}
