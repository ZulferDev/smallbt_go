package optimization

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

// GenerateReport generates a formatted report from optimization results.
func (r *OptimizationReport) GenerateReport() string {
	var sb strings.Builder

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("OPTIMIZATION REPORT\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("Strategy:     %s\n", r.Strategy))
	sb.WriteString(fmt.Sprintf("Symbol:       %s\n", r.Symbol))
	sb.WriteString(fmt.Sprintf("Timeframe:    %s\n", r.Timeframe))
	sb.WriteString(fmt.Sprintf("Period:       %s → %s\n", r.StartTime, r.EndTime))
	sb.WriteString(fmt.Sprintf("Algorithm:    %s\n", r.Algorithm))
	sb.WriteString(fmt.Sprintf("Objective:    %s (%s)\n", r.ObjectiveMetric, r.ObjectiveDirection))
	sb.WriteString(fmt.Sprintf("Total Runs:   %d\n", r.TotalRuns))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\n")

	// Statistics
	sb.WriteString("OPTIMIZATION STATISTICS\n")
	sb.WriteString(fmt.Sprintf("Average %s:   %.4f\n", r.ObjectiveMetric, r.AvgObjectiveValue))
	sb.WriteString(fmt.Sprintf("Std Dev:      %.4f\n", r.StdDevObjective))
	sb.WriteString(fmt.Sprintf("Min %s:     %.4f\n", r.ObjectiveMetric, r.MinObjectiveValue))
	sb.WriteString(fmt.Sprintf("Max %s:     %.4f\n", r.ObjectiveMetric, r.MaxObjectiveValue))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\n")

	// Best result
	sb.WriteString("BEST RESULT\n")
	if r.BestResult != nil {
		sb.WriteString(fmt.Sprintf("Parameters: %s\n", r.BestResult.Parameters.Hash))
		sb.WriteString(fmt.Sprintf("%s: %.4f\n", r.ObjectiveMetric, r.BestResult.ObjectiveValue))
		sb.WriteString("Backtest Metrics:\n")
		if r.BestResult.BacktestResult != nil && r.BestResult.BacktestResult.Metrics != nil {
			m := r.BestResult.BacktestResult.Metrics
			sb.WriteString(fmt.Sprintf("  Total Return:  %.2f%%\n", m.TotalReturn*100))
			sb.WriteString(fmt.Sprintf("  CAGR:          %.2f%%\n", m.CAGR*100))
			sb.WriteString(fmt.Sprintf("  Sharpe:        %.2f\n", m.SharpeRatio))
			sb.WriteString(fmt.Sprintf("  Sortino:       %.2f\n", m.SortinoRatio))
			sb.WriteString(fmt.Sprintf("  Max Drawdown:  %.2f%%\n", m.MaxDrawdown*100))
			sb.WriteString(fmt.Sprintf("  Win Rate:      %.2f%%\n", m.WinRate*100))
			sb.WriteString(fmt.Sprintf("  Profit Factor: %.2f\n", m.ProfitFactor))
			sb.WriteString(fmt.Sprintf("  Expectancy:    %.2fR\n", m.Expectancy))
			sb.WriteString(fmt.Sprintf("  Trades:        %d\n", r.BestResult.BacktestResult.TotalTrades))
		}
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\n")

	// Top 5 results
	sb.WriteString("TOP 5 RESULTS\n")
	sb.WriteString(fmt.Sprintf("%-6s %-30s %s\n", "Rank", "Parameters", fmt.Sprintf("%s", r.ObjectiveMetric)))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for i, result := range r.TopResults {
		if i >= 5 {
			break
		}
		// Truncate hash for display
		hash := result.Parameters.Hash
		if len(hash) > 20 {
			hash = hash[:20] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-6d %-30s %.4f\n", result.Rank, hash, result.ObjectiveValue))
	}
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return sb.String()
}

// SaveJSON saves the optimization report to a JSON file.
func (r *OptimizationReport) SaveJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report to JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write JSON file: %w", err)
	}

	return nil
}

// SaveCSV saves the top results to a CSV file.
func (r *OptimizationReport) SaveCSV(path string) error {
	var sb strings.Builder

	// Header
	sb.WriteString("rank,parameters_hash,objective_value")
	if len(r.Results) > 0 && r.Results[0].BacktestResult != nil {
		sb.WriteString(",total_return,cagr,sharpe,sortino,max_drawdown,win_rate,profit_factor,expectancy,trades")
	}
	sb.WriteString("\n")

	// Data rows
	for _, result := range r.TopResults {
		sb.WriteString(fmt.Sprintf("%d,%s,%.6f", result.Rank, result.Parameters.Hash, result.ObjectiveValue))
		if result.BacktestResult != nil && result.BacktestResult.Metrics != nil {
			m := result.BacktestResult.Metrics
			sb.WriteString(fmt.Sprintf(",%.4f,%.4f,%.4f,%.4f,%.4f,%.4f,%.4f,%.4f,%d",
				m.TotalReturn, m.CAGR, m.SharpeRatio, m.SortinoRatio,
				m.MaxDrawdown, m.WinRate, m.ProfitFactor, m.Expectancy, result.BacktestResult.TotalTrades))
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// TopResultsByMetric returns top N results sorted by a specific metric.
func (r *OptimizationReport) TopResultsByMetric(metric string, n int) []*OptimizationResult {
	if len(r.Results) == 0 {
		return nil
	}

	// Sort by the specified metric
	sorted := make([]*OptimizationResult, len(r.Results))
	copy(sorted, r.Results)

	sort.Slice(sorted, func(i, j int) bool {
		var valI, valJ float64

		switch metric {
		case "sharpe":
			if sorted[i].BacktestResult.Metrics != nil {
				valI = sorted[i].BacktestResult.Metrics.SharpeRatio
			}
			if sorted[j].BacktestResult.Metrics != nil {
				valJ = sorted[j].BacktestResult.Metrics.SharpeRatio
			}
		case "return":
			if sorted[i].BacktestResult.Metrics != nil {
				valI = sorted[i].BacktestResult.Metrics.TotalReturn
			}
			if sorted[j].BacktestResult.Metrics != nil {
				valJ = sorted[j].BacktestResult.Metrics.TotalReturn
			}
		case "cagr":
			if sorted[i].BacktestResult.Metrics != nil {
				valI = sorted[i].BacktestResult.Metrics.CAGR
			}
			if sorted[j].BacktestResult.Metrics != nil {
				valJ = sorted[j].BacktestResult.Metrics.CAGR
			}
		default:
			valI = sorted[i].ObjectiveValue
			valJ = sorted[j].ObjectiveValue
		}

		return valI > valJ
	})

	// Return top N
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// ExportAllResults exports all results as a slice of simplified results.
func (r *OptimizationReport) ExportAllResults() []OptimizationResult {
	results := make([]OptimizationResult, len(r.Results))
	for i, result := range r.Results {
		results[i] = *result
	}
	return results
}

// GetTopParameters returns the top N parameter sets.
func (r *OptimizationReport) GetTopParameters(n int) []ParameterSet {
	if len(r.Results) == 0 {
		return nil
	}

	if n > len(r.Results) {
		n = len(r.Results)
	}

	params := make([]ParameterSet, n)
	for i := 0; i < n; i++ {
		params[i] = r.Results[i].Parameters
	}

	return params
}

// AggregateByParameter aggregates results by a single parameter value.
func (r *OptimizationReport) AggregateByParameter(paramName string) map[float64][]*OptimizationResult {
	aggregated := make(map[float64][]*OptimizationResult)

	for _, result := range r.Results {
		if val, exists := result.Parameters.Values[paramName]; exists {
			aggregated[val] = append(aggregated[val], result)
		}
	}

	return aggregated
}

// CalculateImprovement calculates the improvement from worst to best.
func (r *OptimizationReport) CalculateImprovement() float64 {
	if r.WorstResult == nil || r.BestResult == nil {
		return 0
	}

	if r.WorstResult.ObjectiveValue == 0 {
		return 0
	}

	return (r.BestResult.ObjectiveValue - r.WorstResult.ObjectiveValue) / math.Abs(r.WorstResult.ObjectiveValue)
}

// GetBestParameterValues returns the parameter values from the best result.
func (r *OptimizationReport) GetBestParameterValues() map[string]float64 {
	if r.BestResult == nil {
		return nil
	}
	return r.BestResult.Parameters.Values
}
