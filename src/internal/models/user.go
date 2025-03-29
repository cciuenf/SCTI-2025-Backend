package models

type User struct {
  UserID int `gorm:"primaryKey;autoIncrement"`
  Nome string `gorm:"not null"`
  Idade int `gorm:"default:32"`
  Altura float32 `gorm:"not null"`
  Salario float32 `gorm:"not null"`
}
