package indicator

import (
	"fmt"
)

// EMA implements Exponential Moving Average indicator.
type EMA struct {
	name       string
	period     int
	source     string
	multiplier float64
	prevEMA    *float64 // Cache previous EMA value
	lastBarIdx int      // Last bar index calculated (for cache invalidation)
}

// EMAFactory creates a new EMA indicator from config.
func EMAFactory(config Config) (Indicator, error) {
	if config.Period <= 0 {
		return nil, fmt.Errorf("ema: period must be positive, got %d", config.Period)
	}

	source := config.Source
	if source == "" {
		source = "close"
	}

	// EMA multiplier = 2 / (period + 1)
	multiplier := 2.0 / (float64(config.Period) + 1.0)

	return &EMA{
		name:       "ema",
		period:     config.Period,
		source:     source,
		multiplier: multiplier,
	}, nil
}

// Name returns the indicator name.
func (e *EMA) Name() string {
	return e.name
}

// Calculate computes the EMA value.
func (e *EMA) Calculate(ctx *Context) (Value, error) {
	// Need at least 'period' candles for EMA
	if len(ctx.Candles) < e.period {
		return InvalidValue, nil
	}

	// Check if we've already calculated EMA for this bar index
	if e.lastBarIdx == ctx.BarIndex && e.prevEMA != nil {
		return NewValue(*e.prevEMA), nil
	}

	// For the first value (when len == period), use SMA
	if len(ctx.Candles) == e.period {
		sum := 0.0
		for i := 0; i < e.period; i++ {
			price := SourcePrice(ctx.Candles[i], e.source)
			sum += price
		}
		sma := sum / float64(e.period)
		e.prevEMA = &sma
		e.lastBarIdx = ctx.BarIndex
		return NewValue(sma), nil
	}

	// For subsequent values, use EMA formula:
	// EMA(t) = Price(t) * multiplier + EMA(t-1) * (1 - multiplier)

	// If prevEMA is nil but we have enough data, we need to warm up the cache
	// This happens when we skip to a later bar without calculating from the start
	if e.prevEMA == nil {
		// Calculate SMA for the first 'period' candles to initialize
		sum := 0.0
		for i := 0; i < e.period; i++ {
			price := SourcePrice(ctx.Candles[i], e.source)
			sum += price
		}
		sma := sum / float64(e.period)
		e.prevEMA = &sma

		// Now calculate EMA for remaining candles up to current
		for i := e.period; i < len(ctx.Candles); i++ {
			price := SourcePrice(ctx.Candles[i], e.source)
			ema := price*e.multiplier + *e.prevEMA*(1.0-e.multiplier)
			e.prevEMA = &ema
		}
		e.lastBarIdx = ctx.BarIndex
		return NewValue(*e.prevEMA), nil
	}

	currentPrice := SourcePrice(ctx.Current, e.source)
	ema := currentPrice*e.multiplier + *e.prevEMA*(1.0-e.multiplier)

	// Cache for next calculation
	e.prevEMA = &ema
	e.lastBarIdx = ctx.BarIndex

	return NewValue(ema), nil
}
