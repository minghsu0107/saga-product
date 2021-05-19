package repo

import (
	"io/ioutil"
	"os"
	"time"

	dblog "gorm.io/gorm/logger"

	conf "github.com/minghsu0107/saga-product/config"
	infra_db "github.com/minghsu0107/saga-product/infra/db"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() {
	//writer := os.Stderr
	writer := ioutil.Discard
	config := &conf.Config{
		DBConfig: &conf.DBConfig{
			Dsn:          os.Getenv("DB_DSN"),
			MaxIdleConns: 0,
			MaxOpenConns: 1,
		},
		Logger: &conf.Logger{
			Writer: writer,
			ContextLogger: log.WithFields(log.Fields{
				"app_name": "test",
			}),
			DBLogger: dblog.New(
				&log.Logger{
					Out:       writer,
					Formatter: new(log.TextFormatter),
					Level:     log.DebugLevel,
				},
				dblog.Config{
					SlowThreshold: time.Second,
					LogLevel:      dblog.Info,
					Colorful:      true,
				},
			),
		},
	}

	var err error
	db, err = infra_db.NewDatabaseConnection(config)
	if err != nil {
		panic(err)
	}
}
