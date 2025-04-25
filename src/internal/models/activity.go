package models

import (
	"time"

	"gorm.io/gorm"
)

type ActivityType string

const (
	ActivityPalestra      ActivityType = "palestra"
	ActivityMiniCurso     ActivityType = "mini-curso"
	ActivityVisitaTecnica ActivityType = "visita-tecnica"
)

type Activity struct {
	ID string `gorm:"type:varchar(36);primaryKey;"`

	EventID   string `gorm:"type:varchar(36)" json:"event_id"`
	EventSlug string `gorm:"type:varchar(100)" json:"event_slug"`

	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `json:"description"`
	Speaker     string `json:"speaker"`
	Location    string `json:"location"`
	MaxCapacity int    `gorm:"default:0" json:"max_capacity"` // 0 means unlimited

	Type ActivityType `gorm:"not null" json:"type"`

	StartTime time.Time `gorm:"not null" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`

	IsMandatory bool `gorm:"default:false" json:"is_mandatory"` // If the user needs to be registered automatically
	HasFee      bool `gorm:"default:false" json:"has_fee"`      // If the user needs a token or not to enter

	IsStandalone   bool   `gorm:"default:false" json:"is_standalone"`  // If it can be registered to or exist without an event
	StandaloneSlug string `gorm:"varchar(100)" json:"standalone_slug"` // Used as event slug

	Event         Event                  `gorm:"foreignKey:EventID,EventSlug;references:ID,Slug;constraint:OnDelete:CASCADE"`
	Registrations []ActivityRegistration `gorm:"foreignKey:ActivityID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"autoDeleteTime" json:"deleted_at,omitempty"`
}

type ActivityRegistration struct {
	ID string `gorm:"type:varchar(36);primaryKey;"`

	ActivityID string `gorm:"type:varchar(36);not null" json:"activity_id"`
	UserID     string `gorm:"type:varchar(36);not null" json:"user_id"`
	EventID    string `gorm:"type:varchar(36)" json:"event_id"`
	EventSlug  string `gorm:"type:varchar(100)" json:"event_slug"`

	RegisteredAt time.Time  `gorm:"autoCreateTime" json:"registered_at"`
	HasAttended  bool       `gorm:"default:false" json:"has_attended"`
	AttendedAt   *time.Time `json:"attended_at"`

	IsStandaloneRegistration bool `json:"is_standalone_registration"`

	Activity Activity `gorm:"foreignKey:ActivityID;references:ID;constraint:OnDelete:CASCADE"`
	User     User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"autoDeleteTime" json:"deleted_at,omitempty"`
}

func (Activity) TableName() string {
	return "activities"
}

func (ActivityRegistration) TableName() string {
	return "activity_registrations"
}
