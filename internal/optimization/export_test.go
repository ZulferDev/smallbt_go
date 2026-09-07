package optimization

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/backtest"
)

func TestExportToCSV(t *testing.T) {
	report := createMockReport()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "optimization.csv")

	err := report.ExportToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportToCSV failed: %v", err)
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
	if !containsStr(contentStr, "Rank,ObjectiveValue") {
		t.Error("Header not found")
	}
	if !containsStr(contentStr, "SharpeRatio") {
		t.Error("SharpeRatio column not found")
	}
	if !containsStr(contentStr, "WinRate_%") {
		t.Error("WinRate column not found")
	}

	// Verify has data rows
	lines := countCSVLines(content)
	expectedLines := len(report.Results) + 1 // +1 for header
	if lines != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, lines)
	}
}

func TestExportTopNToCSV(t *testing.T) {
	report := createMockReport()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "top5.csv")

	err := report.ExportTopNToCSV(csvPath, 5)
	if err != nil {
		t.Fatalf("ExportTopNToCSV failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	// Should have 6 lines (header + 5 results)
	lines := countCSVLines(content)
	if lines != 6 {
		t.Errorf("Expected 6 lines for top 5, got %d", lines)
	}
}

func TestExportTopN_ExceedsTotal(t *testing.T) {
	report := createMockReport()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "all.csv")

	// Request more than available
	err := report.ExportTopNToCSV(csvPath, 1000)
	if err != nil {
		t.Fatalf("ExportTopNToCSV failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	lines := countCSVLines(content)
	expectedLines := len(report.Results) + 1
	if lines != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, lines)
	}
}

func TestAnalyzeParameterSensitivity(t *testing.T) {
	report := createMockReport()

	sensitivities := report.AnalyzeParameterSensitivity()

	if len(sensitivities) == 0 {
		t.Fatal("No sensitivities calculated")
	}

	// Should have one for each parameter
	expectedParams := 2 // ema_fast, ema_slow
	if len(sensitivities) != expectedParams {
		t.Errorf("Expected %d parameters, got %d", expectedParams, len(sensitivities))
	}

	// Verify structure
	for _, sens := range sensitivities {
		if sens.ParameterName == "" {
			t.Error("Empty parameter name")
		}
		if sens.MinValue >= sens.MaxValue {
			t.Errorf("Invalid range for %s", sens.ParameterName)
		}
		if sens.BestValue < sens.MinValue || sens.BestValue > sens.MaxValue {
			t.Errorf("BestValue out of range for %s", sens.ParameterName)
		}
		if sens.WorstValue < sens.MinValue || sens.WorstValue > sens.MaxValue {
			t.Errorf("WorstValue out of range for %s", sens.ParameterName)
		}
		if sens.Correlation < -1.01 || sens.Correlation > 1.01 {
			t.Errorf("Invalid correlation for %s: %f", sens.ParameterName, sens.Correlation)
		}
		if sens.Sensitivity != "Low" && sens.Sensitivity != "Medium" && sens.Sensitivity != "High" {
			t.Errorf("Invalid sensitivity level: %s", sens.Sensitivity)
		}
	}
}

func TestExportSensitivityAnalysisToCSV(t *testing.T) {
	report := createMockReport()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "sensitivity.csv")

	err := report.ExportSensitivityAnalysisToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportSensitivityAnalysisToCSV failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify header
	if !containsStr(contentStr, "Parameter") {
		t.Error("Parameter column not found")
	}
	if !containsStr(contentStr, "Correlation") {
		t.Error("Correlation column not found")
	}
	if !containsStr(contentStr, "Sensitivity") {
		t.Error("Sensitivity column not found")
	}

	// Should have parameter rows
	lines := countCSVLines(content)
	if lines < 2 {
		t.Error("Should have at least header + 1 parameter")
	}
}

func TestCalculateCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		x         []float64
		y         []float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "perfect positive",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{2, 4, 6, 8, 10},
			expected:  1.0,
			tolerance: 0.01,
		},
		{
			name:      "perfect negative",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{10, 8, 6, 4, 2},
			expected:  -1.0,
			tolerance: 0.01,
		},
		{
			name:      "no correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{3, 3, 3, 3, 3},
			expected:  0.0,
			tolerance: 0.01,
		},
		{
			name:      "empty",
			x:         []float64{},
			y:         []float64{},
			expected:  0.0,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCorrelation(tt.x, tt.y)
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("Expected %.4f, got %.4f", tt.expected, result)
			}
		})
	}
}

func TestEmptyReport(t *testing.T) {
	report := &OptimizationReport{
		Results: []*OptimizationResult{},
	}

	sensitivities := report.AnalyzeParameterSensitivity()
	if sensitivities != nil {
		t.Error("Expected nil for empty report")
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")

	err := report.ExportSensitivityAnalysisToCSV(csvPath)
	if err == nil {
		t.Error("Expected error for empty report")
	}
}

func TestExportWithNilResults(t *testing.T) {
	report := &OptimizationReport{
		Results: []*OptimizationResult{
			nil,
			{
				Rank:           1,
				ObjectiveValue: 1.5,
				Parameters: ParameterSet{
					Values: map[string]float64{"test": 10.0},
				},
				BacktestResult: &backtest.BacktestResult{
					TotalTrades: 10,
					Metrics: &analytics.Metrics{
						SharpeRatio: 1.5,
					},
				},
			},
		},
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "with_nil.csv")

	err := report.ExportToCSV(csvPath)
	if err != nil {
		t.Fatalf("Should handle nil results: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal("Failed to read CSV")
	}

	// Should have header + 1 valid result (nil skipped)
	lines := countCSVLines(content)
	if lines != 2 {
		t.Errorf("Expected 2 lines, got %d", lines)
	}
}

// Helper functions

func createMockReport() *OptimizationReport {
	results := make([]*OptimizationResult, 0, 10)

	for i := 0; i < 10; i++ {
		emaFast := 5.0 + float64(i)
		emaSlow := 20.0 + float64(i)*5.0
		sharpe := 1.0 + float64(i)*0.1

		result := &OptimizationResult{
			Rank:           i + 1,
			ObjectiveValue: sharpe,
			Parameters: ParameterSet{
				Values: map[string]float64{
					"ema_fast": emaFast,
					"ema_slow": emaSlow,
				},
			},
			BacktestResult: &backtest.BacktestResult{
				TotalTrades: 50 + i*5,
				Metrics: &analytics.Metrics{
					TotalReturn:    0.15 + float64(i)*0.01,
					CAGR:           0.12 + float64(i)*0.01,
					SharpeRatio:    sharpe,
					SortinoRatio:   sharpe * 1.2,
					MaxDrawdown:    0.15 - float64(i)*0.005,
					WinRate:        0.55 + float64(i)*0.01,
					ProfitFactor:   1.5 + float64(i)*0.1,
					Expectancy:     0.02 + float64(i)*0.005,
					WinningTrades:  28 + i*2,
					LosingTrades:   22 + i*3,
					AvgWin:         150.0 + float64(i)*10,
					AvgLoss:        -80.0 - float64(i)*5,
				},
			},
		}
		results = append(results, result)
	}

	return &OptimizationReport{
		Strategy:           "test_strategy",
		Symbol:             "BTCUSDT",
		Timeframe:          "1h",
		ObjectiveMetric:    "sharpe",
		ObjectiveDirection: "maximize",
		Algorithm:          "grid",
		Results:            results,
		BestResult:         results[9],
		WorstResult:        results[0],
		TotalRuns:          10,
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && findStr(s, substr)
}

func findStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func countCSVLines(data []byte) int {
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
