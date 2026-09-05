AGENTS.md

Declarative Quantitative Trading Backtesting Engine

1. Project Identity

Project type: Go-based quantitative trading research and backtesting engine.

Primary goal:

Build a powerful, extensible, and deterministic quantitative trading research engine where trading strategies can be defined declaratively through YAML configuration instead of requiring strategy logic to be hardcoded in Go.

The system must support:

- Simple strategies defined entirely through YAML.
- Complex strategies through composable expressions.
- Custom indicators.
- Custom functions.
- Stateful strategies.
- Multi-timeframe analysis.
- Realistic order execution.
- Risk management.
- Portfolio simulation.
- Backtesting.
- Parameter optimization.
- Walk Forward Analysis.
- Monte Carlo analysis.
- Research/report generation.
- Future paper trading and live trading without redesigning the core architecture.

---

2. Core Product Philosophy

The project is NOT intended to be:

- A clone of Backtrader.
- A collection of hardcoded technical indicators.
- A YAML wrapper around "if/else" statements.
- A toy backtesting script.
- A framework where strategy logic is tightly coupled to the backtest engine.

The project MUST instead be designed as:

«A declarative, programmable quantitative trading research engine.»

The fundamental abstraction is:

Strategy Definition
        ↓
Parser
        ↓
Strategy AST / Intermediate Representation
        ↓
Dependency Resolution
        ↓
Evaluation
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

YAML is the user-facing representation of the strategy language.

The internal engine must NOT depend directly on YAML structures.

---

3. Primary Design Principle

YAML is an Interface, Not the Engine

Do NOT design the system as:

YAML
 ↓
Go structs
 ↓
if type == ...
 ↓
Execute

Instead:

YAML
 ↓
Parser
 ↓
AST / IR
 ↓
Compiler / Resolver
 ↓
Runtime Evaluator
 ↓
Execution Engine

This separation is mandatory.

Future strategy representations may include:

- YAML
- JSON
- CLI-generated strategies
- Web UI
- REST API
- Database
- Another DSL

All of them should be capable of compiling into the same internal strategy representation.

---

4. Product Goals

4.1 Must Have

The first stable version must support:

Data

- OHLCV data.
- Multiple symbols.
- Multiple timeframes.
- Historical data feeds.
- CSV input.
- Parquet input where practical.
- Deterministic chronological iteration.

Indicators

At minimum:

- SMA
- EMA
- RSI
- ATR

Indicator architecture must be extensible.

Expressions

Support:

+
-
*
/
%
>
<
>=
<=
==
!=
AND
OR
NOT

Trading conditions

Support:

cross_above
cross_below
rising
falling
between

Strategy

Support:

- Long entry.
- Short entry.
- Exit.
- Stop loss.
- Take profit.
- Trailing stop.
- Position sizing.

Orders

Support:

- Market.
- Limit.
- Stop.
- Stop-limit.

Execution simulation

Support:

- Fees.
- Slippage.
- Spread where data permits.
- Order status.
- Fill events.

Portfolio

Track:

- Cash.
- Equity.
- Balance.
- Margin where applicable.
- Positions.
- Realized PnL.
- Unrealized PnL.
- Fees.
- Exposure.

Analytics

At minimum:

- Total return.
- CAGR.
- Win rate.
- Profit factor.
- Expectancy.
- Maximum drawdown.
- Sharpe ratio.
- Sortino ratio.
- Number of trades.
- Average trade.
- Average win.
- Average loss.

---

5. Non-Goals for Initial MVP

Do NOT prioritize:

- GUI.
- Mobile application.
- Cloud deployment.
- Distributed backtesting.
- Machine learning.
- Reinforcement learning.
- Automatic strategy generation.
- Exchange execution.
- High-frequency trading.
- Tick-level simulation.

These may be added later.

The architecture should not prevent them, but they should not complicate the MVP unnecessarily.

---

6. High-Level Architecture

The target architecture is:

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

---

7. Repository Architecture

Preferred structure:

.
├── AGENTS.md
├── README.md
├── go.mod
├── go.sum
│
├── cmd/
│   └── trader/
│       └── main.go
│
├── internal/
│   ├── config/
│   │
│   ├── strategy/
│   │   ├── ast/
│   │   ├── parser/
│   │   ├── compiler/
│   │   ├── evaluator/
│   │   ├── state/
│   │   └── registry/
│   │
│   ├── expression/
│   │
│   ├── indicator/
│   │   ├── builtin/
│   │   └── registry/
│   │
│   ├── data/
│   │   ├── feed/
│   │   ├── csv/
│   │   ├── parquet/
│   │   └── resample/
│   │
│   ├── market/
│   │
│   ├── signal/
│   │
│   ├── order/
│   │
│   ├── execution/
│   │
│   ├── broker/
│   │
│   ├── portfolio/
│   │
│   ├── risk/
│   │
│   ├── backtest/
│   │
│   ├── analytics/
│   │
│   ├── optimization/
│   │
│   ├── walkforward/
│   │
│   └── montecarlo/
│
├── pkg/
│
├── strategies/
│   ├── examples/
│   └── tests/
│
├── data/
│
├── tests/
│
├── docs/
│
└── reports/

The exact structure may evolve.

Do not create packages merely to satisfy this tree.

Packages should represent meaningful domain boundaries.

---

8. Domain Model

Core domain objects should include concepts equivalent to:

MarketData
Candle
Tick
Timeframe
Symbol

Indicator
IndicatorValue

Expression
Condition
Signal

Strategy
StrategyState

Order
OrderRequest
OrderStatus
Fill

Position
Portfolio
Account

RiskModel
PositionSizer

Broker
ExecutionModel

Trade
EquityPoint

Backtest
BacktestResult

Metric
Analyzer

Avoid leaking implementation details between domains.

---

9. Event-Driven Architecture

The engine should use an event-driven model.

Conceptually:

MarketEvent
    ↓
Strategy Evaluation
    ↓
SignalEvent
    ↓
Risk Evaluation
    ↓
OrderEvent
    ↓
Execution
    ↓
FillEvent
    ↓
Portfolio Update

Potential event types:

MarketEvent
CandleEvent
SignalEvent
OrderSubmittedEvent
OrderAcceptedEvent
OrderRejectedEvent
OrderFilledEvent
OrderCancelledEvent
PositionOpenedEvent
PositionClosedEvent
PortfolioUpdatedEvent

Do not tightly couple strategy logic to portfolio implementation.

---

10. Strategy DSL

The strategy DSL is one of the most important components.

It must be:

- Declarative.
- Composable.
- Explicit.
- Validatable.
- Extensible.
- Deterministic.
- Independent from YAML implementation details.

---

11. Basic YAML Strategy

Example:

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

  atr:
    type: atr
    period: 14

  volume_avg:
    type: sma
    source: volume
    period: 20

  volume_ratio:
    type: divide
    left: volume
    right: volume_avg

entry:

  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [volume_ratio, 1.2]

exit:

  long:
    any:
      - cross_below: [ema_fast, ema_slow]

risk:

  position_size:
    type: risk_percent
    value: 0.01

  stop_loss:
    type: atr
    period: 14
    multiplier: 1.5

  take_profit:
    type: risk_reward
    ratio: 2

This should compile into an internal representation.

---

12. Expression System

Expressions must be first-class objects.

Examples:

expression:
  type: multiply
  left: close
  right: 1.02

or:

expression: "close * 1.02"

The string expression format may be implemented later.

Internally, expressions should become an AST.

Example:

close * 1.02

        *
       / \
   close  1.02

---

13. Expression Operators

The expression engine should eventually support:

Arithmetic

+
-
*
/
%

Comparison

>
<
>=
<=
==
!=

Logical

AND
OR
NOT

Temporal

previous
ref
shift

Functions

abs
min
max
sqrt
log
exp

Trading functions

cross_above
cross_below
highest
lowest
rising
falling
between

---

14. Historical References

Look-ahead bias is a critical concern.

The DSL must support historical references.

Example:

ref:
  value: ema21
  bars: 1

Meaning:

EMA21[t-1]

Possible syntax:

previous:
  value: close

or:

shift:
  value: close
  bars: 3

The implementation must make the temporal semantics explicit.

---

15. No Look-Ahead Bias

This is a HARD REQUIREMENT.

The engine must never allow future information to influence a decision at time "t".

Bad:

signal[t] uses close[t+1]

Good:

signal[t] uses close[t]
signal[t] uses close[t-1]
signal[t] uses indicator[t]
signal[t] uses indicator[t-n]

Every indicator and expression must have clearly defined evaluation timing.

Tests MUST exist to detect accidental look-ahead.

---

16. Indicator Architecture

Indicators must be registry-based.

Conceptually:

type Indicator interface {
    Name() string
    Calculate(ctx *EvaluationContext) Value
}

The exact interface may evolve.

The architecture must allow:

EMA
SMA
RSI
ATR
MACD
Bollinger Bands
ADX
VWAP
Custom indicators
Composite indicators

without modifying the core evaluator.

---

17. Indicator Registry

Use a registry:

"ema" → EMA implementation
"sma" → SMA implementation
"rsi" → RSI implementation
"atr" → ATR implementation

Avoid:

switch indicator.Type {
case "ema":
case "sma":
case "rsi":
case "atr":
}

inside the core engine.

A switch may exist inside a registry/factory boundary, but indicator implementations must remain independently extensible.

---

18. Composite Indicators

Indicators must be composable.

Example:

indicators:

  ema20:
    type: ema
    period: 20

  ema50:
    type: ema
    period: 50

  ema_distance:
    type: subtract
    left: ema20
    right: ema50

  ema_distance_pct:
    type: divide
    left: ema_distance
    right: ema50

The engine should resolve dependencies automatically.

---

19. Indicator Dependency Graph

Internally:

ema20 ──────┐
            ├── subtract ──→ ema_distance
ema50 ──────┘
                         │
                         ▼
                       divide
                         │
                         ▼
                  ema_distance_pct

The engine should:

1. Parse dependencies.
2. Validate references.
3. Detect cycles.
4. Build a dependency graph.
5. Resolve evaluation order.
6. Cache reusable values where appropriate.
7. Propagate validity status through the dependency graph.

Circular dependencies must produce a clear configuration error.

A composite indicator MUST NOT be considered valid until ALL its dependencies are valid. This prevents signals from being generated using incomplete indicator values during warm-up periods.

---

20. Custom Indicators

The engine must provide an extension mechanism.

At minimum:

registry.Register("my_indicator", implementation)

Custom indicators must not require modifications to the backtest engine.

Future plugin mechanisms may include:

- Go registration.
- Go plugins where appropriate.
- WASM.
- External processes.
- Script runtimes.

Do not implement all of these initially.

Design the interfaces so they remain possible.

---

21. Strategy State

Strategies must eventually support state.

Example:

state:

  setup_valid:
    default: false

Rules:

rules:

  - when:
      cross_above: [ema9, ema21]

    set:
      setup_valid: true

  - when:
      all:
        - eq: [setup_valid, true]
        - gt: [rsi, 50]

    action:
      enter: long

State must be isolated per:

- Strategy instance.
- Symbol.
- Backtest run.

State must never leak between separate backtests.

---

22. Multi-Timeframe Support

Multi-timeframe analysis is a first-class requirement.

Example:

data:

  primary:
    timeframe: 1h

  higher:
    timeframe: 4h

Indicators may specify their timeframe:

indicators:

  ema200_4h:
    type: ema
    timeframe: 4h
    period: 200

  rsi_1h:
    type: rsi
    timeframe: 1h
    period: 14

The engine must correctly align higher-timeframe information with lower-timeframe events.

This alignment must not introduce look-ahead.

---

23. Order Model

The engine must model orders explicitly.

Order types:

Market
Limit
Stop
StopLimit

Order states:

Pending
Accepted
PartiallyFilled
Filled
Cancelled
Rejected
Expired

Order objects should include:

ID
Symbol
Side
Type
Quantity
Price
StopPrice
Time
Status
Fees

---

24. Execution Model

Backtesting must not assume:

signal → instant fill

unless explicitly configured.

Execution should consider:

Price
Volume
Spread
Slippage
Fees
Order type
Intrabar assumptions

The execution model must be replaceable.

---

25. Slippage

Support configurable slippage.

Example:

execution:

  slippage:
    type: percentage
    value: 0.0005

Potential future models:

Fixed
Percentage
Volatility-based
Volume-based
Custom

---

26. Fees

Fees must be configurable.

Example:

fees:

  maker: 0.0002
  taker: 0.0005

The engine must distinguish maker/taker where execution semantics allow it.

---

27. Position Sizing

Position sizing must be independent from signal generation.

Examples:

position_size:

  type: fixed
  value: 100

position_size:

  type: percent_equity
  value: 0.1

position_size:

  type: risk_percent
  value: 0.01

The last one must account for stop-loss distance.

Conceptually:

risk_amount
───────────────
stop_distance
=
position_size

---

28. Risk Engine

Risk management must be a separate domain.

Potential controls:

Risk per trade
Maximum position size
Maximum number of positions
Maximum portfolio exposure
Maximum daily loss
Maximum drawdown
Maximum leverage
Maximum concurrent risk

Risk rules should be evaluated before order submission.

---

29. Stop Loss

Support:

Fixed price
Percentage
ATR-based
Structure-based
Expression-based

Example:

stop_loss:

  type: atr
  period: 14
  multiplier: 1.5

---

30. Take Profit

Support:

Fixed price
Percentage
Risk/reward
Indicator-based
Expression-based
Multiple targets

Example:

take_profit:

  type: risk_reward
  ratio: 2

Future:

take_profit:

  targets:

    - percent: 50
      risk_reward: 1

    - percent: 50
      risk_reward: 3

---

31. Trailing Stop

Support:

Percentage
ATR
Highest/lowest since entry
Expression

Trailing stops must be evaluated using correct chronological data.

---

32. Long and Short

The engine should support both:

LONG
SHORT

where the market model permits it.

Do not assume every market supports shorting.

Market constraints should be represented by the broker/execution layer.

---

33. Portfolio Model

Portfolio must track:

Cash
Equity
Balance
Positions
Exposure
Margin
Realized PnL
Unrealized PnL
Fees

For each position:

Symbol
Side
Quantity
Average entry
Current price
Stop loss
Take profit
Unrealized PnL
Realized PnL

---

34. Backtest Engine

The backtest engine should be deterministic.

Conceptually:

Initialize
 ↓
Load configuration
 ↓
Validate strategy
 ↓
Load data
 ↓
Initialize indicators
 ↓
Initialize portfolio
 ↓
Initialize broker
 ↓
For each chronological event:
      update market
      update indicators
      evaluate strategy
      generate signals
      evaluate risk
      submit orders
      simulate execution
      update portfolio
      record state
 ↓
Finalize
 ↓
Calculate analytics
 ↓
Generate report

---

35. Determinism

Given:

same strategy
same data
same configuration
same seed

the engine must produce the same result.

Randomized components such as Monte Carlo must use explicit seeds when reproducibility is required.

---

36. Data Layer

Data feeds must be abstracted.

Conceptually:

type DataFeed interface {
    Next() (MarketEvent, error)
}

Implementations may include:

CSV
Parquet
Database
Exchange API
WebSocket

Only implement what is necessary for the current phase.

---

37. Data Quality

The engine should validate:

Timestamp ordering
Duplicate timestamps
Missing OHLC values
Invalid OHLC relationships
Negative volume
Invalid prices
Timezone consistency

Invalid data should produce actionable errors.

---

38. Time Handling

All internal timestamps should have explicit timezone semantics.

Prefer:

UTC internally

Convert to local time only at presentation boundaries.

Never silently mix timezones.

---

39. CLI

The CLI is the primary interface during MVP.

Expected commands:

trader validate --strategy strategy.yaml

trader backtest \
  --strategy strategy.yaml \
  --data BTCUSDT.parquet

trader optimize \
  --strategy strategy.yaml \
  --data BTCUSDT.parquet

trader walkforward \
  --strategy strategy.yaml \
  --data BTCUSDT.parquet

trader report \
  --result result.json

Exact command names may evolve.

---

40. Validation

A strategy must be validated before execution.

Validation should detect:

Unknown indicator
Unknown function
Unknown condition
Invalid parameter
Missing parameter
Invalid type
Invalid reference
Circular dependency
Invalid timeframe
Invalid order configuration
Invalid risk configuration

Errors must identify the relevant configuration path.

Example:

strategy.indicators.ema_fast.period:
expected positive integer, got -5

Avoid generic errors such as:

invalid config

---

41. Error Handling

Errors should be:

- Explicit.
- Contextual.
- Actionable.
- Typed where useful.

Use Go error wrapping:

fmt.Errorf("parse strategy indicators: %w", err)

Do not silently ignore errors.

Do not panic for normal user configuration errors.

---

42. Analytics

Analytics must be modular.

Conceptually:

type Analyzer interface {
    Analyze(result BacktestResult) Metric
}

Metrics:

Total Return
CAGR
Sharpe
Sortino
Max Drawdown
Calmar
Win Rate
Profit Factor
Expectancy
Average Trade
Average Win
Average Loss
Exposure
Trade Count

---

43. Equity Curve

Record equity over time.

Required output:

timestamp
equity
cash
drawdown

This data should be exportable.

---

44. Trade Journal

Every completed trade must have a record.

Example:

Trade ID
Symbol
Side
Entry Time
Entry Price
Exit Time
Exit Price
Quantity
Gross PnL
Fees
Net PnL
Return
MAE
MFE
Exit Reason

This is important for quantitative research.

---

45. Output Formats

Backtest results should support machine-readable output.

At minimum:

JSON
CSV

Potential future output:

HTML
SQLite
Parquet

---

46. Optimization

Parameter optimization is a future major feature.

Example:

optimization:

  parameters:

    ema_fast:
      range: [5, 20]
      step: 1

    ema_slow:
      range: [20, 100]
      step: 5

    atr_multiplier:
      range: [1.0, 3.0]
      step: 0.25

  objective:
    type: sharpe

Architecture must allow multiple optimization algorithms:

Grid Search
Random Search
Genetic Algorithm
Bayesian Optimization

Do not implement all initially.

---

47. Walk Forward Analysis

Walk Forward Analysis is an important research feature.

Configuration:

walk_forward:

  enabled: true

  train:
    bars: 2000

  test:
    bars: 500

  step:
    bars: 500

The engine should produce separate results for:

Training
Validation
Testing

and aggregate out-of-sample performance.

---

48. Monte Carlo

Future module.

Potential analysis:

Trade reshuffling
Return reshuffling
Drawdown distribution
Probability of ruin
Confidence intervals

Example:

monte_carlo:

  simulations: 10000

  seed: 42

---

49. Research Integrity

The system must explicitly distinguish:

In-sample
Out-of-sample
Backtest
Paper trade
Live trade

Reports should make it difficult to accidentally interpret optimized in-sample performance as evidence of robustness.

---

50. Avoiding Overfitting

The project should eventually provide research tools such as:

Parameter sensitivity
Walk Forward Analysis
Monte Carlo
Out-of-sample testing
Trade distribution
Drawdown analysis

The system should NOT claim that a profitable backtest proves strategy validity.

---

51. Testing Philosophy

Testing is a first-class requirement.

Tests must exist for:

Unit tests

- Indicators.
- Expressions.
- Conditions.
- Parser.
- Compiler.
- Position sizing.
- Fees.
- Slippage.
- Orders.
- Portfolio accounting.
- Analytics.

Integration tests

- YAML → AST.
- AST → runtime.
- Strategy → signal.
- Signal → order.
- Order → fill.
- Fill → portfolio.
- Complete backtest.

Regression tests

Every discovered bug should ideally become a regression test.

---

52. Numerical Testing

Financial calculations require careful handling.

Tests should account for:

- Floating point precision.
- Rounding.
- Fees.
- Position size precision.
- Price precision.
- Quantity precision.

Avoid arbitrary equality checks for floating point values.

---

53. Look-Ahead Regression Tests

Explicitly test scenarios where future data would incorrectly improve results.

Example:

If candle[t+1] changes,
signal[t] must not change.

This should become a permanent regression test.

---

54. Golden Backtest Tests

Create small deterministic datasets.

Example:

5–20 candles
known strategy
known trades
known PnL

Expected results should be asserted.

These tests are extremely valuable when refactoring the engine.

---

55. Performance

Go is selected partly for performance.

However:

«Correctness > premature optimization.»

Initial priorities:

1. Correctness
2. Determinism
3. Extensibility
4. Testability
5. Performance

Optimize only after profiling.

---

56. Caching

Indicator values may be cached where useful.

However, caching must never alter semantics.

Potential caches:

Indicator result
Expression result
Historical lookup
Dependency graph
Compiled strategy

---

57. Concurrency

Backtests for independent configurations may eventually run concurrently.

Example:

Strategy A ─┐
Strategy B ─┼──→ Worker Pool
Strategy C ─┘

Do not introduce concurrency inside the core event loop unless necessary.

Avoid shared mutable state.

---

58. Extensibility Rules

When implementing a new feature, ask:

«Does this require modifying existing core logic, or can it be registered through an interface?»

Prefer:

Interface
Registry
Dependency Injection
Composition

over:

Huge switch statement
Huge conditional tree
Global mutable state

---

59. Plugin Architecture

Future custom functionality may include:

Custom Indicator
Custom Function
Custom Risk Model
Custom Position Sizer
Custom Execution Model
Custom Analyzer
Custom Data Feed

Each should ideally have a registration mechanism.

---

60. Strategy Versioning

Strategy configuration should support versioning.

Example:

strategy:
  name: ema_volume
  version: "1"

Breaking DSL changes should not silently reinterpret old strategies.

---

61. Configuration Schema

The project should eventually provide a machine-readable schema for strategy YAML.

Benefits:

- Editor autocomplete.
- Validation.
- Documentation.
- Better AI-agent generation.
- Better UX.

---

62. Example Strategy Library

The repository should contain examples:

strategies/examples/

├── sma_cross.yaml
├── ema_cross.yaml
├── ema_volume.yaml
├── rsi_reversal.yaml
├── breakout.yaml
├── trend_following.yaml
├── atr_stop.yaml
├── multi_timeframe.yaml
└── stateful_setup.yaml

Each example should demonstrate a specific capability.

---

63. Documentation

Documentation should include:

Getting Started
Architecture
Strategy DSL
Indicators
Expressions
Conditions
Orders
Risk Management
Multi-Timeframe
Custom Indicators
Backtesting
Optimization
Walk Forward
Analytics
CLI
Developer Guide

Do not document functionality that does not exist.

---

64. CLI UX

CLI output should be readable by humans.

Example:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BACKTEST RESULT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Strategy       EMA Volume Trend
Symbol         BTCUSDT
Timeframe      4H

Period         2020-01-01 → 2026-01-01

Return         +183.42%
CAGR           +19.82%
Sharpe         1.67
Sortino        2.31
Max Drawdown   -21.43%

Trades         428
Win Rate       47.66%
Profit Factor  1.84
Expectancy     +0.43R

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Machine-readable output must remain available.

---

65. Logging

Use structured logging where appropriate.

Log levels:

debug
info
warn
error

Avoid excessive logs inside high-frequency loops unless debug mode is enabled.

---

66. Configuration Immutability

Once a backtest begins, the strategy configuration should be treated as immutable.

Runtime state must live separately from configuration.

Bad:

StrategyConfig mutated during execution

Good:

StrategyConfig
      +
RuntimeState

---

67. Separation of Concerns

Do not mix:

YAML parsing
indicator calculation
order execution
portfolio accounting
analytics

inside the same package or function.

A function doing too many domain responsibilities should be considered a design smell.

---

68. Dependency Direction

Preferred dependency direction:

CLI
 ↓
Application
 ↓
Domain
 ↓
Infrastructure

Infrastructure should not define core trading semantics.

For example:

CSV parser

must not determine:

how a stop-loss works

---

69. Domain First

When implementing features, define domain semantics before implementation details.

Example:

Before implementing ATR stop loss, define:

What price is used?
When is ATR evaluated?
When does stop become active?
What happens on gaps?
What happens when SL and TP are both hit in one candle?

Then implement.

---

70. Intrabar Ambiguity

OHLC candle data does not reveal the exact order of price movement within a candle.

If both:

Stop Loss
Take Profit

are inside the same candle range, the engine must have an explicit policy.

Possible policies:

Conservative
Optimistic
Nearest-first
Worst-case
Best-case
Lower-timeframe reconstruction

Never silently choose a behavior without documenting it.

---

71. Market Gaps

Execution must account for cases where:

requested price

is unavailable because price gaps beyond it.

Stop orders must not always be assumed to fill at the stop price.

---

72. Short-Selling Semantics

For short positions, correctly model:

Entry
Exit
PnL
Fees
Margin
Liquidation where applicable

Do not assume long and short accounting are identical without verification.

---

73. Leverage

Leverage should be modeled separately from position sizing.

Potential future configuration:

account:

  leverage: 5

Do not implement leverage as:

position_size *= leverage

without proper margin semantics.

---

74. Futures / Perpetual Support

The architecture should eventually support:

Funding fees
Margin
Leverage
Liquidation
Mark price
Index price
Maintenance margin

These are future features.

Do not pollute the MVP with incomplete futures semantics.

---

75. Crypto-Oriented Design

The project should be suitable for crypto quantitative research.

Important future concepts:

24/7 market
Funding rates
Exchange-specific fees
Symbol precision
Quantity precision
Minimum order size
Maker/taker fees
Perpetual contracts

But the core engine should remain market-agnostic.

---

76. Strategy Example: User's Typical Strategy

The engine should be capable of representing a strategy such as:

EMA 9 / EMA 21
+
Volume confirmation
+
ATR stop
+
1% risk per trade
+
1:2 R:R

Example:

strategy:
  name: ema_volume_atr

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

  volume_avg:
    type: sma
    source: volume
    period: 20

  volume_ratio:
    type: divide
    left: volume
    right: volume_avg

  atr:
    type: atr
    period: 14

entry:

  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [volume_ratio, 1.2]

risk:

  position_size:
    type: risk_percent
    value: 0.01

  stop_loss:
    type: atr
    indicator: atr
    multiplier: 1.5

  take_profit:
    type: risk_reward
    ratio: 2

The engine must be able to express this without custom Go code.

---

77. Advanced Strategy Example

Eventually support logic such as:

entry:

  long:

    all:

      - gt: [close, ema200]

      - or:

          - all:
              - cross_above: [ema9, ema21]
              - gt: [volume_ratio, 1.2]

          - all:
              - between: [rsi, 40, 50]
              - rising: [rsi]

      - not:
          - gt: [volatility, volatility_limit]

The engine must treat this as a composable expression tree rather than special-case logic.

---

78. Stateful Strategy Example

Eventually:

state:

  breakout_setup:
    default: false

rules:

  - when:
      gt: [close, resistance]

    set:
      breakout_setup: true

  - when:
      all:
        - eq: [breakout_setup, true]
        - gt: [volume_ratio, 1.5]

    action:
      enter: long

  - when:
      bars_since:
        event: breakout_setup
        greater_than: 10

    set:
      breakout_setup: false

The exact syntax can evolve.

The internal architecture must support state machines.

---

79. Future Live Trading Architecture

The eventual architecture should permit:

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

The Strategy DSL should not know whether execution is simulated or real.

---

80. Security

Custom code execution must be treated carefully.

Do not execute arbitrary user-provided Go/Python code without an explicit security boundary.

Future scripting/plugin systems must be sandboxed where possible.

---

81. AI-Agent Development Rules

Coding agents working on this repository MUST:

1. Read "AGENTS.md" before making changes.
2. Understand existing architecture before adding code.
3. Search existing abstractions before creating new ones.
4. Avoid duplicating functionality.
5. Add tests for meaningful behavior.
6. Preserve deterministic behavior.
7. Avoid unnecessary dependencies.
8. Avoid premature optimization.
9. Never silently change strategy semantics.
10. Never introduce look-ahead bias.
11. Document breaking DSL changes.
12. Keep domain boundaries clean.

---

82. AI Agent Implementation Protocol

For every feature:

1. Understand requirement
2. Inspect repository
3. Identify affected domain
4. Design minimal abstraction
5. Implement
6. Add unit tests
7. Add integration tests where appropriate
8. Run formatter
9. Run static analysis
10. Run tests
11. Review for architectural violations
12. Summarize changes

Do not immediately start coding before inspecting the repository.

---

83. AI Agent Must Avoid

Never introduce:

God objects
God functions
Global mutable state
Hidden singleton state
Circular dependencies
Hardcoded strategy logic
Hardcoded indicator lists in core engine
YAML-specific logic in domain layer
Silent error handling
Implicit timezone conversions
Future-data access
Unbounded memory growth

---

84. Definition of Done

A feature is NOT complete merely because:

go build

passes.

A feature is complete when:

Implementation
+
Tests
+
Validation
+
Documentation
+
Architecture consistency

are satisfied.

---

85. Required Quality Gates

Before merging significant changes:

go fmt ./...
go vet ./...
go test ./...

Where applicable:

go test -race ./...

Additional static analysis may be introduced later.

---

86. Phase-Based Development

The project should be developed incrementally.

Phase 0 — Architecture Foundation

Deliver:

Go module
CLI skeleton
Domain model
Package boundaries
Error model
Basic test infrastructure

No complex trading functionality yet.

---

Phase 1 — Market Data

Deliver:

Candle
OHLCV
Time
CSV DataFeed
Deterministic iteration
Data validation

Tests must verify chronological behavior.

---

Phase 2 — Indicator Engine

Deliver:

Indicator interface
Registry
SMA
EMA
RSI
ATR
Indicator dependency resolution

---

Phase 3 — Expression Engine

Deliver:

Arithmetic
Comparison
Logical operators
Cross detection
Historical references
Expression AST

---

Phase 4 — Strategy DSL

Deliver:

YAML parser
Validation
Strategy AST / IR
Indicator configuration
Entry rules
Exit rules

---

Phase 5 — Backtest Core

Deliver:

Event loop
Signals
Orders
Market execution
Portfolio
PnL
Trade records

---

Phase 6 — Realistic Execution

Deliver:

Limit orders
Stop orders
Fees
Slippage
Intrabar ambiguity policy
Order lifecycle

---

Phase 7 — Risk Management

Deliver:

Risk-per-trade
Position sizing
Stop loss
Take profit
Trailing stop
Portfolio risk limits

---

Phase 8 — Analytics

Deliver:

Equity curve
Drawdown
Sharpe
Sortino
Profit factor
Expectancy
Trade statistics
JSON/CSV output

---

Phase 9 — Advanced DSL

Deliver:

State
Functions
Composite indicators
Expression functions
Advanced conditions

---

Phase 10 — Multi-Timeframe

Deliver:

Multiple timeframe feeds
Timeframe alignment
MTF indicators
MTF conditions
Look-ahead-safe synchronization

---

Phase 11 — Optimization

Deliver:

Parameter definitions
Grid search
Optimization metrics
Optimization reports

---

Phase 12 — Walk Forward

Deliver:

Training windows
Testing windows
Rolling windows
Out-of-sample reports

---

Phase 13 — Monte Carlo

Deliver:

Trade reshuffling
Simulation
Drawdown distribution
Confidence intervals

---

Phase 14 — Extensibility

Deliver:

Custom indicators
Custom functions
Custom analyzers
Custom execution models

---

87. MVP Acceptance Criteria

The MVP is considered successful when a user can create a strategy like:

strategy:
  name: ema_volume

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:

  ema9:
    type: ema
    period: 9

  ema21:
    type: ema
    period: 21

  volume_avg:
    type: sma
    source: volume
    period: 20

entry:

  long:
    all:
      - cross_above: [ema9, ema21]
      - gt:
          - volume
          - mul: [volume_avg, 1.2]

risk:

  position_size:
    type: risk_percent
    value: 0.01

  stop_loss:
    type: atr
    period: 14
    multiplier: 1.5

  take_profit:
    type: risk_reward
    ratio: 2

Then run:

trader validate --strategy strategy.yaml

followed by:

trader backtest \
  --strategy strategy.yaml \
  --data BTCUSDT.parquet

and receive:

Return
CAGR
Sharpe
Sortino
Max Drawdown
Win Rate
Profit Factor
Expectancy
Trade Count

with a reproducible trade history.

---

88. Critical Architectural Invariants

The following must remain true throughout development:

Invariant 1

Strategy definitions must not require Go code for ordinary strategies.

Invariant 2

The backtest engine must not know specific strategy logic.

Invariant 3

The expression engine must not know YAML.

Invariant 4

The domain model must not depend on the CLI.

Invariant 5

Indicators must be extensible through registration/composition.

Invariant 6

Execution must be separated from signal generation.

Invariant 7

Risk management must be separated from strategy conditions.

Invariant 8

Backtests must be deterministic.

Invariant 9

Future information must never be accessible during historical evaluation.

Invariant 10

A future live-trading runtime should be able to reuse the strategy and domain layers.

---

89. Architectural Decision Priority

When requirements conflict, prioritize:

1. Correctness
2. No look-ahead bias
3. Determinism
4. Clear domain semantics
5. Extensibility
6. Testability
7. Maintainability
8. Performance
9. Convenience

Do not sacrifice correctness for performance.

Do not sacrifice architecture merely to make the first implementation shorter.

---

90. Long-Term Vision

The final system should evolve into:

              Quant Research Platform
                       │
        ┌──────────────┼──────────────┐
        │              │              │
        ▼              ▼              ▼
    Backtesting   Optimization   Robustness
        │              │              │
        ▼              ▼              ▼
    Strategies      WFA          Monte Carlo
        │
        ▼
    Strategy DSL
        │
        ▼
 ┌───────────────────────┐
 │ Trading Engine Core   │
 └───────────────────────┘
        │
 ┌──────┼──────────┐
 ▼      ▼          ▼
Backtest Paper    Live

The project's ultimate purpose is not merely to answer:

«"Apakah strategy ini profit?"»

It should help answer:

«"Apakah strategy ini logically correct, statistically meaningful, robust terhadap perubahan parameter, realistic terhadap execution, dan memiliki kemungkinan untuk survive di luar sample?"»

That distinction is fundamental to the project.

---

91. Final Instruction to Coding Agents

When uncertain between two implementations:

Prefer the implementation that:

- keeps the domain model clean,
- preserves future extensibility,
- makes semantics explicit,
- is easy to test,
- prevents look-ahead bias,
- does not hardcode strategy behavior,
- does not couple YAML to the engine,
- and can later support optimization and live execution.

Do not optimize for the smallest amount of code.

Optimize for a small, correct, composable foundation.

The project should grow by composition, not by accumulating special cases.
