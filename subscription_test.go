package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// FakeRecorder is a mock recorder for testing aggregated trade subscriptions.
type FakeRecorder struct {
	records []AggTrade
	mu      sync.Mutex
}

func (fr *FakeRecorder) Write(record interface{}) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	agg, ok := record.(AggTrade)
	if !ok {
		return fmt.Errorf("expected AggTrade type, got %T", record)
	}
	fr.records = append(fr.records, agg)
	return nil
}

func (fr *FakeRecorder) GetRecords() []AggTrade {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	return fr.records
}

// FakeLogger is a stub logger that does nothing.
type FakeLogger struct{}

func (fl *FakeLogger) Errorf(format string, args ...interface{}) error {
	fmt.Printf("FakeLogger ERROR: "+format+"\n", args...)
	return nil
}

func (fl *FakeLogger) Infof(format string, args ...interface{}) error {
	fmt.Printf("FakeLogger INFO: "+format+"\n", args...)
	return nil
}

func TestSubscribeAggTrades_WritesAggTradeRecords(t *testing.T) {
	// Create a buffered channel for aggregated trades.
	aggTradeCh := make(chan AggTrade, 5)
	fakeRecorder := &FakeRecorder{}
	fakeLogger := &FakeLogger{}

	// Run SubscribeAggTrades in a separate goroutine (it will exit when the channel is closed).
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		SubscribeAggTrades(aggTradeCh, fakeRecorder, fakeLogger)
	}()

	// Prepare some test aggregated trade messages.
	testAggTrades := []AggTrade{
		{
			EventType:    "aggTrade",
			EventTime:    1000,
			Symbol:       "BTCUSDT",
			AggTradeID:   1,
			Price:        "10000",
			Quantity:     "0.5",
			FirstTradeID: 10,
			LastTradeID:  11,
			TradeTime:    1001,
			IsBuyerMaker: true,
		},
		{
			EventType:    "aggTrade",
			EventTime:    2000,
			Symbol:       "BTCUSDT",
			AggTradeID:   2,
			Price:        "10100",
			Quantity:     "0.3",
			FirstTradeID: 12,
			LastTradeID:  13,
			TradeTime:    2001,
			IsBuyerMaker: false,
		},
	}

	// Send the test messages.
	for _, trade := range testAggTrades {
		aggTradeCh <- trade
	}
	close(aggTradeCh)
	wg.Wait()

	// Verify that the fake recorder captured all sent aggregated trades.
	recorded := fakeRecorder.GetRecords()
	if len(recorded) != len(testAggTrades) {
		t.Errorf("expected %d records, got %d", len(testAggTrades), len(recorded))
	}
	for i, expected := range testAggTrades {
		actual := recorded[i]
		if actual.AggTradeID != expected.AggTradeID || actual.Price != expected.Price {
			t.Errorf("record mismatch at index %d: expected %+v, got %+v", i, expected, actual)
		}
	}
}

// FakeBestPriceRecorder is a mock recorder for testing best price subscriptions.
type FakeBestPriceRecorder struct {
	records []BestPrice
	mu      sync.Mutex
}

func (f *FakeBestPriceRecorder) Write(record interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	bp, ok := record.(BestPrice)
	if !ok {
		return fmt.Errorf("expected BestPrice type, got %T", record)
	}
	f.records = append(f.records, bp)
	return nil
}

func (f *FakeBestPriceRecorder) GetRecords() []BestPrice {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.records
}

func TestSubscribeBestPrice_WritesBestPriceRecords(t *testing.T) {
	bestPriceCh := make(chan BestPrice, 5)
	fakeRecorder := &FakeBestPriceRecorder{}
	fakeLogger := &FakeLogger{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		SubscribeBestPrice(bestPriceCh, fakeRecorder, fakeLogger)
	}()

	testBestPrices := []BestPrice{
		{
			EventType: "bookTicker",
			UpdateID:  1000,
			Symbol:    "BTCUSDT",
			BidPrice:  "50000",
			BidQty:    "1",
			AskPrice:  "50010",
			AskQty:    "0.5",
		},
		{
			EventType: "bookTicker",
			UpdateID:  2000,
			Symbol:    "ETHUSDT",
			BidPrice:  "2000",
			BidQty:    "2",
			AskPrice:  "2010",
			AskQty:    "1.5",
		},
	}

	for _, bp := range testBestPrices {
		bestPriceCh <- bp
	}
	close(bestPriceCh)
	wg.Wait()

	recorded := fakeRecorder.GetRecords()
	t.Logf("Recorded best prices: %+v", recorded)
	if len(recorded) != len(testBestPrices) {
		t.Errorf("expected %d records, got %d", len(testBestPrices), len(recorded))
	}

	for i, expected := range testBestPrices {
		actual := recorded[i]
		if actual.UpdateID != expected.UpdateID || actual.Symbol != expected.Symbol ||
			actual.BidPrice != expected.BidPrice || actual.AskPrice != expected.AskPrice {
			t.Errorf("record mismatch at index %d: expected %+v, got %+v", i, expected, actual)
		}
	}
}

// FakeSnapshotRecorder is a mock recorder for testing snapshot subscriptions.

type FakeSnapshotRecorder struct {
	records []OrderBookSnapshot
	mu      sync.Mutex
}

func (fsr *FakeSnapshotRecorder) Write(record interface{}) error {
	fsr.mu.Lock()
	defer fsr.mu.Unlock()

	snapshot, ok := record.(OrderBookSnapshot)
	if !ok {
		return fmt.Errorf("expected OrderBookSnapshot type, got %T", record)
	}
	fsr.records = append(fsr.records, snapshot)
	return nil
}

func (fsr *FakeSnapshotRecorder) GetRecords() []OrderBookSnapshot {
	fsr.mu.Lock()
	defer fsr.mu.Unlock()
	return fsr.records
}

func TestSubscribeSnapshots_WritesSnapshotRecords(t *testing.T) {
	// Create a buffered channel for OrderBookSnapshot messages
	snapshotCh := make(chan OrderBookSnapshot, 5)
	fakeRecorder := &FakeSnapshotRecorder{}
	fakeLogger := &FakeLogger{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		SubscribeSnapshots(snapshotCh, fakeRecorder, fakeLogger)
	}()

	// Prepare some test snapshot messages
	snapshot1 := OrderBookSnapshot{
		LastUpdateID: 1000,
		Bids:         []PriceLevel{{Price: "50000", Quantity: "1.0"}},
		Asks:         []PriceLevel{{Price: "50005", Quantity: "0.5"}},
	}
	snapshot2 := OrderBookSnapshot{
		LastUpdateID: 2000,
		Bids:         []PriceLevel{{Price: "51000", Quantity: "2.0"}},
		Asks:         []PriceLevel{{Price: "51005", Quantity: "1.5"}},
	}

	// Send the snapshots on the channel
	snapshotCh <- snapshot1
	snapshotCh <- snapshot2
	close(snapshotCh)

	wg.Wait()

	// Validate that the recorder captured both snapshots
	records := fakeRecorder.GetRecords()
	if len(records) != 2 {
		t.Errorf("expected 2 snapshot records, got %d", len(records))
	}

	if records[0].LastUpdateID != snapshot1.LastUpdateID {
		t.Errorf("first snapshot LastUpdateID expected %d, got %d", snapshot1.LastUpdateID, records[0].LastUpdateID)
	}
	if len(records[0].Bids) != len(snapshot1.Bids) || records[0].Bids[0].Price != snapshot1.Bids[0].Price {
		t.Errorf("first snapshot bid mismatch: expected %v, got %v", snapshot1.Bids, records[0].Bids)
	}
	if len(records[0].Asks) != len(snapshot1.Asks) || records[0].Asks[0].Price != snapshot1.Asks[0].Price {
		t.Errorf("first snapshot ask mismatch: expected %v, got %v", snapshot1.Asks, records[0].Asks)
	}

	if records[1].LastUpdateID != snapshot2.LastUpdateID {
		t.Errorf("second snapshot LastUpdateID expected %d, got %d", snapshot2.LastUpdateID, records[1].LastUpdateID)
	}
	if len(records[1].Bids) != len(snapshot2.Bids) || records[1].Bids[0].Price != snapshot2.Bids[0].Price {
		t.Errorf("second snapshot bid mismatch: expected %v, got %v", snapshot2.Bids, records[1].Bids)
	}
	if len(records[1].Asks) != len(snapshot2.Asks) || records[1].Asks[0].Price != snapshot2.Asks[0].Price {
		t.Errorf("second snapshot ask mismatch: expected %v, got %v", snapshot2.Asks, records[1].Asks)
	}
}

// FakeDiffRecorder is a mock recorder for OrderBookDiff events for testing.
type FakeDiffRecorder struct {
	records []OrderBookDiff
	mu      sync.Mutex
}

func (f *FakeDiffRecorder) Write(record interface{}) error {
	diff, ok := record.(OrderBookDiff)
	if !ok {
		return fmt.Errorf("expected OrderBookDiff, got %T", record)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.records = append(f.records, diff)
	return nil
}

func (f *FakeDiffRecorder) GetRecords() []OrderBookDiff {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.records
}

// TestSubscribeOrderBookDiff_SequenceGapDetection simulates a gap in diff message sequence so that SubscribeOrderBookDiff detects the gap,
// triggers the snapshotRequest callback, resets state, and refrains from recording the problematic diff.
func TestSubscribeOrderBookDiff_SequenceGapDetection(t *testing.T) {
	diffCh := make(chan OrderBookDiff, 10)
	snapshotCh := make(chan OrderBookSnapshot, 1)

	fakeDiffRecorder := &FakeDiffRecorder{}
	fakeLogger := &FakeLogger{}

	var snapshotRequestCalled int
	snapshotRequest := func() {
		snapshotRequestCalled++
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		SubscribeOrderBookDiff(diffCh, snapshotCh, fakeDiffRecorder, snapshotRequest, fakeLogger)
	}()

	// Send a snapshot message with LastUpdateID = 100
	snapshotCh <- OrderBookSnapshot{
		LastUpdateID: 100,
		Bids:         []PriceLevel{{Price: "50000", Quantity: "1"}},
		Asks:         []PriceLevel{{Price: "50010", Quantity: "1"}},
	}

	time.Sleep(10 * time.Millisecond)

	// Send a valid diff message with FirstUpdateID = 101 and FinalUpdateID = 101
	validDiff := OrderBookDiff{
		EventType:     "depthUpdate",
		EventTime:     1500,
		Symbol:        "BTCUSDT",
		FirstUpdateID: 101,
		FinalUpdateID: 101,
		Bids:          []PriceLevel{{Price: "50000", Quantity: "0.8"}},
		Asks:          []PriceLevel{{Price: "50010", Quantity: "0.8"}},
	}
	diffCh <- validDiff

	time.Sleep(10 * time.Millisecond)

	// Send a diff message with a gap: expected FirstUpdateID should be 102, but here it is 103, simulating a gap
	gapDiff := OrderBookDiff{
		EventType:     "depthUpdate",
		EventTime:     2000,
		Symbol:        "BTCUSDT",
		FirstUpdateID: 103,
		FinalUpdateID: 103,
		Bids:          []PriceLevel{{Price: "50000", Quantity: "0.5"}},
		Asks:          []PriceLevel{{Price: "50010", Quantity: "0.5"}},
	}
	diffCh <- gapDiff

	time.Sleep(10 * time.Millisecond)

	close(snapshotCh)
	close(diffCh)
	wg.Wait()

	records := fakeDiffRecorder.GetRecords()
	if len(records) != 1 {
		t.Errorf("Expected 1 recorded diff, got %d", len(records))
	} else if records[0].FinalUpdateID != 101 {
		t.Errorf("Recorded diff has wrong FinalUpdateID, expected 101, got %d", records[0].FinalUpdateID)
	}

	if snapshotRequestCalled != 1 {
		t.Errorf("Expected snapshotRequest to be called once, but got %d", snapshotRequestCalled)
	}
}

func TestSubscribeOrderBookDiff_IgnoreOutdatedDiff(t *testing.T) {
	diffCh := make(chan OrderBookDiff, 1)
	snapshotCh := make(chan OrderBookSnapshot, 1)
	fakeDiffRecorder := &FakeDiffRecorder{}
	fakeLogger := &FakeLogger{}

	snapshotRequestCalled := 0
	snapshotRequest := func() {
		snapshotRequestCalled++
	}

	done := make(chan struct{})
	go func() {
		SubscribeOrderBookDiff(diffCh, snapshotCh, fakeDiffRecorder, snapshotRequest, fakeLogger)
		close(done)
	}()

	// Send a snapshot to initialize the state with LastUpdateID = 100
	snapshotCh <- OrderBookSnapshot{
		LastUpdateID: 100,
		Bids:         []PriceLevel{},
		Asks:         []PriceLevel{},
	}
	time.Sleep(50 * time.Millisecond)

	// Send an outdated diff message whose FinalUpdateID (100) does not exceed the snapshot's LastUpdateID
	outdatedDiff := OrderBookDiff{
		EventType:     "depthUpdate",
		EventTime:     1500,
		Symbol:        "BTCUSDT",
		FirstUpdateID: 100,
		FinalUpdateID: 100, // not greater than snapshot's 100
		Bids:          []PriceLevel{{Price: "50000", Quantity: "1"}},
		Asks:          []PriceLevel{{Price: "50010", Quantity: "1"}},
	}
	diffCh <- outdatedDiff
	time.Sleep(50 * time.Millisecond)

	// Close channels to terminate the subscriber
	close(snapshotCh)
	close(diffCh)
	<-done

	if len(fakeDiffRecorder.GetRecords()) != 0 {
		t.Errorf("Expected no diff records to be recorded, but got %d", len(fakeDiffRecorder.GetRecords()))
	}
	if snapshotRequestCalled != 0 {
		t.Errorf("Expected snapshotRequest not to be called, but it was called %d times", snapshotRequestCalled)
	}
}
