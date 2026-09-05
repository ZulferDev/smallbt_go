package main

import (
	"fmt"
	"log"
	"time"

	"github.com/1jehuang/backtest/internal/montecarlo"
)

func main() {
	// Example: Run Monte Carlo analysis on backtest results
	// This demonstrates the Phase 13 requirements:
	// - 10000 simulations (AGENTS.md line 1517)
	// - seed 42 for reproducibility (AGENTS.md line 1232)
	// - Trade reshuffling
	// - Return reshuffling
	// - Bootstrap sampling
	// - Drawdown distribution
	// - Confidence intervals

	// Sample trades from a backtest
	trades := generateSampleTrades()

	fmt.Println("Monte Carlo Analysis Example")
	fmt.Println("=============================")
	fmt.Printf("Original trades: %d\n", len(trades))
	fmt.Printf("Initial capital: $%.2f\n\n", 10000.0)

	// 1. Trade Reshuffling Analysis
	fmt.Println("1. Trade Reshuffling (10000 simulations, seed 42)")
	fmt.Println("   Randomizes trade order while preserving each trade's P&L")
	runTradeReshuffleAnalysis(trades)

	// 2. Return Reshuffling Analysis
	fmt.Println("\n2. Return Reshuffling (10000 simulations, seed 42)")
	fmt.Println("   Randomizes returns while preserving total P&L")
	runReturnReshuffleAnalysis(trades)

	// 3. Bootstrap Analysis
	fmt.Println("\n3. Bootstrap Sampling (10000 simulations, seed 42)")
	fmt.Println("   Samples trades with replacement")
	runBootstrapAnalysis(trades)
}

func generateSampleTrades() []montecarlo.Trade {
	// Generate realistic sample trades
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
		{
			ID:         6,
			EntryTime:  time.Date(2024, 1, 6, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 6, 14, 0, 0, 0, time.UTC),
			EntryPrice: 51500.0,
			ExitPrice:  51200.0,
			Quantity:   0.1,
			GrossPnL:   -30.0,
			Fees:       5.0,
			NetPnL:     -35.0,
			Return:     -0.007,
		},
		{
			ID:         7,
			EntryTime:  time.Date(2024, 1, 7, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 7, 14, 0, 0, 0, time.UTC),
			EntryPrice: 51200.0,
			ExitPrice:  52000.0,
			Quantity:   0.1,
			GrossPnL:   80.0,
			Fees:       5.0,
			NetPnL:     75.0,
			Return:     0.016,
		},
		{
			ID:         8,
			EntryTime:  time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 8, 14, 0, 0, 0, time.UTC),
			EntryPrice: 52000.0,
			ExitPrice:  51800.0,
			Quantity:   0.1,
			GrossPnL:   -20.0,
			Fees:       5.0,
			NetPnL:     -25.0,
			Return:     -0.004,
		},
		{
			ID:         9,
			EntryTime:  time.Date(2024, 1, 9, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 9, 14, 0, 0, 0, time.UTC),
			EntryPrice: 51800.0,
			ExitPrice:  52500.0,
			Quantity:   0.1,
			GrossPnL:   70.0,
			Fees:       5.0,
			NetPnL:     65.0,
			Return:     0.014,
		},
		{
			ID:         10,
			EntryTime:  time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
			EntryPrice: 52500.0,
			ExitPrice:  52300.0,
			Quantity:   0.1,
			GrossPnL:   -20.0,
			Fees:       5.0,
			NetPnL:     -25.0,
			Return:     -0.004,
		},
	}
	return trades
}

func runTradeReshuffleAnalysis(trades []montecarlo.Trade) {
	config := montecarlo.MCConfig{
		Simulations: 10000,
		Seed:        42,
		Type:        montecarlo.TradeReshuffle,
	}

	runner := montecarlo.NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		log.Fatalf("Trade reshuffle failed: %v", err)
	}

	printResults(result)
}

func runReturnReshuffleAnalysis(trades []montecarlo.Trade) {
	config := montecarlo.MCConfig{
		Simulations: 10000,
		Seed:        42,
		Type:        montecarlo.ReturnReshuffle,
	}

	runner := montecarlo.NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		log.Fatalf("Return reshuffle failed: %v", err)
	}

	printResults(result)
}

func runBootstrapAnalysis(trades []montecarlo.Trade) {
	config := montecarlo.MCConfig{
		Simulations: 10000,
		Seed:        42,
		Type:        montecarlo.BootstrapReshuffle,
	}

	runner := montecarlo.NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		log.Fatalf("Bootstrap failed: %v", err)
	}

	printResults(result)
}

func printResults(result *montecarlo.MCResult) {
	stats := result.Statistics

	fmt.Printf("   Simulations:      %d\n", result.Config.Simulations)
	fmt.Printf("   Analysis Type:    %s\n", result.Config.Type.String())
	fmt.Println()
	fmt.Println("   Statistics:")
	fmt.Printf("   - Mean Return:        %+.4f (%.2f%%)\n", stats.MeanReturn, stats.MeanReturn*100)
	fmt.Printf("   - Std Dev Return:     %.4f (%.2f%%)\n", stats.StdDevReturn, stats.StdDevReturn*100)
	fmt.Printf("   - Mean Max Drawdown:  %.4f (%.2f%%)\n", stats.MeanMaxDrawdown, stats.MeanMaxDrawdown*100)
	fmt.Printf("   - Std Dev Drawdown:   %.4f (%.2f%%)\n", stats.StdDevMaxDrawdown, stats.StdDevMaxDrawdown*100)
	fmt.Printf("   - Mean Win Rate:      %.4f (%.2f%%)\n", stats.MeanWinRate, stats.MeanWinRate*100)
	fmt.Printf("   - Mean Sharpe Ratio:  %.4f\n", stats.MeanSharpe)
	fmt.Printf("   - Probability of Ruin: %.4f (%.2f%%)\n", stats.ProbabilityOfRuin, stats.ProbabilityOfRuin*100)
	fmt.Println()
	fmt.Println("   Confidence Intervals:")
	for _, ci := range result.ConfidenceIntervals {
		fmt.Printf("   - %5.1f%%: Return=%+.4f (%.2f%%), Drawdown=%.4f (%.2f%%), WinRate=%.2f%%, Sharpe=%.4f\n",
			ci.Percentile*100,
			ci.TotalReturn, ci.TotalReturn*100,
			ci.MaxDrawdown, ci.MaxDrawdown*100,
			ci.WinRate*100,
			ci.SharpeRatio)
	}
}
