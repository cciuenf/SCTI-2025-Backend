package models

type User struct {
	UserID   string  `gorm:"primaryKey"`
	Nome     string  `gorm:"not null"`
	Email    string  `gorm:"unique;not null"`
	Password string  `gorm:"not null"`
	Idade    int     `gorm:"default:32"`
	Altura   float32 `gorm:"not null"`
	Salario  float32 `gorm:"not null"`
}
