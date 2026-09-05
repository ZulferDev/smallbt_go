package broker

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/execution"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// PaperBroker implements Broker interface for paper trading
// Uses real-time data with simulated execution and realistic latency
type PaperBroker struct {
	orderManager *order.OrderManager
	executor     *execution.SimpleExecutor
	portfolio    *portfolio.Portfolio

	// Paper-specific components
	orderQueue  *OrderQueue
	latencySim  *LatencySimulator
	lastPrices  map[string]float64

	// Background processing
	ticker    *time.Ticker
	stopCh    chan struct{}
	stoppedCh chan struct{}
	wg        sync.WaitGroup

	mu     sync.RWMutex
	closed bool
}

// LatencyConfig configures order latency simulation
type LatencyConfig struct {
	MinLatency time.Duration // Minimum order latency (e.g., 50ms)
	MaxLatency time.Duration // Maximum order latency (e.g., 200ms)
	Seed       int64         // Random seed for reproducibility
}

// DefaultLatencyConfig returns realistic latency settings
func DefaultLatencyConfig() LatencyConfig {
	return LatencyConfig{
		MinLatency: 50 * time.Millisecond,
		MaxLatency: 200 * time.Millisecond,
		Seed:       time.Now().UnixNano(),
	}
}

// NewPaperBroker creates a new paper trading broker
func NewPaperBroker(
	executor *execution.SimpleExecutor,
	portfolio *portfolio.Portfolio,
	latencyConfig LatencyConfig,
) *PaperBroker {
	b := &PaperBroker{
		orderManager: order.NewOrderManager(),
		executor:     executor,
		portfolio:    portfolio,
		orderQueue:   NewOrderQueue(),
		latencySim:   NewLatencySimulator(latencyConfig),
		lastPrices:   make(map[string]float64),
		stopCh:       make(chan struct{}),
		stoppedCh:    make(chan struct{}),
		closed:       false,
	}

	// Start background order processing
	b.startBackgroundProcessing()

	return b
}

// SubmitOrder implements Broker interface for paper trading
// Orders are queued and accepted after simulated latency
func (b *PaperBroker) SubmitOrder(ctx context.Context, ord *order.Order) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return "", ErrBrokerClosed
	}

	// Validate order
	if ord.Quantity <= 0 {
		return "", ErrInvalidQuantity
	}

	// Create order in order manager first
	submitTime := time.Now()
	req := order.OrderRequest{
		Symbol:    ord.Symbol,
		Side:      ord.Side,
		Type:      ord.Type,
		Quantity:  ord.Quantity,
		Price:     ord.Price,
		StopPrice: ord.StopPrice,
	}

	createdOrder, err := b.orderManager.CreateOrder(req, submitTime)
	if err != nil {
		return "", err
	}

	// Calculate when order will be accepted (submit time + latency)
	acceptTime := b.latencySim.SimulateOrderAcceptance(submitTime)

	// Add to order queue
	b.orderQueue.Add(createdOrder, submitTime, acceptTime)

	return createdOrder.ID, nil
}

// ProcessOrderQueue processes pending and accepted orders
// Should be called periodically (e.g., every 100ms) in paper trading loop
func (b *PaperBroker) ProcessOrderQueue(now time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ErrBrokerClosed
	}

	// Step 1: Accept pending orders that have passed their accept time
	pendingOrders := b.orderQueue.GetPendingOrders(now)
	for _, qo := range pendingOrders {
		// Accept order in order manager
		err := b.orderManager.AcceptOrder(qo.Order.ID, now)
		if err != nil {
			continue // Skip this order
		}

		// Update queue status
		b.orderQueue.UpdateStatus(qo.Order.ID, StatusAccepted)
	}

	// Step 2: Try to fill accepted orders
	acceptedOrders := b.orderQueue.GetAcceptedOrders()
	for _, qo := range acceptedOrders {
		// Get current price for symbol
		price, exists := b.lastPrices[string(qo.Order.Symbol)]
		if !exists {
			continue // No price data yet
		}

		// Create a mock candle for execution simulation
		candle := &market.Candle{
			Timestamp: now,
			Close:     price,
			High:      price * 1.001, // Small spread
			Low:       price * 0.999,
			Open:      price,
			Volume:    0,
		}

		// Try to fill order
		fill, err := b.executor.SimulateFill(order.OrderRequest{
			Symbol:    qo.Order.Symbol,
			Side:      qo.Order.Side,
			Type:      qo.Order.Type,
			Quantity:  qo.Order.Quantity - qo.Order.FilledQty,
			Price:     qo.Order.Price,
			StopPrice: qo.Order.StopPrice,
		}, candle)

		if err != nil {
			continue // Order not filled this iteration
		}

		// Fill the order
		fill.OrderID = qo.Order.ID
		err = b.orderManager.FillOrder(qo.Order.ID, fill)
		if err != nil {
			continue
		}

		// Update portfolio with fill
		if qo.Order.Side == order.OrderSideBuy {
			_ = b.portfolio.OpenPosition(
				qo.Order.Symbol,
				portfolio.PositionSideLong,
				fill.Quantity,
				fill.Price,
				now,
			)
		} else if qo.Order.Side == order.OrderSideSell {
			// For now, assume sell closes position
			_ = b.portfolio.ClosePosition(qo.Order.Symbol, fill.Price, now)
		}

		// Update queue status
		updatedOrder, _ := b.orderManager.GetOrder(qo.Order.ID)
		if updatedOrder.Status == order.OrderStatusFilled {
			b.orderQueue.UpdateStatus(qo.Order.ID, StatusFilled)
		}
	}

	return nil
}

// UpdatePrice updates the last known price for a symbol
// Should be called when new market data arrives
func (b *PaperBroker) UpdatePrice(symbol string, price float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastPrices[symbol] = price
}

// CancelOrder implements Broker interface
func (b *PaperBroker) CancelOrder(ctx context.Context, orderID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ErrBrokerClosed
	}

	// Cancel in order manager
	err := b.orderManager.CancelOrder(orderID, time.Now())
	if err != nil {
		return err
	}

	// Remove from queue
	b.orderQueue.Remove(orderID)

	return nil
}

// GetOrder implements Broker interface
func (b *PaperBroker) GetOrder(ctx context.Context, orderID string) (*order.Order, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil, ErrBrokerClosed
	}

	return b.orderManager.GetOrder(orderID)
}

// GetPositions implements Broker interface
func (b *PaperBroker) GetPositions(ctx context.Context) ([]*portfolio.Position, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil, ErrBrokerClosed
	}

	return b.portfolio.GetPositions(), nil
}

// GetBalance implements Broker interface
func (b *PaperBroker) GetBalance(ctx context.Context) (*portfolio.Balance, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil, ErrBrokerClosed
	}

	return &portfolio.Balance{
		Cash:   b.portfolio.Cash,
		Equity: b.portfolio.Equity,
	}, nil
}

// GetLastPrice implements Broker interface
func (b *PaperBroker) GetLastPrice(ctx context.Context, symbol string) (float64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return 0, ErrBrokerClosed
	}

	price, exists := b.lastPrices[symbol]
	if !exists {
		return 0, ErrSymbolNotFound
	}

	return price, nil
}

// startBackgroundProcessing starts the background goroutine for order processing
func (b *PaperBroker) startBackgroundProcessing() {
	b.ticker = time.NewTicker(100 * time.Millisecond)
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		defer close(b.stoppedCh)

		for {
			select {
			case <-b.ticker.C:
				b.processOrderQueueBackground()
			case <-b.stopCh:
				return
			}
		}
	}()
}

// processOrderQueueBackground processes orders in background goroutine
func (b *PaperBroker) processOrderQueueBackground() {
	// Check if closed without holding lock for long
	b.mu.RLock()
	closed := b.closed
	b.mu.RUnlock()

	if closed {
		return
	}

	// ProcessOrderQueue acquires its own lock
	_ = b.ProcessOrderQueue(time.Now())
}

// Close implements Broker interface
func (b *PaperBroker) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return ErrBrokerClosed
	}
	b.closed = true
	b.mu.Unlock()

	// Stop background processing
	if b.ticker != nil {
		b.ticker.Stop()
	}
	close(b.stopCh)

	// Wait for background goroutine to finish
	b.wg.Wait()

	return nil
}

// LatencySimulator simulates realistic order submission latency
type LatencySimulator struct {
	minLatency time.Duration
	maxLatency time.Duration
	rand       *rand.Rand
	mu         sync.Mutex
}

// NewLatencySimulator creates a new latency simulator
func NewLatencySimulator(config LatencyConfig) *LatencySimulator {
	return &LatencySimulator{
		minLatency: config.MinLatency,
		maxLatency: config.MaxLatency,
		rand:       rand.New(rand.NewSource(config.Seed)),
	}
}

// Delay returns a random latency between min and max
func (ls *LatencySimulator) Delay() time.Duration {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if ls.minLatency == ls.maxLatency {
		return ls.minLatency
	}

	rangeMs := int64(ls.maxLatency - ls.minLatency)
	randomMs := ls.rand.Int63n(rangeMs)
	return ls.minLatency + time.Duration(randomMs)
}

// SimulateOrderAcceptance returns when an order submitted at submitTime will be accepted
func (ls *LatencySimulator) SimulateOrderAcceptance(submitTime time.Time) time.Time {
	return submitTime.Add(ls.Delay())
}

// Verify interface compliance at compile time
var _ Broker = (*PaperBroker)(nil)
