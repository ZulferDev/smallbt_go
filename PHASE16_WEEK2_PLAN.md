# Phase 16 Week 2 - Paper Trading Implementation Plan

**Week:** 2 of 4  
**Start Date:** 2026-09-05  
**Duration:** 3-4 weeks (per POST_MVP_PLAN.md)  
**Focus:** Paper Trading with Real-time Simulation

---

## Week 2 Objectives

From POST_MVP_PLAN.md:

- [ ] Implement `PaperBroker`
- [ ] Add latency simulation (50-200ms)
- [ ] Implement order queue
- [ ] Add position reconciliation
- [ ] CLI support for paper trading mode
- [ ] Integration tests for paper mode
- [ ] Documentation update

**Goal:** Enable paper trading with realistic order lifecycle and latency.

---

## Current Foundation (Week 1 Complete)

✅ **Runtime Layer**
- ExecutionMode enum (backtest/paper/live)
- Config with validation
- TimeProvider interface (RealTime ready)

✅ **Broker Interface**
```go
type Broker interface {
    SubmitOrder(ctx context.Context, o *order.Order) (string, error)
    CancelOrder(ctx context.Context, orderID string) error
    GetOrder(ctx context.Context, orderID string) (*order.Order, error)
    GetPositions(ctx context.Context) ([]*portfolio.Position, error)
    GetBalance(ctx context.Context) (*portfolio.Balance, error)
    GetLastPrice(ctx context.Context, symbol string) (float64, error)
    Close() error
}
```

✅ **DataFeed Interface**
```go
type DataFeed interface {
    Subscribe(ctx context.Context, symbols []string) error
    Next(ctx context.Context) (*market.Candle, error)
    Close() error
}
```

---

## PaperBroker Architecture

### Design Principles

1. **Uses real-time data** (not historical)
2. **Simulates execution** (no real exchange)
3. **Realistic latency** (50-200ms delays)
4. **Order lifecycle** (pending → accepted → filled)
5. **Position reconciliation** (tracks state correctly)

### PaperBroker vs SimulatedBroker

| Feature | SimulatedBroker | PaperBroker |
|---------|-----------------|-------------|
| Data | Historical | Real-time |
| Time | HistoricalTime | RealTime |
| Latency | Instant | 50-200ms |
| Order queue | Simple | Realistic |
| Fill logic | Immediate | Delayed |
| Use case | Backtest | Paper trading |

---

## Implementation Plan

### Week 2 Day 1: PaperBroker Core

**File:** `internal/broker/paper.go`

```go
type PaperBroker struct {
    orderManager  *order.OrderManager
    executor      *execution.SimpleExecutor
    portfolio     *portfolio.Portfolio
    
    // Paper-specific
    orderQueue    *OrderQueue
    latencySim    *LatencySimulator
    lastPrices    map[string]float64
    priceFeed     chan PriceUpdate
    
    mu            sync.RWMutex
    closed        bool
}

type LatencySimulator struct {
    minLatency time.Duration  // 50ms
    maxLatency time.Duration  // 200ms
    rand       *rand.Rand
}

type OrderQueue struct {
    pending   map[string]*QueuedOrder
    mu        sync.Mutex
}

type QueuedOrder struct {
    Order        *order.Order
    SubmitTime   time.Time
    AcceptTime   time.Time  // SubmitTime + latency
    Status       string     // "pending", "accepted", "filled"
}
```

**Key Methods:**
```go
func NewPaperBroker(
    executor *execution.SimpleExecutor,
    portfolio *portfolio.Portfolio,
    latencyConfig LatencyConfig,
) *PaperBroker

func (b *PaperBroker) SubmitOrder(ctx context.Context, o *order.Order) (string, error) {
    // 1. Add to order queue with submit time
    // 2. Schedule acceptance after latency
    // 3. Return order ID immediately
}

func (b *PaperBroker) ProcessOrderQueue(now time.Time) {
    // 1. Check pending orders
    // 2. Accept orders past accept time
    // 3. Try to fill accepted orders against current price
}

func (b *PaperBroker) UpdatePrice(symbol string, price float64) {
    // 1. Update last price
    // 2. Trigger fill attempts for accepted orders
}
```

---

### Week 2 Day 2: Latency Simulation

**File:** `internal/broker/latency.go`

```go
type LatencySimulator struct {
    minLatency time.Duration
    maxLatency time.Duration
    rand       *rand.Rand
    mu         sync.Mutex
}

func NewLatencySimulator(min, max time.Duration, seed int64) *LatencySimulator {
    return &LatencySimulator{
        minLatency: min,
        maxLatency: max,
        rand:       rand.New(rand.NewSource(seed)),
    }
}

func (ls *LatencySimulator) Delay() time.Duration {
    ls.mu.Lock()
    defer ls.mu.Unlock()
    
    // Random latency between min and max
    range_ := ls.maxLatency - ls.minLatency
    return ls.minLatency + time.Duration(ls.rand.Int63n(int64(range_)))
}

func (ls *LatencySimulator) SimulateOrderAcceptance(
    submitTime time.Time,
) time.Time {
    return submitTime.Add(ls.Delay())
}
```

**Configuration:**
```yaml
paper_trading:
  latency:
    min: 50ms
    max: 200ms
    seed: 42  # For reproducibility in tests
```

---

### Week 2 Day 3: Order Queue

**File:** `internal/broker/queue.go`

```go
type OrderQueue struct {
    orders map[string]*QueuedOrder
    mu     sync.RWMutex
}

type QueuedOrder struct {
    Order      *order.Order
    SubmitTime time.Time
    AcceptTime time.Time
    Status     OrderQueueStatus
}

type OrderQueueStatus string

const (
    StatusPending  OrderQueueStatus = "pending"   // Submitted, not accepted yet
    StatusAccepted OrderQueueStatus = "accepted"  // Accepted, waiting for fill
    StatusFilled   OrderQueueStatus = "filled"    // Filled
    StatusCancelled OrderQueueStatus = "cancelled"
)

func NewOrderQueue() *OrderQueue {
    return &OrderQueue{
        orders: make(map[string]*QueuedOrder),
    }
}

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

func (q *OrderQueue) GetPendingOrders(now time.Time) []*QueuedOrder {
    q.mu.RLock()
    defer q.mu.RUnlock()
    
    var pending []*QueuedOrder
    for _, qo := range q.orders {
        if qo.Status == StatusPending && now.After(qo.AcceptTime) {
            pending = append(pending, qo)
        }
    }
    return pending
}

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

func (q *OrderQueue) UpdateStatus(orderID string, status OrderQueueStatus) {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    if qo, exists := q.orders[orderID]; exists {
        qo.Status = status
    }
}

func (q *OrderQueue) Remove(orderID string) {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    delete(q.orders, orderID)
}
```

---

### Week 2 Day 4: Position Reconciliation

**File:** `internal/broker/reconciliation.go`

```go
type PositionReconciler struct {
    portfolio *portfolio.Portfolio
}

func NewPositionReconciler(portfolio *portfolio.Portfolio) *PositionReconciler {
    return &PositionReconciler{
        portfolio: portfolio,
    }
}

func (r *PositionReconciler) Reconcile(
    expectedPositions []*portfolio.Position,
) error {
    // Compare expected vs actual positions
    actualPositions := r.portfolio.GetPositions()
    
    // Check for discrepancies
    for _, expected := range expectedPositions {
        actual := findPosition(actualPositions, expected.Symbol)
        if actual == nil {
            return fmt.Errorf("missing position: %s", expected.Symbol)
        }
        
        if actual.Quantity != expected.Quantity {
            return fmt.Errorf("quantity mismatch for %s: expected %.2f, got %.2f",
                expected.Symbol, expected.Quantity, actual.Quantity)
        }
    }
    
    return nil
}

func (r *PositionReconciler) ReconcileWithTrades(trades []order.Fill) error {
    // Verify portfolio state matches filled trades
    for _, trade := range trades {
        // Check if trade is reflected in portfolio
        pos := r.portfolio.GetPosition(trade.Symbol)
        if pos == nil && trade.Quantity > 0 {
            return fmt.Errorf("trade filled but position not opened: %s", trade.Symbol)
        }
    }
    
    return nil
}
```

---

### Week 2 Day 5: Integration & Testing

**Test Plan:**

1. **Unit Tests**
   - LatencySimulator produces delays in range
   - OrderQueue adds/removes orders correctly
   - PositionReconciler detects discrepancies

2. **Integration Tests**
   - PaperBroker accepts orders with latency
   - Orders move through queue: pending → accepted → filled
   - Position reconciliation works after fills

3. **Paper Trading Test**
   - Run strategy in paper mode
   - Verify realistic order lifecycle
   - Verify latency simulation
   - Verify position tracking

**File:** `internal/broker/paper_test.go`

```go
func TestPaperBroker_OrderLifecycle(t *testing.T) {
    // Setup
    executor := execution.NewSimpleExecutor(execution.Config{})
    portfolio := portfolio.NewPortfolio(10000)
    broker := NewPaperBroker(executor, portfolio, LatencyConfig{
        Min: 50 * time.Millisecond,
        Max: 100 * time.Millisecond,
    })
    
    // Submit order
    ctx := context.Background()
    ord := &order.Order{
        ID:       "test-1",
        Symbol:   "BTCUSDT",
        Side:     order.OrderSideBuy,
        Type:     order.OrderTypeMarket,
        Quantity: 1.0,
    }
    
    orderID, err := broker.SubmitOrder(ctx, ord)
    if err != nil {
        t.Fatalf("SubmitOrder failed: %v", err)
    }
    
    // Order should be pending initially
    order, _ := broker.GetOrder(ctx, orderID)
    if order.Status != order.OrderStatusPending {
        t.Errorf("Expected pending status, got %s", order.Status)
    }
    
    // Wait for latency
    time.Sleep(150 * time.Millisecond)
    
    // Process queue
    broker.ProcessOrderQueue(time.Now())
    
    // Order should be accepted now
    order, _ = broker.GetOrder(ctx, orderID)
    if order.Status != order.OrderStatusAccepted {
        t.Errorf("Expected accepted status, got %s", order.Status)
    }
}
```

---

## CLI Integration

### Update CLI to Support Paper Mode

**File:** `cmd/trader/main.go`

```go
var paperCmd = &cobra.Command{
    Use:   "paper",
    Short: "Run strategy in paper trading mode",
    Run: func(cmd *cobra.Command, args []string) {
        // Load strategy
        strategy, err := loadStrategy(strategyPath)
        if err != nil {
            log.Fatal(err)
        }
        
        // Create paper trading config
        config := runtime.Config{
            Mode: runtime.ModePaper,
            DataFeed: runtime.DataFeedConfig{
                Type: "websocket",  // Real-time data
                Params: map[string]string{
                    "exchange": "binance",
                    "symbol":   strategy.Data.Symbol,
                },
            },
            Broker: runtime.BrokerConfig{
                Type: "paper",
                Params: map[string]string{
                    "initial_balance": "10000",
                    "latency_min":     "50ms",
                    "latency_max":     "200ms",
                },
            },
        }
        
        // Run paper trading
        err = runPaperTrading(config, strategy)
        if err != nil {
            log.Fatal(err)
        }
    },
}
```

**Usage:**
```bash
trader paper --strategy=strategy.yaml
```

---

## Week 2 Timeline

| Day | Focus | Deliverable |
|-----|-------|-------------|
| **Day 1** | PaperBroker core | Basic structure + interface |
| **Day 2** | Latency simulation | LatencySimulator working |
| **Day 3** | Order queue | OrderQueue with states |
| **Day 4** | Position reconciliation | Reconciler + tests |
| **Day 5** | Integration | Full paper mode working |
| **Day 6** | CLI + tests | CLI support + integration tests |
| **Day 7** | Documentation | Week 2 report |

---

## Success Criteria

Week 2 is complete when:

1. [ ] `PaperBroker` implements `Broker` interface
2. [ ] Latency simulation works (50-200ms)
3. [ ] Order queue tracks: pending → accepted → filled
4. [ ] Position reconciliation detects discrepancies
5. [ ] CLI supports `trader paper` command
6. [ ] Integration tests pass
7. [ ] Same strategy YAML works in backtest and paper mode
8. [ ] Documentation updated

---

## Risks & Mitigations

### Risk 1: Real-time data not available yet

**Mitigation:** Use CSV with RealTime clock for initial testing. WebSocketFeed in Week 3.

### Risk 2: Concurrency issues

**Mitigation:** Use sync.RWMutex for all shared state. Test with race detector.

### Risk 3: Order queue complexity

**Mitigation:** Start simple, add features incrementally.

---

## Dependencies

**From Week 1 (Complete):**
- ✅ Broker interface
- ✅ DataFeed interface
- ✅ TimeProvider (RealTime ready)
- ✅ ExecutionMode enum

**For Week 3 (Future):**
- WebSocketFeed implementation
- Real-time candle building

---

## Next Steps (Immediate)

1. Create `internal/broker/paper.go`
2. Define PaperBroker struct
3. Implement NewPaperBroker constructor
4. Implement SubmitOrder with order queue
5. Add basic tests

**Let's start implementing!** 🚀

---

## References

- POST_MVP_PLAN.md (Phase 16 Week 2 spec)
- PHASE16_WEEK1_COMPLETE.md (Week 1 foundation)
- ARCHITECTURE.md (Broker interface)
- AGENTS.md §79 (Live trading architecture)
