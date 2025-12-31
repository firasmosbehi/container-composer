package utils

import (
	"fmt"
	"log"
	"os"
)

var (
	// Logger levels
	debugMode   bool
	verboseMode bool
)

// SetDebugMode enables or disables debug mode
func SetDebugMode(enabled bool) {
	debugMode = enabled
}

// SetVerboseMode enables or disables verbose mode
func SetVerboseMode(enabled bool) {
	verboseMode = enabled
}

// Debug prints debug messages when debug mode is enabled
func Debug(format string, args ...interface{}) {
	if debugMode {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Verbose prints verbose messages when verbose mode is enabled
func Verbose(format string, args ...interface{}) {
	if verboseMode || debugMode {
		log.Printf("[VERBOSE] "+format, args...)
	}
}

// Info prints informational messages
func Info(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// Warn prints warning messages
func Warn(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

// Error prints error messages
func Error(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// Fatal prints error message and exits
func Fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[FATAL] "+format+"\n", args...)
	os.Exit(1)
}

// Success prints success messages with color (if terminal supports it)
func Success(format string, args ...interface{}) {
	// TODO: Add color support
	log.Printf("[SUCCESS] "+format, args...)
}