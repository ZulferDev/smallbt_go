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
	strategy        *ast.Strategy
	registry        *indicator.Registry
	indicators      map[string]indicator.Indicator
	context         *indicator.Context
	values          map[string]float64     // Current indicator values
	prevValues      map[string]float64     // Previous indicator values for cross detection
	candles         []market.Candle        // Historical candles
	state           map[string]interface{} // Strategy state variables
}

// NewEvaluator creates a new strategy evaluator.
func NewEvaluator(strat *ast.Strategy, registry *indicator.Registry, symbol market.Symbol, timeframe market.Timeframe) *Evaluator {
	e := &Evaluator{
		strategy:   strat,
		registry:   registry,
		indicators: make(map[string]indicator.Indicator),
		values:     make(map[string]float64),
		prevValues: make(map[string]float64),
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
	if ci.eval.context.BarIndex == 0 {
		fmt.Printf("[DEBUG] Composite indicator '%s': type=%s, left=%s, right=%s, args=%v\n", 
			ci.name, ci.def.Type, ci.def.Left, ci.def.Right, args)
	}
	
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
	// Add candle to history first to get correct barIndex
	e.candles = append(e.candles, candle)
	barIndex := len(e.candles) - 1
	fmt.Printf("[DEBUG] Bar %d: UpdateCandle START\n", barIndex)
	// Save previous indicator values before calculating new ones
	// This is needed for cross detection
	for k, v := range e.values {
		e.prevValues[k] = v
	}

	// Update context
	e.context.Current = candle
	e.context.Candles = e.candles
	e.context.BarIndex = barIndex

	// Calculate basic indicators first
	for name, ind := range e.indicators {
		// Skip composite indicators in first pass
		if _, isComposite := ind.(*compositeIndicator); isComposite {
			continue
		}

		value, err := ind.Calculate(e.context)
		if err != nil {
			return fmt.Errorf("calculate indicator %s: %w", name, err)
		}
		e.context.IndicatorValues[name] = value
		if value.Valid {
			e.values[name] = value.Value
			if barIndex < 25 && (name == "volume_avg" || name == "ema_fast" || name == "ema_slow") {
				fmt.Printf("[DEBUG] Bar %d: Stored basic indicator '%s' = %.4f (Valid=%v)\n", barIndex, name, value.Value, value.Valid)
			}
		} else {
			e.values[name] = 0
			if barIndex < 25 && (name == "volume_avg" || name == "ema_fast" || name == "ema_slow") {
				fmt.Printf("[DEBUG] Bar %d: Stored basic indicator '%s' = 0 (NOT VALID)\n", barIndex, name)
			}
		}
	}

	// Calculate composite indicators second (after basic indicators have values)
	// Use topological sort to handle dependencies between composite indicators
	compositeNames := make([]string, 0)
	for name := range e.indicators {
		if _, isComposite := e.indicators[name].(*compositeIndicator); isComposite {
			compositeNames = append(compositeNames, name)
		}
	}
	
	if barIndex == 0 {
		fmt.Printf("[DEBUG] Found %d composite indicators: %v\n", len(compositeNames), compositeNames)
	}

	// Topologically sort composite indicators based on dependencies
	sortedCompositeNames, err := e.topologicalSortCompositeIndicators(compositeNames)
	if err != nil {
		fmt.Printf("[DEBUG] Error sorting composite indicators: %v\n", err)
		return fmt.Errorf("sort composite indicators: %w", err)
	}
	
	if barIndex == 0 {
		fmt.Printf("[DEBUG] Sorted composite indicators: %v\n", sortedCompositeNames)
	}

	fmt.Printf("[DEBUG] Bar %d: About to loop through %d composite indicators\n", barIndex, len(sortedCompositeNames))
	for _, name := range sortedCompositeNames {
		fmt.Printf("[DEBUG] Bar %d: Processing composite indicator '%s'\n", barIndex, name)
		ind := e.indicators[name]
		if _, isComposite := ind.(*compositeIndicator); !isComposite {
			fmt.Printf("[DEBUG] Skipping non-composite indicator: %s\n", name)
			continue
		}

		value, err := ind.Calculate(e.context)
		if err != nil {
			fmt.Printf("[DEBUG] Bar %d: ERROR calculating composite indicator '%s': %v\n", barIndex, name, err)
			return fmt.Errorf("calculate indicator %s: %w", name, err)
		}
		e.context.IndicatorValues[name] = value
		if value.Valid {
			e.values[name] = value.Value
			fmt.Printf("[DEBUG] Bar %d: Stored composite indicator '%s' value: %.4f\n", barIndex, name, e.values[name])
		} else {
			e.values[name] = 0
			fmt.Printf("[DEBUG] Bar %d: Composite indicator '%s' is invalid, stored 0\n", barIndex, name)
		}
		
		if e.context.BarIndex == 0 {
			fmt.Printf("[DEBUG] Bar %d: About to calculate composite indicator '%s'\n", barIndex, name)
			fmt.Printf("[DEBUG] Bar %d: Stored composite indicator '%s' value: %.4f\n", barIndex, name, e.values[name])
		}
	}
	
	fmt.Printf("[DEBUG] Bar %d: After composite indicator loop: volume_ratio=%v\n", barIndex, e.values["volume_ratio"])
	fmt.Printf("[DEBUG] Bar %d: UpdateCandle END\n", barIndex)

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
	barIndex := len(e.candles) - 1

	// DEBUG: Log when evaluate is called
	fmt.Printf("[DEBUG] Bar %d: Evaluate START\n", barIndex)
	fmt.Printf("[EVALUATE] candle=%d, hasPosition=%v, positionSide=%s, indicators=%d\n", 
		len(e.candles), hasPosition, positionSide, len(e.prevValues))

	// Evaluate entry conditions
	if !hasPosition {
		// Check long entry
		if e.strategy.Entry.Long != nil {
			fmt.Printf("[EVALUATE] Checking long entry condition\n")
			shouldEntry := e.evaluateCondition(e.strategy.Entry.Long)
			fmt.Printf("[EVALUATE] Long entry result: %v\n", shouldEntry)
			if shouldEntry {
				signals = append(signals, signal.Signal{
					Type:      signal.SignalTypeLongEntry,
					Timestamp: currentCandle.Timestamp,
					Price:     currentCandle.Close,
					Reason:    "long entry signal",
				})
				fmt.Printf("[SIGNAL] Generated long entry signal at bar %d, price=%.2f\n", len(e.candles), currentCandle.Close)
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

	fmt.Printf("[DEBUG] Bar %d: Returning %d signals\n", len(e.candles)-1, len(signals))
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
	fmt.Printf("[DEBUG] Bar %d: evaluateFunction(%s, %v)\n", e.context.BarIndex, fn, args)
	if len(args) < 2 {
		fmt.Printf("[DEBUG] evaluateFunction: insufficient args (%d)\n", len(args))
		return false
	}

	// Resolve arguments to values
	values := make([]float64, len(args))
	for i, arg := range args {
		val, err := e.resolveValue(arg)
		if err != nil {
			fmt.Printf("[DEBUG] evaluateFunction: resolveValue error for arg %v: %v\n", arg, err)
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
		// cross_above: [line1, line2]
		// returns true if line1 crossed above line2 between bar[t-1] and bar[t]
		if len(args) < 2 {
			return false
		}
		return e.crossAbove(args[0], args[1])
	case "cross_below":
		// cross_below: [line1, line2]
		// returns true if line1 crossed below line2 between bar[t-1] and bar[t]
		if len(args) < 2 {
			return false
		}
		return e.crossBelow(args[0], args[1])
	case "rising":
		// rising: [value]
		// returns true if value is rising (current > previous)
		if len(args) < 1 {
			return false
		}
		return e.isRising(args[0])
	case "falling":
		// falling: [value]
		// returns true if value is falling (current < previous)
		if len(args) < 1 {
			return false
		}
		return e.isFalling(args[0])
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
		if e.context.BarIndex < 5 && (fn == "divide" || fn == "add" || fn == "multiply" || fn == "subtract") {
			fmt.Printf("[DEBUG] resolveValue: arg[%d]=%v -> value=%.4f\n", i, arg, val)
		}
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
		if e.context.BarIndex < 5 {
			fmt.Printf("[DEBUG] divide: values=%v (from args=%v)\n", values, args)
		}
		if values[1] == 0 {
			// Return 0 instead of error - indicator not ready yet
			fmt.Printf("[DEBUG] divide by zero at bar %d: values[0]=%.4f, values[1]=%.4f\n", e.context.BarIndex, values[0], values[1])
			return 0, nil
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

// topologicalSortCompositeIndicators sorts composite indicators by their dependencies.
func (e *Evaluator) topologicalSortCompositeIndicators(names []string) ([]string, error) {
	// Build dependency graph
	deps := make(map[string][]string)
	for _, name := range names {
		ind, ok := e.indicators[name].(*compositeIndicator)
		if !ok {
			continue
		}

		deps[name] = make([]string, 0)
		if ind.def.Left != "" {
			deps[name] = append(deps[name], ind.def.Left)
		}
		if ind.def.Right != "" {
			deps[name] = append(deps[name], ind.def.Right)
		}
	}

	// Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)
	for _, name := range names {
		inDegree[name] = 0
	}

	for _, name := range names {
		for _, dep := range deps[name] {
			// Only count dependencies on other composite indicators
			if _, isComposite := inDegree[dep]; isComposite {
				inDegree[name]++ // name depends on dep
			}
		}
	}

	queue := make([]string, 0)
	for _, name := range names {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	sorted := make([]string, 0)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		// For each indicator that depends on current
		for _, name := range names {
			for _, dep := range deps[name] {
				if dep == current {
					inDegree[name]--
					if inDegree[name] == 0 {
						queue = append(queue, name)
					}
				}
			}
		}
	}

	// Check for cycles
	if len(sorted) != len(names) {
		return nil, fmt.Errorf("circular dependency detected in composite indicators")
	}

	return sorted, nil
}

// GetIndicatorValue returns the current value of an indicator by name.
func (e *Evaluator) GetIndicatorValue(name string) (float64, error) {
	val, ok := e.values[name]
	if !ok {
		return 0, fmt.Errorf("indicator %q not found", name)
	}

	if val == 0 {
		// Check if indicator exists but has no value yet
		_, hasIndicator := e.indicators[name]
		if hasIndicator {
			return 0, fmt.Errorf("indicator %q has no values yet", name)
		}
	}

	return val, nil
}

// crossAbove detects if line1 crossed above line2 between bar[t-1] and bar[t]
func (e *Evaluator) crossAbove(line1Arg, line2Arg interface{}) bool {
	if len(e.candles) < 2 {
		fmt.Printf("[DEBUG] crossAbove: not enough candles (%d)\n", len(e.candles))
		return false
	}

	// Get current values
	val1Current, err1 := e.resolveValueAt(line1Arg, len(e.candles)-1)
	val2Current, err2 := e.resolveValueAt(line2Arg, len(e.candles)-1)
	if err1 != nil || err2 != nil {
		fmt.Printf("[DEBUG] crossAbove: error getting current values: %v, %v\n", err1, err2)
		return false
	}

	// Get previous values
	val1Prev, err1 := e.resolveValueAt(line1Arg, len(e.candles)-2)
	val2Prev, err2 := e.resolveValueAt(line2Arg, len(e.candles)-2)
	if err1 != nil || err2 != nil {
		fmt.Printf("[DEBUG] crossAbove: error getting previous values: %v, %v\n", err1, err2)
		return false
	}

	result := val1Prev <= val2Prev && val1Current > val2Current
	fmt.Printf("[DEBUG] crossAbove(%v, %v): prev=(%.2f, %.2f) curr=(%.2f, %.2f) -> %v\n",
		line1Arg, line2Arg, val1Prev, val2Prev, val1Current, val2Current, result)

	// cross_above: line1 was below line2, now above
	return result
}

// crossBelow detects if line1 crossed below line2 between bar[t-1] and bar[t]
func (e *Evaluator) crossBelow(line1Arg, line2Arg interface{}) bool {
	if len(e.candles) < 2 {
		return false
	}

	// Get current values
	val1Current, err1 := e.resolveValueAt(line1Arg, len(e.candles)-1)
	val2Current, err2 := e.resolveValueAt(line2Arg, len(e.candles)-1)
	if err1 != nil || err2 != nil {
		return false
	}

	// Get previous values
	val1Prev, err1 := e.resolveValueAt(line1Arg, len(e.candles)-2)
	val2Prev, err2 := e.resolveValueAt(line2Arg, len(e.candles)-2)
	if err1 != nil || err2 != nil {
		return false
	}

	// cross_below: line1 was above line2, now below
	return val1Prev >= val2Prev && val1Current < val2Current
}

// isRising detects if value is rising (current > previous)
func (e *Evaluator) isRising(arg interface{}) bool {
	if len(e.candles) < 2 {
		return false
	}

	valCurrent, err := e.resolveValueAt(arg, len(e.candles)-1)
	if err != nil {
		return false
	}

	valPrev, err := e.resolveValueAt(arg, len(e.candles)-2)
	if err != nil {
		return false
	}

	return valCurrent > valPrev
}

// isFalling detects if value is falling (current < previous)
func (e *Evaluator) isFalling(arg interface{}) bool {
	if len(e.candles) < 2 {
		return false
	}

	valCurrent, err := e.resolveValueAt(arg, len(e.candles)-1)
	if err != nil {
		return false
	}

	valPrev, err := e.resolveValueAt(arg, len(e.candles)-2)
	if err != nil {
		return false
	}

	return valCurrent < valPrev
}

// resolveValueAt resolves the value of an argument at a specific bar index
func (e *Evaluator) resolveValueAt(arg interface{}, barIndex int) (float64, error) {
	if barIndex < 0 || barIndex >= len(e.candles) {
		return 0, fmt.Errorf("bar index out of range: %d", barIndex)
	}

	switch v := arg.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		// For indicator values at previous bars, use prevValues
		// For current bar (last candle), use current values
		currentBarIndex := len(e.candles) - 1
		if barIndex < currentBarIndex {
			// This is a historical request - use prevValues
			if val, ok := e.prevValues[v]; ok {
				return val, nil
			}
		} else {
			// This is current bar - use current values
			if val, ok := e.values[v]; ok {
				return val, nil
			}
		}
		// Check if it's a price field
		candle := e.candles[barIndex]
		switch v {
		case "open":
			return candle.Open, nil
		case "high":
			return candle.High, nil
		case "low":
			return candle.Low, nil
		case "close":
			return candle.Close, nil
		case "volume":
			return candle.Volume, nil
		default:
			return 0, fmt.Errorf("unknown value reference: %s", v)
		}
	default:
		return 0, fmt.Errorf("cannot resolve value: %v", arg)
	}
}
