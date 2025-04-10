package models

import (
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       string         `gorm:"type:varchar(36);primaryKey;"`
	Name     string         `gorm:"not null"`
	LastName string         `gorm:"not null" json:"last_name"`
	Event    string         `gorm:"not null"`
	IsPaid   bool           `gorm:"not null" json:"is_paid"`
	Email    string         `gorm:"unique;not null"`
	Password string         `gorm:"not null"`
	Tokens   []RefreshToken `gorm:"foreignKey:UserID;constrainth:OnDelete:CASCADE"`
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
	ID       string `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"last_name"`
	Event    string `json:"event"`
	IsPaid   bool   `json:"is_paid"`
	jwt.RegisteredClaims
}
