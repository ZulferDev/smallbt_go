package risk

import (
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/order"
	"github.com/1jehuang/backtest/internal/portfolio"
	"github.com/stretchr/testify/assert"
)

func TestRiskManagerCreation(t *testing.T) {
	config := Config{
		MaxTradesPerDay:     5,
		MaxExposurePercent:  50,
		MaxDailyLossPercent: 2,
		MaxDrawdownPercent:  10,
	}

	m := NewManager(config)
	assert.NotNil(t, m)
}

func TestRiskManagerCanEnterTrade(t *testing.T) {
	config := Config{
		MaxTradesPerDay:     5,
		MaxExposurePercent:  50,
		MaxDailyLossPercent: 2,
		MaxDrawdownPercent:  10,
	}

	m := NewManager(config)
	p := portfolio.NewPortfolio(10000.0)

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	can, msg := m.CanEnterTrade(p, req, time.Now())
	assert.True(t, can)
	assert.Empty(t, msg)
}

func TestRiskManagerMaxTradesPerDay(t *testing.T) {
	config := Config{
		MaxTradesPerDay:     2,
		MaxExposurePercent:  50,
		MaxDailyLossPercent: 2,
		MaxDrawdownPercent:  10,
	}

	m := NewManager(config)
	p := portfolio.NewPortfolio(10000.0)

	now := time.Now()
	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	// Record trades
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	m.tradesToday[today] = 2

	can, msg := m.CanEnterTrade(p, req, now)
	assert.False(t, can)
	assert.Contains(t, msg, "max trades per day")
}

func TestRiskManagerExposureLimit(t *testing.T) {
	config := Config{
		MaxTradesPerDay:     5,
		MaxExposurePercent:  50,
		MaxDailyLossPercent: 2,
		MaxDrawdownPercent:  10,
	}

	m := NewManager(config)
	p := portfolio.NewPortfolio(10000.0)

	// Create position with high exposure
	p.OpenPosition(market.Symbol("BTCUSDT"), portfolio.PositionSideLong, 50000.0, time.Now())
	p.Positions[market.Symbol("BTCUSDT")].CurrentPrice = 10000.0 // exposure = 10000

	req := order.OrderRequest{
		Symbol:   market.Symbol("ETHUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	can, msg := m.CanEnterTrade(p, req, time.Now())
	assert.False(t, can)
	assert.Contains(t, msg, "exposure limit")
}

func TestRiskManagerUpdate(t *testing.T) {
	config := Config{
		MaxTradesPerDay:     5,
		MaxExposurePercent:  50,
		MaxDailyLossPercent: 2,
		MaxDrawdownPercent:  10,
	}

	m := NewManager(config)
	p := portfolio.NewPortfolio(10000.0)

	m.Update(p, time.Now())
	assert.Equal(t, 10000.0, m.maxEquity)
}

func TestRiskManagerRecordTrade(t *testing.T) {
	config := Config{
		MaxTradesPerDay:     5,
		MaxExposurePercent:  50,
		MaxDailyLossPercent: 2,
		MaxDrawdownPercent:  10,
	}

	m := NewManager(config)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	m.RecordTrade(true, now)
	assert.Equal(t, 1, m.tradesToday[today])

	m.RecordTrade(false, now)
	assert.Equal(t, 2, m.tradesToday[today])
	assert.Equal(t, 1, m.dailyLossCount)
}

func TestRiskManagerReset(t *testing.T) {
	config := Config{
		MaxTradesPerDay:     5,
		MaxExposurePercent:  50,
		MaxDailyLossPercent: 2,
		MaxDrawdownPercent:  10,
	}

	m := NewManager(config)
	now := time.Now()

	m.RecordTrade(true, now)
	m.maxEquity = 15000.0

	m.Reset()
	assert.Equal(t, 0, len(m.tradesToday))
	assert.Equal(t, 0.0, m.maxEquity)
}

func TestRiskManagerDrawdownLimit(t *testing.T) {
	config := Config{
		MaxTradesPerDay:     5,
		MaxExposurePercent:  50,
		MaxDailyLossPercent: 2,
		MaxDrawdownPercent:  5,
	}

	m := NewManager(config)
	p := portfolio.NewPortfolio(10000.0)

	// Set max equity high
	m.maxEquity = 15000.0

	// Portfolio equity dropped significantly
	p.Equity = 14200.0 // ~5.33% drawdown

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	can, msg := m.CanEnterTrade(p, req, time.Now())
	assert.False(t, can)
	assert.Contains(t, msg, "drawdown limit")
}
