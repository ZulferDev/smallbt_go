# Phase 16 Week 2 Day 2 - Implementation Plan

**Date:** 2026-09-05  
**Focus:** Position Reconciliation & Background Processing  
**Duration:** 3-4 hours  

---

## Day 2 Objectives

1. ✅ **Position Reconciliation** (SKIP - Portfolio already tracks positions correctly)
2. 🔄 **Background Queue Processing** (goroutine-based)
3. 🔄 **Real-time DataFeed Implementation**
4. 🔄 **Integration Tests** (full paper trading loop)

---

## Analysis: Position Reconciliation

**Question:** Do we need a separate PositionReconciler?

**Current State:**
```go
// PaperBroker.ProcessOrderQueue() already:
1. Gets accepted orders
2. Calls executor.Execute()
3. Executor creates Fill
4. Fill updates Portfolio
5. Portfolio tracks positions correctly
```

**Portfolio Methods:**
```go
portfolio.OpenPosition()   // Called on buy fill
portfolio.ClosePosition()  // Called on sell fill
portfolio.GetPositions()   // Returns current positions
portfolio.GetPosition(symbol) // Get specific position
```

**Conclusion:**
- ✅ Portfolio already reconciles automatically through Fill events
- ✅ No separate reconciliation needed for MVP
- ⏳ Can add validation/audit layer later if needed

**Decision:** Skip PositionReconciler for now. Focus on background processing.

---

## Task 1: Background Queue Processing

**Goal:** PaperBroker processes orders continuously in background goroutine.

### Current Architecture

```
User                PaperBroker              OrderQueue
 │                       │                       │
 ├─ SubmitOrder()────────>                       │
 │                       ├─ Add()────────────────>
 │                       │                       │
 │                       │                       │
 ?  When to call ProcessOrderQueue()?           │
 ?  Manual or automatic?                        │
```

### Target Architecture

```
User                PaperBroker              OrderQueue
 │                       │                       │
 ├─ SubmitOrder()────────>                       │
 │                       ├─ Add()────────────────>
 │                       │                       │
 │                  [Background goroutine]       │
 │                       │                       │
 │                  ProcessOrderQueue()          │
 │                  every 100ms                  │
 │                       │                       │
 │                       ├─ Accept pending       │
 │                       ├─ Fill accepted        │
 │                       └─ Update portfolio     │
```

### Implementation

**File:** `internal/broker/paper.go`

```go
type PaperBroker struct {
    // ... existing fields ...
    
    // Background processing
    ticker      *time.Ticker
    stopCh      chan struct{}
    stoppedCh   chan struct{}
    wg          sync.WaitGroup
}

func NewPaperBroker(
    executor *execution.SimpleExecutor,
    portfolio *portfolio.Portfolio,
    config LatencyConfig,
) *PaperBroker {
    b := &PaperBroker{
        orderManager: order.NewOrderManager(),
        executor:     executor,
        portfolio:    portfolio,
        orderQueue:   NewOrderQueue(),
        latencySim:   NewLatencySimulator(config),
        lastPrices:   make(map[string]float64),
        stopCh:       make(chan struct{}),
        stoppedCh:    make(chan struct{}),
    }
    
    // Start background processing
    b.startBackgroundProcessing()
    
    return b
}

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

func (b *PaperBroker) processOrderQueueBackground() {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if b.closed {
        return
    }
    
    _ = b.ProcessOrderQueue(time.Now())
}

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
    
    // Wait for goroutine to finish
    b.wg.Wait()
    
    return nil
}
```

**Tests:**

```go
func TestPaperBroker_BackgroundProcessing(t *testing.T) {
    // Use short fixed latency
    config := LatencyConfig{
        MinLatency: 50 * time.Millisecond,
        MaxLatency: 50 * time.Millisecond,
    }
    
    executor := execution.NewSimpleExecutor(execution.Config{})
    port := portfolio.NewPortfolio(10000)
    broker := NewPaperBroker(executor, port, config)
    defer broker.Close()
    
    // Update price
    broker.UpdatePrice("BTCUSDT", 50000.0)
    
    // Submit order
    ctx := context.Background()
    ord := &order.Order{
        Symbol:   "BTCUSDT",
        Side:     order.OrderSideBuy,
        Type:     order.OrderTypeMarket,
        Quantity: 0.1,
    }
    
    orderID, err := broker.SubmitOrder(ctx, ord)
    if err != nil {
        t.Fatalf("SubmitOrder failed: %v", err)
    }
    
    // Wait for background processing (latency + processing time)
    time.Sleep(200 * time.Millisecond)
    
    // Order should be filled automatically
    qo, exists := broker.orderQueue.Get(orderID)
    if !exists {
        t.Fatal("Order not found")
    }
    
    if qo.Status != StatusFilled {
        t.Errorf("Expected filled, got %s", qo.Status)
    }
    
    // Verify position opened
    positions, err := broker.GetPositions(ctx)
    if err != nil {
        t.Fatalf("GetPositions failed: %v", err)
    }
    
    if len(positions) != 1 {
        t.Errorf("Expected 1 position, got %d", len(positions))
    }
}

func TestPaperBroker_BackgroundProcessing_MultipleOrders(t *testing.T) {
    // Submit multiple orders
    // Verify all processed in background
    // Verify correct order (by accept time)
}

func TestPaperBroker_Close_StopsBackgroundProcessing(t *testing.T) {
    // Create broker
    // Close it
    // Submit order (should fail)
    // Verify goroutine stopped
}
```

---

## Task 2: Real-time DataFeed

**Goal:** Create TickerDataFeed that provides real-time price updates.

### Design

**File:** `internal/broker/ticker_feed.go`

```go
// TickerDataFeed simulates real-time price updates
// Used for paper trading and live trading
type TickerDataFeed struct {
    symbols    []string
    interval   time.Duration
    priceGen   PriceGenerator
    
    ticker     *time.Ticker
    stopCh     chan struct{}
    priceCh    chan PriceUpdate
    
    mu         sync.RWMutex
    closed     bool
}

type PriceUpdate struct {
    Symbol string
    Price  float64
    Time   time.Time
}

type PriceGenerator interface {
    NextPrice(symbol string, prevPrice float64) float64
}

// RandomWalkGenerator simulates price movement
type RandomWalkGenerator struct {
    volatility float64  // e.g., 0.001 = 0.1% per tick
    rand       *rand.Rand
}

func NewTickerDataFeed(
    symbols []string,
    interval time.Duration,
    priceGen PriceGenerator,
) *TickerDataFeed {
    return &TickerDataFeed{
        symbols:  symbols,
        interval: interval,
        priceGen: priceGen,
        stopCh:   make(chan struct{}),
        priceCh:  make(chan PriceUpdate, 100),
    }
}

func (f *TickerDataFeed) Start(initialPrices map[string]float64) {
    f.ticker = time.NewTicker(f.interval)
    
    go func() {
        lastPrices := initialPrices
        
        for {
            select {
            case <-f.ticker.C:
                for _, symbol := range f.symbols {
                    prevPrice := lastPrices[symbol]
                    newPrice := f.priceGen.NextPrice(symbol, prevPrice)
                    lastPrices[symbol] = newPrice
                    
                    f.priceCh <- PriceUpdate{
                        Symbol: symbol,
                        Price:  newPrice,
                        Time:   time.Now(),
                    }
                }
            case <-f.stopCh:
                close(f.priceCh)
                return
            }
        }
    }()
}

func (f *TickerDataFeed) PriceUpdates() <-chan PriceUpdate {
    return f.priceCh
}

func (f *TickerDataFeed) Close() error {
    f.mu.Lock()
    defer f.mu.Unlock()
    
    if f.closed {
        return nil
    }
    f.closed = true
    
    if f.ticker != nil {
        f.ticker.Stop()
    }
    close(f.stopCh)
    
    return nil
}
```

**Simple Implementation for MVP:**

```go
// For MVP, use static prices or simple updates
func NewStaticPriceGenerator(prices map[string]float64) PriceGenerator {
    return &staticGenerator{prices: prices}
}

type staticGenerator struct {
    prices map[string]float64
}

func (g *staticGenerator) NextPrice(symbol string, prevPrice float64) float64 {
    if price, ok := g.prices[symbol]; ok {
        return price
    }
    return prevPrice
}
```

---

## Task 3: Integration Tests

**Goal:** Test full paper trading loop with background processing.

**File:** `internal/broker/integration_test.go`

```go
func TestPaperTrading_FullLoop(t *testing.T) {
    // Setup
    config := LatencyConfig{
        MinLatency: 50 * time.Millisecond,
        MaxLatency: 100 * time.Millisecond,
        Seed:       42,
    }
    
    executor := execution.NewSimpleExecutor(execution.Config{})
    port := portfolio.NewPortfolio(10000)
    broker := NewPaperBroker(executor, port, config)
    defer broker.Close()
    
    // Setup price feed
    feed := NewTickerDataFeed(
        []string{"BTCUSDT"},
        100 * time.Millisecond,
        NewStaticPriceGenerator(map[string]float64{
            "BTCUSDT": 50000.0,
        }),
    )
    defer feed.Close()
    
    feed.Start(map[string]float64{"BTCUSDT": 50000.0})
    
    // Subscribe to price updates
    ctx := context.Background()
    go func() {
        for update := range feed.PriceUpdates() {
            broker.UpdatePrice(update.Symbol, update.Price)
        }
    }()
    
    // Submit buy order
    buyOrder := &order.Order{
        Symbol:   "BTCUSDT",
        Side:     order.OrderSideBuy,
        Type:     order.OrderTypeMarket,
        Quantity: 0.1,
    }
    
    buyID, err := broker.SubmitOrder(ctx, buyOrder)
    if err != nil {
        t.Fatalf("Buy order failed: %v", err)
    }
    
    // Wait for order to fill (latency + processing)
    time.Sleep(300 * time.Millisecond)
    
    // Verify buy filled
    buyQO, _ := broker.orderQueue.Get(buyID)
    if buyQO.Status != StatusFilled {
        t.Errorf("Buy order not filled: %s", buyQO.Status)
    }
    
    // Verify position opened
    positions, _ := broker.GetPositions(ctx)
    if len(positions) != 1 {
        t.Fatalf("Expected 1 position, got %d", len(positions))
    }
    
    if positions[0].Quantity != 0.1 {
        t.Errorf("Expected quantity 0.1, got %.2f", positions[0].Quantity)
    }
    
    // Submit sell order
    sellOrder := &order.Order{
        Symbol:   "BTCUSDT",
        Side:     order.OrderSideSell,
        Type:     order.OrderTypeMarket,
        Quantity: 0.1,
    }
    
    sellID, err := broker.SubmitOrder(ctx, sellOrder)
    if err != nil {
        t.Fatalf("Sell order failed: %v", err)
    }
    
    // Wait for sell to fill
    time.Sleep(300 * time.Millisecond)
    
    // Verify sell filled
    sellQO, _ := broker.orderQueue.Get(sellID)
    if sellQO.Status != StatusFilled {
        t.Errorf("Sell order not filled: %s", sellQO.Status)
    }
    
    // Verify position closed
    positions, _ = broker.GetPositions(ctx)
    if len(positions) != 0 {
        t.Errorf("Expected 0 positions, got %d", len(positions))
    }
}

func TestPaperTrading_MultipleSymbols(t *testing.T) {
    // Test with BTCUSDT and ETHUSDT
    // Submit orders for both
    // Verify independent tracking
}

func TestPaperTrading_PriceMovement(t *testing.T) {
    // Use RandomWalkGenerator
    // Submit limit orders
    // Verify fills at correct prices
}
```

---

## Implementation Order

1. **Add background processing to PaperBroker** (1 hour)
   - Add goroutine fields
   - Implement startBackgroundProcessing()
   - Update Close() to stop goroutine
   - Add tests

2. **Create TickerDataFeed** (1 hour)
   - Implement basic structure
   - Add StaticPriceGenerator for MVP
   - Add simple tests

3. **Integration tests** (1 hour)
   - Full loop test
   - Multi-symbol test
   - Verify position tracking

4. **Documentation** (30 min)
   - Update ARCHITECTURE.md
   - Add usage examples
   - Day 2 completion report

---

## Success Criteria

✅ PaperBroker processes orders automatically in background  
✅ TickerDataFeed provides simulated price updates  
✅ Integration test demonstrates full paper trading loop  
✅ All tests passing (existing + new)  
✅ No breaking changes to existing code  
✅ Documentation updated  

---

## Timeline

**Total:** 3.5 hours

- Background processing: 1 hour
- TickerDataFeed: 1 hour  
- Integration tests: 1 hour
- Documentation: 30 min

---

## Next Steps (Day 3)

After Day 2 complete:
- CLI integration (`trader paper` command)
- Real exchange data integration
- Enhanced price generators
- Performance testing

---

**Status:** Ready to implement
