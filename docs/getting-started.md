# Getting Started with smallbt_go

**A declarative quantitative trading backtesting engine in Go**

---

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Your First Strategy](#your-first-strategy)
5. [Understanding the Results](#understanding-the-results)
6. [Next Steps](#next-steps)

---

## Introduction

smallbt_go is a powerful backtesting engine that lets you define trading strategies using **YAML configuration** instead of writing code. It's designed for:

- **Quantitative traders** testing algorithmic strategies
- **Researchers** analyzing market behavior
- **Developers** building trading systems

### Key Features

- 📝 **Declarative strategies** - Define strategies in YAML
- 🔄 **Data transforms** - Normalize, smooth, and preprocess data
- 📊 **Built-in indicators** - SMA, EMA, RSI, ATR, and more
- 💰 **Realistic execution** - Fees, slippage, market impact
- 📈 **Comprehensive analytics** - Sharpe, Sortino, drawdown, etc.
- 🔍 **Optimization** - Parameter optimization and walk-forward analysis
- 🎲 **Monte Carlo** - Risk analysis and confidence intervals

---

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Install from Source

```bash
# Clone the repository
git clone https://github.com/ZulferDev/smallbt_go.git
cd smallbt_go

# Build the trader CLI
go build -o trader cmd/trader/main.go

# Verify installation
./trader --version
```

### Quick Test

```bash
# Run a simple backtest
./trader backtest \
  --strategy strategies/examples/sma_cross.yaml \
  --data data/BTCUSDT_500h.csv \
  --cash 10000
```

---

## Quick Start

### 1. Prepare Your Data

smallbt_go supports CSV and Parquet formats. CSV format:

```csv
timestamp,open,high,low,close,volume
2024-01-01 00:00:00,42000,42500,41800,42300,150.5
2024-01-01 01:00:00,42300,42800,42200,42600,180.2
...
```

**Requirements:**
- Chronological order (oldest first)
- Valid OHLCV data
- Consistent timeframe

### 2. Create a Strategy

Create `my_first_strategy.yaml`:

```yaml
strategy:
  name: simple_sma_cross
  version: "1"
  description: "Buy when fast SMA crosses above slow SMA"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  # Fast moving average
  sma_fast:
    type: sma
    source: close
    period: 10
  
  # Slow moving average
  sma_slow:
    type: sma
    source: close
    period: 30

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
    value: 0.5
  
  stop_loss:
    type: percentage
    value: 0.02
  
  take_profit:
    type: percentage
    value: 0.04
```

### 3. Run the Backtest

```bash
./trader backtest \
  --strategy my_first_strategy.yaml \
  --data your_data.csv \
  --cash 10000
```

### 4. View Results

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BACKTEST RESULT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy       simple_sma_cross
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

## Your First Strategy

Let's break down the strategy components:

### Strategy Metadata

```yaml
strategy:
  name: simple_sma_cross
  version: "1"
  description: "Buy when fast SMA crosses above slow SMA"
```

- **name**: Unique identifier for your strategy
- **version**: Version control for strategy iterations
- **description**: Human-readable explanation

### Data Configuration

```yaml
data:
  symbol: BTCUSDT
  timeframe: 1h
```

- **symbol**: Trading pair or instrument
- **timeframe**: Data frequency (1m, 5m, 15m, 1h, 4h, 1d)

### Indicators

```yaml
indicators:
  sma_fast:
    type: sma
    source: close
    period: 10
```

Define technical indicators:
- **type**: Indicator type (sma, ema, rsi, atr, etc.)
- **source**: Data field (open, high, low, close, volume)
- **period**: Lookback period

**Available Indicators:**
- `sma` - Simple Moving Average
- `ema` - Exponential Moving Average
- `rsi` - Relative Strength Index
- `atr` - Average True Range
- `macd` - Moving Average Convergence Divergence
- `bollinger` - Bollinger Bands
- `stoch` - Stochastic Oscillator

### Entry Rules

```yaml
entry:
  long:
    all:
      - cross_above: [sma_fast, sma_slow]
```

**Conditions:**
- `all`: All conditions must be true (AND logic)
- `any`: At least one condition true (OR logic)

**Operators:**
- `cross_above`: Value A crosses above value B
- `cross_below`: Value A crosses below value B
- `gt`: Greater than (>)
- `lt`: Less than (<)
- `gte`: Greater than or equal (≥)
- `lte`: Less than or equal (≤)
- `eq`: Equal to (=)

### Exit Rules

```yaml
exit:
  long:
    any:
      - cross_below: [sma_fast, sma_slow]
```

Exit when **any** condition is met.

### Risk Management

```yaml
risk:
  position_size:
    type: percent_equity
    value: 0.5  # 50% of equity
  
  stop_loss:
    type: percentage
    value: 0.02  # 2% stop loss
  
  take_profit:
    type: percentage
    value: 0.04  # 4% take profit
```

**Position Sizing:**
- `percent_equity`: Percentage of current equity
- `fixed`: Fixed quantity
- `risk_percent`: Based on risk per trade

**Stop Loss Types:**
- `percentage`: Fixed percentage from entry
- `atr`: Based on Average True Range
- `trailing`: Trailing stop

**Take Profit Types:**
- `percentage`: Fixed percentage profit
- `risk_reward`: Multiple of stop loss distance

---

## Understanding the Results

### Key Metrics Explained

**Return Metrics:**
- **Return**: Total percentage return over period
- **CAGR**: Compound Annual Growth Rate

**Risk-Adjusted Returns:**
- **Sharpe Ratio**: Return per unit of risk (> 1.0 is good, > 2.0 is excellent)
- **Sortino Ratio**: Like Sharpe but only penalizes downside volatility
- **Max Drawdown**: Largest peak-to-trough decline

**Trading Metrics:**
- **Trades**: Total number of completed trades
- **Win Rate**: Percentage of profitable trades
- **Profit Factor**: Gross profit / Gross loss (> 1.5 is good)
- **Expectancy**: Average profit per trade in R (risk units)

**Portfolio Metrics:**
- **Final Equity**: Ending account value
- **Total Fees**: Cumulative trading fees
- **Net PnL**: Total profit/loss after fees

### What Makes a Good Strategy?

✅ **Good Strategy Indicators:**
- Sharpe ratio > 1.0
- Win rate > 45%
- Profit factor > 1.5
- Max drawdown < 25%
- Expectancy > 0.3R
- Sufficient trades (> 30 for statistical significance)

❌ **Warning Signs:**
- Too few trades (< 10)
- Win rate > 80% (overfitting)
- Max drawdown > 50%
- Negative expectancy
- Profit factor < 1.2

---

## Next Steps

### 1. Explore Example Strategies

```bash
# View example strategies
ls strategies/examples/

# Try different strategies
./trader backtest --strategy strategies/examples/ema_cross.yaml --data your_data.csv
./trader backtest --strategy strategies/examples/rsi_reversal.yaml --data your_data.csv
```

### 2. Add Data Transforms

Preprocess data before strategy execution:

```yaml
data:
  symbol: BTCUSDT
  timeframe: 1h
  
  transforms:
    enabled: true
    transforms:
      # Smooth price data
      - type: ema_smooth
        field: close
        params:
          period: 5
      
      # Normalize to z-scores
      - type: zscore
        field: close
        params:
          window: 20
```

**Available Transforms:**
- `normalize` - Min-max normalization
- `zscore` - Z-score normalization
- `percentile_rank` - Rank-based normalization
- `ema_smooth` - EMA smoothing
- `smooth` - SMA smoothing
- `difference` - First/second/third order differencing
- `clip` - Cap outliers
- `log_returns` - Logarithmic returns
- `scale` - Scale by factor

See [Transform Guide](transforms.md) for details.

### 3. Optimize Parameters

Find optimal indicator periods:

```bash
./trader optimize \
  --strategy my_strategy.yaml \
  --data your_data.csv \
  --param sma_fast.period:5,50,5 \
  --param sma_slow.period:20,100,10
```

### 4. Walk Forward Analysis

Test strategy robustness:

```bash
./trader walkforward \
  --strategy my_strategy.yaml \
  --data your_data.csv \
  --train 1000 \
  --test 200 \
  --step 100
```

### 5. Monte Carlo Analysis

Assess risk and confidence:

```bash
./trader montecarlo \
  --result backtest_result.json \
  --simulations 10000
```

---

## Common Patterns

### Pattern 1: Trend Following

```yaml
indicators:
  ema_fast:
    type: ema
    period: 12
  ema_slow:
    type: ema
    period: 26

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [volume, volume_avg]
```

### Pattern 2: Mean Reversion

```yaml
indicators:
  rsi:
    type: rsi
    period: 14
  
  bollinger:
    type: bollinger
    period: 20

entry:
  long:
    all:
      - lt: [rsi, 30]  # Oversold
      - lt: [close, bollinger.lower]  # Below lower band
```

### Pattern 3: Breakout

```yaml
indicators:
  highest:
    type: highest
    source: high
    period: 20
  
  volume_avg:
    type: sma
    source: volume
    period: 20

entry:
  long:
    all:
      - gt: [close, highest]
      - gt: [volume, volume_avg * 1.5]
```

---

## Tips & Best Practices

### Strategy Development

1. **Start Simple** - Begin with basic strategies, add complexity gradually
2. **Use Enough Data** - At least 1-2 years for reliable results
3. **Mind the Fees** - Include realistic fees (0.1-0.5% per trade)
4. **Test Robustness** - Use walk-forward analysis
5. **Avoid Overfitting** - Don't optimize too many parameters

### Data Quality

1. **Clean Data** - Remove gaps, outliers, errors
2. **Consistent Timeframe** - Don't mix 1h and 4h data
3. **Sufficient History** - More data = better validation
4. **Realistic Volume** - Include volume for better simulation

### Risk Management

1. **Always Use Stops** - Protect against large losses
2. **Size Appropriately** - Don't risk more than 1-2% per trade
3. **Diversify** - Don't put all equity in one trade
4. **Monitor Drawdown** - Exit if drawdown exceeds tolerance

### Performance Analysis

1. **Look Beyond Returns** - Check Sharpe ratio, drawdown
2. **Count Trades** - Need enough trades for significance
3. **Check Distribution** - Are profits from one lucky trade?
4. **Compare to Baseline** - Beat buy-and-hold?

---

## Troubleshooting

### "Insufficient data" Error

**Problem:** Not enough candles for indicator calculation.

**Solution:** 
- Reduce indicator periods
- Use more historical data
- Check data file has enough rows

### "No trades executed"

**Problem:** Entry conditions never met.

**Solution:**
- Simplify entry conditions
- Check indicator values
- Verify data quality
- Test with different parameters

### "Invalid YAML" Error

**Problem:** Syntax error in strategy file.

**Solution:**
- Check indentation (use spaces, not tabs)
- Validate YAML syntax online
- Compare with example strategies

### Poor Performance

**Problem:** Strategy loses money or has low Sharpe ratio.

**Solution:**
- Test on different time periods
- Adjust risk management
- Try different indicators
- Consider market regime

---

## Need Help?

- 📖 [Full Documentation](../README.md)
- 💡 [Strategy Examples](../../strategies/examples/)
- 🔧 [Transform Guide](transforms.md)
- 📊 [Indicator Reference](indicators.md)
- ⚙️ [CLI Reference](cli.md)

---

**Happy Backtesting! 🚀**
