# Phase 16 - Week 1: Core Abstractions

**Week:** 1 of 4  
**Start Date:** 2026-09-05  
**Focus:** Architecture & Interface Design  
**Risk Level:** Low

---

## Week 1 Objectives

From POST_MVP_PLAN.md Phase 16:

- [ ] Define `Broker` interface
- [ ] Define `DataFeed` interface  
- [ ] Refactor `SimulatedBroker` to implement interface
- [ ] Add execution mode configuration

**Goal:** Establish clean abstractions WITHOUT breaking existing backtest functionality.

---

## Current Architecture Analysis

### What We Have (Phase 8)

```
internal/
├── backtest/
│   └── engine.go          # Main backtest loop
├── broker/
│   └── simulated.go       # SimulatedBroker (tightly coupled)
├── data/
│   └── csv/
│       └── reader.go      # CSV data loading
└── portfolio/
    └── portfolio.go       # Position tracking
```

### Problems

1. **SimulatedBroker is concrete**, not interface-based
2. **CSV reader is the only data source**
3. **No execution mode concept** (always backtest)
4. **Time management is implicit** (no abstraction)
5. **Broker and Portfolio are tightly coupled**

### What We Need

```
internal/
├── runtime/               # NEW
│   ├── mode.go           # ExecutionMode enum
│   ├── broker.go         # Broker interface
│   ├── datafeed.go       # DataFeed interface
│   └── time.go           # TimeProvider interface
│
├── broker/
│   ├── broker.go         # Interface definition (moved)
│   ├── simulated.go      # Implements Broker (refactored)
│   ├── paper.go          # NEW (Week 2)
│   └── live.go           # NEW (Week 3)
│
├── data/
│   ├── feed.go           # DataFeed interface (moved)
│   ├── csv/
│   │   └── reader.go     # Implements DataFeed (refactored)
│   └── websocket/        # NEW (Week 3)
│       └── client.go
│
└── backtest/
    └── engine.go         # Uses Broker interface (refactored)
```

---

## Design Decisions

### 1. Broker Interface

**Location:** `internal/broker/broker.go`

**Design:**
```go
package broker

import (
    "context"
    "time"
    
    "github.com/ZulferDev/smallbt_go/internal/order"
    "github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// Broker abstracts order execution (simulated, paper, live)
type Broker interface {
    // Order lifecycle
    SubmitOrder(ctx context.Context, o *order.Order) (string, error)
    CancelOrder(ctx context.Context, orderID string) error
    GetOrder(ctx context.Context, orderID string) (*order.Order, error)
    
    // Position & balance queries
    GetPositions(ctx context.Context) ([]*portfolio.Position, error)
    GetBalance(ctx context.Context) (*portfolio.Balance, error)
    
    // Market data (for order validation)
    GetLastPrice(ctx context.Context, symbol string) (float64, error)
    
    // Lifecycle
    Close() error
}

// OrderID is a unique identifier
type OrderID string
```

**Rationale:**
- Context-aware (cancellation, timeouts)
- Order and Position types already exist
- GetLastPrice needed for order validation
- Close() for cleanup (WebSocket connections, etc)

---

### 2. DataFeed Interface

**Location:** `internal/data/feed.go`

**Design:**
```go
package data

import (
    "context"
    "time"
    
    "github.com/ZulferDev/smallbt_go/internal/market"
)

// DataFeed provides market data (historical or real-time)
type DataFeed interface {
    // Subscribe to symbols (no-op for CSV, required for WebSocket)
    Subscribe(ctx context.Context, symbols []string) error
    
    // Next returns the next candle chronologically
    // Returns io.EOF when no more data
    Next(ctx context.Context) (*market.Candle, error)
    
    // Close releases resources
    Close() error
}

// FeedType distinguishes data sources
type FeedType string

const (
    FeedTypeCSV       FeedType = "csv"
    FeedTypeParquet   FeedType = "parquet"
    FeedTypeWebSocket FeedType = "websocket"
    FeedTypeREST      FeedType = "rest"
)
```

**Rationale:**
- Context for cancellation
- Subscribe() prepares WebSocket, no-op for CSV
- Next() unified interface (historical & real-time)
- io.EOF signals end (backtest done)

---

### 3. ExecutionMode

**Location:** `internal/runtime/mode.go`

**Design:**
```go
package runtime

// ExecutionMode defines how the strategy runs
type ExecutionMode string

const (
    ModeBacktest ExecutionMode = "backtest" // Historical simulation
    ModePaper    ExecutionMode = "paper"    // Real-time simulation
    ModeLive     ExecutionMode = "live"     // Real exchange
)

// Config holds runtime configuration
type Config struct {
    Mode ExecutionMode
    
    // Data source
    DataFeed DataFeedConfig
    
    // Broker
    Broker BrokerConfig
    
    // Risk limits (live mode only)
    RiskLimits *RiskLimits
}

type DataFeedConfig struct {
    Type   string            // "csv", "websocket"
    Params map[string]string // type-specific params
}

type BrokerConfig struct {
    Type   string            // "simulated", "paper", "live"
    Params map[string]string // type-specific params
}

type RiskLimits struct {
    MaxPositionSize  float64
    MaxDailyLoss     float64
    RequireConfirm   bool
}
```

**Rationale:**
- Clear separation of concerns
- Extensible via Params maps
- RiskLimits only in live mode

---

### 4. TimeProvider Interface

**Location:** `internal/runtime/time.go`

**Design:**
```go
package runtime

import "time"

// TimeProvider abstracts time (historical vs real-time)
type TimeProvider interface {
    Now() time.Time
    Sleep(d time.Duration)
    After(d time.Duration) <-chan time.Time
}

// HistoricalTime uses simulated time (backtest)
type HistoricalTime struct {
    current time.Time
}

func (h *HistoricalTime) Now() time.Time {
    return h.current
}

func (h *HistoricalTime) Sleep(d time.Duration) {
    // No-op in backtest
}

func (h *HistoricalTime) After(d time.Duration) <-chan time.Time {
    ch := make(chan time.Time, 1)
    ch <- h.current.Add(d)
    return ch
}

func (h *HistoricalTime) Advance(t time.Time) {
    h.current = t
}

// RealTime uses system time (paper/live)
type RealTime struct{}

func (r *RealTime) Now() time.Time {
    return time.Now()
}

func (r *RealTime) Sleep(d time.Duration) {
    time.Sleep(d)
}

func (r *RealTime) After(d time.Duration) <-chan time.Time {
    return time.After(d)
}
```

**Rationale:**
- Testable (inject HistoricalTime in tests)
- Backtest uses Advance() to control time
- Paper/live use real system clock

---

## Implementation Tasks

### Task 1: Create runtime package

**Files:**
```
internal/runtime/
├── mode.go       # ExecutionMode enum + Config
├── time.go       # TimeProvider interface
├── broker.go     # (will be moved from broker/)
└── datafeed.go   # (will be moved from data/)
```

**Tests:**
```
internal/runtime/
├── mode_test.go
└── time_test.go
```

### Task 2: Define Broker interface

**File:** `internal/broker/broker.go`

**Content:**
- Broker interface definition
- OrderID type
- Error types (OrderNotFound, InsufficientBalance, etc)

**Tests:** Mock-based unit tests

### Task 3: Define DataFeed interface

**File:** `internal/data/feed.go`

**Content:**
- DataFeed interface definition
- FeedType enum
- Common errors (EOF, ConnectionLost, etc)

**Tests:** Mock-based unit tests

### Task 4: Refactor SimulatedBroker

**Current:** `internal/broker/simulated.go` (concrete)

**Goal:** Implement `Broker` interface

**Changes:**
1. Add `SubmitOrder(ctx, order) (string, error)` signature
2. Add `GetOrder(ctx, orderID)` method
3. Add `GetLastPrice(ctx, symbol)` method
4. Add `Close()` method
5. Keep existing logic (fees, slippage, fills)

**Tests:** Update existing tests to verify interface compliance

### Task 5: Refactor CSV Reader

**Current:** `internal/data/csv/reader.go`

**Goal:** Implement `DataFeed` interface

**Changes:**
1. Add `Subscribe(ctx, symbols)` (no-op)
2. Rename `ReadAll()` logic into `Next(ctx)` loop
3. Add `Close()` method
4. Return `io.EOF` when done

**Tests:** Update existing tests

### Task 6: Update Backtest Engine

**File:** `internal/backtest/engine.go`

**Changes:**
1. Accept `Broker` interface instead of concrete SimulatedBroker
2. Accept `DataFeed` interface instead of CSV reader
3. Accept `TimeProvider` interface
4. Use `datafeed.Next(ctx)` instead of for loop

**Tests:** Verify golden test still passes

---

## Testing Strategy

### Unit Tests
- [ ] runtime.Config parsing
- [ ] TimeProvider implementations
- [ ] Broker interface (mock)
- [ ] DataFeed interface (mock)

### Integration Tests
- [ ] Backtest engine + SimulatedBroker (via interface)
- [ ] Backtest engine + CSV feed (via interface)
- [ ] Same strategy, different broker implementations

### Regression Tests
- [ ] **Golden test must still pass**
- [ ] All Phase 8 tests still passing
- [ ] No performance regression

---

## Success Criteria

Week 1 is complete when:

1. [ ] `Broker` interface defined and documented
2. [ ] `DataFeed` interface defined and documented
3. [ ] `ExecutionMode` enum and Config implemented
4. [ ] `TimeProvider` interface implemented
5. [ ] `SimulatedBroker` implements `Broker` interface
6. [ ] CSV reader implements `DataFeed` interface
7. [ ] Backtest engine uses interfaces (not concrete types)
8. [ ] **All existing tests pass** (including golden test)
9. [ ] Architecture document updated
10. [ ] Zero breaking changes to strategy YAML

**Quality bar:** Existing functionality unchanged, interfaces proven with tests.

---

## Risks & Mitigations

### Risk 1: Interface too complex
**Mitigation:** Start minimal, add methods when needed

### Risk 2: Breaking existing tests
**Mitigation:** Refactor incrementally, run tests after each change

### Risk 3: Performance regression
**Mitigation:** Benchmark before/after, interface overhead should be negligible

### Risk 4: Over-abstraction
**Mitigation:** YAGNI - only abstract what we know we need

---

## Rollback Plan

If Week 1 fails:
1. Revert to last working commit (Phase 15 complete)
2. Re-evaluate interface design
3. Consider simpler approach

**Git branch strategy:**
```bash
git checkout -b phase16-week1
# Work here
# When ready:
git checkout main
git merge phase16-week1
```

---

## Documentation Updates

### Files to Update
- [ ] AGENTS.md (add runtime architecture section)
- [ ] README.md (mention execution modes)
- [ ] Architecture diagram (add runtime layer)

### Files to Create
- [ ] ARCHITECTURE.md (new - comprehensive design doc)
- [ ] PHASE16_WEEK1_REPORT.md (at end of week)

---

## Timeline

| Day | Focus | Deliverable |
|-----|-------|-------------|
| **Day 1** | Interface design | Broker & DataFeed interfaces |
| **Day 2** | Runtime package | ExecutionMode + TimeProvider |
| **Day 3** | SimulatedBroker refactor | Implements Broker interface |
| **Day 4** | CSV reader refactor | Implements DataFeed interface |
| **Day 5** | Engine integration | Uses interfaces |
| **Day 6** | Testing & verification | All tests passing |
| **Day 7** | Documentation | Week 1 report |

**Today:** Day 1 (2026-09-05)

---

## Next Steps (Immediate)

1. Create `internal/runtime/` package
2. Define `ExecutionMode` enum
3. Define `TimeProvider` interface
4. Implement `HistoricalTime` and `RealTime`
5. Write tests

**Let's start with runtime package foundation.** 🚀

---

## References

- POST_MVP_PLAN.md (Phase 16 specification)
- AGENTS.md §79 (Live Trading Architecture)
- AGENTS.md §23-24 (Order Model)
- AGENTS.md §36-37 (Data Layer)
- PHASE15_COMPLETE.md (Previous phase baseline)
