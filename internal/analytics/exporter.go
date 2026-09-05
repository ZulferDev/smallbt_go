package analytics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// Exporter handles exporting analytics data to various formats.
type Exporter struct{}

// NewExporter creates a new Exporter.
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportMetricsJSON exports metrics to a JSON file.
func (e *Exporter) ExportMetricsJSON(metrics *Metrics, filePath string) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metrics to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write metrics JSON file: %w", err)
	}

	return nil
}

// ExportMetricsCSV exports metrics to a CSV file.
func (e *Exporter) ExportMetricsCSV(metrics *Metrics, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create metrics CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"metric", "value"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	// Write metrics
	rows := [][]string{
		{"total_return", formatFloat(metrics.TotalReturn)},
		{"cagr", formatFloat(metrics.CAGR)},
		{"sharpe_ratio", formatFloat(metrics.SharpeRatio)},
		{"sortino_ratio", formatFloat(metrics.SortinoRatio)},
		{"calmar_ratio", formatFloat(metrics.CalmarRatio)},
		{"max_drawdown", formatFloat(metrics.MaxDrawdown)},
		{"total_trades", fmt.Sprintf("%d", metrics.TotalTrades)},
		{"winning_trades", fmt.Sprintf("%d", metrics.WinningTrades)},
		{"losing_trades", fmt.Sprintf("%d", metrics.LosingTrades)},
		{"win_rate", formatFloat(metrics.WinRate)},
		{"gross_profit", formatFloat(metrics.GrossProfit)},
		{"gross_loss", formatFloat(metrics.GrossLoss)},
		{"net_profit", formatFloat(metrics.NetProfit)},
		{"profit_factor", formatFloat(metrics.ProfitFactor)},
		{"avg_trade", formatFloat(metrics.AvgTrade)},
		{"avg_win", formatFloat(metrics.AvgWin)},
		{"avg_loss", formatFloat(metrics.AvgLoss)},
		{"largest_win", formatFloat(metrics.LargestWin)},
		{"largest_loss", formatFloat(metrics.LargestLoss)},
		{"expectancy", formatFloat(metrics.Expectancy)},
		{"total_fees", formatFloat(metrics.TotalFees)},
		{"avg_exposure", formatFloat(metrics.AvgExposure)},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	return nil
}

// ExportEquityCurveJSON exports equity curve to a JSON file.
func (e *Exporter) ExportEquityCurveJSON(equityCurve []EquityPoint, filePath string) error {
	data, err := json.MarshalIndent(equityCurve, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal equity curve to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write equity curve JSON file: %w", err)
	}

	return nil
}

// ExportEquityCurveCSV exports equity curve to a CSV file.
func (e *Exporter) ExportEquityCurveCSV(equityCurve []EquityPoint, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create equity curve CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"timestamp", "equity", "cash", "drawdown", "exposure"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	// Write rows
	for _, point := range equityCurve {
		row := []string{
			point.Timestamp.Format(time.RFC3339),
			formatFloat(point.Equity),
			formatFloat(point.Cash),
			formatFloat(point.Drawdown),
			formatFloat(point.Exposure),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	return nil
}

// TradeExport represents a trade for export.
type TradeExport struct {
	ID         string  `json:"id"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	EntryTime  string  `json:"entry_time"`
	EntryPrice float64 `json:"entry_price"`
	ExitTime   string  `json:"exit_time"`
	ExitPrice  float64 `json:"exit_price"`
	Quantity   float64 `json:"quantity"`
	GrossPnL   float64 `json:"gross_pnl"`
	Fees       float64 `json:"fees"`
	NetPnL     float64 `json:"net_pnl"`
	Return     float64 `json:"return"`
	MAE        float64 `json:"mae"`
	MFE        float64 `json:"mfe"`
	ExitReason string  `json:"exit_reason"`
}

// ExportTradesJSON exports trade history to a JSON file.
func (e *Exporter) ExportTradesJSON(trades []portfolio.Trade, filePath string) error {
	exports := make([]TradeExport, len(trades))
	for i, trade := range trades {
		exports[i] = TradeExport{
			ID:         trade.ID,
			Symbol:     string(trade.Symbol),
			Side:       string(trade.Side),
			EntryTime:  trade.EntryTime.Format(time.RFC3339),
			EntryPrice: trade.EntryPrice,
			ExitTime:   trade.ExitTime.Format(time.RFC3339),
			ExitPrice:  trade.ExitPrice,
			Quantity:   trade.Quantity,
			GrossPnL:   trade.GrossPnL,
			Fees:       trade.Fees,
			NetPnL:     trade.NetPnL,
			Return:     trade.Return,
			MAE:        trade.MAE,
			MFE:        trade.MFE,
			ExitReason: trade.ExitReason,
		}
	}

	data, err := json.MarshalIndent(exports, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal trades to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write trades JSON file: %w", err)
	}

	return nil
}

// ExportTradesCSV exports trade history to a CSV file.
func (e *Exporter) ExportTradesCSV(trades []portfolio.Trade, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create trades CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"id", "symbol", "side", "entry_time", "entry_price",
		"exit_time", "exit_price", "quantity", "gross_pnl",
		"fees", "net_pnl", "return", "mae", "mfe", "exit_reason",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	// Write rows
	for _, trade := range trades {
		row := []string{
			trade.ID,
			string(trade.Symbol),
			string(trade.Side),
			trade.EntryTime.Format(time.RFC3339),
			formatFloat(trade.EntryPrice),
			trade.ExitTime.Format(time.RFC3339),
			formatFloat(trade.ExitPrice),
			formatFloat(trade.Quantity),
			formatFloat(trade.GrossPnL),
			formatFloat(trade.Fees),
			formatFloat(trade.NetPnL),
			formatFloat(trade.Return),
			formatFloat(trade.MAE),
			formatFloat(trade.MFE),
			trade.ExitReason,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	return nil
}

// formatFloat formats a float64 to a string with reasonable precision.
func formatFloat(v float64) string {
	return fmt.Sprintf("%.8f", v)
}
