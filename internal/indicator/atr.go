package indicator

import (
	"fmt"
	"math"
)

// ATR implements Average True Range indicator.
type ATR struct {
	name   string
	period int
}

// ATRFactory creates a new ATR indicator from config.
func ATRFactory(config Config) (Indicator, error) {
	if config.Period <= 0 {
		return nil, fmt.Errorf("atr: period must be positive, got %d", config.Period)
	}

	return &ATR{
		name:   "atr",
		period: config.Period,
	}, nil
}

// Name returns the indicator name.
func (a *ATR) Name() string {
	return a.name
}

// Calculate computes the ATR value.
func (a *ATR) Calculate(ctx *Context) (Value, error) {
	// Need at least 'period + 1' candles to calculate true ranges
	if len(ctx.Candles) < a.period+1 {
		return InvalidValue, nil
	}

	// Calculate true ranges for the needed period
	trueRanges := make([]float64, 0, len(ctx.Candles)-1)
	for i := 1; i < len(ctx.Candles); i++ {
		current := ctx.Candles[i]
		previous := ctx.Candles[i-1]

		// True Range = max of:
		// 1. High - Low
		// 2. |High - Previous Close|
		// 3. |Low - Previous Close|
		highLow := current.High - current.Low
		highPrevClose := math.Abs(current.High - previous.Close)
		lowPrevClose := math.Abs(current.Low - previous.Close)

		tr := math.Max(highLow, math.Max(highPrevClose, lowPrevClose))
		trueRanges = append(trueRanges, tr)
	}

	// First ATR: simple average of first 'period' true ranges
	sum := 0.0
	for i := 0; i < a.period; i++ {
		sum += trueRanges[i]
	}
	atr := sum / float64(a.period)

	// If we have more than 'period' true ranges, use smoothed method
	for i := a.period; i < len(trueRanges); i++ {
		// Smoothed ATR: (prev_atr * (period - 1) + current_tr) / period
		atr = (atr*float64(a.period-1) + trueRanges[i]) / float64(a.period)
	}

	return NewValue(atr), nil
}
