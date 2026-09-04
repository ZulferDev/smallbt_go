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
	state      map[string]interface{} // Strategy state variables
}

// NewEvaluator creates a new strategy evaluator.
func NewEvaluator(strat *ast.Strategy, registry *indicator.Registry, symbol market.Symbol, timeframe market.Timeframe) *Evaluator {
	e := &Evaluator{
		strategy:   strat,
		registry:   registry,
		indicators: make(map[string]indicator.Indicator),
		values:     make(map[string]float64),
		context: &indicator.Context{
			Symbol:          symbol,
			Timeframe:       timeframe,
			IndicatorValues: make(map[string]indicator.Value),
		},
		candles: make([]market.Candle, 0),
		state:   make(map[string]interface{}),
	}

	// Initialize state variables from strategy definition
	e.initializeState()

	return e
}

// initializeState sets default values for strategy state variables.
func (e *Evaluator) initializeState() {
	if e.strategy == nil {
		return
	}

	for name, def := range e.strategy.State {
		e.state[name] = def.Default
	}
}

// Initialize initializes indicators defined in the strategy.
func (e *Evaluator) Initialize() error {
	// First pass: create all basic indicators
	for name, def := range e.strategy.Indicators {
		// Skip composite indicators - they'll be handled after basic indicators are created
		if isCompositeIndicator(def.Type) {
			continue
		}

		// Build config from AST fields
		config := indicator.Config{
			Type:   def.Type,
			Period: def.Period,
			Source: def.Source,
		}
		
		// If Period is 0, check Params map
		if config.Period == 0 && def.Params != nil {
			if p, ok := def.Params["period"]; ok {
				switch v := p.(type) {
				case int:
					config.Period = v
				case float64:
					config.Period = int(v)
				}
			}
		}
		
		// If Source is empty, check Params map
		if config.Source == "" && def.Params != nil {
			if s, ok := def.Params["source"]; ok {
				if sourceStr, ok := s.(string); ok {
					config.Source = sourceStr
				}
			}
		}
		
		ind, err := e.registry.Create(config)
		if err != nil {
			return fmt.Errorf("create indicator %s: %w", name, err)
		}
		e.indicators[name] = ind
	}

	// Second pass: create composite indicators
	for name, def := range e.strategy.Indicators {
		if !isCompositeIndicator(def.Type) {
			continue
		}

		// Create a wrapper indicator that delegates to EvaluateExpression
		e.indicators[name] = &compositeIndicator{
			name: name,
			def:  def,
			eval: e,
		}
	}

	return nil
}

// compositeIndicator is a wrapper for composite indicators that use EvaluateExpression.
type compositeIndicator struct {
	name string
	def  ast.IndicatorDef
	eval *Evaluator
}

// Name returns the indicator name.
func (ci *compositeIndicator) Name() string {
	return ci.name
}

// Calculate computes the composite indicator value.
func (ci *compositeIndicator) Calculate(ctx *indicator.Context) (indicator.Value, error) {
	// Extract args from struct fields (Left, Right) or Params map for backward compatibility
	args := make([]interface{}, 0)
	
	// Check struct fields first (new style)
	if ci.def.Left != "" {
		args = append(args, ci.def.Left)
	}
	if ci.def.Right != "" {
		args = append(args, ci.def.Right)
	}
	
	// Fall back to Params map for backward compatibility
	if len(args) == 0 {
		if left, ok := ci.def.Params["left"]; ok {
			args = append(args, left)
		}
		if right, ok := ci.def.Params["right"]; ok {
			args = append(args, right)
		}
		if additional, ok := ci.def.Params["additional"]; ok {
			args = append(args, additional)
		}
	}

	// Use the composite operation type as function name
	value, err := ci.eval.EvaluateExpression(ci.def.Type, args)
	if err != nil {
		return indicator.Value{}, fmt.Errorf("evaluate composite indicator %s: %w", ci.name, err)
	}

	return indicator.Value{
		Valid: true,
		Value: value,
	}, nil
}

// isCompositeIndicator checks if an indicator type is a composite operation.
func isCompositeIndicator(indicatorType string) bool {
	compositeTypes := map[string]bool{
		"add":      true,
		"subtract": true,
		"multiply": true,
		"divide":   true,
		"modulo":   true,
		"abs":      true,
		"min":      true,
		"max":      true,
	}
	_, ok := compositeTypes[indicatorType]
	return ok
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

// EvaluateExpression evaluates a numeric expression (used for composite indicators).
// Returns the result value and any error.
func (e *Evaluator) EvaluateExpression(fn string, args []interface{}) (float64, error) {
	// Validate argument count per function
	switch fn {
	case "abs":
		if len(args) != 1 {
			return 0, fmt.Errorf("abs requires exactly 1 argument, got %d", len(args))
		}
	case "add", "subtract", "multiply", "divide", "modulo":
		if len(args) < 2 {
			return 0, fmt.Errorf("%s requires at least 2 arguments, got %d", fn, len(args))
		}
	case "min", "max":
		if len(args) < 1 {
			return 0, fmt.Errorf("%s requires at least 1 argument, got %d", fn, len(args))
		}
	default:
		if len(args) < 2 {
			return 0, fmt.Errorf("%s requires at least 2 arguments, got %d", fn, len(args))
		}
	}

	// Resolve arguments to values
	values := make([]float64, len(args))
	for i, arg := range args {
		val, err := e.resolveValue(arg)
		if err != nil {
			return 0, err
		}
		values[i] = val
	}

	switch fn {
	case "add":
		result := values[0]
		for i := 1; i < len(values); i++ {
			result += values[i]
		}
		return result, nil
	case "subtract":
		return values[0] - values[1], nil
	case "multiply":
		result := values[0]
		for i := 1; i < len(values); i++ {
			result *= values[i]
		}
		return result, nil
	case "divide":
		if values[1] == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return values[0] / values[1], nil
	case "modulo":
		if values[1] == 0 {
			return 0, fmt.Errorf("modulo by zero")
		}
		return float64(int64(values[0]) % int64(values[1])), nil
	case "abs":
		if values[0] < 0 {
			return -values[0], nil
		}
		return values[0], nil
	case "min":
		result := values[0]
		for i := 1; i < len(values); i++ {
			if values[i] < result {
				result = values[i]
			}
		}
		return result, nil
	case "max":
		result := values[0]
		for i := 1; i < len(values); i++ {
			if values[i] > result {
				result = values[i]
			}
		}
		return result, nil
	default:
		return 0, fmt.Errorf("unknown function: %s", fn)
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
		// Check if it's a state variable
		if stateVal, ok := e.state[v]; ok {
			switch val := stateVal.(type) {
			case float64:
				return val, nil
			case int:
				return float64(val), nil
			case bool:
				if val {
					return 1.0, nil
				}
				return 0.0, nil
			default:
				return 0, fmt.Errorf("state variable %s has unsupported type: %T", v, stateVal)
			}
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

// GetState returns the current value of a state variable.
func (e *Evaluator) GetState(name string) (interface{}, error) {
	val, ok := e.state[name]
	if !ok {
		return nil, fmt.Errorf("state variable not found: %s", name)
	}
	return val, nil
}

// SetState sets the value of a state variable.
func (e *Evaluator) SetState(name string, value interface{}) error {
	// Check if state variable exists
	if _, ok := e.state[name]; !ok {
		return fmt.Errorf("state variable not found: %s", name)
	}
	e.state[name] = value
	return nil
}

// GetStateMap returns a copy of all state variables.
func (e *Evaluator) GetStateMap() map[string]interface{} {
	stateCopy := make(map[string]interface{})
	for k, v := range e.state {
		stateCopy[k] = v
	}
	return stateCopy
}

// SetStateMap updates all state variables from a map.
func (e *Evaluator) SetStateMap(stateMap map[string]interface{}) {
	for k, v := range stateMap {
		e.state[k] = v
	}
}
