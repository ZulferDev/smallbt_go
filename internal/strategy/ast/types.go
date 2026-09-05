package ast

import (
	"github.com/ZulferDev/smallbt_go/internal/expression"
)

// Strategy represents the complete strategy AST.
// This is the internal representation after parsing YAML.
// YAML is just an interface - the engine works with this AST.
type Strategy struct {
	Name        string
	Version     string
	Description string

	// Data configuration
	Data DataConfig

	// Indicators keyed by name
	Indicators map[string]IndicatorDef

	// Entry rules
	Entry EntryRules

	// Exit rules
	Exit ExitRules

	// Risk management
	Risk RiskConfig

	// Execution configuration
	Execution ExecutionConfig

	// State definitions (for stateful strategies)
	State map[string]StateDef
}

// DataConfig defines what data the strategy needs.
type DataConfig struct {
	Symbol    string
	Timeframe string

	// Future: multi-timeframe support
	// Additional map[string]TimeframeConfig
}

// IndicatorDef defines an indicator configuration.
type IndicatorDef struct {
	Name string
	Type string

	// Parameters for the indicator
	Params map[string]interface{}

	// Source field (e.g., "close", "volume")
	Source string

	// Period if applicable
	Period int

	// Timeframe (for multi-timeframe indicators)
	Timeframe string

	// For composite indicators
	Left  string // reference to another indicator
	Right string // reference to another indicator
	Op    string // operation: add, subtract, multiply, divide
}

// EntryRules defines entry conditions.
type EntryRules struct {
	Long  *Condition
	Short *Condition
}

// ExitRules defines exit conditions.
type ExitRules struct {
	Long  *Condition
	Short *Condition
}

// Condition represents a logical condition.
// It can be a simple condition or a composite (all/any).
type Condition struct {
	// Type: "all", "any", "not", "expr", "func"
	Type string

	// For composite conditions
	Conditions []*Condition

	// For expression-based conditions
	Expr expression.Expression

	// For function-based conditions
	Function string
	Args     []interface{}
}

// ExecutionConfig defines execution behavior and order types.
type ExecutionConfig struct {
	// Entry order type: "market", "limit", "stop", "stop_limit"
	EntryOrderType string

	// Exit order type for stop loss and take profit
	ExitOrderType string

	// Slippage configuration
	SlippageType  string  // "percentage", "fixed", "none"
	SlippageValue float64 // percentage (0.0005 = 0.05%) or fixed amount

	// Fees configuration
	FeeMaker float64 // maker fee percentage
	FeeTaker float64 // taker fee percentage

	// Spread configuration (average bid-ask spread)
	Spread float64

	// Intrabar policy for order execution
	IntrabarPolicy string // "conservative", "optimistic", "nearest"
}

// RiskConfig defines risk management rules.
type RiskConfig struct {
	PositionSize PositionSizeConfig
	StopLoss     *StopLossConfig
	TakeProfit   *TakeProfitConfig
	TrailingStop *TrailingStopConfig

	// Portfolio-level risk
	MaxPositions      int
	MaxPortfolioRisk  float64
	MaxDailyLoss      float64
	MaxDrawdown       float64
	MaxConcurrentRisk float64
}

// PositionSizeConfig defines position sizing method.
type PositionSizeConfig struct {
	Type  string  // "fixed", "percent_equity", "risk_percent"
	Value float64 // the amount or percentage

	// For risk-based sizing
	RiskPercent float64
}

// StopLossConfig defines stop loss configuration.
type StopLossConfig struct {
	Type string // "fixed", "percentage", "atr", "expression"

	// For fixed
	Price float64

	// For percentage
	Percentage float64

	// For ATR-based
	Indicator  string
	Multiplier float64

	// For expression-based
	Expr expression.Expression
}

// TakeProfitConfig defines take profit configuration.
type TakeProfitConfig struct {
	Type string // "fixed", "percentage", "risk_reward", "expression"

	// For fixed
	Price float64

	// For percentage
	Percentage float64

	// For risk/reward ratio
	Ratio float64

	// For expression-based
	Expr expression.Expression

	// Future: multiple targets
	// Targets []TargetConfig
}

// TrailingStopConfig defines trailing stop configuration.
type TrailingStopConfig struct {
	Type string // "percentage", "atr", "highest_lowest", "expression"

	// For percentage
	Percentage float64

	// For ATR-based
	Indicator  string
	Multiplier float64

	// For expression-based
	Expr expression.Expression
}

// StateDef defines a state variable for stateful strategies.
type StateDef struct {
	Name    string
	Type    string      // "bool", "float", "int", "string"
	Default interface{} // default value
}

// Rule represents a state transition rule (future feature).
// When condition is met, set state values or trigger actions.
type Rule struct {
	When   *Condition
	Set    map[string]interface{} // state updates
	Action string                 // "enter", "exit", etc.
}
