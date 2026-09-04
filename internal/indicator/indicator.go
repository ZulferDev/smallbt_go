package indicator

import (
	"github.com/1jehuang/backtest/internal/market"
)

// Indicator defines the interface for all technical indicators.
// Indicators must be stateless and deterministic - given the same input,
// they must produce the same output.
type Indicator interface {
	// Name returns the unique identifier for this indicator.
	Name() string

	// Calculate computes the indicator value given the market data context.
	// The context provides access to historical candles and other indicators.
	Calculate(ctx *Context) (Value, error)
}

// Context provides the evaluation context for indicator calculation.
type Context struct {
	// Current candle being evaluated
	Current market.Candle

	// Historical candles (most recent last)
	// Candles[len(Candles)-1] is the current candle
	Candles []market.Candle

	// Symbol being evaluated
	Symbol market.Symbol

	// Timeframe of the data
	Timeframe market.Timeframe

	// IndicatorValues provides access to other computed indicators
	// Key is indicator name, value is the computed value
	IndicatorValues map[string]Value

	// BarIndex is the current bar index (0-based from start of data)
	BarIndex int
}

// Value represents an indicator's computed value.
// It handles the case where an indicator may not have a value
// (e.g., not enough data points).
type Value struct {
	// Valid indicates whether the value is valid
	Valid bool

	// Value is the computed indicator value
	Value float64

	// Metadata for advanced indicators (optional)
	Metadata map[string]float64
}

// NewValue creates a valid indicator value.
func NewValue(v float64) Value {
	return Value{Valid: true, Value: v}
}

// InvalidValue represents an invalid/uncomputed indicator value.
var InvalidValue = Value{Valid: false}

// Config represents the configuration for an indicator.
// This is used by the registry to create indicators with parameters.
type Config struct {
	// Type is the indicator type (e.g., "sma", "ema", "rsi")
	Type string

	// Period is the lookback period for the indicator
	Period int

	// Source specifies which price to use (e.g., "close", "open", "high", "low", "volume")
	Source string

	// Additional parameters for complex indicators
	Params map[string]interface{}
}

// SourcePrice extracts the specified price from a candle.
func SourcePrice(candle market.Candle, source string) float64 {
	switch source {
	case "open":
		return candle.Open
	case "high":
		return candle.High
	case "low":
		return candle.Low
	case "close", "":
		return candle.Close
	case "volume":
		return candle.Volume
	default:
		return candle.Close // default to close
	}
}
