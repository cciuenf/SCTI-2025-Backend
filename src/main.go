package main

import (
	"log"
	"net/http"
	"scti/config"
	"scti/internal/db"
	"scti/internal/router"

	_ "scti/docs"
)

// @title           SCTI 2025 API
// @version         1.0
// @description     API Server for SCTI 2025
// @host            localhost:8080
// @BasePath        /
func main() {
	cfg := config.LoadConfig(".env")
	database := db.Connect(*cfg)
	db.Migrate()

	if cfg.PORT == "" {
		cfg.PORT = "8080"
	}

	mux := router.InitializeMux(database, cfg)

	log.Println("Started server on port: " + cfg.PORT)
	log.Fatal(http.ListenAndServe(":"+cfg.PORT, *mux))
}
