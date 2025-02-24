package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Logger provides lightweight logging and journaling functionality that writes operational messages to a writer.
// It is designed with a functional core and an imperative shell for testability.

type Logger struct {
	w  io.Writer
	mu sync.Mutex
}

// NewLogger creates a new Logger that writes to the provided io.Writer.
func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w}
}

// NewFileLogger creates a Logger that writes to the "journal.txt" file in append mode. It will create the file if it does not exist.
func NewFileLogger() (*Logger, error) {
	if os.Getenv("GO_TEST_MAIN") == "1" {
		return NewLogger(os.Stdout), nil
	}
	f, err := os.OpenFile("journal.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return NewLogger(f), nil
}

// FormatLog is a pure function that returns a formatted log entry using the given level, message, and timestamp.
func FormatLog(level, message string, ts time.Time) string {
	return fmt.Sprintf("[%s] %s: %s", ts.Format("2006-01-02 15:04:05"), level, message)
}

// Log writes a log entry with the specified level and message. It obtains the current UTC timestamp and writes the log entry followed by a newline.
func (l *Logger) Log(level, message string) error {
	timestamp := time.Now().UTC()
	entry := FormatLog(level, message, timestamp)
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := l.w.Write([]byte(entry + "\n"))
	return err
}

// Info logs an informational message.
func (l *Logger) Info(message string) error {
	return l.Log("INFO", message)
}

// Error logs an error message.
func (l *Logger) Error(message string) error {
	return l.Log("ERROR", message)
}

// Logf logs a formatted message with a specified level.
func (l *Logger) Logf(level, format string, a ...interface{}) error {
	return l.Log(level, fmt.Sprintf(format, a...))
}

// Infof logs a formatted informational message.
func (l *Logger) Infof(format string, a ...interface{}) error {
	return l.Logf("INFO", format, a...)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, a ...interface{}) error {
	return l.Logf("ERROR", format, a...)
}
