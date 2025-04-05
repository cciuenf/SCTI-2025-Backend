package models

type User struct {
	UserID   string  `gorm:"type:varchar(36);primaryKey;"`
	Name     string  `gorm:"not null"`
	Email    string  `gorm:"unique;not null"`
	Password string  `gorm:"not null"`
	RefreshToken string `gorm:"type:varchar(36);"`
}

type UserRegister struct {
	Name     string  `gorm:"not null"`
	Email    string  `gorm:"unique;not null"`
	Password string  `gorm:"not null"`
}

type UserLogin struct {
	Email    string  `gorm:"unique;not null"`
	Password string  `gorm:"not null"`
}
