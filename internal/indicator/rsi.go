package indicator

import (
	"fmt"
	"math"
)

// RSI implements Relative Strength Index indicator.
type RSI struct {
	name   string
	period int
	source string
}

// RSIFactory creates a new RSI indicator from config.
func RSIFactory(config Config) (Indicator, error) {
	if config.Period <= 0 {
		return nil, fmt.Errorf("rsi: period must be positive, got %d", config.Period)
	}

	source := config.Source
	if source == "" {
		source = "close"
	}

	return &RSI{
		name:   "rsi",
		period: config.Period,
		source: source,
	}, nil
}

// Name returns the indicator name.
func (r *RSI) Name() string {
	return r.name
}

// Calculate computes the RSI value.
func (r *RSI) Calculate(ctx *Context) (Value, error) {
	// Need at least 'period + 1' candles to calculate price changes
	if len(ctx.Candles) < r.period+1 {
		return InvalidValue, nil
	}

	// Calculate price changes
	changes := make([]float64, 0, len(ctx.Candles)-1)
	for i := 1; i < len(ctx.Candles); i++ {
		current := SourcePrice(ctx.Candles[i], r.source)
		previous := SourcePrice(ctx.Candles[i-1], r.source)
		change := current - previous
		changes = append(changes, change)
	}

	// Calculate average gain and average loss
	avgGain := 0.0
	avgLoss := 0.0

	// First average: simple average of first 'period' changes
	for i := 0; i < r.period; i++ {
		if changes[i] > 0 {
			avgGain += changes[i]
		} else {
			avgLoss += math.Abs(changes[i])
		}
	}

	avgGain /= float64(r.period)
	avgLoss /= float64(r.period)

	// If we have more than 'period' changes, use smoothed method
	for i := r.period; i < len(changes); i++ {
		gain := 0.0
		loss := 0.0
		if changes[i] > 0 {
			gain = changes[i]
		} else {
			loss = math.Abs(changes[i])
		}

		// Smoothed average: (prev_avg * (period - 1) + current) / period
		avgGain = (avgGain*float64(r.period-1) + gain) / float64(r.period)
		avgLoss = (avgLoss*float64(r.period-1) + loss) / float64(r.period)
	}

	// Calculate RSI
	if avgLoss == 0 {
		// No losses means RSI is 100
		return NewValue(100.0), nil
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))

	return NewValue(rsi), nil
}
