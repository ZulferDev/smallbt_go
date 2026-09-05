package expression

import (
	"github.com/ZulferDev/smallbt_go/internal/indicator"
	"testing"
)

// TestLiteralExpr tests the literal expression evaluation.
func TestLiteralExpr(t *testing.T) {
	expr := &LiteralExpr{Value: 42.5}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 42.5 {
		t.Errorf("Expected value 42.5, got: %f", val.Value)
	}
}

// TestIdentifierExpr tests the identifier expression evaluation.
func TestIdentifierExpr(t *testing.T) {
	ctx := &Context{
		IndicatorValues: map[string]indicator.Value{
			"ema20":   {Valid: true, Value: 105.5},
			"ema50":   {Valid: true, Value: 102.3},
			"invalid": {Valid: false, Value: 0},
		},
		CurrentPrices: map[string]float64{
			"close":  103.5,
			"open":   102.0,
			"high":   104.0,
			"low":    101.5,
			"volume": 1000.0,
		},
		BarIndex: 0,
	}

	// Test indicator reference
	indicatorExpr := &IdentifierExpr{Name: "ema20"}
	val, err := indicatorExpr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error for indicator, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid indicator value")
	}

	if val.Value != 105.5 {
		t.Errorf("Expected indicator value 105.5, got: %f", val.Value)
	}

	// Test price reference
	priceExpr := &IdentifierExpr{Name: "close"}
	val, err = priceExpr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error for price, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid price value")
	}

	if val.Value != 103.5 {
		t.Errorf("Expected price value 103.5, got: %f", val.Value)
	}

	// Test invalid indicator
	invalidExpr := &IdentifierExpr{Name: "invalid"}
	val, err = invalidExpr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error for invalid indicator, got: %v", err)
	}

	if val.Valid {
		t.Errorf("Expected invalid value for invalid indicator")
	}
}

// TestBinaryExprAdd tests addition operator.
func TestBinaryExprAdd(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 10.0},
		Op:    OpAdd,
		Right: &LiteralExpr{Value: 20.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 30.0 {
		t.Errorf("Expected 30.0, got: %f", val.Value)
	}
}

// TestBinaryExprSub tests subtraction operator.
func TestBinaryExprSub(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 50.0},
		Op:    OpSub,
		Right: &LiteralExpr{Value: 25.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 25.0 {
		t.Errorf("Expected 25.0, got: %f", val.Value)
	}
}

// TestBinaryExprMul tests multiplication operator.
func TestBinaryExprMul(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 5.0},
		Op:    OpMul,
		Right: &LiteralExpr{Value: 6.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 30.0 {
		t.Errorf("Expected 30.0, got: %f", val.Value)
	}
}

// TestBinaryExprDiv tests division operator.
func TestBinaryExprDiv(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 100.0},
		Op:    OpDiv,
		Right: &LiteralExpr{Value: 4.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 25.0 {
		t.Errorf("Expected 25.0, got: %f", val.Value)
	}
}

// TestBinaryExprDivByZero tests division by zero.
func TestBinaryExprDivByZero(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 100.0},
		Op:    OpDiv,
		Right: &LiteralExpr{Value: 0.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value for division by zero (returns 0)")
	}

	// According to implementation, division by zero returns 0
	if val.Value != 0.0 {
		t.Errorf("Expected 0.0 for division by zero, got: %f", val.Value)
	}
}

// TestBinaryExprGT tests greater than operator.
func TestBinaryExprGT(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 20.0},
		Op:    OpGT,
		Right: &LiteralExpr{Value: 10.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 1.0 {
		t.Errorf("Expected 1.0 (true), got: %f", val.Value)
	}
}

// TestBinaryExprLT tests less than operator.
func TestBinaryExprLT(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 5.0},
		Op:    OpLT,
		Right: &LiteralExpr{Value: 10.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 1.0 {
		t.Errorf("Expected 1.0 (true), got: %f", val.Value)
	}
}

// TestBinaryExprEQ tests equality operator.
func TestBinaryExprEQ(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 10.0},
		Op:    OpEQ,
		Right: &LiteralExpr{Value: 10.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 1.0 {
		t.Errorf("Expected 1.0 (true), got: %f", val.Value)
	}
}

// TestBinaryExprNE tests not equal operator.
func TestBinaryExprNE(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 10.0},
		Op:    OpNE,
		Right: &LiteralExpr{Value: 20.0},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 1.0 {
		t.Errorf("Expected 1.0 (true), got: %f", val.Value)
	}
}

// TestBinaryExprAnd tests logical AND operator.
func TestBinaryExprAnd(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 1.0}, // true
		Op:    OpAnd,
		Right: &LiteralExpr{Value: 1.0}, // true
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 1.0 {
		t.Errorf("Expected 1.0 (true), got: %f", val.Value)
	}
}

// TestBinaryExprAndFalse tests logical AND with false.
func TestBinaryExprAndFalse(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 1.0}, // true
		Op:    OpAnd,
		Right: &LiteralExpr{Value: 0.0}, // false
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 0.0 {
		t.Errorf("Expected 0.0 (false), got: %f", val.Value)
	}
}

// TestBinaryExprOr tests logical OR operator.
func TestBinaryExprOr(t *testing.T) {
	expr := &BinaryExpr{
		Left:  &LiteralExpr{Value: 0.0}, // false
		Op:    OpOr,
		Right: &LiteralExpr{Value: 1.0}, // true
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 1.0 {
		t.Errorf("Expected 1.0 (true), got: %f", val.Value)
	}
}

// TestUnaryExprNot tests logical NOT operator.
func TestUnaryExprNot(t *testing.T) {
	expr := &UnaryExpr{
		Op:   OpNot,
		Expr: &LiteralExpr{Value: 0.0}, // false
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 1.0 {
		t.Errorf("Expected 1.0 (true), got: %f", val.Value)
	}
}

// TestUnaryExprNotTrue tests logical NOT with true.
func TestUnaryExprNotTrue(t *testing.T) {
	expr := &UnaryExpr{
		Op:   OpNot,
		Expr: &LiteralExpr{Value: 1.0}, // true
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 0.0 {
		t.Errorf("Expected 0.0 (false), got: %f", val.Value)
	}
}

// TestUnaryExprNeg tests negation operator.
func TestUnaryExprNeg(t *testing.T) {
	expr := &UnaryExpr{
		Op:   OpNeg,
		Expr: &LiteralExpr{Value: 42.5},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != -42.5 {
		t.Errorf("Expected -42.5, got: %f", val.Value)
	}
}

// TestComplexExpression tests a complex expression tree.
func TestComplexExpression(t *testing.T) {
	// (10 + 20) * (5 - 2) / 3
	expr := &BinaryExpr{
		Left: &BinaryExpr{
			Left:  &LiteralExpr{Value: 10.0},
			Op:    OpAdd,
			Right: &LiteralExpr{Value: 20.0},
		},
		Op: OpMul,
		Right: &BinaryExpr{
			Left:  &LiteralExpr{Value: 5.0},
			Op:    OpSub,
			Right: &LiteralExpr{Value: 2.0},
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		BarIndex:        0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	// (10 + 20) * (5 - 2) = 30 * 3 = 90
	if val.Value != 90.0 {
		t.Errorf("Expected 90.0, got: %f", val.Value)
	}
}

// TestNestedExpression tests nested comparison and logical operators.
func TestNestedExpression(t *testing.T) {
	// (close > 100) AND (close < 200)
	expr := &BinaryExpr{
		Left: &BinaryExpr{
			Left:  &IdentifierExpr{Name: "close"},
			Op:    OpGT,
			Right: &LiteralExpr{Value: 100.0},
		},
		Op: OpAnd,
		Right: &BinaryExpr{
			Left:  &IdentifierExpr{Name: "close"},
			Op:    OpLT,
			Right: &LiteralExpr{Value: 200.0},
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		CurrentPrices: map[string]float64{
			"close": 150.0,
		},
		BarIndex: 0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 1.0 {
		t.Errorf("Expected 1.0 (true), got: %f", val.Value)
	}
}

// TestNestedExpressionFalse tests nested expression that evaluates to false.
func TestNestedExpressionFalse(t *testing.T) {
	// (close > 100) AND (close < 100)
	expr := &BinaryExpr{
		Left: &BinaryExpr{
			Left:  &IdentifierExpr{Name: "close"},
			Op:    OpGT,
			Right: &LiteralExpr{Value: 100.0},
		},
		Op: OpAnd,
		Right: &BinaryExpr{
			Left:  &IdentifierExpr{Name: "close"},
			Op:    OpLT,
			Right: &LiteralExpr{Value: 100.0},
		},
	}

	ctx := &Context{
		IndicatorValues: make(map[string]indicator.Value),
		CurrentPrices: map[string]float64{
			"close": 150.0,
		},
		BarIndex: 0,
	}

	val, err := expr.Evaluate(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !val.Valid {
		t.Errorf("Expected valid value")
	}

	if val.Value != 0.0 {
		t.Errorf("Expected 0.0 (false), got: %f", val.Value)
	}
}
