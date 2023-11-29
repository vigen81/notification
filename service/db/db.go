package db

import (
	`context`
	`fmt`
	
	_ "github.com/go-sql-driver/mysql"
	`go-micro.dev/v4/config/reader`
	`go-micro.dev/v4/logger`
	
	`gitlab.com/healthcare-integration/golang/notification-service/ent/migrate`
	
	// "gitlab.com/healthcare-integration/golang/notification-service/core"
	"gitlab.com/healthcare-integration/golang/notification-service/ent"
)

type database struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Database string `json:"database"`
}

// func Connect() {
//	var err error
//	conf := core.Config.database
//	dsn := fmt.Sprintf(
//		"%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
//		conf.Username,
//		conf.Password,
//		conf.Host,
//		conf.DbName,
//	)
//	dsn = fmt.Sprintf("root:Ap123456!!@tcp(localhost:33063)/hcare?charset=utf8mb4&parseTime=True&loc=Local")
//
//	fmt.Println(dsn)
//	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
//
//	if nil != err {
//		panic(err)
//	}
// }
//
// func DB() *gorm.DB {
//	return db.Debug()
// }
//
// func Migrate() {
//	err := DB().AutoMigrate(
//		&model.Patient{},
//		&model.Session{},
//		&model.VerificationCode{},
//		&model.Notification{},
//		&model.DepartmentID{},
//		&model.Provider{},
//	)
//	if nil != err {
//		panic(err)
//	}
// }
var db *ent.Client

func Client() *ent.Client {
	return db.Debug()
}

var params database

func Configure(r reader.Value) error {
	err := r.Scan(&params)
	if nil != err {
		return err
	}
	return Connect()
	
}

func Connect() (err error) {
	
	if err != nil {
		return err
	}
	
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		params.User,
		params.Password,
		params.Host,
		params.Port,
		params.Database,
	)
	
	db, err = ent.Open("mysql", dsn, ent.Log(func(i ...interface{}) {
		logger.Debug(i...)
	}))
	
	// }
	
	if err != nil {
		logger.Fatalf("failed opening connection to sqlite: %v", err)
	}
	// return
	// Run the auto migration tool.
	if err := db.Debug().Schema.Create(context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
		migrate.WithForeignKeys(true),
	); err != nil {
		logger.Fatalf("failed creating schema resources: %v", err)
	}
	return nil
}
