package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB         string
	DB_NAME    string
	DB_PASS    string
	DB_PORT    string
	DB_USER    string
	DSN        string
	HOST       string
	PORT       string
	JWT_SECRET string
}

var (
	server_host string
	server_port string
	db          string
	db_port     string
	db_user     string
	db_pass     string
	jwtSecret   string
	dsn         string
	systemEmail string
)

func LoadConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load config")
	}

	server_host = os.Getenv("HOST")
	server_port = os.Getenv("PORT")
	db = os.Getenv("DATABASE")
	db_port = os.Getenv("DATABASE_PORT")
	db_user = os.Getenv("DATABASE_USER")
	db_pass = os.Getenv("DATABASE_PASS")
	jwtSecret = os.Getenv("JWT_SECRET")
	systemEmail = os.Getenv("SCTI_EMAIL")

	dsn = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=America/Sao_Paulo", server_host, db_user, db_pass, db, db_port)

	return &Config{
		HOST:       server_host,
		PORT:       server_port,
		DB:         db,
		DB_PORT:    db_port,
		DB_USER:    db_user,
		DB_PASS:    db_pass,
		DSN:        dsn,
		JWT_SECRET: jwtSecret,
	}
}

func GetServerHost() string {
	return server_host
}

func GetServerPort() string {
	return server_port
}

func GetDB() string {
	return db
}

func GetDBPort() string {
	return db_port
}

func GetDBUser() string {
	return db_user
}

func GetDBPass() string {
	return db_pass
}

func GetJWTSecret() string {
	return jwtSecret
}

func GetDSN() string {
	return dsn
}

func GetSystemEmail() string {
	return systemEmail
}
