package db

import (
	"log"
	"os"
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

	testMode := os.Getenv("TEST_MODE") == "true"
	gormCfg := &gorm.Config{}
	if testMode {
		gormCfg.Logger = logger.Default.LogMode(logger.Silent)
	} else {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}

	DB, err = gorm.Open(postgres.Open(cfg.DSN), gormCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to postgres instance")
	return DB
}
