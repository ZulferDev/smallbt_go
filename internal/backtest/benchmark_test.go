package backtest

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// BenchmarkBacktestSmall benchmarks a small backtest (100 candles)
func BenchmarkBacktestSmall(b *testing.B) {
	benchmarkBacktest(b, 100)
}

// BenchmarkBacktestMedium benchmarks a medium backtest (1000 candles)
func BenchmarkBacktestMedium(b *testing.B) {
	benchmarkBacktest(b, 1000)
}

// BenchmarkBacktestLarge benchmarks a large backtest (2000 candles)
func BenchmarkBacktestLarge(b *testing.B) {
	benchmarkBacktest(b, 2000)
}

func benchmarkBacktest(b *testing.B, numCandles int) {
	// Generate test data file
	tmpFile := generateTestCSV(b, numCandles)
	
	// Create backtest config
	config := BacktestConfig{
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		InitialCash:  10000.0,
		StrategyPath: "../../strategies/examples/ema_volume.yaml",
		DataPath:     tmpFile,
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := Run(config)
		if err != nil {
			b.Fatalf("backtest failed: %v", err)
		}
	}
}

// BenchmarkBacktestDifferentStrategies benchmarks multiple strategies
func BenchmarkBacktestEMAVolume(b *testing.B) {
	benchmarkStrategy(b, "../../strategies/examples/ema_volume.yaml", 1000)
}

func BenchmarkBacktestEMACross(b *testing.B) {
	benchmarkStrategy(b, "../../strategies/examples/ema_cross.yaml", 1000)
}

func BenchmarkBacktestSMACross(b *testing.B) {
	benchmarkStrategy(b, "../../strategies/examples/sma_cross.yaml", 1000)
}

func benchmarkStrategy(b *testing.B, strategyPath string, numCandles int) {
	tmpFile := generateTestCSV(b, numCandles)
	
	config := BacktestConfig{
		Symbol:       "BTCUSDT",
		Timeframe:    "1h",
		InitialCash:  10000.0,
		StrategyPath: strategyPath,
		DataPath:     tmpFile,
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := Run(config)
		if err != nil {
			b.Fatalf("backtest failed: %v", err)
		}
	}
}

// Helper: Generate test CSV file with realistic price data
func generateTestCSV(b *testing.B, numCandles int) string {
	b.Helper()
	
	tmpFile := b.TempDir() + "/test_data.csv"
	f, err := os.Create(tmpFile)
	if err != nil {
		b.Fatalf("create temp file: %v", err)
	}
	defer f.Close()
	
	// Write CSV header
	f.WriteString("timestamp,open,high,low,close,volume\n")
	
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	basePrice := 10000.0
	
	for i := 0; i < numCandles; i++ {
		// Simulate realistic price movement with trend and volatility
		trend := float64(i) * 0.5
		noise := float64((i*7)%100-50) * 10.0
		price := basePrice + trend + noise
		
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		open := price
		high := price * 1.015
		low := price * 0.985
		close := price + float64((i*3)%20-10)
		volume := 1000000.0 + float64(i*1000)
		
		f.WriteString(fmt.Sprintf("%s,%.2f,%.2f,%.2f,%.2f,%.2f\n",
			timestamp.Format("2006-01-02 15:04:05"),
			open, high, low, close, volume))
	}
	
	return tmpFile
}
