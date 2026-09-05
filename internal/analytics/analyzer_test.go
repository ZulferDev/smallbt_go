package analytics

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

func TestAnalyzerEmptyTrades(t *testing.T) {
	analyzer := NewAnalyzer()
	input := AnalysisInput{
		InitialCash:  10000,
		FinalEquity:  10000,
		StartTime:    time.Now().AddDate(-1, 0, 0),
		EndTime:      time.Now(),
		TradeHistory: []portfolio.Trade{},
		EquityCurve:  []EquityPoint{},
	}

	metrics := analyzer.Analyze(input)
	if metrics == nil {
		t.Fatal("expected metrics, got nil")
	}

	if metrics.TotalTrades != 0 {
		t.Errorf("expected 0 trades, got %d", metrics.TotalTrades)
	}
}

func TestAnalyzerTotalReturn(t *testing.T) {
	analyzer := NewAnalyzer()
	now := time.Now()
	input := AnalysisInput{
		InitialCash: 10000,
		FinalEquity: 12000,
		StartTime:   now.AddDate(-1, 0, 0),
		EndTime:     now,
		TradeHistory: []portfolio.Trade{
			{NetPnL: 2000},
		},
		EquityCurve: []EquityPoint{
			{Timestamp: now.AddDate(-1, 0, 0), Equity: 10000, Drawdown: 0},
			{Timestamp: now, Equity: 12000, Drawdown: 0},
		},
	}

	metrics := analyzer.Analyze(input)

	expectedReturn := 0.2 // 20%
	if metrics.TotalReturn < expectedReturn-0.01 || metrics.TotalReturn > expectedReturn+0.01 {
		t.Errorf("expected total return ~%.2f, got %.8f", expectedReturn, metrics.TotalReturn)
	}
}

func TestAnalyzerWinRate(t *testing.T) {
	analyzer := NewAnalyzer()
	now := time.Now()
	input := AnalysisInput{
		InitialCash: 10000,
		FinalEquity: 11000,
		StartTime:   now.AddDate(-1, 0, 0),
		EndTime:     now,
		TradeHistory: []portfolio.Trade{
			{NetPnL: 500},
			{NetPnL: -200},
			{NetPnL: 300},
			{NetPnL: -100},
		},
		EquityCurve: []EquityPoint{
			{Timestamp: now.AddDate(-1, 0, 0), Equity: 10000},
			{Timestamp: now, Equity: 11000},
		},
	}

	metrics := analyzer.Analyze(input)

	if metrics.TotalTrades != 4 {
		t.Errorf("expected 4 trades, got %d", metrics.TotalTrades)
	}

	if metrics.WinningTrades != 2 {
		t.Errorf("expected 2 winning trades, got %d", metrics.WinningTrades)
	}

	if metrics.LosingTrades != 2 {
		t.Errorf("expected 2 losing trades, got %d", metrics.LosingTrades)
	}

	expectedWinRate := 0.5
	if metrics.WinRate < expectedWinRate-0.01 || metrics.WinRate > expectedWinRate+0.01 {
		t.Errorf("expected win rate ~%.2f, got %.8f", expectedWinRate, metrics.WinRate)
	}
}

func TestAnalyzerProfitFactor(t *testing.T) {
	analyzer := NewAnalyzer()
	now := time.Now()
	input := AnalysisInput{
		InitialCash: 10000,
		FinalEquity: 10500,
		StartTime:   now.AddDate(-1, 0, 0),
		EndTime:     now,
		TradeHistory: []portfolio.Trade{
			{NetPnL: 1000},
			{NetPnL: -500},
		},
		EquityCurve: []EquityPoint{
			{Timestamp: now.AddDate(-1, 0, 0), Equity: 10000},
			{Timestamp: now, Equity: 10500},
		},
	}

	metrics := analyzer.Analyze(input)

	expectedProfitFactor := 2.0 // 1000 / 500
	if metrics.ProfitFactor < expectedProfitFactor-0.01 || metrics.ProfitFactor > expectedProfitFactor+0.01 {
		t.Errorf("expected profit factor ~%.2f, got %.8f", expectedProfitFactor, metrics.ProfitFactor)
	}
}

func TestAnalyzerSharpeRatio(t *testing.T) {
	analyzer := NewAnalyzer()
	now := time.Now()

	// Create equity curve with consistent positive returns
	equityCurve := []EquityPoint{
		{Timestamp: now.AddDate(0, 0, -10), Equity: 10000},
		{Timestamp: now.AddDate(0, 0, -9), Equity: 10100},
		{Timestamp: now.AddDate(0, 0, -8), Equity: 10200},
		{Timestamp: now.AddDate(0, 0, -7), Equity: 10300},
		{Timestamp: now.AddDate(0, 0, -6), Equity: 10400},
		{Timestamp: now.AddDate(0, 0, -5), Equity: 10500},
		{Timestamp: now.AddDate(0, 0, -4), Equity: 10600},
		{Timestamp: now.AddDate(0, 0, -3), Equity: 10700},
		{Timestamp: now.AddDate(0, 0, -2), Equity: 10800},
		{Timestamp: now.AddDate(0, 0, -1), Equity: 10900},
		{Timestamp: now, Equity: 11000},
	}

	input := AnalysisInput{
		InitialCash:  10000,
		FinalEquity:  11000,
		StartTime:    now.AddDate(0, 0, -10),
		EndTime:      now,
		TradeHistory: []portfolio.Trade{{NetPnL: 1000}},
		EquityCurve:  equityCurve,
		RiskFreeRate: 0.02, // 2% annual
	}

	metrics := analyzer.Analyze(input)

	// With consistent positive returns, Sharpe should be positive
	if metrics.SharpeRatio <= 0 {
		t.Errorf("expected positive Sharpe ratio, got %.8f", metrics.SharpeRatio)
	}
}

func TestAnalyzerMaxDrawdown(t *testing.T) {
	analyzer := NewAnalyzer()
	now := time.Now()

	// Equity curve: 10000 -> 11000 -> 9000 -> 10500
	// Max drawdown: 9000/11000 - 1 = -18.18%
	equityCurve := []EquityPoint{
		{Timestamp: now.AddDate(0, 0, -3), Equity: 10000, Drawdown: 0},
		{Timestamp: now.AddDate(0, 0, -2), Equity: 11000, Drawdown: 0},
		{Timestamp: now.AddDate(0, 0, -1), Equity: 9000, Drawdown: -0.1818},
		{Timestamp: now, Equity: 10500, Drawdown: -0.0454},
	}

	input := AnalysisInput{
		InitialCash:  10000,
		FinalEquity:  10500,
		StartTime:    now.AddDate(0, 0, -3),
		EndTime:      now,
		TradeHistory: []portfolio.Trade{{NetPnL: 500}},
		EquityCurve:  equityCurve,
	}

	metrics := analyzer.Analyze(input)

	// Max drawdown should be around -18%
	if metrics.MaxDrawdown > -0.17 || metrics.MaxDrawdown < -0.19 {
		t.Errorf("expected max drawdown ~-0.18, got %.8f", metrics.MaxDrawdown)
	}
}

func TestAnalyzerExpectancy(t *testing.T) {
	analyzer := NewAnalyzer()
	now := time.Now()
	input := AnalysisInput{
		InitialCash: 10000,
		FinalEquity: 10600,
		StartTime:   now.AddDate(-1, 0, 0),
		EndTime:     now,
		TradeHistory: []portfolio.Trade{
			{NetPnL: 1000},
			{NetPnL: -400},
		},
		EquityCurve: []EquityPoint{
			{Timestamp: now.AddDate(-1, 0, 0), Equity: 10000},
			{Timestamp: now, Equity: 10600},
		},
	}

	metrics := analyzer.Analyze(input)

	// Expectancy = Net Profit / Total Trades = 600 / 2 = 300
	expectedExpectancy := 300.0
	if metrics.Expectancy < expectedExpectancy-1 || metrics.Expectancy > expectedExpectancy+1 {
		t.Errorf("expected expectancy ~%.2f, got %.8f", expectedExpectancy, metrics.Expectancy)
	}
}

func TestCalculateTotalReturn(t *testing.T) {
	tests := []struct {
		name        string
		initialCash float64
		finalEquity float64
		expected    float64
	}{
		{"positive return", 10000, 12000, 0.2},
		{"negative return", 10000, 8000, -0.2},
		{"zero return", 10000, 10000, 0.0},
		{"zero initial", 0, 12000, 0.0},
		{"negative initial", -10000, 12000, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTotalReturn(tt.initialCash, tt.finalEquity)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("expected %.2f, got %.8f", tt.expected, result)
			}
		})
	}
}

func TestCalculateProfitFactor(t *testing.T) {
	tests := []struct {
		name        string
		grossProfit float64
		grossLoss   float64
		expected    float64
	}{
		{"normal case", 1000, -500, 2.0},
		{"no losses", 1000, 0, 1000.0},
		{"no profits", 0, -500, 0.0},
		{"no trades", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateProfitFactor(tt.grossProfit, tt.grossLoss)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("expected %.2f, got %.8f", tt.expected, result)
			}
		})
	}
}

func TestCalculateCAGR(t *testing.T) {
	// 20% return over 1 year
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now()
	result := calculateCAGR(0.2, start, end)

	// CAGR should be approximately 20% for 1 year
	if result < 0.19 || result > 0.21 {
		t.Errorf("expected CAGR ~0.20, got %.8f", result)
	}
}

func TestCalculateReturns(t *testing.T) {
	now := time.Now()
	equityCurve := []EquityPoint{
		{Timestamp: now.AddDate(0, 0, -2), Equity: 10000},
		{Timestamp: now.AddDate(0, 0, -1), Equity: 10100},
		{Timestamp: now, Equity: 10200},
	}

	returns := calculateReturns(equityCurve)

	if len(returns) != 2 {
		t.Errorf("expected 2 returns, got %d", len(returns))
	}

	// First return: (10100 - 10000) / 10000 = 0.01
	if returns[0] < 0.009 || returns[0] > 0.011 {
		t.Errorf("expected first return ~0.01, got %.8f", returns[0])
	}

	// Second return: (10200 - 10100) / 10100 = 0.0099
	if returns[1] < 0.009 || returns[1] > 0.011 {
		t.Errorf("expected second return ~0.0099, got %.8f", returns[1])
	}
}
