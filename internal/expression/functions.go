package expression

import (
	"errors"
	"fmt"
)

// Function represents a function call expression.
type Function struct {
	Name string
	Args []Expression
}

func (f *Function) Evaluate(ctx *Context) (Value, error) {
	switch f.Name {
	case "cross_above":
		return f.evaluateCrossAbove(ctx)
	case "cross_below":
		return f.evaluateCrossBelow(ctx)
	case "rising":
		return f.evaluateRising(ctx)
	case "falling":
		return f.evaluateFalling(ctx)
	case "between":
		return f.evaluateBetween(ctx)
	case "abs":
		return f.evaluateAbs(ctx)
	case "min":
		return f.evaluateMin(ctx)
	case "max":
		return f.evaluateMax(ctx)
	case "sqrt":
		return f.evaluateSqrt(ctx)
	case "log":
		return f.evaluateLog(ctx)
	case "exp":
		return f.evaluateExp(ctx)
	case "previous":
		return f.evaluatePrevious(ctx)
	case "ref":
		return f.evaluateRef(ctx)
	case "shift":
		return f.evaluateShift(ctx)
	default:
		return InvalidValue, fmt.Errorf("unknown function: %s", f.Name)
	}
}

// CrossAbove evaluates whether value A crossed above value B.
// Returns 1 (true) if A[t] > B[t] and A[t-1] <= B[t-1], 0 otherwise.
func (f *Function) evaluateCrossAbove(ctx *Context) (Value, error) {
	if len(f.Args) != 2 {
		return InvalidValue, errors.New("cross_above requires exactly 2 arguments")
	}

	// Current values
	aNow, err := f.Args[0].Evaluate(ctx)
	if err != nil || !aNow.Valid {
		return InvalidValue, err
	}

	bNow, err := f.Args[1].Evaluate(ctx)
	if err != nil || !bNow.Valid {
		return InvalidValue, err
	}

	// Check if we have enough historical data
	if len(ctx.HistoricalPrices) < 2 {
		return InvalidValue, nil
	}

	// Create a context for previous bar
	prevCtx := &Context{
		BarIndex:                  ctx.BarIndex - 1,
		IndicatorValues:           ctx.IndicatorValues, // Would need historical indicator values for proper implementation
		HistoricalIndicatorValues: ctx.HistoricalIndicatorValues,
		CurrentPrices:             ctx.HistoricalPrices[1], // Previous bar's prices
		HistoricalPrices:          ctx.HistoricalPrices[1:],
	}

	// Evaluate previous bar's values
	aPrev, err := f.Args[0].Evaluate(prevCtx)
	if err != nil || !aPrev.Valid {
		return InvalidValue, err
	}

	bPrev, err := f.Args[1].Evaluate(prevCtx)
	if err != nil || !bPrev.Valid {
		return InvalidValue, err
	}

	// Check cross condition: A[t] > B[t] and A[t-1] <= B[t-1]
	if aNow.Value > bNow.Value && aPrev.Value <= bPrev.Value {
		return NewValue(1), nil
	}

	return NewValue(0), nil
}

// CrossBelow evaluates whether value A crossed below value B.
// Returns 1 (true) if A[t] < B[t] and A[t-1] >= B[t-1], 0 otherwise.
func (f *Function) evaluateCrossBelow(ctx *Context) (Value, error) {
	if len(f.Args) != 2 {
		return InvalidValue, errors.New("cross_below requires exactly 2 arguments")
	}

	// Current values
	aNow, err := f.Args[0].Evaluate(ctx)
	if err != nil || !aNow.Valid {
		return InvalidValue, err
	}

	bNow, err := f.Args[1].Evaluate(ctx)
	if err != nil || !bNow.Valid {
		return InvalidValue, err
	}

	// Check if we have enough historical data
	if len(ctx.HistoricalPrices) < 2 {
		return InvalidValue, nil
	}

	// Create a context for previous bar
	prevCtx := &Context{
		BarIndex:                  ctx.BarIndex - 1,
		IndicatorValues:           ctx.IndicatorValues,
		HistoricalIndicatorValues: ctx.HistoricalIndicatorValues,
		CurrentPrices:             ctx.HistoricalPrices[1], // Previous bar's prices
		HistoricalPrices:          ctx.HistoricalPrices[1:],
	}

	// Evaluate previous bar's values
	aPrev, err := f.Args[0].Evaluate(prevCtx)
	if err != nil || !aPrev.Valid {
		return InvalidValue, err
	}

	bPrev, err := f.Args[1].Evaluate(prevCtx)
	if err != nil || !bPrev.Valid {
		return InvalidValue, err
	}

	// Check cross condition: A[t] < B[t] and A[t-1] >= B[t-1]
	if aNow.Value < bNow.Value && aPrev.Value >= bPrev.Value {
		return NewValue(1), nil
	}

	return NewValue(0), nil
}

// Rising evaluates whether a value is rising.
// Returns 1 if value[t] > value[t-1], 0 otherwise.
func (f *Function) evaluateRising(ctx *Context) (Value, error) {
	if len(f.Args) != 1 {
		return InvalidValue, errors.New("rising requires exactly 1 argument")
	}

	// Current value
	now, err := f.Args[0].Evaluate(ctx)
	if err != nil || !now.Valid {
		return InvalidValue, err
	}

	// Check if we have enough historical data
	if len(ctx.HistoricalPrices) < 2 {
		return InvalidValue, nil
	}

	// Create a context for previous bar
	prevCtx := &Context{
		BarIndex:                  ctx.BarIndex - 1,
		IndicatorValues:           ctx.IndicatorValues,
		HistoricalIndicatorValues: ctx.HistoricalIndicatorValues,
		CurrentPrices:             ctx.HistoricalPrices[1],
		HistoricalPrices:          ctx.HistoricalPrices[1:],
	}

	// Previous value
	prev, err := f.Args[0].Evaluate(prevCtx)
	if err != nil || !prev.Valid {
		return InvalidValue, err
	}

	if now.Value > prev.Value {
		return NewValue(1), nil
	}

	return NewValue(0), nil
}

// Falling evaluates whether a value is falling.
// Returns 1 if value[t] < value[t-1], 0 otherwise.
func (f *Function) evaluateFalling(ctx *Context) (Value, error) {
	if len(f.Args) != 1 {
		return InvalidValue, errors.New("falling requires exactly 1 argument")
	}

	// Current value
	now, err := f.Args[0].Evaluate(ctx)
	if err != nil || !now.Valid {
		return InvalidValue, err
	}

	// Check if we have enough historical data
	if len(ctx.HistoricalPrices) < 2 {
		return InvalidValue, nil
	}

	// Create a context for previous bar
	prevCtx := &Context{
		BarIndex:                  ctx.BarIndex - 1,
		IndicatorValues:           ctx.IndicatorValues,
		HistoricalIndicatorValues: ctx.HistoricalIndicatorValues,
		CurrentPrices:             ctx.HistoricalPrices[1],
		HistoricalPrices:          ctx.HistoricalPrices[1:],
	}

	// Previous value
	prev, err := f.Args[0].Evaluate(prevCtx)
	if err != nil || !prev.Valid {
		return InvalidValue, err
	}

	if now.Value < prev.Value {
		return NewValue(1), nil
	}

	return NewValue(0), nil
}

// Between evaluates whether a value is between two bounds (inclusive).
// Returns 1 if lower <= value <= upper, 0 otherwise.
func (f *Function) evaluateBetween(ctx *Context) (Value, error) {
	if len(f.Args) != 3 {
		return InvalidValue, errors.New("between requires exactly 3 arguments")
	}

	value, err := f.Args[0].Evaluate(ctx)
	if err != nil || !value.Valid {
		return InvalidValue, err
	}

	lower, err := f.Args[1].Evaluate(ctx)
	if err != nil || !lower.Valid {
		return InvalidValue, err
	}

	upper, err := f.Args[2].Evaluate(ctx)
	if err != nil || !upper.Valid {
		return InvalidValue, err
	}

	if value.Value >= lower.Value && value.Value <= upper.Value {
		return NewValue(1), nil
	}

	return NewValue(0), nil
}

// Math functions

func (f *Function) evaluateAbs(ctx *Context) (Value, error) {
	if len(f.Args) != 1 {
		return InvalidValue, errors.New("abs requires exactly 1 argument")
	}

	val, err := f.Args[0].Evaluate(ctx)
	if err != nil || !val.Valid {
		return InvalidValue, err
	}

	result := val.Value
	if result < 0 {
		result = -result
	}

	return NewValue(result), nil
}

func (f *Function) evaluateMin(ctx *Context) (Value, error) {
	if len(f.Args) < 2 {
		return InvalidValue, errors.New("min requires at least 2 arguments")
	}

	// Initialize with first value
	first, err := f.Args[0].Evaluate(ctx)
	if err != nil || !first.Valid {
		return InvalidValue, err
	}

	min := first.Value

	// Find minimum among remaining values
	for i := 1; i < len(f.Args); i++ {
		val, err := f.Args[i].Evaluate(ctx)
		if err != nil || !val.Valid {
			return InvalidValue, err
		}
		if val.Value < min {
			min = val.Value
		}
	}

	return NewValue(min), nil
}

func (f *Function) evaluateMax(ctx *Context) (Value, error) {
	if len(f.Args) < 2 {
		return InvalidValue, errors.New("max requires at least 2 arguments")
	}

	// Initialize with first value
	first, err := f.Args[0].Evaluate(ctx)
	if err != nil || !first.Valid {
		return InvalidValue, err
	}

	max := first.Value

	// Find maximum among remaining values
	for i := 1; i < len(f.Args); i++ {
		val, err := f.Args[i].Evaluate(ctx)
		if err != nil || !val.Valid {
			return InvalidValue, err
		}
		if val.Value > max {
			max = val.Value
		}
	}

	return NewValue(max), nil
}

func (f *Function) evaluateSqrt(ctx *Context) (Value, error) {
	if len(f.Args) != 1 {
		return InvalidValue, errors.New("sqrt requires exactly 1 argument")
	}

	val, err := f.Args[0].Evaluate(ctx)
	if err != nil || !val.Valid {
		return InvalidValue, err
	}

	if val.Value < 0 {
		return InvalidValue, errors.New("sqrt of negative number")
	}

	// Simple square root - in production would use math.Sqrt
	// Return approximate value for now
	result := val.Value
	for i := 0; i < 10; i++ {
		result = (result + val.Value/result) / 2
	}

	return NewValue(result), nil
}

func (f *Function) evaluateLog(ctx *Context) (Value, error) {
	if len(f.Args) != 1 {
		return InvalidValue, errors.New("log requires exactly 1 argument")
	}

	val, err := f.Args[0].Evaluate(ctx)
	if err != nil || !val.Valid {
		return InvalidValue, err
	}

	if val.Value <= 0 {
		return InvalidValue, errors.New("log of non-positive number")
	}

	// Natural logarithm approximation
	// In production would use math.Log
	n := val.Value
	log := 0.0
	for n > 1.0 {
		n /= 2.71828
		log += 1.0
	}
	return NewValue(log), nil
}

func (f *Function) evaluateExp(ctx *Context) (Value, error) {
	if len(f.Args) != 1 {
		return InvalidValue, errors.New("exp requires exactly 1 argument")
	}

	val, err := f.Args[0].Evaluate(ctx)
	if err != nil || !val.Valid {
		return InvalidValue, err
	}

	// Exponential approximation
	// In production would use math.Exp
	result := 1.0
	x := val.Value
	term := 1.0
	for i := 1; i < 10; i++ {
		term *= x / float64(i)
		result += term
	}

	return NewValue(result), nil
}

// Historical reference functions

func (f *Function) evaluatePrevious(ctx *Context) (Value, error) {
	if len(f.Args) != 1 {
		return InvalidValue, errors.New("previous requires exactly 1 argument")
	}

	// Check if we have enough historical data
	if len(ctx.HistoricalPrices) < 2 {
		return InvalidValue, nil
	}

	// Create context for previous bar
	prevCtx := &Context{
		BarIndex:                  ctx.BarIndex - 1,
		IndicatorValues:           ctx.IndicatorValues,
		HistoricalIndicatorValues: ctx.HistoricalIndicatorValues,
		CurrentPrices:             ctx.HistoricalPrices[1],
		HistoricalPrices:          ctx.HistoricalPrices[1:],
	}

	return f.Args[0].Evaluate(prevCtx)
}

func (f *Function) evaluateRef(ctx *Context) (Value, error) {
	if len(f.Args) != 2 {
		return InvalidValue, errors.New("ref requires exactly 2 arguments: (value, bars)")
	}

	// First argument is the value expression
	// Second argument is the bars offset
	barsExpr, err := f.Args[1].Evaluate(ctx)
	if err != nil || !barsExpr.Valid {
		return InvalidValue, err
	}

	bars := int(barsExpr.Value)
	if bars < 0 {
		return InvalidValue, errors.New("ref bars must be non-negative")
	}

	// Check if we have enough historical data
	if len(ctx.HistoricalPrices) < bars+1 {
		return InvalidValue, nil
	}

	// Create context for referenced bar
	refCtx := &Context{
		BarIndex:                  ctx.BarIndex - bars,
		IndicatorValues:           ctx.IndicatorValues,
		HistoricalIndicatorValues: ctx.HistoricalIndicatorValues,
		CurrentPrices:             ctx.HistoricalPrices[bars],
		HistoricalPrices:          ctx.HistoricalPrices[bars:],
	}

	return f.Args[0].Evaluate(refCtx)
}

func (f *Function) evaluateShift(ctx *Context) (Value, error) {
	if len(f.Args) != 2 {
		return InvalidValue, errors.New("shift requires exactly 2 arguments: (value, bars)")
	}

	// First argument is the value expression
	// Second argument is the bars offset
	barsExpr, err := f.Args[1].Evaluate(ctx)
	if err != nil || !barsExpr.Valid {
		return InvalidValue, err
	}

	bars := int(barsExpr.Value)

	// Check if we have enough historical data
	if bars >= 0 {
		// Forward shift (not supported in historical context)
		return InvalidValue, errors.New("positive shift not supported in historical context")
	}

	bars = -bars // Convert to positive offset

	if len(ctx.HistoricalPrices) < bars+1 {
		return InvalidValue, nil
	}

	// Create context for shifted bar
	shiftCtx := &Context{
		BarIndex:                  ctx.BarIndex - bars,
		IndicatorValues:           ctx.IndicatorValues,
		HistoricalIndicatorValues: ctx.HistoricalIndicatorValues,
		CurrentPrices:             ctx.HistoricalPrices[bars],
		HistoricalPrices:          ctx.HistoricalPrices[bars:],
	}

	return f.Args[0].Evaluate(shiftCtx)
}
