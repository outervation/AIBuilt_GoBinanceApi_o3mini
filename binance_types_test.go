package main

import (
	"reflect"
	"testing"
)

func TestFieldTags(t *testing.T) {
	expectedTags := map[string]string{
		"EventType":     "name=event_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"EventTime":     "name=event_time, type=INT64",
		"TradeID":       "name=trade_id, type=INT64",
		"Price":         "name=price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"Quantity":      "name=quantity, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"BuyerOrderID":  "name=buyer_order_id, type=INT64",
		"SellerOrderID": "name=seller_order_id, type=INT64",
		"TradeTime":     "name=trade_time, type=INT64",
		"IsBuyerMaker":  "name=is_buyer_maker, type=BOOLEAN",
	}

	tradeType := reflect.TypeOf(Trade{})
	for i := 0; i < tradeType.NumField(); i++ {
		field := tradeType.Field(i)
		tag := field.Tag.Get("parquet")
		expected, ok := expectedTags[field.Name]
		if !ok {
			t.Errorf("Unexpected field %q in Trade struct", field.Name)
		} else if tag != expected {
			t.Errorf("Field %q: expected tag %q, got %q", field.Name, expected, tag)
		}
	}
}

func TestAggTradeFieldTags(t *testing.T) {
	expectedTags := map[string]string{
		"EventType":    "name=event_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"EventTime":    "name=event_time, type=INT64",
		"Symbol":       "name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"AggTradeID":   "name=agg_trade_id, type=INT64",
		"Price":        "name=price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"Quantity":     "name=quantity, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"FirstTradeID": "name=first_trade_id, type=INT64",
		"LastTradeID":  "name=last_trade_id, type=INT64",
		"TradeTime":    "name=trade_time, type=INT64",
		"IsBuyerMaker": "name=is_buyer_maker, type=BOOLEAN",
	}
	aggTradeType := reflect.TypeOf(AggTrade{})
	for i := 0; i < aggTradeType.NumField(); i++ {
		field := aggTradeType.Field(i)
		tag := field.Tag.Get("parquet")
		expected, ok := expectedTags[field.Name]
		if !ok {
			t.Errorf("Unexpected field %q in AggTrade struct", field.Name)
		} else if tag != expected {
			t.Errorf("Field %q: expected tag %q, got %q", field.Name, expected, tag)
		}
	}
}

func TestOrderBookDiffAndPriceLevelTags(t *testing.T) {
	expectedDiffTags := map[string]string{
		"EventType":     "name=event_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"EventTime":     "name=event_time, type=INT64",
		"Symbol":        "name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"FirstUpdateID": "name=first_update_id, type=INT64",
		"FinalUpdateID": "name=final_update_id, type=INT64",
		"Bids":          "name=bids, repetitiontype=REPEATED",
		"Asks":          "name=asks, repetitiontype=REPEATED",
	}
	diffType := reflect.TypeOf(OrderBookDiff{})
	for i := 0; i < diffType.NumField(); i++ {
		field := diffType.Field(i)
		tag := field.Tag.Get("parquet")
		exp, ok := expectedDiffTags[field.Name]
		if !ok {
			t.Errorf("Unexpected field %q in OrderBookDiff struct", field.Name)
		} else if tag != exp {
			t.Errorf("Field %q: expected tag %q, got %q", field.Name, exp, tag)
		}
	}

	expectedPriceLevelTags := map[string]string{
		"Price":    "name=price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"Quantity": "name=quantity, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
	}
	priceLevelType := reflect.TypeOf(PriceLevel{})
	for i := 0; i < priceLevelType.NumField(); i++ {
		field := priceLevelType.Field(i)
		tag := field.Tag.Get("parquet")
		exp, ok := expectedPriceLevelTags[field.Name]
		if !ok {
			t.Errorf("Unexpected field %q in PriceLevel struct", field.Name)
		} else if tag != exp {
			t.Errorf("Field %q in PriceLevel: expected tag %q, got %q", field.Name, exp, tag)
		}
	}
}

func TestBestPriceFieldTags(t *testing.T) {
	expectedTags := map[string]string{
		"EventType": "name=event_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"UpdateID":  "name=update_id, type=INT64",
		"Symbol":    "name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"BidPrice":  "name=bid_price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"BidQty":    "name=bid_qty, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"AskPrice":  "name=ask_price, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
		"AskQty":    "name=ask_qty, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY",
	}
	bestPriceType := reflect.TypeOf(BestPrice{})
	for i := 0; i < bestPriceType.NumField(); i++ {
		field := bestPriceType.Field(i)
		tag := field.Tag.Get("parquet")
		expected, ok := expectedTags[field.Name]
		if !ok {
			t.Errorf("Unexpected field %q in BestPrice struct", field.Name)
		} else if tag != expected {
			t.Errorf("Field %q: expected tag %q, got %q", field.Name, expected, tag)
		}
	}
}

func TestOrderBookSnapshotFieldTags(t *testing.T) {
	expectedTags := map[string]string{
		"LastUpdateID": "name=last_update_id, type=INT64",
		"Bids":         "name=bids, repetitiontype=REPEATED",
		"Asks":         "name=asks, repetitiontype=REPEATED",
	}

	snapshotType := reflect.TypeOf(OrderBookSnapshot{})
	for i := 0; i < snapshotType.NumField(); i++ {
		field := snapshotType.Field(i)
		tag := field.Tag.Get("parquet")
		expected, ok := expectedTags[field.Name]
		if !ok {
			t.Errorf("Unexpected field %q in OrderBookSnapshot struct", field.Name)
		} else if tag != expected {
			t.Errorf("Field %q: expected tag %q, got %q", field.Name, expected, tag)
		}
	}
}
