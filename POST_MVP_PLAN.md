# Post-MVP Development Plan
**Date:** 2026-09-05  
**Status:** Planning Phase  
**Target Version:** 1.0.0

---

## Executive Summary

MVP (v0.1.0) telah selesai dengan 15 fase development lengkap. Document ini merencanakan prioritas development untuk mencapai production-ready v1.0.0 dan fitur-fitur enhancement selanjutnya.

---

## Priority Classification

### 🔴 Critical Path (Required for v1.0.0)
Features yang **harus** ada sebelum production deployment.

### 🟡 High Priority (Target v1.1.0 - v1.3.0)
Features yang sangat meningkatkan usability dan robustness.

### 🟢 Enhancement (Target v1.4.0+)
Features untuk advanced use cases dan ecosystem growth.

---

## Phase 16: Live Trading Architecture 🔴
**Priority:** Critical  
**Target:** v1.0.0  
**Estimated Duration:** 3-4 weeks

### Objectives
- Strategy code berjalan identik di backtest, paper, dan live
- Paper trading dengan real-time data simulation
- Architecture siap untuk exchange integration

### Technical Design

#### 1. Broker Interface Abstraction
```go
type Broker interface {
    // Order lifecycle
    SubmitOrder(ctx context.Context, order *Order) (OrderID, error)
    CancelOrder(ctx context.Context, orderID OrderID) error
    GetOrderStatus(ctx context.Context, orderID OrderID) (*OrderStatus, error)
    
    // Position & balance
    GetPositions(ctx context.Context) ([]Position, error)
    GetBalance(ctx context.Context) (*Balance, error)
    
    // Market data
    GetLastPrice(ctx context.Context, symbol string) (float64, error)
}

// Implementations:
// - SimulatedBroker (existing)
// - PaperBroker (new - uses real-time data)
// - LiveBroker (new - connects to exchange)
```

#### 2. Execution Modes
```yaml
# backtest.yaml (existing)
mode: backtest
data:
  source: csv
  path: data/BTCUSDT.csv

# paper.yaml (new)
mode: paper
data:
  source: websocket
  exchange: binance
  symbol: BTCUSDT
  
account:
  initial_balance: 10000
  simulate_latency: true
  
# live.yaml (new)
mode: live
exchange:
  name: binance
  api_key_env: BINANCE_API_KEY
  api_secret_env: BINANCE_API_SECRET
  
risk_limits:
  max_position_size: 1000
  max_daily_loss: 500
  require_confirmation: true
```

#### 3. Data Feed Abstraction
```go
type DataFeed interface {
    Subscribe(ctx context.Context, symbols []string) error
    NextCandle() (*Candle, error)
    Close() error
}

// Implementations:
// - CSVFeed (existing)
// - WebSocketFeed (new)
// - RestAPIFeed (new)
```

### Implementation Tasks

- [ ] **Week 1: Core Abstractions**
  - [ ] Define `Broker` interface
  - [ ] Define `DataFeed` interface  
  - [ ] Refactor `SimulatedBroker` to implement interface
  - [ ] Add execution mode configuration

- [ ] **Week 2: Paper Trading**
  - [ ] Implement `PaperBroker`
  - [ ] Add latency simulation
  - [ ] Implement order queue
  - [ ] Add position reconciliation

- [ ] **Week 3: Real-Time Data**
  - [ ] Implement `WebSocketFeed` interface
  - [ ] Add connection resilience
  - [ ] Implement data buffering
  - [ ] Add heartbeat/reconnection logic

- [ ] **Week 4: Integration & Testing**
  - [ ] CLI support for paper trading mode
  - [ ] Integration tests for all modes
  - [ ] Documentation update
  - [ ] Example paper trading strategies

### Success Criteria
✅ Same strategy YAML works for backtest, paper, live  
✅ Paper trading simulates realistic latency (50-200ms)  
✅ Position reconciliation handles edge cases  
✅ Tests verify mode-agnostic behavior

---

## Phase 17: Enhanced Data Handling 🔴
**Priority:** Critical  
**Target:** v1.0.0  
**Estimated Duration:** 2-3 weeks

### Objectives
- Support Parquet untuk efficient storage
- Handle market gaps dan missing data
- Multi-timeframe data synchronization

### Technical Design

#### 1. Parquet Support
```go
// internal/data/parquet/reader.go
type ParquetReader struct {
    path string
    file *parquet.File
}

func (r *ParquetReader) Read() ([]*market.Candle, error) {
    // Read Parquet columns: timestamp, open, high, low, close, volume
}
```

#### 2. Data Validation Pipeline
```go
type DataValidator interface {
    Validate(candles []*Candle) (*ValidationReport, error)
}

// Validations:
// - Chronological ordering
// - OHLC consistency (high >= low, etc.)
// - No duplicate timestamps
// - Volume >= 0
// - Price > 0
// - Detect gaps (configurable threshold)
```

#### 3. Gap Handling
```yaml
data:
  source: csv
  path: data/BTCUSDT.csv
  
  validation:
    allow_gaps: true
    max_gap_minutes: 60
    
  handling:
    on_gap: forward_fill  # options: forward_fill, interpolate, skip, error
```

### Implementation Tasks

- [ ] **Week 1: Parquet Integration**
  - [ ] Add parquet library dependency
  - [ ] Implement `ParquetReader`
  - [ ] Add tests with sample parquet files
  - [ ] CLI support for `--data=file.parquet`

- [ ] **Week 2: Data Validation**
  - [ ] Implement validation pipeline
  - [ ] Add gap detection
  - [ ] Create validation report format
  - [ ] CLI command: `trader validate-data`

- [ ] **Week 3: Gap Handling**
  - [ ] Implement forward fill strategy
  - [ ] Add interpolation option
  - [ ] Test with real gapped data
  - [ ] Documentation update

### Success Criteria
✅ Load 5+ years hourly data in <1 second  
✅ Detect and report data quality issues  
✅ Handle market gaps without silent errors  
✅ Support CSV and Parquet formats

---

## Phase 15: Performance & Optimization 🔴
**Priority:** Critical  
**Target:** v1.0.0  
**Estimated Duration:** 2-3 weeks

### Objectives
- Backtest 5+ years hourly data in <2 seconds
- Optimization dengan 100 parameter combinations in <10 seconds
- Memory usage <500MB for typical backtests

### Technical Design

#### 1. Indicator Caching
```go
type CachedIndicator struct {
    base Indicator
    cache map[int]Value  // bar index -> value
    dependencies []string
}

// Cache invalidation on dependency change
```

#### 2. Concurrent Optimization
```go
type WorkerPool struct {
    workers int
    jobs    chan OptimizationJob
    results chan OptimizationResult
}

// Distribute parameter combinations across workers
```

#### 3. Profiling Tools
```bash
trader backtest --strategy=strategy.yaml --data=data.csv --profile=cpu
trader backtest --strategy=strategy.yaml --data=data.csv --profile=mem
```

### Implementation Tasks

- [ ] **Week 1: Profiling Setup**
  - [ ] Add pprof integration
  - [ ] CLI flags for profiling
  - [ ] Benchmark suite for core operations
  - [ ] Identify bottlenecks

- [ ] **Week 2: Caching Implementation**
  - [ ] Implement indicator value caching
  - [ ] Add cache invalidation logic
  - [ ] Verify correctness with tests
  - [ ] Measure performance improvement

- [ ] **Week 3: Concurrent Optimization**
  - [ ] Implement worker pool
  - [ ] Distribute optimization jobs
  - [ ] Aggregate results correctly
  - [ ] Test with large parameter spaces

### Success Criteria
✅ 60%+ speed improvement on typical backtests  
✅ Optimization scales linearly with CPU cores  
✅ Memory usage reduced by 30%+  
✅ No correctness regressions

---

## Phase 19: Advanced Analytics & Reporting 🟡
**Priority:** High  
**Target:** v1.1.0  
**Estimated Duration:** 3 weeks

### Objectives
- HTML reports dengan interactive charts
- Performance attribution analysis
- Publication-quality tearsheets

### Technical Design

#### 1. HTML Report Generation
```go
type ReportGenerator struct {
    template *template.Template
    chartLib string  // "plotly", "chart.js"
}

// Generate interactive HTML with:
// - Equity curve
// - Drawdown chart
// - Monthly returns heatmap
// - Trade distribution
```

#### 2. Performance Attribution
```go
type Attribution struct {
    MarketReturn    float64  // Buy & hold benchmark
    StrategyAlpha   float64  // Strategy excess return
    ExecutionCost   float64  // Fees + slippage impact
    TimingSkill     float64  // Entry/exit timing contribution
}
```

#### 3. Report Templates
```yaml
report:
  format: html  # html, pdf, markdown
  template: tearsheet  # tearsheet, detailed, summary
  
  sections:
    - performance_summary
    - equity_curve
    - drawdown_analysis
    - monthly_returns
    - trade_distribution
    - risk_metrics
    - attribution
```

### Implementation Tasks

- [ ] **Week 1: Template Engine**
  - [ ] Choose HTML template library
  - [ ] Create base tearsheet template
  - [ ] Add chart.js/plotly integration
  - [ ] CLI: `trader report --result=result.json --output=report.html`

- [ ] **Week 2: Advanced Charts**
  - [ ] Equity curve with drawdown overlay
  - [ ] Monthly returns heatmap
  - [ ] Trade distribution histogram
  - [ ] Rolling metrics charts

- [ ] **Week 3: Attribution Analysis**
  - [ ] Implement attribution calculation
  - [ ] Compare vs buy-and-hold benchmark
  - [ ] Calculate execution cost impact
  - [ ] Add to report template

### Success Criteria
✅ HTML report looks professional  
✅ Charts are interactive (zoom, hover)  
✅ Attribution breakdown is accurate  
✅ Report generation takes <5 seconds

---

## Phase 18: Advanced Strategy DSL 🟡
**Priority:** High  
**Target:** v1.2.0  
**Estimated Duration:** 4 weeks

### Objectives
- String-based expression parsing: `close > sma(20) and volume > avg(volume, 20)`
- User-defined functions in YAML
- Enhanced state machine support

### Technical Design

#### 1. Expression Parser
```go
// Parse string expressions into AST
parser := NewExpressionParser()
ast, err := parser.Parse("close > sma(close, 20) and rsi(14) < 30")

// Existing YAML format still supported for backward compatibility
```

#### 2. User-Defined Functions
```yaml
functions:
  volatility:
    formula: "stddev(close, 20) / sma(close, 20)"
    
  volume_surge:
    formula: "volume / sma(volume, 20)"
    
  trend_strength:
    formula: "abs(ema(close, 9) - ema(close, 21)) / atr(14)"

entry:
  long:
    condition: "trend_strength > 2.0 and volume_surge > 1.5"
```

#### 3. Enhanced State Machines
```yaml
state:
  setup_stage:
    type: enum
    values: [none, bullish_structure, confirmed, invalidated]
    default: none
    
  bars_in_setup:
    type: int
    default: 0

rules:
  - name: detect_bullish_structure
    when: "close > ema(close, 200) and rsi(14) > 50"
    then:
      set:
        setup_stage: bullish_structure
        bars_in_setup: 0
        
  - name: confirm_setup
    when: "setup_stage == 'bullish_structure' and cross_above(ema(close, 9), ema(close, 21))"
    then:
      set:
        setup_stage: confirmed
        
  - name: invalidate_setup
    when: "bars_in_setup > 10"
    then:
      set:
        setup_stage: invalidated
        bars_in_setup: 0
        
  - name: enter_on_confirmation
    when: "setup_stage == 'confirmed' and volume > sma(volume, 20) * 1.2"
    then:
      action: enter_long
      set:
        setup_stage: none

  - name: increment_bars
    when: "setup_stage != 'none'"
    then:
      set:
        bars_in_setup: "bars_in_setup + 1"
```

### Implementation Tasks

- [ ] **Week 1: Expression Lexer/Parser**
  - [ ] Define grammar for expression language
  - [ ] Implement tokenizer
  - [ ] Implement parser (recursive descent or parser combinator)
  - [ ] Generate AST from string expressions

- [ ] **Week 2: Function System**
  - [ ] Parser support for function definitions
  - [ ] Function registry
  - [ ] Type checking for user functions
  - [ ] Recursive function resolution

- [ ] **Week 3: State Machine Enhancement**
  - [ ] Enum state types
  - [ ] State transition validation
  - [ ] State-based conditions
  - [ ] State persistence across bars

- [ ] **Week 4: Integration & Testing**
  - [ ] Backward compatibility tests
  - [ ] Complex strategy examples
  - [ ] Documentation update
  - [ ] Migration guide from old to new syntax

### Success Criteria
✅ String expressions parse correctly  
✅ User-defined functions work in expressions  
✅ State machines handle complex logic  
✅ Backward compatible with existing YAML strategies

---

## Phase 22: Exchange Integration 🟡
**Priority:** High  
**Target:** v1.3.0  
**Estimated Duration:** 4-5 weeks

### Objectives
- Integration dengan major exchanges (Binance, Coinbase, etc.)
- Real order submission dengan proper error handling
- Exchange-specific quirks handled correctly

### Technical Design

#### 1. Exchange Abstraction
```go
type Exchange interface {
    // Authentication
    Connect(apiKey, apiSecret string) error
    
    // Market data
    GetCandles(symbol string, timeframe string, limit int) ([]*Candle, error)
    StreamCandles(symbol string, timeframe string) (<-chan *Candle, error)
    
    // Trading
    PlaceOrder(order *Order) (*OrderResult, error)
    CancelOrder(orderID string) error
    GetOrder(orderID string) (*OrderStatus, error)
    
    // Account
    GetBalance() (*Balance, error)
    GetPositions() ([]*Position, error)
    
    // Exchange info
    GetSymbolInfo(symbol string) (*SymbolInfo, error)
    GetTradingFees() (*FeeStructure, error)
}
```

#### 2. CCXT Integration
```go
// internal/exchange/ccxt/binance.go
type BinanceExchange struct {
    client *ccxt.Binance
    rateLimiter *RateLimiter
}

// Handle Binance-specific:
// - Price precision
// - Quantity precision
// - Min notional value
// - Order rate limits
```

#### 3. Live Trading Configuration
```yaml
mode: live

exchange:
  name: binance
  credentials:
    api_key_env: BINANCE_API_KEY
    api_secret_env: BINANCE_API_SECRET
  
  testnet: false  # Use testnet for safe testing

risk_limits:
  max_position_size_usd: 5000
  max_daily_trades: 20
  max_daily_loss_usd: 500
  
  confirmation:
    required: true  # Require manual confirmation for each order
    method: cli  # cli, webhook, api

strategy:
  file: strategies/live/ema_cross.yaml
  
monitoring:
  log_all_orders: true
  notify_on_fill: true
  health_check_interval: 60s
```

### Implementation Tasks

- [ ] **Week 1: CCXT Integration**
  - [ ] Add CCXT library
  - [ ] Implement Binance exchange adapter
  - [ ] Handle authentication
  - [ ] Basic order placement tests (testnet)

- [ ] **Week 2: Exchange-Specific Logic**
  - [ ] Price/quantity precision handling
  - [ ] Min notional value validation
  - [ ] Rate limiting implementation
  - [ ] Error handling for exchange responses

- [ ] **Week 3: Live Trading Safety**
  - [ ] Order confirmation system
  - [ ] Risk limit enforcement
  - [ ] Position reconciliation
  - [ ] Emergency stop mechanism

- [ ] **Week 4: Additional Exchanges**
  - [ ] Coinbase adapter
  - [ ] Bybit adapter (if needed)
  - [ ] Exchange-agnostic tests
  - [ ] Fee structure handling

- [ ] **Week 5: Monitoring & Ops**
  - [ ] Order log persistence
  - [ ] Health check system
  - [ ] Notification system (email/webhook)
  - [ ] Dashboard for live monitoring

### Success Criteria
✅ Orders submit successfully to exchange testnet  
✅ Position reconciliation works correctly  
✅ Risk limits enforced before order submission  
✅ Emergency stop halts all trading immediately  
✅ All exchange-specific quirks handled

---

## Phase 20: Portfolio Analysis 🟢
**Priority:** Enhancement  
**Target:** v1.4.0  
**Estimated Duration:** 3 weeks

### Objectives
- Run multiple strategies simultaneously
- Portfolio-level risk management
- Correlation analysis between strategies

### Technical Design

#### 1. Portfolio Configuration
```yaml
portfolio:
  name: diversified_crypto
  initial_capital: 100000
  
  strategies:
    - name: btc_trend
      file: strategies/btc_trend.yaml
      allocation: 0.4  # 40% of capital
      
    - name: eth_mean_reversion
      file: strategies/eth_mean_reversion.yaml
      allocation: 0.3
      
    - name: sol_breakout
      file: strategies/sol_breakout.yaml
      allocation: 0.3
      
  risk:
    max_correlation: 0.7  # Limit strategy correlation
    rebalance_frequency: weekly
    max_portfolio_drawdown: 0.2
```

#### 2. Portfolio Metrics
```go
type PortfolioMetrics struct {
    TotalReturn        float64
    Sharpe             float64
    DiversificationRatio float64
    StrategyCorrelations map[string]map[string]float64
    
    ByStrategy map[string]*StrategyMetrics
}
```

### Implementation Tasks

- [ ] **Week 1: Portfolio Engine**
  - [ ] Multi-strategy execution
  - [ ] Capital allocation logic
  - [ ] Aggregate portfolio accounting

- [ ] **Week 2: Correlation Analysis**
  - [ ] Calculate strategy return correlations
  - [ ] Diversification ratio
  - [ ] Risk contribution by strategy

- [ ] **Week 3: Rebalancing**
  - [ ] Periodic rebalancing logic
  - [ ] Target allocation enforcement
  - [ ] Rebalancing cost tracking

### Success Criteria
✅ Multiple strategies run simultaneously  
✅ Portfolio metrics calculated correctly  
✅ Correlation warnings when strategies too similar

---

## Phase 21: Machine Learning Integration 🟢
**Priority:** Enhancement  
**Target:** v1.5.0  
**Estimated Duration:** 4 weeks

### Objectives
- Use ML model predictions as indicators
- Feature engineering pipeline
- Look-ahead safe model integration

### Technical Design

#### 1. ML Model Interface
```go
type MLModel interface {
    Predict(features map[string]float64) (float64, error)
    Name() string
}

// Implementations:
// - SKLearnModel (load from pickle)
// - ONNXModel (load from .onnx)
// - CustomModel (user-provided)
```

#### 2. Strategy with ML
```yaml
ml_models:
  trend_predictor:
    type: onnx
    path: models/trend_predictor.onnx
    
    features:
      - name: rsi
        indicator: rsi
        period: 14
        
      - name: ema_distance
        formula: "(ema(close, 9) - ema(close, 21)) / close"
        
      - name: volume_ratio
        formula: "volume / sma(volume, 20)"
        
    output:
      name: trend_score
      threshold: 0.6

entry:
  long:
    all:
      - gt: [trend_score, 0.6]
      - cross_above: [ema9, ema21]
```

### Implementation Tasks

- [ ] **Week 1: Model Interface**
  - [ ] Define ML model interface
  - [ ] ONNX runtime integration
  - [ ] Feature vector construction

- [ ] **Week 2: Feature Engineering**
  - [ ] Feature definition in YAML
  - [ ] Feature calculation pipeline
  - [ ] Look-ahead bias prevention

- [ ] **Week 3: Model Integration**
  - [ ] Model predictions as indicators
  - [ ] Model warming period handling
  - [ ] Error handling for model failures

- [ ] **Week 4: Examples & Docs**
  - [ ] Example ML strategy
  - [ ] Model training guide
  - [ ] Feature engineering best practices

### Success Criteria
✅ ML models load and predict correctly  
✅ Features calculated without look-ahead  
✅ Model predictions usable in strategy logic

---

## Development Workflow

### 1. Branch Strategy
```
main (production)
  ↑
develop (integration)
  ↑
feature/phase-16-live-trading
feature/phase-17-data-handling
feature/phase-15-performance
```

### 2. Release Cadence
- **v1.0.0** (Critical path complete): Target 8-10 weeks
- **v1.1.0** (Reporting): Target +3 weeks
- **v1.2.0** (Advanced DSL): Target +4 weeks
- **v1.3.0** (Exchange integration): Target +5 weeks
- **v1.4.0+** (Enhancements): Iterative releases

### 3. Quality Gates
Before merging any phase:
- [ ] All tests pass
- [ ] Code coverage >85%
- [ ] No race conditions (`go test -race`)
- [ ] Performance benchmarks meet targets
- [ ] Documentation updated
- [ ] CHANGELOG.md updated

---

## Resource Allocation

### Critical Path (Phases 15, 16, 17)
**Estimated Total:** 7-10 weeks  
**Blocking:** Yes - required for v1.0.0

### High Priority (Phases 18, 19, 22)
**Estimated Total:** 11-12 weeks  
**Blocking:** No - can be done post-1.0.0

### Enhancement (Phases 20, 21)
**Estimated Total:** 7 weeks  
**Blocking:** No - advanced features

---

## Risk Management

### Technical Risks
1. **Performance targets missed**
   - Mitigation: Profile early, iterate on hotspots
   
2. **Exchange API breaking changes**
   - Mitigation: Version pinning, API compatibility layer
   
3. **Look-ahead bias in ML integration**
   - Mitigation: Strict feature calculation review, tests

### Schedule Risks
1. **Critical path takes longer than estimated**
   - Mitigation: Prioritize ruthlessly, cut non-essential features
   
2. **Exchange integration complexity**
   - Mitigation: Start with one exchange (Binance), add others iteratively

---

## Success Metrics

### v1.0.0 Launch Criteria
- [ ] Phases 15, 16, 17 complete
- [ ] All existing tests pass
- [ ] Performance targets met
- [ ] Paper trading functional
- [ ] Documentation updated
- [ ] Example strategies tested
- [ ] Migration guide from v0.1.0

### Community Goals (6 months)
- 100+ GitHub stars
- 10+ community-contributed strategies
- 5+ external contributors
- Active discussions on issues/PRs

---

## Next Steps

### Immediate Actions (This Week)
1. Create GitHub issues for Phase 15, 16, 17 tasks
2. Set up project board for tracking
3. Begin Phase 16 design document
4. Profile current backtest performance

### This Month
1. Complete Phase 16 (Live Trading Architecture)
2. Start Phase 17 (Data Handling)
3. Parallel: Begin Phase 15 (Performance)

### This Quarter
1. Release v1.0.0 with critical path complete
2. Begin work on high-priority phases
3. Gather community feedback

---

**Document Version:** 1.0  
**Last Updated:** 2026-09-05  
**Next Review:** After v1.0.0 release
