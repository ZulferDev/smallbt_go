package backtest

import (
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/analytics"
	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/order"
	"github.com/1jehuang/backtest/internal/portfolio"
	"github.com/stretchr/testify/assert"
)

// SimpleTestStrategy is a minimal test strategy for integration testing
type SimpleTestStrategy struct {
	Symbol    string
	BuyPrice  float64
	SellPrice float64
}

func (s *SimpleTestStrategy) Evaluate(candle *market.Candle, state *portfolio.PositionSide) (*order.OrderRequest, error) {
	if candle.Close > s.BuyPrice && *state == "" {
		// Entry signal
		return &order.OrderRequest{
			Symbol:   market.Symbol(s.Symbol),
			Side:     order.OrderSideBuy,
			Type:     order.OrderTypeMarket,
			Quantity: 0.1, // Small position
		}, nil
	}

	if candle.Close < s.SellPrice && *state == portfolio.PositionSideLong {
		// Exit signal
		return &order.OrderRequest{
			Symbol:   market.Symbol(s.Symbol),
			Side:     order.OrderSideSell,
			Type:     order.OrderTypeMarket,
			Quantity: 0.1, // Close full position
		}, nil
	}

	return nil, nil
}

func TestBacktestTypes(t *testing.T) {
	// Test BacktestConfig
	config := BacktestConfig{
		Symbol:       market.Symbol("BTCUSDT"),
		Timeframe:    market.Timeframe("1h"),
		InitialCash:  10000.0,
		StrategyPath: "strategy.yaml",
		DataPath:     "data.parquet",
	}

	assert.Equal(t, market.Symbol("BTCUSDT"), config.Symbol)
	assert.Equal(t, market.Timeframe("1h"), config.Timeframe)
	assert.Equal(t, 10000.0, config.InitialCash)

	// Test BacktestResult
	result := &BacktestResult{
		Config:       config,
		TotalTrades:  5,
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Hour),
		Metrics:      &analytics.Metrics{SharpeRatio: 1.5, WinRate: 0.6},
		TradeHistory: []portfolio.Trade{},
		EquityCurve:  []EquityPoint{},
	}

	assert.Equal(t, 5, result.TotalTrades)
	assert.Equal(t, 1.5, result.Metrics.SharpeRatio)
	assert.Equal(t, 0.6, result.Metrics.WinRate)

	// Test EquityPoint
	ep := EquityPoint{
		Timestamp: time.Now(),
		Equity:    10500.0,
		Cash:      5000.0,
		Drawdown:  -0.05,
		Exposure:  0.5,
	}

	assert.Equal(t, 10500.0, ep.Equity)
	assert.Equal(t, 5000.0, ep.Cash)
	assert.Equal(t, -0.05, ep.Drawdown)
	assert.Equal(t, 0.5, ep.Exposure)
}

func TestBacktestConfigValidation(t *testing.T) {
	// Test valid config
	config := BacktestConfig{
		Symbol:       market.Symbol("BTCUSDT"),
		Timeframe:    market.Timeframe("1h"),
		InitialCash:  10000.0,
		StrategyPath: "strategy.yaml",
		DataPath:     "data.parquet",
	}

	// Basic validation
	assert.NotEmpty(t, config.Symbol)
	assert.NotEmpty(t, config.Timeframe)
	assert.Positive(t, config.InitialCash)
	assert.NotEmpty(t, config.StrategyPath)
	assert.NotEmpty(t, config.DataPath)
}

func TestBacktestResultMetrics(t *testing.T) {
	result := &BacktestResult{
		Metrics: &analytics.Metrics{
			TotalReturn:  0.85,
			CAGR:         0.15,
			SharpeRatio:  1.67,
			SortinoRatio: 2.31,
			MaxDrawdown:  -0.21,
			WinRate:      0.47,
			ProfitFactor: 1.84,
			Expectancy:   0.43,
			TotalTrades:  428,
		},
	}

	// Test metrics access
	assert.Equal(t, 0.85, result.Metrics.TotalReturn)
	assert.Equal(t, 1.67, result.Metrics.SharpeRatio)
	assert.Equal(t, -0.21, result.Metrics.MaxDrawdown)
	assert.Equal(t, 428, result.Metrics.TotalTrades)
}

func TestSimpleTestStrategyEvaluation(t *testing.T) {
	strategy := &SimpleTestStrategy{
		Symbol:    "BTCUSDT",
		BuyPrice:  50000.0,
		SellPrice: 55000.0,
	}

	// Test buy signal
	candle := &market.Candle{
		Timestamp: time.Now(),
		Open:      49500.0,
		High:      51000.0,
		Low:       49000.0,
		Close:     50500.0,
		Volume:    100.0,
	}

	state := portfolio.PositionSide("")
	orderReq, err := strategy.Evaluate(candle, &state)
	assert.NoError(t, err)
	assert.NotNil(t, orderReq)
	assert.Equal(t, order.OrderSideBuy, orderReq.Side)

	// Test sell signal
	state = portfolio.PositionSideLong
	candle2 := &market.Candle{
		Timestamp: time.Now(),
		Open:      55500.0,
		High:      55500.0,
		Low:       54000.0,
		Close:     54500.0,
		Volume:    80.0,
	}

	orderReq, err = strategy.Evaluate(candle2, &state)
	assert.NoError(t, err)
	assert.NotNil(t, orderReq)
	assert.Equal(t, order.OrderSideSell, orderReq.Side)

	// Test no signal
	state = portfolio.PositionSide("")
	candle3 := &market.Candle{
		Timestamp: time.Now(),
		Open:      48000.0,
		High:      49000.0,
		Low:       47000.0,
		Close:     48500.0,
		Volume:    50.0,
	}

	orderReq, err = strategy.Evaluate(candle3, &state)
	assert.NoError(t, err)
	assert.Nil(t, orderReq) // Price below BuyPrice, no signal
}
