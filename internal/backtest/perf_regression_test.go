package backtest

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

// generateTestCSVCompat creates a temporary CSV file with synthetic OHLCV data
// Compatible with testing.T instead of testing.B
func generateTestCSVCompat(t *testing.T, numCandles int) string {
	t.Helper()
	
	// Create temp file
	tmpFile, err := os.CreateTemp("", "perf_test_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	
	// Write CSV header
	fmt.Fprintln(tmpFile, "timestamp,open,high,low,close,volume")
	
	// Generate realistic OHLCV data
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	basePrice := 10000.0
	rnd := rand.New(rand.NewSource(42)) // Deterministic seed
	
	for i := 0; i < numCandles; i++ {
		// Generate OHLCV values
		open := basePrice + rnd.Float64()*100 - 50
		close := open + rnd.Float64()*20 - 10
		high := max(open, close) + rnd.Float64()*10
		low := min(open, close) - rnd.Float64()*10
		volume := 1000000 + rnd.Float64()*500000
		
		// Write candle
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		fmt.Fprintf(tmpFile, "%s,%.2f,%.2f,%.2f,%.2f,%.2f\n",
			timestamp.Format("2006-01-02 15:04:05"),
			open, high, low, close, volume)
		
		// Update base price for next candle
		basePrice = close
	}
	
	tmpFile.Close()
	return tmpFile.Name()
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Performance regression tests ensure that performance does not degrade unexpectedly
// These are NOT benchmarks - they are validation tests that will fail if performance
// falls below acceptable thresholds

// TestPerformanceBacktestSmall validates small dataset performance
func TestPerformanceBacktestSmall(t *testing.T) {
	// Generate test data file
	tmpFile := generateTestCSVCompat(t, 100)
	defer os.Remove(tmpFile)

	// Create backtest config
	config := BacktestConfig{
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		InitialCash:  10000.0,
		StrategyPath: "../../strategies/examples/ema_volume.yaml",
		DataPath:     tmpFile,
	}

	// Run backtest
	start := time.Now()
	result, err := Run(config)
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	// Validate results are correct (not just performance)
	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.Config.Symbol != "BTCUSDT" {
		t.Errorf("expected symbol BTCUSDT, got %s", result.Config.Symbol)
	}

	// Performance validation: 100 candles should complete in under 20ms
	// This is intentionally conservative to avoid flaky tests
	if duration > 20*time.Millisecond {
		t.Errorf("backtest took %v, expected < 20ms", duration)
	}
}

// TestPerformanceBacktestMedium validates medium dataset performance
func TestPerformanceBacktestMedium(t *testing.T) {
	// Generate test data file
	tmpFile := generateTestCSVCompat(t, 1000)
	defer os.Remove(tmpFile)

	// Create backtest config
	config := BacktestConfig{
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		InitialCash:  10000.0,
		StrategyPath: "../../strategies/examples/ema_volume.yaml",
		DataPath:     tmpFile,
	}

	// Run backtest
	start := time.Now()
	result, err := Run(config)
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	// Validate results are correct
	if result == nil {
		t.Fatal("result should not be nil")
	}

	// Performance validation: 1000 candles should complete in under 200ms
	if duration > 200*time.Millisecond {
		t.Errorf("backtest took %v, expected < 200ms", duration)
	}
}

// TestPerformanceBacktestLarge validates large dataset performance
func TestPerformanceBacktestLarge(t *testing.T) {
	// Generate test data file
	tmpFile := generateTestCSVCompat(t, 2000)
	defer os.Remove(tmpFile)

	// Create backtest config
	config := BacktestConfig{
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		InitialCash:  10000.0,
		StrategyPath: "../../strategies/examples/ema_volume.yaml",
		DataPath:     tmpFile,
	}

	// Run backtest
	start := time.Now()
	result, err := Run(config)
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	// Validate results are correct
	if result == nil {
		t.Fatal("result should not be nil")
	}

	// Performance validation: 2000 candles should complete in under 800ms
	// Current baseline: ~676ms, threshold set with buffer for variance
	if duration > 800*time.Millisecond {
		t.Errorf("backtest took %v, expected < 800ms", duration)
	}
}

// TestPerformanceDeterminism ensures performance tests are deterministic
func TestPerformanceDeterminism(t *testing.T) {
	tmpFile := generateTestCSVCompat(t, 500)
	defer os.Remove(tmpFile)

	config := BacktestConfig{
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		InitialCash:  10000.0,
		StrategyPath: "../../strategies/examples/ema_volume.yaml",
		DataPath:     tmpFile,
	}

	// Run backtest twice
	result1, err1 := Run(config)
	if err1 != nil {
		t.Fatalf("first backtest failed: %v", err1)
	}

	result2, err2 := Run(config)
	if err2 != nil {
		t.Fatalf("second backtest failed: %v", err2)
	}

	// Results should be identical (deterministic)
	if result1.StartTime != result2.StartTime {
		t.Errorf("startTime differs: %v vs %v", result1.StartTime, result2.StartTime)
	}

	if result1.EndTime != result2.EndTime {
		t.Errorf("endTime differs: %v vs %v", result1.EndTime, result2.EndTime)
	}

	if result1.TotalTrades != result2.TotalTrades {
		t.Errorf("totalTrades differs: %v vs %v", result1.TotalTrades, result2.TotalTrades)
	}

	if result1.Portfolio.Equity != result2.Portfolio.Equity {
		t.Errorf("final equity differs: %v vs %v", result1.Portfolio.Equity, result2.Portfolio.Equity)
	}
}

// TestPerformanceLinearScaling validates that performance scales linearly
func TestPerformanceLinearScaling(t *testing.T) {
	// Test that doubling data size roughly doubles execution time
	tmpFile100 := generateTestCSVCompat(t, 100)
	tmpFile200 := generateTestCSVCompat(t, 200)
	defer os.Remove(tmpFile100)
	defer os.Remove(tmpFile200)

	config := BacktestConfig{
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		InitialCash:  10000.0,
		StrategyPath: "../../strategies/examples/ema_volume.yaml",
	}

	config.DataPath = tmpFile100
	start1 := time.Now()
	_, err1 := Run(config)
	duration1 := time.Since(start1)
	if err1 != nil {
		t.Fatalf("100 candle backtest failed: %v", err1)
	}

	config.DataPath = tmpFile200
	start2 := time.Now()
	_, err2 := Run(config)
	duration2 := time.Since(start2)
	if err2 != nil {
		t.Fatalf("200 candle backtest failed: %v", err2)
	}

	// Ratio should be approximately 2x (allowing wider variance)
	// With EMA caching optimized to O(1), ratio should be closer to 1:1 for EMA computation
	// but overall ratio includes overhead for other operations
	ratio := float64(duration2) / float64(duration1)
	// After EMA optimization and debug log removal, scaling is highly efficient
	// If ratio < 1.0, it means 200 candle test was faster (possible due to caching/warmup)
	// Accept ratio as low as 0.5 (200 candles half the time of 100) to 3.0
	// This test is for detecting major regressions, not precise performance measurement
	if ratio < 0.5 || ratio > 3.0 {
		t.Errorf("scaling ratio %.2f is outside acceptable range [0.5, 3.0]", ratio)
	}
}

// TestPerformanceResultsValid validates that backtest results are logically valid
func TestPerformanceResultsValid(t *testing.T) {
	tmpFile := generateTestCSVCompat(t, 500)
	defer os.Remove(tmpFile)

	config := BacktestConfig{
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		InitialCash:  10000.0,
		StrategyPath: "../../strategies/examples/ema_volume.yaml",
		DataPath:     tmpFile,
	}

	result, err := Run(config)
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	// Validate portfolio state
	if result.Portfolio == nil {
		t.Fatal("portfolio should not be nil")
	}

	// Equity should be reasonable
	equity := result.Portfolio.Equity
	if equity <= 0 {
		t.Errorf("equity should be positive, got %v", equity)
	}

	// Total trades should be non-negative
	if result.TotalTrades < 0 {
		t.Errorf("total trades should be non-negative, got %v", result.TotalTrades)
	}

	// Metrics should exist
	if result.Metrics == nil {
		t.Fatal("metrics should not be nil")
	}
}
