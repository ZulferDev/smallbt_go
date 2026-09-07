package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/backtest"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
	"github.com/ZulferDev/smallbt_go/internal/report"
)

func main() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("REPORT GENERATION DEMO")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create realistic backtest result
	result := createRealisticBacktestResult()

	// Generate HTML report
	htmlConfig := report.ReportConfig{
		Format:              report.FormatHTML,
		Title:               "Demo Strategy Backtest Report",
		IncludeSummary:      true,
		IncludeMetrics:      true,
		IncludeTradeHistory: true,
		IncludeCSS:          true,
		Theme:               "light",
	}

	htmlGen := report.NewGenerator(htmlConfig)
	htmlReport, err := htmlGen.Generate(result)
	if err != nil {
		log.Fatalf("Generate HTML report failed: %v", err)
	}

	htmlFile := "demo_report.html"
	if err := os.WriteFile(htmlFile, []byte(htmlReport.Content), 0644); err != nil {
		log.Fatalf("Write HTML failed: %v", err)
	}
	fmt.Printf("✅ HTML report generated: %s\n", htmlFile)

	// Generate Markdown report
	mdConfig := report.ReportConfig{
		Format:              report.FormatMarkdown,
		Title:               "Demo Strategy Backtest Report",
		IncludeSummary:      true,
		IncludeMetrics:      true,
		IncludeTradeHistory: true,
	}

	mdGen := report.NewGenerator(mdConfig)
	mdReport, err := mdGen.Generate(result)
	if err != nil {
		log.Fatalf("Generate Markdown report failed: %v", err)
	}

	mdFile := "demo_report.md"
	if err := os.WriteFile(mdFile, []byte(mdReport.Content), 0644); err != nil {
		log.Fatalf("Write Markdown failed: %v", err)
	}
	fmt.Printf("✅ Markdown report generated: %s\n", mdFile)

	// Generate text report (print to console)
	textConfig := report.ReportConfig{
		Format:              report.FormatText,
		Title:               "Demo Strategy Backtest Report",
		IncludeSummary:      true,
		IncludeMetrics:      true,
		IncludeTradeHistory: true,
	}

	textGen := report.NewGenerator(textConfig)
	textReport, err := textGen.Generate(result)
	if err != nil {
		log.Fatalf("Generate text report failed: %v", err)
	}

	fmt.Println("\n" + textReport.Content)
	fmt.Println("\n✅ All report formats generated successfully!")
	fmt.Printf("   - HTML: %s\n", htmlFile)
	fmt.Printf("   - Markdown: %s\n", mdFile)
	fmt.Println("   - Text: printed above")
}

func createRealisticBacktestResult() *backtest.BacktestResult {
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	// Create realistic trade history
	trades := []portfolio.Trade{
		{
			EntryTime:  startTime.Add(24 * time.Hour),
			ExitTime:   startTime.Add(48 * time.Hour),
			Side:       portfolio.PositionSideLong,
			EntryPrice: 42500.0,
			ExitPrice:  44200.0,
			Quantity:   0.25,
			GrossPnL:   425.0,
			Fees:       8.5,
			NetPnL:     416.5,
			Return:     0.04,
			ExitReason: "Take Profit",
		},
		{
			EntryTime:  startTime.Add(72 * time.Hour),
			ExitTime:   startTime.Add(96 * time.Hour),
			Side:       portfolio.PositionSideLong,
			EntryPrice: 44000.0,
			ExitPrice:  43200.0,
			Quantity:   0.25,
			GrossPnL:   -200.0,
			Fees:       8.8,
			NetPnL:     -208.8,
			Return:     -0.019,
			ExitReason: "Stop Loss",
		},
		{
			EntryTime:  startTime.Add(120 * time.Hour),
			ExitTime:   startTime.Add(168 * time.Hour),
			Side:       portfolio.PositionSideLong,
			EntryPrice: 43500.0,
			ExitPrice:  46800.0,
			Quantity:   0.30,
			GrossPnL:   990.0,
			Fees:       13.05,
			NetPnL:     976.95,
			Return:     0.075,
			ExitReason: "Take Profit",
		},
		{
			EntryTime:  startTime.Add(200 * time.Hour),
			ExitTime:   startTime.Add(240 * time.Hour),
			Side:       portfolio.PositionSideLong,
			EntryPrice: 47000.0,
			ExitPrice:  48500.0,
			Quantity:   0.28,
			GrossPnL:   420.0,
			Fees:       13.23,
			NetPnL:     406.77,
			Return:     0.032,
			ExitReason: "Take Profit",
		},
		{
			EntryTime:  startTime.Add(280 * time.Hour),
			ExitTime:   startTime.Add(320 * time.Hour),
			Side:       portfolio.PositionSideLong,
			EntryPrice: 48800.0,
			ExitPrice:  47500.0,
			Quantity:   0.26,
			GrossPnL:   -338.0,
			Fees:       12.54,
			NetPnL:     -350.54,
			Return:     -0.028,
			ExitReason: "Stop Loss",
		},
	}

	return &backtest.BacktestResult{
		Config: backtest.BacktestConfig{
			Symbol:       market.Symbol("BTCUSDT"),
			Timeframe:    market.Timeframe("4h"),
			StartTime:    startTime,
			EndTime:      endTime,
			InitialCash:  10000.0,
			StrategyPath: "demo_strategy.yaml",
		},
		StrategyName: "EMA Crossover with Volume Filter",
		TotalTrades:  45,
		StartTime:    startTime,
		EndTime:      endTime,
		Metrics: &analytics.Metrics{
			TotalReturn:   0.234,
			CAGR:          0.234,
			SharpeRatio:   1.52,
			SortinoRatio:  2.08,
			CalmarRatio:   1.48,
			MaxDrawdown:   0.158,
			AvgDrawdown:   0.048,
			TotalTrades:   45,
			WinningTrades: 27,
			LosingTrades:  18,
			WinRate:       0.60,
			GrossProfit:   3850.0,
			GrossLoss:     -1510.0,
			NetProfit:     2340.0,
			ProfitFactor:  2.55,
			AvgTrade:      52.0,
			AvgWin:        142.59,
			AvgLoss:       -83.89,
		},
		TradeHistory: trades,
		EquityCurve: []backtest.EquityPoint{
			{Timestamp: startTime, Equity: 10000.0, Cash: 10000.0, Drawdown: 0.0},
			{Timestamp: startTime.Add(2160 * time.Hour), Equity: 11170.0, Cash: 11170.0, Drawdown: 0.08},
			{Timestamp: startTime.Add(4320 * time.Hour), Equity: 11850.0, Cash: 11850.0, Drawdown: 0.05},
			{Timestamp: startTime.Add(6480 * time.Hour), Equity: 12120.0, Cash: 12120.0, Drawdown: 0.02},
			{Timestamp: endTime, Equity: 12340.0, Cash: 12340.0, Drawdown: 0.0},
		},
	}
}
