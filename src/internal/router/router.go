package router

import (
	"log"
	"net/http"
	"os"
	"scti/config"
	"scti/internal/handlers"
	mw "scti/internal/middleware"
	repos "scti/internal/repositories"
	"scti/internal/services"

	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"gorm.io/gorm"
)

func InitializeMux(database *gorm.DB, cfg *config.Config) http.Handler {
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("Error creating logs directory: %v\n", err)
	}

	authRepo := repos.NewAuthRepo(database)
	eventRepo := repos.NewEventRepo(database)
	activityRepo := repos.NewActivityRepo(database)
	productRepo := repos.NewProductRepo(database)
	userRepo := repos.NewUserRepo(database)

	// FATAL if fails, system can't exist without super user
	// fatals located in DB func
	authRepo.CreateSuperUser()

	authService := services.NewAuthService(authRepo, cfg.JWT_SECRET)
	eventService := services.NewEventService(eventRepo)
	activityService := services.NewActivityService(activityRepo)
	productService := services.NewProductService(productRepo)
	userService := services.NewUserService(userRepo)

	authHandler := handlers.NewAuthHandler(authService)
	eventHandler := handlers.NewEventHandler(eventService)
	activityHandler := handlers.NewActivityHandler(activityService)
	productHandler := handlers.NewProductHandler(productService)
	userHandler := handlers.NewUsersHandler(userService)

	authMiddleware := mw.AuthMiddleware(authService)

	mux := http.NewServeMux()

	// API documentation routes
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:"+cfg.PORT+"/swagger/doc.json")))

	// Users routes
	mux.Handle("POST /users/create-event-creator", authMiddleware(http.HandlerFunc(userHandler.CreateEventCreator)))
	mux.Handle("GET /users/{id}", authMiddleware(http.HandlerFunc(userHandler.GetUserInfoFromID)))
	mux.Handle("POST /users/batch", authMiddleware(http.HandlerFunc(userHandler.GetUserInfoBatched)))

	// Authentication routes
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /verify-tokens", authHandler.VerifyJWT)
	mux.HandleFunc("POST /forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /change-password", authHandler.ChangePassword)
	mux.Handle("POST /change-name", authMiddleware(http.HandlerFunc(authHandler.ChangeUserName)))
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
	mux.Handle("GET /user-events", authMiddleware(http.HandlerFunc(eventHandler.GetUserEvents)))
	mux.Handle("GET /events/created", authMiddleware(http.HandlerFunc(eventHandler.GetEventsCreatedByUser)))
	mux.Handle("GET /user-accesses", authMiddleware(http.HandlerFunc(activityHandler.GetUserAccesses)))
	mux.Handle("GET /events/{slug}/accesses", authMiddleware(http.HandlerFunc(activityHandler.GetUserAccessesFromEvent)))
	mux.Handle("POST /events", authMiddleware(http.HandlerFunc(eventHandler.CreateEvent)))
	mux.Handle("PATCH /events/{slug}", authMiddleware(http.HandlerFunc(eventHandler.UpdateEvent)))
	mux.Handle("DELETE /events/{slug}", authMiddleware(http.HandlerFunc(eventHandler.DeleteEvent)))
	mux.Handle("POST /events/{slug}/register", authMiddleware(http.HandlerFunc(eventHandler.RegisterToEvent)))
	mux.Handle("POST /events/{slug}/unregister", authMiddleware(http.HandlerFunc(eventHandler.UnregisterFromEvent)))
	mux.Handle("POST /events/{slug}/promote", authMiddleware(http.HandlerFunc(eventHandler.PromoteUserOfEventBySlug)))
	mux.Handle("POST /events/{slug}/demote", authMiddleware(http.HandlerFunc(eventHandler.DemoteUserOfEventBySlug)))

	// Event Activity routes accessed by event slug
	mux.HandleFunc("GET /events/{slug}/activities", activityHandler.GetAllActivitiesFromEvent)
	mux.Handle("GET /user-activities", authMiddleware(http.HandlerFunc(activityHandler.GetUserActivities)))
	mux.Handle("GET /events/{slug}/user-activities", authMiddleware(http.HandlerFunc(activityHandler.GetUserActivitiesFromEvent)))
	mux.Handle("POST /events/{slug}/activity", authMiddleware(http.HandlerFunc(activityHandler.CreateEventActivity)))
	mux.Handle("PATCH /events/{slug}/activity", authMiddleware(http.HandlerFunc(activityHandler.UpdateEventActivity)))
	mux.Handle("DELETE /events/{slug}/activity", authMiddleware(http.HandlerFunc(activityHandler.DeleteEventActivity)))
	mux.Handle("POST /events/{slug}/activity/register", authMiddleware(http.HandlerFunc(activityHandler.RegisterUserToActivity)))
	mux.Handle("POST /events/{slug}/activity/unregister", authMiddleware(http.HandlerFunc(activityHandler.UnregisterUserFromActivity)))
	mux.Handle("GET /events/{slug}/activity/registrations/{id}", authMiddleware(http.HandlerFunc(activityHandler.GetActivityRegistrations)))
	mux.Handle("POST /events/{slug}/activity/attend", authMiddleware(http.HandlerFunc(activityHandler.AttendActivity)))                                      // Only for admins to mark attendance
	mux.Handle("POST /events/{slug}/activity/unattend", authMiddleware(http.HandlerFunc(activityHandler.UnattendActivity)))                                  // Only for master admins and above to mark unattendance
	mux.Handle("POST /events/{slug}/activity/register-standalone", authMiddleware(http.HandlerFunc(activityHandler.RegisterUserToStandaloneActivity)))       // Only if the user is not registered to the event that contains the activity
	mux.Handle("POST /events/{slug}/activity/unregister-standalone", authMiddleware(http.HandlerFunc(activityHandler.UnregisterUserFromStandaloneActivity))) // Only if the user is not registered to the event that contains the activity
	mux.Handle("GET /events/{slug}/activity/attendants/{id}", authMiddleware(http.HandlerFunc(activityHandler.GetActivityAttendants)))

	// Event Product routes accessed by event slug
	mux.Handle("POST /events/{slug}/product", authMiddleware(http.HandlerFunc(productHandler.CreateEventProduct)))
	mux.Handle("PATCH /events/{slug}/product", authMiddleware(http.HandlerFunc(productHandler.UpdateEventProduct)))
	mux.Handle("DELETE /events/{slug}/product", authMiddleware(http.HandlerFunc(productHandler.DeleteEventProduct)))
	mux.Handle("GET /events/{slug}/products", authMiddleware(http.HandlerFunc(productHandler.GetAllProductsFromEvent)))
	mux.Handle("POST /events/{slug}/purchase", authMiddleware(http.HandlerFunc(productHandler.PurchaseProducts)))
	mux.Handle("GET /user-products-relation", authMiddleware(http.HandlerFunc(productHandler.GetUserProductsRelation)))
	mux.Handle("GET /user-products", authMiddleware(http.HandlerFunc(productHandler.GetUserProducts)))
	mux.Handle("GET /user-tokens", authMiddleware(http.HandlerFunc(productHandler.GetUserTokens)))
	mux.Handle("GET /user-purchases", authMiddleware(http.HandlerFunc(productHandler.GetUserPurchases)))

	loggingMux := mw.WithLogging(mux, logsDir)
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // change to localhost:PORT of frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Refresh"},
		AllowCredentials: true,
	}).Handler(loggingMux)

	return corsHandler
}
