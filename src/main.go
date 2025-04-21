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
	log.Println("Started server on port: " + cfg.PORT)
	log.Fatal(http.ListenAndServe(":"+cfg.PORT, mux))
}

func initializeMux(database *gorm.DB, cfg *config.Config) *http.ServeMux {
	authRepo := repos.NewAuthRepo(database)
	eventRepo := repos.NewEventRepo(database)

	authRepo.CreateMasterUser()

	authService := services.NewAuthService(authRepo, cfg.JWT_SECRET)
	eventService := services.NewEventService(eventRepo)

	authHandler := handlers.NewAuthHandler(authService)
	eventHandler := handlers.NewEventHandler(eventService)

	authMiddleware := mw.AuthMiddleware(authService)

	mux := http.NewServeMux()

	// Authentication routes
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /verify-tokens", authHandler.VerifyJWT)
	mux.Handle("POST /logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /refresh-tokens", authMiddleware(http.HandlerFunc(authHandler.GetRefreshTokens)))
	mux.Handle("POST /revoke-refresh-token", authMiddleware(http.HandlerFunc(authHandler.RevokeRefreshToken)))
	mux.Handle("POST /secure-verify-tokens", authMiddleware(http.HandlerFunc(authHandler.VerifyJWT)))

	// Event routes
	mux.HandleFunc("GET /events", eventHandler.GetAllEvents)
	mux.HandleFunc("GET /events/{slug}", eventHandler.GetEventBySlug)
	mux.Handle("POST /events", authMiddleware(http.HandlerFunc(eventHandler.CreateEvent)))
	mux.Handle("PATCH /events", authMiddleware(http.HandlerFunc(eventHandler.UpdateEvent)))
	mux.Handle("PATCH /events/{slug}", authMiddleware(http.HandlerFunc(eventHandler.UpdateEventBySlug)))
	mux.Handle("DELETE /events/{slug}", authMiddleware(http.HandlerFunc(eventHandler.DeleteEventBySlug)))
	mux.Handle("POST /events/{slug}/attend", authMiddleware(http.HandlerFunc(eventHandler.RegisterToEvent)))
	mux.Handle("POST /events/{slug}/unattend", authMiddleware(http.HandlerFunc(eventHandler.UnregisterToEvent)))
	mux.Handle("GET /events/{slug}/attendees", authMiddleware(http.HandlerFunc(eventHandler.GetEventAtendeesBySlug)))

	// Admin routes
	mux.Handle("POST /events/{slug}/promote", authMiddleware(http.HandlerFunc(eventHandler.PromoteUserOfEventBySlug)))
	mux.Handle("POST /events/{slug}/demote", authMiddleware(http.HandlerFunc(eventHandler.DemoteUserOfEventBySlug)))

	return mux
}
