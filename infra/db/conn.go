package db

import (
	conf "github.com/minghsu0107/saga-product/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewDatabaseConnection returns the db connection instance
func NewDatabaseConnection(config *conf.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.DBConfig.Dsn), &gorm.Config{
		Logger: config.Logger.DBLogger,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(config.DBConfig.MaxIdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(config.DBConfig.MaxOpenConns)
	config.Logger.ContextLogger.WithField("type", "setup:db").Info("successful SQL connection")
	return db, nil
}
