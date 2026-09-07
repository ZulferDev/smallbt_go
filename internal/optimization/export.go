package optimization

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
)

// ExportToCSV exports optimization results to CSV format.
func (r *OptimizationReport) ExportToCSV(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Get parameter names from first result
	var paramNames []string
	if len(r.Results) > 0 && r.Results[0] != nil {
		for name := range r.Results[0].Parameters.Values {
			paramNames = append(paramNames, name)
		}
		sort.Strings(paramNames) // Consistent ordering
	}

	// Build header
	header := []string{"Rank", "ObjectiveValue"}
	header = append(header, paramNames...)
	header = append(header,
		"TotalReturn_%",
		"CAGR_%",
		"SharpeRatio",
		"SortinoRatio",
		"MaxDrawdown_%",
		"WinRate_%",
		"ProfitFactor",
		"Expectancy",
		"TotalTrades",
		"WinningTrades",
		"LosingTrades",
		"AvgWin",
		"AvgLoss",
	)

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write results
	for _, result := range r.Results {
		if result == nil {
			continue
		}

		record := []string{
			fmt.Sprintf("%d", result.Rank),
			fmt.Sprintf("%.6f", result.ObjectiveValue),
		}

		// Add parameter values in consistent order
		for _, name := range paramNames {
			value := result.Parameters.Values[name]
			record = append(record, fmt.Sprintf("%.6f", value))
		}

		// Add metrics
		if result.BacktestResult != nil && result.BacktestResult.Metrics != nil {
			m := result.BacktestResult.Metrics
			record = append(record,
				fmt.Sprintf("%.4f", m.TotalReturn*100),
				fmt.Sprintf("%.4f", m.CAGR*100),
				fmt.Sprintf("%.4f", m.SharpeRatio),
				fmt.Sprintf("%.4f", m.SortinoRatio),
				fmt.Sprintf("%.4f", m.MaxDrawdown*100),
				fmt.Sprintf("%.4f", m.WinRate*100),
				fmt.Sprintf("%.4f", m.ProfitFactor),
				fmt.Sprintf("%.6f", m.Expectancy),
				fmt.Sprintf("%d", result.BacktestResult.TotalTrades),
				fmt.Sprintf("%d", m.WinningTrades),
				fmt.Sprintf("%d", m.LosingTrades),
				fmt.Sprintf("%.2f", m.AvgWin),
				fmt.Sprintf("%.2f", m.AvgLoss),
			)
		} else {
			// Fill with zeros if metrics unavailable
			for i := 0; i < 13; i++ {
				record = append(record, "0")
			}
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	return nil
}

// ExportTopNToCSV exports the top N results to CSV.
func (r *OptimizationReport) ExportTopNToCSV(filepath string, n int) error {
	if n <= 0 || n > len(r.Results) {
		n = len(r.Results)
	}

	// Create temporary report with top N
	topReport := &OptimizationReport{
		Strategy:           r.Strategy,
		Symbol:             r.Symbol,
		Timeframe:          r.Timeframe,
		StartTime:          r.StartTime,
		EndTime:            r.EndTime,
		TotalRuns:          n,
		ObjectiveMetric:    r.ObjectiveMetric,
		ObjectiveDirection: r.ObjectiveDirection,
		Algorithm:          r.Algorithm,
		Results:            r.Results[:n],
		BestResult:         r.BestResult,
	}

	return topReport.ExportToCSV(filepath)
}

// ParameterSensitivity represents sensitivity analysis for a parameter.
type ParameterSensitivity struct {
	ParameterName string
	MinValue      float64
	MaxValue      float64
	Range         float64
	BestValue     float64
	WorstValue    float64
	AvgObjective  float64
	StdDevObj     float64
	Correlation   float64
	Sensitivity   string // "Low", "Medium", "High"
}

// AnalyzeParameterSensitivity analyzes how each parameter affects the objective.
func (r *OptimizationReport) AnalyzeParameterSensitivity() []ParameterSensitivity {
	if len(r.Results) == 0 {
		return nil
	}

	// Get all parameter names
	paramNames := make([]string, 0)
	for name := range r.Results[0].Parameters.Values {
		paramNames = append(paramNames, name)
	}
	sort.Strings(paramNames)

	sensitivities := make([]ParameterSensitivity, 0, len(paramNames))

	for _, paramName := range paramNames {
		sens := ParameterSensitivity{
			ParameterName: paramName,
			MinValue:      math.MaxFloat64,
			MaxValue:      -math.MaxFloat64,
		}

		// Collect values and objectives
		var values []float64
		var objectives []float64

		for _, result := range r.Results {
			if result == nil {
				continue
			}

			value := result.Parameters.Values[paramName]
			objective := result.ObjectiveValue

			values = append(values, value)
			objectives = append(objectives, objective)

			if value < sens.MinValue {
				sens.MinValue = value
			}
			if value > sens.MaxValue {
				sens.MaxValue = value
			}

			sens.AvgObjective += objective
		}

		if len(values) == 0 {
			continue
		}

		// Calculate statistics
		sens.Range = sens.MaxValue - sens.MinValue
		sens.AvgObjective /= float64(len(objectives))

		// Calculate standard deviation of objective
		var sumSquaredDiff float64
		for _, obj := range objectives {
			diff := obj - sens.AvgObjective
			sumSquaredDiff += diff * diff
		}
		sens.StdDevObj = math.Sqrt(sumSquaredDiff / float64(len(objectives)))

		// Find best and worst parameter values
		bestObj := -math.MaxFloat64
		worstObj := math.MaxFloat64
		for i, obj := range objectives {
			if obj > bestObj {
				bestObj = obj
				sens.BestValue = values[i]
			}
			if obj < worstObj {
				worstObj = obj
				sens.WorstValue = values[i]
			}
		}

		// Calculate correlation
		sens.Correlation = calculateCorrelation(values, objectives)

		// Determine sensitivity level
		// High std dev or strong correlation = high sensitivity
		if sens.StdDevObj > 0.5 || math.Abs(sens.Correlation) > 0.7 {
			sens.Sensitivity = "High"
		} else if sens.StdDevObj > 0.2 || math.Abs(sens.Correlation) > 0.4 {
			sens.Sensitivity = "Medium"
		} else {
			sens.Sensitivity = "Low"
		}

		sensitivities = append(sensitivities, sens)
	}

	return sensitivities
}

// ExportSensitivityAnalysisToCSV exports parameter sensitivity analysis.
func (r *OptimizationReport) ExportSensitivityAnalysisToCSV(filepath string) error {
	sensitivities := r.AnalyzeParameterSensitivity()
	if len(sensitivities) == 0 {
		return fmt.Errorf("no sensitivity data available")
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Parameter",
		"MinValue",
		"MaxValue",
		"Range",
		"BestValue",
		"WorstValue",
		"AvgObjective",
		"StdDevObjective",
		"Correlation",
		"Sensitivity",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write sensitivity data
	for _, sens := range sensitivities {
		record := []string{
			sens.ParameterName,
			fmt.Sprintf("%.4f", sens.MinValue),
			fmt.Sprintf("%.4f", sens.MaxValue),
			fmt.Sprintf("%.4f", sens.Range),
			fmt.Sprintf("%.4f", sens.BestValue),
			fmt.Sprintf("%.4f", sens.WorstValue),
			fmt.Sprintf("%.6f", sens.AvgObjective),
			fmt.Sprintf("%.6f", sens.StdDevObj),
			fmt.Sprintf("%.4f", sens.Correlation),
			sens.Sensitivity,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	return nil
}

// calculateCorrelation calculates Pearson correlation coefficient.
func calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	n := float64(len(x))

	// Calculate means
	var sumX, sumY float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / n
	meanY := sumY / n

	// Calculate correlation
	var numerator, denomX, denomY float64
	for i := range x {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		denomX += dx * dx
		denomY += dy * dy
	}

	if denomX == 0 || denomY == 0 {
		return 0
	}

	return numerator / math.Sqrt(denomX*denomY)
}
