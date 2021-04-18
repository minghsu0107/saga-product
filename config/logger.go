package config

import (
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	dblog "gorm.io/gorm/logger"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})
}

// Logger is the logger type
type Logger struct {
	Writer        io.Writer
	DBLogger      dblog.Interface
	ContextLogger *log.Entry
}

func newLogger(appName, ginMode string) *Logger {
	var dbLogLevel dblog.LogLevel
	var dbLoggerLevel log.Level

	writer := os.Stderr

	if ginMode == "release" {
		dbLogLevel = dblog.Silent      // disable gorm log
		dbLoggerLevel = log.ErrorLevel // just a placeholder
	} else {
		dbLogLevel = dblog.Info
		dbLoggerLevel = log.DebugLevel
	}

	var dbLogger = &log.Logger{
		Out:       writer,
		Formatter: new(log.TextFormatter),
		Level:     dbLoggerLevel,
	}

	contextLogger := log.WithFields(log.Fields{
		"app_name": appName,
	})

	return &Logger{
		Writer: writer,
		DBLogger: dblog.New(
			dbLogger,
			dblog.Config{
				SlowThreshold: time.Second, // Slow SQL threshold
				LogLevel:      dbLogLevel,  // Log level
				Colorful:      true,        // Enable color
			},
		),
		ContextLogger: contextLogger,
	}
}
