package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM connection using the given DB settings.
func Connect(db DBConfig) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(db.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}
