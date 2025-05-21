package main

import (
	"log"

	"lang-portal/backend_go/internal/api"
	"lang-portal/backend_go/internal/database"
)

func main() {
	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	router := api.SetupRouter(db)

	// Start server
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
