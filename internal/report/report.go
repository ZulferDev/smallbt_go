package report

import (
	"fmt"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/analytics"
	"github.com/ZulferDev/smallbt_go/internal/backtest"
)

// ReportFormat defines the output format for reports
type ReportFormat string

const (
	FormatHTML     ReportFormat = "html"
	FormatMarkdown ReportFormat = "markdown"
	FormatText     ReportFormat = "text"
)

// ReportConfig configures report generation
type ReportConfig struct {
	// Output format
	Format ReportFormat

	// Title of the report
	Title string

	// Include sections
	IncludeSummary      bool
	IncludeMetrics      bool
	IncludeTradeHistory bool
	IncludeEquityCurve  bool
	IncludeDrawdown     bool
	IncludeMonthly      bool

	// Styling options for HTML
	IncludeCSS bool
	Theme      string // "light" or "dark"
}

// DefaultReportConfig returns default configuration
func DefaultReportConfig() ReportConfig {
	return ReportConfig{
		Format:              FormatHTML,
		Title:               "Backtest Report",
		IncludeSummary:      true,
		IncludeMetrics:      true,
		IncludeTradeHistory: true,
		IncludeEquityCurve:  true,
		IncludeDrawdown:     true,
		IncludeMonthly:      true,
		IncludeCSS:          true,
		Theme:               "light",
	}
}

// Report represents a generated backtest report
type Report struct {
	Config ReportConfig
	Result *backtest.BacktestResult
	
	// Generated content
	Content string
	
	// Generation metadata
	GeneratedAt time.Time
}

// Generator generates reports from backtest results
type Generator struct {
	config ReportConfig
}

// NewGenerator creates a new report generator
func NewGenerator(config ReportConfig) *Generator {
	return &Generator{
		config: config,
	}
}

// Generate creates a report from backtest results
func (g *Generator) Generate(result *backtest.BacktestResult) (*Report, error) {
	if result == nil {
		return nil, fmt.Errorf("backtest result is nil")
	}

	generatedAt := time.Now()
	report := &Report{
		Config:      g.config,
		Result:      result,
		GeneratedAt: generatedAt,
	}

	switch g.config.Format {
	case FormatHTML:
		content, err := g.generateHTML(result)
		if err != nil {
			return nil, fmt.Errorf("generate HTML: %w", err)
		}
		report.Content = content

	case FormatMarkdown:
		content, err := g.generateMarkdown(result)
		if err != nil {
			return nil, fmt.Errorf("generate Markdown: %w", err)
		}
		report.Content = content

	case FormatText:
		content, err := g.generateText(result)
		if err != nil {
			return nil, fmt.Errorf("generate Text: %w", err)
		}
		report.Content = content

	default:
		return nil, fmt.Errorf("unsupported format: %s", g.config.Format)
	}

	return report, nil
}

// generateText creates a plain text report
func (g *Generator) generateText(result *backtest.BacktestResult) (string, error) {
	text := ""

	// Header
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	text += fmt.Sprintf("%s\n", g.config.Title)
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"

	// Summary
	if g.config.IncludeSummary {
		text += g.generateSummaryText(result)
		text += "\n"
	}

	// Metrics
	if g.config.IncludeMetrics && result.Metrics != nil {
		text += g.generateMetricsText(result.Metrics)
		text += "\n"
	}

	// Trade History Summary
	if g.config.IncludeTradeHistory && len(result.TradeHistory) > 0 {
		text += g.generateTradeHistorySummaryText(result)
		text += "\n"
	}

	// Footer
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	text += fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

	return text, nil
}

func (g *Generator) generateSummaryText(result *backtest.BacktestResult) string {
	text := "STRATEGY SUMMARY\n"
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	text += fmt.Sprintf("Strategy:       %s\n", result.StrategyName)
	text += fmt.Sprintf("Symbol:         %s\n", result.Config.Symbol)
	text += fmt.Sprintf("Timeframe:      %s\n", result.Config.Timeframe)
	text += fmt.Sprintf("Period:         %s to %s\n", 
		result.Config.StartTime.Format("2006-01-02"),
		result.Config.EndTime.Format("2006-01-02"))
	text += fmt.Sprintf("Initial Cash:   $%.2f\n", result.Config.InitialCash)
	
	finalEquity := result.Config.InitialCash
	if len(result.EquityCurve) > 0 {
		finalEquity = result.EquityCurve[len(result.EquityCurve)-1].Equity
	}
	text += fmt.Sprintf("Final Equity:   $%.2f\n", finalEquity)
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	return text
}

func (g *Generator) generateMetricsText(metrics *analytics.Metrics) string {
	text := "PERFORMANCE METRICS\n"
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	
	text += "\nReturns:\n"
	text += fmt.Sprintf("  Total Return:        %+.2f%%\n", metrics.TotalReturn*100)
	text += fmt.Sprintf("  CAGR:                %+.2f%%\n", metrics.CAGR*100)
	text += fmt.Sprintf("  Sharpe Ratio:        %.2f\n", metrics.SharpeRatio)
	text += fmt.Sprintf("  Sortino Ratio:       %.2f\n", metrics.SortinoRatio)
	
	text += "\nRisk:\n"
	text += fmt.Sprintf("  Max Drawdown:        %.2f%%\n", metrics.MaxDrawdown*100)
	text += fmt.Sprintf("  Calmar Ratio:        %.2f\n", metrics.CalmarRatio)
	text += fmt.Sprintf("  Avg Drawdown:        %.2f%%\n", metrics.AvgDrawdown*100)
	
	text += "\nTrades:\n"
	text += fmt.Sprintf("  Win Rate:            %.2f%%\n", metrics.WinRate*100)
	text += fmt.Sprintf("  Profit Factor:       %.2f\n", metrics.ProfitFactor)
	text += fmt.Sprintf("  Average Trade:       $%.2f\n", metrics.AvgTrade)
	text += fmt.Sprintf("  Average Win:         $%.2f\n", metrics.AvgWin)
	text += fmt.Sprintf("  Average Loss:        $%.2f\n", metrics.AvgLoss)
	
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	return text
}

func (g *Generator) generateTradeHistorySummaryText(result *backtest.BacktestResult) string {
	text := "TRADE HISTORY\n"
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	text += fmt.Sprintf("Total Trades:        %d\n", result.TotalTrades)
	
	winning := 0
	losing := 0
	for _, trade := range result.TradeHistory {
		if trade.NetPnL > 0 {
			winning++
		} else {
			losing++
		}
	}
	
	text += fmt.Sprintf("Winning Trades:      %d\n", winning)
	text += fmt.Sprintf("Losing Trades:       %d\n", losing)
	
	if len(result.TradeHistory) > 0 {
		text += "\nRecent Trades (last 10):\n"
		start := len(result.TradeHistory) - 10
		if start < 0 {
			start = 0
		}
		
		for i := start; i < len(result.TradeHistory); i++ {
			trade := result.TradeHistory[i]
			text += fmt.Sprintf("  %s: %s → %s | PnL: $%.2f (%.2f%%)\n",
				trade.EntryTime.Format("2006-01-02"),
				trade.Side,
				trade.ExitReason,
				trade.NetPnL,
				trade.Return*100)
		}
	}
	
	text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	return text
}
