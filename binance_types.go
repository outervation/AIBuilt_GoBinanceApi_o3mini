package main

import (
	"encoding/json"
	"fmt"
)

// binance_types.go defines all Binance market data types used for recording market data.
// Each type includes the necessary parquet tags for serialization using github.com/xitongsys/parquet-go.

// Trade represents a single trade event from Binance.
// It contains fields like event type, event time, trade ID, price, quantity, buyer/seller order IDs, trade time, and a flag indicating if the buyer was the market maker.
type Trade struct {
	EventType     string `json:"e" parquet:"name=event_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	EventTime     int64  `json:"E" parquet:"name=event_time, type=INT64"`
	TradeID       int64  `json:"t" parquet:"name=trade_id, type=INT64"`
	Price         string `json:"p" parquet:"name=price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Quantity      string `json:"q" parquet:"name=quantity, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	BuyerOrderID  int64  `json:"b" parquet:"name=buyer_order_id, type=INT64"`
	SellerOrderID int64  `json:"a" parquet:"name=seller_order_id, type=INT64"`
	TradeTime     int64  `json:"T" parquet:"name=trade_time, type=INT64"`
	IsBuyerMaker  bool   `json:"m" parquet:"name=is_buyer_maker, type=BOOLEAN"`
}

// AggTrade represents an aggregated trade event from Binance.
// It includes the event type, event time, symbol, aggregated trade ID, price, quantity, first and last trade IDs, trade time, and buyer maker indicator.
type AggTrade struct {
	EventType    string `json:"e" parquet:"name=event_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	EventTime    int64  `json:"E" parquet:"name=event_time, type=INT64"`
	Symbol       string `json:"s" parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	AggTradeID   int64  `json:"a" parquet:"name=agg_trade_id, type=INT64"`
	Price        string `json:"p" parquet:"name=price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Quantity     string `json:"q" parquet:"name=quantity, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	FirstTradeID int64  `json:"f" parquet:"name=first_trade_id, type=INT64"`
	LastTradeID  int64  `json:"l" parquet:"name=last_trade_id, type=INT64"`
	TradeTime    int64  `json:"T" parquet:"name=trade_time, type=INT64"`
	IsBuyerMaker bool   `json:"m" parquet:"name=is_buyer_maker, type=BOOLEAN"`
}

// PriceLevel represents a price level entry in the order book with a price and its associated quantity.
type PriceLevel struct {
	Price    string `parquet:"name=price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Quantity string `parquet:"name=quantity, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

// OrderBookDiff represents a differential update to the order book as received from Binance.
type OrderBookDiff struct {
	EventType     string       `json:"e" parquet:"name=event_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	EventTime     int64        `json:"E" parquet:"name=event_time, type=INT64"`
	Symbol        string       `json:"s" parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	FirstUpdateID int64        `json:"U" parquet:"name=first_update_id, type=INT64"`
	FinalUpdateID int64        `json:"u" parquet:"name=final_update_id, type=INT64"`
	Bids          []PriceLevel `json:"b" parquet:"name=bids, repetitiontype=REPEATED"`
	Asks          []PriceLevel `json:"a" parquet:"name=asks, repetitiontype=REPEATED"`
}
type BestPrice struct {
	EventType string `json:"e" parquet:"name=event_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	UpdateID  int64  `json:"u" parquet:"name=update_id, type=INT64"`
	Symbol    string `json:"s" parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	BidPrice  string `json:"b" parquet:"name=bid_price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	BidQty    string `json:"B" parquet:"name=bid_qty, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	AskPrice  string `json:"a" parquet:"name=ask_price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	AskQty    string `json:"A" parquet:"name=ask_qty, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

// OrderBookSnapshot represents a full snapshot of the order book as obtained via Binance's REST API.
// It includes the last update ID and the complete list of bid and ask price levels.
type OrderBookSnapshot struct {
	LastUpdateID int64        `parquet:"name=last_update_id, type=INT64"`
	Bids         []PriceLevel `parquet:"name=bids, repetitiontype=REPEATED"`
	Asks         []PriceLevel `parquet:"name=asks, repetitiontype=REPEATED"`
}

func (p *PriceLevel) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) != 2 {
		return fmt.Errorf("unexpected length for PriceLevel: %d", len(arr))
	}
	p.Price = arr[0]
	p.Quantity = arr[1]
	return nil
}
