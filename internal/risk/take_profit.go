package risk

import (
	"fmt"

	"github.com/1jehuang/backtest/internal/strategy/ast"
)

// TakeProfitCalculator calculates take profit price based on configuration.
type TakeProfitCalculator struct {
	config ast.TakeProfitConfig
}

// NewTakeProfitCalculator creates a new take profit calculator.
func NewTakeProfitCalculator(config *ast.TakeProfitConfig) *TakeProfitCalculator {
	if config == nil {
		return &TakeProfitCalculator{}
	}
	return &TakeProfitCalculator{config: *config}
}

// Calculate calculates the take profit price for a position.
func (tpc *TakeProfitCalculator) Calculate(
	entryPrice float64,
	side string, // "long" or "short"
	stopLossPrice float64, // For risk/reward ratio calculation
) (float64, error) {
	if tpc.config.Type == "" {
		// No take profit configured
		return 0, nil
	}

	switch tpc.config.Type {
	case "fixed":
		return tpc.config.Price, nil

	case "percentage":
		if side == "long" {
			return entryPrice * (1 + tpc.config.Percentage/100), nil
		}
		return entryPrice * (1 - tpc.config.Percentage/100), nil

	case "risk_reward":
		if stopLossPrice == 0 {
			return 0, fmt.Errorf("stop loss price required for risk/reward take profit")
		}
		risk := absPrice(entryPrice - stopLossPrice)
		reward := risk * tpc.config.Ratio
		if side == "long" {
			return entryPrice + reward, nil
		}
		return entryPrice - reward, nil

	case "expression":
		// Future: evaluate expression
		return 0, fmt.Errorf("expression-based take profit not yet implemented")

	default:
		return 0, fmt.Errorf("unknown take profit type: %s", tpc.config.Type)
	}
}

// IsActive returns true if take profit is configured.
func (tpc *TakeProfitCalculator) IsActive() bool {
	return tpc.config.Type != ""
}
