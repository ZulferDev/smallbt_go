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

	// For the first value, use SMA
	if len(ctx.Candles) == e.period {
		sum := 0.0
		for i := 0; i < e.period; i++ {
			price := SourcePrice(ctx.Candles[i], e.source)
			sum += price
		}
		sma := sum / float64(e.period)
		return NewValue(sma), nil
	}

	// For subsequent values, use EMA formula:
	// EMA(t) = Price(t) * multiplier + EMA(t-1) * (1 - multiplier)

	// We need the previous EMA value
	// To get it, we temporarily calculate with one less candle
	prevCtx := &Context{
		Current:         ctx.Candles[len(ctx.Candles)-2],
		Candles:         ctx.Candles[:len(ctx.Candles)-1],
		Symbol:          ctx.Symbol,
		Timeframe:       ctx.Timeframe,
		IndicatorValues: ctx.IndicatorValues,
		BarIndex:        ctx.BarIndex - 1,
	}

	prevEMA, err := e.Calculate(prevCtx)
	if err != nil {
		return InvalidValue, err
	}

	if !prevEMA.Valid {
		return InvalidValue, nil
	}

	currentPrice := SourcePrice(ctx.Current, e.source)
	ema := currentPrice*e.multiplier + prevEMA.Value*(1.0-e.multiplier)

	return NewValue(ema), nil
}
