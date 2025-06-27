package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DEBUG level for detailed troubleshooting
	DEBUG LogLevel = iota
	// INFO level for general operational information
	INFO
	// WARNING level for potential issues
	WARNING
	// ERROR level for errors that don't cause application failure
	ERROR
	// FATAL level for critical errors that cause application failure
	FATAL
)

var (
	// Logger is the global logger instance
	Logger *CustomLogger
)

// CustomLogger provides enhanced logging capabilities
type CustomLogger struct {
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
	minLevel      LogLevel
}

// InitLogger initializes the global logger
func InitLogger(minLevel LogLevel) {
	Logger = NewCustomLogger(minLevel)
}

// NewCustomLogger creates a new custom logger
func NewCustomLogger(minLevel LogLevel) *CustomLogger {
	return &CustomLogger{
		debugLogger:   log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime),
		infoLogger:    log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		warningLogger: log.New(os.Stdout, "[WARNING] ", log.Ldate|log.Ltime),
		errorLogger:   log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime),
		fatalLogger:   log.New(os.Stderr, "[FATAL] ", log.Ldate|log.Ltime),
		minLevel:      minLevel,
	}
}

// getCallerInfo returns the file name and line number of the caller
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	
	// Extract just the filename from the full path
	parts := strings.Split(file, "/")
	fileName := parts[len(parts)-1]
	
	return fmt.Sprintf("%s:%d", fileName, line)
}

// formatMessage formats a log message with caller information
func formatMessage(message string) string {
	callerInfo := getCallerInfo(3) // Skip 3 levels to get the actual caller
	return fmt.Sprintf("[%s] %s", callerInfo, message)
}

// Debug logs a debug message
func (l *CustomLogger) Debug(format string, v ...interface{}) {
	if l.minLevel <= DEBUG {
		l.debugLogger.Printf(formatMessage(format), v...)
	}
}

// Info logs an info message
func (l *CustomLogger) Info(format string, v ...interface{}) {
	if l.minLevel <= INFO {
		l.infoLogger.Printf(formatMessage(format), v...)
	}
}

// Warning logs a warning message
func (l *CustomLogger) Warning(format string, v ...interface{}) {
	if l.minLevel <= WARNING {
		l.warningLogger.Printf(formatMessage(format), v...)
	}
}

// Error logs an error message
func (l *CustomLogger) Error(format string, v ...interface{}) {
	if l.minLevel <= ERROR {
		l.errorLogger.Printf(formatMessage(format), v...)
	}
}

// Fatal logs a fatal message and exits the application
func (l *CustomLogger) Fatal(format string, v ...interface{}) {
	if l.minLevel <= FATAL {
		l.fatalLogger.Printf(formatMessage(format), v...)
		os.Exit(1)
	}
}

// LogRequest logs an API request
func (l *CustomLogger) LogRequest(method, path, ip string, statusCode int, duration time.Duration) {
	l.Info("Request: %s %s from %s - Status: %d - Duration: %v", 
		method, path, ip, statusCode, duration)
}

// LogBlockchainEvent logs blockchain-related events
func (l *CustomLogger) LogBlockchainEvent(eventType, details string) {
	l.Info("Blockchain Event: %s - %s", eventType, details)
} 