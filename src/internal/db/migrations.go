package db

import (
	"log"
	"scti/internal/models"
)

func Migrate() {
	log.Println("running database migrations...")

	err := DB.AutoMigrate(
		&models.User{},
		&models.UserPass{},
		&models.RefreshToken{},
		&models.Event{},
		&models.EventRegistration{},
		&models.AdminStatus{},
		&models.UserVerification{},
		&models.Activity{},
		&models.ActivityRegistration{},
	)
	if err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Println("database migrated successfully")
}
