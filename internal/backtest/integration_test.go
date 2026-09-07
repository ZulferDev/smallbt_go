package backtest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/integration"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestBacktest_WithTransforms(t *testing.T) {
	// Create test CSV data
	csvPath := createTestCSVFile(t)
	defer os.Remove(csvPath)

	// Create test strategy
	strategyPath := createTestStrategy(t)
	defer os.Remove(strategyPath)

	// Configure transforms
	transformConfig := &integration.TransformConfig{
		Enabled: true,
		Transforms: []integration.TransformSpec{
			{
				Type:  "scale",
				Field: "close",
				Params: map[string]interface{}{
					"factor": 2.0,
				},
			},
			{
				Type:  "normalize",
				Field: "volume",
			},
		},
		BatchSize: 100,
	}

	// Run backtest with transforms
	config := BacktestConfig{
		Symbol:          market.Symbol("BTCUSDT"),
		Timeframe:       market.Timeframe("1h"),
		InitialCash:     10000.0,
		StrategyPath:    strategyPath,
		DataPath:        csvPath,
		TransformConfig: transformConfig,
	}

	result, err := Run(config)
	if err != nil {
		t.Fatalf("Backtest with transforms failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify result structure
	if result.Portfolio == nil {
		t.Error("Expected non-nil portfolio")
	}

	if result.Metrics == nil {
		t.Error("Expected non-nil metrics")
	}

	// Data should be transformed (scaled close prices)
	// If close was 100, after scale(2.0) should be 200
	// Strategy should operate on transformed data
	t.Logf("Backtest completed: %d trades", result.TotalTrades)
	t.Logf("Final equity: %.2f", result.Portfolio.Equity)
}

func TestBacktest_WithoutTransforms(t *testing.T) {
	// Create test data
	csvPath := createTestCSVFile(t)
	defer os.Remove(csvPath)

	strategyPath := createTestStrategy(t)
	defer os.Remove(strategyPath)

	// Run backtest WITHOUT transforms (backward compatible)
	config := BacktestConfig{
		Symbol:          market.Symbol("BTCUSDT"),
		Timeframe:       market.Timeframe("1h"),
		InitialCash:     10000.0,
		StrategyPath:    strategyPath,
		DataPath:        csvPath,
		TransformConfig: nil, // No transforms
	}

	result, err := Run(config)
	if err != nil {
		t.Fatalf("Backtest without transforms failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	t.Logf("Backtest (no transforms) completed: %d trades", result.TotalTrades)
}

func TestBacktest_DisabledTransforms(t *testing.T) {
	csvPath := createTestCSVFile(t)
	defer os.Remove(csvPath)

	strategyPath := createTestStrategy(t)
	defer os.Remove(strategyPath)

	// Disabled transforms should behave like no transforms
	transformConfig := &integration.TransformConfig{
		Enabled: false,
		Transforms: []integration.TransformSpec{
			{Type: "scale", Field: "close"},
		},
	}

	config := BacktestConfig{
		Symbol:          market.Symbol("BTCUSDT"),
		Timeframe:       market.Timeframe("1h"),
		InitialCash:     10000.0,
		StrategyPath:    strategyPath,
		DataPath:        csvPath,
		TransformConfig: transformConfig,
	}

	result, err := Run(config)
	if err != nil {
		t.Fatalf("Backtest with disabled transforms failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	t.Logf("Backtest (disabled transforms) completed: %d trades", result.TotalTrades)
}

func TestBacktest_InvalidTransformConfig(t *testing.T) {
	csvPath := createTestCSVFile(t)
	defer os.Remove(csvPath)

	strategyPath := createTestStrategy(t)
	defer os.Remove(strategyPath)

	// Invalid config: missing required field
	transformConfig := &integration.TransformConfig{
		Enabled: true,
		Transforms: []integration.TransformSpec{
			{
				Type: "scale", // Missing 'field'
				Params: map[string]interface{}{
					"factor": 2.0,
				},
			},
		},
	}

	config := BacktestConfig{
		Symbol:          market.Symbol("BTCUSDT"),
		Timeframe:       market.Timeframe("1h"),
		InitialCash:     10000.0,
		StrategyPath:    strategyPath,
		DataPath:        csvPath,
		TransformConfig: transformConfig,
	}

	_, err := Run(config)
	if err == nil {
		t.Error("Expected error for invalid transform config")
	}
}

// Helper functions

func createTestCSVFile(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test_data.csv")

	file, err := os.Create(csvPath)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer file.Close()

	// Write CSV header
	_, _ = file.WriteString("timestamp,open,high,low,close,volume\n")

	// Write 100 candles
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 100; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Hour)
		price := 45000.0 + float64(i*100)
		high := price + 100
		low := price - 100
		volume := 1000000.0 + float64(i*10000)

		_, _ = file.WriteString(
			ts.Format("2006-01-02 15:04:05") + "," +
				formatFloat(price) + "," +
				formatFloat(high) + "," +
				formatFloat(low) + "," +
				formatFloat(price+50) + "," +
				formatFloat(volume) + "\n",
		)
	}

	return csvPath
}

func createTestStrategy(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	strategyPath := filepath.Join(tmpDir, "test_strategy.yaml")

	strategyYAML := `
strategy:
  name: test_strategy
  version: "1.0"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  sma_fast:
    type: sma
    source: close
    period: 10

  sma_slow:
    type: sma
    source: close
    period: 20

entry:
  long:
    cross_above: [sma_fast, sma_slow]

exit:
  long:
    cross_below: [sma_fast, sma_slow]

risk:
  position_size:
    type: percent_equity
    value: 0.1
`

	err := os.WriteFile(strategyPath, []byte(strategyYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test strategy: %v", err)
	}

	return strategyPath
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
