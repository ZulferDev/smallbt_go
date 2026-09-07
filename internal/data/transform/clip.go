package transform

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// ClipTransform clips (caps) values to a specified range.
// Values below min are set to min, values above max are set to max.
//
// This is useful for:
// - Outlier handling
// - Range limiting
// - Risk control (cap extreme values)
// - Data sanitization
// - Preventing extreme signals
type ClipTransform struct {
	_ TransformConfig
	Field  string  `yaml:"field"` // Which field to clip (close, volume, etc.)
	Min    float64 `yaml:"min"`   // Minimum value (lower bound)
	Max    float64 `yaml:"max"`   // Maximum value (upper bound)
}

// Type returns the transform type identifier
func (t *ClipTransform) Type() string {
	return "clip"
}

// Name returns the name of the transform for debugging/logging
func (t *ClipTransform) Name() string {
	return fmt.Sprintf("clip(field=%s, min=%.2f, max=%.2f)", t.Field, t.Min, t.Max)
}

// Validate checks if the transform configuration is valid
func (t *ClipTransform) Validate() error {
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

	if t.Min >= t.Max {
		return fmt.Errorf("min (%f) must be less than max (%f)", t.Min, t.Max)
	}

	return nil
}

// Apply transforms the data by clipping values to [min, max] range
func (t *ClipTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("insufficient data: need at least 1 candle")
	}

	result := CopyCandles(candles)

	// Clip each value
	for i := range result {
		value := t.getFieldValue(candles[i])
		clipped := t.clip(value)
		t.setFieldValue(result[i], clipped)
	}

	return result, nil
}

// clip clips a value to [min, max] range
func (t *ClipTransform) clip(value float64) float64 {
	if value < t.Min {
		return t.Min
	}
	if value > t.Max {
		return t.Max
	}
	return value
}

// getFieldValue extracts the specified field value from a candle
func (t *ClipTransform) getFieldValue(c *market.Candle) float64 {
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
func (t *ClipTransform) setFieldValue(c *market.Candle, value float64) {
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
