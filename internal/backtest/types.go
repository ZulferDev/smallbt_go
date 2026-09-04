package backtest

import (
	"time"

	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/portfolio"
)

// BacktestConfig holds configuration for a backtest.
type BacktestConfig struct {
	Symbol       market.Symbol
	Timeframe    market.Timeframe
	StartTime    time.Time
	EndTime      time.Time
	InitialCash  float64
	StrategyPath string
	DataPath     string
}

// BacktestResult holds the results of a backtest.
type BacktestResult struct {
	Config       BacktestConfig
	Portfolio    *portfolio.Portfolio
	TotalTrades  int
	StartTime    time.Time
	EndTime      time.Time
	Metrics      map[string]float64
	TradeHistory []portfolio.Trade
	EquityCurve  []EquityPoint
}

// EquityPoint represents a point on the equity curve.
type EquityPoint struct {
	Timestamp time.Time
	Equity    float64
	Cash      float64
	Drawdown  float64
	Exposure  float64
}

// BacktestEngine is the core backtesting engine interface.
type BacktestEngine interface {
	Run(config BacktestConfig) (*BacktestResult, error)
}
