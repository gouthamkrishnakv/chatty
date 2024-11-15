package database

import (
	"github.com/gouthamkrishnakv/chatty/database/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB = nil

// Use this as an injectable DB object
func GetDB() *gorm.DB {
	return db
}

func Init() error {
	sqliteDb, openErr := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if openErr != nil {
		return openErr
	}
	if migrationErr := migrations.RunMigration(sqliteDb); migrationErr != nil {
		return migrationErr
	}
	// Make sure db is connected only after migration
	db = sqliteDb
	return nil
}
