package report

import (
	"strings"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/backtest"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

func TestGenerateTextReport(t *testing.T) {
	result := createMockBacktestResult()
	
	config := DefaultReportConfig()
	config.Format = FormatText
	
	gen := NewGenerator(config)
	report, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	if report == nil {
		t.Fatal("Report is nil")
	}
	
	if report.Content == "" {
		t.Error("Report content is empty")
	}
	
	// Verify key sections are present
	if !strings.Contains(report.Content, "STRATEGY SUMMARY") {
		t.Error("Missing strategy summary section")
	}
	
	if !strings.Contains(report.Content, "PERFORMANCE METRICS") {
		t.Error("Missing performance metrics section")
	}
	
	if !strings.Contains(report.Content, "TRADE HISTORY") {
		t.Error("Missing trade history section")
	}
}

func TestGenerateMarkdownReport(t *testing.T) {
	result := createMockBacktestResult()
	
	config := DefaultReportConfig()
	config.Format = FormatMarkdown
	
	gen := NewGenerator(config)
	report, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Verify markdown formatting
	if !strings.Contains(report.Content, "# Backtest Report") {
		t.Error("Missing markdown title")
	}
	
	if !strings.Contains(report.Content, "## Strategy Summary") {
		t.Error("Missing strategy summary header")
	}
	
	if !strings.Contains(report.Content, "| Parameter | Value |") {
		t.Error("Missing table formatting")
	}
}

func TestGenerateHTMLReport(t *testing.T) {
	result := createMockBacktestResult()
	
	config := DefaultReportConfig()
	config.Format = FormatHTML
	config.IncludeCSS = true
	
	gen := NewGenerator(config)
	report, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Verify HTML structure
	if !strings.Contains(report.Content, "<!DOCTYPE html>") {
		t.Error("Missing DOCTYPE")
	}
	
	if !strings.Contains(report.Content, "<html") {
		t.Error("Missing html tag")
	}
	
	if !strings.Contains(report.Content, "<style>") {
		t.Error("Missing CSS")
	}
	
	if !strings.Contains(report.Content, "<table>") {
		t.Error("Missing tables")
	}
}

func TestGenerateHTMLReportDarkTheme(t *testing.T) {
	result := createMockBacktestResult()
	
	config := DefaultReportConfig()
	config.Format = FormatHTML
	config.Theme = "dark"
	
	gen := NewGenerator(config)
	report, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Verify dark theme CSS is included
	if !strings.Contains(report.Content, "background: #1a1a1a") {
		t.Error("Dark theme CSS not found")
	}
}

func TestGenerateReportWithoutSections(t *testing.T) {
	result := createMockBacktestResult()
	
	config := ReportConfig{
		Format:              FormatText,
		Title:               "Test Report",
		IncludeSummary:      false,
		IncludeMetrics:      false,
		IncludeTradeHistory: false,
	}
	
	gen := NewGenerator(config)
	report, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	// Should still have header and footer
	if !strings.Contains(report.Content, "Test Report") {
		t.Error("Missing title")
	}
}

func TestGenerateReportNilResult(t *testing.T) {
	config := DefaultReportConfig()
	gen := NewGenerator(config)
	
	_, err := gen.Generate(nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestDefaultReportConfig(t *testing.T) {
	config := DefaultReportConfig()
	
	if config.Format != FormatHTML {
		t.Error("Default format should be HTML")
	}
	
	if !config.IncludeSummary {
		t.Error("Default should include summary")
	}
	
	if !config.IncludeMetrics {
		t.Error("Default should include metrics")
	}
	
	if config.Theme != "light" {
		t.Error("Default theme should be light")
	}
}

// Helper function to create mock backtest result
func createMockBacktestResult() *backtest.BacktestResult {
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	
	return &backtest.BacktestResult{
		Config: backtest.BacktestConfig{
			Symbol:       market.Symbol("BTCUSDT"),
			Timeframe:    market.Timeframe("1h"),
			StartTime:    startTime,
			EndTime:      endTime,
			InitialCash:  10000.0,
			StrategyPath: "test_strategy.yaml",
		},
		StrategyName: "Test Strategy",
		TotalTrades:  25,
		StartTime:    startTime,
		EndTime:      endTime,
		Metrics: &analytics.Metrics{
			TotalReturn:  0.185,
			CAGR:         0.15,
			SharpeRatio:  1.45,
			SortinoRatio: 1.85,
			CalmarRatio:  1.25,
			MaxDrawdown:  0.12,
			AvgDrawdown:  0.05,
			WinRate:      0.56,
			ProfitFactor: 1.75,
			AvgTrade:     74.0,
			AvgWin:       185.5,
			AvgLoss:      -95.2,
		},
		TradeHistory: []portfolio.Trade{
			{
				EntryTime:  startTime.Add(24 * time.Hour),
				ExitTime:   startTime.Add(48 * time.Hour),
				Side:       portfolio.PositionSideLong,
				EntryPrice: 45000.0,
				ExitPrice:  46500.0,
				Quantity:   0.2,
				NetPnL:     300.0,
				Return:     0.033,
				ExitReason: "Take Profit",
			},
			{
				EntryTime:  startTime.Add(72 * time.Hour),
				ExitTime:   startTime.Add(96 * time.Hour),
				Side:       portfolio.PositionSideLong,
				EntryPrice: 46000.0,
				ExitPrice:  45500.0,
				Quantity:   0.2,
				NetPnL:     -100.0,
				Return:     -0.011,
				ExitReason: "Stop Loss",
			},
		},
		EquityCurve: []backtest.EquityPoint{
			{Timestamp: startTime, Equity: 10000.0, Cash: 10000.0},
			{Timestamp: endTime, Equity: 11850.0, Cash: 11850.0},
		},
	}
}
