// Package compiler provides the Arc language compiler implementation.
package compiler

import (
	"fmt"
	"os"
	"sync"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// Global logging configuration
const (
	EnableDebugLogging   = true  // Set to false to disable debug logs
	EnableInfoLogging    = true  // Set to false to disable info logs
	EnableWarningLogging = true  // Set to false to disable warning logs
	EnableErrorLogging   = true  // Always keep errors enabled
)

// Logger provides centralized logging for the compiler
type Logger struct {
	mu          sync.Mutex
	prefix      string
	errorCount  int
	warnCount   int
	infoCount   int
	debugCount  int
}

var (
	globalLogger = &Logger{prefix: "[Arc]"}
)

// NewLogger creates a new logger with a custom prefix
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if EnableDebugLogging {
		l.log(LogLevelDebug, format, args...)
		l.mu.Lock()
		l.debugCount++
		l.mu.Unlock()
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	if EnableInfoLogging {
		l.log(LogLevelInfo, format, args...)
		l.mu.Lock()
		l.infoCount++
		l.mu.Unlock()
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	if EnableWarningLogging {
		l.log(LogLevelWarning, format, args...)
		l.mu.Lock()
		l.warnCount++
		l.mu.Unlock()
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if EnableErrorLogging {
		l.log(LogLevelError, format, args...)
		l.mu.Lock()
		l.errorCount++
		l.mu.Unlock()
	}
}

// ErrorAt logs an error at a specific source location
func (l *Logger) ErrorAt(file string, line, column int, format string, args ...interface{}) {
	if EnableErrorLogging {
		message := fmt.Sprintf(format, args...)
		l.log(LogLevelError, "%s:%d:%d: %s", file, line, column, message)
		l.mu.Lock()
		l.errorCount++
		l.mu.Unlock()
	}
}

// WarningAt logs a warning at a specific source location
func (l *Logger) WarningAt(file string, line, column int, format string, args ...interface{}) {
	if EnableWarningLogging {
		message := fmt.Sprintf(format, args...)
		l.log(LogLevelWarning, "%s:%d:%d: %s", file, line, column, message)
		l.mu.Lock()
		l.warnCount++
		l.mu.Unlock()
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var levelStr string
	var output *os.File
	
	switch level {
	case LogLevelDebug:
		levelStr = "DEBUG"
		output = os.Stdout
	case LogLevelInfo:
		levelStr = "INFO"
		output = os.Stdout
	case LogLevelWarning:
		levelStr = "WARN"
		output = os.Stderr
	case LogLevelError:
		levelStr = "ERROR"
		output = os.Stderr
	}

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(output, "%s [%s] %s\n", l.prefix, levelStr, message)
}

// HasErrors returns true if any errors were logged
func (l *Logger) HasErrors() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.errorCount > 0
}

// ErrorCount returns the number of errors logged
func (l *Logger) ErrorCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.errorCount
}

// WarningCount returns the number of warnings logged
func (l *Logger) WarningCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.warnCount
}

// Reset resets all counters
func (l *Logger) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errorCount = 0
	l.warnCount = 0
	l.infoCount = 0
	l.debugCount = 0
}

// PrintSummary prints a summary of logged messages
func (l *Logger) PrintSummary() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.errorCount > 0 || l.warnCount > 0 {
		fmt.Fprintf(os.Stderr, "\n%s Compilation Summary:\n", l.prefix)
		if l.errorCount > 0 {
			fmt.Fprintf(os.Stderr, "  Errors: %d\n", l.errorCount)
		}
		if l.warnCount > 0 {
			fmt.Fprintf(os.Stderr, "  Warnings: %d\n", l.warnCount)
		}
	}
}

// Global logging functions for convenience
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

func Warning(format string, args ...interface{}) {
	globalLogger.Warning(format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

func ErrorAt(file string, line, column int, format string, args ...interface{}) {
	globalLogger.ErrorAt(file, line, column, format, args...)
}

func WarningAt(file string, line, column int, format string, args ...interface{}) {
	globalLogger.WarningAt(file, line, column, format, args...)
}

func HasErrors() bool {
	return globalLogger.HasErrors()
}

func ErrorCount() int {
	return globalLogger.ErrorCount()
}

func WarningCount() int {
	return globalLogger.WarningCount()
}

func PrintSummary() {
	globalLogger.PrintSummary()
}