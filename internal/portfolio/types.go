package portfolio

import (
	"time"

	"github.com/1jehuang/backtest/internal/market"
)

// PositionSide represents the side of a position.
type PositionSide string

const (
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
)

// Position represents an open position.
type Position struct {
	Symbol       market.Symbol
	Side         PositionSide
	Quantity     float64
	EntryPrice   float64
	EntryTime    time.Time
	StopLoss     *float64
	TakeProfit   *float64
	CurrentPrice float64
	CurrentTime  time.Time
}

// UnrealizedPnL calculates unrealized P&L.
func (p *Position) UnrealizedPnL() float64 {
	if p.Side == PositionSideLong {
		return (p.CurrentPrice - p.EntryPrice) * p.Quantity
	}
	return (p.EntryPrice - p.CurrentPrice) * p.Quantity
}

// UnrealizedPnLPercent calculates unrealized P&L percentage.
func (p *Position) UnrealizedPnLPercent() float64 {
	if p.EntryPrice == 0 {
		return 0
	}
	return (p.UnrealizedPnL() / (p.EntryPrice * p.Quantity)) * 100
}

// Trade represents a completed trade.
type Trade struct {
	ID         string
	Symbol     market.Symbol
	Side       PositionSide
	EntryTime  time.Time
	EntryPrice float64
	ExitTime   time.Time
	ExitPrice  float64
	Quantity   float64
	GrossPnL   float64
	Fees       float64
	NetPnL     float64
	Return     float64
	MAE        float64 // Maximum Adverse Excursion
	MFE        float64 // Maximum Favorable Excursion
	ExitReason string
}

// Portfolio represents the trading account portfolio.
type Portfolio struct {
	InitialCash  float64
	Cash         float64
	Equity       float64
	Balance      float64
	Positions    map[market.Symbol]*Position
	ClosedTrades []Trade
	TotalFees    float64
	Timestamp    time.Time
}

// NewPortfolio creates a new portfolio.
func NewPortfolio(initialCash float64) *Portfolio {
	return &Portfolio{
		InitialCash:  initialCash,
		Cash:         initialCash,
		Equity:       initialCash,
		Balance:      initialCash,
		Positions:    make(map[market.Symbol]*Position),
		ClosedTrades: make([]Trade, 0),
	}
}

// UpdateWithCandle updates portfolio values with current market prices.
func (p *Portfolio) UpdateWithCandle(symbol market.Symbol, close float64, timestamp time.Time) {
	if pos, exists := p.Positions[symbol]; exists {
		pos.CurrentPrice = close
		pos.CurrentTime = timestamp
	}
	p.RecalculateEquity()
	p.Timestamp = timestamp
}

// RecalculateEquity recalculates total equity from cash and positions.
func (p *Portfolio) RecalculateEquity() {
	p.Equity = p.Cash
	for _, pos := range p.Positions {
		p.Equity += pos.UnrealizedPnL()
	}
	p.Balance = p.Equity
}

// GetExposure returns total portfolio exposure across all positions.
func (p *Portfolio) GetExposure() float64 {
	var exposure float64
	for _, pos := range p.Positions {
		exposure += pos.Quantity * pos.CurrentPrice
	}
	return exposure
}

// OpenPosition opens a new position.
func (p *Portfolio) OpenPosition(symbol market.Symbol, side PositionSide, entryPrice float64, timestamp time.Time) {
	p.Positions[symbol] = &Position{
		Symbol:       symbol,
		Side:         side,
		Quantity:     1,
		EntryPrice:   entryPrice,
		EntryTime:    timestamp,
		CurrentPrice: entryPrice,
		CurrentTime:  timestamp,
	}
}

// ClosePosition closes an open position and returns a completed trade.
func (p *Portfolio) ClosePosition(symbol market.Symbol, exitPrice float64, timestamp time.Time) *Trade {
	pos, exists := p.Positions[symbol]
	if !exists {
		return nil
	}

	// Calculate PnL
	var grossPnL float64
	if pos.Side == PositionSideLong {
		grossPnL = (exitPrice - pos.EntryPrice) * pos.Quantity
	} else {
		grossPnL = (pos.EntryPrice - exitPrice) * pos.Quantity
	}

	// Estimate fees (simplified)
	fees := (pos.EntryPrice*pos.Quantity + exitPrice*pos.Quantity) * 0.0005

	netPnL := grossPnL - fees
	returnPct := 0.0
	if pos.EntryPrice > 0 {
		returnPct = (exitPrice - pos.EntryPrice) / pos.EntryPrice
		if pos.Side == PositionSideShort {
			returnPct = -returnPct
		}
	}

	trade := &Trade{
		Symbol:     symbol,
		Side:       pos.Side,
		EntryTime:  pos.EntryTime,
		EntryPrice: pos.EntryPrice,
		ExitTime:   timestamp,
		ExitPrice:  exitPrice,
		Quantity:   pos.Quantity,
		GrossPnL:   grossPnL,
		Fees:       fees,
		NetPnL:     netPnL,
		Return:     returnPct,
		ExitReason: "signal",
	}

	// Update cash
	p.Cash += netPnL
	p.TotalFees += fees

	// Remove position
	delete(p.Positions, symbol)

	// Recalculate equity
	p.RecalculateEquity()

	return trade
}
