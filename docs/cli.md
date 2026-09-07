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
| `--commission` | Trading commission rate | 0.001 | `--commission 0.002` |
| `--slippage-model` | Slippage model type | none | `--slippage-model volatility` |
| `--slippage` | Slippage value (model-specific) | 0.0 | `--slippage 0.001` |
| `--slippage-max` | Max slippage for adaptive models | 0.01 | `--slippage-max 0.02` |
| `--output` / `-o` | Output file for results | stdout | `--output result.json` |
| `--format` | Output format (json/csv/text) | text | `--format json` |

**Slippage Models:**

| Model | Description | Use Case |
|-------|-------------|----------|
| `none` | No slippage (perfect execution) | Baseline testing |
| `fixed` | Fixed amount per trade | Conservative estimate |
| `percentage` | Percentage of fill price | Proportional impact |
| `volatility` | Adaptive to market volatility | Realistic simulation |
| `volume` | Based on order/volume ratio | Market impact modeling |

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

# With realistic fees and percentage slippage
trader backtest \
  --strategy strategies/rsi_reversal.yaml \
  --data data/BTCUSDT_1h.csv \
  --commission 0.001 \
  --slippage-model percentage \
  --slippage 0.0005

# With volatility-based slippage (adaptive)
trader backtest \
  --strategy strategies/trend_following.yaml \
  --data data/BTCUSDT_1h.csv \
  --commission 0.001 \
  --slippage-model volatility \
  --slippage 0.001 \
  --slippage-max 0.01

# With volume-based slippage (market impact)
trader backtest \
  --strategy strategies/breakout.yaml \
  --data data/BTCUSDT_1h.csv \
  --commission 0.002 \
  --slippage-model volume \
  --slippage 0.0001 \
  --slippage-max 0.005

# No slippage for baseline comparison
trader backtest \
  --strategy strategies/ema_cross.yaml \
  --data data/BTCUSDT_1h.csv \
  --commission 0.001 \
  --slippage-model none

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
| `--output` / `-o` | Output JSON file | none | `--output wf_results.json` |
| `--csv-windows` | Export window results to CSV | none | `--csv-windows windows.csv` |
| `--csv-aggregate` | Export aggregate metrics to CSV | none | `--csv-aggregate aggregate.csv` |
| `--csv-stability` | Export stability analysis to CSV | none | `--csv-stability stability.csv` |

**Examples:**

**Basic Walk Forward Analysis:**
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

**Export Detailed CSV Reports:**
```bash
trader walkforward \
  --strategy strategies/ema_volume.yaml \
  --data data/ETHUSDT_4h.csv \
  --train 1000 \
  --test 200 \
  --step 200 \
  --csv-windows wf_windows.csv \
  --csv-aggregate wf_aggregate.csv \
  --csv-stability wf_stability.csv
```

**CSV Output Formats:**

**Window Results (`--csv-windows`):**
Contains performance for each walk forward window:
- WindowID, TrainStart, TrainEnd, TestStart, TestEnd
- TrainReturn_%, TrainSharpe, TrainTrades
- TestReturn_%, TestSharpe, TestMaxDrawdown_%, TestWinRate_%, TestProfitFactor, TestTrades
- InSampleOutSampleDelta_%, PerformanceDegradation

**Aggregate Results (`--csv-aggregate`):**
Contains overall out-of-sample performance:
- TotalWindows, TotalTrades, TotalReturn_%, CAGR_%
- SharpeRatio, SortinoRatio, MaxDrawdown_%, CalmarRatio
- WinRate_%, ProfitFactor, Expectancy
- AverageWin, AverageLoss, AvgTradeReturn_%

**Stability Analysis (`--csv-stability`):**
Contains consistency metrics across windows:
- TotalWindows, ProfitableWindows, UnprofitableWindows, ProfitableRate_%
- ConsistencyScore (0-100), AvgReturnStdDev, AvgSharpeStdDev
- BestWindow, BestWindowReturn_%, WorstWindow, WorstWindowReturn_%
- ImprovedWindows, DegradationLow, DegradationModerate, DegradationHigh

**Performance Degradation Classification:**
- **Improved**: Test performance better than training
- **Low**: 0-5% degradation from training to test
- **Moderate**: 5-10% degradation
- **High**: >10% degradation

**Consistency Score Interpretation:**
- **80-100**: Excellent - highly stable across windows
- **60-80**: Good - acceptable consistency
- **40-60**: Fair - moderate instability
- **0-40**: Poor - highly unstable, likely overfit

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
| `--simulations` / `-n` | Number of simulations | 1000 | `--simulations 10000` |
| `--seed` | Random seed for reproducibility | 42 | `--seed 123` |
| `--output` / `-o` | Output JSON file | none | `--output mc_results.json` |
| `--csv-stats` | Export statistics to CSV | none | `--csv-stats mc_stats.csv` |
| `--csv-percentiles` | Export percentiles to CSV | none | `--csv-percentiles percentiles.csv` |
| `--csv-simulations` | Export all simulations to CSV | none | `--csv-simulations sims.csv` |
| `--csv-drawdown` | Export drawdown distribution to CSV | none | `--csv-drawdown dd_dist.csv` |
| `--csv-risk` | Export risk analysis to CSV | none | `--csv-risk risk.csv` |

**Examples:**

**Basic Monte Carlo Analysis:**
```bash
# First run backtest
trader backtest \
  --strategy strategies/sma_cross.yaml \
  --data data/BTCUSDT_1h.csv \
  --output results/backtest.json

# Then run Monte Carlo
trader montecarlo \
  --result results/backtest.json \
  --simulations 10000 \
  --seed 42
```

**Export Detailed CSV Reports:**
```bash
trader montecarlo \
  --result results/backtest.json \
  --simulations 10000 \
  --csv-stats mc_statistics.csv \
  --csv-percentiles mc_percentiles.csv \
  --csv-simulations mc_all_sims.csv \
  --csv-drawdown mc_drawdown.csv \
  --csv-risk mc_risk.csv
```

**CSV Output Formats:**

**Statistics (`--csv-stats`):**
Contains aggregated Monte Carlo statistics:
- Return distribution: Mean, Median, StdDev, Min, Max, P05, P95
- Drawdown distribution: Mean, Median, StdDev, P95
- Win rate distribution: Mean, StdDev, Min, Max
- Sharpe ratio distribution: Mean, StdDev, Min, Max
- Risk metrics: ProbabilityOfRuin, NegativeReturnRatio

**Percentiles (`--csv-percentiles`):**
Contains key percentiles (5th, 25th, 50th, 75th, 95th):
- Percentile, TotalReturn_%, MaxDrawdown_%, WinRate_%, SharpeRatio

**Simulations (`--csv-simulations`):**
Contains all individual simulation results:
- SimulationID, TotalReturn_%, MaxDrawdown_%, TotalTrades
- WinningTrades, LosingTrades, WinRate_%, TotalPnL, SharpeRatio

**Drawdown Distribution (`--csv-drawdown`):**
Contains percentile-based drawdown analysis with interpretations:
- Percentile, MaxDrawdown_%, Interpretation
- Best case (5%), Median (50%), Worst case (95%)

**Risk Analysis (`--csv-risk`):**
Contains comprehensive risk assessment:
- ExpectedReturn_%, WorstCase5Pct_%, BestCase95Pct_%
- ProbabilityProfit_%, ProbabilityLoss_%
- WorstDrawdown95Pct_%, RiskOfRuin_%
- ConsistencyScore with interpretation

**Risk Metrics Interpretation:**

**Probability of Profit:**
- >80%: Very high confidence
- 60-80%: High confidence
- 50-60%: Slight edge
- <50%: Low confidence - high risk

**Risk of Ruin:**
- <1%: Very low risk
- 1-5%: Low risk
- 5-10%: Moderate risk
- >10%: High risk - unacceptable

**Consistency Score (0-100):**
- 80-100: Excellent - very consistent returns
- 60-80: Good - acceptable consistency
- 40-60: Fair - moderate variance
- 0-40: Poor - high variance, unstable

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

---

## Trade Analysis Commands

### export-trades

Export trade history from backtest results to CSV format with comprehensive details.

**Usage:**

```bash
trader export-trades --result <result.json> --output <trades.csv>
```

**Flags:**

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--result` | string | Yes | Path to backtest result JSON file |
| `--output` | string | No | Output CSV file path (default: trades.csv) |

**CSV Columns:**

The exported CSV includes 19 columns:

| Column | Description |
|--------|-------------|
| ID | Unique trade identifier |
| Symbol | Trading symbol |
| Side | Position side (long/short) |
| EntryTime | Entry timestamp (RFC3339) |
| EntryPrice | Entry price |
| ExitTime | Exit timestamp (RFC3339) |
| ExitPrice | Exit price |
| Quantity | Position size |
| Duration_Minutes | Holding time in minutes |
| GrossPnL | Gross profit/loss |
| Fees | Total fees paid |
| NetPnL | Net profit/loss after fees |
| Return_% | Return percentage |
| MAE | Maximum Adverse Excursion (absolute) |
| MFE | Maximum Favorable Excursion (absolute) |
| MAE_% | MAE as percentage of entry price |
| MFE_% | MFE as percentage of entry price |
| ExitReason | Reason for exit (take_profit, stop_loss, etc.) |
| PriceChange_% | Price change percentage |

**Example:**

```bash
# Export trades to CSV
trader export-trades \
  --result backtest_result.json \
  --output my_trades.csv

# View in spreadsheet
libreoffice my_trades.csv
```

---

### analyze-trades

Analyze trade statistics and generate a comprehensive report.

**Usage:**

```bash
trader analyze-trades --result <result.json> [--output <report.txt>]
```

**Flags:**

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--result` | string | Yes | Path to backtest result JSON file |
| `--output` | string | No | Optional file to save analysis report |

**Analysis Includes:**

1. **Trade Breakdown**
   - Total trades
   - Winning trades (count + percentage)
   - Losing trades (count + percentage)
   - Breakeven trades (count + percentage)

2. **Streak Analysis**
   - Maximum winning streak
   - Maximum losing streak
   - Current streak status

3. **Best/Worst Trades**
   - Largest win (dollar amount)
   - Largest loss (dollar amount)
   - Best and worst trade details

4. **Holding Time Statistics**
   - Average holding time
   - Minimum holding time
   - Maximum holding time

5. **MAE/MFE Analysis**
   - Average Maximum Adverse Excursion (%)
   - Average Maximum Favorable Excursion (%)

6. **Exit Reason Distribution**
   - Breakdown by exit reason
   - Count and percentage for each

**Example Output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TRADE ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Trades:      25
├─ Winning:        15 (60.0%)
├─ Losing:         9 (36.0%)
└─ Breakeven:      1 (4.0%)

Streaks:
├─ Max Win Streak:   4
└─ Max Loss Streak:  2

Best Trade:        $450.25
Worst Trade:       $-125.50

Holding Time:
├─ Average:        12.5h
├─ Minimum:        2.3h
└─ Maximum:        2.1d

MAE/MFE Analysis:
├─ Avg MAE:        -0.45%
└─ Avg MFE:        1.23%

Exit Reasons:
├─ take_profit: 12 (48.0%)
├─ stop_loss: 10 (40.0%)
├─ manual: 3 (12.0%)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Examples:**

```bash
# Print analysis to console
trader analyze-trades --result backtest_result.json

# Save analysis to file
trader analyze-trades \
  --result backtest_result.json \
  --output analysis_report.txt

# Use with multiple results
for result in results/*.json; do
  trader analyze-trades --result "$result"
done
```

---

## Complete Workflow Example

Here's a complete workflow using trade analysis:

```bash
# 1. Run backtest
trader backtest \
  --strategy my_strategy.yaml \
  --data historical_data.csv \
  --output results/backtest_2024.json

# 2. Analyze trade statistics
trader analyze-trades \
  --result results/backtest_2024.json \
  --output results/analysis_2024.txt

# 3. Export trades for detailed review
trader export-trades \
  --result results/backtest_2024.json \
  --output results/trades_2024.csv

# 4. Review in spreadsheet
libreoffice results/trades_2024.csv

# 5. Compare with previous results
trader analyze-trades --result results/backtest_2023.json
trader analyze-trades --result results/backtest_2024.json
```

---

## Trade Analysis Use Cases

### 1. Performance Attribution

Identify which exit reasons lead to best performance:

```bash
trader analyze-trades --result result.json | grep "Exit Reasons" -A 5
```

### 2. Trade Duration Analysis

Understand optimal holding periods:

```bash
trader export-trades --result result.json --output trades.csv
# Analyze Duration_Minutes column in spreadsheet
```

### 3. MAE/MFE Optimization

Optimize stop-loss and take-profit levels:

```bash
# Export trades
trader export-trades --result result.json --output trades.csv

# Analyze MAE_% and MFE_% columns to:
# - See how far price typically moves against you (MAE)
# - See peak favorable movement before exit (MFE)
# - Adjust stop-loss if MAE is consistently hit
# - Adjust take-profit if leaving money on table (MFE much larger)
```

### 4. Win Streak Analysis

Monitor consecutive wins/losses for risk management:

```bash
trader analyze-trades --result result.json | grep "Streaks" -A 2
```

### 5. Research Workflow

Document your trading research:

```bash
# Create research folder
mkdir -p research/strategy_v1

# Run backtest
trader backtest \
  --strategy strategy_v1.yaml \
  --data data.csv \
  --output research/strategy_v1/result.json

# Generate analysis
trader analyze-trades \
  --result research/strategy_v1/result.json \
  --output research/strategy_v1/analysis.txt

# Export detailed trades
trader export-trades \
  --result research/strategy_v1/result.json \
  --output research/strategy_v1/trades.csv

# Now you have complete documentation for this version
```


---

## Optimization CSV Export

### Export All Results

Export complete optimization results to CSV format:

```bash
trader optimize \
  --strategy strategy.yaml \
  --data historical.csv \
  --parameters "indicators.ema_fast.period:5:20:1,indicators.ema_slow.period:20:100:5" \
  --objective sharpe \
  --csv optimization_results.csv
```

**CSV Columns:**

The optimization CSV includes:
- Rank
- ObjectiveValue
- All parameter values
- TotalReturn_%
- CAGR_%
- SharpeRatio
- SortinoRatio
- MaxDrawdown_%
- WinRate_%
- ProfitFactor
- Expectancy
- TotalTrades
- WinningTrades
- LosingTrades
- AvgWin
- AvgLoss

---

### Export Top N Results

Export only the best N parameter combinations:

```bash
trader optimize \
  --strategy strategy.yaml \
  --data historical.csv \
  --parameters "indicators.ema_fast.period:5:20:1" \
  --objective sharpe \
  --csv top_results.csv \
  --top 10
```

This exports only the top 10 best parameter combinations, making it easier to focus on the most promising configurations.

---

### Parameter Sensitivity Analysis

Analyze how each parameter affects the optimization objective:

```bash
trader optimize \
  --strategy strategy.yaml \
  --data historical.csv \
  --parameters "indicators.ema_fast.period:5:20:1,indicators.ema_slow.period:20:100:5" \
  --objective sharpe \
  --sensitivity sensitivity_analysis.csv
```

**Sensitivity CSV Columns:**

- Parameter: Parameter name
- MinValue: Minimum value tested
- MaxValue: Maximum value tested
- Range: Value range (max - min)
- BestValue: Parameter value that produced best objective
- WorstValue: Parameter value that produced worst objective
- AvgObjective: Average objective value across all combinations
- StdDevObjective: Standard deviation of objective
- Correlation: Correlation coefficient with objective (-1 to 1)
- Sensitivity: Classification (Low/Medium/High)

**Interpreting Sensitivity:**

- **High Correlation** (|r| > 0.7): Strong relationship with performance
  - Positive correlation: Higher values → better performance
  - Negative correlation: Lower values → better performance

- **High Sensitivity**: Large impact on results
  - Small changes cause large performance swings
  - Requires careful tuning

- **Low Sensitivity**: Minimal impact on results
  - May not need optimization
  - Stable across value range

---

### Complete Optimization Workflow

Export all analysis formats in one command:

```bash
trader optimize \
  --strategy my_strategy.yaml \
  --data BTCUSDT_5000h.csv \
  --parameters "indicators.ema_fast.period:5:20:1,indicators.ema_slow.period:20:100:5" \
  --objective sharpe \
  --parallel 4 \
  --output optimization_full.json \
  --csv optimization_all.csv \
  --top 20 \
  --sensitivity parameter_analysis.csv
```

**Note:** When using `--top N` with `--csv`, the CSV will contain only top N results. For all results, use a separate `--csv` without `--top`.

To export both:
```bash
# Run optimization once, save JSON
trader optimize ... --output results.json

# Then export different formats
# (Feature coming soon: separate export command)
```

---

## Optimization Analysis Examples

### Example 1: Find Optimal EMA Periods

```bash
trader optimize \
  --strategy strategies/ema_cross.yaml \
  --data data/BTCUSDT.csv \
  --parameters "indicators.ema_fast.period:5:30:1,indicators.ema_slow.period:20:200:5" \
  --objective sharpe \
  --parallel 8 \
  --csv ema_optimization.csv \
  --sensitivity ema_sensitivity.csv
```

**Analysis Steps:**
1. Open `ema_sensitivity.csv` in spreadsheet
2. Check correlation for each parameter
3. If `ema_fast` has high correlation but `ema_slow` has low correlation:
   - `ema_fast` needs careful tuning
   - `ema_slow` can use a fixed value
4. Open `ema_optimization.csv`
5. Sort by SharpeRatio descending
6. Review top 10 parameter combinations
7. Check consistency of best parameters

---

### Example 2: Detect Overfitting

**Signs of Overfitting in Sensitivity Analysis:**

1. **Very High Correlation** (|r| > 0.95):
   - May indicate curve-fitting
   - Test on out-of-sample data

2. **Very High Sensitivity**:
   - Small parameter changes cause huge swings
   - Strategy may be unstable

3. **BestValue at Extreme Edges**:
   - If best value is at min or max of range
   - Expand the search range
   - May not have found true optimum

**Example Check:**
```bash
# Run optimization
trader optimize ... --sensitivity sensitivity.csv

# Open sensitivity.csv
# Look for:
# - Correlations very close to ±1.0
# - BestValue = MinValue or MaxValue
# - Sensitivity = "High" for all parameters

# If found, be cautious of overfitting
```

---

### Example 3: Multi-Objective Analysis

While the optimizer uses a single objective, you can analyze trade-offs in the CSV:

```bash
# Optimize for Sharpe
trader optimize \
  --strategy strategy.yaml \
  --data data.csv \
  --parameters "..." \
  --objective sharpe \
  --csv results.csv

# Open results.csv in spreadsheet
# Create scatter plots:
# - SharpeRatio vs MaxDrawdown
# - TotalReturn vs WinRate
# - SharpeRatio vs TotalTrades

# Look for parameter combinations that balance multiple goals
```

---

### Example 4: Parameter Stability Test

Test if parameters are stable across different market conditions:

```bash
# Optimize on 2023 data
trader optimize \
  --strategy strategy.yaml \
  --data data_2023.csv \
  --parameters "..." \
  --objective sharpe \
  --csv results_2023.csv \
  --top 10

# Optimize on 2024 data
trader optimize \
  --strategy strategy.yaml \
  --data data_2024.csv \
  --parameters "..." \
  --objective sharpe \
  --csv results_2024.csv \
  --top 10

# Compare top parameters from both years
# Stable parameters should appear in both top 10 lists
```

---

## Optimization Flags Reference

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--strategy` | string | Strategy YAML file | Required |
| `--data` | string | Historical data file | Required |
| `--parameters` | string | Parameter ranges to optimize | Required |
| `--objective` | string | Optimization objective | sharpe |
| `--direction` | string | maximize or minimize | maximize |
| `--parallel` | int | Number of parallel workers | 1 |
| `--output` | string | Output JSON file | - |
| `--csv` | string | Export all results to CSV | - |
| `--top` | int | Export only top N to CSV | 0 (all) |
| `--sensitivity` | string | Export sensitivity analysis CSV | - |

**Available Objectives:**
- `sharpe`: Sharpe Ratio
- `sortino`: Sortino Ratio
- `return`: Total Return
- `profit_factor`: Profit Factor
- `calmar`: Calmar Ratio
- `expectancy`: Expectancy
- `win_rate`: Win Rate
- `cagr`: CAGR

---

## Best Practices for Optimization

### 1. Start with Coarse Grid

```bash
# First pass: Coarse grid
trader optimize \
  --parameters "ema_fast:5:30:5,ema_slow:20:200:20" \
  --csv coarse.csv

# Identify promising regions in coarse.csv
# Then refine:

# Second pass: Fine grid around best area
trader optimize \
  --parameters "ema_fast:10:20:1,ema_slow:40:80:5" \
  --csv fine.csv
```

### 2. Always Export Sensitivity

Understanding parameter sensitivity helps avoid overfitting:

```bash
trader optimize \
  ... \
  --sensitivity sensitivity.csv
```

Review correlation and sensitivity before trusting results.

### 3. Use Parallel Workers

Speed up optimization with multiple cores:

```bash
trader optimize \
  ... \
  --parallel 8  # Use 8 CPU cores
```

### 4. Validate Out-of-Sample

Never trust in-sample optimization alone:

```bash
# Optimize on training data
trader optimize --data train.csv --csv optimized.csv

# Get best parameters from optimized.csv
# Then test on validation data:
trader backtest --strategy best_params.yaml --data validation.csv
```

### 5. Check for Stability

Stable parameters should:
- Not be at edge of search range
- Have moderate sensitivity (not too high)
- Show consistent performance in top N results
- Work across different time periods


---

## report

Generate professional reports from backtest results in multiple formats.

### Basic Usage

```bash
# Generate HTML report
trader report --result backtest_result.json --format html --output report.html

# Generate Markdown report
trader report --result backtest_result.json --format markdown --output report.md

# Generate text report to console
trader report --result backtest_result.json --format text
```

### Report Formats

#### HTML Report
Professional HTML report with embedded CSS styling:
```bash
trader report \
  --result results/my_strategy.json \
  --format html \
  --output reports/my_strategy.html \
  --title "EMA Crossover Strategy Analysis" \
  --theme dark
```

Features:
- Responsive design
- Light/dark themes
- Professional styling with tables and metrics
- Color-coded performance indicators (green/red)
- Strategy summary, metrics, and trade history

#### Markdown Report
GitHub-compatible Markdown with proper tables:
```bash
trader report \
  --result results/my_strategy.json \
  --format markdown \
  --output reports/my_strategy.md \
  --title "Strategy Performance Report"
```

Features:
- Proper Markdown tables
- Headers and sections
- Easy to version control
- Compatible with GitHub/GitLab

#### Text Report
Console-friendly plain text output:
```bash
trader report \
  --result results/my_strategy.json \
  --format text
```

Features:
- Unicode box-drawing characters
- Aligned columns
- Suitable for terminal display or log files

### Report Sections

All formats include:
1. **Strategy Summary**: Name, symbol, timeframe, period, initial cash, final equity
2. **Performance Metrics**: Returns (Total, CAGR), risk-adjusted metrics (Sharpe, Sortino, Calmar), risk metrics (Max Drawdown, Avg Drawdown), trade statistics (Win Rate, Profit Factor, total trades, avg win/loss)
3. **Trade History**: Summary of all completed trades with entry/exit times, prices, PnL

### Report Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--result` | string | Path to backtest result JSON | Required |
| `--format` | string | Output format: html, markdown, text | html |
| `--output` | string | Output file path (omit for stdout) | - |
| `--title` | string | Report title | "Backtest Report" |
| `--theme` | string | HTML theme: light or dark | light |
| `--no-css` | bool | Exclude CSS from HTML output | false |

### Examples

#### Complete Workflow: Backtest → Report

```bash
# 1. Run backtest and save results
trader backtest \
  --strategy strategies/ema_cross.yaml \
  --data data/BTCUSDT_4h.parquet \
  --output results/ema_cross_2024.json

# 2. Generate HTML report with dark theme
trader report \
  --result results/ema_cross_2024.json \
  --format html \
  --output reports/ema_cross_2024.html \
  --title "EMA Crossover - BTC 2024" \
  --theme dark

# 3. Generate Markdown for documentation
trader report \
  --result results/ema_cross_2024.json \
  --format markdown \
  --output docs/strategy_results.md

# 4. Quick console review
trader report \
  --result results/ema_cross_2024.json \
  --format text
```

#### Batch Report Generation

```bash
#!/bin/bash
# Generate reports for all backtest results

for result in results/*.json; do
  name=$(basename "$result" .json)
  
  # HTML report
  trader report \
    --result "$result" \
    --format html \
    --output "reports/${name}.html" \
    --title "$name"
  
  # Markdown report
  trader report \
    --result "$result" \
    --format markdown \
    --output "reports/${name}.md" \
    --title "$name"
done

echo "Generated reports for $(ls results/*.json | wc -l) backtests"
```

#### Custom Styling

For HTML reports without embedded CSS (to use custom stylesheet):
```bash
trader report \
  --result results/my_strategy.json \
  --format html \
  --output reports/my_strategy.html \
  --no-css

# Then add your own <link> tag in the HTML file
```

### Report Output Examples

#### HTML Output Features
- Strategy name and parameters in header
- Color-coded metrics (green for positive, red for negative)
- Responsive tables that work on mobile
- Dark theme for reduced eye strain
- Professional appearance suitable for client presentations

#### Markdown Output Features
- Proper table formatting with alignment
- Compatible with GitHub/GitLab rendering
- Easy to include in documentation
- Version control friendly

#### Text Output Features
- Clean terminal display
- Unicode box characters for visual separation
- Suitable for logging or email
- No markup overhead

### Integration with Other Commands

```bash
# Backtest → Report
trader backtest --strategy s.yaml --data d.csv --output r.json
trader report --result r.json --format html --output report.html

# Walk Forward → Report (for each window)
trader walkforward --strategy s.yaml --data d.csv --csv-windows windows.csv
# Then generate reports for train/test results

# Monte Carlo → Report (original backtest)
trader montecarlo --result r.json --simulations 1000
trader report --result r.json --format html --output base_report.html
```

### Tips

1. **Use descriptive titles**: Help identify reports later
2. **Dark theme for presentations**: Easier on eyes in dark rooms
3. **Markdown for documentation**: Version control friendly
4. **Text for quick review**: Fast console preview
5. **Batch generation**: Automate report creation for multiple strategies

