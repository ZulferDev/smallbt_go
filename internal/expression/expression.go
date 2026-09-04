package expression

import (
	"github.com/1jehuang/backtest/internal/indicator"
)

// Expression represents a node in the expression AST.
// Expressions are stateless and deterministic.
type Expression interface {
	// Evaluate computes the expression value given the evaluation context.
	Evaluate(ctx *Context) (Value, error)
}

// Context provides the evaluation context for expression evaluation.
type Context struct {
	// Current candle being evaluated
	// (will be populated from indicator.Context in integration)

	// IndicatorValues provides access to computed indicator values
	IndicatorValues map[string]indicator.Value

	// BarIndex is the current bar index (0-based from start of data)
	BarIndex int

	// HistoricalIndicatorValues provides access to historical indicator values
	// Key is indicator name, value is slice of values indexed by bar offset
	// [0] = current bar, [1] = previous bar, etc.
	HistoricalIndicatorValues map[string][]indicator.Value

	// CurrentPrices provides current candle prices
	CurrentPrices map[string]float64 // "open", "high", "low", "close", "volume"

	// HistoricalPrices provides historical candle prices
	// [0] = current bar, [1] = previous bar, etc.
	HistoricalPrices []map[string]float64
}

// Value represents an expression's computed value.
type Value struct {
	// Valid indicates whether the value is valid
	Valid bool

	// Value is the computed value
	Value float64
}

// NewValue creates a valid expression value.
func NewValue(v float64) Value {
	return Value{Valid: true, Value: v}
}

// InvalidValue represents an invalid/uncomputed expression value.
var InvalidValue = Value{Valid: false}

// NodeType represents the type of expression node.
type NodeType string

const (
	NodeLiteral    NodeType = "literal"
	NodeIdentifier NodeType = "identifier"
	NodeBinaryOp   NodeType = "binary_op"
	NodeUnaryOp    NodeType = "unary_op"
	NodeFunction   NodeType = "function"
)

// BinaryOperator represents a binary operator.
type BinaryOperator string

const (
	// Arithmetic
	OpAdd BinaryOperator = "+"
	OpSub BinaryOperator = "-"
	OpMul BinaryOperator = "*"
	OpDiv BinaryOperator = "/"
	OpMod BinaryOperator = "%"

	// Comparison
	OpGT BinaryOperator = ">"
	OpLT BinaryOperator = "<"
	OpGE BinaryOperator = ">="
	OpLE BinaryOperator = "<="
	OpEQ BinaryOperator = "=="
	OpNE BinaryOperator = "!="

	// Logical
	OpAnd BinaryOperator = "and"
	OpOr  BinaryOperator = "or"
)

// UnaryOperator represents a unary operator.
type UnaryOperator string

const (
	OpNot UnaryOperator = "not"
	OpNeg UnaryOperator = "-"
)

// LiteralExpr represents a numeric literal.
type LiteralExpr struct {
	Value float64
}

func (e *LiteralExpr) Evaluate(ctx *Context) (Value, error) {
	return NewValue(e.Value), nil
}

// IdentifierExpr represents a reference to an indicator or price.
type IdentifierExpr struct {
	Name string // e.g., "ema20", "close", "volume"
}

func (e *IdentifierExpr) Evaluate(ctx *Context) (Value, error) {
	// First check if it's a price field
	if price, ok := ctx.CurrentPrices[e.Name]; ok {
		return NewValue(price), nil
	}

	// Then check if it's an indicator
	if val, ok := ctx.IndicatorValues[e.Name]; ok {
		return Value{Valid: val.Valid, Value: val.Value}, nil
	}

	return InvalidValue, nil
}

// BinaryExpr represents a binary operation.
type BinaryExpr struct {
	Left  Expression
	Op    BinaryOperator
	Right Expression
}

func (e *BinaryExpr) Evaluate(ctx *Context) (Value, error) {
	left, err := e.Left.Evaluate(ctx)
	if err != nil {
		return InvalidValue, err
	}
	if !left.Valid {
		return InvalidValue, nil
	}

	right, err := e.Right.Evaluate(ctx)
	if err != nil {
		return InvalidValue, err
	}
	if !right.Valid {
		return InvalidValue, nil
	}

	result, err := applyBinaryOperator(e.Op, left.Value, right.Value)
	if err != nil {
		return InvalidValue, err
	}

	return NewValue(result), nil
}

// UnaryExpr represents a unary operation.
type UnaryExpr struct {
	Op   UnaryOperator
	Expr Expression
}

func (e *UnaryExpr) Evaluate(ctx *Context) (Value, error) {
	val, err := e.Expr.Evaluate(ctx)
	if err != nil {
		return InvalidValue, err
	}
	if !val.Valid {
		return InvalidValue, nil
	}

	result, err := applyUnaryOperator(e.Op, val.Value)
	if err != nil {
		return InvalidValue, err
	}

	return NewValue(result), nil
}

// applyBinaryOperator applies a binary operator to two values.
func applyBinaryOperator(op BinaryOperator, left, right float64) (float64, error) {
	switch op {
	case OpAdd:
		return left + right, nil
	case OpSub:
		return left - right, nil
	case OpMul:
		return left * right, nil
	case OpDiv:
		if right == 0 {
			return 0, nil // Division by zero returns 0 (could also return error)
		}
		return left / right, nil
	case OpMod:
		if right == 0 {
			return 0, nil
		}
		return float64(int(left) % int(right)), nil

	// Comparison operators return 1 (true) or 0 (false)
	case OpGT:
		if left > right {
			return 1, nil
		}
		return 0, nil
	case OpLT:
		if left < right {
			return 1, nil
		}
		return 0, nil
	case OpGE:
		if left >= right {
			return 1, nil
		}
		return 0, nil
	case OpLE:
		if left <= right {
			return 1, nil
		}
		return 0, nil
	case OpEQ:
		if left == right {
			return 1, nil
		}
		return 0, nil
	case OpNE:
		if left != right {
			return 1, nil
		}
		return 0, nil

	// Logical operators
	case OpAnd:
		if left != 0 && right != 0 {
			return 1, nil
		}
		return 0, nil
	case OpOr:
		if left != 0 || right != 0 {
			return 1, nil
		}
		return 0, nil

	default:
		return 0, nil
	}
}

// applyUnaryOperator applies a unary operator to a value.
func applyUnaryOperator(op UnaryOperator, val float64) (float64, error) {
	switch op {
	case OpNot:
		if val == 0 {
			return 1, nil
		}
		return 0, nil
	case OpNeg:
		return -val, nil
	default:
		return val, nil
	}
}
