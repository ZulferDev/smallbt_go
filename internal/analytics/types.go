package analytics

import (
	"time"

	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// Metrics contains all calculated performance metrics.
type Metrics struct {
	// Returns
	TotalReturn float64 `json:"total_return"`
	CAGR        float64 `json:"cagr"`

	// Risk-adjusted returns
	SharpeRatio  float64 `json:"sharpe_ratio"`
	SortinoRatio float64 `json:"sortino_ratio"`
	CalmarRatio  float64 `json:"calmar_ratio"`

	// Risk metrics
	MaxDrawdown     float64   `json:"max_drawdown"`
	MaxDrawdownDate time.Time `json:"max_drawdown_date"`
	AvgDrawdown     float64   `json:"avg_drawdown"`

	// Trade statistics
	TotalTrades   int     `json:"total_trades"`
	WinningTrades int     `json:"winning_trades"`
	LosingTrades  int     `json:"losing_trades"`
	WinRate       float64 `json:"win_rate"`

	// PnL statistics
	GrossProfit  float64 `json:"gross_profit"`
	GrossLoss    float64 `json:"gross_loss"`
	NetProfit    float64 `json:"net_profit"`
	ProfitFactor float64 `json:"profit_factor"`

	// Average metrics
	AvgTrade float64 `json:"avg_trade"`
	AvgWin   float64 `json:"avg_win"`
	AvgLoss  float64 `json:"avg_loss"`

	// Best/Worst
	LargestWin  float64 `json:"largest_win"`
	LargestLoss float64 `json:"largest_loss"`

	// Expectancy
	Expectancy float64 `json:"expectancy"`

	// Exposure
	AvgExposure float64 `json:"avg_exposure"`

	// Fees
	TotalFees float64 `json:"total_fees"`
}

// EquityPoint represents a point in the equity curve.
type EquityPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Equity    float64   `json:"equity"`
	Cash      float64   `json:"cash"`
	Drawdown  float64   `json:"drawdown"`
	Exposure  float64   `json:"exposure"`
}

// AnalysisInput contains all data needed for analysis.
type AnalysisInput struct {
	InitialCash  float64
	FinalEquity  float64
	StartTime    time.Time
	EndTime      time.Time
	TradeHistory []portfolio.Trade
	EquityCurve  []EquityPoint
	RiskFreeRate float64 // Annual risk-free rate (e.g., 0.02 for 2%)
}

// Analyzer is the interface for calculating metrics.
type Analyzer interface {
	Analyze(input AnalysisInput) *Metrics
}
