package logger

import (
	"fmt"
	"log"
)

var verbosityLevel LogLevel = 2

// A logger's level, valued 0-6 where lower levels higher verbosity
type LogLevel uint8

const (
	// Print everything
	LogLevelAll LogLevel = iota

	// Print debug, info, warning, errors and fatal errors
	LogLevelDebug

	// Print info, warning, errors and fatal errors
	LogLevelInfo

	// Print warnings, errors and fatal errors
	LogLevelWarn

	// Print only errors and fatal errors
	LogLevelError

	// Print only fatal errors
	LogLevelFatal

	// Print nothing. Effectively disables logging
	LogLevelNone
)

// getPrintPrefix Gets a log prefix for the given log level
func getPrintPrefix(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	}

	return "DEBUG" // Should not get here
}

// SetLogLevel Sets the logger's verbosity level
func SetLogLevel(level LogLevel) {
	verbosityLevel = level
}

// Prints a message, if level is greater to or equal to the currently set verbosity level
func logger(level LogLevel, format string, v ...interface{}) {
	if level >= verbosityLevel {
		message := fmt.Sprintf("[%s] %s", getPrintPrefix(level), format)
		log.Printf(message, v...)
	}
}

// Prints a debug message
func Debug(format string, v ...interface{}) {
	logger(LogLevelDebug, format, v...)
}

// Prints an info message
func Info(format string, v ...interface{}) {
	logger(LogLevelInfo, format, v...)
}

// Prints a warning message
func Warn(format string, v ...interface{}) {
	logger(LogLevelWarn, format, v...)
}

// Prints an error message
func Error(format string, v ...interface{}) {
	logger(LogLevelError, format, v...)
}

// Prints a fatal message
func Fatal(format string, v ...interface{}) {
	logger(LogLevelFatal, format, v...)
}
