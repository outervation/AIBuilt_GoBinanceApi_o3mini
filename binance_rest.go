package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// orderBookSnapshotResponse defines the JSON structure returned by the Binance REST API.
type orderBookSnapshotResponse struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

// parseOrderBookSnapshot parses raw JSON data into an OrderBookSnapshot struct.
// This function acts as the pure functional core: it has no side effects and can be easily tested.
func parseOrderBookSnapshot(data []byte) (*OrderBookSnapshot, error) {
	var resp orderBookSnapshotResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	bids := make([]PriceLevel, len(resp.Bids))
	for i, bid := range resp.Bids {
		if len(bid) < 2 {
			return nil, errors.New("malformed bid data")
		}
		bids[i] = PriceLevel{
			Price:    bid[0],
			Quantity: bid[1],
		}
	}

	asks := make([]PriceLevel, len(resp.Asks))
	for i, ask := range resp.Asks {
		if len(ask) < 2 {
			return nil, errors.New("malformed ask data")
		}
		asks[i] = PriceLevel{
			Price:    ask[0],
			Quantity: ask[1],
		}
	}

	return &OrderBookSnapshot{
		LastUpdateID: resp.LastUpdateID,
		Bids:         bids,
		Asks:         asks,
	}, nil
}

// FetchOrderBookSnapshot makes an HTTP GET request to Binance's REST API for the order book snapshot
// of the given instrument. It uses the provided http.Client so that it can be easily mocked in tests.
func FetchOrderBookSnapshot(client *http.Client, instrument string) (*OrderBookSnapshot, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/depth?symbol=%s&limit=100", instrument)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	snapshot, err := parseOrderBookSnapshot(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	return snapshot, nil
}

// StartOrderBookSnapshotFetcher periodically fetches the order book snapshot for a given instrument
// at the specified interval. It sends each successfully fetched snapshot to the provided channel.
// The function is designed with a functional core (FetchOrderBookSnapshot and parseOrderBookSnapshot) and an
// imperative shell (ticker-based scheduling and channel handling), enabling easier testing of the core logic.
func StartOrderBookSnapshotFetcher(ctx context.Context, client *http.Client, instrument string, interval time.Duration, out chan<- OrderBookSnapshot) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			snapshot, err := FetchOrderBookSnapshot(client, instrument)
			if err != nil {
				// In a production setting, consider logging this error using the Logger module.
				fmt.Printf("Error fetching snapshot for %s: %v\n", instrument, err)
				continue
			}
			select {
			case out <- *snapshot:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
