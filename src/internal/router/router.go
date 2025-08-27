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
	verifiedOnly := mw.Chain(authMiddleware, mw.IsVerifiedMiddleware())

	mux := http.NewServeMux()

	// API documentation routes
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:"+cfg.PORT+"/swagger/doc.json")))

	// Users routes
	mux.Handle("POST /v1/users/create-event-creator", verifiedOnly(http.HandlerFunc(userHandler.CreateEventCreator)))
	mux.HandleFunc("GET /v1/users/{id}", userHandler.GetUserInfoFromID)
	mux.HandleFunc("POST /v1/users/batch", userHandler.GetUserInfoBatched)

	// Authentication routes
	mux.HandleFunc("POST /v1/register", authHandler.Register)
	mux.HandleFunc("POST /v1/login", authHandler.Login)
	mux.HandleFunc("POST /v1/verify-tokens", authHandler.VerifyJWT)
	mux.HandleFunc("POST /v1/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /v1/change-password", authHandler.ChangePassword)
	mux.Handle("POST /v1/change-name", verifiedOnly(http.HandlerFunc(authHandler.ChangeUserName)))
	mux.Handle("POST /v1/logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /v1/refresh-tokens", authMiddleware(http.HandlerFunc(authHandler.GetRefreshTokens)))
	mux.Handle("POST /v1/revoke-refresh-token", authMiddleware(http.HandlerFunc(authHandler.RevokeRefreshToken)))
	mux.Handle("POST /v1/secure-verify-tokens", authMiddleware(http.HandlerFunc(authHandler.VerifyJWT)))
	mux.Handle("POST /v1/verify-account", authMiddleware(http.HandlerFunc(authHandler.VerifyAccount)))
	mux.Handle("POST /v1/switch-event-creator-status", verifiedOnly(http.HandlerFunc(authHandler.SwitchEventCreatorStatus)))
	mux.Handle("POST /v1/resend-verification-code", authMiddleware(http.HandlerFunc(authHandler.ResendVerificationCode)))
	mux.HandleFunc("POST /v1/force-reauth", authHandler.ForceReAuth)

	// Event routes
	mux.HandleFunc("GET /v1/events/{slug}", eventHandler.GetEvent)
	mux.HandleFunc("GET /v1/events", eventHandler.GetAllEvents)
	mux.HandleFunc("GET /v1/events/public", eventHandler.GetAllPublicEvents)
	mux.Handle("GET /v1/user-events", verifiedOnly(http.HandlerFunc(eventHandler.GetUserEvents)))
	mux.Handle("GET /v1/events/created", verifiedOnly(http.HandlerFunc(eventHandler.GetEventsCreatedByUser)))
	mux.Handle("GET /v1/user-accesses", verifiedOnly(http.HandlerFunc(activityHandler.GetUserAccesses)))
	mux.Handle("GET /v1/events/{slug}/accesses", verifiedOnly(http.HandlerFunc(activityHandler.GetUserAccessesFromEvent)))
	mux.Handle("POST /v1/events", verifiedOnly(http.HandlerFunc(eventHandler.CreateEvent)))
	mux.Handle("PATCH /v1/events/{slug}", verifiedOnly(http.HandlerFunc(eventHandler.UpdateEvent)))
	mux.Handle("DELETE /v1/events/{slug}", verifiedOnly(http.HandlerFunc(eventHandler.DeleteEvent)))
	mux.Handle("POST /v1/events/{slug}/register", verifiedOnly(http.HandlerFunc(eventHandler.RegisterToEvent)))
	mux.Handle("POST /v1/events/{slug}/unregister", verifiedOnly(http.HandlerFunc(eventHandler.UnregisterFromEvent)))
	mux.Handle("POST /v1/events/{slug}/promote", verifiedOnly(http.HandlerFunc(eventHandler.PromoteUserOfEventBySlug)))
	mux.Handle("POST /v1/events/{slug}/demote", verifiedOnly(http.HandlerFunc(eventHandler.DemoteUserOfEventBySlug)))

	// Event Activity routes accessed by event slug
	mux.HandleFunc("GET /v1/events/{slug}/activities", activityHandler.GetAllActivitiesFromEvent)
	mux.Handle("GET /v1/user-activities", verifiedOnly(http.HandlerFunc(activityHandler.GetUserActivities)))
	mux.Handle("GET /v1/user-attended-activities", verifiedOnly(http.HandlerFunc(activityHandler.GetUserAttendedActivities)))
	mux.Handle("GET /v1/events/{slug}/user-activities", verifiedOnly(http.HandlerFunc(activityHandler.GetUserActivitiesFromEvent)))
	mux.Handle("POST /v1/events/{slug}/activity", verifiedOnly(http.HandlerFunc(activityHandler.CreateEventActivity)))
	mux.Handle("PATCH /v1/events/{slug}/activity", verifiedOnly(http.HandlerFunc(activityHandler.UpdateEventActivity)))
	mux.Handle("DELETE /v1/events/{slug}/activity", verifiedOnly(http.HandlerFunc(activityHandler.DeleteEventActivity)))
	mux.Handle("POST /v1/events/{slug}/activity/register", verifiedOnly(http.HandlerFunc(activityHandler.RegisterUserToActivity)))
	mux.Handle("POST /v1/events/{slug}/activity/unregister", verifiedOnly(http.HandlerFunc(activityHandler.UnregisterUserFromActivity)))
	mux.Handle("GET /v1/events/{slug}/activity/registrations/{id}", verifiedOnly(http.HandlerFunc(activityHandler.GetActivityRegistrations)))
	mux.Handle("POST /v1/events/{slug}/activity/attend", verifiedOnly(http.HandlerFunc(activityHandler.AttendActivity)))     // Only for admins to mark attendance
	mux.Handle("POST /v1/events/{slug}/activity/unattend", verifiedOnly(http.HandlerFunc(activityHandler.UnattendActivity))) // Only for master admins and above to mark unattendance
	mux.Handle("GET /v1/events/{slug}/activity/attendants/{id}", verifiedOnly(http.HandlerFunc(activityHandler.GetActivityAttendants)))

	// Event Product routes accessed by event slug
	mux.Handle("POST /v1/events/{slug}/product", verifiedOnly(http.HandlerFunc(productHandler.CreateEventProduct)))
	mux.Handle("PATCH /v1/events/{slug}/product", verifiedOnly(http.HandlerFunc(productHandler.UpdateEventProduct)))
	mux.Handle("DELETE /v1/events/{slug}/product", verifiedOnly(http.HandlerFunc(productHandler.DeleteEventProduct)))
	mux.Handle("GET /v1/events/{slug}/products", authMiddleware(http.HandlerFunc(productHandler.GetAllProductsFromEvent)))
	mux.Handle("POST /v1/events/{slug}/purchase", verifiedOnly(http.HandlerFunc(productHandler.PurchaseProducts)))
	mux.Handle("GET /v1/user-products-relation", verifiedOnly(http.HandlerFunc(productHandler.GetUserProductsRelation)))
	mux.HandleFunc("GET /v1/all-user-products-relation", productHandler.GetAllUserProductsRelation)
	mux.HandleFunc("GET /v1/user-products-global/{id}", productHandler.GetGlobalUserProductsFromID)
	mux.Handle("GET /v1/user-products", verifiedOnly(http.HandlerFunc(productHandler.GetUserProducts)))
	mux.Handle("GET /v1/user-tokens", verifiedOnly(http.HandlerFunc(productHandler.GetUserTokens)))
	mux.Handle("GET /v1/user-purchases", verifiedOnly(http.HandlerFunc(productHandler.GetUserPurchases)))
	mux.Handle("POST /v1/can-gift", verifiedOnly(http.HandlerFunc(productHandler.CanGift)))

	// Payment Only Route
	mux.Handle("POST /v1/events/{slug}/forced-pix", verifiedOnly(http.HandlerFunc(productHandler.ForcedPix)))

	// Webhook routes
	mux.HandleFunc("POST /webhook/mp", productHandler.MPWebhook)

	loggingMux := mw.WithLogging(mux, logsDir)
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // change to localhost:PORT of frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Refresh"},
		AllowCredentials: true,
	}).Handler(loggingMux)

	return corsHandler
}
