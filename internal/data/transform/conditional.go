package transform

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Condition is a function that evaluates if a transform should be applied.
type Condition func(*market.Candle) bool

// ConditionalTransform applies a transform only when condition is met.
type ConditionalTransform struct {
	transform Transform
	condition Condition
	config    TransformConfig
}

// NewConditionalTransform creates a conditional transform.
func NewConditionalTransform(transform Transform, condition Condition) *ConditionalTransform {
	return &ConditionalTransform{
		transform: transform,
		condition: condition,
		config:    DefaultTransformConfig(),
	}
}

// Apply applies the transform only to candles that meet the condition.
func (ct *ConditionalTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	// Filter candles that meet condition
	filtered := make([]*market.Candle, 0, len(candles))
	for _, candle := range candles {
		if ct.condition(candle) {
			filtered = append(filtered, candle)
		}
	}

	if len(filtered) == 0 {
		// No candles meet condition, return original
		return candles, nil
	}

	// Apply transform to filtered candles
	transformed, err := ct.transform.Apply(filtered)
	if err != nil {
		return nil, fmt.Errorf("conditional transform failed: %w", err)
	}

	// Merge back with original (replace filtered candles)
	result := CopyCandles(candles)
	transformIdx := 0
	for i, candle := range candles {
		if ct.condition(candle) && transformIdx < len(transformed) {
			result[i] = transformed[transformIdx]
			transformIdx++
		}
	}

	return result, nil
}

// Name returns the transform name.
func (ct *ConditionalTransform) Name() string {
	return fmt.Sprintf("Conditional(%s)", ct.transform.Name())
}

// Validate validates the transform.
func (ct *ConditionalTransform) Validate() error {
	if ct.transform == nil {
		return fmt.Errorf("%w: transform is nil", ErrInvalidConfig)
	}
	if ct.condition == nil {
		return fmt.Errorf("%w: condition is nil", ErrInvalidConfig)
	}
	return ct.transform.Validate()
}

// Common condition builders

// VolumeAbove returns a condition that checks if volume is above threshold.
func VolumeAbove(threshold float64) Condition {
	return func(c *market.Candle) bool {
		return c.Volume > threshold
	}
}

// VolumeBelow returns a condition that checks if volume is below threshold.
func VolumeBelow(threshold float64) Condition {
	return func(c *market.Candle) bool {
		return c.Volume < threshold
	}
}

// PriceAbove returns a condition that checks if close price is above threshold.
func PriceAbove(threshold float64) Condition {
	return func(c *market.Candle) bool {
		return c.Close > threshold
	}
}

// PriceBelow returns a condition that checks if close price is below threshold.
func PriceBelow(threshold float64) Condition {
	return func(c *market.Candle) bool {
		return c.Close < threshold
	}
}

// PriceInRange returns a condition that checks if price is within range.
func PriceInRange(min, max float64) Condition {
	return func(c *market.Candle) bool {
		return c.Close >= min && c.Close <= max
	}
}

// VolatilityAbove returns a condition that checks if volatility is above threshold.
// Volatility is measured as (High - Low) / Close.
func VolatilityAbove(threshold float64) Condition {
	return func(c *market.Candle) bool {
		if c.Close == 0 {
			return false
		}
		volatility := (c.High - c.Low) / c.Close
		return volatility > threshold
	}
}

// BullishCandle returns a condition that checks if candle is bullish.
func BullishCandle() Condition {
	return func(c *market.Candle) bool {
		return c.Close > c.Open
	}
}

// BearishCandle returns a condition that checks if candle is bearish.
func BearishCandle() Condition {
	return func(c *market.Candle) bool {
		return c.Close < c.Open
	}
}

// AndCondition combines multiple conditions with AND logic.
func AndCondition(conditions ...Condition) Condition {
	return func(c *market.Candle) bool {
		for _, cond := range conditions {
			if !cond(c) {
				return false
			}
		}
		return true
	}
}

// OrCondition combines multiple conditions with OR logic.
func OrCondition(conditions ...Condition) Condition {
	return func(c *market.Candle) bool {
		for _, cond := range conditions {
			if cond(c) {
				return true
			}
		}
		return false
	}
}

// NotCondition negates a condition.
func NotCondition(condition Condition) Condition {
	return func(c *market.Candle) bool {
		return !condition(c)
	}
}

// SwitchTransform applies different transforms based on conditions.
type SwitchTransform struct {
	cases         []switchCase
	defaultCase   Transform
	config        TransformConfig
}

type switchCase struct {
	condition Condition
	transform Transform
}

// NewSwitchTransform creates a switch transform with default case.
func NewSwitchTransform(defaultTransform Transform) *SwitchTransform {
	return &SwitchTransform{
		cases:       make([]switchCase, 0),
		defaultCase: defaultTransform,
		config:      DefaultTransformConfig(),
	}
}

// AddCase adds a condition-transform case.
func (st *SwitchTransform) AddCase(condition Condition, transform Transform) {
	st.cases = append(st.cases, switchCase{
		condition: condition,
		transform: transform,
	})
}

// Apply applies the first matching transform, or default if none match.
func (st *SwitchTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	// Group candles by matching case
	groups := make(map[int][]*market.Candle)
	indices := make(map[int][]int) // Track original indices

	for i, candle := range candles {
		matched := false
		for caseIdx, c := range st.cases {
			if c.condition(candle) {
				groups[caseIdx] = append(groups[caseIdx], candle)
				indices[caseIdx] = append(indices[caseIdx], i)
				matched = true
				break // First matching case
			}
		}
		if !matched {
			// Default case
			groups[-1] = append(groups[-1], candle)
			indices[-1] = append(indices[-1], i)
		}
	}

	// Apply transforms to each group
	result := CopyCandles(candles)
	
	for caseIdx, group := range groups {
		if len(group) == 0 {
			continue
		}

		var transform Transform
		if caseIdx == -1 {
			transform = st.defaultCase
		} else {
			transform = st.cases[caseIdx].transform
		}

		if transform == nil {
			continue // Skip if no transform
		}

		transformed, err := transform.Apply(group)
		if err != nil {
			return nil, fmt.Errorf("switch case %d transform failed: %w", caseIdx, err)
		}

		// Put transformed candles back
		for i, idx := range indices[caseIdx] {
			if i < len(transformed) {
				result[idx] = transformed[i]
			}
		}
	}

	return result, nil
}

// Name returns the transform name.
func (st *SwitchTransform) Name() string {
	return fmt.Sprintf("Switch(%d cases)", len(st.cases))
}

// Validate validates the switch transform.
func (st *SwitchTransform) Validate() error {
	if st.defaultCase != nil {
		if err := st.defaultCase.Validate(); err != nil {
			return fmt.Errorf("default case validation failed: %w", err)
		}
	}

	for i, c := range st.cases {
		if c.condition == nil {
			return fmt.Errorf("%w: case %d has nil condition", ErrInvalidConfig, i)
		}
		if c.transform == nil {
			return fmt.Errorf("%w: case %d has nil transform", ErrInvalidConfig, i)
		}
		if err := c.transform.Validate(); err != nil {
			return fmt.Errorf("case %d validation failed: %w", i, err)
		}
	}

	return nil
}

// FilterTransform filters candles based on a condition.
type FilterTransform struct {
	condition Condition
	config    TransformConfig
}

// NewFilterTransform creates a filter transform.
func NewFilterTransform(condition Condition) *FilterTransform {
	return &FilterTransform{
		condition: condition,
		config:    DefaultTransformConfig(),
	}
}

// Apply filters candles based on condition.
func (ft *FilterTransform) Apply(candles []*market.Candle) ([]*market.Candle, error) {
	if err := ValidateCandles(candles); err != nil {
		return nil, err
	}

	result := make([]*market.Candle, 0, len(candles))
	for _, candle := range candles {
		if ft.condition(candle) {
			result = append(result, candle)
		}
	}

	return result, nil
}

// Name returns the transform name.
func (ft *FilterTransform) Name() string {
	return "Filter"
}

// Validate validates the filter transform.
func (ft *FilterTransform) Validate() error {
	if ft.condition == nil {
		return fmt.Errorf("%w: condition is nil", ErrInvalidConfig)
	}
	return nil
}
