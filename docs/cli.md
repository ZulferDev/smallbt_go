# CLI Reference

**Command-line interface for smallbt_go**

---

## Table of Contents

1. [Overview](#overview)
2. [Global Options](#global-options)
3. [Commands](#commands)
4. [Configuration](#configuration)
5. [Output Formats](#output-formats)
6. [Examples](#examples)

---

## Overview

The `trader` CLI is the primary interface for:

- Running backtests
- Validating strategies
- Optimizing parameters
- Walk-forward analysis
- Monte Carlo simulation

### Basic Usage

```bash
trader [command] [options]
```

### Getting Help

```bash
# General help
trader --help

# Command-specific help
trader backtest --help
trader optimize --help
```

---

## Global Options

Available for all commands:

### `--verbose` / `-v`

Enable verbose logging for debugging.

```bash
trader backtest --strategy my_strategy.yaml --data data.csv --verbose
```

**Output:**
- Detailed execution logs
- Transform values
- Indicator calculations
- Signal generation
- Order execution details

### `--quiet` / `-q`

Suppress all output except errors.

```bash
trader backtest --strategy my_strategy.yaml --data data.csv --quiet
```

### `--version`

Display version information.

```bash
trader --version
```

**Output:**
```
smallbt_go version 0.1.0
Built with Go 1.21.0
```

---

## Commands

### `backtest`

Run a strategy backtest on historical data.

**Synopsis:**
```bash
trader backtest [options]
```

**Required Options:**

| Option | Description | Example |
|--------|-------------|---------|
| `--strategy` / `-s` | Strategy YAML file | `--strategy my_strategy.yaml` |
| `--data` / `-d` | Historical data file (CSV/Parquet) | `--data BTCUSDT.csv` |

**Optional Options:**

| Option | Description | Default | Example |
|--------|-------------|---------|---------|
| `--cash` | Initial cash balance | 10000 | `--cash 50000` |
| `--commission` | Trading commission (%) | 0.001 | `--commission 0.002` |
| `--slippage` | Price slippage (%) | 0.0 | `--slippage 0.001` |
| `--output` / `-o` | Output file for results | stdout | `--output result.json` |
| `--format` | Output format (json/csv/text) | text | `--format json` |

**Examples:**

```bash
# Basic backtest
trader backtest \
  --strategy strategies/sma_cross.yaml \
  --data data/BTCUSDT_1h.csv

# With custom initial cash
trader backtest \
  --strategy strategies/ema_cross.yaml \
  --data data/ETHUSDT_4h.csv \
  --cash 100000

# With realistic fees and slippage
trader backtest \
  --strategy strategies/rsi_reversal.yaml \
  --data data/BTCUSDT_1h.csv \
  --commission 0.001 \
  --slippage 0.0005

# Save results to JSON
trader backtest \
  --strategy strategies/breakout.yaml \
  --data data/BTCUSDT_1h.csv \
  --output results/backtest_001.json \
  --format json
```

**Output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BACKTEST RESULT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy       sma_cross
Symbol         BTCUSDT
Timeframe      1h
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

---

### `validate`

Validate strategy configuration without running backtest.

**Synopsis:**
```bash
trader validate [options]
```

**Required Options:**

| Option | Description | Example |
|--------|-------------|---------|
| `--strategy` / `-s` | Strategy YAML file | `--strategy my_strategy.yaml` |

**Examples:**

```bash
# Validate strategy
trader validate --strategy strategies/new_strategy.yaml
```

**Success Output:**
```
✓ Strategy validation passed
  - Indicators: 4 defined, 0 errors
  - Entry conditions: valid
  - Exit conditions: valid
  - Risk management: valid
  - No circular dependencies
```

**Error Output:**
```
✗ Strategy validation failed

Error in indicators.ema_fast.period:
  expected positive integer, got -5

Error in entry.long:
  unknown condition: cross_over
  did you mean: cross_above?
```

---

### `optimize`

Optimize strategy parameters using grid search.

**Synopsis:**
```bash
trader optimize [options]
```

**Required Options:**

| Option | Description | Example |
|--------|-------------|---------|
| `--strategy` / `-s` | Strategy YAML file | `--strategy my_strategy.yaml` |
| `--data` / `-d` | Historical data file | `--data BTCUSDT.csv` |
| `--param` / `-p` | Parameter to optimize | `--param sma_fast.period:5,50,5` |

**Optional Options:**

| Option | Description | Default | Example |
|--------|-------------|---------|---------|
| `--objective` | Optimization objective | sharpe | `--objective return` |
| `--cash` | Initial cash | 10000 | `--cash 50000` |
| `--output` / `-o` | Output file | stdout | `--output opt_results.json` |

**Parameter Format:**

```
--param <path>:<start>,<end>,<step>
```

Examples:
```bash
--param indicators.sma_fast.period:5,50,5
--param risk.position_size.value:0.1,1.0,0.1
--param indicators.rsi.period:10,20,2
```

**Multiple Parameters:**

```bash
trader optimize \
  --strategy strategies/sma_cross.yaml \
  --data data/BTCUSDT_1h.csv \
  --param indicators.sma_fast.period:5,30,5 \
  --param indicators.sma_slow.period:20,100,10 \
  --objective sharpe
```

**Objectives:**

| Objective | Description | Best For |
|-----------|-------------|----------|
| `return` | Total return | Profit maximization |
| `sharpe` | Sharpe ratio | Risk-adjusted returns |
| `sortino` | Sortino ratio | Downside risk focus |
| `winrate` | Win rate | Trade accuracy |
| `profit_factor` | Profit factor | Profit/loss ratio |
| `expectancy` | Expectancy | Average profit per trade |

**Output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
OPTIMIZATION RESULTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Objective      Sharpe Ratio
Combinations   120
Completed      120 (100%)

Best Parameters:
  sma_fast.period: 15
  sma_slow.period: 50

Best Performance:
  Sharpe:    2.34
  Return:    +68.5%
  Drawdown:  -15.2%
  Trades:    62
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Top 5 Results:
1. sma_fast=15, sma_slow=50  → Sharpe: 2.34
2. sma_fast=10, sma_slow=45  → Sharpe: 2.21
3. sma_fast=20, sma_slow=55  → Sharpe: 2.18
4. sma_fast=15, sma_slow=55  → Sharpe: 2.15
5. sma_fast=10, sma_slow=50  → Sharpe: 2.12
```

---

### `walkforward`

Run walk-forward analysis to test strategy robustness.

**Synopsis:**
```bash
trader walkforward [options]
```

**Required Options:**

| Option | Description | Example |
|--------|-------------|---------|
| `--strategy` / `-s` | Strategy YAML file | `--strategy my_strategy.yaml` |
| `--data` / `-d` | Historical data file | `--data BTCUSDT.csv` |
| `--train` | Training window size (candles) | `--train 1000` |
| `--test` | Test window size (candles) | `--test 200` |
| `--step` | Step size between windows | `--step 100` |

**Optional Options:**

| Option | Description | Default | Example |
|--------|-------------|---------|---------|
| `--param` / `-p` | Parameters to optimize | none | `--param sma_fast.period:5,30,5` |
| `--objective` | Optimization objective | sharpe | `--objective return` |
| `--cash` | Initial cash | 10000 | `--cash 50000` |
| `--output` / `-o` | Output file | stdout | `--output wf_results.json` |

**Example:**

```bash
trader walkforward \
  --strategy strategies/sma_cross.yaml \
  --data data/BTCUSDT_1h.csv \
  --train 2000 \
  --test 500 \
  --step 250 \
  --param indicators.sma_fast.period:5,30,5 \
  --param indicators.sma_slow.period:20,100,10
```

**Output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WALK-FORWARD ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Windows        5
Train Size     2000 candles
Test Size      500 candles
Step Size      250 candles

Out-of-Sample Performance:
  Return:        +32.5%
  Sharpe:        1.45
  Max Drawdown:  -18.3%
  Win Rate:      48.2%
  Trades:        156

Window Results:
1. Train: +45.2% | Test: +28.1% | Degradation: -37.5%
2. Train: +52.1% | Test: +35.6% | Degradation: -31.7%
3. Train: +48.8% | Test: +31.2% | Degradation: -36.1%
4. Train: +51.3% | Test: +33.8% | Degradation: -34.1%
5. Train: +49.6% | Test: +33.9% | Degradation: -31.7%

Avg Degradation: -34.2%
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### `montecarlo`

Run Monte Carlo analysis on backtest results.

**Synopsis:**
```bash
trader montecarlo [options]
```

**Required Options:**

| Option | Description | Example |
|--------|-------------|---------|
| `--result` / `-r` | Backtest result JSON | `--result backtest.json` |

**Optional Options:**

| Option | Description | Default | Example |
|--------|-------------|---------|---------|
| `--simulations` / `-n` | Number of simulations | 10000 | `--simulations 50000` |
| `--confidence` | Confidence level (%) | 95 | `--confidence 99` |
| `--output` / `-o` | Output file | stdout | `--output mc_results.json` |

**Example:**

```bash
# First run backtest
trader backtest \
  --strategy strategies/sma_cross.yaml \
  --data data/BTCUSDT_1h.csv \
  --output results/backtest.json \
  --format json

# Then run Monte Carlo
trader montecarlo \
  --result results/backtest.json \
  --simulations 10000 \
  --confidence 95
```

**Output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MONTE CARLO ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Simulations    10,000
Confidence     95%

Return Distribution:
  Mean:          +42.3%
  Median:        +40.1%
  Std Dev:       18.5%
  
  95% CI:        [+8.2%, +78.4%]
  Probability > 0%:  89.2%
  Probability < -20%: 2.3%

Max Drawdown Distribution:
  Mean:          -15.8%
  Median:        -14.2%
  Worst 5%:      -28.6%
  
  Probability > -20%: 92.1%
  Probability > -30%: 98.7%

Sharpe Ratio Distribution:
  Mean:          1.78
  Median:        1.82
  95% CI:        [0.85, 2.64]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Configuration

### Environment Variables

Configure defaults via environment:

```bash
# Default initial cash
export SMALLBT_CASH=10000

# Default commission
export SMALLBT_COMMISSION=0.001

# Default slippage
export SMALLBT_SLIPPAGE=0.0005

# Data directory
export SMALLBT_DATA_DIR=/path/to/data

# Strategies directory
export SMALLBT_STRATEGY_DIR=/path/to/strategies
```

### Config File

Create `~/.smallbt/config.yaml`:

```yaml
defaults:
  cash: 10000
  commission: 0.001
  slippage: 0.0005

paths:
  data: ~/trading/data
  strategies: ~/trading/strategies
  results: ~/trading/results

output:
  format: json
  verbose: false
```

---

## Output Formats

### Text (Human-Readable)

Default format for terminal display.

```bash
trader backtest --strategy s.yaml --data d.csv --format text
```

### JSON (Machine-Readable)

Structured output for further processing.

```bash
trader backtest --strategy s.yaml --data d.csv --format json
```

**JSON Structure:**
```json
{
  "strategy": {
    "name": "sma_cross",
    "version": "1"
  },
  "period": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  },
  "performance": {
    "return": 0.4532,
    "cagr": 0.4532,
    "sharpe": 1.85,
    "sortino": 2.41,
    "max_drawdown": -0.1243
  },
  "trades": {
    "total": 48,
    "wins": 26,
    "losses": 22,
    "win_rate": 0.5417,
    "profit_factor": 1.89,
    "expectancy": 0.62
  },
  "portfolio": {
    "initial_cash": 10000,
    "final_equity": 14532,
    "total_fees": 145.32
  }
}
```

### CSV (Tabular)

For spreadsheet analysis.

```bash
trader backtest --strategy s.yaml --data d.csv --format csv
```

**CSV Output:**
```csv
metric,value
return,0.4532
cagr,0.4532
sharpe,1.85
sortino,2.41
max_drawdown,-0.1243
trades,48
win_rate,0.5417
profit_factor,1.89
expectancy,0.62
final_equity,14532
```

---

## Examples

### Example 1: Quick Backtest

```bash
trader backtest \
  --strategy strategies/examples/sma_cross.yaml \
  --data data/BTCUSDT_1h.csv
```

### Example 2: Realistic Trading Conditions

```bash
trader backtest \
  --strategy strategies/my_strategy.yaml \
  --data data/ETHUSDT_4h.csv \
  --cash 50000 \
  --commission 0.001 \
  --slippage 0.0005
```

### Example 3: Parameter Optimization

```bash
trader optimize \
  --strategy strategies/rsi_reversal.yaml \
  --data data/BTCUSDT_1h.csv \
  --param indicators.rsi.period:10,20,2 \
  --param entry.long.conditions.0.value:20,40,5 \
  --objective sharpe \
  --output results/optimization.json
```

### Example 4: Walk-Forward Validation

```bash
trader walkforward \
  --strategy strategies/ema_cross.yaml \
  --data data/BTCUSDT_1h.csv \
  --train 2000 \
  --test 500 \
  --step 250 \
  --param indicators.ema_fast.period:5,30,5 \
  --param indicators.ema_slow.period:20,100,10
```

### Example 5: Monte Carlo Risk Assessment

```bash
# Step 1: Backtest
trader backtest \
  --strategy strategies/breakout.yaml \
  --data data/BTCUSDT_1h.csv \
  --output results/backtest.json \
  --format json

# Step 2: Monte Carlo
trader montecarlo \
  --result results/backtest.json \
  --simulations 50000 \
  --confidence 99 \
  --output results/montecarlo.json
```

### Example 6: Batch Processing

```bash
#!/bin/bash
# backtest_all.sh - Backtest multiple strategies

for strategy in strategies/*.yaml; do
  echo "Testing $strategy..."
  trader backtest \
    --strategy "$strategy" \
    --data data/BTCUSDT_1h.csv \
    --output "results/$(basename $strategy .yaml).json" \
    --format json
done
```

### Example 7: Compare Strategies

```bash
# Strategy A
trader backtest -s strategies/sma.yaml -d data.csv -o results/sma.json -f json

# Strategy B
trader backtest -s strategies/ema.yaml -d data.csv -o results/ema.json -f json

# Strategy C
trader backtest -s strategies/rsi.yaml -d data.csv -o results/rsi.json -f json

# Compare
python scripts/compare_results.py results/*.json
```

---

## Tips & Tricks

### 1. Use Aliases

```bash
# ~/.bashrc
alias bt='trader backtest'
alias opt='trader optimize'
alias wf='trader walkforward'
alias mc='trader montecarlo'

# Usage
bt -s my_strategy.yaml -d data.csv
```

### 2. JSON Processing with jq

```bash
# Extract Sharpe ratio
trader backtest -s s.yaml -d d.csv -f json | jq '.performance.sharpe'

# Filter profitable results
trader optimize -s s.yaml -d d.csv -f json | \
  jq '.results[] | select(.return > 0)'
```

### 3. Parallel Execution

```bash
# Backtest multiple symbols in parallel
parallel trader backtest -s strategy.yaml -d data/{}.csv -o results/{}.json ::: \
  BTCUSDT ETHUSDT BNBUSDT ADAUSDT
```

### 4. Progress Monitoring

```bash
# For long optimizations
trader optimize -s s.yaml -d d.csv -p param:1,100,1 --verbose 2>&1 | \
  tee optimization.log
```

---

## Further Reading

- [Getting Started Guide](getting-started.md)
- [Transform Guide](transforms.md)
- [Indicator Reference](indicators.md)
- [Strategy Examples](../../strategies/examples/)

---

**Master the CLI for efficient backtesting! ⚙️**
