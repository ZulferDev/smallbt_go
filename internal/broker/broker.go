package broker

import (
	"fmt"
	"time"

	"github.com/1jehuang/backtest/internal/execution"
	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/order"
)

// Broker manages order submission, execution, and lifecycle.
type Broker struct {
	orderManager  *order.OrderManager
	executor      *execution.SimpleExecutor
	pendingOrders map[string]*order.Order // Orders waiting for execution
}

// NewBroker creates a new broker.
func NewBroker(executor *execution.SimpleExecutor) *Broker {
	return &Broker{
		orderManager:  order.NewOrderManager(),
		executor:      executor,
		pendingOrders: make(map[string]*order.Order),
	}
}

// SubmitOrder submits a new order.
func (b *Broker) SubmitOrder(req order.OrderRequest, timestamp time.Time) (*order.Order, error) {
	ord, err := b.orderManager.CreateOrder(req, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Accept order immediately in backtesting
	err = b.orderManager.AcceptOrder(ord.ID, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to accept order: %w", err)
	}

	// Track pending order
	b.pendingOrders[ord.ID] = ord

	return ord, nil
}

// ProcessPendingOrders processes pending orders against current candle.
func (b *Broker) ProcessPendingOrders(candle *market.Candle, timestamp time.Time) ([]string, error) {
	var filledOrderIDs []string

	for orderID, ord := range b.pendingOrders {
		// Try to fill the order
		fill, err := b.executor.SimulateFill(order.OrderRequest{
			Symbol:    ord.Symbol,
			Side:      ord.Side,
			Type:      ord.Type,
			Quantity:  ord.Quantity - ord.FilledQty,
			Price:     ord.Price,
			StopPrice: ord.StopPrice,
		}, candle)

		if err != nil {
			// Order didn't fill, check if it should expire or remain pending
			continue
		}

		// Fill the order
		fill.OrderID = orderID
		err = b.orderManager.FillOrder(orderID, fill)
		if err != nil {
			return nil, fmt.Errorf("failed to fill order %s: %w", orderID, err)
		}

		// Check if order is fully filled
		updatedOrder, _ := b.orderManager.GetOrder(orderID)
		if updatedOrder.Status == order.OrderStatusFilled {
			filledOrderIDs = append(filledOrderIDs, orderID)
			delete(b.pendingOrders, orderID)
		}
	}

	return filledOrderIDs, nil
}

// CancelOrder cancels a pending order.
func (b *Broker) CancelOrder(orderID string, timestamp time.Time) error {
	err := b.orderManager.CancelOrder(orderID, timestamp)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	delete(b.pendingOrders, orderID)
	return nil
}

// GetOrder retrieves an order by ID.
func (b *Broker) GetOrder(orderID string) (*order.Order, error) {
	return b.orderManager.GetOrder(orderID)
}

// GetPendingOrders returns all pending orders.
func (b *Broker) GetPendingOrders() []*order.Order {
	var pending []*order.Order
	for _, ord := range b.pendingOrders {
		pending = append(pending, ord)
	}
	return pending
}

// GetOrdersBySymbol returns all orders for a symbol.
func (b *Broker) GetOrdersBySymbol(symbol market.Symbol) []*order.Order {
	return b.orderManager.GetOrdersBySymbol(symbol)
}

// GetOrdersByStatus returns all orders with a given status.
func (b *Broker) GetOrdersByStatus(status order.OrderStatus) []*order.Order {
	return b.orderManager.GetOrdersByStatus(status)
}

// GetOrderHistory returns the event history.
func (b *Broker) GetOrderHistory() []order.OrderEvent {
	return b.orderManager.GetHistory()
}

// GetOrderManager returns the underlying order manager.
func (b *Broker) GetOrderManager() *order.OrderManager {
	return b.orderManager
}
