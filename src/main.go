package main

import (
	"log"
	"net/http"
	"scti/config"
	"scti/internal/db"
	"scti/internal/repos"
	"scti/internal/services"
	"scti/internal/handlers"

	"gorm.io/gorm"
)

func main (){
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

func initializeMux(database *gorm.DB, _ *config.Config) *http.ServeMux{
 userRepo := repos.NewUserRepo(database)
  userService := services.NewUserService(userRepo)
  userHandler := handlers.NewUserHandler(userService)

  mux := http.NewServeMux()

  // mux.HandleFunc("GET /me", userHandler.Me)
  mux.HandleFunc("POST /create", userHandler.Create)

  return mux
}
