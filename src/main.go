package main

import (
	"log"
	"net/http"
	"scti/config"
	"scti/internal/db"
	"scti/internal/handlers"
	"scti/internal/repos"
	"scti/internal/services"

	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	database := db.Connect(*cfg)
	db.Migrate()

	mux := initializeMux(database, cfg)
	if cfg.PORT == "" {
		cfg.PORT = "8080"
	}
	log.Println("Started server on port 8080")
	log.Fatal(http.ListenAndServe(":"+cfg.PORT, mux))
}

func initializeMux(database *gorm.DB, cfg *config.Config) *http.ServeMux {
	authRepo := repos.NewAuthRepo(database)

	authService := services.NewAuthService(authRepo, cfg.JWT_SECRET)
	authHandler := handlers.NewAuthHandler(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/verify-tokens", authHandler.VerifyJWT)

	return mux
}
