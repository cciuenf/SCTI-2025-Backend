package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type User struct {
	ID         string   `gorm:"type:varchar(36);primaryKey;" json:"id"`
	Name       string   `gorm:"not null" json:"name"`
	LastName   string   `gorm:"not null" json:"last_name"`
	Email      string   `gorm:"unique;not null" json:"email"`
	IsVerified bool     `gorm:"default:false" json:"is_verified"`
	UserPass   UserPass `gorm:"foreignKey:ID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"autoDeleteTime" json:"deleted_at,omitempty"`

	IsMasterUser bool `gorm:"default:false" json:"is_master_user"`

	// Maybe do these
	// IsUenf  bool   `json:"is_uenf"`
	// Curso   string `json:"curso"`
	// Periodo string `json:"periodo"`

	UserVerification UserVerification `gorm:"foreignKey:ID;references:ID;constraint:OnDelete:CASCADE"`
	Events           []Event          `gorm:"many2many:event_users;constraint:OnDelete:CASCADE"`
	Tokens           []RefreshToken   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type UserPass struct {
	ID       string `gorm:"type:varchar(36);primaryKey" json:"id"`
	Password string `gorm:"not null" json:"password"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"autoDeleteTime" json:"deleted_at,omitempty"`
}

type UserVerification struct {
	ID                 string `gorm:"type:varchar(36);primaryKey" json:"id"`
	VerificationNumber int    `gorm:"not null" json:"verification_number"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"autoDeleteTime" json:"deleted_at,omitempty"`
}

func (UserPass) TableName() string {
	return "user_pass"
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
