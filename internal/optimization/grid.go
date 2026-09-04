package optimization

import (
	"fmt"
	"math"
	"sync"

	"github.com/1jehuang/backtest/internal/backtest"
)

// GridSearch implements grid search optimization algorithm.
type GridSearch struct {
	config OptimizationConfig
}

// NewGridSearch creates a new grid search optimizer.
func NewGridSearch(config OptimizationConfig) *GridSearch {
	return &GridSearch{
		config: config,
	}
}

// GenerateParameterSets generates all parameter combinations for grid search.
func (g *GridSearch) GenerateParameterSets() ([]ParameterSet, error) {
	if len(g.config.Parameters) == 0 {
		return nil, fmt.Errorf("no parameters to optimize")
	}

	// Validate parameter ranges
	for _, p := range g.config.Parameters {
		if p.Step <= 0 {
			return nil, fmt.Errorf("parameter %s: step must be positive", p.Name)
		}
		if p.Start > p.End {
			return nil, fmt.Errorf("parameter %s: start (%v) must be <= end (%v)", p.Name, p.Start, p.End)
		}
	}

	// Generate values for each parameter
	paramValues := make([][]float64, len(g.config.Parameters))
	for i, p := range g.config.Parameters {
		values := generateValues(p)
		paramValues[i] = values
	}

	// Generate all combinations using cartesian product
	sets := cartesianProduct(paramValues, g.config.Parameters)

	return sets, nil
}

// generateValues generates all values for a single parameter.
func generateValues(p ParameterRange) []float64 {
	var values []float64

	// Number of steps
	numSteps := int(math.Ceil((p.End - p.Start) / p.Step))

	for i := 0; i <= numSteps; i++ {
		val := p.Start + float64(i)*p.Step
		if val > p.End {
			val = p.End
		}

		// Round if integer type
		if p.Type == "int" {
			val = math.Round(val)
		}

		values = append(values, val)
	}

	// Remove duplicates (due to rounding)
	unique := make([]float64, 0, len(values))
	seen := make(map[float64]bool)
	for _, v := range values {
		rounded := roundToDecimals(v, 6)
		if !seen[rounded] {
			seen[rounded] = true
			unique = append(unique, v)
		}
	}

	return unique
}

// cartesianProduct generates all combinations of parameter values.
func cartesianProduct(paramValues [][]float64, params []ParameterRange) []ParameterSet {
	if len(paramValues) == 0 {
		return nil
	}

	// Start with first parameter
	result := make([]ParameterSet, 0)
	for _, v := range paramValues[0] {
		result = append(result, ParameterSet{
			Values: map[string]float64{
				params[0].Name: v,
			},
		})
	}

	// Add remaining parameters
	for i := 1; i < len(paramValues); i++ {
		newResult := make([]ParameterSet, 0)
		for _, existing := range result {
			for _, v := range paramValues[i] {
				// Copy existing values
				newValues := make(map[string]float64)
				for k, val := range existing.Values {
					newValues[k] = val
				}
				newValues[params[i].Name] = v

				newResult = append(newResult, ParameterSet{
					Values: newValues,
				})
			}
		}
		result = newResult
	}

	// Generate hashes
	for i := range result {
		result[i].Hash = generateHash(result[i].Values)
	}

	return result
}

// generateHash creates a unique hash for a parameter set.
func generateHash(values map[string]float64) string {
	hash := ""
	for name, val := range values {
		hash += fmt.Sprintf("%s:%.6f;", name, val)
	}
	return hash
}

// roundToDecimals rounds a float to specified decimal places.
func roundToDecimals(val float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(val*multiplier) / multiplier
}

// EstimateTotalCombinations estimates total number of parameter combinations.
func (g *GridSearch) EstimateTotalCombinations() int {
	if len(g.config.Parameters) == 0 {
		return 0
	}

	total := 1
	for _, p := range g.config.Parameters {
		numValues := int(math.Ceil((p.End-p.Start)/p.Step)) + 1
		total *= numValues
	}

	return total
}

// Run executes grid search optimization.
// The evaluator function is called for each parameter set.
func (g *GridSearch) Run(evaluator func(ParameterSet) (*backtest.BacktestResult, error), parallel int) (*OptimizationReport, error) {
	parameterSets, err := g.GenerateParameterSets()
	if err != nil {
		return nil, fmt.Errorf("generate parameter sets: %w", err)
	}

	if len(parameterSets) == 0 {
		return nil, fmt.Errorf("no parameter combinations generated")
	}

	report := &OptimizationReport{
		Strategy:           g.config.StrategyPath,
		Symbol:             string(g.config.BacktestConfig.Symbol),
		Timeframe:          string(g.config.BacktestConfig.Timeframe),
		StartTime:          g.config.BacktestConfig.StartTime.Format("2006-01-02"),
		EndTime:            g.config.BacktestConfig.EndTime.Format("2006-01-02"),
		TotalRuns:          len(parameterSets),
		ObjectiveMetric:    g.config.Objective.Type,
		ObjectiveDirection: g.config.Objective.Direction,
		Algorithm:          "grid",
		Results:            make([]*OptimizationResult, 0, len(parameterSets)),
	}

	// Run evaluations
	if parallel > 1 {
		// Parallel execution
		report.Results = g.runParallel(parameterSets, evaluator, parallel)
	} else {
		// Sequential execution
		report.Results = g.runSequential(parameterSets, evaluator)
	}

	// Calculate objective values
	for _, result := range report.Results {
		result.ObjectiveValue = g.getMetricValue(result.BacktestResult)
	}

	// Sort by objective value
	report.sortResults()

	// Calculate statistics
	report.calculateStatistics()

	return report, nil
}

// runSequential runs evaluations sequentially.
func (g *GridSearch) runSequential(parameterSets []ParameterSet, evaluator func(ParameterSet) (*backtest.BacktestResult, error)) []*OptimizationResult {
	results := make([]*OptimizationResult, 0, len(parameterSets))

	for i, ps := range parameterSets {
		backtestResult, err := evaluator(ps)
		if err != nil {
			// Log error but continue
			fmt.Printf("Warning: parameter set %d failed: %v\n", i, err)
			continue
		}

		results = append(results, &OptimizationResult{
			Parameters:     ps,
			BacktestResult: backtestResult,
		})
	}

	return results
}

// runParallel runs evaluations in parallel with limited concurrency.
func (g *GridSearch) runParallel(parameterSets []ParameterSet, evaluator func(ParameterSet) (*backtest.BacktestResult, error), maxWorkers int) []*OptimizationResult {
	results := make([]*OptimizationResult, len(parameterSets))

	// Create semaphore for limiting concurrency
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, ps := range parameterSets {
		wg.Add(1)
		go func(idx int, pSet ParameterSet) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			backtestResult, err := evaluator(pSet)
			if err != nil {
				fmt.Printf("Warning: parameter set %d failed: %v\n", idx, err)
				return
			}

			mu.Lock()
			results[idx] = &OptimizationResult{
				Parameters:     pSet,
				BacktestResult: backtestResult,
			}
			mu.Unlock()
		}(i, ps)
	}

	wg.Wait()

	// Filter nil results
	filtered := make([]*OptimizationResult, 0, len(results))
	for _, r := range results {
		if r != nil {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

// getMetricValue extracts the objective metric from backtest result.
func (g *GridSearch) getMetricValue(result *backtest.BacktestResult) float64 {
	if result == nil || result.Metrics == nil {
		return 0
	}

	switch g.config.Objective.Type {
	case "sharpe":
		return result.Metrics.SharpeRatio
	case "sortino":
		return result.Metrics.SortinoRatio
	case "return", "total_return":
		return result.Metrics.TotalReturn
	case "profit_factor":
		return result.Metrics.ProfitFactor
	case "calmar":
		if result.Metrics.MaxDrawdown != 0 {
			return result.Metrics.CAGR / result.Metrics.MaxDrawdown
		}
		return 0
	case "win_rate":
		return result.Metrics.WinRate
	case "expectancy":
		return result.Metrics.Expectancy
	case "cagr":
		return result.Metrics.CAGR
	default:
		return result.Metrics.SharpeRatio
	}
}

// sortResults sorts results by objective value.
func (r *OptimizationReport) sortResults() {
	if len(r.Results) == 0 {
		return
	}

	// Sort by objective value
	// For "maximize", higher is better (descending)
	// For "minimize", lower is better (ascending)
	for i := 0; i < len(r.Results); i++ {
		for j := i + 1; j < len(r.Results); j++ {
			shouldSwap := false
			if r.ObjectiveDirection == "minimize" {
				shouldSwap = r.Results[i].ObjectiveValue > r.Results[j].ObjectiveValue
			} else {
				shouldSwap = r.Results[i].ObjectiveValue < r.Results[j].ObjectiveValue
			}
			if shouldSwap {
				r.Results[i], r.Results[j] = r.Results[j], r.Results[i]
			}
		}
	}

	// Assign ranks
	for i := range r.Results {
		r.Results[i].Rank = i + 1
	}

	// Set best and worst
	r.BestResult = r.Results[0]
	r.WorstResult = r.Results[len(r.Results)-1]

	// Set top 10 results
	topN := 10
	if len(r.Results) < topN {
		topN = len(r.Results)
	}
	r.TopResults = r.Results[:topN]
}

// calculateStatistics calculates aggregate statistics.
func (r *OptimizationReport) calculateStatistics() {
	if len(r.Results) == 0 {
		return
	}

	// Calculate mean
	var sum float64
	for _, result := range r.Results {
		sum += result.ObjectiveValue
	}
	r.AvgObjectiveValue = sum / float64(len(r.Results))

	// Calculate std dev
	var sumSqDiff float64
	for _, result := range r.Results {
		diff := result.ObjectiveValue - r.AvgObjectiveValue
		sumSqDiff += diff * diff
	}
	r.StdDevObjective = math.Sqrt(sumSqDiff / float64(len(r.Results)))

	// Min and max
	r.MinObjectiveValue = r.WorstResult.ObjectiveValue
	r.MaxObjectiveValue = r.BestResult.ObjectiveValue
}
