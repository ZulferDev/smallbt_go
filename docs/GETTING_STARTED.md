# Getting Started with Declarative Backtest Engine

## Quick Start

### 1. Installation
```bash
git clone https://github.com/1jehuang/backtest
cd backtest
go build ./cmd/trader
```

### 2. Create Your First Strategy
Create `my_strategy.yaml`:

```yaml
strategy:
  name: sma_crossover
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  sma_fast:
    type: sma
    source: close
    period: 9
  
  sma_slow:
    type: sma
    source: close
    period: 21

entry:
  long:
    all:
      - cross_above: [sma_fast, sma_slow]

exit:
  long:
    any:
      - cross_below: [sma_fast, sma_slow]

risk:
  position_size:
    type: percent_equity
    value: 0.05
  
  stop_loss:
    type: fixed
    value: 0.95  # 5% stop loss
```

### 3. Validate Strategy
```bash
./trader validate --strategy my_strategy.yaml
```

### 4. Run Backtest
```bash
./trader backtest \
  --strategy my_strategy.yaml \
  --data data/BTCUSDT_1h_sample.csv \
  --output results.json
```

### 5. View Results
```bash
cat results.json | jq .
```

## Core Concepts

### Strategy DSL
Strategies are defined in YAML with:
- **Indicators**: Technical indicators like SMA, EMA, RSI, ATR
- **Expressions**: Arithmetic, comparison, and logical operations  
- **Conditions**: Entry and exit rules
- **Risk Management**: Position sizing, stop loss, take profit

### Architecture
The engine follows clean architecture:
- **YAML → AST → Runtime** pipeline
- **Registry-based extensibility**
- **Deterministic backtesting**
- **No look-ahead bias**

### Key Features
- ✅ YAML-defined strategies (no Go code required)
- ✅ Realistic execution (fees, slippage, order types)
- ✅ Risk management (position sizing, stops, trailing stops)
- ✅ Parameter optimization (grid search)
- ✅ Walk Forward Analysis
- ✅ Monte Carlo simulation
- ✅ Multi-timeframe support

## Example Strategies
Check `strategies/examples/` for ready-to-use strategies:
- `ema_volume.yaml`: EMA crossover with volume confirmation
- `stateful_breakout.yaml`: Stateful breakout strategy
- `mtf_test.yaml`: Multi-timeframe example
- `both_directions_test.yaml`: Long/short strategy

## Next Steps
1. Explore the example strategies
2. Modify parameters and test
3. Use optimization to find best parameters
4. Validate robustness with Walk Forward Analysis
5. Analyze risk with Monte Carlo simulation