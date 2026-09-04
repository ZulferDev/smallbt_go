package market

import (
	"time"
)

// Symbol represents a trading symbol.
type Symbol string

// Timeframe represents a trading timeframe.
type Timeframe string

const (
	Timeframe1m  Timeframe = "1m"
	Timeframe5m  Timeframe = "5m"
	Timeframe15m Timeframe = "15m"
	Timeframe30m Timeframe = "30m"
	Timeframe1h  Timeframe = "1h"
	Timeframe4h  Timeframe = "4h"
	Timeframe1d  Timeframe = "1d"
	Timeframe1w  Timeframe = "1w"
	Timeframe1mo Timeframe = "1mo"
)

// Candle represents an OHLCV candle.
type Candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// IsValid checks if candle has valid OHLC relationships.
func (c *Candle) IsValid() bool {
	if c.High < c.Low || c.High < c.Open || c.High < c.Close {
		return false
	}
	if c.Low > c.Open || c.Low > c.Close {
		return false
	}
	if c.Volume < 0 || c.Open < 0 || c.High < 0 || c.Low < 0 || c.Close < 0 {
		return false
	}
	return true
}

// MarketData represents market data for a symbol at a timeframe.
type MarketData struct {
	Symbol    Symbol
	Timeframe Timeframe
	Candles   []Candle
}

// NewMarketData creates a new market data container.
func NewMarketData(symbol Symbol, timeframe Timeframe) *MarketData {
	return &MarketData{
		Symbol:    symbol,
		Timeframe: timeframe,
		Candles:   make([]Candle, 0),
	}
}

// AddCandle adds a candle to the market data.
func (md *MarketData) AddCandle(candle Candle) {
	md.Candles = append(md.Candles, candle)
}

// GetLatest returns the most recent candle.
func (md *MarketData) GetLatest() *Candle {
	if len(md.Candles) == 0 {
		return nil
	}
	return &md.Candles[len(md.Candles)-1]
}

// Length returns the number of candles.
func (md *MarketData) Length() int {
	return len(md.Candles)
}
