package main

import (
	"fmt"
	"os"
	"time"
)

// BuildFileName constructs a Parquet file name based on the provided instrument, data type,
// and the UTC date extracted from the given time.Time value.
// The returned file name format is: <instrument>_<dataType>_<YYYY-MM-DD>.parquet
// For example, BuildFileName("trade", "BTCUSDT", someTime) might return "BTCUSDT_trade_2023-10-15.parquet".
func BuildFileName(dataType string, instrument string, t time.Time) string {
	utcDate := t.UTC().Format("2006-01-02")
	return fmt.Sprintf("%s_%s_%s.parquet", instrument, dataType, utcDate)
}

// FileExists checks if the specified file exists at filePath.
// It returns true if the file exists, and false otherwise.
// This function wraps the os.Stat call, providing an imperative shell for IO,
// while the BuildFileName function remains a pure function for easier testing.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	// For any error that is not "file does not exist", assume the file exists.
	return true
}
