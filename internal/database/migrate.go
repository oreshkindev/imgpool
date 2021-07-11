package database

import (
	"imgpool/internal/services/pool"

	"gorm.io/gorm"
)

// MigrateDB ...
func MigrateDB(conn *gorm.DB) error {
	if e := conn.AutoMigrate(&pool.Imgpool{}); e != nil {
		return e
	}
	return nil
}
