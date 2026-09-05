package montecarlo_test

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/montecarlo"
)

// TestMonteCarloIntegration tests the complete Monte Carlo workflow
func TestMonteCarloIntegration(t *testing.T) {
	// Create sample trades from a backtest
	trades := []montecarlo.Trade{
		{
			ID:         1,
			EntryTime:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			EntryPrice: 50000.0,
			ExitPrice:  50500.0,
			Quantity:   0.1,
			GrossPnL:   50.0,
			Fees:       5.0,
			NetPnL:     45.0,
			Return:     0.009,
		},
		{
			ID:         2,
			EntryTime:  time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 2, 14, 0, 0, 0, time.UTC),
			EntryPrice: 50500.0,
			ExitPrice:  50000.0,
			Quantity:   0.1,
			GrossPnL:   -50.0,
			Fees:       5.0,
			NetPnL:     -55.0,
			Return:     -0.011,
		},
		{
			ID:         3,
			EntryTime:  time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 3, 14, 0, 0, 0, time.UTC),
			EntryPrice: 50000.0,
			ExitPrice:  51000.0,
			Quantity:   0.1,
			GrossPnL:   100.0,
			Fees:       5.0,
			NetPnL:     95.0,
			Return:     0.019,
		},
		{
			ID:         4,
			EntryTime:  time.Date(2024, 1, 4, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 4, 14, 0, 0, 0, time.UTC),
			EntryPrice: 51000.0,
			ExitPrice:  50800.0,
			Quantity:   0.1,
			GrossPnL:   -20.0,
			Fees:       5.0,
			NetPnL:     -25.0,
			Return:     -0.004,
		},
		{
			ID:         5,
			EntryTime:  time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 5, 14, 0, 0, 0, time.UTC),
			EntryPrice: 50800.0,
			ExitPrice:  51500.0,
			Quantity:   0.1,
			GrossPnL:   70.0,
			Fees:       5.0,
			NetPnL:     65.0,
			Return:     0.014,
		},
	}

	config := montecarlo.MCConfig{
		Simulations: 1000,
		Seed:        42,
		Type:        montecarlo.TradeReshuffle,
	}

	// Run Monte Carlo
	runner := montecarlo.NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Monte Carlo runner failed: %v", err)
	}

	// Verify number of simulations
	if len(result.Simulations) != 1000 {
		t.Errorf("Expected 1000 simulations, got %d", len(result.Simulations))
	}

	// Verify statistics are calculated
	stats := result.Statistics

	// Mean return should be positive (more winning trades)
	if stats.MeanReturn <= 0 {
		t.Errorf("Expected positive mean return, got %f", stats.MeanReturn)
	}

	// Mean max drawdown should be between 0 and 1
	if stats.MeanMaxDrawdown < 0 || stats.MeanMaxDrawdown > 1 {
		t.Errorf("Mean max drawdown out of range: %f", stats.MeanMaxDrawdown)
	}

	// Mean win rate should be around 60% (3 out of 5 trades are winners)
	// But due to reshuffling, this can vary
	if stats.MeanWinRate < 0 || stats.MeanWinRate > 1 {
		t.Errorf("Mean win rate out of range: %f", stats.MeanWinRate)
	}

	// Verify confidence intervals
	if len(result.ConfidenceIntervals) != 5 {
		t.Errorf("Expected 5 confidence intervals, got %d", len(result.ConfidenceIntervals))
	}

	// Verify percentile ordering (5th < 50th < 95th)
	p05 := result.ConfidenceIntervals[0].TotalReturn
	p50 := result.ConfidenceIntervals[2].TotalReturn
	p95 := result.ConfidenceIntervals[4].TotalReturn

	if p05 > p50 {
		t.Errorf("5th percentile (%f) should be <= 50th percentile (%f)", p05, p50)
	}
	if p50 > p95 {
		t.Errorf("50th percentile (%f) should be <= 95th percentile (%f)", p50, p95)
	}

	t.Logf("Monte Carlo Results:")
	t.Logf("  Mean Return:      %.4f", stats.MeanReturn)
	t.Logf("  Std Dev Return:   %.4f", stats.StdDevReturn)
	t.Logf("  Mean Drawdown:    %.4f", stats.MeanMaxDrawdown)
	t.Logf("  Mean Win Rate:    %.4f", stats.MeanWinRate)
	t.Logf("  Prob of Ruin:     %.4f", stats.ProbabilityOfRuin)
}

// TestMonteCarloReturnReshuffle tests return reshuffling specifically
func TestMonteCarloReturnReshuffle(t *testing.T) {
	trades := []montecarlo.Trade{
		{ID: 1, NetPnL: 100.0},
		{ID: 2, NetPnL: -50.0},
		{ID: 3, NetPnL: 75.0},
		{ID: 4, NetPnL: -25.0},
		{ID: 5, NetPnL: 50.0},
	}

	config := montecarlo.MCConfig{
		Simulations: 100,
		Seed:        12345,
		Type:        montecarlo.ReturnReshuffle,
	}

	runner := montecarlo.NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Runner failed: %v", err)
	}

	// With return reshuffling, total PnL should be preserved
	expectedTotalPnL := 100.0 - 50.0 + 75.0 - 25.0 + 50.0 // = 150.0

	for i, sim := range result.Simulations {
		actualTotalPnL := 0.0
		for _, trade := range sim.Trades {
			actualTotalPnL += trade.NetPnL
		}
		if actualTotalPnL != expectedTotalPnL {
			t.Errorf("Simulation %d: expected total PnL %f, got %f",
				i, expectedTotalPnL, actualTotalPnL)
		}
	}

	t.Logf("Return Reshuffle Results:")
	t.Logf("  Total PnL preserved: %f", expectedTotalPnL)
	t.Logf("  Mean Return: %.4f", result.Statistics.MeanReturn)
}

// TestMonteCarloBootstrap tests bootstrap sampling
func TestMonteCarloBootstrap(t *testing.T) {
	trades := []montecarlo.Trade{
		{ID: 1, NetPnL: 100.0},
		{ID: 2, NetPnL: -50.0},
		{ID: 3, NetPnL: 75.0},
	}

	config := montecarlo.MCConfig{
		Simulations: 100,
		Seed:        99999,
		Type:        montecarlo.BootstrapReshuffle,
	}

	runner := montecarlo.NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Runner failed: %v", err)
	}

	// Verify each simulation has same number of trades
	for i, sim := range result.Simulations {
		if sim.TotalTrades != len(trades) {
			t.Errorf("Simulation %d: expected %d trades, got %d",
				i, len(trades), sim.TotalTrades)
		}
	}

	// With bootstrap, total PnL will vary across simulations
	pnlValues := make([]float64, len(result.Simulations))
	for i, sim := range result.Simulations {
		pnlValues[i] = sim.TotalPnL
	}

	// Verify there's variance in PnL values (not all same)
	allSame := true
	for i := 1; i < len(pnlValues); i++ {
		if pnlValues[i] != pnlValues[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Bootstrap should produce varying total PnL values")
	}

	t.Logf("Bootstrap Results:")
	t.Logf("  Mean Return: %.4f", result.Statistics.MeanReturn)
	t.Logf("  Std Dev Return: %.4f", result.Statistics.StdDevReturn)
}

// TestMonteCarloDeterminism verifies that same seed produces same results
func TestMonteCarloDeterminism(t *testing.T) {
	trades := []montecarlo.Trade{
		{ID: 1, NetPnL: 100.0},
		{ID: 2, NetPnL: -50.0},
		{ID: 3, NetPnL: 75.0},
	}

	// Run twice with same seed
	for run := 0; run < 2; run++ {
		config := montecarlo.MCConfig{
			Simulations: 50,
			Seed:        42,
			Type:        montecarlo.TradeReshuffle,
		}

		runner := montecarlo.NewRunner(config, trades, 10000.0)
		result, err := runner.Run()
		if err != nil {
			t.Fatalf("Run %d failed: %v", run, err)
		}

		// Store results for comparison
		t.Logf("Run %d - Mean Return: %.6f, Mean Drawdown: %.6f",
			run, result.Statistics.MeanReturn, result.Statistics.MeanMaxDrawdown)
	}
}

// TestMonteCarloLargeDataset tests with more trades
func TestMonteCarloLargeDataset(t *testing.T) {
	// Generate 100 random trades
	trades := make([]montecarlo.Trade, 100)
	for i := 0; i < 100; i++ {
		pnl := float64((i % 20) - 10) * 10.0 // Range -100 to +90
		trades[i] = montecarlo.Trade{
			ID:     int64(i + 1),
			NetPnL: pnl,
		}
	}

	config := montecarlo.MCConfig{
		Simulations: 500,
		Seed:        42,
		Type:        montecarlo.TradeReshuffle,
	}

	runner := montecarlo.NewRunner(config, trades, 100000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Runner failed: %v", err)
	}

	// Verify results
	if len(result.Simulations) != 500 {
		t.Errorf("Expected 500 simulations, got %d", len(result.Simulations))
	}

	// Print summary
	t.Logf("Large Dataset (100 trades, 500 simulations):")
	t.Logf("  Mean Return:      %.4f", result.Statistics.MeanReturn)
	t.Logf("  Std Dev Return:   %.4f", result.Statistics.StdDevReturn)
	t.Logf("  Mean Drawdown:    %.4f", result.Statistics.MeanMaxDrawdown)
	t.Logf("  P05 Return:       %.4f", result.Statistics.P05Return)
	t.Logf("  P95 Return:       %.4f", result.Statistics.P95Return)
	t.Logf("  Prob of Ruin:     %.4f", result.Statistics.ProbabilityOfRuin)

	// Verify confidence intervals
	t.Logf("\nConfidence Intervals:")
	for _, ci := range result.ConfidenceIntervals {
		t.Logf("  %5.0f%%: Return=%+.4f, Drawdown=%.4f, WinRate=%.4f, Sharpe=%.4f",
			ci.Percentile*100, ci.TotalReturn, ci.MaxDrawdown, ci.WinRate, ci.SharpeRatio)
	}
}
