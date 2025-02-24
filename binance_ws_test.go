package main

import (
	"context"
	"strconv"
	"testing"
	"time"
)

// TestListenTradeReceivesValidData connects to Binance's trade websocket for BTCUSDT,
// waits up to 10 seconds for a Trade message, and validates that key fields are non-empty and sane.
func TestListenTradeReceivesValidData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tradeChan := make(chan Trade, 1)

	// Launch ListenTrade in a separate goroutine
	go func() {
		if err := ListenTrade(ctx, "BTCUSDT", tradeChan); err != nil && ctx.Err() == nil {
			t.Errorf("ListenTrade returned error: %v", err)
		}
	}()

	select {
	case trade := <-tradeChan:
		if trade.TradeID == 0 {
			t.Error("TradeID is zero")
		}
		if trade.Price == "" {
			t.Error("Price is empty")
		}
		if trade.Quantity == "" {
			t.Error("Quantity is empty")
		}
		if trade.EventType == "" {
			t.Error("EventType is empty")
		}
		if trade.EventTime == 0 {
			t.Error("EventTime is zero")
		}
	case <-ctx.Done():
		t.Fatal("Timed out waiting for a trade message")
	}
}

// TestListenAggTradeReceivesData tests the aggregated trade websocket for BTCUSDT.
func TestListenAggTradeReceivesData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	aggTradeChan := make(chan AggTrade, 1)
	go func() {
		if err := ListenAggTrade(ctx, "BTCUSDT", aggTradeChan); err != nil && ctx.Err() == nil {
			t.Errorf("ListenAggTrade returned error: %v", err)
		}
	}()

	select {
	case agg := <-aggTradeChan:
		if agg.EventType != "aggTrade" {
			t.Errorf("Expected EventType 'aggTrade', got %s", agg.EventType)
		}
		if agg.AggTradeID == 0 {
			t.Error("AggTradeID is zero")
		}
		if agg.Price == "" {
			t.Error("Price is empty")
		}
		if agg.Quantity == "" {
			t.Error("Quantity is empty")
		}
		if agg.TradeTime == 0 {
			t.Error("TradeTime is zero")
		}
		if agg.FirstTradeID == 0 {
			t.Error("FirstTradeID is zero")
		}
		if agg.LastTradeID == 0 {
			t.Error("LastTradeID is zero")
		}
		if agg.Symbol == "" {
			t.Error("Symbol is empty")
		}
	case <-ctx.Done():
		t.Fatal("Timed out waiting for aggregated trade message")
	}
}

func TestListenOrderBookDiffReceivesValidData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	diffChan := make(chan OrderBookDiff, 1)
	go func() {
		if err := ListenOrderBookDiff(ctx, "BTCUSDT", diffChan); err != nil && ctx.Err() == nil {
			t.Errorf("ListenOrderBookDiff returned error: %v", err)
		}
	}()

	select {
	case diff := <-diffChan:
		if diff.EventType == "" {
			t.Error("OrderBookDiff EventType is empty")
		}
		if diff.EventTime <= 0 {
			t.Errorf("Invalid EventTime: %d", diff.EventTime)
		}
		if len(diff.Bids) == 0 {
			t.Error("Diff Bids array is empty")
		}
		if len(diff.Asks) == 0 {
			t.Error("Diff Asks array is empty")
		}
		if len(diff.Bids) == 0 && len(diff.Asks) == 0 {
			t.Error("Both Bids and Asks are empty")
		}
	case <-ctx.Done():
		t.Fatal("Timed out waiting for order book diff message")
	}
}

// TestListenBestPriceReceivesValidData subscribes to Binance's best price websocket for BTCUSDT,
// waits up to 10 seconds for a BestPrice message, and asserts that key fields are valid and parseable.
func TestListenBestPriceReceivesValidData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bestPriceChan := make(chan BestPrice, 1)
	go func() {
		if err := ListenBestPrice(ctx, "BTCUSDT", bestPriceChan); err != nil && ctx.Err() == nil {
			t.Errorf("ListenBestPrice returned error: %v", err)
		}
	}()

	select {
	case bp := <-bestPriceChan:
		if bp.UpdateID == 0 {
			t.Error("UpdateID is zero")
		}
		if bp.Symbol == "" {
			t.Error("Symbol is empty")
		}
		if bp.BidPrice == "" {
			t.Error("BidPrice is empty")
		}
		if bp.AskPrice == "" {
			t.Error("AskPrice is empty")
		}
		if bp.BidQty == "" {
			t.Error("BidQty is empty")
		}
		if bp.AskQty == "" {
			t.Error("AskQty is empty")
		}
		if _, err := strconv.ParseFloat(bp.BidPrice, 64); err != nil {
			t.Errorf("BidPrice is not a valid float: %v", err)
		}
		if _, err := strconv.ParseFloat(bp.AskPrice, 64); err != nil {
			t.Errorf("AskPrice is not a valid float: %v", err)
		}
		if _, err := strconv.ParseFloat(bp.BidQty, 64); err != nil {
			t.Errorf("BidQty is not a valid float: %v", err)
		}
		if _, err := strconv.ParseFloat(bp.AskQty, 64); err != nil {
			t.Errorf("AskQty is not a valid float: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Timed out waiting for a best price message")
	}
}

// TestWebSocketContextCancellation verifies that the websocket listener exits cleanly when its context is cancelled.
func TestWebSocketContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	// Use the ListenTrade subscription as a representative websocket subscription
	tradeChan := make(chan Trade, 1)

	go func() {
		errChan <- ListenTrade(ctx, "BTCUSDT", tradeChan)
	}()

	// Allow a brief moment for the connection to establish
	time.Sleep(50 * time.Millisecond)
	// Cancel the context to trigger shutdown
	cancel()

	err := <-errChan
	if err != context.Canceled {
		t.Fatalf("Expected error to be context.Canceled, got: %v", err)
	}
}
