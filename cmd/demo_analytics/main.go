package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/backtest"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

func main() {
	// Create mock backtest result with trades
	result := backtest.BacktestResult{
		StrategyName: "test_strategy",
		StartTime:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		TotalTrades:  5,
		TradeHistory: []portfolio.Trade{
			{
				ID:         "TRADE-001",
				Symbol:     market.Symbol("BTCUSDT"),
				Side:       portfolio.PositionSideLong,
				EntryTime:  time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
				EntryPrice: 42000.0,
				ExitTime:   time.Date(2024, 1, 3, 14, 0, 0, 0, time.UTC),
				ExitPrice:  43500.0,
				Quantity:   0.1,
				GrossPnL:   150.0,
				Fees:       3.0,
				NetPnL:     147.0,
				Return:     0.035,
				MAE:        -50.0,
				MFE:        180.0,
				ExitReason: "take_profit",
			},
			{
				ID:         "TRADE-002",
				Symbol:     market.Symbol("BTCUSDT"),
				Side:       portfolio.PositionSideLong,
				EntryTime:  time.Date(2024, 1, 4, 9, 0, 0, 0, time.UTC),
				EntryPrice: 43200.0,
				ExitTime:   time.Date(2024, 1, 4, 16, 0, 0, 0, time.UTC),
				ExitPrice:  42800.0,
				Quantity:   0.1,
				GrossPnL:   -40.0,
				Fees:       2.0,
				NetPnL:     -42.0,
				Return:     -0.0097,
				MAE:        -60.0,
				MFE:        20.0,
				ExitReason: "stop_loss",
			},
			{
				ID:         "TRADE-003",
				Symbol:     market.Symbol("BTCUSDT"),
				Side:       portfolio.PositionSideLong,
				EntryTime:  time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
				EntryPrice: 42500.0,
				ExitTime:   time.Date(2024, 1, 6, 15, 0, 0, 0, time.UTC),
				ExitPrice:  44800.0,
				Quantity:   0.15,
				GrossPnL:   345.0,
				Fees:       4.5,
				NetPnL:     340.5,
				Return:     0.054,
				MAE:        -80.0,
				MFE:        400.0,
				ExitReason: "take_profit",
			},
			{
				ID:         "TRADE-004",
				Symbol:     market.Symbol("BTCUSDT"),
				Side:       portfolio.PositionSideLong,
				EntryTime:  time.Date(2024, 1, 7, 10, 0, 0, 0, time.UTC),
				EntryPrice: 44500.0,
				ExitTime:   time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC),
				ExitPrice:  46200.0,
				Quantity:   0.12,
				GrossPnL:   204.0,
				Fees:       3.8,
				NetPnL:     200.2,
				Return:     0.038,
				MAE:        -45.0,
				MFE:        230.0,
				ExitReason: "take_profit",
			},
			{
				ID:         "TRADE-005",
				Symbol:     market.Symbol("BTCUSDT"),
				Side:       portfolio.PositionSideLong,
				EntryTime:  time.Date(2024, 1, 9, 8, 0, 0, 0, time.UTC),
				EntryPrice: 45800.0,
				ExitTime:   time.Date(2024, 1, 9, 14, 0, 0, 0, time.UTC),
				ExitPrice:  45300.0,
				Quantity:   0.08,
				GrossPnL:   -40.0,
				Fees:       1.8,
				NetPnL:     -41.8,
				Return:     -0.011,
				MAE:        -55.0,
				MFE:        15.0,
				ExitReason: "stop_loss",
			},
		},
	}

	// Save to JSON
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	_ = os.WriteFile("/tmp/demo_result.json", resultJSON, 0644)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("DEMO: Trade Export & Analysis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\nGenerated backtest result with %d trades\n", len(result.TradeHistory))
	fmt.Println("Result saved to: /tmp/demo_result.json")

	// Test CSV export
	fmt.Println("\n1. Exporting to CSV...")
	exporter := analytics.NewTradeJournalExporter(result.TradeHistory)
	if err := exporter.ExportCSV("/tmp/demo_trades.csv"); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("✓ CSV exported to: /tmp/demo_trades.csv")

	// Show CSV preview
	csvData, _ := os.ReadFile("/tmp/demo_trades.csv")
	lines := 0
	for i, b := range csvData {
		if b == '\n' {
			lines++
			if lines == 3 {
				fmt.Printf("\nCSV Preview (first 3 lines):\n%s...\n", string(csvData[:i]))
				break
			}
		}
	}

	// Test trade analysis
	fmt.Println("\n2. Analyzing trades...")
	analysis := analytics.AnalyzeTrades(result.TradeHistory)
	report := analytics.FormatAnalysisReport(analysis)
	fmt.Println(report)

	// Save report
	_ = os.WriteFile("/tmp/demo_analysis.txt", []byte(report), 0644)
	fmt.Println("✓ Analysis saved to: /tmp/demo_analysis.txt")

	fmt.Println("\nYou can now test the CLI commands:")
	fmt.Println("  ./trader export-trades --result /tmp/demo_result.json --output /tmp/trades.csv")
	fmt.Println("  ./trader analyze-trades --result /tmp/demo_result.json")
	fmt.Println()
}
