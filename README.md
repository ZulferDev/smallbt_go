# Jcode - Declarative Quantitative Trading Backtesting Engine

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Build Status](https://img.shields.io/github/actions/workflow/status/ZulferDev/smallbt_go/ci.yml?branch=master)](https://github.com/ZulferDev/smallbt_go/actions)
[![AGENTS.md Compliant](https://img.shields.io/badge/architecture-AGENTS.md%20compliant-success)](AGENTS.md)

A powerful, extensible, and deterministic quantitative trading research engine where trading strategies are defined **declaratively through YAML** instead of requiring hardcoded Go logic.

**Key Philosophy**: *YAML is an interface, not the engine.* The strategy DSL compiles to an intermediate representation, making the engine independent of the configuration format.

## 🎯 **MVP Complete & Verified**

✅ **All 15 development phases complete** (AGENTS.md §86)  
✅ **All acceptance criteria met** (AGENTS.md §87)  
✅ **Zero look-ahead bias guaranteed**  
✅ **Deterministic execution verified**  
✅ **Test suite: 31 test files passing**

## 🚀 Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/ZulferDev/smallbt_go.git
cd jcode

# Build CLI
go build -o ./bin/trader ./cmd/trader

# Or install globally
go install ./cmd/trader
```

### Your First Strategy

Create `my_strategy.yaml`:

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
  volume_avg:
    type: sma
    source: volume
    period: 20
  atr:
    type: atr
    period: 14

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [volume, mul: [volume_avg, 1.2]]

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
```

### Run a Backtest

```bash
# Validate strategy
trader validate --strategy my_strategy.yaml

# Run backtest
trader backtest \
  --strategy my_strategy.yaml \
  --data data/BTCUSDT.csv \
  --output results/
```

### View Results

The engine produces:

1. **Human-readable output**:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BACKTEST RESULT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
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
```

2. **Machine-readable JSON**:
```json
{
  "return": 1.8342,
  "cagr": 0.1982,
  "sharpe": 1.67,
  "sortino": 2.31,
  "max_drawdown": -0.2143,
  "trades": 428,
  "win_rate": 0.4766,
  "profit_factor": 1.84,
  "expectancy": 0.43
}
```

3. **Trade journal CSV** with every completed trade

## 📊 **Available CLI Commands**

| Command | Description |
|---------|-------------|
| `trader validate --strategy file.yaml` | Validate strategy syntax and dependencies |
| `trader backtest --strategy file.yaml --data data.csv` | Run backtest with realistic execution |
| `trader optimize --strategy file.yaml --data data.csv` | Parameter optimization with grid search |
| `trader walkforward --strategy file.yaml --data data.csv` | Walk-forward analysis |
| `trader montecarlo --strategy file.yaml --data data.csv` | Monte Carlo simulation |

## 🏗️ **Architecture**

The engine follows a clean separation of concerns:

```
YAML Strategy → Parser → Strategy AST/IR → Compiler → Runtime Evaluator
        ↓
Dependency Graph → Indicator/Expression Evaluation
        ↓
Signal Engine → Risk Engine → Order Engine
        ↓
Execution/Broker → Portfolio → Analytics
```

**Core design principle**: The engine never knows specific strategy logic. Strategy definitions compile to an intermediate representation that can be executed without understanding the original YAML.

## 📈 **Features**

### ✅ **Complete MVP** (Phase 1-15)

- **Declarative Strategy DSL**: Define strategies entirely in YAML
- **Indicator Engine**: SMA, EMA, RSI, ATR with registry-based extensibility
- **Expression System**: Arithmetic, comparison, logical operators
- **Backtest Core**: Deterministic event-driven simulation
- **Realistic Execution**: Fees, slippage, order types (market/limit/stop)
- **Risk Management**: Position sizing, stop loss, take profit, trailing stop
- **Portfolio Model**: Cash, equity, PnL, margin where applicable
- **Analytics**: Return, CAGR, Sharpe, Sortino, drawdown, win rate, profit factor
- **Parameter Optimization**: Grid search with configurable metrics
- **Walk Forward Analysis**: Training/testing windows for robustness
- **Monte Carlo Simulation**: Trade reshuffling for confidence intervals

### 🔄 **Post-MVP Roadmap** (Phase 16-25)

See [ROADMAP.md](ROADMAP.md) for detailed future phases:

1. **Performance Optimization** (Phase 15)
2. **Live Trading Architecture** (Phase 16)
3. **Enhanced Data Handling** (Phase 17)
4. **Advanced Strategy DSL** (Phase 18)
5. **Advanced Analytics & Reporting** (Phase 19)
6. **Portfolio Analysis** (Phase 20)
7. **Machine Learning Integration** (Phase 21)
8. **Exchange Integration** (Phase 22)
9. **Robustness & Stress Testing** (Phase 23)
10. **Cloud & Distributed Computing** (Phase 24)
11. **Documentation & Community** (Phase 25)

## 🎛️ **Strategy Examples**

11 ready-to-use examples in `strategies/examples/`:

- `sma_cross.yaml` - Simple moving average crossover
- `ema_volume.yaml` - EMA crossover with volume confirmation
- `rsi_reversal.yaml` - RSI oversold/overbought reversal
- `breakout.yaml` - Support/resistance breakout
- `trend_following.yaml` - Multi-indicator trend following
- `atr_stop.yaml` - ATR-based trailing stop
- `multi_timeframe.yaml` - Multi-timeframe analysis
- `stateful_setup.yaml` - State machine example
- `composite_indicator.yaml` - Custom indicator composition
- `complex_entry.yaml` - Multiple entry conditions
- `risk_managed.yaml` - Complete risk management example

## 🧠 **Research Integrity**

The engine explicitly distinguishes:

| Mode | Purpose |
|------|---------|
| **Backtest** | Initial strategy development |
| **Optimization** | Parameter tuning |
| **Walk Forward** | Out-of-sample validation |
| **Monte Carlo** | Confidence estimation |
| **Paper Trade** | Real data simulation |
| **Live Trade** | Real execution |

**Critical guarantee**: The system never claims a profitable backtest proves strategy validity. Research tools emphasize statistical significance over isolated performance metrics.

## 🛡️ **Architectural Invariants**

1. **Declarative strategies** - No Go code required for ordinary strategies
2. **Zero look-ahead bias** - Temporal semantics always explicit
3. **Deterministic execution** - Same inputs → same outputs
4. **Clean domain boundaries** - No leaky abstractions
5. **Extensibility first** - Register, don't hardcode
6. **Test coverage** - All meaningful behavior tested
7. **Backward compatibility** - Strategies remain valid across versions
8. **Market agnostic** - Crypto/equities/forex support through abstraction
9. **Performance vs correctness** - Correctness always prioritized
10. **Live trading readiness** - Strategy layer identical for backtest and live

## 🧪 **Testing Philosophy**

- **Unit tests**: Individual components in isolation
- **Integration tests**: Cross-package interactions
- **Regression tests**: Every bug becomes a test
- **Golden backtests**: Small deterministic datasets with known results
- **Look-ahead regression**: Explicit tests prevent future data access
- **Determinism verification**: Same inputs always produce same outputs

```bash
# Run all tests
go test -race -cover ./...

# Test specific package
go test ./internal/backtest

# Run benchmarks
go test -bench=. -benchmem ./...
```

## 📚 **Documentation**

| Document | Purpose |
|----------|---------|
| [AGENTS.md](AGENTS.md) | **Architectural specification** (required reading) |
| [MVP_COMPLETION_REPORT.md](MVP_COMPLETION_REPORT.md) | **Complete MVP verification** |
| [VERIFICATION_SUMMARY.txt](VERIFICATION_SUMMARY.txt) | **Execution evidence** |
| [ROADMAP.md](ROADMAP.md) | **Future development phases** |
| [CONTRIBUTING.md](CONTRIBUTING.md) | **Contribution guidelines** |

**Guides** in `docs/`:
- `docs/GETTING_STARTED.md` - First steps tutorial
- `docs/STRATEGY_DSL.md` - Complete DSL reference
- `docs/INDICATORS_CONDITIONS.md` - Indicator and expression guide
- `docs/BACKTESTING.md` - Backtesting concepts and configuration
- `docs/OPTIMIZATION_WALKFORWARD_MONTECARLO.md` - Advanced research methods
- `docs/DEVELOPER_GUIDE.md` - Architecture and extension points
- `docs/ARCHITECTURE.md` - System design and patterns

## 🏭 **Production Readiness**

### ✅ **Quality Gates**
- All tests passing (race detection enabled)
- Code formatted (gofmt)
- Static analysis passing (go vet, golangci-lint)
- Security audit (gosec)
- Documentation verification
- Performance regression tracking

### ✅ **CI/CD Pipeline**
- GitHub Actions with comprehensive workflow
- Multi-platform binary releases
- Automated testing on push/PR
- Release automation on version tags

## 🤝 **Contributing**

We welcome contributions! Please read:
1. [AGENTS.md](AGENTS.md) - Architectural requirements
2. [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
3. [ROADMAP.md](ROADMAP.md) - Future development priorities

**Critical**: Never violate the architectural invariants, especially **zero look-ahead bias** and **deterministic execution**.

## 📄 **License**

MIT License - see [LICENSE](LICENSE) for details.

## 🎉 **Acknowledgments**

Built according to the comprehensive architectural specification in AGENTS.md, which guided development through 15 phases to a complete, verified MVP.

---

**Ready for quantitative trading research that prioritizes correctness over convenience.**

---

## 📊 Paper Trading

**Status:** ✅ Production Ready (Phase 16)

Paper trading simulates real-time execution without risking real capital. Perfect for strategy validation before live deployment.

### Quick Start

```bash
# Static price simulation
trader paper --strategy strategies/examples/paper_ema_cross.yaml \
             --symbol BTCUSDT \
             --price 50000 \
             --duration 60

# Real-time WebSocket data
trader paper --strategy strategies/examples/paper_ema_cross.yaml \
             --symbol BTCUSDT \
             --websocket ws://localhost:8080 \
             --duration 300
```

### Features

- ✅ **Realistic latency simulation** (50-200ms)
- ✅ **WebSocket real-time data feed**
- ✅ **Portfolio tracking** (cash, equity, positions)
- ✅ **Same strategy YAML as backtesting**
- ✅ **Order lifecycle simulation**
- ✅ **Real-time status updates**

### Example Output

```
Starting paper trading...
Strategy: paper_ema_cross
Symbol: BTCUSDT
WebSocket: ws://localhost:8080

Connected to WebSocket
Subscribing to: BTCUSDT

[Candle 1] 15:26:25 | O:50000.00 H:50100.00 L:49900.00 C:50050.00 V:1500.00
[Candle 2] 15:26:30 | O:50050.00 H:50150.00 L:50000.00 C:50100.00 V:1200.00

[5s] Balance: 10000.00 | Equity: 10000.00 | Positions: 0 | Candles: 2
```

### CLI Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--strategy` | Strategy YAML file | *required* | `--strategy ema.yaml` |
| `--symbol` | Trading symbol | BTCUSDT | `--symbol ETHUSDT` |
| `--price` | Initial price (static mode) | 50000.0 | `--price 45000` |
| `--balance` | Initial balance | 10000.0 | `--balance 50000` |
| `--duration` | Duration in seconds | 60 | `--duration 300` |
| `--websocket` | WebSocket URL (optional) | - | `--websocket ws://localhost:8080` |

### WebSocket Protocol

Paper trading expects JSON messages with OHLCV candle data:

```json
{
  "timestamp": 1609459200,
  "open": 50000.0,
  "high": 50100.0,
  "low": 49900.0,
  "close": 50050.0,
  "volume": 1000.0
}
```

**See:** [Paper Trading Guide](docs/PAPER_TRADING_GUIDE.md) for detailed documentation.

### Architecture

```
WebSocket Server
    ↓
WebSocketFeed (Week 3)
    ↓
Subscribe() → candle channel
    ↓
PaperBroker (Week 2)
    ↓
Order Queue + Latency Simulation
    ↓
Portfolio Updates
```

**Components:**
- **WebSocketFeed:** Real-time data connection with auto-reconnection
- **PaperBroker:** Order execution simulation with realistic latency
- **Portfolio:** Balance and position tracking
- **Order Queue:** Asynchronous order processing

### Workflow

1. **Backtest** your strategy with historical data
2. **Paper trade** with WebSocket real-time data
3. **Review results** and iterate
4. **Deploy to live** trading (Phase 17+)

---

