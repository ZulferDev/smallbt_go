package indicator

import (
	"fmt"
)

// SMA implements Simple Moving Average indicator.
type SMA struct {
	name   string
	period int
	source string
}

// SMAFactory creates a new SMA indicator from config.
func SMAFactory(config Config) (Indicator, error) {
	if config.Period <= 0 {
		return nil, fmt.Errorf("sma: period must be positive, got %d", config.Period)
	}

	source := config.Source
	if source == "" {
		source = "close"
	}

	return &SMA{
		name:   "sma",
		period: config.Period,
		source: source,
	}, nil
}

// Name returns the indicator name.
func (s *SMA) Name() string {
	return s.name
}

// Calculate computes the SMA value.
func (s *SMA) Calculate(ctx *Context) (Value, error) {
	// Need at least 'period' candles for SMA
	if len(ctx.Candles) < s.period {
		return InvalidValue, nil
	}

	// Calculate sum of the last 'period' values
	sum := 0.0
	startIdx := len(ctx.Candles) - s.period

	for i := startIdx; i < len(ctx.Candles); i++ {
		price := SourcePrice(ctx.Candles[i], s.source)
		sum += price
	}

	// Calculate average
	avg := sum / float64(s.period)

	return NewValue(avg), nil
}
