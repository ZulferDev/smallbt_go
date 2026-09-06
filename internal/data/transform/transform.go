package transform

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Transform represents a data transformation function.
// Transforms are composable and can be chained together.
type Transform interface {
	// Apply applies the transformation to the input candles.
	// Returns transformed candles or an error.
	Apply(candles []*market.Candle) ([]*market.Candle, error)

	// Name returns the name of the transform for debugging/logging.
	Name() string

	// Validate validates the transform configuration.
	Validate() error
}

// TransformFunc is a function type that implements Transform.
// Useful for simple stateless transforms.
type TransformFunc func([]*market.Candle) ([]*market.Candle, error)

// Apply implements Transform interface.
func (f TransformFunc) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	return f(candles)
}

// Name returns "TransformFunc".
func (f TransformFunc) Name() string {
	return "TransformFunc"
}

// Validate always returns nil for TransformFunc.
func (f TransformFunc) Validate() error {
	return nil
}

// TransformChain represents a chain of transforms applied sequentially.
type TransformChain struct {
	transforms []Transform
	name       string
}

// NewTransformChain creates a new transform chain.
func NewTransformChain(transforms ...Transform) *TransformChain {
	return &TransformChain{
		transforms: transforms,
		name:       "TransformChain",
	}
}

// Apply applies all transforms in sequence.
func (tc *TransformChain) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if len(tc.transforms) == 0 {
		return candles, nil
	}

	result := candles
	var err error

	for i, transform := range tc.transforms {
		result, err = transform.Apply(result)
		if err != nil {
			return nil, fmt.Errorf("transform %d (%s) failed: %w", i, transform.Name(), err)
		}
	}

	return result, nil
}

// Name returns the chain name.
func (tc *TransformChain) Name() string {
	return tc.name
}

// Validate validates all transforms in the chain.
func (tc *TransformChain) Validate() error {
	for i, transform := range tc.transforms {
		if err := transform.Validate(); err != nil {
			return fmt.Errorf("transform %d (%s) validation failed: %w", i, transform.Name(), err)
		}
	}
	return nil
}

// Add adds a transform to the chain.
func (tc *TransformChain) Add(transform Transform) {
	tc.transforms = append(tc.transforms, transform)
}

// Len returns the number of transforms in the chain.
func (tc *TransformChain) Len() int {
	return len(tc.transforms)
}

// Get returns the transform at index i.
func (tc *TransformChain) Get(i int) Transform {
	if i < 0 || i >= len(tc.transforms) {
		return nil
	}
	return tc.transforms[i]
}

// Clear removes all transforms from the chain.
func (tc *TransformChain) Clear() {
	tc.transforms = nil
}

// Clone creates a copy of the transform chain.
func (tc *TransformChain) Clone() *TransformChain {
	clone := &TransformChain{
		transforms: make([]Transform, len(tc.transforms)),
		name:       tc.name,
	}
	copy(clone.transforms, tc.transforms)
	return clone
}

// TransformConfig contains common configuration for transforms.
type TransformConfig struct {
	// SkipInvalid skips invalid candles instead of returning error.
	SkipInvalid bool

	// PreserveCount ensures output has same count as input.
	// For transforms that reduce count (e.g., log returns), pads with nil.
	PreserveCount bool
}

// DefaultTransformConfig returns default configuration.
func DefaultTransformConfig() TransformConfig {
	return TransformConfig{
		SkipInvalid:   false,
		PreserveCount: false,
	}
}

// Errors
var (
	// ErrEmptyInput is returned when input is empty.
	ErrEmptyInput = fmt.Errorf("empty input")

	// ErrInvalidCandle is returned when a candle is invalid.
	ErrInvalidCandle = fmt.Errorf("invalid candle")

	// ErrInsufficientData is returned when there's not enough data for transform.
	ErrInsufficientData = fmt.Errorf("insufficient data")

	// ErrInvalidConfig is returned when configuration is invalid.
	ErrInvalidConfig = fmt.Errorf("invalid configuration")
)

// Helper functions

// ValidateCandles validates that candles are not nil and have valid values.
func ValidateCandles(candles []*market.Candle) error {
	if len(candles) == 0 {
		return ErrEmptyInput
	}

	for i, candle := range candles {
		if candle == nil {
			return fmt.Errorf("candle at index %d is nil: %w", i, ErrInvalidCandle)
		}
	}

	return nil
}

// CopyCandles creates a deep copy of candles.
func CopyCandles(candles []*market.Candle) []*market.Candle {
	if candles == nil {
		return nil
	}

	result := make([]*market.Candle, len(candles))
	for i, candle := range candles {
		if candle != nil {
			copied := *candle
			result[i] = &copied
		}
	}

	return result
}
