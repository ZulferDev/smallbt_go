package expression

import (
	"github.com/ZulferDev/smallbt_go/internal/indicator"
	"testing"
)

// TestFunctionCrossAbove tests cross_above function.
func TestFunctionCrossAbove(t *testing.T) {
	fn := &Function{
		Name: "cross_above",
		Args: []Expression{
			&IdentifierExpr{Name: "ema_fast"},
			&IdentifierExpr{Name: "ema_slow"},
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{
			"ema_fast": {Valid: true, Value: 105.0},
			"ema_slow": {Valid: true, Value: 103.0},
		},
		BarIndex:                  1,
		HistoricalIndicatorValues: make(map[string][]indicator.Value),
		HistoricalPrices: []map[string]float64{
			{"close": 105.0, "ema_fast": 105.0, "ema_slow": 103.0}, // current
			{"close": 102.0, "ema_fast": 102.0, "ema_slow": 104.0}, // previous
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Current: 105 > 103 (true), Previous: 102 <= 104 (true) -> cross_above = true
	if !val.Valid || val.Value != 1.0 {
		t.Errorf("Expected cross_above to be true (1.0), got: %v", val)
	}
}

// TestFunctionCrossAboveNoCondition tests cross_above when condition not met.
func TestFunctionCrossAboveNoCondition(t *testing.T) {
	fn := &Function{
		Name: "cross_above",
		Args: []Expression{
			&IdentifierExpr{Name: "ema_fast"},
			&IdentifierExpr{Name: "ema_slow"},
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{
			"ema_fast": {Valid: true, Value: 105.0},
			"ema_slow": {Valid: true, Value: 103.0},
		},
		BarIndex:                  1,
		HistoricalIndicatorValues: make(map[string][]indicator.Value),
		HistoricalPrices: []map[string]float64{
			{"close": 105.0, "ema_fast": 105.0, "ema_slow": 103.0}, // current
			{"close": 106.0, "ema_fast": 106.0, "ema_slow": 102.0}, // previous
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Current: 105 > 103, Previous: 106 > 102 (already crossed) -> cross_above = false
	if !val.Valid || val.Value != 0.0 {
		t.Errorf("Expected cross_above to be false (0.0), got: %v", val)
	}
}

// TestFunctionCrossBelow tests cross_below function.
func TestFunctionCrossBelow(t *testing.T) {
	fn := &Function{
		Name: "cross_below",
		Args: []Expression{
			&IdentifierExpr{Name: "ema_fast"},
			&IdentifierExpr{Name: "ema_slow"},
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{
			"ema_fast": {Valid: true, Value: 102.0},
			"ema_slow": {Valid: true, Value: 104.0},
		},
		BarIndex:                  1,
		HistoricalIndicatorValues: make(map[string][]indicator.Value),
		HistoricalPrices: []map[string]float64{
			{"close": 102.0, "ema_fast": 102.0, "ema_slow": 104.0}, // current
			{"close": 105.0, "ema_fast": 105.0, "ema_slow": 103.0}, // previous
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Current: 102 < 104 (true), Previous: 105 >= 103 (true) -> cross_below = true
	if !val.Valid || val.Value != 1.0 {
		t.Errorf("Expected cross_below to be true (1.0), got: %v", val)
	}
}

// TestFunctionRising tests rising function.
func TestFunctionRising(t *testing.T) {
	fn := &Function{
		Name: "rising",
		Args: []Expression{
			&IdentifierExpr{Name: "close"},
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{},
		CurrentPrices: map[string]float64{
			"close": 105.0,
		},
		BarIndex: 1,
		HistoricalPrices: []map[string]float64{
			{"close": 105.0}, // current
			{"close": 102.0}, // previous
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Current: 105 > previous: 102 -> rising = true
	if !val.Valid || val.Value != 1.0 {
		t.Errorf("Expected rising to be true (1.0), got: %v", val)
	}
}

// TestFunctionRisingFalse tests rising when value is falling.
func TestFunctionRisingFalse(t *testing.T) {
	fn := &Function{
		Name: "rising",
		Args: []Expression{
			&IdentifierExpr{Name: "close"},
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{},
		CurrentPrices: map[string]float64{
			"close": 102.0,
		},
		BarIndex: 1,
		HistoricalPrices: []map[string]float64{
			{"close": 102.0}, // current
			{"close": 105.0}, // previous
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Current: 102 < previous: 105 -> rising = false
	if !val.Valid || val.Value != 0.0 {
		t.Errorf("Expected rising to be false (0.0), got: %v", val)
	}
}

// TestFunctionFalling tests falling function.
func TestFunctionFalling(t *testing.T) {
	fn := &Function{
		Name: "falling",
		Args: []Expression{
			&IdentifierExpr{Name: "close"},
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{},
		CurrentPrices: map[string]float64{
			"close": 102.0,
		},
		BarIndex: 1,
		HistoricalPrices: []map[string]float64{
			{"close": 102.0}, // current
			{"close": 105.0}, // previous
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Current: 102 < previous: 105 -> falling = true
	if !val.Valid || val.Value != 1.0 {
		t.Errorf("Expected falling to be true (1.0), got: %v", val)
	}
}

// TestFunctionBetween tests between function.
func TestFunctionBetween(t *testing.T) {
	fn := &Function{
		Name: "between",
		Args: []Expression{
			&LiteralExpr{Value: 50.0},
			&LiteralExpr{Value: 40.0},
			&LiteralExpr{Value: 60.0},
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// 50 is between 40 and 60 -> true
	if !val.Valid || val.Value != 1.0 {
		t.Errorf("Expected between to be true (1.0), got: %v", val)
	}
}

// TestFunctionBetweenFalse tests between when value is outside range.
func TestFunctionBetweenFalse(t *testing.T) {
	fn := &Function{
		Name: "between",
		Args: []Expression{
			&LiteralExpr{Value: 70.0},
			&LiteralExpr{Value: 40.0},
			&LiteralExpr{Value: 60.0},
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// 70 is not between 40 and 60 -> false
	if !val.Valid || val.Value != 0.0 {
		t.Errorf("Expected between to be false (0.0), got: %v", val)
	}
}

// TestFunctionAbs tests abs function.
func TestFunctionAbs(t *testing.T) {
	fn := &Function{
		Name: "abs",
		Args: []Expression{
			&LiteralExpr{Value: -42.5},
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid || val.Value != 42.5 {
		t.Errorf("Expected abs to be 42.5, got: %v", val)
	}
}

// TestFunctionMin tests min function.
func TestFunctionMin(t *testing.T) {
	fn := &Function{
		Name: "min",
		Args: []Expression{
			&LiteralExpr{Value: 10.0},
			&LiteralExpr{Value: 5.0},
			&LiteralExpr{Value: 15.0},
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid || val.Value != 5.0 {
		t.Errorf("Expected min to be 5.0, got: %v", val)
	}
}

// TestFunctionMax tests max function.
func TestFunctionMax(t *testing.T) {
	fn := &Function{
		Name: "max",
		Args: []Expression{
			&LiteralExpr{Value: 10.0},
			&LiteralExpr{Value: 5.0},
			&LiteralExpr{Value: 15.0},
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid || val.Value != 15.0 {
		t.Errorf("Expected max to be 15.0, got: %v", val)
	}
}

// TestFunctionPrevious tests previous function.
func TestFunctionPrevious(t *testing.T) {
	fn := &Function{
		Name: "previous",
		Args: []Expression{
			&IdentifierExpr{Name: "close"},
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{},
		CurrentPrices: map[string]float64{
			"close": 105.0,
		},
		BarIndex: 1,
		HistoricalPrices: []map[string]float64{
			{"close": 105.0}, // current
			{"close": 102.0}, // previous
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// previous(close) should return the previous bar's close value
	if !val.Valid || val.Value != 102.0 {
		t.Errorf("Expected previous(close) to be 102.0, got: %v", val)
	}
}

// TestFunctionRef tests ref function with bar offset.
func TestFunctionRef(t *testing.T) {
	fn := &Function{
		Name: "ref",
		Args: []Expression{
			&IdentifierExpr{Name: "close"},
			&LiteralExpr{Value: 2.0}, // 2 bars back
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{},
		CurrentPrices: map[string]float64{
			"close": 105.0,
		},
		BarIndex: 2,
		HistoricalPrices: []map[string]float64{
			{"close": 105.0}, // current (bar 2)
			{"close": 103.0}, // bar 1
			{"close": 100.0}, // bar 0 (2 bars back)
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// ref(close, 2) should return close from 2 bars back
	if !val.Valid || val.Value != 100.0 {
		t.Errorf("Expected ref(close, 2) to be 100.0, got: %v", val)
	}
}

// TestFunctionShift tests shift function (negative offset).
func TestFunctionShift(t *testing.T) {
	fn := &Function{
		Name: "shift",
		Args: []Expression{
			&IdentifierExpr{Name: "close"},
			&LiteralExpr{Value: -1.0}, // 1 bar back
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{},
		CurrentPrices: map[string]float64{
			"close": 105.0,
		},
		BarIndex: 1,
		HistoricalPrices: []map[string]float64{
			{"close": 105.0}, // current
			{"close": 102.0}, // 1 bar back
		},
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// shift(close, -1) should return close from 1 bar back
	if !val.Valid || val.Value != 102.0 {
		t.Errorf("Expected shift(close, -1) to be 102.0, got: %v", val)
	}
}

// TestFunctionInsufficientData tests function with insufficient historical data.
func TestFunctionInsufficientData(t *testing.T) {
	fn := &Function{
		Name: "rising",
		Args: []Expression{
			&IdentifierExpr{Name: "close"},
		},
	}

	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{},
		CurrentPrices: map[string]float64{
			"close": 105.0,
		},
		BarIndex:         0,
		HistoricalPrices: []map[string]float64{}, // No historical data
	}

	val, err := fn.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should return invalid value when insufficient data
	if val.Valid {
		t.Errorf("Expected invalid value when insufficient historical data")
	}
}

// TestFunctionWrongArgCount tests function with wrong argument count.
func TestFunctionWrongArgCount(t *testing.T) {
	fn := &Function{
		Name: "cross_above",
		Args: []Expression{
			&LiteralExpr{Value: 10.0},
			&LiteralExpr{Value: 20.0},
			&LiteralExpr{Value: 30.0}, // Extra arg
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := fn.Evaluate(ctx)

	if err == nil {
		t.Errorf("Expected error for wrong arg count")
	}

	if val.Valid {
		t.Errorf("Expected invalid value")
	}
}

// TestFunctionUnknown tests unknown function call.
func TestFunctionUnknown(t *testing.T) {
	fn := &Function{
		Name: "unknown_function",
		Args: []Expression{},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := fn.Evaluate(ctx)

	if err == nil {
		t.Errorf("Expected error for unknown function")
	}

	if val.Valid {
		t.Errorf("Expected invalid value")
	}
}
