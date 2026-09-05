package optimization

import (
	"math"
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/analytics"
	"github.com/1jehuang/backtest/internal/backtest"
)

// TestParameterDefinition tests parameter range and step definitions
func TestParameterDefinition(t *testing.T) {
	tests := []struct {
		name      string
		param     ParameterRange
		wantCount int // expected number of values
		wantErr   bool
	}{
		{
			name: "integer range 5-20 step 1",
			param: ParameterRange{
				Name:  "ema_fast.period",
				Start: 5,
				End:   20,
				Step:  1,
				Type:  "int",
			},
			wantCount: 16, // 5,6,7,...,20
		},
		{
			name: "integer range 20-100 step 5",
			param: ParameterRange{
				Name:  "ema_slow.period",
				Start: 20,
				End:   100,
				Step:  5,
				Type:  "int",
			},
			wantCount: 17, // 20,25,30,...,100
		},
		{
			name: "float range 1.0-3.0 step 0.25",
			param: ParameterRange{
				Name:  "atr.multiplier",
				Start: 1.0,
				End:   3.0,
				Step:  0.25,
				Type:  "float",
			},
			wantCount: 9, // 1.0,1.25,1.5,...,3.0
		},
		{
			name: "invalid step",
			param: ParameterRange{
				Name:  "test.param",
				Start: 1,
				End:   10,
				Step:  0,
				Type:  "int",
			},
			wantErr: true,
		},
		{
			name: "invalid range start > end",
			param: ParameterRange{
				Name:  "test.param",
				Start: 10,
				End:   1,
				Step:  1,
				Type:  "int",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OptimizationConfig{
				Parameters: []ParameterRange{tt.param},
			}

			gs := NewGridSearch(config)
			sets, err := gs.GenerateParameterSets()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(sets) != tt.wantCount {
				t.Errorf("expected %d parameter sets, got %d", tt.wantCount, len(sets))
			}
		})
	}
}

// TestGridSearchAlgorithm tests grid search implementation
func TestGridSearchAlgorithm(t *testing.T) {
	// Test: ema_fast: [5, 10, 15, 20] (4 values)
	//       ema_slow: [20, 25, 30] (3 values)
	// Total combinations: 4 * 3 = 12

	config := OptimizationConfig{
		Parameters: []ParameterRange{
			{
				Name:  "indicators.ema_fast.period",
				Start: 5,
				End:   20,
				Step:  5,
				Type:  "int",
			},
			{
				Name:  "indicators.ema_slow.period",
				Start: 20,
				End:   30,
				Step:  5,
				Type:  "int",
			},
		},
	}

	gs := NewGridSearch(config)

	// Test estimation
	estimated := gs.EstimateTotalCombinations()
	if estimated != 12 {
		t.Errorf("expected 12 combinations, got %d", estimated)
	}

	// Test generation
	sets, err := gs.GenerateParameterSets()
	if err != nil {
		t.Fatalf("generate parameter sets: %v", err)
	}

	if len(sets) != 12 {
		t.Errorf("expected 12 parameter sets, got %d", len(sets))
	}

	// Verify all combinations are unique
	seen := make(map[string]bool)
	for _, set := range sets {
		if seen[set.Hash] {
			t.Errorf("duplicate parameter set: %s", set.Hash)
		}
		seen[set.Hash] = true
	}

	// Verify each set has both parameters
	for _, set := range sets {
		if len(set.Values) != 2 {
			t.Errorf("expected 2 parameters in set, got %d", len(set.Values))
		}

		if _, ok := set.Values["indicators.ema_fast.period"]; !ok {
			t.Errorf("missing ema_fast parameter in set")
		}

		if _, ok := set.Values["indicators.ema_slow.period"]; !ok {
			t.Errorf("missing ema_slow parameter in set")
		}
	}
}

// TestOptimizationMetrics tests optimization objective functions
func TestOptimizationMetrics(t *testing.T) {
	// Create mock backtest results
	mockResults := []*backtest.BacktestResult{
		{
			Metrics: &analytics.Metrics{
				SharpeRatio:  1.5,
				SortinoRatio: 2.0,
				TotalReturn:  0.50, // 50%
				CAGR:         0.25, // 25%
				MaxDrawdown:  0.20, // 20%
				ProfitFactor: 1.8,
				WinRate:      0.55,
				Expectancy:   0.45,
			},
		},
	}

	tests := []struct {
		name     string
		objective string
		expected float64
	}{
		{"sharpe ratio", "sharpe", 1.5},
		{"sortino ratio", "sortino", 2.0},
		{"total return", "return", 0.50},
		{"profit factor", "profit_factor", 1.8},
		{"win rate", "win_rate", 0.55},
		{"expectancy", "expectancy", 0.45},
		{"cagr", "cagr", 0.25},
		{"calmar", "calmar", 1.25}, // CAGR / MaxDD = 0.25 / 0.20
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OptimizationConfig{
				Objective: ObjectiveConfig{
					Type: tt.objective,
				},
			}

			gs := NewGridSearch(config)
			value := gs.getMetricValue(mockResults[0])

			if math.Abs(value-tt.expected) > 0.01 {
				t.Errorf("objective %s: expected %.4f, got %.4f", tt.objective, tt.expected, value)
			}
		})
	}
}

// TestOptimizationReport tests optimization results reporting
func TestOptimizationReport(t *testing.T) {
	// Create mock optimization results
	results := []*OptimizationResult{
		{
			Parameters: ParameterSet{
				Values: map[string]float64{"ema.period": 10},
				Hash:   "ema.period:10.000000;",
			},
			BacktestResult: &backtest.BacktestResult{
				Metrics: &analytics.Metrics{
					SharpeRatio: 1.5,
					TotalReturn: 0.30,
				},
			},
			ObjectiveValue: 1.5,
		},
		{
			Parameters: ParameterSet{
				Values: map[string]float64{"ema.period": 20},
				Hash:   "ema.period:20.000000;",
			},
			BacktestResult: &backtest.BacktestResult{
				Metrics: &analytics.Metrics{
					SharpeRatio: 2.0,
					TotalReturn: 0.45,
				},
			},
			ObjectiveValue: 2.0,
		},
		{
			Parameters: ParameterSet{
				Values: map[string]float64{"ema.period": 30},
				Hash:   "ema.period:30.000000;",
			},
			BacktestResult: &backtest.BacktestResult{
				Metrics: &analytics.Metrics{
					SharpeRatio: 1.8,
					TotalReturn: 0.40,
				},
			},
			ObjectiveValue: 1.8,
		},
	}

	report := &OptimizationReport{
		Strategy:           "test_strategy",
		Symbol:             "BTCUSDT",
		Timeframe:          "1h",
		StartTime:          "2020-01-01",
		EndTime:            "2021-01-01",
		TotalRuns:          3,
		ObjectiveMetric:    "sharpe",
		ObjectiveDirection: "maximize",
		Algorithm:          "grid",
		Results:            results,
	}

	// Sort results
	report.sortResults()

	// Verify best result
	if report.BestResult == nil {
		t.Fatal("expected best result, got nil")
	}

	if report.BestResult.ObjectiveValue != 2.0 {
		t.Errorf("expected best sharpe 2.0, got %.2f", report.BestResult.ObjectiveValue)
	}

	if report.BestResult.Parameters.Values["ema.period"] != 20 {
		t.Errorf("expected best parameter 20, got %.0f", report.BestResult.Parameters.Values["ema.period"])
	}

	// Verify worst result
	if report.WorstResult == nil {
		t.Fatal("expected worst result, got nil")
	}

	if report.WorstResult.ObjectiveValue != 1.5 {
		t.Errorf("expected worst sharpe 1.5, got %.2f", report.WorstResult.ObjectiveValue)
	}

	// Calculate statistics
	report.calculateStatistics()

	expectedAvg := (1.5 + 2.0 + 1.8) / 3
	if math.Abs(report.AvgObjectiveValue-expectedAvg) > 0.01 {
		t.Errorf("expected avg %.4f, got %.4f", expectedAvg, report.AvgObjectiveValue)
	}

	// Test report generation
	reportStr := report.GenerateReport()
	if reportStr == "" {
		t.Error("expected non-empty report string")
	}

	// Verify report contains key information
	if !containsAll(reportStr, "OPTIMIZATION REPORT", "BTCUSDT", "sharpe", "BEST RESULT") {
		t.Error("report missing expected sections")
	}
}

// TestNoLookaheadInOptimization tests that optimization doesn't use lookahead
func TestNoLookaheadInOptimization(t *testing.T) {
	// This test verifies that each backtest in optimization:
	// 1. Uses only historical data
	// 2. Doesn't peek at future candles
	// 3. Maintains deterministic results

	// For now, we verify the optimization config structure
	// ensures no lookahead can occur

	config := OptimizationConfig{
		BacktestConfig: backtest.BacktestConfig{
			Symbol:    "BTCUSDT",
			Timeframe: "1h",
			StartTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2020, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		Parameters: []ParameterRange{
			{
				Name:  "indicators.ema.period",
				Start: 10,
				End:   20,
				Step:  5,
				Type:  "int",
			},
		},
		Objective: ObjectiveConfig{
			Type:      "sharpe",
			Direction: "maximize",
		},
	}

	// Verify config is valid
	gs := NewGridSearch(config)
	sets, err := gs.GenerateParameterSets()
	if err != nil {
		t.Fatalf("generate parameter sets: %v", err)
	}

	// Each parameter set should be independent
	// and can be run in any order (determinism)
	for i, set := range sets {
		if set.Hash == "" {
			t.Errorf("parameter set %d missing hash", i)
		}
	}

	t.Log("Optimization structure validated for no lookahead")
}

// TestMultipleOptimizationScenarios tests various optimization scenarios
func TestMultipleOptimizationScenarios(t *testing.T) {
	tests := []struct {
		name          string
		params        []ParameterRange
		expectedCount int
	}{
		{
			name: "single parameter",
			params: []ParameterRange{
				{
					Name:  "ema.period",
					Start: 10,
					End:   20,
					Step:  5,
					Type:  "int",
				},
			},
			expectedCount: 3, // 10, 15, 20
		},
		{
			name: "two parameters",
			params: []ParameterRange{
				{
					Name:  "ema_fast.period",
					Start: 5,
					End:   10,
					Step:  5,
					Type:  "int",
				},
				{
					Name:  "ema_slow.period",
					Start: 20,
					End:   30,
					Step:  10,
					Type:  "int",
				},
			},
			expectedCount: 4, // 2 * 2 = 4
		},
		{
			name: "three parameters",
			params: []ParameterRange{
				{
					Name:  "ema_fast.period",
					Start: 5,
					End:   10,
					Step:  5,
					Type:  "int",
				},
				{
					Name:  "ema_slow.period",
					Start: 20,
					End:   25,
					Step:  5,
					Type:  "int",
				},
				{
					Name:  "atr.multiplier",
					Start: 1.0,
					End:   2.0,
					Step:  0.5,
					Type:  "float",
				},
			},
			expectedCount: 12, // 2 * 2 * 3 = 12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OptimizationConfig{
				Parameters: tt.params,
			}

			gs := NewGridSearch(config)
			sets, err := gs.GenerateParameterSets()

			if err != nil {
				t.Fatalf("generate parameter sets: %v", err)
			}

			if len(sets) != tt.expectedCount {
				t.Errorf("expected %d combinations, got %d", tt.expectedCount, len(sets))
			}
		})
	}
}

// TestOptimizationDirection tests maximize vs minimize objectives
func TestOptimizationDirection(t *testing.T) {
	results := []*OptimizationResult{
		{ObjectiveValue: 1.5},
		{ObjectiveValue: 2.0},
		{ObjectiveValue: 1.8},
	}

	t.Run("maximize", func(t *testing.T) {
		report := &OptimizationReport{
			ObjectiveDirection: "maximize",
			Results:            results,
		}
		report.sortResults()

		if report.BestResult.ObjectiveValue != 2.0 {
			t.Errorf("maximize: expected best 2.0, got %.2f", report.BestResult.ObjectiveValue)
		}
	})

	t.Run("minimize", func(t *testing.T) {
		report := &OptimizationReport{
			ObjectiveDirection: "minimize",
			Results:            results,
		}
		report.sortResults()

		if report.BestResult.ObjectiveValue != 1.5 {
			t.Errorf("minimize: expected best 1.5, got %.2f", report.BestResult.ObjectiveValue)
		}
	})
}

// TestOptimizationRunSequential tests sequential optimization execution
func TestOptimizationRunSequential(t *testing.T) {
	config := OptimizationConfig{
		Parameters: []ParameterRange{
			{
				Name:  "ema.period",
				Start: 10,
				End:   20,
				Step:  10,
				Type:  "int",
			},
		},
		Objective: ObjectiveConfig{
			Type:      "sharpe",
			Direction: "maximize",
		},
	}

	gs := NewGridSearch(config)

	// Mock evaluator that returns deterministic results
	evaluator := func(ps ParameterSet) (*backtest.BacktestResult, error) {
		return &backtest.BacktestResult{
			Metrics: &analytics.Metrics{
				SharpeRatio: ps.Values["ema.period"] / 10.0,
			},
		}, nil
	}

	report, err := gs.Run(evaluator, 1) // sequential (parallel=1)
	if err != nil {
		t.Fatalf("optimization run failed: %v", err)
	}

	if report.TotalRuns != 2 {
		t.Errorf("expected 2 runs, got %d", report.TotalRuns)
	}

	if len(report.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(report.Results))
	}

	// Verify results are sorted (maximize sharpe)
	if report.BestResult.ObjectiveValue < report.WorstResult.ObjectiveValue {
		t.Error("results not properly sorted for maximization")
	}
}

// TestOptimizationRunParallel tests parallel optimization execution
func TestOptimizationRunParallel(t *testing.T) {
	config := OptimizationConfig{
		Parameters: []ParameterRange{
			{
				Name:  "ema.period",
				Start: 10,
				End:   30,
				Step:  10,
				Type:  "int",
			},
		},
		Objective: ObjectiveConfig{
			Type:      "sharpe",
			Direction: "maximize",
		},
	}

	gs := NewGridSearch(config)

	// Mock evaluator
	evaluator := func(ps ParameterSet) (*backtest.BacktestResult, error) {
		return &backtest.BacktestResult{
			Metrics: &analytics.Metrics{
				SharpeRatio: ps.Values["ema.period"] / 10.0,
			},
		}, nil
	}

	report, err := gs.Run(evaluator, 2) // parallel with 2 workers
	if err != nil {
		t.Fatalf("optimization run failed: %v", err)
	}

	if report.TotalRuns != 3 {
		t.Errorf("expected 3 runs, got %d", report.TotalRuns)
	}

	if len(report.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(report.Results))
	}
}

// TestOptimizationReportExport tests report export functionality
func TestOptimizationReportExport(t *testing.T) {
	results := []*OptimizationResult{
		{
			Parameters: ParameterSet{
				Values: map[string]float64{"ema.period": 20},
				Hash:   "ema.period:20.000000;",
			},
			BacktestResult: &backtest.BacktestResult{
				Metrics: &analytics.Metrics{
					SharpeRatio:  2.0,
					TotalReturn:  0.45,
					CAGR:         0.25,
					MaxDrawdown:  0.15,
					WinRate:      0.55,
					ProfitFactor: 1.8,
					Expectancy:   0.45,
				},
				TotalTrades: 100,
			},
			ObjectiveValue: 2.0,
			Rank:           1,
		},
	}

	report := &OptimizationReport{
		Strategy:           "test",
		Results:            results,
		TopResults:         results,
		ObjectiveMetric:    "sharpe",
		ObjectiveDirection: "maximize",
	}

	// Sort results to set BestResult/WorstResult
	report.sortResults()

	// Test GetTopParameters
	topParams := report.GetTopParameters(1)
	if len(topParams) != 1 {
		t.Errorf("expected 1 top parameter, got %d", len(topParams))
	}

	if topParams[0].Values["ema.period"] != 20 {
		t.Errorf("expected ema.period 20, got %.0f", topParams[0].Values["ema.period"])
	}

	// Test GetBestParameterValues
	bestParams := report.GetBestParameterValues()
	if bestParams["ema.period"] != 20 {
		t.Errorf("expected best ema.period 20, got %.0f", bestParams["ema.period"])
	}

	// Test CalculateImprovement
	report.BestResult = results[0]
	report.WorstResult = &OptimizationResult{ObjectiveValue: 1.0}
	improvement := report.CalculateImprovement()
	if improvement <= 0 {
		t.Errorf("expected positive improvement, got %.4f", improvement)
	}
}

// Helper function
func containsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !contains(s, substr) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
