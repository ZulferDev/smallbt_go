package portfolio

import (
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/market"
	"github.com/stretchr/testify/assert"
)

func TestPortfolioInitialization(t *testing.T) {
	p := NewPortfolio(10000.0)
	assert.Equal(t, 10000.0, p.Cash)
	assert.Equal(t, 10000.0, p.Equity)
	assert.Equal(t, 0, len(p.Positions))
}

func TestPortfolioOpenPosition(t *testing.T) {
	p := NewPortfolio(10000.0)
	symbol := market.Symbol("BTCUSDT")

	p.OpenPosition(symbol, PositionSideLong, 0.1, 50000.0, time.Now())

	assert.Equal(t, 1, len(p.Positions))
	pos := p.Positions[symbol]
	assert.Equal(t, PositionSideLong, pos.Side)
	assert.Equal(t, 50000.0, pos.EntryPrice)
	assert.Equal(t, 0.1, pos.Quantity)
}

func TestPortfolioClosePositionLong(t *testing.T) {
	p := NewPortfolio(10000.0)
	symbol := market.Symbol("BTCUSDT")
	now := time.Now()

	// Open long position
	p.OpenPosition(symbol, PositionSideLong, 0.1, 50000.0, now)

	// Close at higher price (profit)
	trade := p.ClosePosition(symbol, 60000.0, now.Add(time.Hour))

	assert.NotNil(t, trade)
	assert.Equal(t, 0, len(p.Positions))
	assert.Greater(t, p.Cash, 10000.0)   // Should have profit
	assert.Equal(t, p.Cash, p.Equity)    // Equity should equal cash when no positions
	assert.Greater(t, p.Equity, 10000.0) // Equity should be higher than initial
	assert.Equal(t, 0.1, trade.Quantity)
}

func TestPortfolioClosePositionShort(t *testing.T) {
	p := NewPortfolio(10000.0)
	symbol := market.Symbol("BTCUSDT")
	now := time.Now()

	// Open short position
	p.OpenPosition(symbol, PositionSideShort, 0.1, 50000.0, now)

	// Close at lower price (profit)
	trade := p.ClosePosition(symbol, 40000.0, now.Add(time.Hour))

	assert.NotNil(t, trade)
	assert.Equal(t, 0, len(p.Positions))
	assert.Greater(t, p.Cash, 10000.0)
	assert.Equal(t, 0.1, trade.Quantity)
}

func TestPortfolioPnLCalculation(t *testing.T) {
	p := NewPortfolio(10000.0)
	symbol := market.Symbol("BTCUSDT")
	now := time.Now()

	p.OpenPosition(symbol, PositionSideLong, 0.1, 50000.0, now)
	trade := p.ClosePosition(symbol, 55000.0, now.Add(time.Hour))

	assert.NotNil(t, trade)
	assert.Greater(t, trade.GrossPnL, 0.0) // Profit
}

func TestPortfolioMultiplePositions(t *testing.T) {
	p := NewPortfolio(50000.0)
	now := time.Now()

	p.OpenPosition(market.Symbol("BTCUSDT"), PositionSideLong, 0.1, 50000.0, now)
	p.OpenPosition(market.Symbol("ETHUSDT"), PositionSideLong, 1.0, 3000.0, now)

	assert.Equal(t, 2, len(p.Positions))
}

func TestPortfolioEquity(t *testing.T) {
	p := NewPortfolio(10000.0)
	symbol := market.Symbol("BTCUSDT")
	now := time.Now()

	p.OpenPosition(symbol, PositionSideLong, 0.1, 50000.0, now)
	p.Positions[symbol].CurrentPrice = 55000.0 // Price moved up
	p.RecalculateEquity()

	assert.Greater(t, p.Equity, p.Cash)
}

func TestPortfolioGetExposure(t *testing.T) {
	p := NewPortfolio(10000.0)
	now := time.Now()

	p.OpenPosition(market.Symbol("BTCUSDT"), PositionSideLong, 0.1, 50000.0, now)
	p.Positions[market.Symbol("BTCUSDT")].CurrentPrice = 55000.0
	p.OpenPosition(market.Symbol("ETHUSDT"), PositionSideLong, 1.0, 3000.0, now)
	p.Positions[market.Symbol("ETHUSDT")].CurrentPrice = 3200.0

	exposure := p.GetExposure()
	assert.Equal(t, 55000.0*0.1+3200.0*1.0, exposure)
}
