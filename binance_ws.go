package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
	"log"
)

// const BASE_STREAM = "stream.binance.com"
const BASE_STREAM = "data-stream.binance.vision"

func safeReadMessage(conn *websocket.Conn) (int, []byte, error) {
	var mt int
	var msg []byte
	var err error
	defer func() {
		if r := recover(); r != nil {
			mt, msg = 0, nil
			err = fmt.Errorf("panic recovered in safeReadMessage: %v", r)
			log.Printf("Recovered in safeReadMessage: %v", r)
		}
	}()
	mt, msg, err = conn.ReadMessage()
	return mt, msg, err
}

// https://github.com/gorilla/websocket/issues/474

// listenWebSocket is a helper function that connects to the given WebSocket URL,
// reads messages in a loop (checking for context cancellation), and calls the handler
// function for each message.
// readResult holds the result of a WebSocket read operation.
type readResult struct {
	mt  int
	msg []byte
	err error
}

// listenWebSocket connects to the given WebSocket URL, then spawns a goroutine
// to continuously read messages, sending them over a channel. The main goroutine
// waits for either context cancellation or messages from that channel.
func listenWebSocket(ctx context.Context, url string, handler func([]byte) error) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to dial websocket %s: %w", url, err)
	}
	log.Printf("Successfully connected to %s", url)
	defer conn.Close()

	readCh := make(chan readResult)

	go func() {
		defer close(readCh)

		for {
			// Blocking read with no deadline
			mt, msg, err := safeReadMessage(conn)

			// If an error occurs, send it down the channel and break out
			if err != nil {
				readCh <- readResult{mt: mt, msg: msg, err: err}
				return
			}

			// Otherwise, send the successfully read message
			readCh <- readResult{mt: mt, msg: msg, err: nil}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// Context canceled; return
			return ctx.Err()

		case rr, ok := <-readCh:
			if !ok {
				return fmt.Errorf("Websocket read goroutine for %s ended unexpectedly", url)
			}

			// If the read result had an error, handle it
			if rr.err != nil {
				log.Printf("Websocket read error: %v", rr.err)
				return rr.err
			}

			// No error, so handle the message
			// log.Printf("Read message: %s", string(rr.msg))
			if err := handler(rr.msg); err != nil {
				log.Printf("handler error: %v", err)
			}
		}
	}
}

// ListenTrade subscribes to Binance trade events for the given symbol using a dedicated WebSocket connection.
// Incoming messages are unmarshaled into Trade structs (defined in binance_types.go) and pushed onto the provided channel.
func ListenTrade(ctx context.Context, symbol string, out chan<- Trade) error {
	url := fmt.Sprintf("wss://%s:9443/ws/%s@trade", BASE_STREAM, strings.ToLower(symbol))
	return listenWebSocket(ctx, url, func(msg []byte) error {
		var combined struct {
			Stream string          `json:"stream"`
			Data   json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(msg, &combined); err == nil && combined.Stream != "" {
			msg = combined.Data
		}
		var trade Trade
		if err := json.Unmarshal(msg, &trade); err != nil {
			return fmt.Errorf("failed to unmarshal Trade: %w, raw message: %s", err, msg)
		}
		if trade.EventType != "trade" {
			return nil
		}
		out <- trade
		return nil
	})
}

// ListenAggTrade subscribes to Binance aggregated trade events for the given symbol.
func ListenAggTrade(ctx context.Context, symbol string, out chan<- AggTrade) error {
	url := fmt.Sprintf("wss://%s:9443/ws/%s@aggTrade", BASE_STREAM, strings.ToLower(symbol))
	return listenWebSocket(ctx, url, func(msg []byte) error {
		var aggTrade AggTrade
		if err := json.Unmarshal(msg, &aggTrade); err != nil {
			return fmt.Errorf("failed to unmarshal AggTrade: %w, raw message: %s", err, msg)
		}
		out <- aggTrade
		return nil
	})
}

// ListenOrderBookDiff subscribes to Binance order book diff events for the given symbol.
func ListenOrderBookDiff(ctx context.Context, symbol string, out chan<- OrderBookDiff) error {
	url := fmt.Sprintf("wss://%s:9443/ws/%s@depth", BASE_STREAM, strings.ToLower(symbol))
	return listenWebSocket(ctx, url, func(msg []byte) error {
		var diff OrderBookDiff
		if err := json.Unmarshal(msg, &diff); err != nil {
			return fmt.Errorf("failed to unmarshal OrderBookDiff: %w, raw message: %s", err, msg)
		}
		out <- diff
		return nil
	})
}

// ListenBestPrice subscribes to Binance best price (book ticker) events for the given symbol.
func ListenBestPrice(ctx context.Context, symbol string, out chan<- BestPrice) error {
	url := fmt.Sprintf("wss://%s:9443/ws/%s@bookTicker", BASE_STREAM, strings.ToLower(symbol))
	return listenWebSocket(ctx, url, func(msg []byte) error {
		var best BestPrice
		if err := json.Unmarshal(msg, &best); err != nil {
			return fmt.Errorf("failed to unmarshal BestPrice: %w, raw message: %s", err, msg)
		}
		out <- best
		return nil
	})
}
