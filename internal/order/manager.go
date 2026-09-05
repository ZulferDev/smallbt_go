package order

import (
	"fmt"
	"sync"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// OrderEvent represents an event in an order's lifecycle.
type OrderEvent struct {
	OrderID   string
	Type      string // "created", "accepted", "filled", "rejected", "cancelled"
	Status    OrderStatus
	Timestamp time.Time
	Details   map[string]interface{}
}

// OrderManager manages order lifecycle and history.
type OrderManager struct {
	mu       sync.RWMutex
	orders   map[string]*Order
	history  []OrderEvent
	nextID   int64
	sequence map[string]int64 // Track sequence per symbol
}

// NewOrderManager creates a new order manager.
func NewOrderManager() *OrderManager {
	return &OrderManager{
		orders:   make(map[string]*Order),
		history:  make([]OrderEvent, 0),
		nextID:   1,
		sequence: make(map[string]int64),
	}
}

// CreateOrder creates a new order and records the event.
func (om *OrderManager) CreateOrder(req OrderRequest, timestamp time.Time) (*Order, error) {
	om.mu.Lock()
	defer om.mu.Unlock()

	if req.Quantity <= 0 {
		return nil, fmt.Errorf("invalid quantity: %f", req.Quantity)
	}

	// Validate order type constraints
	if req.Type == OrderTypeLimit && req.Price == nil {
		return nil, fmt.Errorf("limit order requires price")
	}
	if req.Type == OrderTypeStop && req.StopPrice == nil {
		return nil, fmt.Errorf("stop order requires stop price")
	}
	if req.Type == OrderTypeStopLimit && (req.Price == nil || req.StopPrice == nil) {
		return nil, fmt.Errorf("stop-limit order requires both price and stop price")
	}

	// Generate order ID
	orderID := fmt.Sprintf("%s_%d_%d", req.Symbol, om.sequence[string(req.Symbol)], om.nextID)
	om.nextID++
	om.sequence[string(req.Symbol)]++

	order := &Order{
		ID:          orderID,
		Symbol:      req.Symbol,
		Side:        req.Side,
		Type:        req.Type,
		Quantity:    req.Quantity,
		Price:       req.Price,
		StopPrice:   req.StopPrice,
		Status:      OrderStatusPending,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		FilledQty:   0,
		FilledPrice: 0,
		Fees:        0,
	}

	om.orders[orderID] = order

	// Record event
	om.recordEvent(OrderEvent{
		OrderID:   orderID,
		Type:      "created",
		Status:    OrderStatusPending,
		Timestamp: timestamp,
		Details: map[string]interface{}{
			"symbol":   req.Symbol,
			"side":     req.Side,
			"type":     req.Type,
			"quantity": req.Quantity,
		},
	})

	return order, nil
}

// AcceptOrder marks an order as accepted.
func (om *OrderManager) AcceptOrder(orderID string, timestamp time.Time) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	order, ok := om.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status != OrderStatusPending {
		return fmt.Errorf("cannot accept order with status: %s", order.Status)
	}

	order.Status = OrderStatusAccepted
	order.UpdatedAt = timestamp

	om.recordEvent(OrderEvent{
		OrderID:   orderID,
		Type:      "accepted",
		Status:    OrderStatusAccepted,
		Timestamp: timestamp,
		Details:   map[string]interface{}{},
	})

	return nil
}

// FillOrder marks an order as filled with the given fill details.
func (om *OrderManager) FillOrder(orderID string, fill *Fill) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	order, ok := om.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status == OrderStatusFilled {
		return fmt.Errorf("order already filled: %s", orderID)
	}

	if order.Status == OrderStatusCancelled || order.Status == OrderStatusRejected {
		return fmt.Errorf("cannot fill cancelled/rejected order: %s", orderID)
	}

	// Update fill quantities
	newFilledQty := order.FilledQty + fill.Quantity
	if newFilledQty > order.Quantity {
		return fmt.Errorf("fill quantity exceeds order quantity: %f > %f", newFilledQty, order.Quantity)
	}

	// Calculate weighted average fill price
	if order.FilledQty > 0 {
		order.FilledPrice = (order.FilledPrice*order.FilledQty + fill.Price*fill.Quantity) / newFilledQty
	} else {
		order.FilledPrice = fill.Price
	}

	order.FilledQty = newFilledQty
	order.Fees += fill.Fees
	order.UpdatedAt = fill.Timestamp

	// Determine status
	if order.FilledQty >= order.Quantity {
		order.Status = OrderStatusFilled
	} else {
		order.Status = OrderStatusPartiallyFilled
	}

	om.recordEvent(OrderEvent{
		OrderID:   orderID,
		Type:      "filled",
		Status:    order.Status,
		Timestamp: fill.Timestamp,
		Details: map[string]interface{}{
			"filled_quantity": fill.Quantity,
			"fill_price":      fill.Price,
			"total_filled":    order.FilledQty,
			"fees":            fill.Fees,
			"slippage":        fill.Slippage,
		},
	})

	return nil
}

// RejectOrder marks an order as rejected.
func (om *OrderManager) RejectOrder(orderID string, reason string, timestamp time.Time) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	order, ok := om.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status != OrderStatusPending && order.Status != OrderStatusAccepted {
		return fmt.Errorf("cannot reject order with status: %s", order.Status)
	}

	order.Status = OrderStatusRejected
	order.UpdatedAt = timestamp

	om.recordEvent(OrderEvent{
		OrderID:   orderID,
		Type:      "rejected",
		Status:    OrderStatusRejected,
		Timestamp: timestamp,
		Details: map[string]interface{}{
			"reason": reason,
		},
	})

	return nil
}

// CancelOrder marks an order as cancelled.
func (om *OrderManager) CancelOrder(orderID string, timestamp time.Time) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	order, ok := om.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status == OrderStatusFilled || order.Status == OrderStatusCancelled || order.Status == OrderStatusRejected {
		return fmt.Errorf("cannot cancel order with status: %s", order.Status)
	}

	order.Status = OrderStatusCancelled
	order.UpdatedAt = timestamp

	om.recordEvent(OrderEvent{
		OrderID:   orderID,
		Type:      "cancelled",
		Status:    OrderStatusCancelled,
		Timestamp: timestamp,
		Details: map[string]interface{}{
			"filled_qty": order.FilledQty,
		},
	})

	return nil
}

// ExpireOrder marks an order as expired.
func (om *OrderManager) ExpireOrder(orderID string, timestamp time.Time) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	order, ok := om.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status == OrderStatusFilled || order.Status == OrderStatusCancelled || order.Status == OrderStatusRejected {
		return fmt.Errorf("cannot expire order with status: %s", order.Status)
	}

	order.Status = OrderStatusExpired
	order.UpdatedAt = timestamp

	om.recordEvent(OrderEvent{
		OrderID:   orderID,
		Type:      "expired",
		Status:    OrderStatusExpired,
		Timestamp: timestamp,
		Details:   map[string]interface{}{},
	})

	return nil
}

// GetOrder returns an order by ID.
func (om *OrderManager) GetOrder(orderID string) (*Order, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	order, ok := om.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	return order, nil
}

// GetOrdersBySymbol returns all orders for a symbol.
func (om *OrderManager) GetOrdersBySymbol(symbol market.Symbol) []*Order {
	om.mu.RLock()
	defer om.mu.RUnlock()

	var result []*Order
	for _, order := range om.orders {
		if order.Symbol == symbol {
			result = append(result, order)
		}
	}
	return result
}

// GetOrdersByStatus returns all orders with a given status.
func (om *OrderManager) GetOrdersByStatus(status OrderStatus) []*Order {
	om.mu.RLock()
	defer om.mu.RUnlock()

	var result []*Order
	for _, order := range om.orders {
		if order.Status == status {
			result = append(result, order)
		}
	}
	return result
}

// GetHistory returns the event history.
func (om *OrderManager) GetHistory() []OrderEvent {
	om.mu.RLock()
	defer om.mu.RUnlock()

	// Return a copy
	result := make([]OrderEvent, len(om.history))
	copy(result, om.history)
	return result
}

// recordEvent records an order event (internal, must hold lock).
func (om *OrderManager) recordEvent(event OrderEvent) {
	om.history = append(om.history, event)
}

// GetOrderCount returns the total number of orders.
func (om *OrderManager) GetOrderCount() int {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return len(om.orders)
}

// GetEventCount returns the total number of events.
func (om *OrderManager) GetEventCount() int {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return len(om.history)
}
