package execution

import (
	"fmt"
	"math/rand"

	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
)

// Config holds execution configuration.
type Config struct {
	SlippageType   string // "percentage", "fixed", "none" (deprecated, use SlippageModel)
	SlippageValue  float64
	SlippageModel  SlippageModel // New: use interface-based slippage model
	FeeMaker       float64
	FeeTaker       float64
	Spread         float64 // Average spread
	IntrabarPolicy string  // "optimistic", "conservative", "nearest"
	Seed           int64   // Random seed for reproducibility
}

// SimpleExecutor simulates order execution in backtesting.
type SimpleExecutor struct {
	config        Config
	rng           *rand.Rand
	lastCandle    *market.Candle
	slippageModel SlippageModel
}

// NewSimpleExecutor creates a new simple execution simulator.
func NewSimpleExecutor(config Config) *SimpleExecutor {
	executor := &SimpleExecutor{
		config: config,
		rng:    rand.New(rand.NewSource(config.Seed)),
	}

	// Use provided SlippageModel if set, otherwise create from legacy config
	if config.SlippageModel != nil {
		executor.slippageModel = config.SlippageModel
	} else {
		// Fallback to legacy slippage config
		switch config.SlippageType {
		case "fixed":
			executor.slippageModel = NewFixedSlippageModel(config.SlippageValue)
		case "percentage":
			executor.slippageModel = NewPercentageSlippageModel(config.SlippageValue)
		case "none":
			executor.slippageModel = NewNoSlippageModel()
		default:
			executor.slippageModel = NewNoSlippageModel()
		}
	}

	return executor
}

// SetCurrentCandle sets the current candle for execution.
func (e *SimpleExecutor) SetCurrentCandle(candle *market.Candle) {
	e.lastCandle = candle
}

// SimulateFill simulates order fill based on order type and candle data.
func (e *SimpleExecutor) SimulateFill(req order.OrderRequest, candle *market.Candle) (*order.Fill, error) {
	if candle == nil {
		return nil, fmt.Errorf("no candle data available")
	}

	// Calculate fill price based on order type
	var fillPrice float64
	switch req.Type {
	case order.OrderTypeMarket:
		fillPrice = e.calculateMarketFillPrice(req, candle)
	case order.OrderTypeLimit:
		fillPrice = e.calculateLimitFillPrice(req, candle)
	case order.OrderTypeStop:
		fillPrice = e.calculateStopFillPrice(req, candle)
	case order.OrderTypeStopLimit:
		fillPrice = e.calculateStopLimitFillPrice(req, candle)
	default:
		return nil, fmt.Errorf("unsupported order type: %s", req.Type)
	}

	if fillPrice <= 0 {
		return nil, fmt.Errorf("order not filled")
	}

	// Calculate slippage
	slippage := e.calculateSlippage(req, fillPrice)
	fillPrice += slippage

	// Calculate fees
	fees := e.calculateFees(req, fillPrice)

	return &order.Fill{
		OrderID:   "", // Will be set by caller
		Symbol:    req.Symbol,
		Side:      req.Side,
		Quantity:  req.Quantity,
		Price:     fillPrice,
		Timestamp: candle.Timestamp,
		Fees:      fees,
		Slippage:  slippage,
	}, nil
}

// calculateMarketFillPrice calculates market order fill price.
func (e *SimpleExecutor) calculateMarketFillPrice(req order.OrderRequest, candle *market.Candle) float64 {
	// Conservative: use worst price for the side
	if req.Side == order.OrderSideBuy {
		// Buy at high price (worst for buyer)
		return candle.High
	} else {
		// Sell at low price (worst for seller)
		return candle.Low
	}
}

// calculateLimitFillPrice calculates limit order fill price.
func (e *SimpleExecutor) calculateLimitFillPrice(req order.OrderRequest, candle *market.Candle) float64 {
	if req.Price == nil {
		return 0
	}
	limitPrice := *req.Price

	// Check if limit price is within candle range
	if req.Side == order.OrderSideBuy {
		// Buy limit: we want to buy at or below limit price
		if limitPrice >= candle.Low {
			// Fill at limit price if it's within range
			// Conservative: assume worst fill (highest price within limit)
			if limitPrice >= candle.High {
				return candle.High
			}
			return limitPrice
		}
	} else {
		// Sell limit: we want to sell at or above limit price
		if limitPrice <= candle.High {
			// Fill at limit price if it's within range
			// Conservative: assume worst fill (lowest price within limit)
			if limitPrice <= candle.Low {
				return candle.Low
			}
			return limitPrice
		}
	}

	return 0
}

// calculateStopFillPrice calculates stop order fill price.
func (e *SimpleExecutor) calculateStopFillPrice(req order.OrderRequest, candle *market.Candle) float64 {
	if req.StopPrice == nil {
		return 0
	}
	stopPrice := *req.StopPrice

	// Check if stop price is within candle range
	if req.Side == order.OrderSideBuy {
		// Buy stop: triggered when price rises above stop price
		if stopPrice <= candle.High {
			// Fill at stop price if triggered
			// Conservative: assume worst fill (highest price above stop)
			if stopPrice <= candle.Low {
				return candle.Low
			}
			return stopPrice
		}
	} else {
		// Sell stop: triggered when price falls below stop price
		if stopPrice >= candle.Low {
			// Fill at stop price if triggered
			// Conservative: assume worst fill (lowest price below stop)
			if stopPrice >= candle.High {
				return candle.High
			}
			return stopPrice
		}
	}

	return 0
}

// calculateStopLimitFillPrice calculates stop-limit order fill price.
func (e *SimpleExecutor) calculateStopLimitFillPrice(req order.OrderRequest, candle *market.Candle) float64 {
	if req.StopPrice == nil || req.Price == nil {
		return 0
	}
	stopPrice := *req.StopPrice

	// First check if stop is triggered
	stopTriggered := false
	if req.Side == order.OrderSideBuy && stopPrice <= candle.High {
		stopTriggered = true
	} else if req.Side == order.OrderSideSell && stopPrice >= candle.Low {
		stopTriggered = true
	}

	if !stopTriggered {
		return 0
	}

	// Once stop is triggered, it becomes a limit order
	reqCopy := req
	reqCopy.Type = order.OrderTypeLimit
	return e.calculateLimitFillPrice(reqCopy, candle)
}

// calculateSlippage calculates slippage for a fill using the configured model.
func (e *SimpleExecutor) calculateSlippage(req order.OrderRequest, fillPrice float64) float64 {
	if e.slippageModel == nil {
		return 0
	}

	// Use the slippage model with current candle context
	slippage, err := e.slippageModel.CalculateSlippage(req, fillPrice, e.lastCandle)
	if err != nil {
		// If error, fall back to no slippage
		return 0
	}

	return slippage
}

// calculateFees calculates fees for a fill.
func (e *SimpleExecutor) calculateFees(req order.OrderRequest, fillPrice float64) float64 {
	// Market orders are taker, limit orders are maker
	var feeRate float64
	switch req.Type {
	case order.OrderTypeMarket:
		feeRate = e.config.FeeTaker
	case order.OrderTypeLimit:
		feeRate = e.config.FeeMaker
	default:
		feeRate = e.config.FeeTaker
	}

	notional := fillPrice * req.Quantity
	return notional * feeRate
}

// ApplySpread applies spread to price.
func (e *SimpleExecutor) ApplySpread(price float64, side order.OrderSide) float64 {
	if side == order.OrderSideBuy {
		// Buyers pay spread (higher price)
		return price * (1 + e.config.Spread/2)
	} else {
		// Sellers receive spread (lower price)
		return price * (1 - e.config.Spread/2)
	}
}
