package montecarlo

import (
	"math"
	"testing"
	"time"
)

// TestTradeReshuffle tests the trade reshuffling algorithm
func TestTradeReshuffle(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100},
		{ID: 2, NetPnL: -50},
		{ID: 3, NetPnL: 75},
		{ID: 4, NetPnL: -25},
	}

	reshuffler := NewReshuffler(42) // deterministic seed

	// Test 1: Shuffled trades should have same length
	shuffled := reshuffler.ShuffleTrades(trades)
	if len(shuffled) != len(trades) {
		t.Errorf("expected %d trades, got %d", len(trades), len(shuffled))
	}

	// Test 2: Shuffled trades should contain same IDs
	idMap := make(map[int64]bool)
	for _, trade := range shuffled {
		idMap[trade.ID] = true
	}
	for _, trade := range trades {
		if !idMap[trade.ID] {
			t.Errorf("trade ID %d missing from shuffled result", trade.ID)
		}
	}

	// Test 3: Multiple shuffles should produce different orders
	shuffled2 := reshuffler.ShuffleTrades(trades)
	allSame := true
	for i := range shuffled {
		if shuffled[i].ID != shuffled2[i].ID {
			allSame = false
			break
		}
	}
	if allSame && len(trades) > 1 {
		// This is statistically unlikely with a good random source
		t.Log("Warning: two consecutive shuffles produced identical order")
	}
}

// TestReturnReshuffle tests the return reshuffling algorithm
func TestReturnReshuffle(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100, Fees: 1},
		{ID: 2, NetPnL: -50, Fees: 1},
		{ID: 3, NetPnL: 75, Fees: 1},
	}

	reshuffler := NewReshuffler(42)
	shuffled := reshuffler.ShuffleReturns(trades)

	// Test 1: Same length
	if len(shuffled) != len(trades) {
		t.Errorf("expected %d trades, got %d", len(trades), len(shuffled))
	}

	// Test 2: NetPnL values should be permuted
	pnlSum := 0.0
	for _, trade := range shuffled {
		pnlSum += trade.NetPnL
	}
	expectedSum := 100.0 - 50.0 + 75.0
	if pnlSum != expectedSum {
		t.Errorf("expected total PnL %f, got %f", expectedSum, pnlSum)
	}
}

// TestBootstrapTrades tests the bootstrap sampling algorithm
func TestBootstrapTrades(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100},
		{ID: 2, NetPnL: -50},
		{ID: 3, NetPnL: 75},
	}

	reshuffler := NewReshuffler(42)
	sampled := reshuffler.BootstrapTrades(trades, 100)

	// Test 1: Sample size should match requested
	if len(sampled) != 100 {
		t.Errorf("expected 100 trades, got %d", len(sampled))
	}

	// Test 2: All sampled trades should come from original set
	idMap := make(map[int64]bool)
	for _, trade := range trades {
		idMap[trade.ID] = true
	}
	for _, trade := range sampled {
		if !idMap[trade.ID] {
			t.Errorf("sampled trade ID %d not in original set", trade.ID)
		}
	}
}

// TestBuildEquityCurve tests equity curve construction
func TestBuildEquityCurve(t *testing.T) {
	initialCapital := 10000.0
	trades := []Trade{
		{NetPnL: 100},
		{NetPnL: -50},
		{NetPnL: 75},
	}

	curve := BuildEquityCurve(initialCapital, trades)

	// Test 1: Length should be trades + 1
	if len(curve) != len(trades)+1 {
		t.Errorf("expected %d points, got %d", len(trades)+1, len(curve))
	}

	// Test 2: Initial equity
	if curve[0] != initialCapital {
		t.Errorf("expected initial equity %f, got %f", initialCapital, curve[0])
	}

	// Test 3: Final equity
	expectedFinal := initialCapital + 100 - 50 + 75
	if curve[len(curve)-1] != expectedFinal {
		t.Errorf("expected final equity %f, got %f", expectedFinal, curve[len(curve)-1])
	}
}

// TestCalculateDrawdown tests drawdown calculation
func TestCalculateDrawdown(t *testing.T) {
	tests := []struct {
		name     string
		curve    []float64
		expected float64
	}{
		{
			name:     "empty curve",
			curve:    []float64{},
			expected: 0.0,
		},
		{
			name:     "no drawdown",
			curve:    []float64{100, 110, 120, 130},
			expected: 0.0,
		},
		{
			name:     "simple drawdown",
			curve:    []float64{100, 90, 80, 90},
			expected: 0.2, // (100-80)/100
		},
		{
			name:     "partial recovery",
			curve:    []float64{100, 120, 100, 110},
			expected: 1.0 / 6.0, // (120-100)/120
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDrawdown(tt.curve)
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected drawdown %f, got %f", tt.expected, result)
			}
		})
	}
}

// TestCalculateWinRate tests win rate calculation
func TestCalculateWinRate(t *testing.T) {
	tests := []struct {
		name     string
		trades   []Trade
		expected float64
	}{
		{
			name:     "empty trades",
			trades:   []Trade{},
			expected: 0.0,
		},
		{
			name: "50% win rate",
			trades: []Trade{
				{NetPnL: 100},
				{NetPnL: -50},
			},
			expected: 0.5,
		},
		{
			name: "75% win rate",
			trades: []Trade{
				{NetPnL: 100},
				{NetPnL: 50},
				{NetPnL: 75},
				{NetPnL: -25},
			},
			expected: 0.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateWinRate(tt.trades)
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected win rate %f, got %f", tt.expected, result)
			}
		})
	}
}

// TestCalculatePercentile tests percentile calculation
func TestCalculatePercentile(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		percentile float64
		expected   float64
	}{
		{0.0, 1.0},
		{0.5, 5.5},
		{1.0, 10.0},
		{0.25, 3.25},
		{0.75, 7.75},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := CalculatePercentile(values, tt.percentile)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("percentile %f: expected %f, got %f", tt.percentile, tt.expected, result)
			}
		})
	}
}

// TestMCRunner tests the main Monte Carlo runner
func TestMCRunner(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100, Fees: 1, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour)},
		{ID: 2, NetPnL: -50, Fees: 1, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour)},
		{ID: 3, NetPnL: 75, Fees: 1, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour)},
		{ID: 4, NetPnL: -25, Fees: 1, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour)},
	}

	config := MCConfig{
		Simulations: 100,
		Seed:        42,
		Type:        TradeReshuffle,
	}

	runner := NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	// Test 1: Correct number of simulations
	if len(result.Simulations) != 100 {
		t.Errorf("expected 100 simulations, got %d", len(result.Simulations))
	}

	// Test 2: Statistics should be calculated
	if result.Statistics.MeanReturn == 0 {
		t.Error("mean return should not be zero")
	}

	// Test 3: Confidence intervals should exist
	if len(result.ConfidenceIntervals) != 5 {
		t.Errorf("expected 5 confidence intervals, got %d", len(result.ConfidenceIntervals))
	}

	// Test 4: Probability of ruin should be calculated
	if result.Statistics.ProbabilityOfRuin < 0 || result.Statistics.ProbabilityOfRuin > 1 {
		t.Errorf("probability of ruin should be between 0 and 1, got %f", result.Statistics.ProbabilityOfRuin)
	}
}

// TestDeterministicSeeding tests that same seed produces same results
func TestDeterministicSeeding(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100},
		{ID: 2, NetPnL: -50},
		{ID: 3, NetPnL: 75},
	}

	config := MCConfig{
		Simulations: 10,
		Seed:        12345,
		Type:        TradeReshuffle,
	}

	runner1 := NewRunner(config, trades, 10000.0)
	result1, err := runner1.Run()
	if err != nil {
		t.Fatalf("runner1 failed: %v", err)
	}

	runner2 := NewRunner(config, trades, 10000.0)
	result2, err := runner2.Run()
	if err != nil {
		t.Fatalf("runner2 failed: %v", err)
	}

	// Test: Results should be identical with same seed
	if result1.Statistics.MeanReturn != result2.Statistics.MeanReturn {
		t.Errorf("deterministic seed failed: %f != %f",
			result1.Statistics.MeanReturn, result2.Statistics.MeanReturn)
	}

	if result1.Statistics.MeanMaxDrawdown != result2.Statistics.MeanMaxDrawdown {
		t.Errorf("deterministic drawdown failed: %f != %f",
			result1.Statistics.MeanMaxDrawdown, result2.Statistics.MeanMaxDrawdown)
	}
}

// TestProbabilityOfRuin tests probability of ruin calculation
func TestProbabilityOfRuin(t *testing.T) {
	// Create trades with mixed outcomes
	trades := []Trade{
		{NetPnL: -1000}, // big loss
		{NetPnL: -500},
		{NetPnL: 200},
		{NetPnL: -300},
	}

	config := MCConfig{
		Simulations: 100,
		Seed:        42,
		Type:        TradeReshuffle,
	}

	runner := NewRunner(config, trades, 1000.0) // small capital relative to losses
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	// With small capital and large losses, probability of ruin should be significant
	if result.Statistics.ProbabilityOfRuin < 0 || result.Statistics.ProbabilityOfRuin > 1 {
		t.Errorf("probability of ruin out of range: %f", result.Statistics.ProbabilityOfRuin)
	}
}

// TestExportToJSON tests JSON export
func TestExportToJSON(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100},
		{ID: 2, NetPnL: -50},
	}

	config := MCConfig{
		Simulations: 10,
		Seed:        42,
		Type:        TradeReshuffle,
	}

	runner := NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	// Test JSON export (just verify it doesn't error)
	err = result.ExportToJSON("/tmp/test_mc_result.json")
	if err != nil {
		t.Errorf("JSON export failed: %v", err)
	}
}

// TestExportToCSV tests CSV export
func TestExportToCSV(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100},
		{ID: 2, NetPnL: -50},
	}

	config := MCConfig{
		Simulations: 10,
		Seed:        42,
		Type:        TradeReshuffle,
	}

	runner := NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	// Test CSV export (just verify it doesn't error)
	err = result.ExportToCSV("/tmp/test_mc_result.csv")
	if err != nil {
		t.Errorf("CSV export failed: %v", err)
	}
}

// TestExportToText tests text export
func TestExportToText(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100},
		{ID: 2, NetPnL: -50},
	}

	config := MCConfig{
		Simulations: 10,
		Seed:        42,
		Type:        TradeReshuffle,
	}

	runner := NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	// Test text export (just verify it returns non-empty string)
	text := result.ExportToText()
	if len(text) == 0 {
		t.Error("text export returned empty string")
	}
}

// TestMCAnalysisTypeString tests string representation
func TestMCAnalysisTypeString(t *testing.T) {
	tests := []struct {
		t        MCAnalysisType
		expected string
	}{
		{TradeReshuffle, "TradeReshuffle"},
		{ReturnReshuffle, "ReturnReshuffle"},
		{BootstrapReshuffle, "BootstrapReshuffle"},
		{MCAnalysisType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.t.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.t.String())
			}
		})
	}
}

// TestMCConfigValidate tests configuration validation
func TestMCConfigValidate(t *testing.T) {
	// Invalid: zero simulations
	config := MCConfig{Simulations: 0}
	err := config.Validate()
	if err == nil {
		t.Error("expected error for zero simulations")
	}

	// Valid: positive simulations
	config = MCConfig{Simulations: 100}
	err = config.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestEmptyTrades tests behavior with empty trade list
func TestEmptyTrades(t *testing.T) {
	config := MCConfig{
		Simulations: 10,
		Seed:        42,
		Type:        TradeReshuffle,
	}

	runner := NewRunner(config, []Trade{}, 10000.0)
	_, err := runner.Run()
	if err == nil {
		t.Error("expected error for empty trades")
	}
}

// TestReturnReshuffleType tests Monte Carlo with return reshuffling
func TestReturnReshuffleType(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100},
		{ID: 2, NetPnL: -50},
		{ID: 3, NetPnL: 75},
	}

	config := MCConfig{
		Simulations: 50,
		Seed:        42,
		Type:        ReturnReshuffle,
	}

	runner := NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	// Verify simulations ran
	if len(result.Simulations) != 50 {
		t.Errorf("expected 50 simulations, got %d", len(result.Simulations))
	}

	// Total PnL should be preserved across all simulations
	expectedTotalPnL := 100.0 - 50.0 + 75.0
	for i, sim := range result.Simulations {
		if math.Abs(sim.TotalPnL-expectedTotalPnL) > 0.01 {
			t.Errorf("simulation %d: expected total PnL %f, got %f",
				i, expectedTotalPnL, sim.TotalPnL)
		}
	}
}

// TestBootstrapReshuffleType tests Monte Carlo with bootstrap sampling
func TestBootstrapReshuffleType(t *testing.T) {
	trades := []Trade{
		{ID: 1, NetPnL: 100},
		{ID: 2, NetPnL: -50},
		{ID: 3, NetPnL: 75},
	}

	config := MCConfig{
		Simulations: 50,
		Seed:        42,
		Type:        BootstrapReshuffle,
	}

	runner := NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	// Verify simulations ran
	if len(result.Simulations) != 50 {
		t.Errorf("expected 50 simulations, got %d", len(result.Simulations))
	}

	// Each simulation should have same number of trades as input
	for i, sim := range result.Simulations {
		if sim.TotalTrades != len(trades) {
			t.Errorf("simulation %d: expected %d trades, got %d",
				i, len(trades), sim.TotalTrades)
		}
	}
}
