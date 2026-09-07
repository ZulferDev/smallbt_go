package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/backtest"
)

// generateHTML creates an HTML report
func (g *Generator) generateHTML(result *backtest.BacktestResult) (string, error) {
	var html strings.Builder

	// HTML header
	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html lang=\"en\">\n")
	html.WriteString("<head>\n")
	html.WriteString("  <meta charset=\"UTF-8\">\n")
	html.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	html.WriteString(fmt.Sprintf("  <title>%s</title>\n", g.config.Title))

	// Include CSS if requested
	if g.config.IncludeCSS {
		html.WriteString("  <style>\n")
		html.WriteString(g.getCSS())
		html.WriteString("  </style>\n")
	}

	html.WriteString("</head>\n")
	html.WriteString("<body>\n")
	html.WriteString("  <div class=\"container\">\n")

	// Header
	html.WriteString(fmt.Sprintf("    <h1>%s</h1>\n", g.config.Title))
	html.WriteString(fmt.Sprintf("    <p class=\"generated\">Generated: %s</p>\n",
		time.Now().Format("2006-01-02 15:04:05")))
	html.WriteString("    <hr>\n")

	// Summary
	if g.config.IncludeSummary {
		html.WriteString(g.generateSummaryHTML(result))
	}

	// Metrics
	if g.config.IncludeMetrics && result.Metrics != nil {
		html.WriteString(g.generateMetricsHTML(result.Metrics))
	}

	// Trade History
	if g.config.IncludeTradeHistory && len(result.TradeHistory) > 0 {
		html.WriteString(g.generateTradeHistoryHTML(result))
	}

	// Footer
	html.WriteString("  </div>\n")
	html.WriteString("</body>\n")
	html.WriteString("</html>\n")

	return html.String(), nil
}

func (g *Generator) generateSummaryHTML(result *backtest.BacktestResult) string {
	var html strings.Builder

	html.WriteString("    <section class=\"summary\">\n")
	html.WriteString("      <h2>Strategy Summary</h2>\n")
	html.WriteString("      <table>\n")
	html.WriteString("        <tr><th>Parameter</th><th>Value</th></tr>\n")
	html.WriteString(fmt.Sprintf("        <tr><td>Strategy</td><td>%s</td></tr>\n", result.StrategyName))
	html.WriteString(fmt.Sprintf("        <tr><td>Symbol</td><td>%s</td></tr>\n", result.Config.Symbol))
	html.WriteString(fmt.Sprintf("        <tr><td>Timeframe</td><td>%s</td></tr>\n", result.Config.Timeframe))
	html.WriteString(fmt.Sprintf("        <tr><td>Period</td><td>%s to %s</td></tr>\n",
		result.Config.StartTime.Format("2006-01-02"),
		result.Config.EndTime.Format("2006-01-02")))
	html.WriteString(fmt.Sprintf("        <tr><td>Initial Cash</td><td>$%.2f</td></tr>\n", result.Config.InitialCash))
	finalEquity := result.Config.InitialCash
	if len(result.EquityCurve) > 0 {
		finalEquity = result.EquityCurve[len(result.EquityCurve)-1].Equity
	}
	html.WriteString(fmt.Sprintf("        <tr><td>Final Equity</td><td class=\"highlight\">$%.2f</td></tr>\n", finalEquity))
	html.WriteString(fmt.Sprintf("        <tr><td>Total Trades</td><td>%d</td></tr>\n", result.TotalTrades))
	html.WriteString("      </table>\n")
	html.WriteString("    </section>\n")

	return html.String()
}

func (g *Generator) generateMetricsHTML(metrics *analytics.Metrics) string {
	var html strings.Builder

	html.WriteString("    <section class=\"metrics\">\n")
	html.WriteString("      <h2>Performance Metrics</h2>\n")

	// Returns
	html.WriteString("      <h3>Returns</h3>\n")
	html.WriteString("      <table>\n")
	html.WriteString("        <tr><th>Metric</th><th>Value</th></tr>\n")

	returnClass := "positive"
	if metrics.TotalReturn < 0 {
		returnClass = "negative"
	}
	html.WriteString(fmt.Sprintf("        <tr><td>Total Return</td><td class=\"%s\">%+.2f%%</td></tr>\n",
		returnClass, metrics.TotalReturn*100))
	html.WriteString(fmt.Sprintf("        <tr><td>CAGR</td><td>%+.2f%%</td></tr>\n", metrics.CAGR*100))
	html.WriteString(fmt.Sprintf("        <tr><td>Sharpe Ratio</td><td>%.2f</td></tr>\n", metrics.SharpeRatio))
	html.WriteString(fmt.Sprintf("        <tr><td>Sortino Ratio</td><td>%.2f</td></tr>\n", metrics.SortinoRatio))
	html.WriteString("      </table>\n")

	// Risk
	html.WriteString("      <h3>Risk Metrics</h3>\n")
	html.WriteString("      <table>\n")
	html.WriteString("        <tr><th>Metric</th><th>Value</th></tr>\n")
	html.WriteString(fmt.Sprintf("        <tr><td>Max Drawdown</td><td class=\"negative\">%.2f%%</td></tr>\n",
		metrics.MaxDrawdown*100))
	html.WriteString(fmt.Sprintf("        <tr><td>Calmar Ratio</td><td>%.2f</td></tr>\n", metrics.CalmarRatio))
	html.WriteString(fmt.Sprintf("        <tr><td>Avg Drawdown</td><td>%.2f%%</td></tr>\n", metrics.AvgDrawdown*100))
	html.WriteString("      </table>\n")

	// Trade Statistics
	html.WriteString("      <h3>Trade Statistics</h3>\n")
	html.WriteString("      <table>\n")
	html.WriteString("        <tr><th>Metric</th><th>Value</th></tr>\n")
	html.WriteString(fmt.Sprintf("        <tr><td>Win Rate</td><td>%.2f%%</td></tr>\n", metrics.WinRate*100))
	html.WriteString(fmt.Sprintf("        <tr><td>Profit Factor</td><td>%.2f</td></tr>\n", metrics.ProfitFactor))
	html.WriteString(fmt.Sprintf("        <tr><td>Average Trade</td><td>$%.2f</td></tr>\n", metrics.AvgTrade))
	html.WriteString(fmt.Sprintf("        <tr><td>Average Win</td><td class=\"positive\">$%.2f</td></tr>\n", metrics.AvgWin))
	html.WriteString(fmt.Sprintf("        <tr><td>Average Loss</td><td class=\"negative\">$%.2f</td></tr>\n", metrics.AvgLoss))
	html.WriteString("      </table>\n")

	html.WriteString("    </section>\n")

	return html.String()
}

func (g *Generator) generateTradeHistoryHTML(result *backtest.BacktestResult) string {
	var html strings.Builder

	winning := 0
	losing := 0
	for _, trade := range result.TradeHistory {
		if trade.NetPnL > 0 {
			winning++
		} else {
			losing++
		}
	}

	html.WriteString("    <section class=\"trades\">\n")
	html.WriteString("      <h2>Trade History</h2>\n")
	html.WriteString(fmt.Sprintf("      <p><strong>Total Trades:</strong> %d</p>\n", result.TotalTrades))
	html.WriteString(fmt.Sprintf("      <p><strong>Winning Trades:</strong> %d</p>\n", winning))
	html.WriteString(fmt.Sprintf("      <p><strong>Losing Trades:</strong> %d</p>\n", losing))

	html.WriteString("      <h3>Recent Trades (Last 10)</h3>\n")
	html.WriteString("      <table>\n")
	html.WriteString("        <tr><th>Date</th><th>Side</th><th>Exit Reason</th><th>PnL</th><th>Return %</th></tr>\n")

	start := len(result.TradeHistory) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(result.TradeHistory); i++ {
		trade := result.TradeHistory[i]
		pnlClass := "positive"
		if trade.NetPnL < 0 {
			pnlClass = "negative"
		}

		html.WriteString(fmt.Sprintf("        <tr><td>%s</td><td>%s</td><td>%s</td><td class=\"%s\">$%.2f</td><td class=\"%s\">%.2f%%</td></tr>\n",
			trade.EntryTime.Format("2006-01-02"),
			trade.Side,
			trade.ExitReason,
			pnlClass,
			trade.NetPnL,
			pnlClass,
			trade.Return*100))
	}

	html.WriteString("      </table>\n")
	html.WriteString("    </section>\n")

	return html.String()
}

func (g *Generator) getCSS() string {
	if g.config.Theme == "dark" {
		return darkThemeCSS
	}
	return lightThemeCSS
}

const lightThemeCSS = `
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
      line-height: 1.6;
      color: #333;
      background: #f5f5f5;
      padding: 20px;
    }
    .container {
      max-width: 1200px;
      margin: 0 auto;
      background: white;
      padding: 40px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
      border-radius: 8px;
    }
    h1 {
      color: #2c3e50;
      margin-bottom: 10px;
      font-size: 32px;
    }
    h2 {
      color: #34495e;
      margin: 30px 0 20px 0;
      padding-bottom: 10px;
      border-bottom: 2px solid #3498db;
      font-size: 24px;
    }
    h3 {
      color: #7f8c8d;
      margin: 20px 0 10px 0;
      font-size: 18px;
    }
    .generated {
      color: #95a5a6;
      font-style: italic;
      margin-bottom: 20px;
    }
    hr {
      border: none;
      border-top: 1px solid #ecf0f1;
      margin: 30px 0;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      margin: 20px 0;
    }
    th, td {
      text-align: left;
      padding: 12px;
      border-bottom: 1px solid #ecf0f1;
    }
    th {
      background: #3498db;
      color: white;
      font-weight: 600;
    }
    tr:hover {
      background: #f8f9fa;
    }
    .positive {
      color: #27ae60;
      font-weight: 600;
    }
    .negative {
      color: #e74c3c;
      font-weight: 600;
    }
    .highlight {
      font-weight: 700;
      font-size: 18px;
    }
    section {
      margin: 30px 0;
    }
`

const darkThemeCSS = `
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
      line-height: 1.6;
      color: #e0e0e0;
      background: #1a1a1a;
      padding: 20px;
    }
    .container {
      max-width: 1200px;
      margin: 0 auto;
      background: #2d2d2d;
      padding: 40px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.3);
      border-radius: 8px;
    }
    h1 {
      color: #ffffff;
      margin-bottom: 10px;
      font-size: 32px;
    }
    h2 {
      color: #f0f0f0;
      margin: 30px 0 20px 0;
      padding-bottom: 10px;
      border-bottom: 2px solid #3498db;
      font-size: 24px;
    }
    h3 {
      color: #b0b0b0;
      margin: 20px 0 10px 0;
      font-size: 18px;
    }
    .generated {
      color: #808080;
      font-style: italic;
      margin-bottom: 20px;
    }
    hr {
      border: none;
      border-top: 1px solid #404040;
      margin: 30px 0;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      margin: 20px 0;
    }
    th, td {
      text-align: left;
      padding: 12px;
      border-bottom: 1px solid #404040;
    }
    th {
      background: #3498db;
      color: white;
      font-weight: 600;
    }
    tr:hover {
      background: #363636;
    }
    .positive {
      color: #2ecc71;
      font-weight: 600;
    }
    .negative {
      color: #e74c3c;
      font-weight: 600;
    }
    .highlight {
      font-weight: 700;
      font-size: 18px;
    }
    section {
      margin: 30px 0;
    }
`
