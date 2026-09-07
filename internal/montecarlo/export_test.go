package montecarlo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportStatisticsToCSV(t *testing.T) {
	mcr := createMockMCResult()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "mc_statistics.csv")

	err := mcr.ExportStatisticsToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportStatisticsToCSV failed: %v", err)
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

	// Verify key sections
	if !containsString(contentStr, "MeanReturn_%") {
		t.Error("MeanReturn not found")
	}
	if !containsString(contentStr, "MedianReturn_%") {
		t.Error("MedianReturn not found")
	}
	if !containsString(contentStr, "MaxDrawdown") {
		t.Error("MaxDrawdown not found")
	}
	if !containsString(contentStr, "ProbabilityOfRuin") {
		t.Error("ProbabilityOfRuin not found")
	}
}

func TestExportPercentilesToCSV(t *testing.T) {
	mcr := createMockMCResult()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "percentiles.csv")

	err := mcr.ExportPercentilesToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportPercentilesToCSV failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify header columns
	if !containsString(contentStr, "Percentile") {
		t.Error("Percentile column not found")
	}
	if !containsString(contentStr, "TotalReturn_%") {
		t.Error("TotalReturn column not found")
	}
	if !containsString(contentStr, "SharpeRatio") {
		t.Error("SharpeRatio column not found")
	}

	// Verify has data rows
	lines := countLines(content)
	expectedLines := len(mcr.ConfidenceIntervals) + 1 // +1 for header
	if lines != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, lines)
	}
}

func TestExportSimulationsToCSV(t *testing.T) {
	mcr := createMockMCResult()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "simulations.csv")

	err := mcr.ExportSimulationsToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportSimulationsToCSV failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify columns
	if !containsString(contentStr, "SimulationID") {
		t.Error("SimulationID column not found")
	}
	if !containsString(contentStr, "WinRate_%") {
		t.Error("WinRate column not found")
	}

	// Verify row count
	lines := countLines(content)
	expectedLines := len(mcr.Simulations) + 1 // +1 for header
	if lines != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, lines)
	}
}

func TestExportDrawdownDistributionToCSV(t *testing.T) {
	mcr := createMockMCResult()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "drawdown_dist.csv")

	err := mcr.ExportDrawdownDistributionToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportDrawdownDistributionToCSV failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify columns
	if !containsString(contentStr, "Percentile") {
		t.Error("Percentile column not found")
	}
	if !containsString(contentStr, "MaxDrawdown_%") {
		t.Error("MaxDrawdown column not found")
	}
	if !containsString(contentStr, "Interpretation") {
		t.Error("Interpretation column not found")
	}
}

func TestAnalyzeRisk(t *testing.T) {
	mcr := createMockMCResult()

	analysis := mcr.AnalyzeRisk()
	if analysis == nil {
		t.Fatal("Risk analysis returned nil")
	}

	// Verify expected return is reasonable
	if analysis.ExpectedReturn < -1.0 || analysis.ExpectedReturn > 5.0 {
		t.Errorf("Expected return out of reasonable range: %.4f", analysis.ExpectedReturn)
	}

	// Verify probability of profit + loss = 1.0
	total := analysis.ProbabilityProfit + analysis.ProbabilityLoss
	if total < 0.99 || total > 1.01 {
		t.Errorf("Probabilities don't sum to 1.0: %.4f", total)
	}

	// Verify consistency score is in valid range
	if analysis.ConsistencyScore < 0 || analysis.ConsistencyScore > 100 {
		t.Errorf("Consistency score out of range: %.2f", analysis.ConsistencyScore)
	}

	// Verify worst case is worse than expected
	if analysis.WorstCase5Pct > analysis.ExpectedReturn {
		t.Error("Worst case should be worse than expected return")
	}

	// Verify best case is better than expected
	if analysis.BestCase95Pct < analysis.ExpectedReturn {
		t.Error("Best case should be better than expected return")
	}
}

func TestExportRiskAnalysisToCSV(t *testing.T) {
	mcr := createMockMCResult()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "risk_analysis.csv")

	err := mcr.ExportRiskAnalysisToCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportRiskAnalysisToCSV failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify key risk metrics
	if !containsString(contentStr, "ExpectedReturn") {
		t.Error("ExpectedReturn not found")
	}
	if !containsString(contentStr, "WorstCase5Pct") {
		t.Error("WorstCase5Pct not found")
	}
	if !containsString(contentStr, "ProbabilityProfit") {
		t.Error("ProbabilityProfit not found")
	}
	if !containsString(contentStr, "RiskOfRuin") {
		t.Error("RiskOfRuin not found")
	}
	if !containsString(contentStr, "ConsistencyScore") {
		t.Error("ConsistencyScore not found")
	}
}

func TestExportStatistics_NilResult(t *testing.T) {
	var mcr *MCResult

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "nil.csv")

	err := mcr.ExportStatisticsToCSV(csvPath)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestExportPercentiles_EmptyIntervals(t *testing.T) {
	mcr := &MCResult{
		ConfidenceIntervals: []ConfidenceLevel{},
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")

	err := mcr.ExportPercentilesToCSV(csvPath)
	if err == nil {
		t.Error("Expected error for empty confidence intervals")
	}
}

func TestExportSimulations_EmptyResults(t *testing.T) {
	mcr := &MCResult{
		Simulations: []SimulationResult{},
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")

	err := mcr.ExportSimulationsToCSV(csvPath)
	if err == nil {
		t.Error("Expected error for empty simulations")
	}
}

func TestAnalyzeRisk_NilResult(t *testing.T) {
	var mcr *MCResult

	analysis := mcr.AnalyzeRisk()
	if analysis != nil {
		t.Error("Expected nil for nil result")
	}
}

func TestGetAnalysisTypeName(t *testing.T) {
	tests := []struct {
		typ      MCAnalysisType
		expected string
	}{
		{TradeReshuffle, "TradeReshuffle"},
		{ReturnReshuffle, "ReturnReshuffle"},
		{BootstrapReshuffle, "BootstrapReshuffle"},
		{MCAnalysisType(99), "Unknown"},
	}

	for _, tt := range tests {
		name := getAnalysisTypeName(tt.typ)
		if name != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, name)
		}
	}
}

func TestRiskInterpretations(t *testing.T) {
	// Test return interpretation
	tests := []struct {
		ret      float64
		contains string
	}{
		{0.25, "Strong"},
		{0.15, "Good"},
		{0.05, "Positive"},
		{-0.05, "Negative"},
	}

	for _, tt := range tests {
		interp := getReturnInterpretation(tt.ret)
		if !containsString(interp, tt.contains) {
			t.Errorf("For return %.2f, expected '%s' in interpretation, got '%s'",
				tt.ret, tt.contains, interp)
		}
	}
}

func TestProbabilityInterpretation(t *testing.T) {
	tests := []struct {
		prob     float64
		contains string
	}{
		{0.85, "Very high"},
		{0.70, "High"},
		{0.55, "Slight"},
		{0.40, "Low"},
	}

	for _, tt := range tests {
		interp := getProbabilityInterpretation(tt.prob)
		if !containsString(interp, tt.contains) {
			t.Errorf("For probability %.2f, expected '%s' in interpretation, got '%s'",
				tt.prob, tt.contains, interp)
		}
	}
}

func TestRiskOfRuinInterpretation(t *testing.T) {
	tests := []struct {
		risk     float64
		contains string
	}{
		{0.005, "Very low"},
		{0.03, "Low"},
		{0.07, "Moderate"},
		{0.15, "High"},
	}

	for _, tt := range tests {
		interp := getRiskOfRuinInterpretation(tt.risk)
		if !containsString(interp, tt.contains) {
			t.Errorf("For risk %.3f, expected '%s' in interpretation, got '%s'",
				tt.risk, tt.contains, interp)
		}
	}
}

func TestConsistencyInterpretation(t *testing.T) {
	tests := []struct {
		score    float64
		contains string
	}{
		{85, "Excellent"},
		{70, "Good"},
		{50, "Fair"},
		{30, "Poor"},
	}

	for _, tt := range tests {
		interp := getConsistencyInterpretation(tt.score)
		if !containsString(interp, tt.contains) {
			t.Errorf("For score %.0f, expected '%s' in interpretation, got '%s'",
				tt.score, tt.contains, interp)
		}
	}
}

func TestMultipleSimulations(t *testing.T) {
	// Create result with many simulations
	mcr := &MCResult{
		Config: MCConfig{
			Simulations: 1000,
			Seed:        42,
			Type:        TradeReshuffle,
		},
		Simulations: make([]SimulationResult, 1000),
		Statistics: MCStatistics{
			MeanReturn:          0.15,
			MedianReturn:        0.14,
			StdDevReturn:        0.08,
			P05Return:           0.02,
			P95Return:           0.28,
			MeanMaxDrawdown:     0.12,
			P95MaxDrawdown:      0.22,
			ProbabilityOfRuin:   0.01,
			NegativeReturnCount: 50,
			NegativeReturnRatio: 0.05,
		},
		ConfidenceIntervals: []ConfidenceLevel{
			{Percentile: 0.05, TotalReturn: 0.02, MaxDrawdown: 0.08},
			{Percentile: 0.50, TotalReturn: 0.14, MaxDrawdown: 0.12},
			{Percentile: 0.95, TotalReturn: 0.28, MaxDrawdown: 0.22},
		},
	}

	// Initialize simulations with varying results
	for i := 0; i < 1000; i++ {
		ret := 0.15 + float64(i-500)*0.0003 // Varying returns
		dd := 0.12 + float64(i)*0.0001      // Varying drawdowns

		mcr.Simulations[i] = SimulationResult{
			TotalReturn:   ret,
			MaxDrawdown:   dd,
			TotalTrades:   100,
			WinningTrades: 55,
			LosingTrades:  45,
			WinRate:       0.55,
			Sharpe:        1.5,
		}
	}

	tmpDir := t.TempDir()

	// Test all exports work with many simulations
	if err := mcr.ExportStatisticsToCSV(filepath.Join(tmpDir, "stats.csv")); err != nil {
		t.Errorf("Export statistics failed: %v", err)
	}

	if err := mcr.ExportSimulationsToCSV(filepath.Join(tmpDir, "sims.csv")); err != nil {
		t.Errorf("Export simulations failed: %v", err)
	}

	if err := mcr.ExportPercentilesToCSV(filepath.Join(tmpDir, "percentiles.csv")); err != nil {
		t.Errorf("Export percentiles failed: %v", err)
	}

	// Verify simulation count in CSV
	content, _ := os.ReadFile(filepath.Join(tmpDir, "sims.csv"))
	lines := countLines(content)
	if lines != 1001 { // 1000 simulations + header
		t.Errorf("Expected 1001 lines, got %d", lines)
	}
}

// Helper functions

func createMockMCResult() *MCResult {
	return &MCResult{
		Config: MCConfig{
			Simulations: 5,
			Seed:        42,
			Type:        TradeReshuffle,
		},
		Simulations: []SimulationResult{
			{TotalReturn: 0.12, MaxDrawdown: 0.08, TotalTrades: 50, WinningTrades: 28, LosingTrades: 22, WinRate: 0.56, Sharpe: 1.4},
			{TotalReturn: 0.08, MaxDrawdown: 0.12, TotalTrades: 50, WinningTrades: 26, LosingTrades: 24, WinRate: 0.52, Sharpe: 1.1},
			{TotalReturn: -0.02, MaxDrawdown: 0.18, TotalTrades: 50, WinningTrades: 23, LosingTrades: 27, WinRate: 0.46, Sharpe: 0.8},
			{TotalReturn: 0.15, MaxDrawdown: 0.10, TotalTrades: 50, WinningTrades: 30, LosingTrades: 20, WinRate: 0.60, Sharpe: 1.6},
			{TotalReturn: 0.10, MaxDrawdown: 0.11, TotalTrades: 50, WinningTrades: 27, LosingTrades: 23, WinRate: 0.54, Sharpe: 1.3},
		},
		Statistics: MCStatistics{
			MeanReturn:          0.086,
			StdDevReturn:        0.06,
			MinReturn:           -0.02,
			MaxReturn:           0.15,
			MedianReturn:        0.10,
			P05Return:           -0.02,
			P95Return:           0.15,
			MeanMaxDrawdown:     0.118,
			StdDevMaxDrawdown:   0.035,
			MinMaxDrawdown:      0.08,
			MaxMaxDrawdown:      0.18,
			MedianMaxDrawdown:   0.11,
			P95MaxDrawdown:      0.18,
			MeanWinRate:         0.536,
			StdDevWinRate:       0.048,
			MinWinRate:          0.46,
			MaxWinRate:          0.60,
			MeanSharpe:          1.24,
			StdDevSharpe:        0.28,
			MinSharpe:           0.8,
			MaxSharpe:           1.6,
			ProbabilityOfRuin:   0.0,
			NegativeReturnCount: 1,
			NegativeReturnRatio: 0.2,
		},
		ConfidenceIntervals: []ConfidenceLevel{
			{Percentile: 0.05, TotalReturn: -0.02, MaxDrawdown: 0.08, WinRate: 0.46, SharpeRatio: 0.8},
			{Percentile: 0.25, TotalReturn: 0.08, MaxDrawdown: 0.10, WinRate: 0.52, SharpeRatio: 1.1},
			{Percentile: 0.50, TotalReturn: 0.10, MaxDrawdown: 0.11, WinRate: 0.54, SharpeRatio: 1.3},
			{Percentile: 0.75, TotalReturn: 0.12, MaxDrawdown: 0.12, WinRate: 0.56, SharpeRatio: 1.4},
			{Percentile: 0.95, TotalReturn: 0.15, MaxDrawdown: 0.18, WinRate: 0.60, SharpeRatio: 1.6},
		},
	}
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
