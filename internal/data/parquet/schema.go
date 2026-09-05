package parquet

import (
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CandleParquet represents a candle in Parquet format.
// Uses Parquet struct tags for schema definition.
type CandleParquet struct {
	Timestamp int64   `parquet:"name=timestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	Open      float64 `parquet:"name=open, type=DOUBLE"`
	High      float64 `parquet:"name=high, type=DOUBLE"`
	Low       float64 `parquet:"name=low, type=DOUBLE"`
	Close     float64 `parquet:"name=close, type=DOUBLE"`
	Volume    float64 `parquet:"name=volume, type=DOUBLE"`
}

// ToMarketCandle converts CandleParquet to market.Candle.
func (c *CandleParquet) ToMarketCandle() *market.Candle {
	return &market.Candle{
		Timestamp: time.UnixMilli(c.Timestamp),
		Open:      c.Open,
		High:      c.High,
		Low:       c.Low,
		Close:     c.Close,
		Volume:    c.Volume,
	}
}

// FromMarketCandle converts market.Candle to CandleParquet.
func FromMarketCandle(candle *market.Candle) *CandleParquet {
	return &CandleParquet{
		Timestamp: candle.Timestamp.UnixMilli(),
		Open:      candle.Open,
		High:      candle.High,
		Low:       candle.Low,
		Close:     candle.Close,
		Volume:    candle.Volume,
	}
}
