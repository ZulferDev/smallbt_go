# smallbt_go Architecture

**Project:** Declarative Quantitative Trading Backtesting Engine  
**Language:** Go  
**Status:** Active Development - Phase 16 (Live Trading Architecture)  
**Version:** Pre-v1.0.0

---

## Table of Contents

1. [Overview](#overview)
2. [Design Principles](#design-principles)
3. [Core Architecture](#core-architecture)
4. [Runtime Layer](#runtime-layer)
5. [Data Layer](#data-layer)
6. [Strategy Layer](#strategy-layer)
7. [Execution Layer](#execution-layer)
8. [Portfolio Layer](#portfolio-layer)
9. [Analytics Layer](#analytics-layer)
10. [Future Architecture](#future-architecture)

---

## Overview

smallbt_go is a quantitative trading research engine where strategies are defined declaratively through YAML configuration instead of hardcoded logic. The system supports backtesting, parameter optimization, walk-forward analysis, and is designed to support paper trading and live trading without architectural redesign.

### Key Characteristics

- **Declarative:** Strategies defined in YAML, not Go code
- **Deterministic:** Same input produces same output
- **Extensible:** Indicators, functions, and data sources are pluggable
- **Fast:** <2s backtest for 5 years hourly data (Phase 15)
- **Mode-agnostic:** Same strategy runs in backtest, paper, and live modes

---

## Design Principles

### 1. YAML is an Interface, Not the Engine

```
YAML Strategy Definition
        ↓
    Parser
        ↓
Strategy AST / IR
        ↓
Dependency Resolution
        ↓
    Evaluator
        ↓
    Signal
        ↓
    Risk
        ↓
    Order
        ↓
  Execution
        ↓
  Portfolio
        ↓
  Analytics
```

The engine does NOT depend on YAML structures. YAML compiles to an intermediate representation (AST) that the engine evaluates.

### 2. Interface-Based Design

All critical components are interface-based:
- `Broker` - abstracts order execution (simulated/paper/live)
- `DataFeed` - abstracts market data (CSV/WebSocket/REST)
- `TimeProvider` - abstracts time (historical/real-time)

This enables:
- Easy testing (mock implementations)
- Multiple execution modes without code changes
- Future extensibility (new data sources, exchanges)

### 3. Backward Compatibility

Breaking changes are avoided. When refactoring:
- New interfaces wrap existing implementations
- Old APIs maintained via adapter/wrapper patterns
- Tests verify identical behavior

### 4. No Look-Ahead Bias

The system guarantees:
- `signal[t]` only uses data from `t` and earlier
- Indicators have explicit evaluation timing
- Tests detect accidental look-ahead

---

## Core Architecture

### High-Level Layers

```
┌─────────────────────────────────────────┐
│         Strategy Definition             │
│            (YAML / JSON)                │
└──────────────┬──────────────────────────┘
               ↓
┌─────────────────────────────────────────┐
│          Parser & Compiler              │
│    (YAML → AST → Intermediate Rep)      │
└──────────────┬──────────────────────────┘
               ↓
┌─────────────────────────────────────────┐
│         Runtime Layer (NEW)             │
│   ExecutionMode | TimeProvider          │
└──────────────┬──────────────────────────┘
               ↓
┌──────────────────────┬──────────────────┐
│                      │                  │
│   Data Layer         │   Execution      │
│   (DataFeed)         │   (Broker)       │
│                      │                  │
└──────────┬───────────┴─────────┬────────┘
           ↓                     ↓
┌─────────────────┐    ┌──────────────────┐
│   Indicators    │    │   Portfolio      │
│   (Cached)      │    │   (Positions)    │
└─────────────────┘    └──────────────────┘
           ↓                     ↓
┌─────────────────────────────────────────┐
│            Analytics & Reports          │
└─────────────────────────────────────────┘
```

---

## Runtime Layer

**Added:** Phase 16 Week 1  
**Location:** `internal/runtime/`

The runtime layer abstracts execution mode (backtest vs paper vs live) and time management.

### ExecutionMode

```go
type ExecutionMode string

const (
    ModeBacktest ExecutionMode = "backtest" // Historical simulation
    ModePaper    ExecutionMode = "paper"    // Real-time simulation
    ModeLive     ExecutionMode = "live"     // Real exchange
)
```

### Config

```go
type Config struct {
    Mode       ExecutionMode
    DataFeed   DataFeedConfig
    Broker     BrokerConfig
    RiskLimits *RiskLimits  // Required for live mode
}
```

### TimeProvider Interface

```go
type TimeProvider interface {
    Now() time.Time
    Sleep(d time.Duration)
    After(d time.Duration) <-chan time.Time
}
```

**Implementations:**
- `HistoricalTime` - controlled time for backtesting
- `RealTime` - system clock for paper/live trading

**Why:** Enables deterministic backtests (controlled time) and real-time paper/live trading (system clock) using the same engine code.

---

## Data Layer

**Location:** `internal/data/`

### DataFeed Interface

**Added:** Phase 16 Week 1

```go
type DataFeed interface {
    Subscribe(ctx context.Context, symbols []string) error
    Next(ctx context.Context) (*market.Candle, error)
    Close() error
}
```

**Implementations:**
- `CSVDataFeed` - historical data from CSV files
- `ParquetDataFeed` - historical data from Parquet (Phase 17)
- `WebSocketFeed` - real-time data via WebSocket (Phase 16 Week 3)
- `RESTFeed` - real-time data via REST API (future)

**Design:**
- Context-aware for cancellation/timeout
- Returns `io.EOF` when done (backtest)
- Blocks until next candle (real-time)

### CSV Implementation

```go
type CSVDataFeed struct {
    feed   *CSVFeed   // Wraps existing CSV reader
    closed bool
}
```

**Features:**
- Validates chronological order
- Validates OHLC relationships
- Auto-detects timestamp formats
- Supports custom column mappings

---

## Strategy Layer

**Location:** `internal/strategy/`

### Strategy Definition (YAML)

```yaml
strategy:
  name: ema_volume_trend
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema_fast:
    type: ema
    source: close
    period: 9
  
  ema_slow:
    type: ema
    source: close
    period: 21

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [volume, 1000000]

risk:
  position_size:
    type: risk_percent
    value: 0.01
  
  stop_loss:
    type: atr
    period: 14
    multiplier: 1.5
```

### Strategy AST

The parser converts YAML to an Abstract Syntax Tree (AST):

```go
type Strategy struct {
    Name        string
    Version     string
    Data        DataConfig
    Indicators  map[string]IndicatorConfig
    Entry       EntryConfig
    Exit        ExitConfig
    Risk        RiskConfig
    Execution   ExecutionConfig
}
```

### Evaluator

**Location:** `internal/strategy/evaluator/`

Two evaluator implementations:
- `Evaluator` - stateless, recomputes all indicators
- `CachedEvaluator` - stateful, incremental updates (Phase 15)

**CachedEvaluator Performance:**
- 37.6x faster than stateless (Phase 15)
- O(n) vs O(n²) complexity
- Incremental indicator updates

---

## Execution Layer

**Location:** `internal/broker/`, `internal/execution/`, `internal/order/`

### Broker Interface

**Added:** Phase 16 Week 1

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

**Implementations:**
- `SimulatedBroker` - backtest execution (Phase 16)
- `LegacyBroker` - backward compatibility wrapper (Phase 16)
- `PaperBroker` - paper trading (Phase 16 Week 2)
- `LiveBroker` - exchange execution (Phase 16 Week 3-4)

### SimulatedBroker

```go
type SimulatedBroker struct {
    orderManager  *order.OrderManager
    executor      *execution.SimpleExecutor
    pendingOrders map[string]*order.Order
    portfolio     *portfolio.Portfolio
    lastPrices    map[string]float64
}
```

**Features:**
- Context-aware methods
- Validates orders before submission
- Simulates fills via `SimpleExecutor`
- Tracks last prices per symbol

### LegacyBroker

**Added:** Phase 16 Week 1 Day 3

Wrapper that maintains old API for backward compatibility:

```go
type LegacyBroker struct {
    broker   *SimulatedBroker
    orderMgr *order.OrderManager
}

// Old API signature
func (b *LegacyBroker) SubmitOrder(req OrderRequest, timestamp time.Time) (*Order, error)
```

**Why:** Avoids refactoring 746-line engine while using new SimulatedBroker internally.

### Order Model

```go
type Order struct {
    ID          string
    Symbol      market.Symbol
    Side        OrderSide    // buy/sell
    Type        OrderType    // market/limit/stop/stop_limit
    Quantity    float64
    Price       *float64
    StopPrice   *float64
    Status      OrderStatus
    CreatedAt   time.Time
    FilledQty   float64
    FilledPrice float64
    Fees        float64
}
```

### Execution Simulation

**Location:** `internal/execution/`

```go
type SimpleExecutor struct {
    config Config // slippage, fees, spread
}

func (e *SimpleExecutor) SimulateFill(req OrderRequest, candle *Candle) (*Fill, error)
```

**Simulates:**
- Market orders - fill at close +/- slippage
- Limit orders - fill if price reached
- Stop orders - trigger when stop price hit
- Fees (maker/taker)
- Slippage (percentage or fixed)

**Intrabar Policy:**
- Conservative - stop hit first if ambiguous
- Optimistic - limit filled first if ambiguous

---

## Portfolio Layer

**Location:** `internal/portfolio/`

### Portfolio

```go
type Portfolio struct {
    InitialCash  float64
    Cash         float64
    Equity       float64
    Balance      float64
    Positions    map[market.Symbol]*Position
    ClosedTrades []Trade
    TotalFees    float64
    Timestamp    time.Time
}
```

### Position

```go
type Position struct {
    Symbol       market.Symbol
    Side         PositionSide  // long/short
    Quantity     float64
    EntryPrice   float64
    EntryTime    time.Time
    StopLoss     *float64
    TakeProfit   *float64
    CurrentPrice float64
}
```

**Methods:**
- `UnrealizedPnL()` - current P&L
- `UnrealizedPnLPercent()` - percentage return

### Trade

```go
type Trade struct {
    ID         string
    Symbol     market.Symbol
    Side       PositionSide
    EntryTime  time.Time
    EntryPrice float64
    ExitTime   time.Time
    ExitPrice  float64
    Quantity   float64
    GrossPnL   float64
    Fees       float64
    NetPnL     float64
    Return     float64
    MAE        float64  // Maximum Adverse Excursion
    MFE        float64  // Maximum Favorable Excursion
    ExitReason string
}
```

### Balance (NEW - Phase 16)

```go
type Balance struct {
    Cash   float64
    Equity float64
}
```

---

## Analytics Layer

**Location:** `internal/analytics/`

### Metrics

Core metrics calculated:
- Total Return
- CAGR (Compound Annual Growth Rate)
- Sharpe Ratio
- Sortino Ratio
- Maximum Drawdown
- Calmar Ratio
- Win Rate
- Profit Factor
- Expectancy
- Average Trade / Win / Loss

### Equity Curve

```go
type EquityPoint struct {
    Timestamp time.Time
    Equity    float64
    Cash      float64
    Drawdown  float64
}
```

### Trade Journal

All closed trades exported with full details for research analysis.

---

## Indicator Layer

**Location:** `internal/indicator/`

### Indicator Interface

```go
type Indicator interface {
    Name() string
    Calculate(candles []market.Candle) []Value
}
```

### CachedIndicator Interface (Phase 15)

```go
type CachedIndicator interface {
    Indicator
    Update(candle market.Candle, state *StateManager) (Value, error)
}
```

**Performance:** Incremental updates enable 37.6x speedup.

### Built-in Indicators

- SMA (Simple Moving Average)
- EMA (Exponential Moving Average)
- RSI (Relative Strength Index)
- ATR (Average True Range)
- MACD (Moving Average Convergence Divergence)
- Bollinger Bands
- ADX (Average Directional Index)

### Registry

```go
type Registry struct {
    indicators map[string]IndicatorFactory
}

func (r *Registry) Register(name string, factory IndicatorFactory)
func (r *Registry) Create(name string, params map[string]interface{}) (Indicator, error)
```

**Extensibility:** Custom indicators registered without modifying core engine.

---

## Risk Layer

**Location:** `internal/risk/`

### Position Sizing

```go
type PositionSizer struct {
    config PositionSizeConfig
}

func (ps *PositionSizer) Calculate(
    price float64,
    portfolio *portfolio.Portfolio,
    stopDistance float64,
) float64
```

**Methods:**
- `fixed` - fixed quantity
- `percent_equity` - percentage of equity
- `risk_percent` - risk-based (accounts for stop distance)

### Risk Management

```go
type Manager struct {
    config Config
}

func (m *Manager) ValidateTrade(
    portfolio *portfolio.Portfolio,
    exposure float64,
) error
```

**Checks:**
- Maximum trades per day
- Maximum exposure percentage
- Maximum daily loss
- Maximum drawdown

---

## Future Architecture

### Week 2: Paper Trading (In Progress)

```
Runtime
   ├── Backtest (done)
   ├── Paper (Week 2)
   └── Live (Week 3-4)

Broker
   ├── SimulatedBroker (done)
   ├── PaperBroker (Week 2)
   └── LiveBroker (Week 3-4)

DataFeed
   ├── CSVDataFeed (done)
   ├── WebSocketFeed (Week 3)
   └── RESTFeed (future)
```

### Week 3-4: Live Trading

```
LiveBroker
   ├── Exchange API client (Binance/Bybit)
   ├── Order state tracking
   ├── Position reconciliation
   └── Rate limiting

WebSocketFeed
   ├── Connection management
   ├── Reconnection logic
   ├── Candle building (tick → OHLCV)
   └── Data validation
```

### Post-MVP Features

- Multi-timeframe analysis (Phase 16)
- Parameter optimization (Phase 18)
- Walk Forward Analysis (Phase 19)
- Monte Carlo simulation (Phase 20)
- Advanced analytics (Phase 19)
- Web dashboard (future)

---

## Package Structure

```
.
├── cmd/
│   └── trader/              # CLI entry point
│
├── internal/
│   ├── runtime/             # NEW - Execution modes
│   │   ├── mode.go         # ExecutionMode, Config
│   │   ├── time.go         # TimeProvider interface
│   │   └── errors.go
│   │
│   ├── data/                # Data layer
│   │   ├── feed.go         # DataFeed interface (NEW)
│   │   ├── csv/            # CSV implementation
│   │   └── parquet/        # Parquet implementation
│   │
│   ├── strategy/            # Strategy layer
│   │   ├── ast/            # Abstract Syntax Tree
│   │   ├── parser/         # YAML → AST
│   │   ├── compiler/       # AST validation
│   │   └── evaluator/      # Signal generation
│   │
│   ├── indicator/           # Indicator layer
│   │   ├── builtin/        # Built-in indicators
│   │   ├── cached/         # Cached implementations (Phase 15)
│   │   └── registry/       # Indicator registry
│   │
│   ├── broker/              # Execution layer
│   │   ├── broker.go       # Broker interface (NEW)
│   │   ├── simulated.go    # SimulatedBroker (NEW)
│   │   └── legacy.go       # LegacyBroker wrapper (NEW)
│   │
│   ├── execution/           # Order execution simulation
│   ├── order/               # Order management
│   ├── portfolio/           # Portfolio tracking
│   ├── risk/                # Risk management
│   ├── analytics/           # Performance metrics
│   ├── backtest/            # Backtest engine
│   │
│   ├── optimization/        # Parameter optimization (Phase 18)
│   ├── walkforward/         # Walk Forward Analysis (Phase 19)
│   └── montecarlo/          # Monte Carlo simulation (Phase 20)
│
├── strategies/              # Example strategies
│   └── examples/
│
└── tests/                   # Integration tests
```

---

## Key Design Decisions

### Decision 1: Interface-Based Broker (Phase 16)

**Context:** Need to support backtest, paper, and live trading.

**Options:**
1. Single concrete Broker with mode flags
2. Interface with multiple implementations

**Chosen:** Interface-based design

**Rationale:**
- Testable (mock implementations)
- Extensible (new exchanges)
- Clean separation of concerns
- Mode-agnostic strategy code

### Decision 2: LegacyBroker Wrapper (Phase 16 Day 3)

**Context:** Engine uses old Broker API, new interface has different signature.

**Options:**
1. Refactor 746-line engine
2. Create wrapper maintaining old API

**Chosen:** LegacyBroker wrapper (94 lines)

**Rationale:**
- Low risk (1-line engine change)
- Zero breaking changes
- Incremental migration possible
- Proven by tests (100% pass)

### Decision 3: Cached Indicators (Phase 15)

**Context:** Backtest was slow (50.4s for 5 years).

**Options:**
1. Optimize algorithm
2. Incremental state-based updates

**Chosen:** CachedIndicator with incremental updates

**Result:** 37.6x speedup (50.4s → 1.34s)

### Decision 4: Defer DataFeed Engine Integration (Phase 16 Day 3)

**Context:** Engine works with candle arrays, DataFeed uses Next() iterator.

**Options:**
1. Refactor engine to use DataFeed.Next()
2. Keep current approach, use DataFeed for real-time only

**Chosen:** Defer engine refactor

**Rationale:**
- Current approach works perfectly
- No user-facing benefit now
- Reduces risk
- Focus on Week 2 (Paper Trading)

---

## Testing Strategy

### Unit Tests
- Indicator calculations
- Expression evaluation
- Order logic
- Portfolio accounting
- Analytics formulas

### Integration Tests
- Strategy parsing → AST
- AST → Evaluator → Signals
- Signals → Orders → Fills
- Complete backtest flow

### Golden Tests (Phase 8)
- Known strategy + known data = known results
- Verified on every change
- Prevents regressions

### Performance Tests (Phase 15)
- Benchmark indicator calculations
- Profile CPU/memory
- Regression detection

---

## Performance Characteristics

| Operation | Time | Complexity |
|-----------|------|------------|
| Load 5yr hourly CSV | ~100ms | O(n) |
| Parse strategy YAML | <10ms | O(indicators) |
| Backtest 5yr (cached) | 1.34s | O(n) |
| Backtest 5yr (stateless) | 50.4s | O(n²) |
| Single indicator update | <1μs | O(1) |

**Target:** <2s for 5 years hourly data ✅ ACHIEVED (Phase 15)

---

## References

- AGENTS.md - Project philosophy and requirements
- POST_MVP_PLAN.md - Development roadmap
- PHASE15_COMPLETE.md - Performance optimization
- PHASE16_WEEK1_PROGRESS.md - Live trading architecture
- README.md - Getting started guide

---

**Last Updated:** 2026-09-05 (Phase 16 Week 1)  
**Architecture Status:** Stable and extensible  
**Next Evolution:** Paper Trading (Week 2)
