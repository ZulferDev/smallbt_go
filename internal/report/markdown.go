package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/backtest"
)

// generateMarkdown creates a Markdown report
func (g *Generator) generateMarkdown(result *backtest.BacktestResult) (string, error) {
	var md strings.Builder

	// Title
	md.WriteString(fmt.Sprintf("# %s\n\n", g.config.Title))
	md.WriteString(fmt.Sprintf("*Generated: %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))
	md.WriteString("---\n\n")

	// Summary
	if g.config.IncludeSummary {
		md.WriteString(g.generateSummaryMarkdown(result))
		md.WriteString("\n")
	}

	// Metrics
	if g.config.IncludeMetrics && result.Metrics != nil {
		md.WriteString(g.generateMetricsMarkdown(result.Metrics))
		md.WriteString("\n")
	}

	// Trade History
	if g.config.IncludeTradeHistory && len(result.TradeHistory) > 0 {
		md.WriteString(g.generateTradeHistoryMarkdown(result))
		md.WriteString("\n")
	}

	return md.String(), nil
}

func (g *Generator) generateSummaryMarkdown(result *backtest.BacktestResult) string {
	var md strings.Builder

	md.WriteString("## Strategy Summary\n\n")
	md.WriteString("| Parameter | Value |\n")
	md.WriteString("|-----------|-------|\n")
	md.WriteString(fmt.Sprintf("| Strategy | %s |\n", result.StrategyName))
	md.WriteString(fmt.Sprintf("| Symbol | %s |\n", result.Config.Symbol))
	md.WriteString(fmt.Sprintf("| Timeframe | %s |\n", result.Config.Timeframe))
	md.WriteString(fmt.Sprintf("| Period | %s to %s |\n",
		result.Config.StartTime.Format("2006-01-02"),
		result.Config.EndTime.Format("2006-01-02")))
	md.WriteString(fmt.Sprintf("| Initial Cash | $%.2f |\n", result.Config.InitialCash))
	finalEquity := result.Config.InitialCash
	if len(result.EquityCurve) > 0 {
		finalEquity = result.EquityCurve[len(result.EquityCurve)-1].Equity
	}
	md.WriteString(fmt.Sprintf("| Final Equity | $%.2f |\n", finalEquity))
	md.WriteString(fmt.Sprintf("| Total Trades | %d |\n", result.TotalTrades))

	return md.String()
}

func (g *Generator) generateMetricsMarkdown(metrics *analytics.Metrics) string {
	var md strings.Builder

	md.WriteString("## Performance Metrics\n\n")

	// Returns
	md.WriteString("### Returns\n\n")
	md.WriteString("| Metric | Value |\n")
	md.WriteString("|--------|-------|\n")
	md.WriteString(fmt.Sprintf("| Total Return | **%+.2f%%** |\n", metrics.TotalReturn*100))
	md.WriteString(fmt.Sprintf("| CAGR | %+.2f%% |\n", metrics.CAGR*100))
	md.WriteString(fmt.Sprintf("| Sharpe Ratio | %.2f |\n", metrics.SharpeRatio))
	md.WriteString(fmt.Sprintf("| Sortino Ratio | %.2f |\n", metrics.SortinoRatio))
	md.WriteString("\n")

	// Risk
	md.WriteString("### Risk Metrics\n\n")
	md.WriteString("| Metric | Value |\n")
	md.WriteString("|--------|-------|\n")
	md.WriteString(fmt.Sprintf("| Max Drawdown | **%.2f%%** |\n", metrics.MaxDrawdown*100))
	md.WriteString(fmt.Sprintf("| Calmar Ratio | %.2f |\n", metrics.CalmarRatio))
	md.WriteString(fmt.Sprintf("| Avg Drawdown | %.2f%% |\n", metrics.AvgDrawdown*100))
	md.WriteString("\n")

	// Trade Statistics
	md.WriteString("### Trade Statistics\n\n")
	md.WriteString("| Metric | Value |\n")
	md.WriteString("|--------|-------|\n")
	md.WriteString(fmt.Sprintf("| Win Rate | %.2f%% |\n", metrics.WinRate*100))
	md.WriteString(fmt.Sprintf("| Profit Factor | %.2f |\n", metrics.ProfitFactor))
	md.WriteString(fmt.Sprintf("| Average Trade | $%.2f |\n", metrics.AvgTrade))
	md.WriteString(fmt.Sprintf("| Average Win | $%.2f |\n", metrics.AvgWin))
	md.WriteString(fmt.Sprintf("| Average Loss | $%.2f |\n", metrics.AvgLoss))

	return md.String()
}

func (g *Generator) generateTradeHistoryMarkdown(result *backtest.BacktestResult) string {
	var md strings.Builder

	md.WriteString("## Trade History\n\n")

	winning := 0
	losing := 0
	for _, trade := range result.TradeHistory {
		if trade.NetPnL > 0 {
			winning++
		} else {
			losing++
		}
	}

	md.WriteString(fmt.Sprintf("- **Total Trades:** %d\n", result.TotalTrades))
	md.WriteString(fmt.Sprintf("- **Winning Trades:** %d\n", winning))
	md.WriteString(fmt.Sprintf("- **Losing Trades:** %d\n\n", losing))

	// Recent trades table
	md.WriteString("### Recent Trades (Last 10)\n\n")
	md.WriteString("| Date | Side | Exit Reason | PnL | Return % |\n")
	md.WriteString("|------|------|-------------|-----|----------|\n")

	start := len(result.TradeHistory) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(result.TradeHistory); i++ {
		trade := result.TradeHistory[i]
		pnlSign := ""
		if trade.NetPnL > 0 {
			pnlSign = "✅"
		} else {
			pnlSign = "❌"
		}

		md.WriteString(fmt.Sprintf("| %s | %s | %s | %s $%.2f | %.2f%% |\n",
			trade.EntryTime.Format("2006-01-02"),
			trade.Side,
			trade.ExitReason,
			pnlSign,
			trade.NetPnL,
			trade.Return*100))
	}

	return md.String()
}
