package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/internal/config"
)

var Database *ent.Client

func NewDatabase(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func NewEntClient(db *sql.DB, logger *logrus.Logger) (*ent.Client, error) {
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))

	// Run migrations
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	Database = client
	if logger != nil {
		logger.Info("Database connection established and schema created")
	}
	return client, nil
}
