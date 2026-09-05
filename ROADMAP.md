# Post-MVP Roadmap

## Project Status
✅ **MVP Complete**: All 15 development phases finished, all acceptance criteria met.

---

## Phase 15: Performance & Optimization

### Goals
- Profile backtest engine for bottlenecks
- Implement caching strategies without breaking semantics
- Support concurrent independent backtests
- Reduce memory footprint for large datasets

### Key Tasks
- [ ] Add CPU/memory profiling to CLI
- [ ] Implement indicator value caching with invalidation
- [ ] Add expression result caching
- [ ] Implement worker pool for parameter optimization
- [ ] Benchmarking suite for core operations

### Success Criteria
- Backtest 5+ years of hourly data in <2 seconds
- Optimization on 100 parameter combinations in <10 seconds
- Memory usage <500MB for typical backtests

---

## Phase 16: Live Trading Architecture

### Goals
- Design abstraction layer that supports live execution
- Create paper trading interface
- Implement basic WebSocket data feed support

### Key Tasks
- [ ] Define LiveBroker interface separate from SimulatedBroker
- [ ] Implement paper trading mode
- [ ] Add real-time data feed abstraction
- [ ] Create order submission queue
- [ ] Add position reconciliation logic

### Success Criteria
- Strategy code runs identically in backtest, paper, and live
- Paper trading produces realistic execution simulation
- Architecture ready for exchange integration

---

## Phase 17: Enhanced Data Handling

### Goals
- Support multiple data sources simultaneously
- Add Parquet support for efficient storage
- Implement data validation and cleaning pipeline

### Key Tasks
- [ ] Implement Parquet reader/writer
- [ ] Add data normalization (handling gaps, holidays)
- [ ] Create multi-source synchronization
- [ ] Add data quality metrics dashboard
- [ ] Support intrabar tick data

### Success Criteria
- Load 10+ years of data in <1 second
- Auto-detect and handle market gaps
- Support tick, minute, hourly, daily data simultaneously

---

## Phase 18: Advanced Strategy DSL

### Goals
- Add string-based expression parsing
- Implement strategy state machines
- Support advanced temporal logic
- Add custom function definitions

### Key Tasks
- [ ] Implement expression parser for: `close > sma(20) and volume > avg(volume, 20)`
- [ ] State machine DSL with transitions and rules
- [ ] User-defined functions in YAML
- [ ] Indicator composition operators
- [ ] Macro expansion system

### Success Criteria
- Complex nested conditions work correctly
- State machines handle edge cases properly
- User-defined functions integrate with type system

---

## Phase 19: Advanced Analytics & Reporting

### Goals
- Generate publication-quality reports
- Add advanced statistical analysis
- Implement performance attribution
- Create interactive visualizations

### Key Tasks
- [ ] HTML report generation with charts
- [ ] PDF export capability
- [ ] Performance attribution (market, strategy, execution)
- [ ] Risk factor analysis
- [ ] Strategy tearsheet templates

### Success Criteria
- HTML reports display equity curve, drawdown, monthly returns
- PDF export maintains all formatting
- Attribution shows source of returns

---

## Phase 20: Portfolio Analysis

### Goals
- Support multi-strategy portfolios
- Implement correlation analysis
- Add portfolio risk metrics

### Key Tasks
- [ ] Portfolio aggregation layer
- [ ] Cross-strategy correlation tracking
- [ ] Portfolio-level risk limits
- [ ] Rebalancing logic
- [ ] Tax-aware position management

### Success Criteria
- Multiple strategies can be combined with proper accounting
- Portfolio metrics (diversification, correlation) computed correctly
- Rebalancing rules enforced

---

## Phase 21: Machine Learning Integration

### Goals
- Create interface for ML models in strategies
- Support feature engineering pipelines
- Add model training utilities

### Key Tasks
- [ ] ML model interface for predictions
- [ ] Feature engineering DSL
- [ ] Backtesting with model predictions
- [ ] Model drift detection
- [ ] Retraining pipeline support

### Success Criteria
- ML models can be used as indicators
- Feature engineering stays look-ahead safe
- Model performance tracked separately from strategy

---

## Phase 22: Exchange Integration

### Goals
- Support real exchange APIs
- Add paper trading with real data
- Implement risk limits at exchange level

### Key Tasks
- [ ] CCXT integration for multi-exchange support
- [ ] Real order submission framework
- [ ] Exchange-specific fee models
- [ ] Rate limiting and API resilience
- [ ] Position reconciliation with exchange

### Success Criteria
- Paper trading uses real exchange data
- Live trading can submit orders to exchange
- Handles exchange-specific quirks (precision, limits, fees)

---

## Phase 23: Robustness & Stress Testing

### Goals
- Add scenario analysis
- Implement stress testing framework
- Support adversarial analysis

### Key Tasks
- [ ] Scenario DSL (market crashes, gaps, etc.)
- [ ] Worst-case drawdown analysis
- [ ] Regime change testing
- [ ] Black swan simulations
- [ ] Sensitivity analysis extension

### Success Criteria
- Stress tests reveal strategy vulnerabilities
- Scenario analysis produces actionable insights
- Strategy robustness quantified and reported

---

## Phase 24: Cloud & Distributed Computing

### Goals
- Support distributed optimization
- Add cloud data storage integration
- Implement parallel backtesting

### Key Tasks
- [ ] Distributed optimization framework
- [ ] AWS/GCP/Azure integration
- [ ] Kubernetes deployment support
- [ ] Distributed data storage (S3, GCS)
- [ ] Result aggregation and reporting

### Success Criteria
- Optimization scales to 1000+ parameter combinations
- Multi-machine optimization coordinated properly
- Results reproducible across runs

---

## Phase 25: Documentation & Community

### Goals
- Create comprehensive user guide
- Build example strategy library
- Establish contribution guidelines

### Key Tasks
- [ ] User guide (100+ pages)
- [ ] Video tutorials
- [ ] Example strategy collection (50+ strategies)
- [ ] API documentation generation
- [ ] Contributing guidelines and templates

### Success Criteria
- New users can create strategies in <1 hour
- 50+ example strategies covering all features
- Community contributions accepted and integrated

---

## Known Limitations (Post-MVP)

### Current Constraints
1. **Single-threaded backtest**: Event loop is inherently sequential
2. **OHLC-only candle data**: Intrabar ambiguity unresolved
3. **No leverage modeling**: Futures trading requires extensions
4. **Limited data sources**: CSV and basic feeds only
5. **No ML integration**: Models can't influence strategy decisions
6. **Single machine only**: No distributed optimization yet
7. **No exchange integration**: Paper trading only
8. **Static risk limits**: No dynamic risk adjustment mid-backtest

### Design Debt to Address
1. Expression AST could benefit from JIT compilation
2. Portfolio accounting could use event sourcing
3. Risk engine could be more modular
4. Indicator caching needs invalidation strategy
5. CLI could support interactive REPL mode

---

## Success Metrics for Post-MVP

| Metric | Target | Current |
|--------|--------|---------|
| Backtest speed (5Y hourly) | <2s | ~5s |
| Memory usage | <500MB | ~800MB |
| Optimization concurrency | 16+ workers | 1 worker |
| Data sources supported | 5+ | 1 (CSV) |
| Strategy examples | 50+ | 11 |
| Exchange integrations | 3+ | 0 |
| Test coverage | >90% | ~85% |
| Documentation pages | 50+ | 12 |

---

## Maintenance Tasks (Ongoing)

- [ ] Dependency updates (quarterly)
- [ ] Security audits (quarterly)
- [ ] Performance regression tests
- [ ] Community issue triage (weekly)
- [ ] Documentation updates with changes
- [ ] Example strategy updates for new features

---

## Architecture Principles (Post-MVP)

All future work must maintain:

1. **Declarative strategies** - No Go code for ordinary strategies
2. **Zero look-ahead bias** - Temporal semantics always explicit
3. **Deterministic execution** - Same inputs → same outputs
4. **Clean domain boundaries** - No leaky abstractions
5. **Extensibility first** - Register, don't hardcode
6. **Test coverage** - All meaningful behavior tested
7. **Backward compatibility** - Strategies remain valid across versions

---

## Critical Path to 1.0

1. **Phase 15** (Performance) - Required for production use
2. **Phase 16** (Live Trading) - Required for deployment
3. **Phase 17** (Data Handling) - Required for real-world trading
4. **Phase 19** (Reporting) - Required for stakeholder communication
5. **Phase 22** (Exchange Integration) - Required for live execution

Phases 18, 20, 21, 23, 24, 25 are enhancement layers.

---

Generated: 2026-09-05T06:33:10Z
