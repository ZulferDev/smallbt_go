package risk

import (
	"testing"

	"github.com/1jehuang/backtest/internal/portfolio"
	"github.com/1jehuang/backtest/internal/strategy/ast"
	"github.com/stretchr/testify/assert"
)

func TestPositionSizer_Fixed(t *testing.T) {
	config := ast.PositionSizeConfig{
		Type:  "fixed",
		Value: 10.0,
	}
	ps := NewPositionSizer(config)

	quantity, err := ps.CalculateQuantity(50000, 100000, 0)

	assert.NoError(t, err)
	assert.Equal(t, 10.0, quantity)
}

func TestPositionSizer_PercentEquity(t *testing.T) {
	config := ast.PositionSizeConfig{
		Type:  "percent_equity",
		Value: 5.0, // 5% of equity
	}
	ps := NewPositionSizer(config)

	// 5% of 100000 = 5000
	// At 50000 entry price: 5000 / 50000 = 0.1
	quantity, err := ps.CalculateQuantity(50000, 100000, 0)

	assert.NoError(t, err)
	assert.InEpsilon(t, 0.1, quantity, 0.0001)
}

func TestPositionSizer_RiskPercent_Long(t *testing.T) {
	config := ast.PositionSizeConfig{
		Type:        "risk_percent",
		RiskPercent: 1.0, // 1% of equity
	}
	ps := NewPositionSizer(config)

	// Entry: 50000, Stop: 48000, Risk: 2000
	// 1% of 100000 = 1000 risk
	// Quantity = 1000 / 2000 = 0.5
	quantity, err := ps.CalculateQuantity(50000, 100000, 48000)

	assert.NoError(t, err)
	assert.Equal(t, 0.5, quantity)
}

func TestPositionSizer_RiskPercent_Short(t *testing.T) {
	config := ast.PositionSizeConfig{
		Type:        "risk_percent",
		RiskPercent: 1.0, // 1% of equity
	}
	ps := NewPositionSizer(config)

	// Entry: 50000, Stop: 52000, Risk: 2000
	// 1% of 100000 = 1000 risk
	// Quantity = 1000 / 2000 = 0.5
	quantity, err := ps.CalculateQuantity(50000, 100000, 52000)

	assert.NoError(t, err)
	assert.Equal(t, 0.5, quantity)
}

func TestPositionSizer_InvalidStopLoss(t *testing.T) {
	config := ast.PositionSizeConfig{
		Type:        "risk_percent",
		RiskPercent: 1.0,
	}
	ps := NewPositionSizer(config)

	// Entry equals stop loss - invalid
	_, err := ps.CalculateQuantity(50000, 100000, 50000)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be different from entry price")
}

func TestPositionSizer_InsufficientCash(t *testing.T) {
	config := ast.PositionSizeConfig{
		Type:  "fixed",
		Value: 10.0,
	}
	ps := NewPositionSizer(config)

	portfolio := portfolio.NewPortfolio(100000) // 100k cash
	// 10 * 50000 = 500000 > 100000
	err := ps.ValidateForPortfolio(10, 50000, portfolio)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient cash")
}

func TestStopLossCalculator_Fixed(t *testing.T) {
	config := ast.StopLossConfig{
		Type:  "fixed",
		Price: 48000,
	}
	slc := NewStopLossCalculator(&config)

	price, err := slc.Calculate(50000, "long", 0)

	assert.NoError(t, err)
	assert.Equal(t, 48000.0, price)
}

func TestStopLossCalculator_Percentage_Long(t *testing.T) {
	config := ast.StopLossConfig{
		Type:       "percentage",
		Percentage: 4.0, // 4%
	}
	slc := NewStopLossCalculator(&config)

	// 50000 * (1 - 0.04) = 48000
	price, err := slc.Calculate(50000, "long", 0)

	assert.NoError(t, err)
	assert.Equal(t, 48000.0, price)
}

func TestStopLossCalculator_Percentage_Short(t *testing.T) {
	config := ast.StopLossConfig{
		Type:       "percentage",
		Percentage: 4.0, // 4%
	}
	slc := NewStopLossCalculator(&config)

	// 50000 * (1 + 0.04) = 52000
	price, err := slc.Calculate(50000, "short", 0)

	assert.NoError(t, err)
	assert.Equal(t, 52000.0, price)
}

func TestStopLossCalculator_ATR_Long(t *testing.T) {
	config := ast.StopLossConfig{
		Type:       "atr",
		Multiplier: 1.5,
	}
	slc := NewStopLossCalculator(&config)

	// 50000 - (2000 * 1.5) = 47000
	price, err := slc.Calculate(50000, "long", 2000)

	assert.NoError(t, err)
	assert.Equal(t, 47000.0, price)
}

func TestStopLossCalculator_ATR_Short(t *testing.T) {
	config := ast.StopLossConfig{
		Type:       "atr",
		Multiplier: 1.5,
	}
	slc := NewStopLossCalculator(&config)

	// 50000 + (2000 * 1.5) = 53000
	price, err := slc.Calculate(50000, "short", 2000)

	assert.NoError(t, err)
	assert.Equal(t, 53000.0, price)
}

func TestStopLossCalculator_NoConfig(t *testing.T) {
	slc := NewStopLossCalculator(nil)

	price, err := slc.Calculate(50000, "long", 0)

	assert.NoError(t, err)
	assert.Equal(t, 0.0, price)
	assert.False(t, slc.IsActive())
}

func TestTrailingStopCalculator_Percentage_Long(t *testing.T) {
	config := ast.TrailingStopConfig{
		Type:       "percentage",
		Percentage: 2.0,
	}
	tsc := NewTrailingStopCalculator(&config)

	// First update - set initial stop
	stop, err := tsc.UpdateTrailingStop(0, 52000, 48000, "long", 0)
	assert.NoError(t, err)
	assert.Equal(t, 52000*0.98, stop) // 50960

	// Price moves up to 54000
	stop, err = tsc.UpdateTrailingStop(stop, 54000, 48000, "long", 0)
	assert.NoError(t, err)
	assert.Equal(t, 54000*0.98, stop) // 52920
}

func TestTrailingStopCalculator_ATR_Short(t *testing.T) {
	config := ast.TrailingStopConfig{
		Type:       "atr",
		Multiplier: 1.5,
	}
	tsc := NewTrailingStopCalculator(&config)

	// Short position: low since entry = 48000, atr = 2000
	// New stop = 48000 + (2000 * 1.5) = 51000
	stop, err := tsc.UpdateTrailingStop(0, 52000, 48000, "short", 2000)
	assert.NoError(t, err)
	assert.Equal(t, 51000.0, stop)
}
