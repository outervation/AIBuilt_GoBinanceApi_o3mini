package main

import (
	"net/http"
	"testing"
	"time"
)

func TestParseOrderBookSnapshot_ValidInput(t *testing.T) {
	validJSON := "{\"lastUpdateId\":12345, \"bids\": [[\"100.0\", \"1.0\"]], \"asks\": [[\"101.0\", \"2.0\"]]}"

	snapshot, err := parseOrderBookSnapshot([]byte(validJSON))
	if err != nil {
		t.Fatalf("Error parsing snapshot: %v", err)
	}

	if snapshot.LastUpdateID != 12345 {
		t.Errorf("Expected LastUpdateID 12345, got %d", snapshot.LastUpdateID)
	}

	if len(snapshot.Bids) != 1 {
		t.Errorf("Expected 1 bid, got %d", len(snapshot.Bids))
	} else {
		if snapshot.Bids[0].Price != "100.0" {
			t.Errorf("Expected bid price '100.0', got %s", snapshot.Bids[0].Price)
		}
		if snapshot.Bids[0].Quantity != "1.0" {
			t.Errorf("Expected bid quantity '1.0', got %s", snapshot.Bids[0].Quantity)
		}
	}

	if len(snapshot.Asks) != 1 {
		t.Errorf("Expected 1 ask, got %d", len(snapshot.Asks))
	} else {
		if snapshot.Asks[0].Price != "101.0" {
			t.Errorf("Expected ask price '101.0', got %s", snapshot.Asks[0].Price)
		}
		if snapshot.Asks[0].Quantity != "2.0" {
			t.Errorf("Expected ask quantity '2.0', got %s", snapshot.Asks[0].Quantity)
		}
	}
}

func TestParseOrderBookSnapshot_InvalidInput(t *testing.T) {
	// Test with a malformed JSON string (e.g., missing a closing brace)
	malformedJSON := "{\"lastUpdateId\":12345, \"bids\": [[\"100.0\", \"1.0\"]], \"asks\": [[\"101.0\", \"2.0\"]"
	if _, err := parseOrderBookSnapshot([]byte(malformedJSON)); err == nil {
		t.Fatalf("Expected error when parsing malformed JSON, but got nil")
	}

	// Test with incomplete JSON data: bids array element missing quantity
	incompleteJSON := "{\"lastUpdateId\":12345, \"bids\": [[\"100.0\"]], \"asks\": [[\"101.0\", \"2.0\"]]}"
	if _, err := parseOrderBookSnapshot([]byte(incompleteJSON)); err == nil {
		t.Fatalf("Expected error when parsing incomplete JSON, but got nil")
	}
}

func TestFetchOrderBookSnapshot_LiveData(t *testing.T) {
	// Create an HTTP client with a timeout of 10 seconds
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	snapshot, err := FetchOrderBookSnapshot(client, "BTCUSDT")
	if err != nil {
		t.Fatalf("Failed to fetch snapshot from live API: %v", err)
	}

	if snapshot.LastUpdateID == 0 {
		t.Fatalf("Expected non-zero LastUpdateID, got %d", snapshot.LastUpdateID)
	}

	if len(snapshot.Bids) == 0 {
		t.Fatalf("Expected non-empty bids array")
	}

	if len(snapshot.Asks) == 0 {
		t.Fatalf("Expected non-empty asks array")
	}
}
