package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
)

// custom error type for status codes
type StatusError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *StatusError) Error() string {
	return e.Message
}

func NewStatusError(message string, code int) *StatusError {
	return &StatusError{
		Message: message,
		Code:    code,
	}
}

// ErrorResponse is a struct for the error response
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// HandleError is a function that handles errors and writes the response
// It takes a http.ResponseWriter, an error, a message, a level and a status code
// It checks if the error is a StatusError, if so it sets the status code
func HandleError(w http.ResponseWriter, err error, message string, level string, statusCode int) {
	// if error is status error, set status code
	logError(err, message, level)
	if statusErr, ok := err.(*StatusError); ok {
		statusCode = statusErr.Code
		err = fmt.Errorf("%s", statusErr.Message)
	} else {
		statusCode, err = sanitizeError(err, message, statusCode)
	}
	writeErrorResponse(w, err.Error(), statusCode)
}

func logError(err error, message string, level string) {
	logEntry := Logger.WithFields(map[string]interface{}{
		"level": level,
		"error": err.Error(),
		"stack": string(debug.Stack()),
	})
	if message != "" {
		logEntry = logEntry.WithField("message", message)
	}
	switch strings.ToLower(level) {
	case "warn", "warning":
		logEntry.Warn("Handled error")
	case "info":
		logEntry.Info("Handled error")
	default:
		logEntry.Error("Handled error")
	}
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
	})
}

func sanitizeError(err error, message string, statusCode int) (int, error) {
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate key value violates unique constraint") {
		return http.StatusConflict, fmt.Errorf("record already exists")
	}
	if strings.Contains(errStr, "no rows in result set") {
		return http.StatusNotFound, fmt.Errorf("no record found")

	}
	if strings.Contains(errStr, "violates foreign key constraint") {
		return http.StatusBadRequest, fmt.Errorf("referenced entity does not exist or is invalid")
	}

	if strings.Contains(errStr, "violates not-null constraint") {
		column := strings.Split(errStr, "\"")[1]
		return http.StatusBadRequest, fmt.Errorf("missing required field: %s", column)
	}

	if strings.Contains(errStr, "invalid input syntax for type") {
		return http.StatusBadRequest, fmt.Errorf("invalid data format provided")
	}

	if strings.Contains(errStr, "deadlock detected") {
		return http.StatusInternalServerError, fmt.Errorf("a temporary issue occurred, please try again")
	}

	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "connection timeout") {
		return http.StatusServiceUnavailable, fmt.Errorf("unable to connect to the database, please try again later")
	}

	if strings.Contains(errStr, "syntax error at or near") {
		return http.StatusInternalServerError, fmt.Errorf("an internal error occurred, please contact support")
	}

	if strings.Contains(errStr, "current transaction is aborted") {
		return http.StatusInternalServerError, fmt.Errorf("an error occurred during the operation, please try again")
	}

	if strings.Contains(errStr, "value too long for type") {
		return http.StatusBadRequest, fmt.Errorf("input value exceeds the allowed length")
	}

	if strings.Contains(errStr, "permission denied") {
		return http.StatusForbidden, fmt.Errorf("you do not have permission to perform this action")
	}

	if strings.Contains(errStr, "context deadline exceeded") {
		return http.StatusGatewayTimeout, fmt.Errorf("the operation timed out, please try again")
	}

	if strings.Contains(errStr, "invalid input syntax for type json") {
		return http.StatusBadRequest, fmt.Errorf("invalid JSON format provided")
	}
	return statusCode, fmt.Errorf("%s", message)
}
