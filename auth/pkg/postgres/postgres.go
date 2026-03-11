package postgres

import (
	"auth/configs"
	"auth/internal/app/core/helpers/errorhandler"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	logging "gorm.io/gorm/logger"
)

func NewClient() *gorm.DB {
	newLogger := logging.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logging.Config{
			SlowThreshold:             time.Minute,
			LogLevel:                  logging.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(generateURI()), &gorm.Config{
		Logger: newLogger,
	})

	errorhandler.Fatal(err, "failed to connect database")

	psqlDB, err := db.DB()
	errorhandler.Fatal(err, "failed to get db object")

	psqlDB.SetMaxIdleConns(16)
	psqlDB.SetMaxOpenConns(32)
	psqlDB.SetConnMaxIdleTime(2 * time.Minute)
	psqlDB.SetConnMaxLifetime(45 * time.Minute)

	return db
}

func generateURI() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%v/%s?sslmode=%s&application_name=%s",
		configs.Config.DB.Postgres.Connection,
		configs.Config.DB.Postgres.Username,
		configs.Config.DB.Postgres.Password,
		configs.Config.DB.Postgres.Host,
		configs.Config.DB.Postgres.Port,
		configs.Config.DB.Postgres.Database,
		configs.Config.DB.Postgres.SSLMode,
		fmt.Sprintf("%s_%s", configs.Config.App.Environment, configs.Config.App.Name),
	)
}
