package db

import (
	"log"
	"scti/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg config.Config) *gorm.DB {
	var err error
	if cfg.DSN == "" {
		log.Fatalf("dsn was empty")
	}
	DB, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to postgres instance")
	return DB
}
