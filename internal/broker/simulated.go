package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/execution"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// SimulatedBroker implements Broker interface for backtesting
// It simulates order execution against historical market data
type SimulatedBroker struct {
	orderManager  *order.OrderManager
	executor      *execution.SimpleExecutor
	pendingOrders map[string]*order.Order // Orders waiting for execution
	portfolio     *portfolio.Portfolio
	lastPrices    map[string]float64 // symbol -> last price
	closed        bool
}

// NewSimulatedBroker creates a new simulated broker for backtesting
func NewSimulatedBroker(executor *execution.SimpleExecutor, portfolio *portfolio.Portfolio) *SimulatedBroker {
	return &SimulatedBroker{
		orderManager:  order.NewOrderManager(),
		executor:      executor,
		pendingOrders: make(map[string]*order.Order),
		portfolio:     portfolio,
		lastPrices:    make(map[string]float64),
		closed:        false,
	}
}

// SubmitOrder implements Broker interface
func (b *SimulatedBroker) SubmitOrder(ctx context.Context, ord *order.Order) (string, error) {
	if b.closed {
		return "", ErrBrokerClosed
	}

	// Validate order
	if ord.Quantity <= 0 {
		return "", ErrInvalidQuantity
	}

	// In backtest, orders are accepted immediately
	// Real execution happens in ProcessPendingOrders
	b.pendingOrders[ord.ID] = ord

	return ord.ID, nil
}

// CancelOrder implements Broker interface
func (b *SimulatedBroker) CancelOrder(ctx context.Context, orderID string) error {
	if b.closed {
		return ErrBrokerClosed
	}

	ord, exists := b.pendingOrders[orderID]
	if !exists {
		return ErrOrderNotFound
	}

	if ord.Status == order.OrderStatusFilled {
		return ErrOrderAlreadyFilled
	}

	if ord.Status == order.OrderStatusCancelled {
		return ErrOrderAlreadyCancelled
	}

	// Cancel the order
	err := b.orderManager.CancelOrder(orderID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	delete(b.pendingOrders, orderID)
	return nil
}

// GetOrder implements Broker interface
func (b *SimulatedBroker) GetOrder(ctx context.Context, orderID string) (*order.Order, error) {
	if b.closed {
		return nil, ErrBrokerClosed
	}

	ord, err := b.orderManager.GetOrder(orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	return ord, nil
}

// GetPositions implements Broker interface
func (b *SimulatedBroker) GetPositions(ctx context.Context) ([]*portfolio.Position, error) {
	if b.closed {
		return nil, ErrBrokerClosed
	}

	return b.portfolio.GetPositions(), nil
}

// GetBalance implements Broker interface
func (b *SimulatedBroker) GetBalance(ctx context.Context) (*portfolio.Balance, error) {
	if b.closed {
		return nil, ErrBrokerClosed
	}

	balance := &portfolio.Balance{
		Cash:   b.portfolio.Cash,
		Equity: b.portfolio.Equity,
	}

	return balance, nil
}

// GetLastPrice implements Broker interface
func (b *SimulatedBroker) GetLastPrice(ctx context.Context, symbol string) (float64, error) {
	if b.closed {
		return 0, ErrBrokerClosed
	}

	price, exists := b.lastPrices[symbol]
	if !exists {
		return 0, ErrSymbolNotFound
	}

	return price, nil
}

// Close implements Broker interface
func (b *SimulatedBroker) Close() error {
	b.closed = true
	return nil
}

// ProcessPendingOrders processes pending orders against current candle
// This is called by the backtest engine on each new candle
func (b *SimulatedBroker) ProcessPendingOrders(candle *market.Candle, symbol string, timestamp time.Time) ([]string, error) {
	if b.closed {
		return nil, ErrBrokerClosed
	}

	// Update last price
	b.lastPrices[symbol] = candle.Close

	var filledOrderIDs []string

	for orderID, ord := range b.pendingOrders {
		// Only process orders for this symbol
		if string(ord.Symbol) != symbol {
			continue
		}

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

// GetPendingOrders returns all pending orders (for backtest engine)
func (b *SimulatedBroker) GetPendingOrders() []*order.Order {
	var pending []*order.Order
	for _, ord := range b.pendingOrders {
		pending = append(pending, ord)
	}
	return pending
}

// GetOrderManager returns the underlying order manager (for backtest engine)
func (b *SimulatedBroker) GetOrderManager() *order.OrderManager {
	return b.orderManager
}

// Verify interface compliance at compile time
var _ Broker = (*SimulatedBroker)(nil)
