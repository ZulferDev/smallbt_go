package montecarlo

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ExportToJSON exports Monte Carlo results to JSON file (backward compatibility)
func (mcr *MCResult) ExportToJSON(filepath string) error {
	data, err := json.MarshalIndent(mcr, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// ExportToCSV exports basic statistics to CSV (backward compatibility)
func (mcr *MCResult) ExportToCSV(filepath string) error {
	return mcr.ExportStatisticsToCSV(filepath)
}

// ExportToText generates a text report (backward compatibility)
func (mcr *MCResult) ExportToText() string {
	if mcr == nil {
		return "No Monte Carlo results"
	}

	var sb strings.Builder

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("MONTE CARLO SIMULATION RESULTS\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("Simulations: %d\n", len(mcr.Simulations)))
	sb.WriteString(fmt.Sprintf("Analysis Type: %s\n", getAnalysisTypeName(mcr.Config.Type)))
	sb.WriteString(fmt.Sprintf("Seed: %d\n\n", mcr.Config.Seed))

	stats := mcr.Statistics

	sb.WriteString("Return Distribution:\n")
	sb.WriteString(fmt.Sprintf("  Mean: %.2f%%\n", stats.MeanReturn*100))
	sb.WriteString(fmt.Sprintf("  Median: %.2f%%\n", stats.MedianReturn*100))
	sb.WriteString(fmt.Sprintf("  Std Dev: %.2f%%\n", stats.StdDevReturn*100))
	sb.WriteString(fmt.Sprintf("  Min: %.2f%%\n", stats.MinReturn*100))
	sb.WriteString(fmt.Sprintf("  Max: %.2f%%\n\n", stats.MaxReturn*100))

	sb.WriteString("Drawdown Distribution:\n")
	sb.WriteString(fmt.Sprintf("  Mean Max DD: %.2f%%\n", stats.MeanMaxDrawdown*100))
	sb.WriteString(fmt.Sprintf("  95th Percentile: %.2f%%\n\n", stats.P95MaxDrawdown*100))

	sb.WriteString("Risk Metrics:\n")
	sb.WriteString(fmt.Sprintf("  Probability of Ruin: %.2f%%\n", stats.ProbabilityOfRuin*100))
	sb.WriteString(fmt.Sprintf("  Negative Returns: %d (%.1f%%)\n", 
		stats.NegativeReturnCount, stats.NegativeReturnRatio*100))

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return sb.String()
}

// ExportStatisticsToCSV exports aggregated Monte Carlo statistics to CSV.
func (mcr *MCResult) ExportStatisticsToCSV(filepath string) error {
	if mcr == nil {
		return fmt.Errorf("no Monte Carlo results to export")
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Metric", "Value"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	stats := mcr.Statistics

	// Write statistics
	metrics := [][]string{
		{"Simulations", fmt.Sprintf("%d", len(mcr.Simulations))},
		{"AnalysisType", getAnalysisTypeName(mcr.Config.Type)},
		{"Seed", fmt.Sprintf("%d", mcr.Config.Seed)},
		{"", ""},
		{"=== Return Distribution ===", ""},
		{"MeanReturn_%", fmt.Sprintf("%.2f", stats.MeanReturn*100)},
		{"MedianReturn_%", fmt.Sprintf("%.2f", stats.MedianReturn*100)},
		{"StdDevReturn_%", fmt.Sprintf("%.2f", stats.StdDevReturn*100)},
		{"MinReturn_%", fmt.Sprintf("%.2f", stats.MinReturn*100)},
		{"MaxReturn_%", fmt.Sprintf("%.2f", stats.MaxReturn*100)},
		{"P05Return_%", fmt.Sprintf("%.2f", stats.P05Return*100)},
		{"P95Return_%", fmt.Sprintf("%.2f", stats.P95Return*100)},
		{"", ""},
		{"=== Drawdown Distribution ===", ""},
		{"MeanMaxDrawdown_%", fmt.Sprintf("%.2f", stats.MeanMaxDrawdown*100)},
		{"MedianMaxDrawdown_%", fmt.Sprintf("%.2f", stats.MedianMaxDrawdown*100)},
		{"StdDevMaxDrawdown_%", fmt.Sprintf("%.2f", stats.StdDevMaxDrawdown*100)},
		{"MinMaxDrawdown_%", fmt.Sprintf("%.2f", stats.MinMaxDrawdown*100)},
		{"MaxMaxDrawdown_%", fmt.Sprintf("%.2f", stats.MaxMaxDrawdown*100)},
		{"P95MaxDrawdown_%", fmt.Sprintf("%.2f", stats.P95MaxDrawdown*100)},
		{"", ""},
		{"=== Win Rate Distribution ===", ""},
		{"MeanWinRate_%", fmt.Sprintf("%.2f", stats.MeanWinRate*100)},
		{"StdDevWinRate_%", fmt.Sprintf("%.2f", stats.StdDevWinRate*100)},
		{"MinWinRate_%", fmt.Sprintf("%.2f", stats.MinWinRate*100)},
		{"MaxWinRate_%", fmt.Sprintf("%.2f", stats.MaxWinRate*100)},
		{"", ""},
		{"=== Sharpe Ratio Distribution ===", ""},
		{"MeanSharpe", fmt.Sprintf("%.4f", stats.MeanSharpe)},
		{"StdDevSharpe", fmt.Sprintf("%.4f", stats.StdDevSharpe)},
		{"MinSharpe", fmt.Sprintf("%.4f", stats.MinSharpe)},
		{"MaxSharpe", fmt.Sprintf("%.4f", stats.MaxSharpe)},
		{"", ""},
		{"=== Risk Metrics ===", ""},
		{"ProbabilityOfRuin_%", fmt.Sprintf("%.2f", stats.ProbabilityOfRuin*100)},
		{"NegativeReturnCount", fmt.Sprintf("%d", stats.NegativeReturnCount)},
		{"NegativeReturnRatio_%", fmt.Sprintf("%.2f", stats.NegativeReturnRatio*100)},
	}

	for _, metric := range metrics {
		if err := writer.Write(metric); err != nil {
			return fmt.Errorf("write metric: %w", err)
		}
	}

	return nil
}

// ExportPercentilesToCSV exports percentile analysis to CSV.
func (mcr *MCResult) ExportPercentilesToCSV(filepath string) error {
	if mcr == nil || len(mcr.ConfidenceIntervals) == 0 {
		return fmt.Errorf("no confidence intervals to export")
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
		"Percentile",
		"TotalReturn_%",
		"MaxDrawdown_%",
		"WinRate_%",
		"SharpeRatio",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Sort by percentile
	intervals := make([]ConfidenceLevel, len(mcr.ConfidenceIntervals))
	copy(intervals, mcr.ConfidenceIntervals)
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].Percentile < intervals[j].Percentile
	})

	// Write percentile data
	for _, ci := range intervals {
		record := []string{
			fmt.Sprintf("%.0f", ci.Percentile*100),
			fmt.Sprintf("%.2f", ci.TotalReturn*100),
			fmt.Sprintf("%.2f", ci.MaxDrawdown*100),
			fmt.Sprintf("%.2f", ci.WinRate*100),
			fmt.Sprintf("%.4f", ci.SharpeRatio),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	return nil
}

// ExportSimulationsToCSV exports all simulation results to CSV.
func (mcr *MCResult) ExportSimulationsToCSV(filepath string) error {
	if mcr == nil || len(mcr.Simulations) == 0 {
		return fmt.Errorf("no simulations to export")
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
		"SimulationID",
		"TotalReturn_%",
		"MaxDrawdown_%",
		"TotalTrades",
		"WinningTrades",
		"LosingTrades",
		"WinRate_%",
		"TotalPnL",
		"SharpeRatio",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write simulation results
	for i, sim := range mcr.Simulations {
		record := []string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%.2f", sim.TotalReturn*100),
			fmt.Sprintf("%.2f", sim.MaxDrawdown*100),
			fmt.Sprintf("%d", sim.TotalTrades),
			fmt.Sprintf("%d", sim.WinningTrades),
			fmt.Sprintf("%d", sim.LosingTrades),
			fmt.Sprintf("%.2f", sim.WinRate*100),
			fmt.Sprintf("%.2f", sim.TotalPnL),
			fmt.Sprintf("%.4f", sim.Sharpe),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	return nil
}

// ExportDrawdownDistributionToCSV exports drawdown distribution analysis.
func (mcr *MCResult) ExportDrawdownDistributionToCSV(filepath string) error {
	if mcr == nil || len(mcr.Simulations) == 0 {
		return fmt.Errorf("no simulations for drawdown analysis")
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Collect all max drawdowns
	drawdowns := make([]float64, len(mcr.Simulations))
	for i, sim := range mcr.Simulations {
		drawdowns[i] = sim.MaxDrawdown
	}

	// Sort for percentile calculation
	sort.Float64s(drawdowns)

	// Write header
	header := []string{"Percentile", "MaxDrawdown_%", "Interpretation"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Key percentiles
	percentiles := []float64{0.05, 0.10, 0.25, 0.50, 0.75, 0.90, 0.95, 0.99}

	for _, p := range percentiles {
		idx := int(p * float64(len(drawdowns)))
		if idx >= len(drawdowns) {
			idx = len(drawdowns) - 1
		}

		interpretation := getDrawdownInterpretation(p)

		record := []string{
			fmt.Sprintf("%.0f", p*100),
			fmt.Sprintf("%.2f", drawdowns[idx]*100),
			interpretation,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	return nil
}

// RiskAnalysis provides risk assessment based on Monte Carlo results.
type RiskAnalysis struct {
	WorstCase5Pct        float64 // 5th percentile return
	BestCase95Pct        float64 // 95th percentile return
	ProbabilityProfit    float64 // % of simulations with positive return
	ProbabilityLoss      float64 // % of simulations with negative return
	ExpectedReturn       float64 // mean return
	WorstDrawdown95Pct   float64 // 95th percentile max drawdown
	RiskOfRuin           float64 // probability of catastrophic loss
	ConsistencyScore     float64 // 0-100, based on return variance
}

// AnalyzeRisk performs comprehensive risk analysis.
func (mcr *MCResult) AnalyzeRisk() *RiskAnalysis {
	if mcr == nil || len(mcr.Simulations) == 0 {
		return nil
	}

	stats := mcr.Statistics

	// Calculate probability of profit
	profitCount := len(mcr.Simulations) - stats.NegativeReturnCount
	probProfit := float64(profitCount) / float64(len(mcr.Simulations))

	// Consistency score based on return variance
	// Lower variance = higher consistency
	// Normalize: assume typical std dev is 0.2 (20%), max reasonable is 1.0
	normalizedVolatility := stats.StdDevReturn / 1.0
	if normalizedVolatility > 1.0 {
		normalizedVolatility = 1.0
	}
	consistencyScore := (1.0 - normalizedVolatility) * 100

	analysis := &RiskAnalysis{
		WorstCase5Pct:      stats.P05Return,
		BestCase95Pct:      stats.P95Return,
		ProbabilityProfit:  probProfit,
		ProbabilityLoss:    stats.NegativeReturnRatio,
		ExpectedReturn:     stats.MeanReturn,
		WorstDrawdown95Pct: stats.P95MaxDrawdown,
		RiskOfRuin:         stats.ProbabilityOfRuin,
		ConsistencyScore:   consistencyScore,
	}

	return analysis
}

// ExportRiskAnalysisToCSV exports risk analysis to CSV.
func (mcr *MCResult) ExportRiskAnalysisToCSV(filepath string) error {
	analysis := mcr.AnalyzeRisk()
	if analysis == nil {
		return fmt.Errorf("no risk analysis available")
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Metric", "Value", "Interpretation"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write risk metrics
	metrics := [][]string{
		{"ExpectedReturn_%", fmt.Sprintf("%.2f", analysis.ExpectedReturn*100), getReturnInterpretation(analysis.ExpectedReturn)},
		{"WorstCase5Pct_%", fmt.Sprintf("%.2f", analysis.WorstCase5Pct*100), "5% chance of worse outcome"},
		{"BestCase95Pct_%", fmt.Sprintf("%.2f", analysis.BestCase95Pct*100), "5% chance of better outcome"},
		{"", "", ""},
		{"ProbabilityProfit_%", fmt.Sprintf("%.2f", analysis.ProbabilityProfit*100), getProbabilityInterpretation(analysis.ProbabilityProfit)},
		{"ProbabilityLoss_%", fmt.Sprintf("%.2f", analysis.ProbabilityLoss*100), ""},
		{"", "", ""},
		{"WorstDrawdown95Pct_%", fmt.Sprintf("%.2f", analysis.WorstDrawdown95Pct*100), "95% chance of smaller drawdown"},
		{"RiskOfRuin_%", fmt.Sprintf("%.2f", analysis.RiskOfRuin*100), getRiskOfRuinInterpretation(analysis.RiskOfRuin)},
		{"", "", ""},
		{"ConsistencyScore", fmt.Sprintf("%.2f", analysis.ConsistencyScore), getConsistencyInterpretation(analysis.ConsistencyScore)},
	}

	for _, metric := range metrics {
		if err := writer.Write(metric); err != nil {
			return fmt.Errorf("write metric: %w", err)
		}
	}

	return nil
}

// Helper functions

func getAnalysisTypeName(t MCAnalysisType) string {
	switch t {
	case TradeReshuffle:
		return "TradeReshuffle"
	case ReturnReshuffle:
		return "ReturnReshuffle"
	case BootstrapReshuffle:
		return "BootstrapReshuffle"
	default:
		return "Unknown"
	}
}

func getDrawdownInterpretation(percentile float64) string {
	if percentile <= 0.05 {
		return "Best case (5% of simulations)"
	} else if percentile <= 0.25 {
		return "Better than average"
	} else if percentile <= 0.50 {
		return "Median outcome"
	} else if percentile <= 0.75 {
		return "Worse than average"
	} else if percentile <= 0.95 {
		return "Poor outcome"
	}
	return "Worst case (95% of simulations)"
}

func getReturnInterpretation(ret float64) string {
	if ret > 0.2 {
		return "Strong expected return"
	} else if ret > 0.1 {
		return "Good expected return"
	} else if ret > 0 {
		return "Positive expected return"
	}
	return "Negative expected return - high risk"
}

func getProbabilityInterpretation(prob float64) string {
	if prob > 0.8 {
		return "Very high confidence"
	} else if prob > 0.6 {
		return "High confidence"
	} else if prob > 0.5 {
		return "Slight edge"
	}
	return "Low confidence - high risk"
}

func getRiskOfRuinInterpretation(risk float64) string {
	if risk < 0.01 {
		return "Very low risk"
	} else if risk < 0.05 {
		return "Low risk"
	} else if risk < 0.10 {
		return "Moderate risk"
	}
	return "High risk - unacceptable"
}

func getConsistencyInterpretation(score float64) string {
	if score > 80 {
		return "Excellent - very consistent returns"
	} else if score > 60 {
		return "Good - acceptable consistency"
	} else if score > 40 {
		return "Fair - moderate variance"
	}
	return "Poor - high variance, unstable"
}
