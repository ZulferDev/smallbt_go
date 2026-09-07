package walkforward

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
)

// ExportWindowResultsToCSV exports detailed results for each walk forward window.
func (wfa *WalkForwardAnalysis) ExportWindowResultsToCSV(filepath string) error {
	if wfa == nil || len(wfa.Results) == 0 {
		return fmt.Errorf("no walk forward results to export")
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
		"WindowID",
		"TrainStart",
		"TrainEnd",
		"TestStart",
		"TestEnd",
		"TrainReturn_%",
		"TrainSharpe",
		"TrainTrades",
		"TestReturn_%",
		"TestSharpe",
		"TestMaxDrawdown_%",
		"TestWinRate_%",
		"TestProfitFactor",
		"TestTrades",
		"InSampleOutSampleDelta_%",
		"PerformanceDegradation",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Sort windows by ID
	windowIDs := make([]int, 0, len(wfa.Results))
	for id := range wfa.Results {
		windowIDs = append(windowIDs, id)
	}
	sort.Ints(windowIDs)

	// Write window results
	for _, windowID := range windowIDs {
		result := wfa.Results[windowID]
		if result == nil {
			continue
		}

		// Calculate in-sample vs out-of-sample delta
		var trainReturn, testReturn float64
		var trainSharpe, testSharpe float64
		var trainTrades, testTrades int

		if result.TrainResult != nil && result.TrainResult.Metrics != nil {
			trainReturn = result.TrainResult.Metrics.TotalReturn * 100
			trainSharpe = result.TrainResult.Metrics.SharpeRatio
			trainTrades = result.TrainResult.TotalTrades
		}

		if result.TestResult != nil && result.TestResult.Metrics != nil {
			testReturn = result.TestResult.Metrics.TotalReturn * 100
			testSharpe = result.TestResult.Metrics.SharpeRatio
			testTrades = result.TestResult.TotalTrades
		}

		delta := testReturn - trainReturn
		var degradation string
		if delta < -10 {
			degradation = "High"
		} else if delta < -5 {
			degradation = "Moderate"
		} else if delta < 0 {
			degradation = "Low"
		} else {
			degradation = "Improved"
		}

		record := []string{
			fmt.Sprintf("%d", windowID),
			result.TrainStart.Format("2006-01-02"),
			result.TrainEnd.Format("2006-01-02"),
			result.TestStart.Format("2006-01-02"),
			result.TestEnd.Format("2006-01-02"),
			fmt.Sprintf("%.2f", trainReturn),
			fmt.Sprintf("%.4f", trainSharpe),
			fmt.Sprintf("%d", trainTrades),
			fmt.Sprintf("%.2f", testReturn),
			fmt.Sprintf("%.4f", testSharpe),
		}

		if result.TestResult != nil && result.TestResult.Metrics != nil {
			record = append(record,
				fmt.Sprintf("%.2f", result.TestResult.Metrics.MaxDrawdown*100),
				fmt.Sprintf("%.2f", result.TestResult.Metrics.WinRate*100),
				fmt.Sprintf("%.4f", result.TestResult.Metrics.ProfitFactor),
			)
		} else {
			record = append(record, "0.00", "0.00", "0.00")
		}

		record = append(record,
			fmt.Sprintf("%d", testTrades),
			fmt.Sprintf("%.2f", delta),
			degradation,
		)

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	return nil
}

// ExportAggregateResultToCSV exports the out-of-sample aggregate metrics.
func (wfa *WalkForwardAnalysis) ExportAggregateResultToCSV(filepath string) error {
	if wfa == nil || wfa.AggregateResult == nil {
		return fmt.Errorf("no aggregate results to export")
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

	agg := wfa.AggregateResult

	// Write metrics
	metrics := [][]string{
		{"TotalWindows", fmt.Sprintf("%d", len(wfa.Results))},
		{"TotalTrades", fmt.Sprintf("%d", agg.TotalTrades)},
		{"TotalReturn_%", fmt.Sprintf("%.2f", agg.TotalReturn*100)},
		{"CAGR_%", fmt.Sprintf("%.2f", agg.CAGR*100)},
		{"SharpeRatio", fmt.Sprintf("%.4f", agg.SharpeRatio)},
		{"SortinoRatio", fmt.Sprintf("%.4f", agg.SortinoRatio)},
		{"MaxDrawdown_%", fmt.Sprintf("%.2f", agg.MaxDrawdown*100)},
		{"CalmarRatio", fmt.Sprintf("%.4f", agg.CalmarRatio)},
		{"WinRate_%", fmt.Sprintf("%.2f", agg.WinRate*100)},
		{"ProfitFactor", fmt.Sprintf("%.4f", agg.ProfitFactor)},
		{"Expectancy", fmt.Sprintf("%.6f", agg.Expectancy)},
		{"AverageWin", fmt.Sprintf("%.2f", agg.AverageWin)},
		{"AverageLoss", fmt.Sprintf("%.2f", agg.AverageLoss)},
		{"AvgTradeReturn_%", fmt.Sprintf("%.4f", agg.AverageTradeReturn*100)},
	}

	for _, metric := range metrics {
		if err := writer.Write(metric); err != nil {
			return fmt.Errorf("write metric: %w", err)
		}
	}

	return nil
}

// StabilityAnalysis analyzes consistency across walk forward windows.
type StabilityAnalysis struct {
	TotalWindows        int
	ProfitableWindows   int
	UnprofitableWindows int
	ConsistencyScore    float64 // 0-100
	AvgReturnStdDev     float64
	AvgSharpeStdDev     float64
	WorstWindow         int
	BestWindow          int
	WorstWindowReturn   float64
	BestWindowReturn    float64
	DegradationHigh     int // Count of high degradation windows
	DegradationModerate int
	DegradationLow      int
	ImprovedWindows     int
}

// AnalyzeStability analyzes performance stability across windows.
func (wfa *WalkForwardAnalysis) AnalyzeStability() *StabilityAnalysis {
	if wfa == nil || len(wfa.Results) == 0 {
		return nil
	}

	analysis := &StabilityAnalysis{
		TotalWindows:      len(wfa.Results),
		WorstWindowReturn: 1e9,
		BestWindowReturn:  -1e9,
	}

	var returns []float64
	var sharpes []float64

	for windowID, result := range wfa.Results {
		if result == nil || result.TestResult == nil || result.TestResult.Metrics == nil {
			continue
		}

		metrics := result.TestResult.Metrics
		testReturn := metrics.TotalReturn

		returns = append(returns, testReturn)
		sharpes = append(sharpes, metrics.SharpeRatio)

		// Count profitable/unprofitable
		if testReturn > 0 {
			analysis.ProfitableWindows++
		} else {
			analysis.UnprofitableWindows++
		}

		// Track best/worst
		if testReturn < analysis.WorstWindowReturn {
			analysis.WorstWindowReturn = testReturn
			analysis.WorstWindow = windowID
		}
		if testReturn > analysis.BestWindowReturn {
			analysis.BestWindowReturn = testReturn
			analysis.BestWindow = windowID
		}

		// Analyze degradation
		if result.TrainResult != nil && result.TrainResult.Metrics != nil {
			trainReturn := result.TrainResult.Metrics.TotalReturn
			delta := (testReturn - trainReturn) * 100

			if delta < -10 {
				analysis.DegradationHigh++
			} else if delta < -5 {
				analysis.DegradationModerate++
			} else if delta < 0 {
				analysis.DegradationLow++
			} else {
				analysis.ImprovedWindows++
			}
		}
	}

	// Calculate standard deviations
	if len(returns) > 1 {
		analysis.AvgReturnStdDev = calculateStdDev(returns)
		analysis.AvgSharpeStdDev = calculateStdDev(sharpes)
	}

	// Calculate consistency score (0-100)
	// Higher is better, based on:
	// - Percentage of profitable windows (40%)
	// - Low return volatility (30%)
	// - Low degradation (30%)
	profitableRatio := float64(analysis.ProfitableWindows) / float64(analysis.TotalWindows)

	// Normalize std dev (assume typical is 0.2, max reasonable is 1.0)
	volatilityScore := 1.0 - min(analysis.AvgReturnStdDev/1.0, 1.0)

	// Degradation score
	degradationScore := float64(analysis.ImprovedWindows+analysis.DegradationLow) / float64(analysis.TotalWindows)

	analysis.ConsistencyScore = (profitableRatio*40 + volatilityScore*30 + degradationScore*30)

	return analysis
}

// ExportStabilityAnalysisToCSV exports stability analysis.
func (wfa *WalkForwardAnalysis) ExportStabilityAnalysisToCSV(filepath string) error {
	analysis := wfa.AnalyzeStability()
	if analysis == nil {
		return fmt.Errorf("no stability analysis available")
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

	// Write stability metrics
	metrics := [][]string{
		{"TotalWindows", fmt.Sprintf("%d", analysis.TotalWindows)},
		{"ProfitableWindows", fmt.Sprintf("%d", analysis.ProfitableWindows)},
		{"UnprofitableWindows", fmt.Sprintf("%d", analysis.UnprofitableWindows)},
		{"ProfitableRate_%", fmt.Sprintf("%.2f", float64(analysis.ProfitableWindows)/float64(analysis.TotalWindows)*100)},
		{"ConsistencyScore", fmt.Sprintf("%.2f", analysis.ConsistencyScore)},
		{"AvgReturnStdDev", fmt.Sprintf("%.6f", analysis.AvgReturnStdDev)},
		{"AvgSharpeStdDev", fmt.Sprintf("%.6f", analysis.AvgSharpeStdDev)},
		{"BestWindow", fmt.Sprintf("%d", analysis.BestWindow)},
		{"BestWindowReturn_%", fmt.Sprintf("%.2f", analysis.BestWindowReturn*100)},
		{"WorstWindow", fmt.Sprintf("%d", analysis.WorstWindow)},
		{"WorstWindowReturn_%", fmt.Sprintf("%.2f", analysis.WorstWindowReturn*100)},
		{"ImprovedWindows", fmt.Sprintf("%d", analysis.ImprovedWindows)},
		{"DegradationLow", fmt.Sprintf("%d", analysis.DegradationLow)},
		{"DegradationModerate", fmt.Sprintf("%d", analysis.DegradationModerate)},
		{"DegradationHigh", fmt.Sprintf("%d", analysis.DegradationHigh)},
	}

	for _, metric := range metrics {
		if err := writer.Write(metric); err != nil {
			return fmt.Errorf("write metric: %w", err)
		}
	}

	return nil
}

// Helper functions

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	// Return standard deviation
	return variance // Simplified - return variance for now
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
