package models

import (
	"time"

	"gorm.io/gorm"
)

type ActivityLevel string

const (
	ActivityEasy   ActivityLevel = "easy"
	ActivityMedium ActivityLevel = "medium"
	ActivityHard   ActivityLevel = "hard"
)

type Activity struct {
	ID string `gorm:"type:varchar(36);primaryKey" example:"550e8400-e29b-41d4-a716-446655440000"`

	EventID string `gorm:"type:varchar(36);index" json:"event_id" example:"550e8400-e29b-41d4-a716-446655440001"`

	Name         string        `gorm:"type:varchar(100);not null" json:"name" example:"Workshop de Go"`
	Description  string        `json:"description" example:"Workshop introdutório sobre a linguagem Go"`
	Speaker      string        `json:"speaker" example:"John Doe"`
	Location     string        `json:"location" example:"Sala 101"`
	Requirements string        `gorm:"type:varchar(1024)" json:"requirements"`
	Level        ActivityLevel `gorm:"not null" json:"level"`

	// Changed from int to boolean flags for capacity management
	HasUnlimitedCapacity bool `gorm:"default:false" json:"has_unlimited_capacity" example:"true"` // Whether activity has unlimited capacity
	MaxCapacity          int  `gorm:"default:30" json:"max_capacity" example:"30"`                // Max capacity when HasUnlimitedCapacity is false

	Type ActivityType `gorm:"not null" json:"type" example:"palestra"`

	StartTime time.Time `gorm:"not null" json:"start_time" example:"2024-10-15T14:00:00Z"`
	EndTime   time.Time `gorm:"not null" json:"end_time" example:"2024-10-15T16:00:00Z"`

	// Access control
	IsMandatory bool `gorm:"default:false" json:"is_mandatory" example:"true"` // If users need to be registered automatically
	HasFee      bool `gorm:"default:false" json:"has_fee" example:"true"`      // If an event ticket or token is required
	NeedsToken  bool `gorm:"default:false" json:"needs_token" example:"true"`  // If a token is required for this activity

	// Visibility and blocking
	IsHidden  bool `gorm:"default:false" json:"is_hidden" example:"false"`  // Whether the activity is hidden from search/listings
	IsBlocked bool `gorm:"default:false" json:"is_blocked" example:"false"` // Whether the activity is blocked from interactions

	// Relationships
	Registrants []User `gorm:"many2many:activity_registrations;constraint:OnDelete:CASCADE" json:"-"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at" example:"2024-10-15T14:00:00Z"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at" example:"2024-10-15T14:00:00Z"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Activity) TableName() string {
	return "activities"
}

type ActivityRegistration struct {
	ActivityID string `gorm:"type:varchar(36);primaryKey" json:"activity_id"`
	UserID     string `gorm:"type:varchar(36);primaryKey" json:"user_id"`

	RegisteredAt time.Time  `gorm:"autoCreateTime" json:"registered_at"`
	AttendedAt   *time.Time `json:"attended_at"` // Time of attendance, null if not attended yet

	// Access method tracking
	AccessMethod string  `gorm:"type:varchar(20)" json:"access_method"` // "event", "product", "token", or "direct"
	ProductID    *string `gorm:"type:varchar(36)" json:"product_id"`    // Which product was used (if applicable)
	TokenID      *string `gorm:"type:varchar(36)" json:"token_id"`      // Which token was used (if applicable)

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (ActivityRegistration) TableName() string {
	return "activity_registrations"
}

type ActivityType string

const (
	ActivityPalestra      ActivityType = "palestra"
	ActivityMiniCurso     ActivityType = "mini-curso"
	ActivityVisitaTecnica ActivityType = "visita-tecnica"
)

type AccessMethod string

const (
	AccessMethodDirect  AccessMethod = "direct"
	AccessMethodEvent   AccessMethod = "event"
	AccessMethodProduct AccessMethod = "product"
	AccessMethodToken   AccessMethod = "token"
)

// ----------------- Request and Response Models ----------------- //

type CreateActivityRequest struct {
	Name                 string        `json:"name" example:"Workshop de Go"`
	Description          string        `json:"description" example:"Workshop introdutório sobre a linguagem Go"`
	Speaker              string        `json:"speaker" example:"John Doe"`
	Location             string        `json:"location" example:"Sala 101"`
	Type                 ActivityType  `json:"type" example:"palestra"`
	StartTime            time.Time     `json:"start_time" example:"2024-10-15T14:00:00Z"`
	EndTime              time.Time     `json:"end_time" example:"2024-10-15T16:00:00Z"`
	HasUnlimitedCapacity bool          `json:"has_unlimited_capacity" example:"false"`
	MaxCapacity          int           `json:"max_capacity" example:"30"`
	IsMandatory          bool          `json:"is_mandatory" example:"false"`
	HasFee               bool          `json:"has_fee" example:"false"`
	IsHidden             bool          `json:"is_hidden" example:"false"`
	IsBlocked            bool          `json:"is_blocked" example:"false"`
	Level                ActivityLevel `json:"level" example:"easy"`
	Requirements         string        `json:"requirements" example:"VSCode e Python 3.12"`
}

type ActivityUpdateRequest struct {
	ActivityID           string        `json:"activity_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name                 string        `json:"name" example:"Workshop de Go"`
	Description          string        `json:"description" example:"Workshop introdutório sobre a linguagem Go"`
	Speaker              string        `json:"speaker" example:"John Doe"`
	Location             string        `json:"location" example:"Sala 101"`
	Type                 ActivityType  `json:"type" example:"palestra"`
	StartTime            time.Time     `json:"start_time" example:"2024-10-15T14:00:00Z"`
	EndTime              time.Time     `json:"end_time" example:"2024-10-15T16:00:00Z"`
	HasUnlimitedCapacity bool          `json:"has_unlimited_capacity" example:"false"`
	MaxCapacity          int           `json:"max_capacity" example:"30"`
	IsMandatory          bool          `json:"is_mandatory" example:"false"`
	HasFee               bool          `json:"has_fee" example:"false"`
	IsHidden             bool          `json:"is_hidden" example:"false"`
	IsBlocked            bool          `json:"is_blocked" example:"false"`
	Level                ActivityLevel `json:"level" example:"easy"`
	Requirements         string        `json:"requirements" example:"VSCode e Python 3.12"`
}

type ActivityRegistrationRequest struct {
	ActivityID string `json:"activity_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID     string `json:"user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"` // Optional, used for admin actions on other users
}

type ActivityDeleteRequest struct {
	ActivityID string `json:"activity_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}
