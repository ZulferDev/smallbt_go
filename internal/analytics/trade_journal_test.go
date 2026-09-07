package analytics

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

func TestTradeJournalExporter_ExportCSV(t *testing.T) {
	trades := []portfolio.Trade{
		{
			ID:         "TRADE-001",
			Symbol:     market.Symbol("BTCUSDT"),
			Side:       portfolio.PositionSideLong,
			EntryTime:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EntryPrice: 50000.0,
			ExitTime:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			ExitPrice:  51000.0,
			Quantity:   0.1,
			GrossPnL:   100.0,
			Fees:       2.0,
			NetPnL:     98.0,
			Return:     0.0196,
			MAE:        -50.0,
			MFE:        150.0,
			ExitReason: "take_profit",
		},
		{
			ID:         "TRADE-002",
			Symbol:     market.Symbol("ETHUSDT"),
			Side:       portfolio.PositionSideShort,
			EntryTime:  time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			EntryPrice: 3000.0,
			ExitTime:   time.Date(2024, 1, 2, 16, 0, 0, 0, time.UTC),
			ExitPrice:  2950.0,
			Quantity:   1.0,
			GrossPnL:   50.0,
			Fees:       1.0,
			NetPnL:     49.0,
			Return:     0.0163,
			MAE:        -20.0,
			MFE:        60.0,
			ExitReason: "stop_loss",
		},
	}

	exporter := NewTradeJournalExporter(trades)

	// Create temp directory
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "trades.csv")

	// Export CSV
	err := exporter.ExportCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Fatalf("CSV file not created")
	}

	// Read and verify content
	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	contentStr := string(content)

	// Verify header
	if !containsString(contentStr, "ID,Symbol,Side,EntryTime") {
		t.Error("CSV header not found")
	}

	// Verify trade data
	if !containsString(contentStr, "TRADE-001") {
		t.Error("Trade 001 not found in CSV")
	}
	if !containsString(contentStr, "TRADE-002") {
		t.Error("Trade 002 not found in CSV")
	}
	if !containsString(contentStr, "take_profit") {
		t.Error("Exit reason not found in CSV")
	}
}

func TestTradeJournalExporter_ExportCSV_EmptyTrades(t *testing.T) {
	exporter := NewTradeJournalExporter([]portfolio.Trade{})

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")

	err := exporter.ExportCSV(csvPath)
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	// Verify file exists with header only
	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Read CSV failed: %v", err)
	}

	if len(content) == 0 {
		t.Error("CSV file is empty")
	}
}

func TestAnalyzeTrades(t *testing.T) {
	trades := []portfolio.Trade{
		// Winning trade
		{
			ID:         "WIN-1",
			EntryTime:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			EntryPrice: 50000.0,
			NetPnL:     100.0,
			MAE:        -50.0,
			MFE:        150.0,
			ExitReason: "take_profit",
		},
		// Losing trade
		{
			ID:         "LOSS-1",
			EntryTime:  time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 2, 14, 0, 0, 0, time.UTC),
			EntryPrice: 51000.0,
			NetPnL:     -80.0,
			MAE:        -100.0,
			MFE:        20.0,
			ExitReason: "stop_loss",
		},
		// Another winning trade
		{
			ID:         "WIN-2",
			EntryTime:  time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 3, 18, 0, 0, 0, time.UTC),
			EntryPrice: 52000.0,
			NetPnL:     200.0,
			MAE:        -30.0,
			MFE:        250.0,
			ExitReason: "take_profit",
		},
		// Breakeven trade
		{
			ID:         "BE-1",
			EntryTime:  time.Date(2024, 1, 4, 10, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 4, 11, 0, 0, 0, time.UTC),
			EntryPrice: 53000.0,
			NetPnL:     0.0,
			MAE:        -10.0,
			MFE:        10.0,
			ExitReason: "manual",
		},
	}

	analysis := AnalyzeTrades(trades)

	// Test basic counts
	if analysis.TotalTrades != 4 {
		t.Errorf("Expected 4 total trades, got %d", analysis.TotalTrades)
	}
	if analysis.WinningTrades != 2 {
		t.Errorf("Expected 2 winning trades, got %d", analysis.WinningTrades)
	}
	if analysis.LosingTrades != 1 {
		t.Errorf("Expected 1 losing trade, got %d", analysis.LosingTrades)
	}
	if analysis.BreakevenTrades != 1 {
		t.Errorf("Expected 1 breakeven trade, got %d", analysis.BreakevenTrades)
	}

	// Test best/worst trades
	if analysis.BestTrade == nil || analysis.BestTrade.ID != "WIN-2" {
		t.Errorf("Best trade should be WIN-2")
	}
	if analysis.LargestWin != 200.0 {
		t.Errorf("Expected largest win 200.0, got %.2f", analysis.LargestWin)
	}
	if analysis.WorstTrade == nil || analysis.WorstTrade.ID != "LOSS-1" {
		t.Errorf("Worst trade should be LOSS-1")
	}
	if analysis.LargestLoss != -80.0 {
		t.Errorf("Expected largest loss -80.0, got %.2f", analysis.LargestLoss)
	}

	// Test streaks
	if analysis.MaxWinStreak < 1 {
		t.Errorf("Expected max win streak >= 1, got %d", analysis.MaxWinStreak)
	}

	// Test holding times
	if analysis.AvgHoldingTime == 0 {
		t.Error("Average holding time should not be zero")
	}
	if analysis.MinHoldingTime == 0 {
		t.Error("Min holding time should not be zero")
	}
	if analysis.MaxHoldingTime == 0 {
		t.Error("Max holding time should not be zero")
	}

	// Test MAE/MFE
	if analysis.AvgMAE >= 0 {
		t.Errorf("Average MAE should be negative, got %.2f", analysis.AvgMAE)
	}
	if analysis.AvgMFE <= 0 {
		t.Errorf("Average MFE should be positive, got %.2f", analysis.AvgMFE)
	}

	// Test exit reasons
	if len(analysis.ExitReasons) == 0 {
		t.Error("Exit reasons should not be empty")
	}
	if analysis.ExitReasons["take_profit"] != 2 {
		t.Errorf("Expected 2 take_profit exits, got %d", analysis.ExitReasons["take_profit"])
	}
	if analysis.ExitReasons["stop_loss"] != 1 {
		t.Errorf("Expected 1 stop_loss exit, got %d", analysis.ExitReasons["stop_loss"])
	}
}

func TestAnalyzeTrades_EmptyTrades(t *testing.T) {
	analysis := AnalyzeTrades([]portfolio.Trade{})

	if analysis.TotalTrades != 0 {
		t.Errorf("Expected 0 total trades, got %d", analysis.TotalTrades)
	}
	if analysis.WinningTrades != 0 {
		t.Errorf("Expected 0 winning trades, got %d", analysis.WinningTrades)
	}
	if analysis.BestTrade != nil {
		t.Error("Best trade should be nil for empty trades")
	}
	if analysis.WorstTrade != nil {
		t.Error("Worst trade should be nil for empty trades")
	}
}

func TestAnalyzeTrades_AllWinners(t *testing.T) {
	trades := []portfolio.Trade{
		{NetPnL: 100.0, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour), EntryPrice: 50000.0, MAE: -10, MFE: 110},
		{NetPnL: 150.0, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour), EntryPrice: 50000.0, MAE: -20, MFE: 170},
		{NetPnL: 200.0, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour), EntryPrice: 50000.0, MAE: -15, MFE: 220},
	}

	analysis := AnalyzeTrades(trades)

	if analysis.WinningTrades != 3 {
		t.Errorf("Expected 3 winning trades, got %d", analysis.WinningTrades)
	}
	if analysis.LosingTrades != 0 {
		t.Errorf("Expected 0 losing trades, got %d", analysis.LosingTrades)
	}
	if analysis.MaxWinStreak != 3 {
		t.Errorf("Expected max win streak 3, got %d", analysis.MaxWinStreak)
	}
	if analysis.MaxLossStreak != 0 {
		t.Errorf("Expected max loss streak 0, got %d", analysis.MaxLossStreak)
	}
}

func TestAnalyzeTrades_AllLosers(t *testing.T) {
	trades := []portfolio.Trade{
		{NetPnL: -100.0, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour), EntryPrice: 50000.0, MAE: -110, MFE: 10},
		{NetPnL: -150.0, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour), EntryPrice: 50000.0, MAE: -170, MFE: 20},
		{NetPnL: -200.0, EntryTime: time.Now(), ExitTime: time.Now().Add(time.Hour), EntryPrice: 50000.0, MAE: -220, MFE: 15},
	}

	analysis := AnalyzeTrades(trades)

	if analysis.WinningTrades != 0 {
		t.Errorf("Expected 0 winning trades, got %d", analysis.WinningTrades)
	}
	if analysis.LosingTrades != 3 {
		t.Errorf("Expected 3 losing trades, got %d", analysis.LosingTrades)
	}
	if analysis.MaxLossStreak != 3 {
		t.Errorf("Expected max loss streak 3, got %d", analysis.MaxLossStreak)
	}
}

func TestFormatAnalysisReport(t *testing.T) {
	trades := []portfolio.Trade{
		{
			NetPnL:     100.0,
			EntryTime:  time.Now(),
			ExitTime:   time.Now().Add(2 * time.Hour),
			EntryPrice: 50000.0,
			MAE:        -50.0,
			MFE:        150.0,
			ExitReason: "take_profit",
		},
		{
			NetPnL:     -80.0,
			EntryTime:  time.Now(),
			ExitTime:   time.Now().Add(4 * time.Hour),
			EntryPrice: 51000.0,
			MAE:        -100.0,
			MFE:        20.0,
			ExitReason: "stop_loss",
		},
	}

	analysis := AnalyzeTrades(trades)
	report := FormatAnalysisReport(analysis)

	// Verify report contains key sections
	if !containsString(report, "TRADE ANALYSIS") {
		t.Error("Report should contain TRADE ANALYSIS header")
	}
	if !containsString(report, "Total Trades:") {
		t.Error("Report should contain Total Trades")
	}
	if !containsString(report, "Winning:") {
		t.Error("Report should contain Winning trades")
	}
	if !containsString(report, "Losing:") {
		t.Error("Report should contain Losing trades")
	}
	if !containsString(report, "Best Trade:") {
		t.Error("Report should contain Best Trade")
	}
	if !containsString(report, "Worst Trade:") {
		t.Error("Report should contain Worst Trade")
	}
	if !containsString(report, "MAE/MFE Analysis:") {
		t.Error("Report should contain MAE/MFE Analysis")
	}
	if !containsString(report, "Exit Reasons:") {
		t.Error("Report should contain Exit Reasons")
	}
	if !containsString(report, "take_profit") {
		t.Error("Report should contain exit reason details")
	}
}

func TestFormatAnalysisReport_EmptyTrades(t *testing.T) {
	analysis := AnalyzeTrades([]portfolio.Trade{})
	report := FormatAnalysisReport(analysis)

	if !containsString(report, "No trades to analyze") {
		t.Error("Report should indicate no trades")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"30 seconds", 30 * time.Second, "30s"},
		{"5 minutes", 5 * time.Minute, "5.0m"},
		{"2.5 hours", 150 * time.Minute, "2.5h"},
		{"3 days", 72 * time.Hour, "3.0d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
