# Backtesting Guide

## Overview

Backtesting simulates strategy performance on historical data with realistic execution modeling.

## Basic Backtest

```bash
./trader backtest \
  --strategy strategy.yaml \
  --data data/BTCUSDT_1h.csv \
  --output results.json
```

## Data Format

### CSV Format

Required columns:
```
timestamp,open,high,low,close,volume
```

Example:
```csv
timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,42000.0,42500.0,41800.0,42300.0,1500.5
2024-01-01T01:00:00Z,42300.0,42800.0,42100.0,42600.0,1800.2
```

### Timestamp Format

Supported formats:
- ISO8601: `2024-01-01T00:00:00Z`
- Unix timestamp: `1704067200`
- RFC3339: `2024-01-01T00:00:00+00:00`

All times are handled in UTC internally.

## Execution Model

### Order Types

#### Market Order
Executed immediately at next available price:
```yaml
entry:
  long:
    order_type: market
```

#### Limit Order
Executed when price reaches limit:
```yaml
entry:
  long:
    order_type: limit
    limit_price: close * 0.99  # 1% below close
```

#### Stop Order
Triggered when price reaches stop, executed as market:
```yaml
exit:
  long:
    order_type: stop
    stop_price: entry_price * 0.95
```

### Fees

Configure maker/taker fees:
```yaml
execution:
  fees:
    maker: 0.0002  # 0.02%
    taker: 0.0005  # 0.05%
```

Market orders pay taker fees. Limit orders pay maker fees if they rest in the order book.

### Slippage

Simulate price slippage:
```yaml
execution:
  slippage:
    type: percentage
    value: 0.0005  # 0.05% slippage
```

Slippage models:
- `percentage`: Fixed percentage of price
- `fixed`: Fixed amount in quote currency
- `volatility`: Based on ATR (future)

### Intrabar Execution

When both stop loss and take profit are within the same candle:

**Policy**: Worst-case (conservative)
- For long: Stop loss triggers first
- For short: Stop loss triggers first

This prevents overly optimistic results.

## Portfolio Management

### Initial Capital

```bash
./trader backtest \
  --strategy strategy.yaml \
  --data data.csv \
  --initial-capital 10000
```

Default: $10,000

### Position Sizing

See [Strategy DSL - Risk Management](STRATEGY_DSL.md#risk-management)

### Multiple Positions

By default, only one position per symbol is allowed. Future versions may support multiple concurrent positions.

## Result Analysis

### Metrics

The backtest produces comprehensive metrics:

#### Returns
- **Total Return**: Overall percentage return
- **CAGR**: Compound Annual Growth Rate
- **Best Trade**: Largest winning trade
- **Worst Trade**: Largest losing trade
- **Average Trade**: Mean trade return

#### Risk
- **Max Drawdown**: Largest peak-to-trough decline
- **Sharpe Ratio**: Risk-adjusted return (annualized)
- **Sortino Ratio**: Downside risk-adjusted return
- **Calmar Ratio**: CAGR / Max Drawdown
- **Volatility**: Standard deviation of returns

#### Trade Statistics
- **Total Trades**: Number of completed trades
- **Win Rate**: Percentage of winning trades
- **Profit Factor**: Gross profit / Gross loss
- **Expectancy**: Average profit per trade
- **Average Win**: Mean winning trade
- **Average Loss**: Mean losing trade
- **Largest Win**: Best single trade
- **Largest Loss**: Worst single trade

#### Exposure
- **Time in Market**: Percentage of time with open position
- **Average Trade Duration**: Mean holding period

### Output Formats

#### JSON
```bash
./trader backtest --output results.json
```

Machine-readable format with all metrics and trade history.

#### CSV
```bash
./trader backtest --output results.csv
```

Tabular format for spreadsheet analysis.

### Equity Curve

The equity curve is included in JSON output:

```json
{
  "equity_curve": [
    {"timestamp": "2024-01-01T00:00:00Z", "equity": 10000.0, "drawdown": 0.0},
    {"timestamp": "2024-01-02T00:00:00Z", "equity": 10500.0, "drawdown": 0.0}
  ]
}
```

### Trade Journal

Every trade is recorded:

```json
{
  "trades": [
    {
      "id": "trade-001",
      "symbol": "BTCUSDT",
      "side": "long",
      "entry_time": "2024-01-01T00:00:00Z",
      "entry_price": 42000.0,
      "exit_time": "2024-01-02T00:00:00Z",
      "exit_price": 43000.0,
      "quantity": 0.5,
      "gross_pnl": 500.0,
      "fees": 5.0,
      "net_pnl": 495.0,
      "return": 0.0238,
      "exit_reason": "stop_loss"
    }
  ]
}
```

## Validation

### Strategy Validation

Always validate before backtesting:

```bash
./trader validate --strategy strategy.yaml
```

This checks:
- YAML syntax
- Required fields
- Indicator dependencies
- Invalid parameters
- Circular references

### Data Validation

The engine validates data quality:
- Chronological order
- No duplicate timestamps
- Valid OHLC relationships (high >= close >= low)
- No negative values
- No missing required fields

Invalid data produces clear error messages.

## Determinism

Backtests are deterministic by default:
- Same strategy + same data = same results
- No random behavior without explicit seed
- Reproducible across runs

For Monte Carlo with randomization:
```bash
./trader montecarlo --seed 42
```

## Look-Ahead Bias Prevention

The engine prevents look-ahead bias:
- Indicators use only past data
- Conditions evaluated chronologically
- No future data access
- Historical references explicit (`previous`, `shift`)

## Common Pitfalls

### 1. Insufficient Warmup Period

Indicators need warmup data:
```yaml
indicators:
  ema200:
    type: ema
    period: 200  # Needs 200 bars before valid
```

First 200 bars may not generate signals.

### 2. Overfitting

Optimizing on the same data you test on leads to overfitting. Use Walk Forward Analysis:

```bash
./trader walkforward --strategy strategy.yaml --data data.csv
```

### 3. Ignoring Fees and Slippage

Always include realistic fees and slippage:
```yaml
execution:
  fees:
    maker: 0.0002
    taker: 0.0005
  slippage:
    type: percentage
    value: 0.0005
```

### 4. Survivor Bias

Testing only on successful assets (like BTC) ignores failed assets. Test across multiple assets where possible.

### 5. Small Sample Size

A strategy with 10 trades has insufficient statistical significance. Aim for 100+ trades.

## Best Practices

### 1. Out-of-Sample Testing

Always reserve data for out-of-sample testing:
- Train: 70% of data
- Test: 30% of data

Or use Walk Forward Analysis.

### 2. Multiple Timeframes

Test strategies across different timeframes:
```bash
for tf in 1h 4h 1d; do
  ./trader backtest --strategy strategy_${tf}.yaml --data data_${tf}.csv
done
```

### 3. Multiple Assets

Test across multiple symbols to verify robustness.

### 4. Stress Testing

Test during different market conditions:
- Bull markets
- Bear markets
- Sideways markets
- High volatility
- Low volatility

### 5. Monte Carlo Analysis

Verify statistical robustness:
```bash
./trader montecarlo \
  --strategy strategy.yaml \
  --data data.csv \
  --simulations 1000
```

### 6. Parameter Sensitivity

Test parameter variations:
```bash
./trader optimize \
  --strategy strategy.yaml \
  --data data.csv
```

## Example Workflow

```bash
# 1. Validate strategy
./trader validate --strategy ema_cross.yaml

# 2. Initial backtest
./trader backtest \
  --strategy ema_cross.yaml \
  --data BTCUSDT_1h.csv \
  --output initial_results.json

# 3. Optimize parameters
./trader optimize \
  --strategy ema_cross.yaml \
  --data BTCUSDT_1h.csv

# 4. Walk Forward Analysis
./trader walkforward \
  --strategy ema_cross_optimized.yaml \
  --data BTCUSDT_1h.csv

# 5. Monte Carlo validation
./trader montecarlo \
  --strategy ema_cross_optimized.yaml \
  --data BTCUSDT_1h.csv \
  --simulations 1000

# 6. Review results
cat results.json | jq .
```

## Troubleshooting

### No Trades Generated

Check:
1. Sufficient warmup data for indicators
2. Conditions are achievable
3. Position sizing allows trades (not too small/large)
4. Data covers expected signal periods

### Unrealistic Results

Check:
1. Fees and slippage configured
2. No look-ahead bias
3. Sufficient data
4. Realistic position sizing

### Errors During Backtest

Check:
1. Data format matches CSV specification
2. All indicator dependencies satisfied
3. Valid strategy YAML
4. Sufficient system resources

## Performance Tips

### 1. Data Format

Parquet files load faster than CSV for large datasets (future support).

### 2. Indicator Caching

The engine automatically caches indicator values. No manual optimization needed.

### 3. Parallel Backtests

Run multiple backtests in parallel:
```bash
./trader backtest --strategy strategy1.yaml --data data.csv &
./trader backtest --strategy strategy2.yaml --data data.csv &
wait
```

### 4. Memory Usage

For very large datasets, consider:
- Splitting data into chunks
- Using streaming data feeds (future)
- Reducing indicator lookback periods where possible