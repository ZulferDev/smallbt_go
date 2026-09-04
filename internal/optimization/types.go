package optimization

import (
	"github.com/1jehuang/backtest/internal/backtest"
	"github.com/1jehuang/backtest/internal/strategy/ast"
)

// ParameterRange defines the range for a parameter to optimize.
type ParameterRange struct {
	Name  string        `json:"name"`
	Start float64       `json:"start"`
	End   float64       `json:"end"`
	Step  float64       `json:"step"`
	Type  string        `json:"type"` // "int" or "float"
	Path  string        `json:"path"` // e.g., "indicators.ema_fast.period"
}

// OptimizationConfig defines the optimization parameters and objectives.
type OptimizationConfig struct {
	// The base backtest configuration
	BacktestConfig backtest.BacktestConfig

	// Strategy path
	StrategyPath string

	// Parameters to optimize
	Parameters []ParameterRange

	// Optimization objective
	Objective ObjectiveConfig

	// Algorithm: "grid", "random" (future: genetic, bayesian)
	Algorithm string
}

// ObjectiveConfig defines the optimization objective.
type ObjectiveConfig struct {
	// Type: "sharpe", "sortino", "return", "profit_factor", "calmar"
	Type string

	// Direction: "maximize" or "minimize"
	Direction string
}

// ParameterSet represents a specific combination of parameter values.
type ParameterSet struct {
	Values map[string]float64 // parameter name -> value
	Hash   string             // unique hash of this parameter set
}

// OptimizationResult represents the result of a single backtest with specific parameters.
type OptimizationResult struct {
	Parameters     ParameterSet
	BacktestResult *backtest.BacktestResult
	ObjectiveValue float64 // the metric value being optimized
	Rank           int     // ranking among all results
}

// OptimizationReport contains aggregated results from optimization run.
type OptimizationReport struct {
	Strategy          string
	Symbol            string
	Timeframe         string
	StartTime         string
	EndTime           string
	TotalRuns         int
	ObjectiveMetric   string
	ObjectiveDirection string
	Algorithm         string

	// All results sorted by objective metric
	Results []*OptimizationResult

	// Best result
	BestResult *OptimizationResult

	// Worst result
	WorstResult *OptimizationResult

	// Statistics
	AvgObjectiveValue float64
	StdDevObjective   float64
	MinObjectiveValue float64
	MaxObjectiveValue float64

	// Top N results
	TopResults []*OptimizationResult
}

// StrategyModifier creates a modified strategy with new parameter values.
type StrategyModifier interface {
	ModifyStrategy(strategy *ast.Strategy, parameterSet ParameterSet) (*ast.Strategy, error)
}
