package models

import (
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID         string `gorm:"type:varchar(36);primaryKey;"`
	Name       string `gorm:"not null"`
	LastName   string `gorm:"not null" json:"last_name"`
	Email      string `gorm:"unique;not null"`
	Password   string `gorm:"not null"`
	IsVerified bool   `json:"is_verified"`

	IsMasterUser bool `json:"is_master_user"`

	// Maybe do these
	// IsUenf  bool   `json:"is_uenf"`
	// Curso   string `json:"curso"`
	// Periodo string `json:"periodo"`

	Events []Event        `gorm:"many2many:event_users;constraint:OnDelete:CASCADE"`
	Tokens []RefreshToken `gorm:"foreignKey:UserID;constrainth:OnDelete:CASCADE"`
}

type AdminType string

const (
	AdminTypeMaster AdminType = "master_admin"
	AdminTypeNormal AdminType = "admin"
)

type AdminStatus struct {
	gorm.Model
	UserID    string    `gorm:"type:varchar(36)"`
	EventID   string    `gorm:"type:varchar(36)"`
	EventSlug string    `gorm:"type:varchar(100)"`
	AdminType AdminType `gorm:"type:varchar(20)"`
}

type UserRegister struct {
	gorm.Model
	Name     string `gorm:"not null"`
	LastName string `gorm:"not null" json:"last_name"`
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}

type UserLogin struct {
	gorm.Model
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}

type RefreshToken struct {
	gorm.Model
	UserID   string `gorm:"type:varchar(36);" json:"user_id"`
	TokenStr string `gorm:"type:varchar(1024);" json:"token_str"`
}

type UserClaims struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	LastName    string `json:"last_name"`
	IsVerified  bool   `json:"is_verified"`
	AdminStatus string `json:"admin_status"`
	jwt.RegisteredClaims
}
