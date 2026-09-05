package indicator

import (
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CachedIndicator extends Indicator with stateful, incremental updates.
// Unlike stateless Indicators that recalculate on every call, CachedIndicators
// maintain internal state and update incrementally for O(1) performance.
type CachedIndicator interface {
	Indicator

	// Update incrementally updates the indicator with a new candle.
	// Returns the computed value and whether it's valid (warm).
	// This is called once per bar in chronological order.
	Update(candle market.Candle, prevCandle *market.Candle) (Value, error)

	// IsWarm returns true when the indicator has enough data to produce valid values.
	IsWarm() bool

	// Reset clears the indicator's state for a new backtest run.
	Reset()

	// WarmupPeriod returns the number of bars needed before the indicator is warm.
	WarmupPeriod() int
}

// StateManager manages stateful indicator instances for a strategy evaluation.
// It ensures indicators are properly initialized, updated, and isolated between runs.
type StateManager struct {
	indicators map[string]CachedIndicator
	values     map[string]Value
	barIndex   int
}

// NewStateManager creates a new indicator state manager.
func NewStateManager() *StateManager {
	return &StateManager{
		indicators: make(map[string]CachedIndicator),
		values:     make(map[string]Value),
		barIndex:   -1,
	}
}

// Register adds a cached indicator to the manager.
func (sm *StateManager) Register(name string, indicator CachedIndicator) {
	sm.indicators[name] = indicator
	sm.values[name] = InvalidValue
}

// Update updates all indicators with a new candle.
// This should be called once per bar in chronological order.
func (sm *StateManager) Update(candle market.Candle, prevCandle *market.Candle) error {
	sm.barIndex++

	for name, indicator := range sm.indicators {
		value, err := indicator.Update(candle, prevCandle)
		if err != nil {
			return err
		}
		sm.values[name] = value
	}

	return nil
}

// GetValue returns the current cached value for an indicator.
func (sm *StateManager) GetValue(name string) Value {
	if value, ok := sm.values[name]; ok {
		return value
	}
	return InvalidValue
}

// AllWarm returns true when all indicators have warmed up.
func (sm *StateManager) AllWarm() bool {
	for _, indicator := range sm.indicators {
		if !indicator.IsWarm() {
			return false
		}
	}
	return true
}

// Reset resets all indicators for a new backtest run.
func (sm *StateManager) Reset() {
	for _, indicator := range sm.indicators {
		indicator.Reset()
	}
	for name := range sm.values {
		sm.values[name] = InvalidValue
	}
	sm.barIndex = -1
}

// MaxWarmupPeriod returns the maximum warmup period across all indicators.
func (sm *StateManager) MaxWarmupPeriod() int {
	maxPeriod := 0
	for _, indicator := range sm.indicators {
		period := indicator.WarmupPeriod()
		if period > maxPeriod {
			maxPeriod = period
		}
	}
	return maxPeriod
}
