package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func InitLogger() {
	Logger = logrus.New()

	// Set log format
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z",
	})

	// Set log level
	Logger.SetLevel(logrus.InfoLevel)

	// Set output
	logFile := "/logs/ai-service/app.log"
	if file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		Logger.SetOutput(file)
	} else {
		Logger.SetOutput(os.Stdout)
		Logger.Warnf("Failed to open log file, using stdout: %v", err)
	}

	// Also log to stdout in development
	if os.Getenv("ENV") == "development" {
		Logger.SetOutput(os.Stdout)
	}
}
