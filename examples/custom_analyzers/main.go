package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// WinStreakAnalyzer calculates the maximum consecutive winning trades.
type WinStreakAnalyzer struct{}

func (w *WinStreakAnalyzer) Name() string {
	return "max_win_streak"
}

func (w *WinStreakAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
	if len(input.TradeHistory) == 0 {
		return 0, nil
	}

	maxStreak := 0
	currentStreak := 0

	for _, trade := range input.TradeHistory {
		if trade.NetPnL > 0 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
	}

	return maxStreak, nil
}

// LossStreakAnalyzer calculates the maximum consecutive losing trades.
type LossStreakAnalyzer struct{}

func (l *LossStreakAnalyzer) Name() string {
	return "max_loss_streak"
}

func (l *LossStreakAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
	if len(input.TradeHistory) == 0 {
		return 0, nil
	}

	maxStreak := 0
	currentStreak := 0

	for _, trade := range input.TradeHistory {
		if trade.NetPnL <= 0 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
	}

	return maxStreak, nil
}

// AvgTradeDurationAnalyzer calculates average trade duration in hours.
type AvgTradeDurationAnalyzer struct{}

func (a *AvgTradeDurationAnalyzer) Name() string {
	return "avg_trade_duration_hours"
}

func (a *AvgTradeDurationAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
	if len(input.TradeHistory) == 0 {
		return 0.0, nil
	}

	totalDuration := 0.0
	for _, trade := range input.TradeHistory {
		duration := trade.ExitTime.Sub(trade.EntryTime).Hours()
		totalDuration += duration
	}

	avgHours := totalDuration / float64(len(input.TradeHistory))
	return avgHours, nil
}

// RiskRewardAnalyzer calculates average risk-reward ratio using MAE/MFE.
type RiskRewardAnalyzer struct{}

func (r *RiskRewardAnalyzer) Name() string {
	return "avg_risk_reward_ratio"
}

func (r *RiskRewardAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
	if len(input.TradeHistory) == 0 {
		return 0.0, nil
	}

	totalRR := 0.0
	count := 0

	for _, trade := range input.TradeHistory {
		if math.Abs(trade.MAE) > 0.0001 { // Avoid division by near-zero
			rr := math.Abs(trade.MFE / trade.MAE)
			totalRR += rr
			count++
		}
	}

	if count == 0 {
		return 0.0, nil
	}

	return totalRR / float64(count), nil
}

// TradeBreakdown contains statistics broken down by trade side.
type TradeBreakdown struct {
	LongTrades    int     `json:"long_trades"`
	ShortTrades   int     `json:"short_trades"`
	LongWinRate   float64 `json:"long_win_rate"`
	ShortWinRate  float64 `json:"short_win_rate"`
	LongNetProfit float64 `json:"long_net_profit"`
	ShortNetProfit float64 `json:"short_net_profit"`
}

// SideBreakdownAnalyzer analyzes long vs short performance.
type SideBreakdownAnalyzer struct{}

func (s *SideBreakdownAnalyzer) Name() string {
	return "side_breakdown"
}

func (s *SideBreakdownAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
	if len(input.TradeHistory) == 0 {
		return TradeBreakdown{}, nil
	}

	breakdown := TradeBreakdown{}
	longWins := 0
	shortWins := 0

	for _, trade := range input.TradeHistory {
		if trade.Side == portfolio.PositionSideLong {
			breakdown.LongTrades++
			breakdown.LongNetProfit += trade.NetPnL
			if trade.NetPnL > 0 {
				longWins++
			}
		} else if trade.Side == portfolio.PositionSideShort {
			breakdown.ShortTrades++
			breakdown.ShortNetProfit += trade.NetPnL
			if trade.NetPnL > 0 {
				shortWins++
			}
		}
	}

	if breakdown.LongTrades > 0 {
		breakdown.LongWinRate = float64(longWins) / float64(breakdown.LongTrades)
	}
	if breakdown.ShortTrades > 0 {
		breakdown.ShortWinRate = float64(shortWins) / float64(breakdown.ShortTrades)
	}

	return breakdown, nil
}

// ReturnStdDevAnalyzer calculates standard deviation of returns.
type ReturnStdDevAnalyzer struct{}

func (r *ReturnStdDevAnalyzer) Name() string {
	return "return_std_dev"
}

func (r *ReturnStdDevAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
	if len(input.TradeHistory) == 0 {
		return 0.0, nil
	}

	// Calculate mean return
	sum := 0.0
	for _, trade := range input.TradeHistory {
		sum += trade.Return
	}
	mean := sum / float64(len(input.TradeHistory))

	// Calculate variance
	variance := 0.0
	for _, trade := range input.TradeHistory {
		diff := trade.Return - mean
		variance += diff * diff
	}
	variance /= float64(len(input.TradeHistory))

	// Standard deviation
	stdDev := math.Sqrt(variance)
	return stdDev, nil
}

// MonthlyStats contains monthly performance statistics.
type MonthlyStats struct {
	Month      string  `json:"month"`
	Trades     int     `json:"trades"`
	NetProfit  float64 `json:"net_profit"`
	WinRate    float64 `json:"win_rate"`
}

// MonthlyBreakdownAnalyzer analyzes performance by month.
type MonthlyBreakdownAnalyzer struct{}

func (m *MonthlyBreakdownAnalyzer) Name() string {
	return "monthly_breakdown"
}

func (m *MonthlyBreakdownAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
	if len(input.TradeHistory) == 0 {
		return []MonthlyStats{}, nil
	}

	// Group trades by month
	monthlyData := make(map[string]*MonthlyStats)

	for _, trade := range input.TradeHistory {
		month := trade.ExitTime.Format("2006-01")
		
		if _, exists := monthlyData[month]; !exists {
			monthlyData[month] = &MonthlyStats{
				Month: month,
			}
		}

		stats := monthlyData[month]
		stats.Trades++
		stats.NetProfit += trade.NetPnL
		if trade.NetPnL > 0 {
			// Count for win rate calculation later
		}
	}

	// Convert to slice and calculate win rates
	results := make([]MonthlyStats, 0, len(monthlyData))
	for month, stats := range monthlyData {
		// Recalculate win rate
		wins := 0
		for _, trade := range input.TradeHistory {
			if trade.ExitTime.Format("2006-01") == month && trade.NetPnL > 0 {
				wins++
			}
		}
		if stats.Trades > 0 {
			stats.WinRate = float64(wins) / float64(stats.Trades)
		}
		results = append(results, *stats)
	}

	return results, nil
}

func main() {
	fmt.Println("=== Custom Analyzer Demo ===")
	fmt.Println()

	// Register all custom analyzers
	fmt.Println("Registering custom analyzers...")
	
	if err := analytics.Register(&WinStreakAnalyzer{}); err != nil {
		log.Fatalf("Failed to register WinStreakAnalyzer: %v", err)
	}
	
	if err := analytics.Register(&LossStreakAnalyzer{}); err != nil {
		log.Fatalf("Failed to register LossStreakAnalyzer: %v", err)
	}
	
	if err := analytics.Register(&AvgTradeDurationAnalyzer{}); err != nil {
		log.Fatalf("Failed to register AvgTradeDurationAnalyzer: %v", err)
	}
	
	if err := analytics.Register(&RiskRewardAnalyzer{}); err != nil {
		log.Fatalf("Failed to register RiskRewardAnalyzer: %v", err)
	}
	
	if err := analytics.Register(&SideBreakdownAnalyzer{}); err != nil {
		log.Fatalf("Failed to register SideBreakdownAnalyzer: %v", err)
	}
	
	if err := analytics.Register(&ReturnStdDevAnalyzer{}); err != nil {
		log.Fatalf("Failed to register ReturnStdDevAnalyzer: %v", err)
	}
	
	if err := analytics.Register(&MonthlyBreakdownAnalyzer{}); err != nil {
		log.Fatalf("Failed to register MonthlyBreakdownAnalyzer: %v", err)
	}

	fmt.Printf("Registered %d custom analyzers\n\n", len(analytics.List()))

	// Create sample trade history
	now := time.Now()
	trades := []portfolio.Trade{
		{Symbol: market.Symbol("BTCUSDT"), Side: portfolio.PositionSideLong, EntryTime: now.Add(-100 * time.Hour), ExitTime: now.Add(-96 * time.Hour), NetPnL: 150, Return: 0.015, MAE: -50, MFE: 180},
		{Symbol: market.Symbol("BTCUSDT"), Side: portfolio.PositionSideLong, EntryTime: now.Add(-90 * time.Hour), ExitTime: now.Add(-85 * time.Hour), NetPnL: 200, Return: 0.02, MAE: -30, MFE: 220},
		{Symbol: market.Symbol("BTCUSDT"), Side: portfolio.PositionSideLong, EntryTime: now.Add(-80 * time.Hour), ExitTime: now.Add(-76 * time.Hour), NetPnL: -80, Return: -0.008, MAE: -100, MFE: 20},
		{Symbol: market.Symbol("BTCUSDT"), Side: portfolio.PositionSideShort, EntryTime: now.Add(-70 * time.Hour), ExitTime: now.Add(-65 * time.Hour), NetPnL: 120, Return: 0.012, MAE: -40, MFE: 140},
		{Symbol: market.Symbol("BTCUSDT"), Side: portfolio.PositionSideLong, EntryTime: now.Add(-60 * time.Hour), ExitTime: now.Add(-55 * time.Hour), NetPnL: 180, Return: 0.018, MAE: -60, MFE: 200},
		{Symbol: market.Symbol("BTCUSDT"), Side: portfolio.PositionSideLong, EntryTime: now.Add(-50 * time.Hour), ExitTime: now.Add(-46 * time.Hour), NetPnL: 90, Return: 0.009, MAE: -35, MFE: 110},
		{Symbol: market.Symbol("BTCUSDT"), Side: portfolio.PositionSideShort, EntryTime: now.Add(-40 * time.Hour), ExitTime: now.Add(-36 * time.Hour), NetPnL: -50, Return: -0.005, MAE: -70, MFE: 15},
		{Symbol: market.Symbol("BTCUSDT"), Side: portfolio.PositionSideLong, EntryTime: now.Add(-30 * time.Hour), ExitTime: now.Add(-25 * time.Hour), NetPnL: 220, Return: 0.022, MAE: -45, MFE: 240},
	}

	// Create analysis input
	input := analytics.AnalysisInput{
		InitialCash:  10000,
		FinalEquity:  10830,
		StartTime:    now.Add(-100 * time.Hour),
		EndTime:      now,
		TradeHistory: trades,
		RiskFreeRate: 0.02,
	}

	fmt.Println("Sample Trade History:")
	fmt.Printf("  Total trades: %d\n", len(trades))
	fmt.Printf("  Initial cash: $%.2f\n", input.InitialCash)
	fmt.Printf("  Final equity: $%.2f\n\n", input.FinalEquity)

	// Calculate all custom metrics
	fmt.Println("Calculating custom metrics...")
	fmt.Println()
	results := analytics.CalculateAll(input)

	// Display results
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("CUSTOM ANALYZER RESULTS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Win/Loss streaks
	if maxWinStreak, ok := results["max_win_streak"].(int); ok {
		fmt.Printf("Max Win Streak:       %d trades\n", maxWinStreak)
	}
	if maxLossStreak, ok := results["max_loss_streak"].(int); ok {
		fmt.Printf("Max Loss Streak:      %d trades\n", maxLossStreak)
	}

	// Average trade duration
	if avgDuration, ok := results["avg_trade_duration_hours"].(float64); ok {
		fmt.Printf("Avg Trade Duration:   %.2f hours\n", avgDuration)
	}

	// Risk-reward ratio
	if rr, ok := results["avg_risk_reward_ratio"].(float64); ok {
		fmt.Printf("Avg Risk/Reward:      %.2f\n", rr)
	}

	// Return std dev
	if stdDev, ok := results["return_std_dev"].(float64); ok {
		fmt.Printf("Return Std Dev:       %.4f\n\n", stdDev)
	}

	// Side breakdown
	if breakdown, ok := results["side_breakdown"].(TradeBreakdown); ok {
		fmt.Println("Side Breakdown:")
		fmt.Printf("  Long Trades:        %d\n", breakdown.LongTrades)
		fmt.Printf("  Long Win Rate:      %.2f%%\n", breakdown.LongWinRate*100)
		fmt.Printf("  Long Net Profit:    $%.2f\n", breakdown.LongNetProfit)
		fmt.Printf("  Short Trades:       %d\n", breakdown.ShortTrades)
		fmt.Printf("  Short Win Rate:     %.2f%%\n", breakdown.ShortWinRate*100)
		fmt.Printf("  Short Net Profit:   $%.2f\n\n", breakdown.ShortNetProfit)
	}

	// Monthly breakdown
	if monthly, ok := results["monthly_breakdown"].([]MonthlyStats); ok && len(monthly) > 0 {
		fmt.Println("Monthly Breakdown:")
		for _, stats := range monthly {
			fmt.Printf("  %s: %d trades, $%.2f profit, %.2f%% win rate\n",
				stats.Month, stats.Trades, stats.NetProfit, stats.WinRate*100)
		}
		fmt.Println()
	}

	// List all registered analyzers
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Registered Analyzers:")
	for _, name := range analytics.List() {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\n=== Demo Complete ===")
}
