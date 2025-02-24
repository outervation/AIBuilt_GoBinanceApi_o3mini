package main

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestFormatLog verifies that FormatLog produces a correctly formatted log entry for a given timestamp, level, and message.
func TestFormatLog(t *testing.T) {
	ts := time.Date(2023, 10, 15, 12, 34, 56, 0, time.UTC)
	result := FormatLog("INFO", "test message", ts)
	expected := "[2023-10-15 12:34:56] INFO: test message"
	if result != expected {
		t.Errorf("FormatLog() = %q, want %q", result, expected)
	}
}

func TestLoggerLog(t *testing.T) {
	var buf bytes.Buffer
	lg := NewLogger(&buf)
	level := "TEST"
	message := "test log event"
	if err := lg.Log(level, message); err != nil {
		t.Fatalf("Logger.Log returned error: %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Fatalf("No output captured from Logger.Log")
	}
	output = strings.TrimSuffix(output, "\n")
	// Validate that output starts with the expected formatted timestamp and level/message
	if len(output) < 21 { // minimal "[<timestamp>]" length check
		t.Fatalf("Output too short to contain valid log entry: %q", output)
	}
	if output[0] != '[' {
		t.Errorf("Output does not start with '[': %s", output)
	}
	// Extract timestamp from output
	closeIdx := strings.Index(output, "]")
	if closeIdx == -1 {
		t.Errorf("Output missing closing bracket: %s", output)
	}
	tsStr := output[1:closeIdx]
	ts, err := time.Parse("2006-01-02 15:04:05", tsStr)
	if err != nil {
		t.Errorf("Timestamp parsing failed for %q: %v", tsStr, err)
	}
	now := time.Now().UTC()
	diff := now.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2*time.Second {
		t.Errorf("Logged timestamp %v differs from current time %v by more than 2 seconds", ts, now)
	}
	expectedSuffix := "] " + level + ": " + message
	if !strings.HasSuffix(output, expectedSuffix) {
		t.Errorf("Logged output does not have expected suffix. Expected: %q, Got: %q", expectedSuffix, output)
	}
}

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	lg := NewLogger(&buf)
	message := "info log event"
	if err := lg.Info(message); err != nil {
		t.Fatalf("Logger.Info returned error: %v", err)
	}
	output := strings.TrimSuffix(buf.String(), "\n")
	if output == "" {
		t.Fatalf("No output captured from Logger.Info")
	}
	expectedSuffix := "] INFO: " + message
	if !strings.HasSuffix(output, expectedSuffix) {
		t.Errorf("Logger.Info output does not have expected suffix. Expected: %q, Got: %q", expectedSuffix, output)
	}
	if output[0] != '[' {
		t.Fatalf("Logger.Info output does not start with '[': %q", output)
	}
	closeIdx := strings.Index(output, "]")
	if closeIdx == -1 {
		t.Fatalf("Logger.Info output missing closing bracket")
	}
	tsStr := output[1:closeIdx]
	ts, err := time.Parse("2006-01-02 15:04:05", tsStr)
	if err != nil {
		t.Fatalf("Timestamp parsing failed for %q: %v", tsStr, err)
	}
	now := time.Now().UTC()
	diff := now.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2*time.Second {
		t.Errorf("Logged timestamp %v differs from current time %v by more than 2 seconds", ts, now)
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	lg := NewLogger(&buf)
	message := "error log event"
	if err := lg.Error(message); err != nil {
		t.Fatalf("Logger.Error returned error: %v", err)
	}
	output := strings.TrimSuffix(buf.String(), "\n")
	if output == "" {
		t.Fatalf("No output captured from Logger.Error")
	}
	if output[0] != '[' {
		t.Fatalf("Logger.Error output does not start with '[': %q", output)
	}
	closeIdx := strings.Index(output, "]")
	if closeIdx == -1 {
		t.Fatalf("Logger.Error output missing closing bracket")
	}
	tsStr := output[1:closeIdx]
	ts, err := time.Parse("2006-01-02 15:04:05", tsStr)
	if err != nil {
		t.Fatalf("Timestamp parsing failed for %q: %v", tsStr, err)
	}
	now := time.Now().UTC()
	diff := now.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2*time.Second {
		t.Errorf("Logged timestamp %v differs from current time %v by more than 2 seconds", ts, now)
	}
	expectedSuffix := "] ERROR: " + message
	if !strings.HasSuffix(output, expectedSuffix) {
		t.Errorf("Logger.Error output does not have expected suffix. Expected: %q, Got: %q", expectedSuffix, output)
	}
}

func TestLoggerConcurrency(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	const goroutines = 10
	const messages = 100

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messages; j++ {
				msg := fmt.Sprintf("goroutine=%d message=%d", id, j)
				if err := logger.Info(msg); err != nil {
					t.Errorf("Logger.Info failed: %v", err)
				}
			}
		}(i)
	}
	wg.Wait()

	entries := strings.Split(strings.TrimSpace(buf.String()), "\n")
	expected := goroutines * messages
	if len(entries) != expected {
		t.Fatalf("Expected %d log entries, got %d", expected, len(entries))
	}

	for _, entry := range entries {
		if len(entry) == 0 {
			t.Errorf("Empty log entry encountered")
			continue
		}
		if entry[0] != '[' {
			t.Errorf("Log entry does not start with '[': %s", entry)
		}
		idx := strings.Index(entry, "]")
		if idx == -1 {
			t.Errorf("Log entry missing closing bracket: %s", entry)
			continue
		}
		timestampStr := entry[1:idx]
		if _, err := time.Parse("2006-01-02 15:04:05", timestampStr); err != nil {
			t.Errorf("Invalid timestamp in log entry: %s; error: %v", entry, err)
		}
		if !strings.Contains(entry, "] INFO: ") {
			t.Errorf("Log entry missing expected pattern '] INFO: ': %s", entry)
		}
		if !strings.Contains(entry, "goroutine=") {
			t.Errorf("Log entry missing expected 'goroutine=' content: %s", entry)
		}
	}
}
