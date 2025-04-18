package db

import (
	"log"
	"scti/internal/models"
)

func Migrate() {
	log.Println("running database migrations...")

	err := DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Event{},
		&models.EventUser{},
	)
	if err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Println("Database migrated successfully")
}
