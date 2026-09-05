package risk

import (
	"fmt"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/order"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// Config holds risk management configuration.
type Config struct {
	MaxTradesPerDay     int
	MaxExposurePercent  float64
	MaxDailyLossPercent float64
	MaxDrawdownPercent  float64
}

// Manager manages risk for trading.
type Manager struct {
	config          Config
	tradesToday     map[time.Time]int
	dailyEquityHigh float64
	maxEquity       float64
	dailyLossCount  int
}

// NewManager creates a new risk manager.
func NewManager(config Config) *Manager {
	return &Manager{
		config:      config,
		tradesToday: make(map[time.Time]int),
		maxEquity:   0,
	}
}

// CanEnterTrade checks if entering a trade is allowed given current risk state.
func (m *Manager) CanEnterTrade(portfolio *portfolio.Portfolio, orderReq order.OrderRequest, timestamp time.Time) (bool, string) {
	// Check trades per day limit
	date := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, timestamp.Location())
	if m.tradesToday[date] >= m.config.MaxTradesPerDay {
		return false, fmt.Sprintf("max trades per day reached (%d/%d)", m.tradesToday[date], m.config.MaxTradesPerDay)
	}

	// Check exposure limit
	if m.config.MaxExposurePercent > 0 {
		exposure := portfolio.GetExposure()
		maxExposure := portfolio.Equity * m.config.MaxExposurePercent / 100
		if exposure >= maxExposure {
			return false, fmt.Sprintf("exposure limit reached: %.2f >= %.2f", exposure, maxExposure)
		}
	}

	// Check drawdown limit
	if m.config.MaxDrawdownPercent > 0 && m.maxEquity > 0 {
		currentDD := (m.maxEquity - portfolio.Equity) / m.maxEquity * 100
		if currentDD > m.config.MaxDrawdownPercent {
			return false, fmt.Sprintf("drawdown limit reached: %.2f%% > %.2f%%", currentDD, m.config.MaxDrawdownPercent)
		}
	}

	// Check daily loss limit
	if m.dailyLossCount > 2 && m.dailyEquityHigh > 0 {
		dailyLoss := (m.dailyEquityHigh - portfolio.Equity) / m.dailyEquityHigh * 100
		if dailyLoss > m.config.MaxDailyLossPercent {
			return false, fmt.Sprintf("daily loss limit reached: %.2f%% > %.2f%%", dailyLoss, m.config.MaxDailyLossPercent)
		}
	}

	return true, ""
}

// Update updates risk state with portfolio changes.
func (m *Manager) Update(portfolio *portfolio.Portfolio, timestamp time.Time) {
	// Update max equity
	if portfolio.Equity > m.maxEquity {
		m.maxEquity = portfolio.Equity
	}

	// Update daily equity high
	if portfolio.Equity > m.dailyEquityHigh {
		m.dailyEquityHigh = portfolio.Equity
	}

	// Reset daily stats if date changed
	if len(m.tradesToday) > 0 {
		var lastDate time.Time
		for d := range m.tradesToday {
			if d.After(lastDate) {
				lastDate = d
			}
		}
		if lastDate.Day() != timestamp.Day() {
			m.dailyEquityHigh = 0
			m.dailyLossCount = 0
		}
	}
}

// RecordTrade records a completed trade.
func (m *Manager) RecordTrade(profit bool, timestamp time.Time) {
	date := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, timestamp.Location())
	m.tradesToday[date]++
	if !profit {
		m.dailyLossCount++
	}
}

// Reset resets the risk manager.
func (m *Manager) Reset() {
	m.tradesToday = make(map[time.Time]int)
	m.dailyEquityHigh = 0
	m.maxEquity = 0
	m.dailyLossCount = 0
}
