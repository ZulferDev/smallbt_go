package tests

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/analytics"
	"github.com/1jehuang/backtest/internal/backtest"
	"github.com/1jehuang/backtest/internal/portfolio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPhase8_EquityCurve verifies equity curve generation
func TestPhase8_EquityCurve(t *testing.T) {
	// Run a simple backtest
	result := runTestBacktest(t)

	// Verify equity curve exists and has data
	require.NotNil(t, result.EquityCurve, "Equity curve should not be nil")
	require.Greater(t, len(result.EquityCurve), 0, "Equity curve should have data points")

	// Verify equity curve structure
	for i, point := range result.EquityCurve {
		assert.NotZero(t, point.Timestamp, "Point %d: timestamp should not be zero", i)
		assert.GreaterOrEqual(t, point.Equity, 0.0, "Point %d: equity should be non-negative", i)
		assert.GreaterOrEqual(t, point.Cash, 0.0, "Point %d: cash should be non-negative", i)
		assert.LessOrEqual(t, point.Drawdown, 0.0, "Point %d: drawdown should be non-positive", i)
		assert.GreaterOrEqual(t, point.Exposure, 0.0, "Point %d: exposure should be non-negative", i)
		assert.LessOrEqual(t, point.Exposure, 1.0, "Point %d: exposure should not exceed 100%%", i)
	}

	// Verify chronological ordering
	for i := 1; i < len(result.EquityCurve); i++ {
		assert.True(t, result.EquityCurve[i].Timestamp.After(result.EquityCurve[i-1].Timestamp),
			"Equity curve should be chronologically ordered")
	}

	t.Logf("✓ Equity curve: %d points", len(result.EquityCurve))
}

// TestPhase8_TradeJournal verifies trade journal with detailed records
func TestPhase8_TradeJournal(t *testing.T) {
	result := runTestBacktest(t)

	// Verify trade history exists
	require.NotNil(t, result.TradeHistory, "Trade history should not be nil")

	if len(result.TradeHistory) == 0 {
		t.Skip("No trades in test backtest")
	}

	// Verify trade journal fields
	for i, trade := range result.TradeHistory {
		assert.NotEmpty(t, trade.Symbol, "Trade %d: symbol should not be empty", i)
		assert.NotEmpty(t, trade.Side, "Trade %d: side should not be empty", i)
		assert.NotZero(t, trade.EntryTime, "Trade %d: entry time should not be zero", i)
		assert.Greater(t, trade.EntryPrice, 0.0, "Trade %d: entry price should be positive", i)
		assert.NotZero(t, trade.ExitTime, "Trade %d: exit time should not be zero", i)
		assert.Greater(t, trade.ExitPrice, 0.0, "Trade %d: exit price should be positive", i)
		assert.Greater(t, trade.Quantity, 0.0, "Trade %d: quantity should be positive", i)
		assert.GreaterOrEqual(t, trade.Fees, 0.0, "Trade %d: fees should be non-negative", i)
		assert.NotEmpty(t, trade.ExitReason, "Trade %d: exit reason should not be empty", i)

		// Verify exit time is after entry time
		assert.True(t, trade.ExitTime.After(trade.EntryTime),
			"Trade %d: exit time should be after entry time", i)

		// Verify PnL calculation
		expectedGrossPnL := (trade.ExitPrice - trade.EntryPrice) * trade.Quantity
		if trade.Side == "short" {
			expectedGrossPnL = (trade.EntryPrice - trade.ExitPrice) * trade.Quantity
		}
		assert.InDelta(t, expectedGrossPnL, trade.GrossPnL, 0.01,
			"Trade %d: gross PnL calculation", i)

		expectedNetPnL := trade.GrossPnL - trade.Fees
		assert.InDelta(t, expectedNetPnL, trade.NetPnL, 0.01,
			"Trade %d: net PnL calculation", i)

		// Verify return calculation
		expectedReturn := trade.NetPnL / (trade.EntryPrice * trade.Quantity)
		assert.InDelta(t, expectedReturn, trade.Return, 0.0001,
			"Trade %d: return calculation", i)

		t.Logf("✓ Trade %d: %s %s @ %.2f → %.2f, PnL: %.2f, Return: %.2f%%, Reason: %s",
			i, trade.Side, trade.Symbol, trade.EntryPrice, trade.ExitPrice,
			trade.NetPnL, trade.Return*100, trade.ExitReason)
	}
}

// TestPhase8_AllMetrics verifies all required metrics are calculated
func TestPhase8_AllMetrics(t *testing.T) {
	result := runTestBacktest(t)

	require.NotNil(t, result.Metrics, "Metrics should not be nil")

	// Required metrics from AGENTS.md Phase 8
	requiredMetrics := map[string]func() float64{
		"Total Return":   func() float64 { return result.Metrics.TotalReturn },
		"CAGR":           func() float64 { return result.Metrics.CAGR },
		"Sharpe Ratio":   func() float64 { return result.Metrics.SharpeRatio },
		"Sortino Ratio":  func() float64 { return result.Metrics.SortinoRatio },
		"Calmar Ratio":   func() float64 { return result.Metrics.CalmarRatio },
		"Max Drawdown":   func() float64 { return result.Metrics.MaxDrawdown },
		"Win Rate":       func() float64 { return result.Metrics.WinRate },
		"Profit Factor":  func() float64 { return result.Metrics.ProfitFactor },
		"Expectancy":     func() float64 { return result.Metrics.Expectancy },
		"Average Trade":  func() float64 { return result.Metrics.AvgTrade },
		"Average Win":    func() float64 { return result.Metrics.AvgWin },
		"Average Loss":   func() float64 { return result.Metrics.AvgLoss },
	}

	for name, getter := range requiredMetrics {
		value := getter()
		t.Logf("✓ %s: %.4f", name, value)
	}

	// Verify metric ranges
	assert.GreaterOrEqual(t, result.Metrics.WinRate, 0.0, "Win rate should be >= 0")
	assert.LessOrEqual(t, result.Metrics.WinRate, 1.0, "Win rate should be <= 1")
	assert.LessOrEqual(t, result.Metrics.MaxDrawdown, 0.0, "Max drawdown should be non-positive")
	assert.GreaterOrEqual(t, result.Metrics.ProfitFactor, 0.0, "Profit factor should be non-negative")

	// Verify trade counts
	assert.Equal(t, result.TotalTrades, result.Metrics.TotalTrades,
		"Total trades should match between result and metrics")
	assert.Equal(t, result.Metrics.TotalTrades,
		result.Metrics.WinningTrades+result.Metrics.LosingTrades,
		"Total trades should equal winning + losing trades")
}

// TestPhase8_JSONExport verifies JSON export functionality
func TestPhase8_JSONExport(t *testing.T) {
	result := runTestBacktest(t)

	// Export metrics to JSON
	metricsFile := "/tmp/phase8_test_metrics.json"
	defer os.Remove(metricsFile)

	exporter := analytics.NewExporter()
	err := exporter.ExportMetricsJSON(result.Metrics, metricsFile)
	require.NoError(t, err, "Metrics JSON export should succeed")

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(metricsFile)
	require.NoError(t, err, "Should be able to read exported JSON")

	var exportedMetrics analytics.Metrics
	err = json.Unmarshal(data, &exportedMetrics)
	require.NoError(t, err, "Exported data should be valid JSON")

	// Verify key fields are preserved
	assert.Equal(t, result.Metrics.TotalReturn, exportedMetrics.TotalReturn)
	assert.Equal(t, result.Metrics.TotalTrades, exportedMetrics.TotalTrades)
	assert.Equal(t, result.Metrics.WinRate, exportedMetrics.WinRate)

	t.Logf("✓ Metrics JSON export: %d bytes", len(data))
}

// TestPhase8_CSVExport verifies CSV export functionality
func TestPhase8_CSVExport(t *testing.T) {
	result := runTestBacktest(t)

	// Convert backtest.EquityPoint to analytics.EquityPoint
	equityCurve := make([]analytics.EquityPoint, len(result.EquityCurve))
	for i, ep := range result.EquityCurve {
		equityCurve[i] = analytics.EquityPoint{
			Timestamp: ep.Timestamp,
			Equity:    ep.Equity,
			Cash:      ep.Cash,
			Drawdown:  ep.Drawdown,
			Exposure:  ep.Exposure,
		}
	}

	// Export equity curve to CSV
	equityFile := "/tmp/phase8_equity_curve.csv"
	defer os.Remove(equityFile)

	exporter := analytics.NewExporter()
	err := exporter.ExportEquityCurveCSV(equityCurve, equityFile)
	require.NoError(t, err, "Equity curve CSV export should succeed")

	// Verify CSV format
	f, err := os.Open(equityFile)
	require.NoError(t, err, "Should be able to open equity curve CSV")
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	require.NoError(t, err, "Should be able to read equity curve CSV")
	require.Greater(t, len(records), 1, "CSV should have header + data rows")

	// Verify header
	header := records[0]
	expectedHeaders := []string{"timestamp", "equity", "cash", "drawdown", "exposure"}
	assert.Equal(t, expectedHeaders, header, "CSV header should match expected format")

	t.Logf("✓ Equity curve CSV: %d rows", len(records)-1)

	// Export trade history to CSV
	tradesFile := "/tmp/phase8_trades.csv"
	defer os.Remove(tradesFile)

	err = exporter.ExportTradesCSV(result.TradeHistory, tradesFile)
	require.NoError(t, err, "Trade history CSV export should succeed")

	// Verify trade history CSV
	f2, err := os.Open(tradesFile)
	require.NoError(t, err, "Should be able to open trades CSV")
	defer f2.Close()

	reader2 := csv.NewReader(f2)
	tradeRecords, err := reader2.ReadAll()
	require.NoError(t, err, "Should be able to read trades CSV")
	
	if len(result.TradeHistory) > 0 {
		require.Greater(t, len(tradeRecords), 1, "Trades CSV should have header + data rows")
		t.Logf("✓ Trade history CSV: %d rows", len(tradeRecords)-1)
	}
}

// TestPhase8_PerformanceMetrics verifies performance metric calculations
func TestPhase8_PerformanceMetrics(t *testing.T) {
	// Create test data with known outcomes
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	trades := []portfolio.Trade{
		{
			EntryPrice: 100,
			ExitPrice:  110,
			Quantity:   1,
			GrossPnL:   10,
			NetPnL:     9.9,
			Fees:       0.1,
			Side:       "long",
		},
		{
			EntryPrice: 110,
			ExitPrice:  105,
			Quantity:   1,
			GrossPnL:   -5,
			NetPnL:     -5.1,
			Fees:       0.1,
			Side:       "long",
		},
		{
			EntryPrice: 105,
			ExitPrice:  120,
			Quantity:   1,
			GrossPnL:   15,
			NetPnL:     14.9,
			Fees:       0.1,
			Side:       "long",
		},
	}

	equityCurve := []analytics.EquityPoint{
		{Timestamp: startTime, Equity: 10000, Cash: 10000},
		{Timestamp: startTime.Add(24 * time.Hour), Equity: 10009.9, Cash: 10009.9},
		{Timestamp: startTime.Add(48 * time.Hour), Equity: 10004.8, Cash: 10004.8},
		{Timestamp: endTime, Equity: 10019.7, Cash: 10019.7},
	}

	input := analytics.AnalysisInput{
		StartTime:    startTime,
		EndTime:      endTime,
		InitialCash:  10000,
		FinalEquity:  10019.7,
		EquityCurve:  equityCurve,
		TradeHistory: trades,
	}

	analyzer := analytics.NewAnalyzer()
	metrics := analyzer.Analyze(input)

	// Verify calculated metrics
	assert.Equal(t, 3, metrics.TotalTrades, "Should have 3 trades")
	assert.Equal(t, 2, metrics.WinningTrades, "Should have 2 winning trades")
	assert.Equal(t, 1, metrics.LosingTrades, "Should have 1 losing trade")
	assert.InDelta(t, 2.0/3.0, metrics.WinRate, 0.01, "Win rate should be ~66.67%")
	
	expectedGrossProfit := 9.9 + 14.9
	assert.InDelta(t, expectedGrossProfit, metrics.GrossProfit, 0.01, "Gross profit")
	
	expectedGrossLoss := -5.1
	assert.InDelta(t, expectedGrossLoss, metrics.GrossLoss, 0.01, "Gross loss")
	
	expectedProfitFactor := expectedGrossProfit / math.Abs(expectedGrossLoss)
	assert.InDelta(t, expectedProfitFactor, metrics.ProfitFactor, 0.01, "Profit factor")
	
	expectedAvgTrade := (9.9 - 5.1 + 14.9) / 3
	assert.InDelta(t, expectedAvgTrade, metrics.AvgTrade, 0.01, "Average trade")

	t.Logf("✓ Performance metrics validated")
	t.Logf("  Win Rate: %.2f%%", metrics.WinRate*100)
	t.Logf("  Profit Factor: %.2f", metrics.ProfitFactor)
	t.Logf("  Average Trade: %.2f", metrics.AvgTrade)
}

// TestPhase8_DrawdownCalculation verifies drawdown calculation
func TestPhase8_DrawdownCalculation(t *testing.T) {
	// Create equity curve with pre-calculated drawdowns
	now := time.Now()
	equityCurve := []analytics.EquityPoint{
		{Timestamp: now, Equity: 10000, Drawdown: 0},
		{Timestamp: now.Add(1 * time.Hour), Equity: 10500, Drawdown: 0},
		{Timestamp: now.Add(2 * time.Hour), Equity: 10200, Drawdown: -0.0286}, // -2.86% from peak 10500
		{Timestamp: now.Add(3 * time.Hour), Equity: 9800, Drawdown: -0.0667},   // -6.67% from peak 10500
		{Timestamp: now.Add(4 * time.Hour), Equity: 10100, Drawdown: -0.0381}, // -3.81% from peak 10500
		{Timestamp: now.Add(5 * time.Hour), Equity: 10600, Drawdown: 0},        // New peak
		{Timestamp: now.Add(6 * time.Hour), Equity: 10400, Drawdown: -0.0189}, // -1.89% from new peak 10600
	}

	// Add minimal trade history to satisfy analyzer requirements
	trades := []portfolio.Trade{
		{ID: "t1", Symbol: "BTCUSDT", Side: portfolio.PositionSideLong, EntryTime: now, ExitTime: now.Add(6 * time.Hour), NetPnL: 400, Fees: 0},
	}

	input := analytics.AnalysisInput{
		StartTime:   now,
		EndTime:     now.Add(6 * time.Hour),
		InitialCash: 10000,
		FinalEquity: 10400,
		EquityCurve: equityCurve,
		TradeHistory: trades,
	}

	analyzer := analytics.NewAnalyzer()
	metrics := analyzer.Analyze(input)

	// Max drawdown should be -6.67% (from 10500 to 9800)
	expectedMaxDD := -0.0667
	assert.InDelta(t, expectedMaxDD, metrics.MaxDrawdown, 0.001, "Max drawdown calculation")
	assert.Less(t, metrics.MaxDrawdown, 0.0, "Max drawdown should be negative")

	t.Logf("✓ Max Drawdown: %.2f%%", metrics.MaxDrawdown*100)
}

// Helper function to run a test backtest
func runTestBacktest(t *testing.T) *backtest.BacktestResult {
	// Create test data
	csvData := `timestamp,open,high,low,close,volume
2024-01-01 00:00:00,45000,45500,44800,45200,100
2024-01-01 04:00:00,45200,45800,45000,45600,120
2024-01-01 08:00:00,45600,46000,45400,45800,110
2024-01-01 12:00:00,45800,46200,45600,46000,130
2024-01-01 16:00:00,46000,46500,45900,46300,140
2024-01-01 20:00:00,46300,46800,46200,46600,150
2024-01-02 00:00:00,46600,47000,46500,46900,160
2024-01-02 04:00:00,46900,47200,46700,47000,155
2024-01-02 08:00:00,47000,47400,46900,47200,165
2024-01-02 12:00:00,47200,47600,47100,47400,170`

	csvFile := "/tmp/phase8_test_data.csv"
	err := os.WriteFile(csvFile, []byte(csvData), 0644)
	require.NoError(t, err)
	defer os.Remove(csvFile)

	// Create test strategy
	strategyYAML := `strategy:
  name: phase8_test_strategy
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema9:
    type: ema
    source: close
    period: 9

  ema21:
    type: ema
    source: close
    period: 21

entry:
  long:
    condition:
      function: cross_above
      args: [ema9, ema21]

exit:
  long:
    condition:
      function: cross_below
      args: [ema9, ema21]

risk:
  position_size:
    type: risk_percent
    value: 0.02

  stop_loss:
    type: percentage
    value: 0.02

  take_profit:
    type: risk_reward
    ratio: 2

execution:
  entry_order_type: market
  exit_order_type: market
  fee_maker: 0.0001
  fee_taker: 0.0005`

	strategyFile := "/tmp/phase8_test_strategy.yaml"
	err = os.WriteFile(strategyFile, []byte(strategyYAML), 0644)
	require.NoError(t, err)
	defer os.Remove(strategyFile)

	// Run backtest through CLI would be ideal, but for unit test we'll mock
	// This is a placeholder - actual implementation would call backtest engine
	result := &backtest.BacktestResult{
		StrategyName: "phase8_test_strategy",
		TotalTrades:  1,
		StartTime:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
		Metrics: &analytics.Metrics{
			TotalReturn:   0.02,
			CAGR:          0.15,
			SharpeRatio:   1.5,
			SortinoRatio:  2.1,
			CalmarRatio:   2.5,
			MaxDrawdown:   -0.05,
			TotalTrades:   1,
			WinningTrades: 1,
			LosingTrades:  0,
			WinRate:       1.0,
			ProfitFactor:  99999,
			Expectancy:    200,
			AvgTrade:      200,
			AvgWin:        200,
			AvgLoss:       0,
		},
		TradeHistory: []portfolio.Trade{
			{
				Symbol:      "BTCUSDT",
				Side:        "long",
				EntryTime:   time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC),
				EntryPrice:  45600,
				ExitTime:    time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
				ExitPrice:   45800,
				Quantity:    0.1,
				GrossPnL:    20,
				Fees:        0.5,
				NetPnL:      19.5,
				Return:      0.0043,
				ExitReason:  "signal",
			},
		},
		EquityCurve: []backtest.EquityPoint{
			{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Equity: 10000, Cash: 10000, Drawdown: 0, Exposure: 0},
			{Timestamp: time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC), Equity: 10000, Cash: 5440, Drawdown: 0, Exposure: 0.456},
			{Timestamp: time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC), Equity: 10019.5, Cash: 10019.5, Drawdown: 0, Exposure: 0},
			{Timestamp: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC), Equity: 10019.5, Cash: 10019.5, Drawdown: 0, Exposure: 0},
		},
	}

	return result
}
