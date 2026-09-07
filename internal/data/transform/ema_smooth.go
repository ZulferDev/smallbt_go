package transform

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// EMASmooth smooths price data using exponential moving average.
// Formula: EMA[t] = α * value[t] + (1 - α) * EMA[t-1]
// where α = 2 / (period + 1)
//
// EMA gives more weight to recent values compared to SMA.
// This is useful for:
// - Noise reduction with trend responsiveness
// - Lag reduction compared to SMA
// - Signal smoothing while preserving recent moves
// - Trend detection with faster reaction
type EMASmooth struct {
	_ TransformConfig
	Field  string `yaml:"field"`  // Which field to smooth (close, volume, etc.)
	Period int    `yaml:"period"` // EMA period
}

// Type returns the transform type identifier
func (t *EMASmooth) Type() string {
	return "ema_smooth"
}

// Name returns the name of the transform for debugging/logging
func (t *EMASmooth) Name() string {
	return fmt.Sprintf("ema_smooth(field=%s, period=%d)", t.Field, t.Period)
}

// Validate checks if the transform configuration is valid
func (t *EMASmooth) Validate() error {
	if t.Field == "" {
		return fmt.Errorf("field is required")
	}

	validFields := []string{"open", "high", "low", "close", "volume"}
	valid := false
	for _, f := range validFields {
		if t.Field == f {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid field %s, must be one of: %v", t.Field, validFields)
	}

	if t.Period < 2 {
		return fmt.Errorf("period must be at least 2, got %d", t.Period)
	}

	return nil
}

// Apply transforms the data by applying EMA smoothing
func (t *EMASmooth) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	if len(candles) < 1 {
		return nil, fmt.Errorf("insufficient data: need at least 1 candle")
	}

	result := CopyCandles(candles)

	// Calculate smoothing factor (alpha)
	alpha := 2.0 / float64(t.Period+1)

	// Initialize EMA with first value
	ema := t.getFieldValue(candles[0])
	t.setFieldValue(result[0], ema)

	// Calculate EMA for remaining values
	for i := 1; i < len(result); i++ {
		currentValue := t.getFieldValue(candles[i])
		ema = alpha*currentValue + (1-alpha)*ema
		t.setFieldValue(result[i], ema)
	}

	return result, nil
}

// getFieldValue extracts the specified field value from a candle
func (t *EMASmooth) getFieldValue(c *market.Candle) float64 {
	switch t.Field {
	case "open":
		return c.Open
	case "high":
		return c.High
	case "low":
		return c.Low
	case "close":
		return c.Close
	case "volume":
		return c.Volume
	default:
		return 0
	}
}

// setFieldValue sets the specified field value in a candle
func (t *EMASmooth) setFieldValue(c *market.Candle, value float64) {
	switch t.Field {
	case "open":
		c.Open = value
	case "high":
		c.High = value
	case "low":
		c.Low = value
	case "close":
		c.Close = value
	case "volume":
		c.Volume = value
	}
}
