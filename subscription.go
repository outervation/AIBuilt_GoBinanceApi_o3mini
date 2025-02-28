package main

// RecorderWriter defines the minimal interface for writing records.
type RecorderWriter interface {
	Write(record interface{}) error
}

// LoggerInterface defines the minimal interface for logging required by subscription functions.
type LoggerInterface interface {
	Errorf(format string, args ...interface{}) error
	Infof(format string, args ...interface{}) error
}

// ProcessOrderBookDiffMessage processes an OrderBookDiff message based on the current snapshot's LastUpdateID and the last processed diff update ID.
// It returns whether the diff should be recorded, the new last processed update ID, and whether a sequence gap was detected.
func ProcessOrderBookDiffMessage(diff OrderBookDiff, lastSnapshotId, lastProcessedId int64) (bool, int64, bool) {
	if diff.FinalUpdateID <= lastSnapshotId {
		return false, lastProcessedId, false
	}
	if lastProcessedId == lastSnapshotId {
		if diff.FirstUpdateID > lastSnapshotId+1 {
			return false, lastProcessedId, true
		}
	} else {
		if diff.FirstUpdateID != lastProcessedId+1 {
			return false, lastProcessedId, true
		}
	}
	return true, diff.FinalUpdateID, false
}

// SubscribeTrades listens to the trade channel and writes each Trade to the provided RecorderWriter.
func SubscribeTrades(tradeCh <-chan Trade, recorder RecorderWriter, logger LoggerInterface) {
	for trade := range tradeCh {
		if err := recorder.Write(trade); err != nil {
			logger.Errorf("error writing trade: %v", err)
		}
	}
}

// SubscribeAggTrades listens to the aggregated trade channel and writes each AggTrade to the provided RecorderWriter.
func SubscribeAggTrades(aggTradeCh <-chan AggTrade, recorder RecorderWriter, logger LoggerInterface) {
	for aggTrade := range aggTradeCh {
		if err := recorder.Write(aggTrade); err != nil {
			logger.Errorf("error writing aggregated trade: %v", err)
		}
	}
}

// SubscribeBestPrice listens to the best price channel and writes each BestPrice to the provided RecorderWriter.
func SubscribeBestPrice(bestPriceCh <-chan BestPrice, recorder RecorderWriter, logger LoggerInterface) {
	for bestPrice := range bestPriceCh {
		if err := recorder.Write(bestPrice); err != nil {
			logger.Errorf("error writing best price: %v", err)
		}
	}
}

// SubscribeSnapshots listens to the order book snapshot channel and writes each OrderBookSnapshot to the provided RecorderWriter.
func SubscribeSnapshots(snapshotCh <-chan OrderBookSnapshot, recorder RecorderWriter, logger LoggerInterface) {
	for snapshot := range snapshotCh {
		if err := recorder.Write(snapshot); err != nil {
			logger.Errorf("error writing order book snapshot: %v", err)
		}
	}
}

// SubscribeOrderBookDiff listens to the order book diff channel alongside the snapshot channel.
// It applies filtering rules to ensure that outdated diff messages are discarded and sequence gaps trigger a new snapshot request.
func SubscribeOrderBookDiff(diffCh <-chan OrderBookDiff, snapshotCh <-chan OrderBookSnapshot, diffRecorder RecorderWriter, snapshotRequest func(), logger LoggerInterface) {
	var lastSnapshotId int64 = 0
	var lastProcessedId int64 = 0
	for {
		select {
		case snapshot, ok := <-snapshotCh:
			if !ok {
				logger.Errorf("snapshot channel closed")
				return
			}
			lastSnapshotId = snapshot.LastUpdateID
			lastProcessedId = snapshot.LastUpdateID
			logger.Infof("Received new snapshot with LastUpdateID: %d", lastSnapshotId)
		case diff, ok := <-diffCh:
			if !ok {
				logger.Errorf("order book diff channel closed")
				return
			}
			if lastSnapshotId == 0 {
				logger.Infof("No snapshot received yet; skipping diff message with FinalUpdateID: %d", diff.FinalUpdateID)
				continue
			}
			recordMsg, newProcessedId, gapDetected := ProcessOrderBookDiffMessage(diff, lastSnapshotId, lastProcessedId)
			if gapDetected {
				logger.Errorf("Sequence gap detected: expected %d but got %d. Triggering new snapshot request.", lastProcessedId+1, diff.FirstUpdateID)
				snapshotRequest()
				lastSnapshotId = 0
				lastProcessedId = 0
				continue
			}
			if recordMsg {
				if err := diffRecorder.Write(diff); err != nil {
					logger.Errorf("error writing order book diff: %v", err)
				}
				lastProcessedId = newProcessedId
			} else {
				logger.Infof("Discarded outdated diff with FinalUpdateID: %d (Snapshot LastUpdateID: %d)", diff.FinalUpdateID, lastSnapshotId)
			}
		}
	}
}
