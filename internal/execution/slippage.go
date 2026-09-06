package execution

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
)

// SlippageModel defines the interface for slippage calculation models.
type SlippageModel interface {
	// CalculateSlippage computes slippage for an order.
	// Returns the slippage amount (positive for adverse, negative for favorable).
	CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error)
	
	// Name returns the model name for logging/debugging.
	Name() string
}

// FixedSlippageModel applies a fixed slippage amount.
type FixedSlippageModel struct {
	Amount float64 // Fixed slippage amount
}

func (m *FixedSlippageModel) Name() string {
	return "fixed"
}

func (m *FixedSlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
	// Apply fixed slippage in direction of order
	if req.Side == order.OrderSideBuy {
		return m.Amount, nil // Adverse: pay more
	}
	return -m.Amount, nil // Adverse: receive less
}

// PercentageSlippageModel applies a percentage-based slippage.
type PercentageSlippageModel struct {
	Percentage float64 // Slippage as percentage (e.g., 0.001 = 0.1%)
}

func (m *PercentageSlippageModel) Name() string {
	return "percentage"
}

func (m *PercentageSlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
	slippage := fillPrice * m.Percentage
	
	// Apply in direction of order
	if req.Side == order.OrderSideBuy {
		return slippage, nil // Pay more
	}
	return -slippage, nil // Receive less
}

// VolatilitySlippageModel calculates slippage based on market volatility.
// Uses the candle's price range (high - low) as a proxy for volatility.
type VolatilitySlippageModel struct {
	VolatilityFactor float64 // Multiplier for volatility (e.g., 0.1 = 10% of range)
	MinSlippage      float64 // Minimum slippage (absolute)
	MaxSlippage      float64 // Maximum slippage (absolute)
}

func (m *VolatilitySlippageModel) Name() string {
	return "volatility"
}

func (m *VolatilitySlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
	if candle == nil {
		return 0, fmt.Errorf("candle required for volatility-based slippage")
	}
	
	// Calculate price range as volatility proxy
	priceRange := candle.High - candle.Low
	if priceRange <= 0 {
		// No volatility, use minimum slippage
		return m.applyDirection(m.MinSlippage, req.Side), nil
	}
	
	// Calculate slippage based on volatility
	slippage := priceRange * m.VolatilityFactor
	
	// Clamp to [min, max]
	if slippage < m.MinSlippage {
		slippage = m.MinSlippage
	}
	if m.MaxSlippage > 0 && slippage > m.MaxSlippage {
		slippage = m.MaxSlippage
	}
	
	return m.applyDirection(slippage, req.Side), nil
}

func (m *VolatilitySlippageModel) applyDirection(slippage float64, side order.OrderSide) float64 {
	if side == order.OrderSideBuy {
		return slippage // Pay more
	}
	return -slippage // Receive less
}

// VolumeSlippageModel calculates slippage based on order size relative to volume.
// Larger orders relative to volume incur more slippage (market impact).
type VolumeSlippageModel struct {
	ImpactFactor float64 // Impact per unit of volume ratio
	MinSlippage  float64 // Minimum slippage
	MaxSlippage  float64 // Maximum slippage
}

func (m *VolumeSlippageModel) Name() string {
	return "volume"
}

func (m *VolumeSlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
	if candle == nil {
		return 0, fmt.Errorf("candle required for volume-based slippage")
	}
	
	if candle.Volume <= 0 {
		// No volume data, use minimum slippage
		return m.applyDirection(m.MinSlippage, req.Side), nil
	}
	
	// Calculate order size as fraction of volume
	// Assuming req.Quantity is in base currency units
	orderNotional := req.Quantity * fillPrice
	volumeNotional := candle.Volume * fillPrice // Approximate
	
	volumeRatio := orderNotional / volumeNotional
	
	// Calculate slippage based on volume ratio
	slippage := volumeRatio * m.ImpactFactor * fillPrice
	
	// Clamp to [min, max]
	if slippage < m.MinSlippage {
		slippage = m.MinSlippage
	}
	if m.MaxSlippage > 0 && slippage > m.MaxSlippage {
		slippage = m.MaxSlippage
	}
	
	return m.applyDirection(slippage, req.Side), nil
}

func (m *VolumeSlippageModel) applyDirection(slippage float64, side order.OrderSide) float64 {
	if side == order.OrderSideBuy {
		return slippage // Pay more
	}
	return -slippage // Receive less
}

// NoSlippageModel applies no slippage (perfect execution).
type NoSlippageModel struct{}

func (m *NoSlippageModel) Name() string {
	return "none"
}

func (m *NoSlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
	return 0, nil
}
