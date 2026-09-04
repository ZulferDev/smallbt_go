package resample

import (
	"fmt"

	"github.com/1jehuang/backtest/internal/market"
)

// Resample resamples candles to a different timeframe.
func Resample(candles []*market.Candle, tf market.Timeframe) ([]*market.Candle, error) {
	if len(candles) == 0 {
		return candles, nil
	}

	// For MVP, only support daily (no resampling needed)
	// Full implementation would handle 1h, 4h, 1d, etc.
	if tf == market.Timeframe1d || tf == "" {
		return candles, nil
	}

	return nil, fmt.Errorf("resample to %s not yet implemented", tf)
}
