# Architecture Overview

## High-Level Design

The backtest engine follows clean architecture principles with clear separation of concerns:

```
                    ┌────────────────────┐
                    │ Strategy Definition│
                    │ YAML / JSON / API  │
                    └──────────┬─────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │       Parser       │
                    └──────────┬─────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │ Strategy AST / IR  │
                    └──────────┬─────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
              ▼                ▼                ▼
        Indicator Graph   Expression Tree   State Machine
              │                │                │
              └────────────────┼────────────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │ Strategy Evaluator │
                    └──────────┬─────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │    Signal Engine   │
                    └──────────┬─────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │     Risk Engine    │
                    └──────────┬─────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │    Order Engine    │
                    └──────────┬─────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │  Execution/Broker  │
                    └──────────┬─────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │     Portfolio      │
                    └──────────┬─────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │ Analytics/Reports  │
                    └────────────────────┘
```

## Core Packages

### `internal/strategy`
Strategy definition, parsing, and compilation.

- **`parser`**: YAML → AST conversion
- **`ast`**: Abstract Syntax Tree types
- **`compiler`**: AST → runtime representation
- **`evaluator`**: Strategy execution during backtest

### `internal/data`
Market data handling.

- **`csv`**: CSV data feed parsing
- **`feed`**: Data feed interface
- **`resample`**: Timeframe resampling

### `internal/indicator`
Technical indicators with registry pattern.

- **`registry`**: Indicator factory
- **Built-in indicators**: SMA, EMA, RSI, ATR
- **Composite indicators**: Via expression engine

### `internal/expression`
Expression evaluation (arithmetic, logic, conditions).

- **Expression AST**
- **Evaluator**
- **Built-in functions**: cross_above, cross_below, rising, falling, between

### `internal/signal`
Signal generation from conditions.

### `internal/order`
Order types and lifecycle.

- Market, Limit, Stop, Stop-Limit
- Order states: Pending, Accepted, Filled, Cancelled, Rejected

### `internal/execution`
Realistic order execution simulation.

- Fees (maker/taker)
- Slippage
- Order filling
- Intrabar ambiguity handling

### `internal/portfolio`
Portfolio state management.

- Positions
- Cash
- PnL (realized and unrealized)
- Equity curve
- Trade history

### `internal/risk`
Risk management rules.

- Position sizing
- Stop loss
- Take profit
- Trailing stops

### `internal/backtest`
Main backtest engine.

- Event loop
- Chronological iteration
- State management

### `internal/analytics`
Result analysis.

- Metrics: Sharpe, Sortino, drawdown, etc.
- Trade statistics
- Equity curve analysis

### `internal/optimization`
Parameter optimization.

- Grid search
- Objective functions

### `internal/walkforward`
Walk Forward Analysis.

- Train/test window management
- Out-of-sample analysis

### `internal/montecarlo`
Monte Carlo simulation.

- Trade reshuffling
- Statistical analysis

## Key Design Principles

### 1. YAML is an Interface, Not the Engine

```
YAML → Parser → AST → Compiler → Runtime Evaluator
```

This separation ensures:
- Future strategy representations (JSON, CLI, Web UI) can compile to the same AST
- Domain logic is independent of YAML structure
- Easy to validate and transform

### 2. Registry-Based Extensibility

Indicators, functions, and analyzers use the registry pattern:

```go
registry.Register("ema", &EMAImplementation{})
```

This allows adding new indicators without modifying the core engine.

### 3. Deterministic Backtesting

All randomized operations use explicit seeds. Same inputs produce same outputs.

### 4. No Look-Ahead Bias

The architecture makes look-ahead impossible:
- Data flows chronologically
- Future data is never accessed during evaluation
- Historical references use explicit `ref` / `shift` operators

### 5. Event-Driven Architecture

```
MarketEvent → Strategy Evaluation → SignalEvent → Risk Evaluation
    ↓            ↓                      ↓              ↓
OrderEvent → Execution → FillEvent → Portfolio Update
```

Each stage is decoupled and testable independently.

### 6. Clean Domain Boundaries

Each package represents a domain with minimal coupling:

```
CLI
 ↓
Application (backtest engine)
 ↓
Domain (strategy, portfolio, indicators)
 ↓
Infrastructure (CSV parser, execution model)
```

## Data Flow During Backtest

1. **Initialization**
   - Load strategy YAML
   - Parse into AST
   - Validate configuration
   - Initialize indicators
   - Initialize portfolio

2. **For each candle (chronologically)**
   - Update market data
   - Calculate indicators
   - Evaluate entry conditions
   - Evaluate exit conditions
   - Generate signals
   - Apply risk rules
   - Submit orders
   - Simulate execution
   - Update portfolio
   - Record metrics

3. **Finalization**
   - Calculate analytics
   - Generate reports

## Configuration Immutability

Strategy configuration is immutable during execution:

```go
StrategyConfig (immutable)
    +
RuntimeState (mutable)
    =
BacktestState
```

This prevents accidental strategy modifications during backtesting.

## Dependency Injection

The engine uses dependency injection to allow:
- Custom data feeds
- Custom execution models
- Custom indicators
- Custom analyzers

## Testing Strategy

Tests verify:
- Unit: Individual components (indicators, expressions, conditions)
- Integration: YAML → backtest pipeline
- Regression: Specific bugs and edge cases
- Golden: Known datasets with expected results

## Performance Considerations

1. **Indicator caching**: Calculated values cached per candle
2. **Dependency resolution**: Calculated once during initialization
3. **Lazy evaluation**: Conditions evaluated only when needed
4. **Memory efficiency**: Streaming data feeds for large datasets

## Future Extensions

The architecture supports:
- Live trading (swap execution engine)
- Paper trading (same as backtest)
- Multi-asset portfolios (extend portfolio model)
- Machine learning (custom indicator plugins)
- Cloud deployment (stateless design)

Without requiring core engine redesign.