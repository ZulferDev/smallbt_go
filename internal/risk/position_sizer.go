package risk

import (
	"fmt"

	"github.com/1jehuang/backtest/internal/portfolio"
	"github.com/1jehuang/backtest/internal/strategy/ast"
)

// PositionSizer calculates position size based on risk configuration.
type PositionSizer struct {
	config ast.PositionSizeConfig
}

// NewPositionSizer creates a new position sizer.
func NewPositionSizer(config ast.PositionSizeConfig) *PositionSizer {
	return &PositionSizer{config: config}
}

// CalculateQuantity calculates the quantity to trade based on the position sizing method.
// For risk-based sizing, stopLossPrice is required.
func (ps *PositionSizer) CalculateQuantity(
	entryPrice float64,
	accountEquity float64,
	stopLossPrice float64,
) (float64, error) {
	switch ps.config.Type {
	case "fixed":
		return ps.config.Value, nil

	case "percent_equity":
		// Allocate a percentage of equity
		allocation := accountEquity * ps.config.Value / 100
		quantity := allocation / entryPrice
		return quantity, nil

	case "risk_percent":
		// Risk a percentage of equity based on stop loss distance
		if stopLossPrice == 0 {
			return 0, fmt.Errorf("stop loss price required for risk_percent sizing")
		}

		riskAmount := accountEquity * ps.config.RiskPercent / 100
		priceDistance := absPrice(entryPrice - stopLossPrice)

		if priceDistance <= 0 {
			return 0, fmt.Errorf("stop loss price must be different from entry price")
		}

		quantity := riskAmount / priceDistance
		return quantity, nil

	default:
		return 0, fmt.Errorf("unknown position size type: %s", ps.config.Type)
	}
}

// absPrice returns the absolute difference between two prices.
func absPrice(diff float64) float64 {
	if diff < 0 {
		return -diff
	}
	return diff
}

// ValidateForPortfolio checks if the position size is valid for the current portfolio.
func (ps *PositionSizer) ValidateForPortfolio(quantity float64, entryPrice float64, portfolio *portfolio.Portfolio) error {
	notionalValue := quantity * entryPrice

	// Check if we have enough cash
	if notionalValue > portfolio.Cash {
		return fmt.Errorf("insufficient cash: need %.2f, have %.2f", notionalValue, portfolio.Cash)
	}

	return nil
}
