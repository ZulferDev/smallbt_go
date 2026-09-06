package transform

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// MovingAverageSmoothTransform smooths price data using simple moving average.
// Formula: average of last N values
// Output count: len(input) for SMA (first N-1 values are padded)
type MovingAverageSmoothTransform struct {
	config TransformConfig
	period int
	field  string // "close", "open", "high", "low"
}

// NewMovingAverageSmoothTransform creates a new moving average smooth transform.
func NewMovingAverageSmoothTransform(period int, field string) *MovingAverageSmoothTransform {
	if field == "" {
		field = "close"
	}
	return &MovingAverageSmoothTransform{
		config: DefaultTransformConfig(),
		period: period,
		field:  field,
	}
}

// Apply applies moving average smoothing.
func (t *MovingAverageSmoothTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	if len(candles) < t.period {
		return nil, fmt.Errorf("%w: need at least %d candles for MA(%d)", ErrInsufficientData, t.period, t.period)
	}

	// Deep copy
	result := CopyCandles(candles)

	// Calculate moving average for each position
	for i := t.period - 1; i < len(result); i++ {
		sum := 0.0
		for j := 0; j < t.period; j++ {
			sum += getFieldValue(candles[i-j], t.field)
		}
		avg := sum / float64(t.period)

		// Set the smoothed value
		setFieldValue(result[i], t.field, avg)
	}

	// For first N-1 values, keep original or use partial average
	for i := 0; i < t.period-1; i++ {
		// Use expanding average for first values
		sum := 0.0
		count := i + 1
		for j := 0; j <= i; j++ {
			sum += getFieldValue(candles[j], t.field)
		}
		avg := sum / float64(count)
		setFieldValue(result[i], t.field, avg)
	}

	return result, nil
}

// Name returns the transform name.
func (t *MovingAverageSmoothTransform) Name() string {
	return fmt.Sprintf("MASmooth(%d,%s)", t.period, t.field)
}

// Validate validates the transform configuration.
func (t *MovingAverageSmoothTransform) Validate() error {
	if t.period < 2 {
		return fmt.Errorf("%w: period must be >= 2", ErrInvalidConfig)
	}

	validFields := map[string]bool{
		"open":   true,
		"high":   true,
		"low":    true,
		"close":  true,
		"volume": true,
	}

	if !validFields[t.field] {
		return fmt.Errorf("%w: invalid field %s", ErrInvalidConfig, t.field)
	}

	return nil
}

// DifferenceTransform calculates first-order difference.
// Formula: value[i] - value[i-1]
// Output count: len(input) - 1
type DifferenceTransform struct {
	config TransformConfig
	field  string // "close", "open", "high", "low"
}

// NewDifferenceTransform creates a new difference transform.
func NewDifferenceTransform(field string) *DifferenceTransform {
	if field == "" {
		field = "close"
	}
	return &DifferenceTransform{
		config: DefaultTransformConfig(),
		field:  field,
	}
}

// Apply calculates first difference.
func (t *DifferenceTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	if len(candles) < 2 {
		return nil, fmt.Errorf("%w: need at least 2 candles for difference", ErrInsufficientData)
	}

	// Result has len(input) - 1
	result := make([]*market.Candle, len(candles)-1)

	for i := 1; i < len(candles); i++ {
		prev := getFieldValue(candles[i-1], t.field)
		curr := getFieldValue(candles[i], t.field)

		diff := curr - prev

		// Create new candle with difference
		result[i-1] = &market.Candle{
			Timestamp: candles[i].Timestamp,
			Open:      diff,
			High:      diff,
			Low:       diff,
			Close:     diff,
			Volume:    candles[i].Volume,
		}
	}

	return result, nil
}

// Name returns the transform name.
func (t *DifferenceTransform) Name() string {
	return fmt.Sprintf("Difference(%s)", t.field)
}

// Validate validates the transform configuration.
func (t *DifferenceTransform) Validate() error {
	validFields := map[string]bool{
		"open":   true,
		"high":   true,
		"low":    true,
		"close":  true,
		"volume": true,
	}

	if !validFields[t.field] {
		return fmt.Errorf("%w: invalid field %s", ErrInvalidConfig, t.field)
	}

	return nil
}

// ScaleTransform scales values by a constant factor.
// Formula: value * factor
type ScaleTransform struct {
	config TransformConfig
	factor float64
	field  string // "close", "open", "high", "low", "volume", "all"
}

// NewScaleTransform creates a new scale transform.
func NewScaleTransform(factor float64, field string) *ScaleTransform {
	return &ScaleTransform{
		config: DefaultTransformConfig(),
		factor: factor,
		field:  field,
	}
}

// Apply applies scaling.
func (t *ScaleTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	// Deep copy
	result := CopyCandles(candles)

	if t.field == "" || t.field == "all" {
		// Scale all OHLCV fields
		for _, candle := range result {
			candle.Open *= t.factor
			candle.High *= t.factor
			candle.Low *= t.factor
			candle.Close *= t.factor
			candle.Volume *= t.factor
		}
	} else {
		// Scale specific field
		for _, candle := range result {
			value := getFieldValue(candle, t.field)
			setFieldValue(candle, t.field, value*t.factor)
		}
	}

	return result, nil
}

// Name returns the transform name.
func (t *ScaleTransform) Name() string {
	if t.field == "" || t.field == "all" {
		return fmt.Sprintf("Scale(%.2f,all)", t.factor)
	}
	return fmt.Sprintf("Scale(%.2f,%s)", t.factor, t.field)
}

// Validate validates the transform configuration.
func (t *ScaleTransform) Validate() error {
	validFields := map[string]bool{
		"":       true,
		"all":    true,
		"open":   true,
		"high":   true,
		"low":    true,
		"close":  true,
		"volume": true,
	}

	if !validFields[t.field] {
		return fmt.Errorf("%w: invalid field %s", ErrInvalidConfig, t.field)
	}

	return nil
}
