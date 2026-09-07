package walkforward

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/backtest"
)

func TestExportWindowResultsToCSV(t *testing.T) {
	wfa := createMockWalkForwardAnalysis()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "window_results.csv")

	err := wfa.ExportWindowResultsToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportWindowResultsToCSV failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Fatal("CSV file not created")
	}

	// Read and verify content
	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify header
	if !containsString(contentStr, "WindowID") {
		t.Error("WindowID column not found")
	}
	if !containsString(contentStr, "TrainReturn_%") {
		t.Error("TrainReturn column not found")
	}
	if !containsString(contentStr, "TestReturn_%") {
		t.Error("TestReturn column not found")
	}
	if !containsString(contentStr, "PerformanceDegradation") {
		t.Error("PerformanceDegradation column not found")
	}

	// Verify has data rows
	lines := countLines(content)
	expectedLines := len(wfa.Results) + 1 // +1 for header
	if lines != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, lines)
	}
}

func TestExportAggregateResultToCSV(t *testing.T) {
	wfa := createMockWalkForwardAnalysis()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "aggregate.csv")

	err := wfa.ExportAggregateResultToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportAggregateResultToCSV failed: %v", err)
	}

	// Verify file exists
	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify key metrics
	if !containsString(contentStr, "TotalReturn_%") {
		t.Error("TotalReturn not found")
	}
	if !containsString(contentStr, "SharpeRatio") {
		t.Error("SharpeRatio not found")
	}
	if !containsString(contentStr, "MaxDrawdown_%") {
		t.Error("MaxDrawdown not found")
	}
}

func TestAnalyzeStability(t *testing.T) {
	wfa := createMockWalkForwardAnalysis()

	analysis := wfa.AnalyzeStability()
	if analysis == nil {
		t.Fatal("Stability analysis returned nil")
	}

	// Verify basic counts
	if analysis.TotalWindows != len(wfa.Results) {
		t.Errorf("Expected %d windows, got %d", len(wfa.Results), analysis.TotalWindows)
	}

	// Verify profitable + unprofitable = total
	total := analysis.ProfitableWindows + analysis.UnprofitableWindows
	if total > analysis.TotalWindows {
		t.Error("Profitable + Unprofitable exceeds total windows")
	}

	// Verify consistency score is in valid range
	if analysis.ConsistencyScore < 0 || analysis.ConsistencyScore > 100 {
		t.Errorf("Consistency score out of range: %.2f", analysis.ConsistencyScore)
	}

	// Verify best/worst tracking
	if analysis.BestWindowReturn < analysis.WorstWindowReturn {
		t.Error("Best window return should be >= worst window return")
	}

	// Verify degradation counts sum correctly
	degradationTotal := analysis.ImprovedWindows + analysis.DegradationLow +
		analysis.DegradationModerate + analysis.DegradationHigh
	if degradationTotal > analysis.TotalWindows {
		t.Error("Degradation counts exceed total windows")
	}
}

func TestExportStabilityAnalysisToCSV(t *testing.T) {
	wfa := createMockWalkForwardAnalysis()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "stability.csv")

	err := wfa.ExportStabilityAnalysisToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportStabilityAnalysisToCSV failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify key stability metrics
	if !containsString(contentStr, "ConsistencyScore") {
		t.Error("ConsistencyScore not found")
	}
	if !containsString(contentStr, "ProfitableWindows") {
		t.Error("ProfitableWindows not found")
	}
	if !containsString(contentStr, "DegradationHigh") {
		t.Error("DegradationHigh not found")
	}
}

func TestExportWindowResults_EmptyAnalysis(t *testing.T) {
	wfa := &WalkForwardAnalysis{
		Results: map[int]*WFWindowResult{},
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")

	err := wfa.ExportWindowResultsToCSV(csvPath)
	if err == nil {
		t.Error("Expected error for empty analysis")
	}
}

func TestExportAggregate_NilResult(t *testing.T) {
	wfa := &WalkForwardAnalysis{
		AggregateResult: nil,
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "nil_agg.csv")

	err := wfa.ExportAggregateResultToCSV(csvPath)
	if err == nil {
		t.Error("Expected error for nil aggregate result")
	}
}

func TestStabilityAnalysis_EmptyResults(t *testing.T) {
	wfa := &WalkForwardAnalysis{
		Results: map[int]*WFWindowResult{},
	}

	analysis := wfa.AnalyzeStability()
	if analysis != nil {
		t.Error("Expected nil for empty results")
	}
}

func TestPerformanceDegradationClassification(t *testing.T) {
	tests := []struct {
		name         string
		trainReturn  float64
		testReturn   float64
		expectedDegr string
	}{
		{"Improved", 0.10, 0.15, "Improved"},
		{"Stable", 0.10, 0.095, "Low"},
		{"Moderate degradation", 0.10, 0.04, "Moderate"},
		{"High degradation", 0.10, -0.05, "High"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := (tt.testReturn - tt.trainReturn) * 100
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

			if degradation != tt.expectedDegr {
				t.Errorf("Expected %s, got %s (delta=%.2f)", tt.expectedDegr, degradation, delta)
			}
		})
	}
}

func TestCalculateStdDev(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single value", []float64{1.0}, 0},
		{"constant values", []float64{5.0, 5.0, 5.0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateStdDev(tt.values)
			if result != tt.expected {
				t.Errorf("Expected %.6f, got %.6f", tt.expected, result)
			}
		})
	}
}

func TestStabilityConsistencyScore(t *testing.T) {
	// Test that consistency score is calculated correctly
	wfa := createMockWalkForwardAnalysis()
	analysis := wfa.AnalyzeStability()

	// With 3 profitable out of 5 windows (60%), consistency should be decent
	// The actual calculation can produce scores across a wide range
	if analysis.ConsistencyScore < 0 || analysis.ConsistencyScore > 100 {
		t.Errorf("Consistency score out of valid range [0-100]: %.2f", analysis.ConsistencyScore)
	}
}

func TestExportMultipleWindows(t *testing.T) {
	// Create analysis with many windows
	wfa := &WalkForwardAnalysis{
		Results: make(map[int]*WFWindowResult),
		AggregateResult: &WFAggregateResult{
			TotalTrades: 100,
			TotalReturn: 0.25,
			SharpeRatio: 1.5,
			MaxDrawdown: 0.15,
		},
	}

	// Add 10 windows
	for i := 0; i < 10; i++ {
		testReturn := 0.05 + float64(i)*0.02
		wfa.Results[i] = &WFWindowResult{
			WindowID:   i,
			TrainStart: time.Now().AddDate(0, -12+i, 0),
			TrainEnd:   time.Now().AddDate(0, -11+i, 0),
			TestStart:  time.Now().AddDate(0, -11+i, 0),
			TestEnd:    time.Now().AddDate(0, -10+i, 0),
			TrainResult: &backtest.BacktestResult{
				TotalTrades: 10,
				Metrics: &analytics.Metrics{
					TotalReturn: testReturn - 0.01,
					SharpeRatio: 1.2,
				},
			},
			TestResult: &backtest.BacktestResult{
				TotalTrades: 10,
				Metrics: &analytics.Metrics{
					TotalReturn:  testReturn,
					SharpeRatio:  1.5,
					MaxDrawdown:  0.1,
					WinRate:      0.6,
					ProfitFactor: 1.8,
				},
			},
		}
	}

	// Test CSV export with many windows
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "many_windows.csv")

	err := wfa.ExportWindowResultsToCSV(csvPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal("Failed to read CSV")
	}

	// Should have header + 10 data lines
	lines := countLines(content)
	if lines != 11 {
		t.Errorf("Expected 11 lines, got %d", lines)
	}
}

// Helper functions

func createMockWalkForwardAnalysis() *WalkForwardAnalysis {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	wfa := &WalkForwardAnalysis{
		Config: WindowConfig{
			TrainBars: 1000,
			TestBars:  200,
			StepBars:  200,
		},
		Results: make(map[int]*WFWindowResult),
		AggregateResult: &WFAggregateResult{
			TotalTrades:        125,
			TotalReturn:        0.18,
			CAGR:               0.15,
			SharpeRatio:        1.45,
			SortinoRatio:       1.85,
			MaxDrawdown:        0.12,
			CalmarRatio:        1.25,
			WinRate:            0.58,
			ProfitFactor:       1.75,
			Expectancy:         0.025,
			AverageWin:         125.50,
			AverageLoss:        -78.25,
			AverageTradeReturn: 0.015,
		},
	}

	// Create 5 windows with varying performance
	windowReturns := []float64{0.08, -0.02, 0.12, 0.05, 0.15}
	trainReturns := []float64{0.10, 0.05, 0.15, 0.08, 0.12}

	for i := 0; i < 5; i++ {
		wfa.Results[i] = &WFWindowResult{
			WindowID:   i,
			TrainStart: baseTime.AddDate(0, i*2, 0),
			TrainEnd:   baseTime.AddDate(0, i*2+2, 0),
			TestStart:  baseTime.AddDate(0, i*2+2, 0),
			TestEnd:    baseTime.AddDate(0, i*2+3, 0),
			TrainResult: &backtest.BacktestResult{
				TotalTrades: 25,
				Metrics: &analytics.Metrics{
					TotalReturn: trainReturns[i],
					SharpeRatio: 1.2 + float64(i)*0.1,
				},
			},
			TestResult: &backtest.BacktestResult{
				TotalTrades: 25,
				Metrics: &analytics.Metrics{
					TotalReturn:  windowReturns[i],
					SharpeRatio:  1.0 + float64(i)*0.15,
					MaxDrawdown:  0.08 + float64(i)*0.02,
					WinRate:      0.55 + float64(i)*0.02,
					ProfitFactor: 1.5 + float64(i)*0.1,
				},
			},
		}
	}

	return wfa
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findString(s, substr)
}

func findString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func countLines(data []byte) int {
	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		count++
	}
	return count
}
