package transform

import (
	"fmt"
	"math"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// NormalizeTransform normalizes OHLC values using min-max scaling.
// Formula: (value - min) / (max - min)
// Output range: [0, 1]
type NormalizeTransform struct {
	config TransformConfig
	field  string // "close", "open", "high", "low", "volume"
}

// NewNormalizeTransform creates a new normalize transform.
// Field specifies which field to normalize: "close", "open", "high", "low", "volume".
// Empty field normalizes all OHLC fields.
func NewNormalizeTransform(field string) *NormalizeTransform {
	return &NormalizeTransform{
		config: DefaultTransformConfig(),
		field:  field,
	}
}

// Apply applies min-max normalization.
func (t *NormalizeTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	// Deep copy to avoid modifying original
	result := CopyCandles(candles)

	if t.field == "" || t.field == "all" {
		// Normalize all OHLC fields
		t.normalizeField(result, "open")
		t.normalizeField(result, "high")
		t.normalizeField(result, "low")
		t.normalizeField(result, "close")
		t.normalizeField(result, "volume")
	} else {
		t.normalizeField(result, t.field)
	}

	return result, nil
}

// normalizeField normalizes a specific field.
func (t *NormalizeTransform) normalizeField(candles []*market.Candle, field string) {
	if len(candles) == 0 {
		return
	}

	// Find min and max
	min, max := math.MaxFloat64, -math.MaxFloat64

	for _, candle := range candles {
		value := getFieldValue(candle, field)
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}

	// Avoid division by zero
	if max == min {
		// All values are the same, set to 0.5
		for _, candle := range candles {
			setFieldValue(candle, field, 0.5)
		}
		return
	}

	// Normalize: (value - min) / (max - min)
	scale := max - min
	for _, candle := range candles {
		value := getFieldValue(candle, field)
		normalized := (value - min) / scale
		setFieldValue(candle, field, normalized)
	}
}

// Name returns the transform name.
func (t *NormalizeTransform) Name() string {
	if t.field == "" || t.field == "all" {
		return "Normalize(all)"
	}
	return fmt.Sprintf("Normalize(%s)", t.field)
}

// Validate validates the transform configuration.
func (t *NormalizeTransform) Validate() error {
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

// LogReturnsTransform calculates logarithmic returns.
// Formula: ln(price[i] / price[i-1])
// Output count: len(input) - 1
type LogReturnsTransform struct {
	config TransformConfig
	field  string // "close", "open", "high", "low"
}

// NewLogReturnsTransform creates a new log returns transform.
func NewLogReturnsTransform(field string) *LogReturnsTransform {
	if field == "" {
		field = "close"
	}
	return &LogReturnsTransform{
		config: DefaultTransformConfig(),
		field:  field,
	}
}

// Apply calculates log returns.
func (t *LogReturnsTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	if len(candles) < 2 {
		return nil, fmt.Errorf("%w: need at least 2 candles for log returns", ErrInsufficientData)
	}

	// Result has len(input) - 1
	result := make([]*market.Candle, len(candles)-1)

	for i := 1; i < len(candles); i++ {
		prev := getFieldValue(candles[i-1], t.field)
		curr := getFieldValue(candles[i], t.field)

		if prev <= 0 || curr <= 0 {
			return nil, fmt.Errorf("%w: price must be positive for log returns", ErrInvalidCandle)
		}

		logReturn := math.Log(curr / prev)

		// Create new candle with log return as close
		result[i-1] = &market.Candle{
			Timestamp: candles[i].Timestamp,
			Open:      logReturn,
			High:      logReturn,
			Low:       logReturn,
			Close:     logReturn,
			Volume:    candles[i].Volume,
		}
	}

	return result, nil
}

// Name returns the transform name.
func (t *LogReturnsTransform) Name() string {
	return fmt.Sprintf("LogReturns(%s)", t.field)
}

// Validate validates the transform configuration.
func (t *LogReturnsTransform) Validate() error {
	validFields := map[string]bool{
		"open":  true,
		"high":  true,
		"low":   true,
		"close": true,
	}

	if !validFields[t.field] {
		return fmt.Errorf("%w: invalid field %s", ErrInvalidConfig, t.field)
	}

	return nil
}

// PercentageChangeTransform calculates percentage change.
// Formula: (price[i] - price[i-1]) / price[i-1] * 100
// Output count: len(input) - 1
type PercentageChangeTransform struct {
	config TransformConfig
	field  string // "close", "open", "high", "low"
}

// NewPercentageChangeTransform creates a new percentage change transform.
func NewPercentageChangeTransform(field string) *PercentageChangeTransform {
	if field == "" {
		field = "close"
	}
	return &PercentageChangeTransform{
		config: DefaultTransformConfig(),
		field:  field,
	}
}

// Apply calculates percentage change.
func (t *PercentageChangeTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	if len(candles) < 2 {
		return nil, fmt.Errorf("%w: need at least 2 candles for percentage change", ErrInsufficientData)
	}

	// Result has len(input) - 1
	result := make([]*market.Candle, len(candles)-1)

	for i := 1; i < len(candles); i++ {
		prev := getFieldValue(candles[i-1], t.field)
		curr := getFieldValue(candles[i], t.field)

		if prev == 0 {
			return nil, fmt.Errorf("%w: previous price cannot be zero", ErrInvalidCandle)
		}

		pctChange := ((curr - prev) / prev) * 100

		// Create new candle with percentage change
		result[i-1] = &market.Candle{
			Timestamp: candles[i].Timestamp,
			Open:      pctChange,
			High:      pctChange,
			Low:       pctChange,
			Close:     pctChange,
			Volume:    candles[i].Volume,
		}
	}

	return result, nil
}

// Name returns the transform name.
func (t *PercentageChangeTransform) Name() string {
	return fmt.Sprintf("PercentageChange(%s)", t.field)
}

// Validate validates the transform configuration.
func (t *PercentageChangeTransform) Validate() error {
	validFields := map[string]bool{
		"open":  true,
		"high":  true,
		"low":   true,
		"close": true,
	}

	if !validFields[t.field] {
		return fmt.Errorf("%w: invalid field %s", ErrInvalidConfig, t.field)
	}

	return nil
}

// Helper functions

func getFieldValue(candle *market.Candle, field string) float64 {
	switch field {
	case "open":
		return candle.Open
	case "high":
		return candle.High
	case "low":
		return candle.Low
	case "close":
		return candle.Close
	case "volume":
		return candle.Volume
	default:
		return candle.Close
	}
}

func setFieldValue(candle *market.Candle, field string, value float64) {
	switch field {
	case "open":
		candle.Open = value
	case "high":
		candle.High = value
	case "low":
		candle.Low = value
	case "close":
		candle.Close = value
	case "volume":
		candle.Volume = value
	}
}
