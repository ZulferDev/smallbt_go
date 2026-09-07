package transform

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// DifferencingTransform computes the difference between consecutive values
// Formula: diff[t] = value[t] - value[t-1]
//
// This is useful for:
// - Making time series stationary
// - Removing trends
// - Velocity/rate of change analysis
// - Momentum strategies
//
// Can be applied multiple times for higher-order differencing:
// - First difference: price change
// - Second difference: acceleration
type DifferencingTransform struct {
	_ TransformConfig
	Field  string `yaml:"field"` // Which field to difference (close, volume, etc.)
	Order  int    `yaml:"order"` // Order of differencing (1 = first difference, 2 = second, etc.)
}

// Type returns the transform type identifier
func (t *DifferencingTransform) Type() string {
	return "difference"
}

// Name returns the name of the transform for debugging/logging
func (t *DifferencingTransform) Name() string {
	return fmt.Sprintf("difference(field=%s, order=%d)", t.Field, t.Order)
}

// Validate checks if the transform configuration is valid
func (t *DifferencingTransform) Validate() error {
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

	if t.Order < 1 {
		return fmt.Errorf("order must be at least 1, got %d", t.Order)
	}

	if t.Order > 3 {
		return fmt.Errorf("order must be at most 3, got %d (higher orders rarely useful)", t.Order)
	}

	return nil
}

// Apply transforms the data by computing differences
func (t *DifferencingTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	if len(candles) < t.Order+1 {
		return nil, fmt.Errorf("insufficient data: need at least %d candles for order %d differencing, got %d",
			t.Order+1, t.Order, len(candles))
	}

	result := CopyCandles(candles)

	// Apply differencing order times
	for order := 1; order <= t.Order; order++ {
		result = t.applyOneDifference(result)
	}

	return result, nil
}

// applyOneDifference applies first-order differencing
func (t *DifferencingTransform) applyOneDifference(candles []*market.Candle) []*market.Candle {
	if len(candles) < 2 {
		return candles
	}

	result := CopyCandles(candles)

	// First value has no previous value, set to 0
	t.setFieldValue(result[0], 0)

	// For each subsequent candle, compute difference
	for i := 1; i < len(result); i++ {
		current := t.getFieldValue(candles[i])
		previous := t.getFieldValue(candles[i-1])
		diff := current - previous
		t.setFieldValue(result[i], diff)
	}

	return result
}

// getFieldValue extracts the specified field value from a candle
func (t *DifferencingTransform) getFieldValue(c *market.Candle) float64 {
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
func (t *DifferencingTransform) setFieldValue(c *market.Candle, value float64) {
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
