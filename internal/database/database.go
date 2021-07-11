package database

import (
	"fmt"
	"imgpool/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDatabase ...
func InitDatabase(c *config.Config) (*gorm.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Table, c.Database.Pass)

	conn, e := gorm.Open(postgres.Open(connStr))
	if e != nil {
		return conn, e
	}

	return conn, nil
}
