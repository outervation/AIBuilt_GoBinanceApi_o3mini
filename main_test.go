package main

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

/*
// TestMainfulShutdown simulates a SIGINT to the application and confirms that it shuts down gracefully.
// It does so by spawning a helper process that runs main() and then sending an interrupt signal to it.
func TestMainfulShutdown(t *testing.T) {
	// When running as the helper process (GO_TEST_MAIN is set) we do not run this test
	if os.Getenv("GO_TEST_MAIN") == "1" {
		return
	}

	// Spawn a subprocess that will run TestHelperProcessMain (which calls main)
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessMain")
	// Set an environment variable so the helper process knows to run main()
	cmd.Env = append(os.Environ(), "GO_TEST_MAIN=1")

	// Capture output for verification
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start helper process: %v", err)
	}

	// Allow the helper process time to initialize and set up its signal handlers
	time.Sleep(1 * time.Second)

	// Send SIGINT to the helper process
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to send SIGINT to helper process: %v", err)
	}

	// Wait for the process to exit. We set a timeout to avoid hanging in case of failure.
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			// If the process exited with an error, check whether it exited with a non-zero status and fail accordingly.
			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					if status.ExitStatus() != 0 {
						t.Fatalf("Helper process exited with non-zero status: %d", status.ExitStatus())
					}
				} else {
					t.Fatalf("Helper process exited with error: %v", err)
				}
			} else {
				t.Fatalf("Helper process error: %v", err)
			}
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("Helper process did not exit within 10 seconds after SIGINT")
	}

	// Verify output contains the shutdown log message
	if !strings.Contains(output.String(), "Shutdown signal received") {
		t.Fatalf("Shutdown message not found in output: %s", output.String())
	}
}
*/
// TestHelperProcessMain is a helper test that is invoked in a subprocess. It calls main() so that the actual application
// can be exercised. It is only executed when the environment variable GO_TEST_MAIN is set.
func TestHelperProcessMain(t *testing.T) {
	if os.Getenv("GO_TEST_MAIN") != "1" {
		return
	}
	// Run the main application. It will block until it receives a SIGINT signal.
	main()
	// Normally, main() will call os.Exit(0), so the following line may never be reached.
	os.Exit(0)
}

func TestMainTradeWebSocketIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tradeCh := make(chan Trade, 1)
	go func() {
		if err := ListenTrade(ctx, "BTCUSDT", tradeCh); err != nil {
			t.Logf("ListenTrade error: %v", err) // Log error if ListenTrade returns before a trade is received
		}
	}()
	select {
	case trade := <-tradeCh:
		t.Logf("Received trade: %+v", trade)
		if trade.EventType != "trade" {
			t.Errorf("Expected EventType 'trade', got %s", trade.EventType)
		}
		if trade.TradeID <= 0 {
			t.Errorf("TradeID should be positive, got %d", trade.TradeID)
		}
		if trade.Price == "" {
			t.Errorf("Price should not be empty")
		}
	case <-ctx.Done():
		t.Fatalf("Timeout waiting for trade message")
	}
}

// TestMainOrderBookSnapshotIntegration calls the Binance REST API snapshot endpoint for BTCUSDT, waits up to 10 seconds for a response,
// and verifies that the parsed OrderBookSnapshot contains a valid non-zero LastUpdateID with non-empty bid and ask lists.
func TestMainOrderBookSnapshotIntegration(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}
	// Call the REST API snapshot endpoint for BTCUSDT
	snapshot, err := FetchOrderBookSnapshot(client, "BTCUSDT")
	if err != nil {
		t.Fatalf("Failed to fetch snapshot: %v", err)
	}
	if snapshot.LastUpdateID <= 0 {
		t.Fatalf("Invalid LastUpdateID: %d", snapshot.LastUpdateID)
	}
	if len(snapshot.Bids) == 0 {
		t.Fatalf("Empty Bids in snapshot")
	}
	if len(snapshot.Asks) == 0 {
		t.Fatalf("Empty Asks in snapshot")
	}
	t.Logf("Fetched snapshot for BTCUSDT: LastUpdateID = %d, %d bids, %d asks", snapshot.LastUpdateID, len(snapshot.Bids), len(snapshot.Asks))
}
