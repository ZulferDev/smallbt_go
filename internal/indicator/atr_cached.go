package indicator

import (
	"fmt"
	"math"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CachedATR implements ATR with incremental state-based calculation.
// This provides O(1) per-bar performance vs O(n) for the stateless version.
type CachedATR struct {
	name   string
	period int

	// State
	currentATR float64
	prevCandle *market.Candle
	barsSeen   int
	isWarm     bool

	// Warmup buffer for initial ATR calculation
	trBuffer []float64
}

// NewCachedATR creates a new cached ATR indicator.
func NewCachedATR(period int) (*CachedATR, error) {
	if period <= 0 {
		return nil, fmt.Errorf("atr: period must be positive, got %d", period)
	}

	return &CachedATR{
		name:     "atr",
		period:   period,
		trBuffer: make([]float64, 0, period),
	}, nil
}

// CachedATRFactory creates a new cached ATR indicator from config.
func CachedATRFactory(config Config) (CachedIndicator, error) {
	if config.Period <= 0 {
		return nil, fmt.Errorf("atr: period must be positive, got %d", config.Period)
	}

	return NewCachedATR(config.Period)
}

// Name returns the indicator name.
func (a *CachedATR) Name() string {
	return a.name
}

// Calculate implements the stateless Indicator interface for backward compatibility.
// This still does full recalculation but allows CachedATR to satisfy both interfaces.
func (a *CachedATR) Calculate(ctx *Context) (Value, error) {
	// Fallback to stateless calculation
	// In practice, Update() should be used instead for performance
	if len(ctx.Candles) < a.period+1 {
		return InvalidValue, nil
	}

	trueRanges := make([]float64, 0, len(ctx.Candles)-1)
	for i := 1; i < len(ctx.Candles); i++ {
		current := ctx.Candles[i]
		previous := ctx.Candles[i-1]

		tr := calculateTrueRange(&current, &previous)
		trueRanges = append(trueRanges, tr)
	}

	sum := 0.0
	for i := 0; i < a.period; i++ {
		sum += trueRanges[i]
	}
	atr := sum / float64(a.period)

	for i := a.period; i < len(trueRanges); i++ {
		atr = (atr*float64(a.period-1) + trueRanges[i]) / float64(a.period)
	}

	return NewValue(atr), nil
}

// Update incrementally updates the ATR with a new candle.
// This is the high-performance path: O(1) instead of O(n).
func (a *CachedATR) Update(candle market.Candle, prevCandle *market.Candle) (Value, error) {
	// First bar: just store it, no TR yet
	if prevCandle == nil {
		a.prevCandle = &candle
		a.barsSeen = 1
		return InvalidValue, nil
	}

	// Calculate true range for this bar
	tr := calculateTrueRange(&candle, prevCandle)

	if !a.isWarm {
		// Warmup phase: collect TRs until we have 'period' of them
		a.trBuffer = append(a.trBuffer, tr)
		a.barsSeen++

		if len(a.trBuffer) >= a.period {
			// Calculate initial ATR: simple average of first 'period' TRs
			sum := 0.0
			for _, val := range a.trBuffer {
				sum += val
			}
			a.currentATR = sum / float64(a.period)
			a.isWarm = true

			// Clear buffer to free memory
			a.trBuffer = nil

			a.prevCandle = &candle
			return NewValue(a.currentATR), nil
		}

		a.prevCandle = &candle
		return InvalidValue, nil
	}

	// Incremental update using smoothed ATR formula:
	// ATR[t] = (ATR[t-1] * (period - 1) + TR[t]) / period
	a.currentATR = (a.currentATR*float64(a.period-1) + tr) / float64(a.period)
	a.barsSeen++
	a.prevCandle = &candle

	return NewValue(a.currentATR), nil
}

// IsWarm returns true when the indicator has enough data.
func (a *CachedATR) IsWarm() bool {
	return a.isWarm
}

// Reset clears the indicator state for a new backtest run.
func (a *CachedATR) Reset() {
	a.currentATR = 0
	a.prevCandle = nil
	a.barsSeen = 0
	a.isWarm = false
	a.trBuffer = make([]float64, 0, a.period)
}

// WarmupPeriod returns the number of bars needed before the indicator is warm.
func (a *CachedATR) WarmupPeriod() int {
	return a.period + 1 // Need period + 1 bars (1 for previous candle + period TRs)
}

// calculateTrueRange computes the true range for a candle.
func calculateTrueRange(current, previous *market.Candle) float64 {
	// True Range = max of:
	// 1. High - Low
	// 2. |High - Previous Close|
	// 3. |Low - Previous Close|
	highLow := current.High - current.Low
	highPrevClose := math.Abs(current.High - previous.Close)
	lowPrevClose := math.Abs(current.Low - previous.Close)

	return math.Max(highLow, math.Max(highPrevClose, lowPrevClose))
}
