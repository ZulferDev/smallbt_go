# smallbt_go - Declarative Quantitative Trading Backtesting Engine

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Build Status](https://img.shields.io/github/actions/workflow/status/ZulferDev/smallbt_go/ci.yml?branch=master)](https://github.com/ZulferDev/smallbt_go/actions)
[![AGENTS.md Compliant](https://img.shields.io/badge/architecture-AGENTS.md%20compliant-success)](AGENTS.md)

A powerful, extensible, and deterministic quantitative trading research engine where trading strategies are defined **declaratively through YAML** instead of requiring hardcoded Go logic.

**Key Philosophy**: *YAML is an interface, not the engine.* The strategy DSL compiles to an intermediate representation, making the engine independent of the configuration format.

---

## 🎯 Project Status

✅ **Core Engine:** Production-ready  
✅ **Transform System:** Complete (8 types)  
✅ **Slippage Models:** Complete (5 models)  
✅ **Documentation:** Comprehensive (3,882 lines)  
✅ **Test Suite:** 26/26 packages passing  
✅ **Zero look-ahead bias guaranteed**  
✅ **Deterministic execution verified**

---

## 🚀 Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/ZulferDev/smallbt_go.git
cd smallbt_go

# Build CLI
go build -o trader cmd/trader/main.go

# Or install globally
go install ./cmd/trader
```

### Your First Strategy

Create `my_strategy.yaml`:

```yaml
strategy:
  name: ema_volume_trend
  version: "1"
  description: "EMA crossover with volume confirmation"

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

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [volume, volume_avg * 1.2]

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
```

### Run a Backtest

```bash
# Validate strategy
trader validate --strategy my_strategy.yaml

# Run backtest
trader backtest \
  --strategy my_strategy.yaml \
  --data data/BTCUSDT_1h.csv \
  --cash 10000
```

### View Results

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BACKTEST RESULT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy       ema_volume_trend
Symbol         BTCUSDT
Timeframe      4h

Period         2024-01-01 → 2024-12-31

Return         +45.32%
CAGR           +45.32%
Sharpe         1.85
Sortino        2.41
Max Drawdown   -12.43%

Trades         48
Win Rate       54.17%
Profit Factor  1.89
Expectancy     +0.62R

Final Equity   $14,532.00
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**See:** [Getting Started Guide](docs/getting-started.md) for complete tutorial.

---

## 📊 Features

### Core Engine

- ✅ **Declarative Strategy DSL** - Define strategies entirely in YAML
- ✅ **Event-Driven Architecture** - Clean separation of concerns
- ✅ **Zero Look-Ahead Bias** - Temporal semantics always explicit
- ✅ **Deterministic Execution** - Reproducible results guaranteed
- ✅ **Extensible Design** - Registry-based plugin architecture

### Technical Indicators

**Built-in Indicators (10+):**
- Trend: SMA, EMA
- Momentum: RSI, MACD, Stochastic
- Volatility: ATR, Bollinger Bands
- Volume: Volume SMA, Volume Ratio
- Custom: Composable indicator pipelines

**See:** [Indicator Reference](docs/indicators.md)

### Data Transforms (NEW! 🎉)

**Preprocessing Pipeline:**
- **Statistical:** Z-score, Percentile rank, Normalization
- **Smoothing:** SMA smooth, EMA smooth
- **Time Series:** Differencing (1st/2nd/3rd order), Log returns
- **Outlier Control:** Clipping, Min-max bounds
- **Scaling:** Custom multipliers

**Example:**
```yaml
data:
  transforms:
    enabled: true
    transforms:
      - type: clip          # Remove outliers
        field: close
        params:
          min: 40000
          max: 60000
      
      - type: ema_smooth    # Smooth noise
        field: close
        params:
          period: 5
      
      - type: zscore        # Normalize
        field: close
        params:
          window: 20
```

**See:** [Transform Guide](docs/transforms.md)

### Execution Models

**Realistic Simulation:**
- ✅ **Commission Models** - Maker/taker fees
- ✅ **5 Slippage Models:**
  - Fixed slippage
  - Percentage-based
  - Volatility-adaptive (NEW! 🎉)
  - Volume-based market impact (NEW! 🎉)
  - No slippage (baseline)
- ✅ **Order Types** - Market, Limit, Stop, Stop-Limit
- ✅ **Fill Simulation** - Realistic order lifecycle

### Risk Management

- ✅ **Position Sizing** - Fixed, percent equity, risk-based
- ✅ **Stop Loss** - Percentage, ATR-based, trailing
- ✅ **Take Profit** - Fixed, risk-reward ratio, multiple targets
- ✅ **Portfolio Limits** - Max exposure, max positions

### Advanced Analysis

- ✅ **Parameter Optimization** - Grid search
- ✅ **Walk-Forward Analysis** - Out-of-sample validation
- ✅ **Monte Carlo Simulation** - Risk assessment
- ✅ **Comprehensive Analytics** - 20+ performance metrics

---

## 📚 Documentation

### User Guides

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | Complete beginner to advanced tutorial |
| [CSV Format Guide](docs/CSV_FORMAT.md) | Flexible CSV formats, troubleshooting, and best practices |
| [Transform Guide](docs/transforms.md) | Data preprocessing and normalization |
| [CLI Reference](docs/cli.md) | Complete command-line reference |
| [Indicator Reference](docs/indicators.md) | All indicators with formulas and examples |
| [Best Practices](docs/best-practices.md) | Strategy development guidelines |

### Architecture Documentation

| Document | Purpose |
|----------|---------|
| [AGENTS.md](AGENTS.md) | **Complete architectural specification** |
| [ROADMAP.md](ROADMAP.md) | Future development phases |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution guidelines |

### Phase Reports

- [Phase 17 Complete](docs/reports/phase17_complete.md) - Data transforms foundation
- [Phase 18 Complete](docs/reports/phase_18_complete.md) - Advanced transforms
- [Phase 19 Session](docs/reports/phase_19_session_complete.md) - Slippage models & docs

---

## 🎛️ CLI Commands

| Command | Description | Example |
|---------|-------------|---------|
| `validate` | Validate strategy configuration | `trader validate --strategy s.yaml` |
| `backtest` | Run backtest on historical data | `trader backtest --strategy s.yaml --data d.csv` |
| `optimize` | Parameter optimization | `trader optimize --strategy s.yaml --param period:5,30,5` |
| `walkforward` | Walk-forward analysis | `trader walkforward --train 2000 --test 500` |
| `montecarlo` | Monte Carlo simulation | `trader montecarlo --result backtest.json` |

**See:** [CLI Reference](docs/cli.md) for complete documentation.

---

## 💡 Strategy Examples

**20+ ready-to-use strategies in `strategies/examples/`:**

### Basic Strategies
- `sma_cross.yaml` - Simple moving average crossover
- `ema_cross.yaml` - Exponential moving average crossover
- `rsi_reversal.yaml` - RSI oversold/overbought reversal

### Advanced Strategies
- `ema_volume.yaml` - EMA with volume confirmation
- `breakout.yaml` - Support/resistance breakout
- `trend_following.yaml` - Multi-indicator trend following
- `atr_stop.yaml` - ATR-based trailing stop

### Transform Examples (NEW!)
- `zscore_mean_reversion.yaml` - Z-score normalization strategy
- `difference_momentum.yaml` - Differencing for momentum
- `ema_smooth_trend.yaml` - Smoothed trend following
- `percentile_strength.yaml` - Percentile-based signals
- `clip_robust.yaml` - Outlier-resistant strategy
- `pipeline_statistical.yaml` - Multi-stage transform pipeline

---

## 🏗️ Architecture

### High-Level Flow

```
YAML Strategy
    ↓
Parser
    ↓
Strategy AST/IR
    ↓
Dependency Resolution
    ↓
Indicator/Expression Evaluation
    ↓
Signal Generation
    ↓
Risk Management
    ↓
Order Execution
    ↓
Portfolio Tracking
    ↓
Analytics & Reporting
```

### Core Design Principles

1. **YAML is an Interface, Not the Engine**
   - Strategy DSL compiles to intermediate representation
   - Engine independent of configuration format

2. **Zero Look-Ahead Bias**
   - Future information never accessible
   - Explicit temporal semantics
   - Regression tests prevent violations

3. **Deterministic Execution**
   - Same inputs → same outputs
   - Reproducible results
   - Verified through testing

4. **Extensibility First**
   - Registry-based architecture
   - Plugin system for custom components
   - No hardcoded strategy logic

5. **Clean Domain Boundaries**
   - Separation of concerns
   - No leaky abstractions
   - Testable components

**See:** [AGENTS.md](AGENTS.md) for complete architectural specification.

---

## 🧪 Testing & Quality

### Test Coverage

- **Unit Tests:** Individual component testing
- **Integration Tests:** Cross-package interactions
- **E2E Tests:** Complete strategy workflows
- **Regression Tests:** Prevent bug reintroduction
- **Look-Ahead Tests:** Prevent future data access

### Quality Metrics

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run specific package
go test ./internal/data/transform

# Run benchmarks
go test -bench=. ./...
```

**Current Status:**
- ✅ 26/26 packages passing
- ✅ 132+ total tests
- ✅ Zero regressions
- ✅ Race condition free

---

## 🎯 Performance Metrics

### Analytics Provided

**Return Metrics:**
- Total Return, CAGR

**Risk-Adjusted:**
- Sharpe Ratio, Sortino Ratio, Calmar Ratio

**Risk Metrics:**
- Maximum Drawdown, Volatility, VaR

**Trading Metrics:**
- Win Rate, Profit Factor, Expectancy
- Average Trade, Average Win, Average Loss
- Trade Count, Exposure Time

**Portfolio Metrics:**
- Final Equity, Total Fees, Net PnL

---

## 🛡️ Research Integrity

The system explicitly distinguishes different testing phases:

| Phase | Purpose | Data |
|-------|---------|------|
| **Backtest** | Initial development | Historical |
| **Optimization** | Parameter tuning | In-sample |
| **Walk-Forward** | Robustness validation | Out-of-sample |
| **Monte Carlo** | Risk assessment | Simulated |
| **Paper Trade** | Real-time simulation | Live (simulated) |
| **Live Trade** | Real execution | Live (real) |

**Critical:** The system never claims profitable backtests prove strategy validity. Emphasis on statistical significance and out-of-sample performance.

---

## 🚀 Roadmap

### Completed Phases ✅

- ✅ **Phase 1-15:** Core MVP (AGENTS.md §1-86)
- ✅ **Phase 16:** Paper Trading
- ✅ **Phase 17:** Transform Foundation
- ✅ **Phase 18:** Advanced Transforms
- ✅ **Phase 19:** Enhanced Backtesting (In Progress)

### Upcoming Phases

- 🔄 **Phase 19:** Advanced execution models (partial fills)
- 📋 **Phase 20:** Portfolio analysis tools
- 📋 **Phase 21:** Machine learning integration
- 📋 **Phase 22:** Exchange integration (live trading)
- 📋 **Phase 23:** Stress testing & robustness
- 📋 **Phase 24:** Cloud deployment
- 📋 **Phase 25:** Community & ecosystem

**See:** [ROADMAP.md](ROADMAP.md) for detailed future development.

---

## 🤝 Contributing

We welcome contributions! Please:

1. Read [AGENTS.md](AGENTS.md) - Architectural requirements
2. Read [CONTRIBUTING.md](CONTRIBUTING.md) - Guidelines
3. Check [ROADMAP.md](ROADMAP.md) - Development priorities

**Critical Rules:**
- Never violate zero look-ahead bias
- Maintain deterministic execution
- Add tests for all features
- Follow existing architecture patterns

---

## 📖 Learning Resources

### For Beginners
1. Start with [Getting Started Guide](docs/getting-started.md)
2. Read example strategies in `strategies/examples/`
3. Run backtests with sample data
4. Experiment with parameters

### For Advanced Users
1. Read [AGENTS.md](AGENTS.md) for architecture
2. Explore [Transform Guide](docs/transforms.md)
3. Study [Best Practices](docs/best-practices.md)
4. Use walk-forward and Monte Carlo analysis

### For Developers
1. Read [AGENTS.md](AGENTS.md) thoroughly
2. Study package structure in `internal/`
3. Review test files for patterns
4. Understand extension points

---

## 🎓 Key Concepts

### Declarative Strategies

Define trading logic **without writing code:**

```yaml
entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [volume, volume_avg * 1.2]
      - gt: [close, ema_200]
```

No Go/Python code required. Pure configuration.

### Transform Pipelines

Preprocess data before strategy execution:

```yaml
transforms:
  - type: clip         # 1. Remove outliers
  - type: ema_smooth   # 2. Smooth noise
  - type: zscore       # 3. Normalize
```

Sequential processing for clean signals.

### Risk-First Design

Risk management is mandatory, not optional:

```yaml
risk:
  position_size:
    type: risk_percent
    value: 0.01        # Risk 1% per trade
  
  stop_loss:
    type: atr
    multiplier: 1.5     # ATR-based stop
```

Every trade has defined risk.

---

## 🌟 Why smallbt_go?

### vs Backtrader (Python)
- ✅ **10-100x faster** (Go vs Python)
- ✅ **Declarative config** (YAML vs hardcoded)
- ✅ **Type safety** (compile-time checks)
- ✅ **Better concurrency** (goroutines vs threads)

### vs QuantConnect
- ✅ **Self-hosted** (no cloud lock-in)
- ✅ **Open source** (MIT license)
- ✅ **Simple deployment** (single binary)
- ✅ **No vendor lock-in**

### vs Zipline
- ✅ **Still maintained** (active development)
- ✅ **Modern architecture** (clean design)
- ✅ **Better docs** (comprehensive guides)
- ✅ **Extensible** (plugin system)

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built according to the comprehensive architectural specification in [AGENTS.md](AGENTS.md), which guided development through 19 phases to a production-ready, well-documented system.

Special thanks to the quantitative trading community for inspiration and best practices.

---

## 📞 Support

- 📖 **Documentation:** [docs/](docs/)
- 🐛 **Issues:** [GitHub Issues](https://github.com/ZulferDev/smallbt_go/issues)
- 💬 **Discussions:** [GitHub Discussions](https://github.com/ZulferDev/smallbt_go/discussions)

---

**Ready for quantitative trading research that prioritizes correctness over convenience.**

**Built with ❤️ for the quant community.**

---
