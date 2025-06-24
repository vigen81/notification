package main

import (
	"context"
	"log"

	"gitlab.smartbet.am/golang/notification/internal/config"
	"gitlab.smartbet.am/golang/notification/internal/db"
)

func main() {
	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create Ent client and run migrations
	client, err := db.NewEntClient(database, nil)
	if err != nil {
		log.Fatalf("Failed to create ent client: %v", err)
	}
	defer client.Close()

	// Run the schema migration
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("Database migration completed successfully")
}
