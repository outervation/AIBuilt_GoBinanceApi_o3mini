package main

import (
	"os"
	"testing"
	"time"
)

// TestBuildFileNameFormatsCorrectly verifies that BuildFileName produces a file name
// in the format <instrument>_<dataType>_<YYYY-MM-DD>.parquet
func TestBuildFileNameFormatsCorrectly(t *testing.T) {
	// Use a fixed UTC time
	fixedTime := time.Date(2023, time.October, 15, 12, 34, 56, 0, time.UTC)

	// Expected file name per the format: instrument_dataType_date.parquet
	expected := "BTCUSDT_trade_2023-10-15.parquet"

	// Notice that the BuildFileName function expects dataType first and instrument second.
	actual := BuildFileName("trade", "BTCUSDT", fixedTime)

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestFileExistsReturnsFalseForNonexistent(t *testing.T) {
	// Create a file name that is highly unlikely to exist
	fakeFileName := "nonexistent_12345.parquet"
	if FileExists(fakeFileName) {
		t.Errorf("Expected FileExists(%q) to return false, but got true", fakeFileName)
	}
}

func TestFileExistsReturnsTrueForExistingFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "temp_test_file_*.parquet")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()

	if !FileExists(tempFileName) {
		t.Errorf("Expected FileExists(%q) to return true, but got false", tempFileName)
	}

	if err := os.Remove(tempFileName); err != nil {
		t.Errorf("failed to remove temp file %q: %v", tempFileName, err)
	}
}
