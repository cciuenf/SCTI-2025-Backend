package models

import (
	"time"

	"gorm.io/gorm"
)

type QRCode struct {
	UserID string `gorm:"type:varchar(36);primaryKey" json:"user_id"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Re-implementing admin system
type AdminType string

const (
	AdminTypeMaster AdminType = "master_admin"
	AdminTypeNormal AdminType = "admin"
)

// AdminStatus represents user admin status for events
type AdminStatus struct {
	gorm.Model
	UserID    string    `gorm:"type:varchar(36)"`
	EventID   string    `gorm:"type:varchar(36)"`
	AdminType AdminType `gorm:"type:varchar(20)"`
}

func (Activity) TableName() string {
	return "activities"
}

func (Product) TableName() string {
	return "products"
}

func (ActivityRegistration) TableName() string {
	return "activity_registrations"
}

func (Purchase) TableName() string {
	return "purchases"
}

func (UserProduct) TableName() string {
	return "user_products"
}

func (ProductBundle) TableName() string {
	return "product_bundles"
}

func (AccessTarget) TableName() string {
	return "access_targets"
}

func (QRCode) TableName() string {
	return "qr_codes"
}

func (UserToken) TableName() string {
	return "user_tokens"
}

func (AdminStatus) TableName() string {
	return "admin_statuses"
}
