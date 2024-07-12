package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"sync"
	"time"
)

var (
	defaultMaxOpenConns    = 100
	defaultMaxIdleConns    = 2
	defaultConnMaxLifetime = time.Hour * 1
)

// Database singleton struct
type Database struct {
	*gorm.DB
}

// A package-level variable to hold the singleton instance
var (
	dbInstance *Database
	dbonce     sync.Once
)

// GetDB returns the singleton database instance
func GetDB() *Database {
	dbonce.Do(func() {
		dialector := postgres.New(postgres.Config{
			DriverName: "pgx",
			DSN:        os.Getenv("DATABASE_URL"),
		})

		db, err := gorm.Open(dialector, &gorm.Config{
			PrepareStmt: true,
		})
		if err != nil {
			log.Fatal("Failed to connect to the database:", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal("Failed to connect to the database:", err)
		}

		sqlDB.SetMaxOpenConns(defaultMaxOpenConns)
		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		sqlDB.SetConnMaxLifetime(defaultConnMaxLifetime)
		sqlDB.SetMaxIdleConns(defaultMaxIdleConns)

		dbInstance = &Database{
			DB: db,
		}
	})
	return dbInstance
}
