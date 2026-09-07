package transform

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// PercentileRankTransform converts values to their percentile rank (0-100)
// within a rolling window.
//
// Formula: percentile = (rank / (n-1)) * 100
// where rank = number of values <= current value
//
// This is useful for:
// - Relative strength analysis
// - Position within range (overbought/oversold)
// - Non-parametric normalization
// - Rank-based strategies
// - Distribution-agnostic comparison
type PercentileRankTransform struct {
	_      TransformConfig
	Field  string `yaml:"field"`  // Which field to transform (close, volume, etc.)
	Window int    `yaml:"window"` // Rolling window for percentile calculation
}

// Type returns the transform type identifier
func (t *PercentileRankTransform) Type() string {
	return "percentile_rank"
}

// Name returns the name of the transform for debugging/logging
func (t *PercentileRankTransform) Name() string {
	return fmt.Sprintf("percentile_rank(field=%s, window=%d)", t.Field, t.Window)
}

// Validate checks if the transform configuration is valid
func (t *PercentileRankTransform) Validate() error {
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

	if t.Window < 2 {
		return fmt.Errorf("window must be at least 2, got %d", t.Window)
	}

	return nil
}

// Apply transforms the data by calculating percentile ranks
func (t *PercentileRankTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	if len(candles) < t.Window {
		return nil, fmt.Errorf("insufficient data: need at least %d candles, got %d", t.Window, len(candles))
	}

	result := CopyCandles(candles)

	// Calculate percentile rank for each candle
	for i := range result {
		if i < t.Window-1 {
			// Not enough data yet, set to 50 (neutral)
			t.setFieldValue(result[i], 50.0)
			continue
		}

		// Get window values
		windowStart := i - t.Window + 1
		values := make([]float64, t.Window)
		for j := 0; j < t.Window; j++ {
			values[j] = t.getFieldValue(candles[windowStart+j])
		}

		// Calculate percentile rank
		currentValue := t.getFieldValue(candles[i])
		percentile := t.calculatePercentileRank(values, currentValue)
		t.setFieldValue(result[i], percentile)
	}

	return result, nil
}

// calculatePercentileRank computes the percentile rank of value in values
func (t *PercentileRankTransform) calculatePercentileRank(values []float64, value float64) float64 {
	if len(values) == 0 {
		return 50.0 // Neutral if no data
	}

	// Count how many values are less than or equal to current value
	rank := 0
	for _, v := range values {
		if v <= value {
			rank++
		}
	}

	// Convert rank to percentile (0-100)
	// Use (rank-1) to get 0-based rank, then divide by (n-1)
	// This gives 0% for minimum, 100% for maximum
	n := len(values)
	if n == 1 {
		return 50.0 // Single value is at 50th percentile
	}

	percentile := float64(rank-1) / float64(n-1) * 100.0

	// Clamp to [0, 100]
	if percentile < 0 {
		percentile = 0
	}
	if percentile > 100 {
		percentile = 100
	}

	return percentile
}

// getFieldValue extracts the specified field value from a candle
func (t *PercentileRankTransform) getFieldValue(c *market.Candle) float64 {
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
func (t *PercentileRankTransform) setFieldValue(c *market.Candle, value float64) {
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
