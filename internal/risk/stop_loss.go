package risk

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/strategy/ast"
)

// StopLossCalculator calculates stop loss price based on configuration.
type StopLossCalculator struct {
	config ast.StopLossConfig
}

// NewStopLossCalculator creates a new stop loss calculator.
func NewStopLossCalculator(config *ast.StopLossConfig) *StopLossCalculator {
	if config == nil {
		return &StopLossCalculator{}
	}
	return &StopLossCalculator{config: *config}
}

// Calculate calculates the stop loss price for a position.
// For ATR-based, atrValue must be provided.
func (slc *StopLossCalculator) Calculate(
	entryPrice float64,
	side string, // "long" or "short"
	atrValue float64,
) (float64, error) {
	if slc.config.Type == "" {
		// No stop loss configured
		return 0, nil
	}

	switch slc.config.Type {
	case "fixed":
		return slc.config.Price, nil

	case "percentage":
		if side == "long" {
			return entryPrice * (1 - slc.config.Percentage/100), nil
		}
		return entryPrice * (1 + slc.config.Percentage/100), nil

	case "atr":
		if atrValue == 0 {
			return 0, fmt.Errorf("ATR value required for ATR-based stop loss")
		}
		distance := atrValue * slc.config.Multiplier
		if side == "long" {
			return entryPrice - distance, nil
		}
		return entryPrice + distance, nil

	case "expression":
		// Future: evaluate expression
		return 0, fmt.Errorf("expression-based stop loss not yet implemented")

	default:
		return 0, fmt.Errorf("unknown stop loss type: %s", slc.config.Type)
	}
}

// IsActive returns true if stop loss is configured.
func (slc *StopLossCalculator) IsActive() bool {
	return slc.config.Type != ""
}

// TrailingStopCalculator calculates trailing stop price.
type TrailingStopCalculator struct {
	config ast.TrailingStopConfig
}

// NewTrailingStopCalculator creates a new trailing stop calculator.
func NewTrailingStopCalculator(config *ast.TrailingStopConfig) *TrailingStopCalculator {
	if config == nil {
		return &TrailingStopCalculator{}
	}
	return &TrailingStopCalculator{config: *config}
}

// UpdateTrailingStop updates the trailing stop price based on price movement.
func (tsc *TrailingStopCalculator) UpdateTrailingStop(
	currentStop float64,
	highSinceEntry float64,
	lowSinceEntry float64,
	side string,
	atrValue float64,
) (float64, error) {
	if tsc.config.Type == "" {
		return currentStop, nil
	}

	switch tsc.config.Type {
	case "percentage":
		if side == "long" {
			newStop := highSinceEntry * (1 - tsc.config.Percentage/100)
			if newStop > currentStop {
				return newStop, nil
			}
		} else {
			newStop := lowSinceEntry * (1 + tsc.config.Percentage/100)
			if newStop < currentStop || currentStop == 0 {
				return newStop, nil
			}
		}
		return currentStop, nil

	case "atr":
		if atrValue == 0 {
			return currentStop, fmt.Errorf("ATR value required for ATR-based trailing stop")
		}
		distance := atrValue * tsc.config.Multiplier
		if side == "long" {
			newStop := highSinceEntry - distance
			if newStop > currentStop {
				return newStop, nil
			}
		} else {
			newStop := lowSinceEntry + distance
			if newStop < currentStop || currentStop == 0 {
				return newStop, nil
			}
		}
		return currentStop, nil

	default:
		return currentStop, fmt.Errorf("unknown trailing stop type: %s", tsc.config.Type)
	}
}

// IsActive returns true if trailing stop is configured.
func (tsc *TrailingStopCalculator) IsActive() bool {
	return tsc.config.Type != ""
}
