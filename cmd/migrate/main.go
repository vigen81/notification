package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent"
)

func main() {
	// Create a basic logger for migration
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting database migration...")

	// Get database configuration from environment variables or use defaults
	host := getEnv("DB_HOST", "localhost")
	port := "3306"
	user := getEnv("DB_USER", "notification_user")
	password := getEnv("DB_PASSWORD", "notification_pass")
	dbname := getEnv("DB_NAME", "notification_db")

	// Build MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(300 * time.Second)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	logger.Info("Database connection established successfully")

	// Create Ent driver and client
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	// Enable debug mode
	client = client.Debug()

	// Run migrations
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	logger.Info("Migration completed successfully!")
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
