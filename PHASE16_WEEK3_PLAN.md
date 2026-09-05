# Phase 16 Week 3 - Real-Time Data Feed Implementation Plan

**Date:** 2026-09-05  
**Phase:** 16 (Production Readiness)  
**Week:** 3  
**Duration:** 8 hours (2 hours × 4 days)  
**Status:** PLANNED  

---

## Objectives

Implement WebSocket real-time data feed with connection resilience, buffering, and reconnection logic for paper trading and future live trading support.

---

## Requirements (from POST_MVP_PLAN.md)

- [ ] Implement `WebSocketFeed` interface
- [ ] Add connection resilience
- [ ] Implement data buffering
- [ ] Add heartbeat/reconnection logic

---

## Technical Design

### 1. WebSocket Feed Interface

```go
// internal/data/feed/websocket.go
type WebSocketFeed struct {
    url           string
    symbols       []string
    timeframe     time.Duration
    
    conn          *websocket.Conn
    reconnecting  bool
    reconnectDelay time.Duration
    maxReconnects int
    
    buffer        *CandleBuffer
    subscribers   []chan *market.Candle
    
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
    
    lastPing      time.Time
    pingInterval  time.Duration
    pongTimeout   time.Duration
}

func NewWebSocketFeed(url string, symbols []string, timeframe time.Duration) *WebSocketFeed
func (f *WebSocketFeed) Connect() error
func (f *WebSocketFeed) Subscribe() <-chan *market.Candle
func (f *WebSocketFeed) Close() error
```

### 2. Connection Resilience

- Automatic reconnection with exponential backoff
- Max reconnect attempts (default: 10)
- Graceful degradation on repeated failures
- Connection state management

### 3. Data Buffering

```go
// internal/data/feed/buffer.go
type CandleBuffer struct {
    mu       sync.RWMutex
    candles  []*market.Candle
    maxSize  int
    overflow chan *market.Candle
}

func NewCandleBuffer(maxSize int) *CandleBuffer
func (b *CandleBuffer) Push(candle *market.Candle) error
func (b *CandleBuffer) Drain() []*market.Candle
```

### 4. Heartbeat & Reconnection

- Ping interval: 30s
- Pong timeout: 10s
- Automatic reconnection on timeout
- Heartbeat monitoring goroutine

---

## Implementation Strategy

### Day 1 (2h)
**Focus:** Core WebSocket Feed + Connection Management

**Tasks:**
1. Create `internal/data/feed/websocket.go`
2. Implement `WebSocketFeed` struct
3. Implement `Connect()` with basic connection
4. Implement `Close()` with graceful shutdown
5. Add unit tests for connection lifecycle

**Deliverables:**
- WebSocketFeed skeleton (200-250 lines)
- Basic connection tests (100-150 lines)
- Tests passing

### Day 2 (2h)
**Focus:** Reconnection Logic + Resilience

**Tasks:**
1. Implement exponential backoff reconnection
2. Add connection state machine (disconnected/connecting/connected/reconnecting)
3. Implement max retry logic
4. Add reconnection tests
5. Test failure scenarios

**Deliverables:**
- Reconnection logic (+150 lines)
- State machine implementation
- Reconnection tests (100-150 lines)
- All tests passing

### Day 3 (2h)
**Focus:** Data Buffering + Heartbeat

**Tasks:**
1. Create `internal/data/feed/buffer.go`
2. Implement `CandleBuffer` with thread-safety
3. Implement heartbeat/ping-pong mechanism
4. Add heartbeat monitoring goroutine
5. Add buffer tests

**Deliverables:**
- CandleBuffer implementation (100-150 lines)
- Heartbeat logic (+100 lines)
- Buffer tests (100-150 lines)
- All tests passing

### Day 4 (2h)
**Focus:** Integration + Documentation

**Tasks:**
1. Integrate WebSocketFeed with PaperBroker
2. Add CLI support for WebSocket data source
3. Create integration test with mock WebSocket server
4. Update ARCHITECTURE.md
5. Create completion report

**Deliverables:**
- Integration code (+100 lines)
- Integration tests (150-200 lines)
- ARCHITECTURE.md update (+200 lines)
- Week 3 completion report
- All tests passing

---

## Testing Strategy

### Unit Tests
- Connection lifecycle (connect/disconnect)
- Reconnection with various failure scenarios
- Buffer operations (push/drain/overflow)
- Heartbeat timeout detection
- State transitions

### Integration Tests
- Full WebSocket flow with mock server
- Reconnection during active data flow
- Buffer behavior under load
- Integration with PaperBroker

### Edge Cases
- Connection drops during data transmission
- Multiple reconnection attempts
- Buffer overflow
- Concurrent subscribers
- Graceful shutdown during reconnection

---

## Success Criteria

✅ WebSocketFeed connects and receives data  
✅ Automatic reconnection works with exponential backoff  
✅ Data buffering prevents data loss during reconnection  
✅ Heartbeat detects stale connections  
✅ Integration with PaperBroker works  
✅ All tests passing (unit + integration)  
✅ Documentation complete  

---

## Architecture Impact

### New Components
```
internal/data/feed/
├── websocket.go      (WebSocketFeed implementation)
├── buffer.go         (CandleBuffer implementation)
├── websocket_test.go (Unit tests)
└── integration_test.go (Integration tests)
```

### Modified Components
- `cmd/trader/main.go` - Add WebSocket data source CLI flag
- `internal/broker/paper.go` - Support WebSocket feed
- `ARCHITECTURE.md` - Document real-time data architecture

---

## Risk Assessment

### Technical Risks
1. **WebSocket library choice** - Mitigation: Use gorilla/websocket (standard)
2. **Race conditions in buffer** - Mitigation: Comprehensive mutex usage + tests
3. **Reconnection loops** - Mitigation: Max retry limit + exponential backoff
4. **Memory leaks from goroutines** - Mitigation: Proper context cancellation + WaitGroup

### Scope Risks
- **Exchange-specific protocols** - Mitigation: Start with generic protocol, document extension points
- **Authentication/API keys** - Mitigation: Phase 16 Week 4 (not Week 3)

---

## Dependencies

### External Libraries
- `github.com/gorilla/websocket` - WebSocket client (already standard)

### Internal Dependencies
- `internal/market` - Candle types
- `internal/broker` - PaperBroker integration

---

## Notes

- Focus on generic WebSocket protocol first
- Exchange-specific implementations (Binance, etc.) deferred to Week 4
- Authentication/authorization deferred to Week 4
- This week establishes the foundation for real-time data

---

**Plan Created:** 2026-09-05 14:50 UTC  
**Estimated Completion:** 2026-09-05 22:50 UTC (8 hours)
