package transform

import (
	"fmt"
	"math"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// ZScoreTransform normalizes data to z-scores (standard deviations from mean)
// Z-score = (x - μ) / σ
// where μ is the mean and σ is the standard deviation
//
// This is useful for:
// - Statistical analysis
// - Comparing different timeframes
// - Identifying outliers (|z| > 2 or 3)
// - Mean reversion strategies
type ZScoreTransform struct {
	_ TransformConfig
	Field  string `yaml:"field"`  // Which field to normalize (close, volume, etc.)
	Window int    `yaml:"window"` // Rolling window for mean/stddev calculation
}

// Type returns the transform type identifier
func (t *ZScoreTransform) Type() string {
	return "zscore"
}

// Name returns the name of the transform for debugging/logging
func (t *ZScoreTransform) Name() string {
	return fmt.Sprintf("zscore(field=%s, window=%d)", t.Field, t.Window)
}

// Validate checks if the transform configuration is valid
func (t *ZScoreTransform) Validate() error {
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

// Apply transforms the data by calculating z-scores
func (t *ZScoreTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
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

	// For each candle, calculate z-score using rolling window
	for i := range result {
		if i < t.Window-1 {
			// Not enough data yet, set to 0
			t.setFieldValue(result[i], 0)
			continue
		}

		// Calculate mean and stddev over window
		windowStart := i - t.Window + 1
		values := make([]float64, t.Window)
		for j := 0; j < t.Window; j++ {
			values[j] = t.getFieldValue(candles[windowStart+j])
		}

		mean := calculateMean(values)
		stddev := calculateStdDev(values, mean)

		// Calculate z-score
		currentValue := t.getFieldValue(candles[i])
		var zscore float64
		if stddev > 0 {
			zscore = (currentValue - mean) / stddev
		} else {
			// If stddev is 0 (all values same), z-score is 0
			zscore = 0
		}

		t.setFieldValue(result[i], zscore)
	}

	return result, nil
}

// getFieldValue extracts the specified field value from a candle
func (t *ZScoreTransform) getFieldValue(c *market.Candle) float64 {
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
func (t *ZScoreTransform) setFieldValue(c *market.Candle, value float64) {
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

// calculateMean computes the arithmetic mean of values
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStdDev computes the standard deviation
func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}
