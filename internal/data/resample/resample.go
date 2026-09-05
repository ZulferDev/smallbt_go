package resample

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Resample resamples candles to a different timeframe.
func Resample(candles []*market.Candle, tf market.Timeframe) ([]*market.Candle, error) {
	if len(candles) == 0 {
		return candles, nil
	}

	// For MVP, support 1h and daily without resampling
	// Full implementation would handle resampling from any timeframe
	if tf == market.Timeframe1d || tf == market.Timeframe1h || tf == "" {
		return candles, nil
	}

	return nil, fmt.Errorf("resample to %s not yet implemented", tf)
}
