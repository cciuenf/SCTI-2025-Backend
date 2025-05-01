package models

import (
	"time"

	"gorm.io/gorm"
)

// Activity represents a scheduled activity that can be part of an event or standalone
type Activity struct {
	ID string `gorm:"type:varchar(36);primaryKey"`

	// Event relationship - nullable for standalone activities
	EventID *string `gorm:"type:varchar(36);index" json:"event_id"`

	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `json:"description"`
	Speaker     string `json:"speaker"`
	Location    string `json:"location"`

	// Changed from int to boolean flags for capacity management
	HasUnlimitedCapacity bool `gorm:"default:false" json:"has_unlimited_capacity"` // Whether activity has unlimited capacity
	MaxCapacity          int  `gorm:"default:30" json:"max_capacity"`              // Max capacity when HasUnlimitedCapacity is false

	Type ActivityType `gorm:"not null" json:"type"`

	StartTime time.Time `gorm:"not null" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`

	// Access control
	IsMandatory bool `gorm:"default:false" json:"is_mandatory"` // If users need to be registered automatically
	HasFee      bool `gorm:"default:false" json:"has_fee"`      // If a token is required for this activity

	// Visibility and blocking
	IsHidden  bool `gorm:"default:false" json:"is_hidden"`  // Whether the activity is hidden from search/listings
	IsBlocked bool `gorm:"default:false" json:"is_blocked"` // Whether the activity is blocked from interactions

	// Standalone properties
	IsStandalone   bool   `gorm:"default:false" json:"is_standalone"`                           // If it can be registered for independently
	StandaloneSlug string `gorm:"type:varchar(100);unique;default:null" json:"standalone_slug"` // Used when standalone

	// Relationships
	Registrants []User `gorm:"many2many:activity_registrations;constraint:OnDelete:CASCADE" json:"-"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type ActivityType string

const (
	ActivityPalestra      ActivityType = "palestra"
	ActivityMiniCurso     ActivityType = "mini-curso"
	ActivityVisitaTecnica ActivityType = "visita-tecnica"
)

type AccessMethod string

const (
	AccessMethodDirect AccessMethod = "direct"
	AccessMethodEvent  AccessMethod = "event"
)

// ----------------- Request and Response Models ----------------- //

type ActivityRegistrationRequest struct {
	ActivityID string `json:"activity_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID     string `json:"user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"` // Optional, used for admin actions on other users
}

type ActivityUpdateRequest struct {
	ActivityID string                `json:"activity_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Activity   CreateActivityRequest `json:"activity"`
}

type CreateActivityRequest struct {
	Name                 string       `json:"name" example:"Workshop de Go"`
	Description          string       `json:"description" example:"Workshop introdutório sobre a linguagem Go"`
	Speaker              string       `json:"speaker" example:"John Doe"`
	Location             string       `json:"location" example:"Sala 101"`
	Type                 ActivityType `json:"type" example:"palestra"`
	StartTime            time.Time    `json:"start_time" example:"2024-10-15T14:00:00Z"`
	EndTime              time.Time    `json:"end_time" example:"2024-10-15T16:00:00Z"`
	HasUnlimitedCapacity bool         `json:"has_unlimited_capacity" example:"false"`
	MaxCapacity          int          `json:"max_capacity" example:"30"`
	IsMandatory          bool         `json:"is_mandatory" example:"false"`
	HasFee               bool         `json:"has_fee" example:"false"`
	IsStandalone         bool         `json:"is_standalone" example:"false"`
	StandaloneSlug       string       `json:"standalone_slug" example:"workshop-go-2024"`
	IsHidden             bool         `json:"is_hidden" example:"false"`
	IsBlocked            bool         `json:"is_blocked" example:"false"`
}

type ActivityDeleteRequest struct {
	ActivityID string `json:"activity_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type GetAttendeesRequest struct {
	ID string `json:"id" example:"18d03d08-267b-4b27-b5bc-e423e2489202"`
}
