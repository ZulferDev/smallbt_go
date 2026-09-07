package analytics

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// TradeJournalExporter exports trade history to various formats.
type TradeJournalExporter struct {
	trades []portfolio.Trade
}

// NewTradeJournalExporter creates a new trade journal exporter.
func NewTradeJournalExporter(trades []portfolio.Trade) *TradeJournalExporter {
	return &TradeJournalExporter{
		trades: trades,
	}
}

// ExportCSV exports trade history to CSV format with comprehensive details.
func (e *TradeJournalExporter) ExportCSV(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID",
		"Symbol",
		"Side",
		"EntryTime",
		"EntryPrice",
		"ExitTime",
		"ExitPrice",
		"Quantity",
		"Duration_Minutes",
		"GrossPnL",
		"Fees",
		"NetPnL",
		"Return_%",
		"MAE",
		"MFE",
		"MAE_%",
		"MFE_%",
		"ExitReason",
		"PriceChange_%",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write trades
	for _, trade := range e.trades {
		duration := trade.ExitTime.Sub(trade.EntryTime).Minutes()
		priceChange := ((trade.ExitPrice - trade.EntryPrice) / trade.EntryPrice) * 100

		// Calculate MAE/MFE as percentages
		maePercent := (trade.MAE / trade.EntryPrice) * 100
		mfePercent := (trade.MFE / trade.EntryPrice) * 100

		record := []string{
			trade.ID,
			string(trade.Symbol),
			string(trade.Side),
			trade.EntryTime.Format(time.RFC3339),
			fmt.Sprintf("%.6f", trade.EntryPrice),
			trade.ExitTime.Format(time.RFC3339),
			fmt.Sprintf("%.6f", trade.ExitPrice),
			fmt.Sprintf("%.6f", trade.Quantity),
			fmt.Sprintf("%.2f", duration),
			fmt.Sprintf("%.6f", trade.GrossPnL),
			fmt.Sprintf("%.6f", trade.Fees),
			fmt.Sprintf("%.6f", trade.NetPnL),
			fmt.Sprintf("%.4f", trade.Return*100),
			fmt.Sprintf("%.6f", trade.MAE),
			fmt.Sprintf("%.6f", trade.MFE),
			fmt.Sprintf("%.4f", maePercent),
			fmt.Sprintf("%.4f", mfePercent),
			trade.ExitReason,
			fmt.Sprintf("%.4f", priceChange),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	return nil
}

// TradeAnalysis provides advanced trade statistics.
type TradeAnalysis struct {
	TotalTrades     int
	WinningTrades   int
	LosingTrades    int
	BreakevenTrades int

	// Streaks
	CurrentStreak int
	MaxWinStreak  int
	MaxLossStreak int

	// Best/Worst
	BestTrade   *portfolio.Trade
	WorstTrade  *portfolio.Trade
	LargestWin  float64
	LargestLoss float64

	// Holding periods
	AvgHoldingTime time.Duration
	MinHoldingTime time.Duration
	MaxHoldingTime time.Duration

	// MAE/MFE analysis
	AvgMAE        float64
	AvgMFE        float64
	AvgMAEPercent float64
	AvgMFEPercent float64

	// Exit reasons breakdown
	ExitReasons map[string]int
}

// AnalyzeTrades performs comprehensive trade analysis.
func AnalyzeTrades(trades []portfolio.Trade) *TradeAnalysis {
	if len(trades) == 0 {
		return &TradeAnalysis{
			ExitReasons: make(map[string]int),
		}
	}

	analysis := &TradeAnalysis{
		TotalTrades:    len(trades),
		ExitReasons:    make(map[string]int),
		MinHoldingTime: time.Duration(1<<63 - 1), // Max duration
	}

	var (
		totalHoldingTime time.Duration
		totalMAE         float64
		totalMFE         float64
		totalMAEPercent  float64
		totalMFEPercent  float64
		currentStreak    int
		lastWasWin       bool
		maxWinStreak     int
		maxLossStreak    int
	)

	for i := range trades {
		trade := &trades[i]

		// Win/Loss/Breakeven
		if trade.NetPnL > 0 {
			analysis.WinningTrades++

			// Streak tracking
			if lastWasWin {
				currentStreak++
			} else {
				currentStreak = 1
				lastWasWin = true
			}
			if currentStreak > maxWinStreak {
				maxWinStreak = currentStreak
			}

			// Best trade
			if analysis.BestTrade == nil || trade.NetPnL > analysis.BestTrade.NetPnL {
				analysis.BestTrade = trade
				analysis.LargestWin = trade.NetPnL
			}
		} else if trade.NetPnL < 0 {
			analysis.LosingTrades++

			// Streak tracking
			if !lastWasWin && i > 0 {
				currentStreak++
			} else {
				currentStreak = 1
				lastWasWin = false
			}
			if currentStreak > maxLossStreak {
				maxLossStreak = currentStreak
			}

			// Worst trade
			if analysis.WorstTrade == nil || trade.NetPnL < analysis.WorstTrade.NetPnL {
				analysis.WorstTrade = trade
				analysis.LargestLoss = trade.NetPnL
			}
		} else {
			analysis.BreakevenTrades++
		}

		// Holding time
		holdingTime := trade.ExitTime.Sub(trade.EntryTime)
		totalHoldingTime += holdingTime
		if holdingTime < analysis.MinHoldingTime {
			analysis.MinHoldingTime = holdingTime
		}
		if holdingTime > analysis.MaxHoldingTime {
			analysis.MaxHoldingTime = holdingTime
		}

		// MAE/MFE
		totalMAE += trade.MAE
		totalMFE += trade.MFE
		totalMAEPercent += (trade.MAE / trade.EntryPrice) * 100
		totalMFEPercent += (trade.MFE / trade.EntryPrice) * 100

		// Exit reasons
		if trade.ExitReason != "" {
			analysis.ExitReasons[trade.ExitReason]++
		}
	}

	// Calculate averages
	analysis.AvgHoldingTime = totalHoldingTime / time.Duration(len(trades))
	analysis.AvgMAE = totalMAE / float64(len(trades))
	analysis.AvgMFE = totalMFE / float64(len(trades))
	analysis.AvgMAEPercent = totalMAEPercent / float64(len(trades))
	analysis.AvgMFEPercent = totalMFEPercent / float64(len(trades))
	analysis.MaxWinStreak = maxWinStreak
	analysis.MaxLossStreak = maxLossStreak
	analysis.CurrentStreak = currentStreak

	return analysis
}

// FormatAnalysisReport generates a human-readable analysis report.
func FormatAnalysisReport(analysis *TradeAnalysis) string {
	if analysis.TotalTrades == 0 {
		return "No trades to analyze"
	}

	report := fmt.Sprintf(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TRADE ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Trades:      %d
├─ Winning:        %d (%.1f%%)
├─ Losing:         %d (%.1f%%)
└─ Breakeven:      %d (%.1f%%)

Streaks:
├─ Max Win Streak:   %d
└─ Max Loss Streak:  %d

Best Trade:        $%.2f
Worst Trade:       $%.2f

Holding Time:
├─ Average:        %s
├─ Minimum:        %s
└─ Maximum:        %s

MAE/MFE Analysis:
├─ Avg MAE:        %.2f%%
└─ Avg MFE:        %.2f%%

Exit Reasons:
`,
		analysis.TotalTrades,
		analysis.WinningTrades, float64(analysis.WinningTrades)/float64(analysis.TotalTrades)*100,
		analysis.LosingTrades, float64(analysis.LosingTrades)/float64(analysis.TotalTrades)*100,
		analysis.BreakevenTrades, float64(analysis.BreakevenTrades)/float64(analysis.TotalTrades)*100,
		analysis.MaxWinStreak,
		analysis.MaxLossStreak,
		analysis.LargestWin,
		analysis.LargestLoss,
		formatDuration(analysis.AvgHoldingTime),
		formatDuration(analysis.MinHoldingTime),
		formatDuration(analysis.MaxHoldingTime),
		analysis.AvgMAEPercent,
		analysis.AvgMFEPercent,
	)

	for reason, count := range analysis.ExitReasons {
		percent := float64(count) / float64(analysis.TotalTrades) * 100
		report += fmt.Sprintf("├─ %s: %d (%.1f%%)\n", reason, count, percent)
	}

	report += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

	return report
}

// formatDuration formats a duration in human-readable format.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}
