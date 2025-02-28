package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	// Local project packages (all files are part of package main)
	// Note: Logger, NewFileLogger, NewRecorder, BuildFileName etc. are defined in other files.
)

func main() {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal channel to gracefully shut down on SIGINT or SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize logger
	logger, err := NewFileLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Hardcoded instruments for initial testing
	instruments := []string{"BTCUSDT"}
	batchSize := 1

	// HTTP client for REST API calls

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// For each instrument, set up pipelines
	for _, instrument := range instruments {
		// Create channels for different data types with buffering
		tradeCh := make(chan Trade, 100)
		aggTradeCh := make(chan AggTrade, 100)
		diffCh := make(chan OrderBookDiff, 100)
		bestPriceCh := make(chan BestPrice, 100)

		// Create channels for snapshots
		// We'll use a raw snapshot channel which is fanned out to two separate channels: one for order book diff filtering and one for recording snapshots

		rawSnapshotCh := make(chan OrderBookSnapshot, 10)
		snapshotDiffCh := make(chan OrderBookSnapshot, 10)
		snapshotRecCh := make(chan OrderBookSnapshot, 10)

		// Fan-out routine: reads from rawSnapshotCh and sends snapshots to both diff and recording channels
		go func() {
			for snapshot := range rawSnapshotCh {
				snapshotDiffCh <- snapshot
				snapshotRecCh <- snapshot
			}
		}()

		// Create Recorder instances for each market data type
		tradeRecorder, err := NewRecorder(instrument, "trade", &Trade{}, batchSize)
		if err != nil {
			logger.Errorf("Failed to create trade recorder for %s: %v", instrument, err)
			continue
		}
		aggTradeRecorder, err := NewRecorder(instrument, "aggTrade", &AggTrade{}, batchSize)
		if err != nil {
			logger.Errorf("Failed to create aggTrade recorder for %s: %v", instrument, err)
			continue
		}
		diffRecorder, err := NewRecorder(instrument, "orderBookDiff", &OrderBookDiff{}, batchSize)
		if err != nil {
			logger.Errorf("Failed to create order book diff recorder for %s: %v", instrument, err)
			continue
		}
		bestPriceRecorder, err := NewRecorder(instrument, "bestPrice", &BestPrice{}, batchSize)
		if err != nil {
			logger.Errorf("Failed to create best price recorder for %s: %v", instrument, err)
			continue
		}
		snapshotRecorder, err := NewRecorder(instrument, "snapshot", &OrderBookSnapshot{}, batchSize)
		if err != nil {
			logger.Errorf("Failed to create snapshot recorder for %s: %v", instrument, err)
			continue
		}

		// Define snapshot request callback for order book diff subscription
		snapshotRequest := func() {
			go func() {
				snapshot, err := FetchOrderBookSnapshot(client, instrument)
				if err != nil {
					logger.Errorf("Snapshot request failed for %s: %v", instrument, err)
					return
				}
				// Send the fetched snapshot into the raw snapshot channel
				rawSnapshotCh <- *snapshot
			}()
		}

		// Start Binance WebSocket connections in separate goroutines
		go func(inst string) {
			if err := ListenTrade(ctx, inst, tradeCh); err != nil {
				logger.Errorf("ListenTrade error for %s: %v", inst, err)
				cancel()
			}
		}(instrument)

		go func(inst string) {
			if err := ListenAggTrade(ctx, inst, aggTradeCh); err != nil {
				logger.Errorf("ListenAggTrade error for %s: %v", inst, err)
				cancel()
			}
		}(instrument)

		go func(inst string) {
			if err := ListenOrderBookDiff(ctx, inst, diffCh); err != nil {
				logger.Errorf("ListenOrderBookDiff error for %s: %v", inst, err)
				cancel()
			}
		}(instrument)

		go func(inst string) {
			if err := ListenBestPrice(ctx, inst, bestPriceCh); err != nil {
				logger.Errorf("ListenBestPrice error for %s: %v", inst, err)
				cancel()
			}
		}(instrument)

		// Start REST snapshot fetcher (runs every 1 minute)
		go func(inst string) {
			if err := StartOrderBookSnapshotFetcher(ctx, client, inst, 1*time.Minute, rawSnapshotCh); err != nil {
				logger.Errorf("Snapshot fetcher error for %s: %v", inst, err)
				cancel()
			}
		}(instrument)
		// Start subscription handlers to process incoming messages and record them
		go SubscribeTrades(tradeCh, tradeRecorder, logger)
		go SubscribeAggTrades(aggTradeCh, aggTradeRecorder, logger)
		go SubscribeBestPrice(bestPriceCh, bestPriceRecorder, logger)
		go SubscribeSnapshots(snapshotRecCh, snapshotRecorder, logger)
		go SubscribeOrderBookDiff(diffCh, snapshotDiffCh, diffRecorder, snapshotRequest, logger)
		snapshotRequest()
	}

	// Wait for termination signal
	<-sigChan
	logger.Infof("Shutdown signal received. Cancelling context and closing application.")
	cancel()

	// Allow some time for goroutines to finish (flushing buffers etc.)
	time.Sleep(10 * time.Second)

	os.Exit(0)
}
