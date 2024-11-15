package migrations

import (
	"errors"

	"github.com/gouthamkrishnakv/chatty/database/models"
	"gorm.io/gorm"
)

var dbModels = []interface{}{
	&models.User{},
	&models.Message{},
}

func RunMigration(db *gorm.DB) error {
	if db == nil {
		return errors.New("database connection is nil")
	}
	return db.AutoMigrate(dbModels...)
}
