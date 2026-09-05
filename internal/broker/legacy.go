package broker

import (
	"context"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/execution"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// LegacyBroker wraps SimulatedBroker to provide backward compatibility
// with the old Broker API used by the backtest engine
type LegacyBroker struct {
	broker    *SimulatedBroker
	orderMgr  *order.OrderManager
}

// NewBroker creates a legacy broker for backward compatibility with engine
// This maintains the old API: SubmitOrder(req, timestamp) while using new SimulatedBroker internally
func NewBroker(executor *execution.SimpleExecutor) *LegacyBroker {
	orderMgr := order.NewOrderManager()
	portfolio := portfolio.NewPortfolio(0) // Portfolio managed externally by engine
	simBroker := NewSimulatedBroker(executor, portfolio)
	
	return &LegacyBroker{
		broker:   simBroker,
		orderMgr: orderMgr,
	}
}

// SubmitOrder submits an order using the old API signature
// Converts OrderRequest to Order and delegates to SimulatedBroker
func (b *LegacyBroker) SubmitOrder(req order.OrderRequest, timestamp time.Time) (*order.Order, error) {
	// Create order through order manager
	ord, err := b.orderMgr.CreateOrder(req, timestamp)
	if err != nil {
		return nil, err
	}

	// Accept order immediately (backtest behavior)
	err = b.orderMgr.AcceptOrder(ord.ID, timestamp)
	if err != nil {
		return nil, err
	}

	// Submit to new broker interface
	ctx := context.Background()
	_, err = b.broker.SubmitOrder(ctx, ord)
	if err != nil {
		return nil, err
	}

	return ord, nil
}

// ProcessPendingOrders processes orders using old API signature
// Note: symbol is derived from the first order or passed candle
func (b *LegacyBroker) ProcessPendingOrders(candle *market.Candle, timestamp time.Time) ([]string, error) {
	// Determine symbol from pending orders or use a default
	// In practice, backtest is single-symbol, so we can infer it
	symbol := "BTCUSDT" // Default, will be overridden by actual orders
	
	// Get pending orders to determine symbol
	pendingOrders := b.broker.GetPendingOrders()
	if len(pendingOrders) > 0 {
		symbol = string(pendingOrders[0].Symbol)
	}
	
	return b.broker.ProcessPendingOrders(candle, symbol, timestamp)
}

// CancelOrder cancels an order using old API
func (b *LegacyBroker) CancelOrder(orderID string, timestamp time.Time) error {
	ctx := context.Background()
	return b.broker.CancelOrder(ctx, orderID)
}

// GetOrder retrieves an order using old API
func (b *LegacyBroker) GetOrder(orderID string) (*order.Order, error) {
	ctx := context.Background()
	return b.broker.GetOrder(ctx, orderID)
}

// GetPendingOrders returns pending orders
func (b *LegacyBroker) GetPendingOrders() []*order.Order {
	return b.broker.GetPendingOrders()
}

// GetOrderManager returns the order manager
func (b *LegacyBroker) GetOrderManager() *order.OrderManager {
	return b.orderMgr
}
