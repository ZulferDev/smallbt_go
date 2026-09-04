package backtest

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/1jehuang/backtest/internal/data/csv"
	"github.com/1jehuang/backtest/internal/data/parquet"
	"github.com/1jehuang/backtest/internal/data/resample"
	"github.com/1jehuang/backtest/internal/execution"
	"github.com/1jehuang/backtest/internal/indicator"
	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/portfolio"
	"github.com/1jehuang/backtest/internal/risk"
	"github.com/1jehuang/backtest/internal/signal"
	"github.com/1jehuang/backtest/internal/strategy/ast"
	"github.com/1jehuang/backtest/internal/strategy/evaluator"
	"github.com/1jehuang/backtest/internal/strategy/parser"
)

// Run runs a backtest with the given configuration.
func Run(config BacktestConfig) (*BacktestResult, error) {
	// Step 1: Load strategy
	strategyAST, err := loadStrategy(config.StrategyPath)
	if err != nil {
		return nil, fmt.Errorf("load strategy: %w", err)
	}

	// Step 2: Load and prepare data
	candles, err := loadCandles(config.DataPath, config.Timeframe)
	if err != nil {
		return nil, fmt.Errorf("load candles: %w", err)
	}

	// Filter candles by date range
	if !config.StartTime.IsZero() {
		candles = filterCandlesByTime(candles, config.StartTime, time.Time{})
	}
	if !config.EndTime.IsZero() {
		candles = filterCandlesByTime(candles, time.Time{}, config.EndTime)
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("no candles after filtering")
	}

	// Step 3: Initialize components
	registry := indicator.NewRegistry()
	registry = indicator.BuiltinRegistry()

	// Create strategy evaluator
	evaluator := evaluator.NewEvaluator(strategyAST, registry, config.Symbol, config.Timeframe)
	if err := evaluator.Initialize(); err != nil {
		return nil, fmt.Errorf("initialize evaluator: %w", err)
	}

	// Create portfolio
	portfolio := portfolio.NewPortfolio(config.InitialCash)

	// Create risk manager
	riskConfig := risk.Config{
		MaxTradesPerDay:     10,
		MaxExposurePercent:  1.0, // 100%
		MaxDailyLossPercent: 0.05,
		MaxDrawdownPercent:  0.20,
	}
	riskManager := risk.NewManager(riskConfig)

	// Create execution config
	execConfig := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005, // 0.05% slippage
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Spread:        0.0001,
		Seed:          42, // Fixed seed for reproducibility
	}
	executor := execution.NewSimpleExecutor(execConfig)

	// Step 4: Run backtest loop
	tradeHistory, equityCurve, result := runBacktestLoop(
		candles, evaluator, portfolio, riskManager, executor, config,
	)

	// Step 5: Calculate analytics
	metrics := calculateMetrics(result)

	// Step 6: Finalize result
	return &BacktestResult{
		Config:       config,
		Portfolio:    portfolio,
		TotalTrades:  len(tradeHistory),
		StartTime:    candles[0].Timestamp,
		EndTime:      candles[len(candles)-1].Timestamp,
		Metrics:      metrics,
		TradeHistory: tradeHistory,
		EquityCurve:  equityCurve,
	}, nil
}

func loadStrategy(path string) (*ast.Strategy, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Parse YAML
	p := parser.NewParser()
	strategyAST, err := p.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse strategy: %w", err)
	}

	return strategyAST, nil
}

func loadCandles(path string, tf market.Timeframe) ([]*market.Candle, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Determine file type and load
	ext := filepath.Ext(absPath)
	var candles []market.Candle

	switch ext {
	case ".csv":
		feed, err := csv.NewCSVFeed(absPath, csv.DefaultCSVConfig(market.Symbol(""), tf))
		if err != nil {
			return nil, fmt.Errorf("load csv: %w", err)
		}
		candles = feed.GetCandles()
	case ".parquet":
		reader := parquet.NewReader(absPath)
		candles, err = reader.Read()
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("load candles: %w", err)
	}

	// Convert []market.Candle to []*market.Candle
	result := make([]*market.Candle, len(candles))
	for i := range candles {
		result[i] = &candles[i]
	}

	// Resample if needed
	if tf != "" && tf != market.Timeframe1d {
		result, err = resample.Resample(result, tf)
		if err != nil {
			return nil, fmt.Errorf("resample: %w", err)
		}
	}

	return result, nil
}

func filterCandlesByTime(candles []*market.Candle, start, end time.Time) []*market.Candle {
	var result []*market.Candle
	for _, c := range candles {
		if !start.IsZero() && c.Timestamp.Before(start) {
			continue
		}
		if !end.IsZero() && !c.Timestamp.Before(end) {
			continue
		}
		result = append(result, c)
	}
	return result
}

type backtestState struct {
	hasPosition  bool
	positionSide portfolio.PositionSide
	currentTrade *portfolio.Trade
	entryPrice   float64
	entryTime    time.Time
	stopLoss     *float64
	takeProfit   *float64
}

func runBacktestLoop(
	candles []*market.Candle,
	evaluator *evaluator.Evaluator,
	portfolioInstance *portfolio.Portfolio,
	riskManager *risk.Manager,
	executor *execution.SimpleExecutor,
	config BacktestConfig,
) ([]portfolio.Trade, []EquityPoint, *BacktestResult) {
	var trades []portfolio.Trade
	var equityCurve []EquityPoint
	var state backtestState

	for _, candle := range candles {
		executor.SetCurrentCandle(candle)

		// Evaluate strategy
		signals, err := evaluator.Evaluate(state.hasPosition, string(state.positionSide))
		if err != nil {
			continue
		}

		// Process signals
		for _, sig := range signals {
			switch sig.Type {
			case signal.SignalTypeLongEntry:
				// Create position
				portfolioInstance.OpenPosition(
					config.Symbol,
					portfolio.PositionSideLong,
					sig.Price,
					candle.Timestamp,
				)
				state.hasPosition = true
				state.positionSide = portfolio.PositionSideLong
				state.entryPrice = sig.Price
				state.entryTime = candle.Timestamp
				state.stopLoss, state.takeProfit = nil, nil

			case signal.SignalTypeShortEntry:
				// Create position
				portfolioInstance.OpenPosition(
					config.Symbol,
					portfolio.PositionSideShort,
					sig.Price,
					candle.Timestamp,
				)
				state.hasPosition = true
				state.positionSide = portfolio.PositionSideShort
				state.entryPrice = sig.Price
				state.entryTime = candle.Timestamp
				state.stopLoss, state.takeProfit = nil, nil

			case signal.SignalTypeLongExit:
				// Close position
				trade := portfolioInstance.ClosePosition(
					config.Symbol,
					sig.Price,
					candle.Timestamp,
				)
				trades = append(trades, *trade)
				state.hasPosition = false
				state.positionSide = ""

			case signal.SignalTypeShortExit:
				// Close position
				trade := portfolioInstance.ClosePosition(
					config.Symbol,
					sig.Price,
					candle.Timestamp,
				)
				trades = append(trades, *trade)
				state.hasPosition = false
				state.positionSide = ""
			}
		}

		// Update portfolio with current candle
		portfolioInstance.UpdateWithCandle(config.Symbol, candle.Close, candle.Timestamp)

		// Record equity point
		equityCurve = append(equityCurve, EquityPoint{
			Timestamp: candle.Timestamp,
			Equity:    portfolioInstance.Equity,
			Cash:      portfolioInstance.Cash,
			Drawdown:  calculateDrawdown(portfolioInstance),
			Exposure:  portfolioInstance.GetExposure(),
		})
	}

	return trades, equityCurve, nil
}

func calculateDrawdown(portfolioInstance *portfolio.Portfolio) float64 {
	if portfolioInstance.InitialCash == 0 {
		return 0
	}
	return (portfolioInstance.Equity - portfolioInstance.InitialCash) / portfolioInstance.InitialCash
}

func calculateMetrics(result *BacktestResult) map[string]float64 {
	metrics := make(map[string]float64)

	if len(result.TradeHistory) == 0 {
		metrics["total_return"] = 0
		metrics["cagr"] = 0
		metrics["sharpe_ratio"] = 0
		metrics["sortino_ratio"] = 0
		metrics["max_drawdown"] = 0
		metrics["win_rate"] = 0
		metrics["profit_factor"] = 0
		metrics["expectancy"] = 0
		metrics["average_trade"] = 0
		metrics["average_win"] = 0
		metrics["average_loss"] = 0
		metrics["win_count"] = 0
		metrics["loss_count"] = 0
		return metrics
	}

	// Total return
	totalReturn := (result.Portfolio.Equity - result.Portfolio.InitialCash) / result.Portfolio.InitialCash
	metrics["total_return"] = totalReturn

	// CAGR
	days := result.EndTime.Sub(result.StartTime).Hours() / 24
	if days > 0 {
		metrics["cagr"] = pow(1+totalReturn, 365/days) - 1
	}

	// Trade statistics
	var totalWin, totalLoss float64
	winCount, lossCount := 0, 0
	for _, trade := range result.TradeHistory {
		if trade.NetPnL > 0 {
			totalWin += trade.NetPnL
			winCount++
		} else {
			totalLoss += trade.NetPnL
			lossCount++
		}
	}

	metrics["win_count"] = float64(winCount)
	metrics["loss_count"] = float64(lossCount)
	metrics["win_rate"] = float64(winCount) / float64(len(result.TradeHistory))

	if winCount > 0 {
		metrics["average_win"] = totalWin / float64(winCount)
	}
	if lossCount > 0 {
		metrics["average_loss"] = totalLoss / float64(lossCount)
	}
	if len(result.TradeHistory) > 0 {
		metrics["average_trade"] = (totalWin + totalLoss) / float64(len(result.TradeHistory))
	}

	// Profit factor
	if totalLoss != 0 {
		metrics["profit_factor"] = totalWin / -totalLoss
	} else {
		metrics["profit_factor"] = totalWin
	}

	// Expectancy
	if len(result.TradeHistory) > 0 {
		var expectancy float64
		for _, trade := range result.TradeHistory {
			expectancy += trade.NetPnL
		}
		metrics["expectancy"] = expectancy / float64(len(result.TradeHistory))
	}

	// Max drawdown
	maxDD := 0.0
	for _, point := range result.EquityCurve {
		if point.Drawdown < maxDD {
			maxDD = point.Drawdown
		}
	}
	metrics["max_drawdown"] = maxDD

	// Sharpe ratio (simplified)
	if len(result.EquityCurve) > 1 {
		var returns []float64
		for i := 1; i < len(result.EquityCurve); i++ {
			prev := result.EquityCurve[i-1].Equity
			curr := result.EquityCurve[i].Equity
			if prev > 0 {
				returns = append(returns, (curr-prev)/prev)
			}
		}
		metrics["sharpe_ratio"] = calculateSharpe(returns)
	}

	return metrics
}

func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}

func calculateSharpe(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Calculate mean
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// Calculate std dev
	var variance float64
	for _, r := range returns {
		diff := r - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))
	stdDev := sqrt(variance)

	if stdDev == 0 {
		return 0
	}

	// Sharpe ratio (assuming 252 trading days, 0 risk-free rate)
	return mean / stdDev * sqrt(252)
}

func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	result := x
	for i := 0; i < 10; i++ {
		result = (result + x/result) / 2
	}
	return result
}
