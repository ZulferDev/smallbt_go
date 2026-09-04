package backtest

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/1jehuang/backtest/internal/analytics"
	"github.com/1jehuang/backtest/internal/broker"
	"github.com/1jehuang/backtest/internal/data/csv"
	"github.com/1jehuang/backtest/internal/data/parquet"
	"github.com/1jehuang/backtest/internal/data/resample"
	"github.com/1jehuang/backtest/internal/execution"
	"github.com/1jehuang/backtest/internal/indicator"
	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/order"
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

	// Create risk calculators from strategy config
	var positionSizer *risk.PositionSizer
	var stopLossCalc *risk.StopLossCalculator
	var takeProfitCalc *risk.TakeProfitCalculator
	var trailingStopCalc *risk.TrailingStopCalculator

	if strategyAST.Risk.PositionSize.Type != "" {
		// Position sizer
		positionSizer = risk.NewPositionSizer(strategyAST.Risk.PositionSize)
	}

	// Stop loss (pointer check)
	if strategyAST.Risk.StopLoss != nil {
		stopLossCalc = risk.NewStopLossCalculator(strategyAST.Risk.StopLoss)
	}

	// Take profit (pointer check)
	if strategyAST.Risk.TakeProfit != nil {
		takeProfitCalc = risk.NewTakeProfitCalculator(strategyAST.Risk.TakeProfit)
	}

	// Trailing stop
	if strategyAST.Risk.TrailingStop != nil {
		trailingStopCalc = risk.NewTrailingStopCalculator(strategyAST.Risk.TrailingStop)
	}

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

	// Create broker
	brokerInstance := broker.NewBroker(executor)

	// Step 4: Run backtest loop
	tradeHistory, equityCurve, err := runBacktestLoop(
		candles, evaluator, portfolio, riskManager, executor, brokerInstance, config, strategyAST,
		positionSizer, stopLossCalc, takeProfitCalc, trailingStopCalc,
	)
	if err != nil {
		return nil, fmt.Errorf("backtest loop: %w", err)
	}

	// Step 5: Calculate analytics
	result := &BacktestResult{
		Config:       config,
		Portfolio:    portfolio,
		TotalTrades:  len(tradeHistory),
		StartTime:    candles[0].Timestamp,
		EndTime:      candles[len(candles)-1].Timestamp,
		TradeHistory: tradeHistory,
		EquityCurve:  equityCurve,
	}
	result.Metrics = calculateMetrics(result)

	return result, nil
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
	brokerInstance *broker.Broker,
	config BacktestConfig,
	strategyAST *ast.Strategy,
	positionSizer *risk.PositionSizer,
	stopLossCalc *risk.StopLossCalculator,
	takeProfitCalc *risk.TakeProfitCalculator,
	trailingStopCalc *risk.TrailingStopCalculator,
) ([]portfolio.Trade, []EquityPoint, error) {
	var trades []portfolio.Trade
	var equityCurve []EquityPoint
	state := backtestState{
		hasPosition:  false,
		positionSide: "",
		entryPrice:   0,
		entryTime:    time.Time{},
		stopLoss:     nil,
		takeProfit:   nil,
	}

	for _, candle := range candles {
		executor.SetCurrentCandle(candle)

		// Process pending orders from previous candles
		filledOrders, err := brokerInstance.ProcessPendingOrders(candle, candle.Timestamp)
		if err != nil {
			continue
		}

		// Handle filled orders
		for _, orderID := range filledOrders {
			ord, _ := brokerInstance.GetOrder(orderID)
			if ord == nil {
				continue
			}

			// Update portfolio based on filled order
			if ord.Side == order.OrderSideBuy {
				if !state.hasPosition {
					// Open long position
					err := portfolioInstance.OpenPosition(
						config.Symbol,
						portfolio.PositionSideLong,
						ord.Quantity,
						ord.FilledPrice,
						candle.Timestamp,
					)
					if err != nil {
						// Log error but continue - this shouldn't happen with proper capital
						continue
					}
					state.hasPosition = true
					state.positionSide = portfolio.PositionSideLong
					state.entryPrice = ord.FilledPrice
					state.entryTime = candle.Timestamp

					// Calculate stop loss and take profit
					if stopLossCalc != nil {
						// Get ATR value if needed
						atrValue := 0.0
						if strategyAST.Risk.StopLoss != nil && strategyAST.Risk.StopLoss.Type == "atr" && strategyAST.Risk.StopLoss.Indicator != "" {
							indicatorValues := evaluator.GetIndicatorValues()
							if val, ok := indicatorValues[strategyAST.Risk.StopLoss.Indicator]; ok {
								atrValue = val
							}
						}
						slSide := "long"
						if state.positionSide == portfolio.PositionSideShort {
							slSide = "short"
						}
						sl, err := stopLossCalc.Calculate(ord.FilledPrice, slSide, atrValue)
						if err != nil {
							return nil, nil, fmt.Errorf("calculate stop loss: %w", err)
						}
						if sl > 0 {
							state.stopLoss = &sl
						}
					}
					if takeProfitCalc != nil {
						// Get stop loss price for risk/reward calculation
						slPrice := 0.0
						if state.stopLoss != nil {
							slPrice = *state.stopLoss
						}
						tpSide := "long"
						if state.positionSide == portfolio.PositionSideShort {
							tpSide = "short"
						}
						tp, err := takeProfitCalc.Calculate(ord.FilledPrice, tpSide, slPrice)
						if err != nil {
							return nil, nil, fmt.Errorf("calculate take profit: %w", err)
						}
						if tp > 0 {
							state.takeProfit = &tp
						}
					}
				}
			} else if ord.Side == order.OrderSideSell {
				if state.hasPosition && state.positionSide == portfolio.PositionSideLong {
					// Close long position
					trade := portfolioInstance.ClosePosition(
						config.Symbol,
						ord.FilledPrice,
						candle.Timestamp,
					)
					trades = append(trades, *trade)
					state.hasPosition = false
					state.positionSide = ""
					state.stopLoss = nil
					state.takeProfit = nil
				} else if !state.hasPosition {
					// Open short position
					err := portfolioInstance.OpenPosition(
						config.Symbol,
						portfolio.PositionSideShort,
						ord.Quantity,
						ord.FilledPrice,
						candle.Timestamp,
					)
					if err != nil {
						// Log error but continue - this shouldn't happen with proper capital
						continue
					}
					state.hasPosition = true
					state.positionSide = portfolio.PositionSideShort
					state.entryPrice = ord.FilledPrice
					state.entryTime = candle.Timestamp

					// Calculate stop loss and take profit
					if stopLossCalc != nil {
						// Get ATR value if needed
						atrValue := 0.0
						if strategyAST.Risk.StopLoss != nil && strategyAST.Risk.StopLoss.Type == "atr" && strategyAST.Risk.StopLoss.Indicator != "" {
							indicatorValues := evaluator.GetIndicatorValues()
							if val, ok := indicatorValues[strategyAST.Risk.StopLoss.Indicator]; ok {
								atrValue = val
							}
						}
						slSide := "long"
						if state.positionSide == portfolio.PositionSideShort {
							slSide = "short"
						}
						sl, err := stopLossCalc.Calculate(ord.FilledPrice, slSide, atrValue)
						if err != nil {
							return nil, nil, fmt.Errorf("calculate stop loss: %w", err)
						}
						if sl > 0 {
							state.stopLoss = &sl
						}
					}
					if takeProfitCalc != nil {
						// Get stop loss price for risk/reward calculation
						slPrice := 0.0
						if state.stopLoss != nil {
							slPrice = *state.stopLoss
						}
						tpSide := "long"
						if state.positionSide == portfolio.PositionSideShort {
							tpSide = "short"
						}
						tp, err := takeProfitCalc.Calculate(ord.FilledPrice, tpSide, slPrice)
						if err != nil {
							return nil, nil, fmt.Errorf("calculate take profit: %w", err)
						}
						if tp > 0 {
							state.takeProfit = &tp
						}
					}
				}
			}
		}

		// Update evaluator with current candle FIRST
		// This ensures indicators are calculated before evaluating conditions
		evaluator.UpdateCandle(*candle)

		// Evaluate strategy
		signals, err := evaluator.Evaluate(state.hasPosition, string(state.positionSide))
		if err != nil {
			continue
		}

		// Process signals and submit orders
		for _, sig := range signals {
			switch sig.Type {
			case signal.SignalTypeLongEntry:
				if !state.hasPosition {
					// Calculate position size
					quantity := 1.0
					var stopDist float64
					var sl float64
					if stopLossCalc != nil {
						// Get ATR value if needed
						atrValue := 0.0
						if strategyAST.Risk.StopLoss != nil && strategyAST.Risk.StopLoss.Type == "atr" && strategyAST.Risk.StopLoss.Indicator != "" {
							indicatorValues := evaluator.GetIndicatorValues()
							if val, ok := indicatorValues[strategyAST.Risk.StopLoss.Indicator]; ok {
								atrValue = val
							}
						}
						// Estimate stop distance for position sizing
						var err error
						sl, err = stopLossCalc.Calculate(candle.Close, "long", atrValue)
						if err == nil && sl > 0 {
							stopDist = absPrice(candle.Close - sl)
						}
					}
					if positionSizer != nil {
						slPrice := 0.0
						if stopDist > 0 {
							if candle.Close > sl {
								slPrice = sl
							} else {
								slPrice = candle.Close - stopDist
							}
						}
						qty, err := positionSizer.CalculateQuantity(candle.Close, portfolioInstance.Equity, slPrice)
						if err == nil && qty > 0 {
							quantity = qty
						}
					}

					// Submit market buy order
					req := order.OrderRequest{
						Symbol:   config.Symbol,
						Side:     order.OrderSideBuy,
						Type:     order.OrderTypeMarket,
						Quantity: quantity,
					}
					_, err := brokerInstance.SubmitOrder(req, candle.Timestamp)
					if err != nil {
						continue
					}
				}

			case signal.SignalTypeShortEntry:
				if !state.hasPosition {
					// Calculate position size
					quantity := 1.0
					var stopDist float64
					if stopLossCalc != nil {
						// Get ATR value if needed
						atrValue := 0.0
						if strategyAST.Risk.StopLoss != nil && strategyAST.Risk.StopLoss.Type == "atr" && strategyAST.Risk.StopLoss.Indicator != "" {
							indicatorValues := evaluator.GetIndicatorValues()
							if val, ok := indicatorValues[strategyAST.Risk.StopLoss.Indicator]; ok {
								atrValue = val
							}
						}
						// Estimate stop distance for position sizing
						sl, err := stopLossCalc.Calculate(candle.Close, "short", atrValue)
						if err == nil && sl > 0 {
							stopDist = absPrice(sl - candle.Close)
						}
					}
					if positionSizer != nil {
						slPrice := 0.0
						if stopDist > 0 {
							slPrice = candle.Close + stopDist
						}
						qty, err := positionSizer.CalculateQuantity(candle.Close, portfolioInstance.Equity, slPrice)
						if err == nil && qty > 0 {
							quantity = qty
						}
					}

					// Submit market sell order
					req := order.OrderRequest{
						Symbol:   config.Symbol,
						Side:     order.OrderSideSell,
						Type:     order.OrderTypeMarket,
						Quantity: quantity,
					}
					_, err := brokerInstance.SubmitOrder(req, candle.Timestamp)
					if err != nil {
						continue
					}
				}

			case signal.SignalTypeLongExit:
				if state.hasPosition && state.positionSide == portfolio.PositionSideLong {
					// Submit market sell order to close
					req := order.OrderRequest{
						Symbol:   config.Symbol,
						Side:     order.OrderSideSell,
						Type:     order.OrderTypeMarket,
						Quantity: 1.0,
					}
					_, err := brokerInstance.SubmitOrder(req, candle.Timestamp)
					if err != nil {
						continue
					}
				}

			case signal.SignalTypeShortExit:
				if state.hasPosition && state.positionSide == portfolio.PositionSideShort {
					// Submit market buy order to close
					req := order.OrderRequest{
						Symbol:   config.Symbol,
						Side:     order.OrderSideBuy,
						Type:     order.OrderTypeMarket,
						Quantity: 1.0,
					}
					_, err := brokerInstance.SubmitOrder(req, candle.Timestamp)
					if err != nil {
						continue
					}
				}
			}
		}

		// Check stop loss and take profit if position is open
		if state.hasPosition {
			if state.stopLoss != nil {
				if state.positionSide == portfolio.PositionSideLong && candle.Low <= *state.stopLoss {
					// Stop loss hit for long position
					req := order.OrderRequest{
						Symbol:    config.Symbol,
						Side:      order.OrderSideSell,
						Type:      order.OrderTypeStop,
						Quantity:  1.0,
						StopPrice: state.stopLoss,
					}
					brokerInstance.SubmitOrder(req, candle.Timestamp)
				} else if state.positionSide == portfolio.PositionSideShort && candle.High >= *state.stopLoss {
					// Stop loss hit for short position
					req := order.OrderRequest{
						Symbol:    config.Symbol,
						Side:      order.OrderSideBuy,
						Type:      order.OrderTypeStop,
						Quantity:  1.0,
						StopPrice: state.stopLoss,
					}
					brokerInstance.SubmitOrder(req, candle.Timestamp)
				}
			}

			if state.takeProfit != nil {
				if state.positionSide == portfolio.PositionSideLong && candle.High >= *state.takeProfit {
					// Take profit hit for long position
					req := order.OrderRequest{
						Symbol:   config.Symbol,
						Side:     order.OrderSideSell,
						Type:     order.OrderTypeLimit,
						Quantity: 1.0,
						Price:    state.takeProfit,
					}
					brokerInstance.SubmitOrder(req, candle.Timestamp)
				} else if state.positionSide == portfolio.PositionSideShort && candle.Low <= *state.takeProfit {
					// Take profit hit for short position
					req := order.OrderRequest{
						Symbol:   config.Symbol,
						Side:     order.OrderSideBuy,
						Type:     order.OrderTypeLimit,
						Quantity: 1.0,
						Price:    state.takeProfit,
					}
					brokerInstance.SubmitOrder(req, candle.Timestamp)
				}
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

func calculateMetrics(result *BacktestResult) *analytics.Metrics {
	analyzer := analytics.NewAnalyzer()

	// Convert backtest.EquityPoint to analytics.EquityPoint
	equityCurve := make([]analytics.EquityPoint, len(result.EquityCurve))
	for i, ep := range result.EquityCurve {
		equityCurve[i] = analytics.EquityPoint{
			Timestamp: ep.Timestamp,
			Equity:    ep.Equity,
			Cash:      ep.Cash,
			Drawdown:  ep.Drawdown,
			Exposure:  ep.Exposure,
		}
	}

	input := analytics.AnalysisInput{
		StartTime:    result.StartTime,
		EndTime:      result.EndTime,
		InitialCash:  result.Portfolio.InitialCash,
		FinalEquity:  result.Portfolio.Equity,
		EquityCurve:  equityCurve,
		TradeHistory: result.TradeHistory,
	}

	return analyzer.Analyze(input)
}

// absPrice returns absolute value for price calculations
func absPrice(value float64) float64 {
	return math.Abs(value)
}
