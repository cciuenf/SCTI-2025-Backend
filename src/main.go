package main

import (
	"log"
	"net/http"
	"scti/config"
	"scti/internal/db"
	"scti/internal/handlers"
	mw "scti/internal/middleware"
	repos "scti/internal/repositories"
	"scti/internal/services"

	_ "scti/docs"

	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"gorm.io/gorm"
)

// @title           SCTI 2025 API
// @version         1.0
// @description     API Server for SCTI 2025
// @host            localhost:8080
// @BasePath        /
func main() {
	cfg := config.LoadConfig()
	database := db.Connect(*cfg)
	db.Migrate()

	mux := initializeMux(database, cfg)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // change to localhost:PORT of frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(mux)

	if cfg.PORT == "" {
		cfg.PORT = "8080"
	}
	log.Println("Started server on port: " + cfg.PORT)
	log.Fatal(http.ListenAndServe(":"+cfg.PORT, corsHandler))
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

	// API documentation routes
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:"+cfg.PORT+"/swagger/doc.json")))

	// Authentication routes
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /verify-tokens", authHandler.VerifyJWT)
	mux.Handle("POST /logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /refresh-tokens", authMiddleware(http.HandlerFunc(authHandler.GetRefreshTokens)))
	mux.Handle("POST /revoke-refresh-token", authMiddleware(http.HandlerFunc(authHandler.RevokeRefreshToken)))
	mux.Handle("POST /secure-verify-tokens", authMiddleware(http.HandlerFunc(authHandler.VerifyJWT)))
	mux.Handle("POST /verify-account", authMiddleware(http.HandlerFunc(authHandler.VerifyAccount)))

	// Event routes
	mux.HandleFunc("GET /events", eventHandler.GetAllEvents)
	mux.HandleFunc("GET /events/{slug}", eventHandler.GetEventBySlug)
	mux.HandleFunc("GET /events/{slug}/activities", eventHandler.GetEventBySlugWithActivities)
	mux.Handle("POST /events", authMiddleware(http.HandlerFunc(eventHandler.CreateEvent)))
	mux.Handle("PATCH /events", authMiddleware(http.HandlerFunc(eventHandler.UpdateEvent)))
	mux.Handle("PATCH /events/{slug}", authMiddleware(http.HandlerFunc(eventHandler.UpdateEventBySlug)))
	mux.Handle("DELETE /events/{slug}", authMiddleware(http.HandlerFunc(eventHandler.DeleteEventBySlug)))
	mux.Handle("POST /events/{slug}/attend", authMiddleware(http.HandlerFunc(eventHandler.RegisterToEvent)))
	mux.Handle("POST /events/{slug}/unattend", authMiddleware(http.HandlerFunc(eventHandler.UnregisterFromEvent)))
	mux.Handle("GET /events/{slug}/attendees", authMiddleware(http.HandlerFunc(eventHandler.GetEventAttendeesBySlug)))
	// Event Activity routes
	mux.HandleFunc("GET /activities/{slug}", eventHandler.GetAllActivitiesFromEvent)
	mux.Handle("POST /events/{slug}/activity", authMiddleware(http.HandlerFunc(eventHandler.CreateEventActivity)))
	mux.Handle("POST /events/{slug}/activity/register", authMiddleware(http.HandlerFunc(eventHandler.RegisterUserToActivity)))
	mux.Handle("POST /events/{slug}/activity/unregister", authMiddleware(http.HandlerFunc(eventHandler.UnregisterUserFromActivity)))
	// Standalone activity routes
	mux.Handle("POST /activity", authMiddleware(http.HandlerFunc(eventHandler.CreateEventActivity)))
	mux.Handle("POST /activity/register", authMiddleware(http.HandlerFunc(eventHandler.RegisterUserToActivity)))
	mux.Handle("POST /activity/unregister", authMiddleware(http.HandlerFunc(eventHandler.UnregisterUserFromActivity)))

	// Admin routes
	mux.Handle("POST /events/{slug}/promote", authMiddleware(http.HandlerFunc(eventHandler.PromoteUserOfEventBySlug)))
	mux.Handle("POST /events/{slug}/demote", authMiddleware(http.HandlerFunc(eventHandler.DemoteUserOfEventBySlug)))

	return mux
}
