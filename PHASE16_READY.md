# Phase 16 - Live Trading Architecture - READY TO START

**Phase 15 Status:** ✅ COMPLETE  
**Preparation Date:** 2026-09-05  
**Estimated Duration:** 3-4 weeks

---

## Phase 15 Summary - Foundation Complete

### What We Achieved
- **37.6x speedup** (50.4s → 1.34s)
- **Target exceeded by 33%** (<2s goal)
- **Golden test passed** (100% correctness)
- **Production-ready** backtest engine

### Infrastructure Ready
- ✅ Event-driven architecture
- ✅ Clean domain boundaries
- ✅ Comprehensive test suite
- ✅ Performance profiling tools
- ✅ Cached indicator system

---

## Phase 16 Overview

**Goal:** Design and implement foundation for live trading without redesigning core architecture.

**From AGENTS.md §79:**
```
                    Strategy
                       │
          ┌────────────┴────────────┐
          ▼                         ▼
   Backtest Runtime          Live Runtime
          │                         │
   Simulated Broker          Real Broker
          │                         │
          └────────────┬────────────┘
                       ▼
                   Portfolio
```

**Key Principle:** Strategy DSL must not know whether execution is simulated or real.

---

## Phase 16 Acceptance Criteria

From POST_MVP_PLAN.md:

### 1. Architecture (Week 1)
- [ ] Define Runtime interface (Backtest vs Live)
- [ ] Define Broker interface (Simulated vs Real)
- [ ] Separate time management (Historical vs Real-time)
- [ ] Event routing abstraction
- [ ] State persistence design

### 2. Real-time Data (Week 2)
- [ ] WebSocket data feed interface
- [ ] Real-time candle builder
- [ ] Multi-symbol real-time support
- [ ] Connection management
- [ ] Data quality validation

### 3. Order Management (Week 3)
- [ ] Order state machine
- [ ] Order lifecycle tracking
- [ ] Exchange API abstraction
- [ ] Rate limiting
- [ ] Error handling & retry logic

### 4. Risk & Safety (Week 4)
- [ ] Live risk controls
- [ ] Kill switch mechanism
- [ ] Position reconciliation
- [ ] Heartbeat monitoring
- [ ] Emergency shutdown

---

## Key Design Decisions Needed

### 1. Broker Abstraction
**Question:** How to unify simulated vs real execution?

**Options:**
```go
type Broker interface {
    SubmitOrder(order *Order) error
    CancelOrder(orderID string) error
    GetOrder(orderID string) (*Order, error)
    GetPositions() ([]*Position, error)
    GetBalance() (*Balance, error)
}

// Implementations:
// - SimulatedBroker (current)
// - LiveBroker (new)
```

### 2. Time Management
**Question:** How to handle historical vs real-time?

**Options:**
```go
type TimeProvider interface {
    Now() time.Time
    Sleep(duration time.Duration)
    After(duration time.Duration) <-chan time.Time
}

// Implementations:
// - HistoricalTime (backtest)
// - RealTime (live)
```

### 3. Data Feed
**Question:** How to unify CSV vs WebSocket?

**Options:**
```go
type DataFeed interface {
    Subscribe(symbols []string) error
    Next() (*MarketEvent, error)
    Close() error
}

// Implementations:
// - CSVFeed (current)
// - WebSocketFeed (new)
```

### 4. State Management
**Question:** How to persist strategy state?

**Options:**
- File-based (JSON/SQLite)
- In-memory with snapshots
- Database (PostgreSQL)

---

## Architecture Goals

### Must Have
1. **Zero strategy code changes** for live vs backtest
2. **Same evaluation engine** for both modes
3. **Same indicators** for both modes
4. **Same signal generation** for both modes
5. **Same risk management** for both modes

### Separation of Concerns
```
Strategy Layer (unchanged)
      ↓
Evaluation Layer (unchanged)
      ↓
Signal Layer (unchanged)
      ↓
Runtime Layer (NEW - backtest vs live)
      ↓
Broker Layer (ABSTRACTED)
      ↓
Execution Layer (DIFFERENT)
```

### Critical Invariants
1. No look-ahead bias (same as backtest)
2. Deterministic evaluation (where possible)
3. Testable without real exchange
4. Graceful degradation
5. Observable state

---

## Implementation Strategy

### Week 1: Architecture Foundation
**Focus:** Define interfaces, no implementation yet

**Deliverables:**
- Runtime interface design
- Broker interface design
- Time provider interface
- Data feed abstraction
- Architecture document
- Interface tests (mocks)

### Week 2: Real-time Data
**Focus:** Implement WebSocket data handling

**Deliverables:**
- WebSocket client
- Candle builder (tick → OHLCV)
- Connection management
- Reconnection logic
- Data validation

### Week 3: Order Management
**Focus:** Implement real broker interaction

**Deliverables:**
- Exchange API client (Binance/Bybit)
- Order state tracking
- Fill detection
- Position sync
- Error handling

### Week 4: Risk & Safety
**Focus:** Production safety controls

**Deliverables:**
- Live risk limits
- Kill switch
- Heartbeat monitor
- Position reconciliation
- Emergency procedures document

---

## Testing Strategy

### Unit Tests
- Runtime interface implementations
- Broker interface implementations
- Time provider logic
- Order state machine

### Integration Tests
- Strategy → Live Runtime → Mock Broker
- Real-time data → Indicator → Signal
- Order submission → Fill detection
- Position tracking

### Simulation Tests
- Mock exchange with WebSocket
- Latency simulation
- Connection failures
- Partial fills
- Order rejections

### Manual Tests (Paper Trading)
- Real exchange testnet
- Real WebSocket data
- Real order lifecycle
- Real position tracking
- No real money

---

## Risks & Mitigations

### Risk 1: Architecture Too Complex
**Mitigation:** Start with simplest working design, refactor when necessary

### Risk 2: Exchange-Specific Issues
**Mitigation:** Support only Binance initially, abstract later

### Risk 3: Real-time Performance
**Mitigation:** Reuse Phase 15 cached indicators

### Risk 4: State Synchronization
**Mitigation:** Reconcile positions on every tick

### Risk 5: Network Failures
**Mitigation:** Graceful degradation, never lose state

---

## Success Criteria

Phase 16 is complete when:

1. [ ] Strategy from Phase 8 runs **unchanged** in live mode
2. [ ] Same signals generated (backtest vs live with same historical data)
3. [ ] Orders submitted to testnet successfully
4. [ ] Positions tracked correctly
5. [ ] Kill switch works
6. [ ] All tests passing
7. [ ] Architecture document complete
8. [ ] Paper trading guide written

**Quality Bar:** Must be safe enough for paper trading on testnet.

---

## Out of Scope (Future Phases)

**Not in Phase 16:**
- Production exchange credentials
- Real money trading
- Multi-account support
- Advanced order types (iceberg, etc)
- Machine learning integration
- Web dashboard
- Mobile app

**But architecture should not prevent them.**

---

## Prerequisites

From Phase 15:
- ✅ Fast backtest engine (1.34s)
- ✅ Cached indicators
- ✅ Event-driven architecture
- ✅ Clean domain model
- ✅ Comprehensive tests

**All prerequisites met.**

---

## Next Steps

1. Read POST_MVP_PLAN.md Phase 16 section
2. Review AGENTS.md §79 (Live Trading Architecture)
3. Design Runtime interface
4. Design Broker interface
5. Create architecture document
6. Implement Week 1 deliverables

---

## Questions to Answer

Before starting implementation:

1. Which exchange API to support first? (Binance/Bybit)
2. WebSocket library? (gorilla/websocket vs nhooyr.io/websocket)
3. State persistence? (JSON file vs SQLite)
4. Configuration format? (extend strategy.yaml?)
5. Logging approach? (structured logging?)

**Decision:** Start minimal, extend when needed.

---

## Estimated Timeline

| Week | Deliverable | Risk |
|------|-------------|------|
| 1 | Architecture & Interfaces | Low |
| 2 | Real-time Data | Medium |
| 3 | Order Management | High |
| 4 | Risk & Safety | Medium |

**Total:** 3-4 weeks (same as Phase 15)

---

## Phase 15 → Phase 16 Transition

**What stays the same:**
- Strategy DSL
- Indicator engine
- Expression evaluator
- Signal generation
- Risk management logic

**What changes:**
- Runtime (backtest → live)
- Broker (simulated → real)
- Data (CSV → WebSocket)
- Time (historical → real-time)

**Impact:** Minimal - good architecture pays off.

---

## Confidence Level

**Architecture Design:** High (clear separation exists)  
**Implementation:** Medium (new WebSocket/API territory)  
**Testing:** High (simulation possible)  
**Safety:** High (testnet + kill switch)

**Overall:** Medium-High

---

**Phase 16 is the most critical phase before v1.0.0. Let's build it right.** 🚀

---

## References

- POST_MVP_PLAN.md (Phase 16 details)
- AGENTS.md §79 (Live Trading Architecture)
- AGENTS.md §16-34 (Order/Execution/Broker model)
- GitHub Issue #2 (Live Trading - Critical)
