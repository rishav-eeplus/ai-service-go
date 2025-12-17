package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// CustomLogger wraps logrus.Logger with additional functionality
type CustomLogger struct {
	*logrus.Logger
}

var Logger *CustomLogger

// InitLogger initializes the custom logger
func InitLogger() {
	baseLogger := logrus.New()

	// Set log format
	baseLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z",
	})

	// Set log level
	baseLogger.SetLevel(logrus.InfoLevel)

	// Set output
	logFile := "/logs/ai-service/app.log"
	if file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		baseLogger.SetOutput(file)
	} else {
		baseLogger.SetOutput(os.Stdout)
		fmt.Printf("Failed to open log file, using stdout: %v\n", err)
	}

	Logger = &CustomLogger{Logger: baseLogger}
}

// WithContext returns a logger with context fields
func (l *CustomLogger) WithContext(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// LogRequest logs HTTP request information
func (l *CustomLogger) LogRequest(method, path string, statusCode int, duration float64) {
	l.WithFields(logrus.Fields{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration,
		"type":        "request",
	}).Info("HTTP Request")
}

// LogError logs error with additional context
func (l *CustomLogger) LogError(err error, context map[string]interface{}) {
	fields := logrus.Fields{"error": err.Error(), "type": "error"}
	for k, v := range context {
		fields[k] = v
	}
	l.WithFields(fields).Error("Error occurred")
}

// LogSuccess logs success message with context
func (l *CustomLogger) LogSuccess(message string, context map[string]interface{}) {
	fields := logrus.Fields{"type": "success"}
	for k, v := range context {
		fields[k] = v
	}
	l.WithFields(fields).Info(message)
}
