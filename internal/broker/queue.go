package broker

import (
	"sync"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/order"
)

// OrderQueue manages order lifecycle for paper trading
type OrderQueue struct {
	orders map[string]*QueuedOrder
	mu     sync.RWMutex
}

// QueuedOrder represents an order in the paper trading queue
type QueuedOrder struct {
	Order      *order.Order
	SubmitTime time.Time
	AcceptTime time.Time
	Status     OrderQueueStatus
}

// OrderQueueStatus represents the status of an order in the queue
type OrderQueueStatus string

const (
	// StatusPending means order is submitted but not yet accepted
	StatusPending OrderQueueStatus = "pending"

	// StatusAccepted means order is accepted and waiting for fill
	StatusAccepted OrderQueueStatus = "accepted"

	// StatusFilled means order is completely filled
	StatusFilled OrderQueueStatus = "filled"

	// StatusCancelled means order was cancelled
	StatusCancelled OrderQueueStatus = "cancelled"
)

// NewOrderQueue creates a new order queue
func NewOrderQueue() *OrderQueue {
	return &OrderQueue{
		orders: make(map[string]*QueuedOrder),
	}
}

// Add adds an order to the queue with submit and accept times
func (q *OrderQueue) Add(order *order.Order, submitTime, acceptTime time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.orders[order.ID] = &QueuedOrder{
		Order:      order,
		SubmitTime: submitTime,
		AcceptTime: acceptTime,
		Status:     StatusPending,
	}
}

// GetPendingOrders returns orders that are pending and past their accept time
func (q *OrderQueue) GetPendingOrders(now time.Time) []*QueuedOrder {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var pending []*QueuedOrder
	for _, qo := range q.orders {
		if qo.Status == StatusPending && !now.Before(qo.AcceptTime) {
			pending = append(pending, qo)
		}
	}
	return pending
}

// GetAcceptedOrders returns all accepted orders waiting for fill
func (q *OrderQueue) GetAcceptedOrders() []*QueuedOrder {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var accepted []*QueuedOrder
	for _, qo := range q.orders {
		if qo.Status == StatusAccepted {
			accepted = append(accepted, qo)
		}
	}
	return accepted
}

// UpdateStatus updates the status of an order in the queue
func (q *OrderQueue) UpdateStatus(orderID string, status OrderQueueStatus) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if qo, exists := q.orders[orderID]; exists {
		qo.Status = status
	}
}

// Remove removes an order from the queue
func (q *OrderQueue) Remove(orderID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.orders, orderID)
}

// Get retrieves a queued order by ID
func (q *OrderQueue) Get(orderID string) (*QueuedOrder, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	qo, exists := q.orders[orderID]
	return qo, exists
}

// Count returns the total number of orders in the queue
func (q *OrderQueue) Count() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.orders)
}

// CountByStatus returns the number of orders with the given status
func (q *OrderQueue) CountByStatus(status OrderQueueStatus) int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	count := 0
	for _, qo := range q.orders {
		if qo.Status == status {
			count++
		}
	}
	return count
}

// Clear removes all orders from the queue
func (q *OrderQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.orders = make(map[string]*QueuedOrder)
}
