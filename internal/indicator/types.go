package indicator

import (
	"github.com/1jehuang/backtest/internal/market"
)

// IndicatorValue represents the output of an indicator calculation.
type IndicatorValue struct {
	Value     float64
	IsValid   bool
	Timestamp int64 // Bar index or timestamp
}

// Indicator is the interface that all indicators must implement.
type Indicator interface {
	// Name returns the indicator name.
	Name() string

	// Add adds a new value to the indicator.
	Add(value float64)

	// Value returns the current indicator value.
	Value() IndicatorValue

	// Reset resets the indicator to initial state.
	Reset()

	// WarmupPeriod returns the number of bars needed before the indicator is valid.
	WarmupPeriod() int
}

// IndicatorRegistry manages indicator implementations.
type IndicatorRegistry struct {
	indicators map[string]func(params map[string]interface{}) Indicator
}

// NewIndicatorRegistry creates a new indicator registry.
func NewIndicatorRegistry() *IndicatorRegistry {
	return &IndicatorRegistry{
		indicators: make(map[string]func(params map[string]interface{}) Indicator),
	}
}

// Register registers a new indicator type.
func (r *IndicatorRegistry) Register(name string, factory func(params map[string]interface{}) Indicator) {
	r.indicators[name] = factory
}

// Create creates an indicator by name.
func (r *IndicatorRegistry) Create(name string, params map[string]interface{}) Indicator {
	factory, exists := r.indicators[name]
	if !exists {
		return nil
	}
	return factory(params)
}

// Exists checks if an indicator type is registered.
func (r *IndicatorRegistry) Exists(name string) bool {
	_, exists := r.indicators[name]
	return exists
}

// List returns all registered indicator names.
func (r *IndicatorRegistry) List() []string {
	names := make([]string, 0, len(r.indicators))
	for name := range r.indicators {
		names = append(names, name)
	}
	return names
}

// EvaluationContext provides context for indicator evaluation.
type EvaluationContext struct {
	Candle    market.Candle
	BarIndex  int
	Timestamp int64
	Symbol    market.Symbol
	Timeframe market.Timeframe
	Values    map[string]float64 // Other indicator/field values
}

// NewEvaluationContext creates a new evaluation context.
func NewEvaluationContext(candle market.Candle, barIndex int, symbol market.Symbol, timeframe market.Timeframe) *EvaluationContext {
	return &EvaluationContext{
		Candle:    candle,
		BarIndex:  barIndex,
		Symbol:    symbol,
		Timeframe: timeframe,
		Values:    make(map[string]float64),
	}
}
