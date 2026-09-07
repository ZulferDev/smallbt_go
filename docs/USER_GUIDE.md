# SmallBT User Guide

Complete guide to using the SmallBT declarative quantitative trading backtesting engine.

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [Strategy Configuration](#strategy-configuration)
5. [Data Formats](#data-formats)
6. [Complete Workflow Examples](#complete-workflow-examples)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Installation

### Prerequisites

- Go 1.24 or higher
- Git

### Build from Source

```bash
git clone https://github.com/ZulferDev/smallbt_go.git
cd smallbt_go
go build -o trader ./cmd/trader
```

### Verify Installation

```bash
./trader --version
./trader --help
```

---

## Quick Start

### 1. Create Your First Strategy

Create a file `my_strategy.yaml`:

```yaml
strategy:
  name: simple_ema_cross
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

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

exit:
  long:
    any:
      - cross_below: [ema_fast, ema_slow]

risk:
  position_size:
    type: percent_equity
    value: 0.1
  
  max_positions: 1
```

### 2. Validate Your Strategy

```bash
./trader validate my_strategy.yaml
```

Expected output:
```
✅ Strategy 'simple_ema_cross' (v1) validated successfully
   Symbol: BTCUSDT, Timeframe: 1h
   Indicators: 2
   Entry rules: 1
```

### 3. Run Your First Backtest

```bash
./trader backtest \
  --strategy my_strategy.yaml \
  --data data/BTCUSDT.csv \
  --output results/backtest_result.json
```

---

## CLI Commands

### `validate` - Validate Strategy Configuration

Validates a strategy YAML file for syntax errors, missing fields, circular dependencies, and invalid parameters.

**Usage:**
```bash
./trader validate [OPTIONS] <strategy.yaml>
```

**Examples:**

```bash
# Basic validation
./trader validate strategies/ema_cross.yaml

# Verbose output
./trader validate --verbose strategies/complex_strategy.yaml
```

**What it checks:**
- YAML syntax
- Required fields
- Indicator dependencies
- Circular references
- Parameter types and ranges
- Entry/exit condition validity

---

### `validate-transforms` - Validate Data Transforms

Validates data transformation pipeline in a strategy by applying it to sample data.

**Usage:**
```bash
./trader validate-transforms --strategy <strategy.yaml> --data <data.csv>
```

**Options:**
- `--strategy` - Path to strategy YAML file
- `--data` - Path to sample data CSV file
- `--rows` - Number of rows to process (default: 100)

**Example:**

```bash
./trader validate-transforms \
  --strategy strategies/with_transforms.yaml \
  --data data/BTCUSDT.csv \
  --rows 50
```

**Sample output:**
```
Validating transforms with 50 rows of data...
✅ Transform 1: resample_4h - OK (50 → 12 bars)
✅ Transform 2: filter_volume - OK (12 → 10 bars)
✅ Transform 3: add_signals - OK (10 → 10 bars)
All transforms validated successfully!
```

---

### `backtest` - Run Backtests

Runs a complete backtest with a strategy and historical data.

**Usage:**
```bash
./trader backtest [OPTIONS]
```

**Required Options:**
- `--strategy <file>` - Path to strategy YAML file
- `--data <file>` - Path to market data CSV or Parquet file

**Optional Options:**
- `--output <file>` - Output JSON file path (default: none)
- `--csv <file>` - Export trades to CSV
- `--symbol <string>` - Trading symbol (default: from strategy or BTCUSDT)
- `--cash <float>` - Initial cash (default: 10000)
- `--start <date>` - Start date (YYYY-MM-DD)
- `--end <date>` - End date (YYYY-MM-DD)

**Examples:**

#### Basic Backtest
```bash
./trader backtest \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv
```

#### With Date Range
```bash
./trader backtest \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-01-01 \
  --end 2023-12-31
```

#### Save Results to File
```bash
./trader backtest \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --output results/my_backtest.json \
  --csv results/trades.csv
```

#### Custom Initial Capital
```bash
./trader backtest \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --cash 100000
```

**Output Explanation:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BACKTEST RESULT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy       ema_volume_atr
Symbol         BTCUSDT
Timeframe      1h
Period         2023-01-01 → 2023-12-31
Runtime        45.2ms

Return         +23.45%        ← Total return
CAGR           +23.45%        ← Annualized return
Sharpe         1.85           ← Risk-adjusted return (>1 is good)
Sortino        2.31           ← Downside risk-adjusted return
Max Drawdown   -12.34%        ← Largest peak-to-trough decline

Trades         142            ← Total number of trades
Win Rate       54.23%         ← Percentage of winning trades
Profit Factor  1.65           ← Gross profit / Gross loss (>1 profitable)
Expectancy     +0.38R         ← Average R-multiple per trade

Final Equity   $12,345.67
Total Fees     $234.56
Net PnL        $2,111.11
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### `optimize` - Parameter Optimization

Optimizes strategy parameters using grid search to find the best combination.

**Usage:**
```bash
./trader optimize [OPTIONS]
```

**Required Options:**
- `--strategy <file>` - Path to strategy YAML file
- `--data <file>` - Path to market data file
- `--parameters <string>` - Parameter ranges in format: `name:start:end:step`

**Optional Options:**
- `--output <file>` - Output JSON file path
- `--csv <file>` - Export all results to CSV
- `--sensitivity <file>` - Export parameter sensitivity analysis to CSV
- `--objective <metric>` - Optimization objective (default: sharpe)
  - Options: `sharpe`, `sortino`, `return`, `profit_factor`, `win_rate`, `expectancy`
- `--direction <string>` - Optimization direction: `maximize` or `minimize` (default: maximize)
- `--parallel <int>` - Number of parallel workers (default: 1)
- `--top <int>` - Export only top N results to CSV (0 = all)
- `--cash <float>` - Initial cash (default: 10000)
- `--start <date>` - Start date (YYYY-MM-DD)
- `--end <date>` - End date (YYYY-MM-DD)

**Examples:**

#### Single Parameter Optimization
```bash
./trader optimize \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --parameters "ema_fast.period:5:20:1" \
  --objective sharpe
```

#### Multiple Parameters
```bash
./trader optimize \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --parameters "ema_fast.period:5:15:2,ema_slow.period:20:50:5,volume_ratio:1.0:2.0:0.1" \
  --objective sharpe \
  --parallel 4
```

#### Save Top Results
```bash
./trader optimize \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --parameters "ema_fast.period:5:20:1,ema_slow.period:20:60:5" \
  --objective sharpe \
  --output results/optimization.json \
  --csv results/all_results.csv \
  --top 10 \
  --parallel 8
```

#### With Sensitivity Analysis
```bash
./trader optimize \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --parameters "ema_fast.period:5:20:1,ema_slow.period:20:60:5" \
  --objective sharpe \
  --sensitivity results/sensitivity.csv \
  --parallel 4
```

**Parameter Format:**

```
name:start:end:step
```

- `name`: Parameter path in strategy (e.g., `ema_fast.period`, `risk.stop_loss.multiplier`)
- `start`: Starting value
- `end`: Ending value (inclusive)
- `step`: Step size

**Output Example:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PARAMETER OPTIMIZATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy:    strategies/ema_volume.yaml
Data:        data/BTCUSDT.csv
Symbol:      BTCUSDT
Objective:   sharpe (maximize)
Algorithm:   Grid Search
Parameters:  2
  - ema_fast.period: [5.00 to 15.00, step 2.00]
  - ema_slow.period: [20.00 to 50.00, step 5.00]

Total Combinations: 42
Parallel Workers:   4
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Progress: [████████████████████] 42/42 (100%)

BEST RESULT
Parameters: ema_fast.period:9;ema_slow.period:25;
Sharpe: 2.15
Backtest Metrics:
  Total Return:  +34.56%
  CAGR:          +34.56%
  Sharpe:        2.15
  Sortino:       2.87
  Max Drawdown:  -8.45%
  Win Rate:      58.33%
  Profit Factor: 2.12
  Expectancy:    +0.52R
  Trades:        96

TOP 5 RESULTS
Rank   Parameters                     Sharpe
------------------------------------------------------------
1      ema_fast.period:9;ema_...      2.15
2      ema_fast.period:7;ema_...      2.08
3      ema_fast.period:11;ema...      1.98
4      ema_fast.period:9;ema_...      1.95
5      ema_fast.period:7;ema_...      1.89

Optimization completed in 3.2s
```

---

### `walkforward` - Walk Forward Analysis

Performs Walk Forward Analysis to test strategy robustness and avoid overfitting.

**Usage:**
```bash
./trader walkforward [OPTIONS]
```

**Required Options:**
- `--strategy <file>` - Path to strategy YAML file
- `--data <file>` - Path to market data file

**Optional Options:**
- `--output <file>` - Output JSON file path
- `--train <int>` - Number of bars for training period (default: 1000)
- `--test <int>` - Number of bars for testing period (default: 200)
- `--step <int>` - Number of bars to step forward (default: same as test)
- `--csv-windows <file>` - Export window results to CSV
- `--csv-aggregate <file>` - Export aggregate metrics to CSV
- `--csv-stability <file>` - Export stability analysis to CSV
- `--cash <float>` - Initial cash (default: 10000)
- `--symbol <string>` - Trading symbol

**Examples:**

#### Basic Walk Forward
```bash
./trader walkforward \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --train 1000 \
  --test 200 \
  --step 200
```

#### With All Exports
```bash
./trader walkforward \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --train 2000 \
  --test 500 \
  --step 250 \
  --output results/walkforward.json \
  --csv-windows results/wf_windows.csv \
  --csv-aggregate results/wf_aggregate.csv \
  --csv-stability results/wf_stability.csv
```

**How Walk Forward Works:**

```
Data Timeline:
|----Train----|--Test--|----Train----|--Test--|----Train----|--Test--|
    1000       200        1000       200        1000       200
    
Window 1: Train on bars 1-1000,    test on bars 1001-1200
Window 2: Train on bars 201-1200,  test on bars 1201-1400
Window 3: Train on bars 401-1400,  test on bars 1401-1600
...
```

**Output Example:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WALK FORWARD ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy:       strategies/ema_volume.yaml
Symbol:         BTCUSDT
Timeframe:      1h
Total Bars:     5000
Train Bars:     1000
Test Bars:      200
Step Bars:      200
Windows:        20

WINDOW RESULTS
Window  Period              Return    Sharpe   MaxDD    Trades
------  ------------------  --------  -------  -------  ------
1       2023-01-01→2023-02  +12.3%    1.85     -5.2%    15
2       2023-02-01→2023-03  +8.7%     1.52     -6.8%    12
3       2023-03-01→2023-04  -2.1%     -0.42    -8.3%    14
...

AGGREGATE METRICS
Out-of-Sample Performance:
  Total Return:      +45.67%
  CAGR:             +45.67%
  Sharpe:           1.23
  Sortino:          1.67
  Max Drawdown:     -15.43%
  Win Rate:         52.34%
  Profit Factor:    1.45
  Total Trades:     280

STABILITY ANALYSIS
Metric Consistency:
  Return Stability:     0.72  ← Consistency across windows (0-1)
  Sharpe Stability:     0.68
  Drawdown Stability:   0.81
  
Window Performance:
  Profitable Windows:   65%   ← Percentage of profitable windows
  Avg Win Window:       +8.5%
  Avg Loss Window:      -3.2%
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### `montecarlo` - Monte Carlo Simulation

Runs Monte Carlo simulation on backtest results to estimate confidence intervals and probability distributions.

**Usage:**
```bash
./trader montecarlo [OPTIONS]
```

**Required Options:**
- `--strategy <file>` - Path to strategy YAML file
- `--data <file>` - Path to market data file

**OR**

- `--result <file>` - Path to previous backtest result JSON

**Optional Options:**
- `--output <file>` - Output JSON file path
- `--simulations <int>` - Number of Monte Carlo runs (default: 10000)
- `--seed <int>` - Random seed for reproducibility
- `--method <string>` - Simulation method: `trade_reshuffle` or `return_reshuffle` (default: trade_reshuffle)
- `--cash <float>` - Initial cash (default: 10000)

**Examples:**

#### Run Monte Carlo from Strategy
```bash
./trader montecarlo \
  --strategy strategies/ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --simulations 10000 \
  --output results/montecarlo.json
```

#### Run from Previous Backtest Result
```bash
./trader montecarlo \
  --result results/backtest.json \
  --simulations 10000 \
  --seed 42 \
  --output results/montecarlo.json
```

#### Different Simulation Method
```bash
./trader montecarlo \
  --result results/backtest.json \
  --simulations 10000 \
  --method return_reshuffle \
  --output results/montecarlo.json
```

**Simulation Methods:**

1. **trade_reshuffle** (default): Reshuffles the order of trades
   - Preserves individual trade characteristics
   - Tests sensitivity to trade sequence
   
2. **return_reshuffle**: Reshuffles individual returns
   - More aggressive randomization
   - Tests broader outcome distribution

**Output Example:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MONTE CARLO SIMULATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy:       strategies/ema_volume.yaml
Method:         trade_reshuffle
Simulations:    10000
Seed:           42

FINAL RETURN DISTRIBUTION
  Original:     +23.45%
  Mean:         +22.87%
  Median:       +23.12%
  Std Dev:      8.34%
  
Percentiles:
  5th:          +8.23%    ← 5% chance of returns below this
  25th:         +17.45%
  50th:         +23.12%
  75th:         +28.67%
  95th:         +38.91%   ← 5% chance of returns above this

MAX DRAWDOWN DISTRIBUTION
  Original:     -12.34%
  Mean:         -13.87%
  Median:       -13.12%
  Worst Case:   -28.45%   ← 95th percentile worst drawdown
  
SHARPE RATIO DISTRIBUTION
  Original:     1.85
  Mean:         1.78
  Median:       1.81
  
PROBABILITY ANALYSIS
  P(Return > 0):        87.5%   ← Chance of profit
  P(Return > 10%):      78.2%
  P(Return > 20%):      54.3%
  P(Drawdown < -20%):   8.7%    ← Chance of severe drawdown
  
CONFIDENCE INTERVALS (95%)
  Final Return:   [+8.23%, +38.91%]
  Max Drawdown:   [-7.23%, -28.45%]
  Sharpe Ratio:   [0.45, 2.98]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### `paper` - Paper Trading

Runs paper trading simulation with real-time or simulated market data.

**Usage:**
```bash
./trader paper [OPTIONS]
```

**Required Options:**
- `--strategy <file>` - Path to strategy YAML file
- `--symbol <string>` - Trading symbol
- `--price <float>` - Starting price (for simulation mode)

**OR**

- `--feed <string>` - Live data feed source (e.g., `binance`, `coinbase`)

**Optional Options:**
- `--cash <float>` - Initial cash (default: 10000)
- `--duration <duration>` - Trading duration (e.g., `1h`, `24h`, `7d`)
- `--interval <duration>` - Update interval (e.g., `1s`, `5s`, `1m`)
- `--output <file>` - Output JSON file path
- `--csv <file>` - Export trades to CSV
- `--verbose` - Verbose logging

**Examples:**

#### Simulated Paper Trading
```bash
./trader paper \
  --strategy strategies/ema_volume.yaml \
  --symbol BTCUSDT \
  --price 50000 \
  --duration 24h \
  --interval 1m
```

#### With Custom Settings
```bash
./trader paper \
  --strategy strategies/ema_volume.yaml \
  --symbol BTCUSDT \
  --price 45000 \
  --cash 100000 \
  --duration 48h \
  --interval 5m \
  --output results/paper_trading.json \
  --csv results/paper_trades.csv \
  --verbose
```

**Output Example:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PAPER TRADING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy:       ema_volume_atr
Symbol:         BTCUSDT
Starting Price: $50,000
Initial Cash:   $10,000
Duration:       24h
Interval:       1m

[2023-06-15 10:00:00] Paper trading started
[2023-06-15 10:05:23] LONG ENTRY @ $50,250 | Size: 0.198 BTC | Risk: 1%
[2023-06-15 10:15:45] Price: $50,450 | Unrealized P&L: +$39.60
[2023-06-15 11:23:12] LONG EXIT @ $51,100 | P&L: +$168.30 (+1.69%)
[2023-06-15 14:45:00] LONG ENTRY @ $49,800 | Size: 0.204 BTC
...

Current Status:
  Elapsed:        8h 30m
  Current Price:  $51,234
  Equity:         $10,567.89
  Open Positions: 1
  Total Trades:   12
  Win Rate:       58.33%
  Total P&L:      +$567.89 (+5.68%)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Press Ctrl+C to stop paper trading...
```

---

### `report` - Generate Reports

Generates detailed reports from backtest results in various formats.

**Usage:**
```bash
./trader report [OPTIONS]
```

**Required Options:**
- `--result <file>` - Path to backtest result JSON file

**Optional Options:**
- `--format <string>` - Report format: `text`, `markdown`, `html`, `json` (default: text)
- `--output <file>` - Output file path (default: stdout)
- `--template <file>` - Custom report template
- `--charts` - Include charts (HTML format only)

**Examples:**

#### Text Report (Console)
```bash
./trader report --result results/backtest.json
```

#### Markdown Report
```bash
./trader report \
  --result results/backtest.json \
  --format markdown \
  --output reports/backtest_report.md
```

#### HTML Report with Charts
```bash
./trader report \
  --result results/backtest.json \
  --format html \
  --output reports/backtest_report.html \
  --charts
```

---

### `export-trades` - Export Trade History

Exports trade history from backtest results to CSV format.

**Usage:**
```bash
./trader export-trades [OPTIONS]
```

**Required Options:**
- `--result <file>` - Path to backtest result JSON file
- `--output <file>` - Output CSV file path

**Optional Options:**
- `--filter <string>` - Filter trades: `winning`, `losing`, `all` (default: all)
- `--min-pnl <float>` - Minimum P&L filter
- `--max-pnl <float>` - Maximum P&L filter

**Examples:**

#### Export All Trades
```bash
./trader export-trades \
  --result results/backtest.json \
  --output results/all_trades.csv
```

#### Export Only Winning Trades
```bash
./trader export-trades \
  --result results/backtest.json \
  --output results/winning_trades.csv \
  --filter winning
```

#### Export with P&L Filter
```bash
./trader export-trades \
  --result results/backtest.json \
  --output results/large_trades.csv \
  --min-pnl 100
```

**CSV Output Format:**

```csv
TradeID,Symbol,Side,EntryTime,EntryPrice,ExitTime,ExitPrice,Quantity,GrossPnL,Fees,NetPnL,Return,MAE,MFE,ExitReason
1,BTCUSDT,LONG,2023-01-15T10:00:00Z,20000,2023-01-15T14:30:00Z,20450,0.5,225.00,4.50,220.50,1.13%,-50.00,300.00,take_profit
2,BTCUSDT,LONG,2023-01-16T09:15:00Z,20100,2023-01-16T12:45:00Z,19850,0.5,-125.00,4.50,-129.50,-0.64%,-200.00,50.00,stop_loss
...
```

---

### `analyze-trades` - Analyze Trade Statistics

Analyzes trade history and generates detailed statistics and patterns.

**Usage:**
```bash
./trader analyze-trades [OPTIONS]
```

**Required Options:**
- `--result <file>` - Path to backtest result JSON file

**Optional Options:**
- `--output <file>` - Output analysis report to file
- `--csv <file>` - Export analysis to CSV
- `--group-by <string>` - Group analysis by: `day`, `week`, `month`, `hour`

**Examples:**

#### Basic Analysis
```bash
./trader analyze-trades --result results/backtest.json
```

#### Group by Time Period
```bash
./trader analyze-trades \
  --result results/backtest.json \
  --group-by month \
  --output results/monthly_analysis.txt
```

**Output Example:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TRADE ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Backtest:  results/backtest.json
Strategy:  ema_volume_atr
Period:    2023-01-01 → 2023-12-31

OVERVIEW
Total Trades:         142
Winning Trades:       77 (54.23%)
Losing Trades:        65 (45.77%)

PERFORMANCE METRICS
Gross Profit:         $5,234.56
Gross Loss:          -$3,123.45
Net Profit:           $2,111.11
Profit Factor:        1.68
Expectancy:           $14.87 per trade

RETURN STATISTICS
Average Win:          +2.34%
Average Loss:         -1.23%
Win/Loss Ratio:       1.90
Largest Win:          +8.45%
Largest Loss:         -3.21%

TRADE DURATION
Average:              4h 23m
Median:               3h 45m
Shortest:             1h 12m
Longest:              12h 34m

EXIT ANALYSIS
Take Profit:          45 (31.69%)
Stop Loss:            38 (26.76%)
Trailing Stop:        23 (16.20%)
Signal Exit:          36 (25.35%)

DRAWDOWN ANALYSIS
Maximum Drawdown:     -12.34%
Average Drawdown:     -3.45%
Max Drawdown Duration: 18 days
Recovery Time:        23 days

CONSECUTIVE TRADES
Max Consecutive Wins:  7
Max Consecutive Losses: 5
Avg Streak Length:     2.3

TIME OF DAY ANALYSIS
Best Hour:    10:00-11:00 (Win Rate: 67%)
Worst Hour:   22:00-23:00 (Win Rate: 38%)

MONTHLY BREAKDOWN
Month       Trades  Win%    Net P&L     Return
------      ------  -----   --------    ------
2023-01     12      58.3%   +$234.56    +2.35%
2023-02     15      46.7%   -$123.45    -1.22%
2023-03     14      64.3%   +$456.78    +4.72%
...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Strategy Configuration

### Complete Strategy Structure

```yaml
strategy:
  name: string              # Strategy name (required)
  version: string           # Version (required)
  description: string       # Optional description

data:
  symbol: string           # Trading symbol (required)
  timeframe: string        # Timeframe: 1m, 5m, 15m, 1h, 4h, 1d, etc. (required)
  
  # Optional data transformations
  transforms:
    - type: resample
      timeframe: 4h
    - type: filter
      condition: gt
      field: volume
      value: 1000

indicators:
  indicator_name:
    type: string            # Indicator type (required)
    source: string          # Source field: close, high, low, volume, etc.
    period: integer         # Period (if applicable)
    # Additional indicator-specific parameters

entry:
  long:                     # Long entry conditions
    all:                    # All conditions must be true (AND)
      - condition
    any:                    # Any condition must be true (OR)
      - condition
      
  short:                    # Short entry conditions (optional)
    all:
      - condition

exit:
  long:                     # Long exit conditions
    any:
      - condition
      
  short:                    # Short exit conditions
    any:
      - condition

risk:
  position_size:            # Position sizing
    type: string           # fixed, percent_equity, risk_percent, etc.
    value: float
    
  stop_loss:               # Stop loss configuration
    type: string           # fixed, percent, atr, etc.
    value: float           # or multiplier for ATR
    
  take_profit:             # Take profit configuration
    type: string           # fixed, percent, risk_reward, etc.
    value: float           # or ratio
    
  trailing_stop:           # Optional trailing stop
    type: string
    value: float
    
  max_positions: integer   # Maximum concurrent positions
  max_daily_loss: float    # Maximum daily loss (0-1)
  max_portfolio_risk: float # Maximum portfolio risk (0-1)

state:                      # Optional stateful variables
  variable_name:
    type: bool|int|float|string
    default: value
```

### Available Indicators

#### Moving Averages
- `sma` - Simple Moving Average
- `ema` - Exponential Moving Average
- `wma` - Weighted Moving Average
- `vwap` - Volume Weighted Average Price

#### Momentum
- `rsi` - Relative Strength Index
- `macd` - Moving Average Convergence Divergence
- `stoch` - Stochastic Oscillator
- `cci` - Commodity Channel Index
- `roc` - Rate of Change

#### Volatility
- `atr` - Average True Range
- `bb` - Bollinger Bands
- `kc` - Keltner Channels
- `dc` - Donchian Channels

#### Volume
- `obv` - On Balance Volume
- `vwma` - Volume Weighted Moving Average
- `mfi` - Money Flow Index

#### Trend
- `adx` - Average Directional Index
- `aroon` - Aroon Indicator
- `psar` - Parabolic SAR

#### Custom/Composite
- `add` - Addition
- `subtract` - Subtraction
- `multiply` - Multiplication
- `divide` - Division
- `highest` - Highest value over period
- `lowest` - Lowest value over period

### Available Conditions

#### Comparison
- `gt` - Greater than
- `gte` - Greater than or equal
- `lt` - Less than
- `lte` - Less than or equal
- `eq` - Equal
- `neq` - Not equal

#### Crossovers
- `cross_above` - Value crosses above another
- `cross_below` - Value crosses below another

#### Trends
- `rising` - Value is rising over period
- `falling` - Value is falling over period

#### Ranges
- `between` - Value is between two values
- `outside` - Value is outside range

#### Logical
- `all` - All conditions must be true (AND)
- `any` - Any condition must be true (OR)
- `not` - Negates condition

---

## Data Formats

### CSV Format

The engine supports flexible CSV formats with automatic column detection from headers.

**Quick Start - Standard Format:**
```csv
timestamp,open,high,low,close,volume
2023-01-01T00:00:00Z,20000.00,20100.00,19900.00,20050.00,1234.56
2023-01-01T01:00:00Z,20050.00,20150.00,20000.00,20100.00,2345.67
```

**Key Features:**
- ✓ Automatic column detection from headers
- ✓ Flexible column ordering (any order supported)
- ✓ Case-insensitive headers (TIMESTAMP, Open, HIGH, etc.)
- ✓ Alternative names (time, date, datetime, vol)
- ✓ Multiple timestamp formats including milliseconds

**Supported Timestamp Formats:**
- RFC3339: `2023-01-01T00:00:00Z`
- ISO8601: `2023-01-01T00:00:00+00:00`
- With milliseconds: `2023-01-01 00:00:00.000`
- Unix timestamp: `1672531200` (seconds or milliseconds)

**For complete CSV format documentation, troubleshooting, and advanced usage, see:**
→ [CSV_FORMAT.md](CSV_FORMAT.md)

### Parquet Format

Same schema as CSV but in Apache Parquet format for better performance with large datasets.

---

## Complete Workflow Examples

### Example 1: Basic EMA Crossover Strategy

**Step 1: Create strategy file `ema_cross.yaml`**

```yaml
strategy:
  name: ema_crossover
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  ema9:
    type: ema
    source: close
    period: 9

  ema21:
    type: ema
    source: close
    period: 21

entry:
  long:
    all:
      - cross_above: [ema9, ema21]

exit:
  long:
    any:
      - cross_below: [ema9, ema21]

risk:
  position_size:
    type: percent_equity
    value: 0.1
  max_positions: 1
```

**Step 2: Validate**
```bash
./trader validate ema_cross.yaml
```

**Step 3: Backtest**
```bash
./trader backtest \
  --strategy ema_cross.yaml \
  --data data/BTCUSDT.csv \
  --output results/ema_cross.json
```

**Step 4: Analyze**
```bash
./trader analyze-trades --result results/ema_cross.json
```

---

### Example 2: RSI Mean Reversion with Risk Management

**Create `rsi_meanrev.yaml`:**

```yaml
strategy:
  name: rsi_mean_reversion
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  rsi:
    type: rsi
    source: close
    period: 14
    
  atr:
    type: atr
    period: 14
    
  ema200:
    type: ema
    source: close
    period: 200

entry:
  long:
    all:
      - lt: [rsi, 30]           # RSI oversold
      - gt: [close, ema200]     # Above long-term trend

exit:
  long:
    any:
      - gt: [rsi, 70]           # RSI overbought

risk:
  position_size:
    type: risk_percent
    value: 0.01                 # Risk 1% per trade
    
  stop_loss:
    type: atr
    indicator: atr
    multiplier: 2.0             # 2x ATR stop
    
  take_profit:
    type: risk_reward
    ratio: 3                    # 3:1 risk/reward
    
  max_positions: 3
  max_daily_loss: 0.03          # Max 3% daily loss
```

**Run complete workflow:**

```bash
# 1. Validate
./trader validate rsi_meanrev.yaml

# 2. Backtest
./trader backtest \
  --strategy rsi_meanrev.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-01-01 \
  --end 2023-06-30 \
  --output results/rsi_train.json

# 3. Walk Forward Analysis
./trader walkforward \
  --strategy rsi_meanrev.yaml \
  --data data/BTCUSDT.csv \
  --train 2000 \
  --test 500 \
  --output results/rsi_wf.json

# 4. Monte Carlo
./trader montecarlo \
  --result results/rsi_train.json \
  --simulations 10000 \
  --output results/rsi_mc.json

# 5. Generate Report
./trader report \
  --result results/rsi_train.json \
  --format markdown \
  --output reports/rsi_report.md
```

---

### Example 3: Multi-Indicator Strategy with Optimization

**Create `multi_indicator.yaml`:**

```yaml
strategy:
  name: multi_indicator_trend
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  ema_fast:
    type: ema
    source: close
    period: 9
    
  ema_slow:
    type: ema
    source: close
    period: 21
    
  rsi:
    type: rsi
    source: close
    period: 14
    
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
      - gt: [rsi, 50]
      - gt: [volume_ratio, 1.2]

exit:
  long:
    any:
      - cross_below: [ema_fast, ema_slow]
      - lt: [rsi, 40]

risk:
  position_size:
    type: risk_percent
    value: 0.015
    
  stop_loss:
    type: atr
    indicator: atr
    multiplier: 1.5
    
  take_profit:
    type: risk_reward
    ratio: 2.5
    
  trailing_stop:
    type: atr
    indicator: atr
    multiplier: 1.0
    
  max_positions: 2
```

**Optimize parameters:**

```bash
./trader optimize \
  --strategy multi_indicator.yaml \
  --data data/BTCUSDT.csv \
  --parameters "ema_fast.period:5:15:1,ema_slow.period:20:30:2,rsi.period:10:20:2,volume_ratio:1.0:2.0:0.1" \
  --objective sharpe \
  --parallel 8 \
  --output results/optimization.json \
  --csv results/all_combinations.csv \
  --top 20
```

**Test best parameters with Walk Forward:**

```bash
# Update strategy with best parameters from optimization
# Then run walk forward

./trader walkforward \
  --strategy multi_indicator_optimized.yaml \
  --data data/BTCUSDT.csv \
  --train 3000 \
  --test 750 \
  --step 375 \
  --csv-windows results/wf_windows.csv \
  --output results/wf_final.json
```

---

## Best Practices

### 1. Strategy Development Workflow

1. **Start Simple**
   - Begin with basic indicators and conditions
   - Validate strategy syntax before backtesting
   
2. **Use Train/Test Split**
   - Train on 70% of data
   - Test on remaining 30%
   - Never optimize on test data
   
3. **Validate with Walk Forward**
   - Always run Walk Forward Analysis
   - Check for consistency across windows
   - Look for degradation in out-of-sample performance
   
4. **Assess Risk with Monte Carlo**
   - Run Monte Carlo with 10,000+ simulations
   - Check worst-case scenarios (5th percentile)
   - Verify probability of acceptable returns

5. **Paper Trade Before Live**
   - Run paper trading for at least 2 weeks
   - Compare paper results with backtest expectations
   - Monitor for execution differences

### 2. Parameter Optimization

**Do:**
- Use reasonable parameter ranges based on market characteristics
- Optimize on in-sample data only
- Validate on out-of-sample data
- Use Walk Forward Analysis to avoid overfitting
- Consider parameter stability (sensitivity analysis)

**Don't:**
- Optimize on entire dataset
- Use extremely narrow parameter ranges
- Cherry-pick parameters based on best single backtest
- Ignore parameter sensitivity
- Over-optimize (too many parameters)

### 3. Risk Management

**Essential Rules:**
- Never risk more than 1-2% per trade
- Use stop losses on every trade
- Limit maximum daily loss (3-5%)
- Limit maximum concurrent positions
- Scale position size with volatility (use ATR)

### 4. Data Quality

**Checklist:**
- Verify no missing timestamps
- Check for duplicate data
- Validate OHLC relationships (H >= O,C,L)
- Ensure sufficient data (minimum 1000 bars)
- Use consistent timezone (UTC recommended)

### 5. Performance Metrics

**Key Metrics to Monitor:**
- **Sharpe Ratio**: > 1.0 good, > 2.0 excellent
- **Profit Factor**: > 1.5 acceptable, > 2.0 good
- **Max Drawdown**: < 20% acceptable, < 10% good
- **Win Rate**: Not critical alone, consider with risk/reward
- **Expectancy**: Must be positive

---

## Troubleshooting

### Common Issues

#### "Strategy validation failed: circular dependency"

**Cause:** Indicators reference each other in a loop.

**Solution:** Check indicator dependencies. Each indicator can only reference indicators defined above it.

```yaml
# Bad - circular dependency
indicators:
  ind_a:
    type: add
    left: close
    right: ind_b  # References ind_b which doesn't exist yet
  ind_b:
    type: add
    left: close
    right: ind_a  # References ind_a - circular!

# Good - proper dependency order
indicators:
  ind_a:
    type: ema
    source: close
    period: 9
  ind_b:
    type: add
    left: close
    right: ind_a  # ind_a already defined above
```

#### "Insufficient data for indicator warmup"

**Cause:** Not enough historical data to calculate indicators with large periods.

**Solution:** Ensure your data has at least `max(indicator_periods) + 100` bars.

```bash
# Check data length
wc -l data/BTCUSDT.csv

# If using EMA 200, need at least 300 bars of data
```

#### "No trades generated"

**Possible causes:**
1. Entry conditions too strict
2. Indicator values not in expected range
3. Data quality issues
4. Risk limits preventing entries

**Debug steps:**

```bash
# 1. Validate strategy
./trader validate strategy.yaml

# 2. Check indicator values (add debug logging in strategy)
# 3. Relax entry conditions temporarily to test
# 4. Check risk parameters (position size, max positions)
```

#### "Optimization taking too long"

**Solutions:**

1. Reduce parameter ranges:
```bash
# Instead of
--parameters "ema:5:50:1"  # 46 combinations

# Use
--parameters "ema:5:50:5"  # 10 combinations
```

2. Use parallel processing:
```bash
--parallel 8  # Use 8 CPU cores
```

3. Reduce data size for initial exploration:
```bash
--start 2023-01-01 --end 2023-03-31  # Test on 3 months first
```

#### "Walk Forward: insufficient bars"

**Cause:** Data doesn't have enough bars for train + test windows.

**Solution:** Reduce window sizes or get more data:

```bash
# Calculate required bars
required = train + test
# With step < test, total_required = train + test + (num_windows * step)

# Adjust parameters
./trader walkforward \
  --train 500 \    # Reduced from 1000
  --test 100 \     # Reduced from 200
  --step 50
```

#### "Monte Carlo: not enough trades"

**Cause:** Backtest generated too few trades for meaningful Monte Carlo analysis.

**Solution:** Need at least 30 trades for useful Monte Carlo results. Either:
- Use more data
- Adjust strategy to generate more trades
- Lower entry condition thresholds

---

## Additional Resources

### Example Strategies

Check `strategies/examples/` for ready-to-use strategy templates:

- `ema_cross.yaml` - Simple EMA crossover
- `ema_volume.yaml` - EMA with volume filter
- `rsi_bb.yaml` - RSI Bollinger Bands mean reversion
- `breakout.yaml` - Donchian channel breakout
- `trend_following.yaml` - Multi-timeframe trend following

### Documentation

- `docs/ARCHITECTURE.md` - System architecture overview
- `docs/STRATEGY_DSL.md` - Complete DSL reference
- `docs/INDICATORS_CONDITIONS.md` - All indicators and conditions
- `docs/BACKTESTING.md` - Backtesting deep dive
- `docs/OPTIMIZATION_WALKFORWARD_MONTECARLO.md` - Advanced analysis
- `docs/DEVELOPER_GUIDE.md` - Extending the system

### Getting Help

1. Check documentation in `docs/`
2. Review example strategies in `strategies/examples/`
3. Run commands with `--help` flag
4. Check GitHub Issues: https://github.com/ZulferDev/smallbt_go/issues

---

## Quick Reference

### Command Cheat Sheet

```bash
# Validate
./trader validate strategy.yaml

# Backtest
./trader backtest --strategy S.yaml --data D.csv --output R.json

# Optimize
./trader optimize --strategy S.yaml --data D.csv --parameters "p:1:10:1" --parallel 4

# Walk Forward
./trader walkforward --strategy S.yaml --data D.csv --train 1000 --test 200

# Monte Carlo
./trader montecarlo --result R.json --simulations 10000

# Paper Trading
./trader paper --strategy S.yaml --symbol BTCUSDT --price 50000 --duration 24h

# Reports
./trader report --result R.json --format markdown --output report.md
./trader analyze-trades --result R.json
./trader export-trades --result R.json --output trades.csv
```

### Strategy Template

```yaml
strategy:
  name: my_strategy
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  # Your indicators here

entry:
  long:
    all:
      - # Your conditions

exit:
  long:
    any:
      - # Your conditions

risk:
  position_size:
    type: risk_percent
    value: 0.01
  stop_loss:
    type: atr
    indicator: atr
    multiplier: 2.0
  take_profit:
    type: risk_reward
    ratio: 2.0
  max_positions: 1
```

---

**Version:** 1.0  
**Last Updated:** 2026-09-07  
**License:** MIT
