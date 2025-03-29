package config

import (
  "fmt"
  "log"
  "os"
  "github.com/joho/godotenv"
)

type Config struct  {
  DB string
  DB_NAME string
  DB_PASS string
  DB_PORT string
  DB_USER string
  DSN string
  HOST string
  PORT string
}

func LoadConfig () *Config {
  err:= godotenv.Load(".env")
  if err != nil {
    log.Fatalf("Failed to load config")
  }

  host := os.Getenv("HOST")
  port := os.Getenv("PORT")
  db_port := os.Getenv("DATABASE_PORT")
  user := os.Getenv("DATABASE_USER")
  pass := os.Getenv("DATABASE_PASS")
  db := os.Getenv("DATABASE")

  dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=America/Sao_Paulo", host, user, pass, db, db_port)

  return &Config{
    HOST: host,
    PORT: port,
    DB_PORT: db_port,
    DB: db,
    DB_USER: user,
    DB_PASS: pass,
    DSN: dsn,
  }
}
