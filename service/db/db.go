package db

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/healthcare-integration/golang/notification-service/ent"
	"gitlab.com/healthcare-integration/golang/notification-service/ent/migrate"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/logger"
	// "gitlab.com/healthcare-integration/golang/storage-service/core"
)

type database struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Database string `json:"database"`
}

var db *ent.Client

func Client() *ent.Client {
	return db
}

var params database

func Configure(r reader.Value) error {
	err := r.Scan(&params)
	if nil != err {
		return err
	}
	return Connect()

}
func Tx(handler func(tx *ent.Tx) error) error {
	tx, err := Client().Tx(context.Background())
	if err != nil {
		return err
	}

	err = handler(tx)

	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()

}

func Connect() (err error) {

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		params.User,
		params.Password,
		params.Host,
		params.Port,
		params.Database,
	)

	//db, err = ent.Open("mysql", dsn, ent.Debug())
	db, err = ent.Open("mysql", dsn)

	if err != nil {
		logger.Fatalf("failed opening connection to sqlite: %v", err)
	}
	// return
	// Restart the auto migration tool.
	if err := db.Debug().Schema.Create(context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
		migrate.WithForeignKeys(true),
	); err != nil {
		logger.Fatalf("failed creating schema resources: %v", err)
	}

	return nil
}
