package evaluator

import (
	"fmt"

	"github.com/1jehuang/backtest/internal/indicator"
	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/signal"
	"github.com/1jehuang/backtest/internal/strategy/ast"
)

// Evaluator evaluates a strategy against market data.
type Evaluator struct {
	strategy   *ast.Strategy
	registry   *indicator.Registry
	indicators map[string]indicator.Indicator
	context    *indicator.Context
	values     map[string]float64 // Current indicator values
	candles    []market.Candle    // Historical candles
}

// NewEvaluator creates a new strategy evaluator.
func NewEvaluator(strat *ast.Strategy, registry *indicator.Registry, symbol market.Symbol, timeframe market.Timeframe) *Evaluator {
	return &Evaluator{
		strategy:   strat,
		registry:   registry,
		indicators: make(map[string]indicator.Indicator),
		values:     make(map[string]float64),
		context: &indicator.Context{
			Symbol:    symbol,
			Timeframe: timeframe,
		},
		candles: make([]market.Candle, 0),
	}
}

// Initialize initializes indicators defined in the strategy.
func (e *Evaluator) Initialize() error {
	for name, def := range e.strategy.Indicators {
		// Create indicator based on type
		config := indicator.Config{
			Type:   def.Type,
			Params: def.Params,
		}
		ind, err := e.registry.Create(config)
		if err != nil {
			return fmt.Errorf("create indicator %s: %w", name, err)
		}
		e.indicators[name] = ind
	}
	return nil
}

// UpdateCandle updates the evaluator with a new candle and calculates indicators.
func (e *Evaluator) UpdateCandle(candle market.Candle) error {
	// Add candle to history
	e.candles = append(e.candles, candle)

	// Update context
	e.context.Current = candle
	e.context.Candles = e.candles
	e.context.BarIndex = len(e.candles) - 1

	// Calculate all indicators
	for name, ind := range e.indicators {
		value, err := ind.Calculate(e.context)
		if err != nil {
			return fmt.Errorf("calculate indicator %s: %w", name, err)
		}
		e.context.IndicatorValues[name] = value
		if value.Valid {
			e.values[name] = value.Value
		} else {
			e.values[name] = 0
		}
	}

	return nil
}

// Evaluate evaluates the strategy and returns signals.
func (e *Evaluator) Evaluate(hasPosition bool, positionSide string) ([]signal.Signal, error) {
	var signals []signal.Signal

	// Check if we have enough data
	if len(e.candles) == 0 {
		return signals, nil
	}

	currentCandle := e.candles[len(e.candles)-1]

	// Evaluate entry conditions
	if !hasPosition {
		// Check long entry
		if e.strategy.Entry.Long != nil {
			if e.evaluateCondition(e.strategy.Entry.Long) {
				signals = append(signals, signal.Signal{
					Type:      signal.SignalTypeLongEntry,
					Timestamp: currentCandle.Timestamp,
					Price:     currentCandle.Close,
					Reason:    "long entry signal",
				})
			}
		}

		// Check short entry
		if e.strategy.Entry.Short != nil {
			if e.evaluateCondition(e.strategy.Entry.Short) {
				signals = append(signals, signal.Signal{
					Type:      signal.SignalTypeShortEntry,
					Timestamp: currentCandle.Timestamp,
					Price:     currentCandle.Close,
					Reason:    "short entry signal",
				})
			}
		}
	} else {
		// Evaluate exit conditions
		// Check long exit
		if positionSide == "long" && e.strategy.Exit.Long != nil {
			if e.evaluateCondition(e.strategy.Exit.Long) {
				signals = append(signals, signal.Signal{
					Type:      signal.SignalTypeLongExit,
					Timestamp: currentCandle.Timestamp,
					Price:     currentCandle.Close,
					Reason:    "long exit signal",
				})
			}
		}

		// Check short exit
		if positionSide == "short" && e.strategy.Exit.Short != nil {
			if e.evaluateCondition(e.strategy.Exit.Short) {
				signals = append(signals, signal.Signal{
					Type:      signal.SignalTypeShortExit,
					Timestamp: currentCandle.Timestamp,
					Price:     currentCandle.Close,
					Reason:    "short exit signal",
				})
			}
		}
	}

	return signals, nil
}

// evaluateCondition evaluates a condition recursively.
func (e *Evaluator) evaluateCondition(cond *ast.Condition) bool {
	if cond == nil {
		return false
	}

	switch cond.Type {
	case "all":
		// All conditions must be true
		for _, c := range cond.Conditions {
			if !e.evaluateCondition(c) {
				return false
			}
		}
		return len(cond.Conditions) > 0

	case "any":
		// At least one condition must be true
		for _, c := range cond.Conditions {
			if e.evaluateCondition(c) {
				return true
			}
		}
		return false

	case "not":
		// Negate inner condition
		if len(cond.Conditions) > 0 {
			return !e.evaluateCondition(cond.Conditions[0])
		}
		return false

	case "func":
		// Evaluate function-based condition
		return e.evaluateFunction(cond.Function, cond.Args)

	default:
		return false
	}
}

// evaluateFunction evaluates a function-based condition.
func (e *Evaluator) evaluateFunction(fn string, args []interface{}) bool {
	if len(args) < 2 {
		return false
	}

	// Resolve arguments to values
	values := make([]float64, len(args))
	for i, arg := range args {
		val, err := e.resolveValue(arg)
		if err != nil {
			return false
		}
		values[i] = val
	}

	switch fn {
	case "gt":
		return values[0] > values[1]
	case "lt":
		return values[0] < values[1]
	case "ge":
		return values[0] >= values[1]
	case "le":
		return values[0] <= values[1]
	case "eq":
		return values[0] == values[1]
	case "ne":
		return values[0] != values[1]
	case "cross_above":
		// TODO: Implement cross detection with history
		return false
	case "cross_below":
		// TODO: Implement cross detection with history
		return false
	case "rising":
		// TODO: Implement rising detection with history
		return false
	case "falling":
		// TODO: Implement falling detection with history
		return false
	case "between":
		if len(values) >= 3 {
			return values[0] >= values[1] && values[0] <= values[2]
		}
		return false
	default:
		return false
	}
}

// resolveValue resolves an argument to a numeric value.
func (e *Evaluator) resolveValue(arg interface{}) (float64, error) {
	switch v := arg.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		// Check if it's an indicator value
		if val, ok := e.values[v]; ok {
			return val, nil
		}
		// Check if it's a price field
		if len(e.candles) == 0 {
			return 0, fmt.Errorf("no candles available")
		}
		currentCandle := e.candles[len(e.candles)-1]
		switch v {
		case "open":
			return currentCandle.Open, nil
		case "high":
			return currentCandle.High, nil
		case "low":
			return currentCandle.Low, nil
		case "close":
			return currentCandle.Close, nil
		case "volume":
			return currentCandle.Volume, nil
		default:
			return 0, fmt.Errorf("unknown value reference: %s", v)
		}
	default:
		return 0, fmt.Errorf("cannot resolve value: %v", arg)
	}
}

// GetIndicatorValues returns current indicator values.
func (e *Evaluator) GetIndicatorValues() map[string]float64 {
	return e.values
}

// GetCandleCount returns the number of candles processed.
func (e *Evaluator) GetCandleCount() int {
	return len(e.candles)
}
