package main

import (
	"log"
	"net/http"
	"scti/config"
	"scti/internal/db"
	"scti/internal/handlers"
	mw "scti/internal/middleware"
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

	authMiddleware := mw.AuthMiddleware(authService)

	mux := http.NewServeMux()

	// Authentication routes
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /verify-tokens", authHandler.VerifyJWT)
	mux.Handle("POST /logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /refresh-tokens", authMiddleware(http.HandlerFunc(authHandler.GetRefreshTokens)))
	mux.Handle("POST /revoke-refresh-token", authMiddleware(http.HandlerFunc(authHandler.RevokeRefreshToken)))

	return mux
}
