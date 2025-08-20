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
		&models.Product{},
		&models.Purchase{},
		&models.UserProduct{},
		&models.UserToken{},
		&models.ProductBundle{},
		&models.AccessTarget{},
		&models.PixPurchase{},
	)
	if err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Println("database migrated successfully")
}
