package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("Error creating logs directory: %v\n", err)
	}

	loggingMux := mw.WithLogging(mux, logsDir)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // change to localhost:PORT of frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(loggingMux)

	if cfg.PORT == "" {
		cfg.PORT = "8080"
	}
	log.Println("Started server on port: " + cfg.PORT)
	log.Println("Request logs will be saved to: " + filepath.Join(logsDir))
	log.Fatal(http.ListenAndServe(":"+cfg.PORT, corsHandler))
}

func initializeMux(database *gorm.DB, cfg *config.Config) *http.ServeMux {
	authRepo := repos.NewAuthRepo(database)
	eventRepo := repos.NewEventRepo(database)
	activityRepo := repos.NewActivityRepo(database)
	productRepo := repos.NewProductRepo(database)

	authRepo.CreateSuperUser()

	authService := services.NewAuthService(authRepo, cfg.JWT_SECRET)
	eventService := services.NewEventService(eventRepo)
	activityService := services.NewActivityService(activityRepo)
	productService := services.NewProductService(productRepo)

	authHandler := handlers.NewAuthHandler(authService)
	eventHandler := handlers.NewEventHandler(eventService)
	activityHandler := handlers.NewActivityHandler(activityService)
	productHandler := handlers.NewProductHandler(productService)

	authMiddleware := mw.AuthMiddleware(authService)

	mux := http.NewServeMux()

	// API documentation routes
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:"+cfg.PORT+"/swagger/doc.json")))

	// Authentication routes
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /verify-tokens", authHandler.VerifyJWT)
	mux.HandleFunc("POST /forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /change-password", authHandler.ChangePassword)
	mux.Handle("POST /logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /refresh-tokens", authMiddleware(http.HandlerFunc(authHandler.GetRefreshTokens)))
	mux.Handle("POST /revoke-refresh-token", authMiddleware(http.HandlerFunc(authHandler.RevokeRefreshToken)))
	mux.Handle("POST /secure-verify-tokens", authMiddleware(http.HandlerFunc(authHandler.VerifyJWT)))
	mux.Handle("POST /verify-account", authMiddleware(http.HandlerFunc(authHandler.VerifyAccount)))
	mux.Handle("POST /switch-event-creator-status", authMiddleware(http.HandlerFunc(authHandler.SwitchEventCreatorStatus)))

	// Event routes
	mux.HandleFunc("GET /events/{slug}", eventHandler.GetEvent)
	mux.HandleFunc("GET /events", eventHandler.GetAllEvents)
	mux.HandleFunc("GET /events/public", eventHandler.GetAllPublicEvents)
	mux.Handle("GET /events/created", authMiddleware(http.HandlerFunc(eventHandler.GetEventsCreatedByUser)))
	mux.Handle("POST /events", authMiddleware(http.HandlerFunc(eventHandler.CreateEvent)))
	mux.Handle("PATCH /events/{slug}", authMiddleware(http.HandlerFunc(eventHandler.UpdateEvent)))
	mux.Handle("DELETE /events/{slug}", authMiddleware(http.HandlerFunc(eventHandler.DeleteEvent)))
	mux.Handle("POST /events/{slug}/register", authMiddleware(http.HandlerFunc(eventHandler.RegisterToEvent)))
	mux.Handle("POST /events/{slug}/unregister", authMiddleware(http.HandlerFunc(eventHandler.UnregisterFromEvent)))
	mux.Handle("POST /events/{slug}/promote", authMiddleware(http.HandlerFunc(eventHandler.PromoteUserOfEventBySlug)))
	mux.Handle("POST /events/{slug}/demote", authMiddleware(http.HandlerFunc(eventHandler.DemoteUserOfEventBySlug)))

	// Event Activity routes accessed by event slug
	mux.HandleFunc("GET /events/{slug}/activities", activityHandler.GetAllActivitiesFromEvent)
	mux.Handle("POST /events/{slug}/activity", authMiddleware(http.HandlerFunc(activityHandler.CreateEventActivity)))
	mux.Handle("PATCH /events/{slug}/activity", authMiddleware(http.HandlerFunc(activityHandler.UpdateEventActivity)))
	mux.Handle("DELETE /events/{slug}/activity", authMiddleware(http.HandlerFunc(activityHandler.DeleteEventActivity)))
	mux.Handle("POST /events/{slug}/activity/register", authMiddleware(http.HandlerFunc(activityHandler.RegisterUserToActivity)))
	mux.Handle("POST /events/{slug}/activity/unregister", authMiddleware(http.HandlerFunc(activityHandler.UnregisterUserFromActivity)))
	mux.Handle("GET /events/{slug}/activity/registrations", authMiddleware(http.HandlerFunc(activityHandler.GetActivityRegistrations)))
	mux.Handle("POST /events/{slug}/activity/attend", authMiddleware(http.HandlerFunc(activityHandler.AttendActivity)))                                      // Only for admins to mark attendance
	mux.Handle("POST /events/{slug}/activity/unattend", authMiddleware(http.HandlerFunc(activityHandler.UnattendActivity)))                                  // Only for master admins and above to mark unattendance
	mux.Handle("POST /events/{slug}/activity/register-standalone", authMiddleware(http.HandlerFunc(activityHandler.RegisterUserToStandaloneActivity)))       // Only if the user is not registered to the event that contains the activity
	mux.Handle("POST /events/{slug}/activity/unregister-standalone", authMiddleware(http.HandlerFunc(activityHandler.UnregisterUserFromStandaloneActivity))) // Only if the user is not registered to the event that contains the activity

	// Event Product routes accessed by event slug
	mux.Handle("POST /events/{slug}/product", authMiddleware(http.HandlerFunc(productHandler.CreateEventProduct)))
	mux.Handle("PATCH /events/{slug}/product", authMiddleware(http.HandlerFunc(productHandler.UpdateEventProduct)))
	mux.Handle("DELETE /events/{slug}/product", authMiddleware(http.HandlerFunc(productHandler.DeleteEventProduct)))
	mux.Handle("GET /events/{slug}/products", authMiddleware(http.HandlerFunc(productHandler.GetAllProductsFromEvent)))
	mux.Handle("POST /events/{slug}/purchase", authMiddleware(http.HandlerFunc(productHandler.PurchaseProducts)))
	mux.Handle("POST /events/{slug}/try-purchase", authMiddleware(http.HandlerFunc(productHandler.TryPurchaseProducts)))
	return mux
}
