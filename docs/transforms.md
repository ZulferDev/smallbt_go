# Transform Guide

**Data preprocessing and transformation for advanced strategies**

---

## Table of Contents

1. [Introduction](#introduction)
2. [Transform Basics](#transform-basics)
3. [Statistical Transforms](#statistical-transforms)
4. [Smoothing Transforms](#smoothing-transforms)
5. [Time Series Transforms](#time-series-transforms)
6. [Outlier Handling](#outlier-handling)
7. [Transform Pipelines](#transform-pipelines)
8. [Best Practices](#best-practices)

---

## Introduction

Transforms preprocess your data **before** strategy execution. They enable:

- **Noise reduction** - Smooth price data
- **Normalization** - Compare different timeframes
- **Statistical analysis** - Z-scores, percentiles
- **Time series** - Stationarity, differencing
- **Outlier control** - Cap extreme values

### When to Use Transforms

✅ **Use transforms when:**
- Price data is noisy
- Need statistical indicators
- Comparing different symbols/timeframes
- Building mean reversion strategies
- Need trend removal

❌ **Don't use transforms when:**
- Strategy works on raw prices
- Need original price levels
- Unsure of mathematical implications
- Testing for first time (baseline first)

---

## Transform Basics

### Configuration

Add transforms to your strategy YAML:

```yaml
data:
  symbol: BTCUSDT
  timeframe: 1h
  
  transforms:
    enabled: true
    transforms:
      - type: zscore
        field: close
        params:
          window: 20
```

### Structure

Each transform has:
- **type**: Transform name
- **field**: Which OHLCV field to transform
- **params**: Transform-specific parameters

### Supported Fields

- `open` - Opening price
- `high` - Highest price
- `low` - Lowest price
- `close` - Closing price (most common)
- `volume` - Trading volume

### Execution Order

Transforms are applied sequentially:

```yaml
transforms:
  - type: smooth      # 1. Smooth first
  - type: difference  # 2. Then difference
  - type: zscore      # 3. Then normalize
```

---

## Statistical Transforms

### Z-Score Normalization

Converts values to standard deviations from mean.

**Formula:** `z = (x - μ) / σ`

**Configuration:**
```yaml
- type: zscore
  field: close
  params:
    window: 20  # Rolling window size
```

**Output:**
- `z = 0` → At mean
- `z > 2` → 2 std deviations above mean (overbought)
- `z < -2` → 2 std deviations below mean (oversold)

**Use Cases:**
- Mean reversion strategies
- Outlier detection
- Statistical arbitrage
- Comparing different timeframes

**Example Strategy:**
```yaml
data:
  transforms:
    - type: zscore
      field: close
      params:
        window: 20

indicators:
  rsi:
    type: rsi
    period: 14

entry:
  long:
    all:
      - lt: [close, -2]    # Z-score < -2 (oversold)
      - lt: [rsi, 30]      # RSI confirms
```

**Properties:**
- Parametric (assumes normal distribution)
- Unbounded output (-∞ to +∞)
- Warm-up period: window - 1 candles

---

### Percentile Rank

Converts values to percentile within rolling window (0-100%).

**Formula:** `percentile = (rank / (n-1)) * 100`

**Configuration:**
```yaml
- type: percentile_rank
  field: close
  params:
    window: 20
```

**Output:**
- `0%` → Minimum in window
- `50%` → Median
- `100%` → Maximum in window

**Use Cases:**
- Relative strength analysis
- Overbought/oversold (>80% / <20%)
- Rank-based strategies
- Distribution-agnostic normalization

**Example Strategy:**
```yaml
data:
  transforms:
    - type: percentile_rank
      field: close
      params:
        window: 20

entry:
  long:
    all:
      - lt: [close, 20]     # Below 20th percentile
      - gt: [volume, volume_avg]
```

**Properties:**
- Non-parametric (no distribution assumption)
- Bounded output (0-100)
- Warm-up period: window - 1 candles

**Comparison:**

| Aspect | Z-Score | Percentile |
|--------|---------|------------|
| Type | Parametric | Non-parametric |
| Range | Unbounded | 0-100% |
| Assumption | Normal dist | None |
| Outliers | Sensitive | Robust |

---

## Smoothing Transforms

### SMA Smoothing

Simple moving average smoothing.

**Configuration:**
```yaml
- type: smooth
  field: close
  params:
    period: 5
```

**Use Cases:**
- Noise reduction
- Trend clarity
- Stable signals

**Properties:**
- Equal weights
- More lag
- Better stability

---

### EMA Smoothing

Exponential moving average smoothing.

**Formula:** `EMA[t] = α * value[t] + (1-α) * EMA[t-1]`  
where `α = 2/(period+1)`

**Configuration:**
```yaml
- type: ema_smooth
  field: close
  params:
    period: 12
```

**Use Cases:**
- Noise reduction with responsiveness
- Trend following
- Less lag than SMA

**Example Strategy:**
```yaml
data:
  transforms:
    - type: ema_smooth
      field: close
      params:
        period: 12

indicators:
  ema_fast:
    type: ema
    source: close
    period: 5
  
  ema_slow:
    type: ema
    source: close
    period: 13

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
```

**Properties:**
- Exponential weights (recent data emphasized)
- Less lag than SMA
- No warm-up period

**Comparison:**

| Aspect | SMA | EMA |
|--------|-----|-----|
| Weights | Equal | Exponential |
| Lag | More | Less |
| Responsiveness | Lower | Higher |
| Best For | Stability | Trend detection |

---

## Time Series Transforms

### Differencing

Calculates change between consecutive values.

**Formula:** `diff[t] = value[t] - value[t-1]`

**Configuration:**
```yaml
- type: difference
  field: close
  params:
    order: 1  # 1=velocity, 2=acceleration, 3=jerk
```

**Order Meanings:**
- **Order 1**: Velocity (price change)
- **Order 2**: Acceleration (change of change)
- **Order 3**: Jerk (change of acceleration)

**Use Cases:**
- Remove trends
- Make time series stationary
- Momentum strategies
- Rate of change analysis

**Example Strategy:**
```yaml
data:
  transforms:
    - type: difference
      field: close
      params:
        order: 1  # First-order: velocity

indicators:
  momentum_avg:
    type: sma
    source: close  # Already differenced
    period: 10

entry:
  long:
    all:
      - gt: [close, 0]              # Positive momentum
      - gt: [close, momentum_avg]   # Accelerating
```

**Properties:**
- First difference removes linear trends
- Second difference removes quadratic trends
- Invertible (cumsum reverses)
- No warm-up period

**Mathematical Properties:**
```
Original prices: 100, 105, 110, 115, 120
1st difference:  0,   5,   5,   5,   5    (velocity)
2nd difference:  0,   5,   0,   0,   0    (acceleration)
```

---

### Log Returns

Logarithmic returns for percentage changes.

**Formula:** `log_return[t] = log(value[t] / value[t-1])`

**Configuration:**
```yaml
- type: log_returns
  field: close
```

**Use Cases:**
- Percentage change analysis
- Statistical properties
- Comparing different symbols

**Properties:**
- Time-additive
- Symmetric for gains/losses
- Requires positive prices

---

## Outlier Handling

### Clipping

Caps values to [min, max] range.

**Configuration:**
```yaml
- type: clip
  field: close
  params:
    min: 40000
    max: 60000
```

**Use Cases:**
- Remove extreme outliers
- Range limiting
- Robust signal generation
- Prevent flash crash impact

**Example Strategy:**
```yaml
data:
  transforms:
    # Clip outliers first
    - type: clip
      field: close
      params:
        min: 40000
        max: 60000
    
    # Then smooth
    - type: ema_smooth
      field: close
      params:
        period: 10

indicators:
  ema_fast:
    type: ema
    period: 9
  
  ema_slow:
    type: ema
    period: 21

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
```

**Properties:**
- Idempotent: `clip(clip(x)) = clip(x)`
- Preserves order
- No warm-up period
- Fast (O(n))

---

## Transform Pipelines

### Multi-Stage Processing

Combine transforms for powerful preprocessing:

### Pipeline 1: Statistical Momentum

```yaml
transforms:
  - type: difference      # Calculate velocity
    field: close
    params:
      order: 1
  
  - type: zscore         # Normalize to z-scores
    field: close
    params:
      window: 20
```

**Result:** Z-scored velocity (statistical momentum signal)

---

### Pipeline 2: Robust Trend Following

```yaml
transforms:
  - type: clip           # Remove outliers first
    field: close
    params:
      min: 40000
      max: 60000
  
  - type: ema_smooth     # Then smooth
    field: close
    params:
      period: 12
```

**Result:** Outlier-resistant, smoothed prices

---

### Pipeline 3: Relative Strength with Outlier Control

```yaml
transforms:
  - type: clip           # Cap extremes
    field: close
    params:
      min: 0
      max: 100000
  
  - type: percentile_rank  # Rank normalize
    field: close
    params:
      window: 20
```

**Result:** Outlier-resistant percentile ranks

---

### Pipeline 4: Stationary Statistical Analysis

```yaml
transforms:
  - type: difference     # Make stationary
    field: close
    params:
      order: 1
  
  - type: smooth         # Smooth velocity
    field: close
    params:
      period: 5
  
  - type: percentile_rank  # Rank normalize
    field: close
    params:
      window: 20
```

**Result:** Percentile-ranked smoothed velocity

---

## Best Practices

### 1. Start Simple

```yaml
# ❌ DON'T start with complex pipelines
transforms:
  - type: clip
  - type: smooth
  - type: difference
  - type: zscore

# ✅ DO start simple, add complexity gradually
transforms:
  - type: smooth
    field: close
    params:
      period: 5
```

### 2. Understand Mathematical Implications

```yaml
# ❌ DON'T use transforms you don't understand
transforms:
  - type: zscore  # What does z-score mean for my strategy?

# ✅ DO understand the transformation
# Z-score converts to standard deviations from mean
# Useful for mean reversion, outlier detection
```

### 3. Test With and Without

Always compare:
1. Baseline (no transforms)
2. With transforms

```bash
# Baseline
./trader backtest --strategy baseline.yaml --data data.csv

# With transforms
./trader backtest --strategy with_transforms.yaml --data data.csv
```

### 4. Mind the Warm-up Period

Transforms with windows need warm-up:

```yaml
# This needs 20 candles before valid signals
- type: zscore
  field: close
  params:
    window: 20
```

Ensure your data has enough history.

### 5. Validate Transform Output

Check transformed values make sense:

```bash
# Enable verbose logging
./trader backtest --strategy my_strategy.yaml --data data.csv --verbose
```

### 6. Consider Order

Transform order matters:

```yaml
# Different results!
# Option A: Smooth then normalize
transforms:
  - type: smooth
  - type: zscore

# Option B: Normalize then smooth
transforms:
  - type: zscore
  - type: smooth
```

Generally: Clean → Transform → Normalize

### 7. Use Appropriate Fields

```yaml
# ✅ Good: Close price for signals
- type: zscore
  field: close

# ⚠️ Careful: Volume has different properties
- type: zscore
  field: volume  # Ensure this makes sense for your strategy
```

---

## Common Patterns

### Pattern 1: Noise Reduction

```yaml
transforms:
  - type: ema_smooth
    field: close
    params:
      period: 5
```

### Pattern 2: Mean Reversion

```yaml
transforms:
  - type: zscore
    field: close
    params:
      window: 20
```

### Pattern 3: Momentum

```yaml
transforms:
  - type: difference
    field: close
    params:
      order: 1
```

### Pattern 4: Robust Signals

```yaml
transforms:
  - type: clip
    field: close
    params:
      min: 40000
      max: 60000
  
  - type: percentile_rank
    field: close
    params:
      window: 20
```

---

## Troubleshooting

### "Insufficient data" Error

**Problem:** Not enough candles for transform window.

**Solution:**
- Reduce window size
- Use more historical data
- Check transform requirements

### Unexpected Results

**Problem:** Strategy behaves strangely after adding transforms.

**Solution:**
- Check transformed values
- Verify indicator calculations
- Compare with non-transformed baseline
- Ensure indicators use correct source

### Performance Degradation

**Problem:** Strategy performs worse with transforms.

**Solution:**
- Transforms may not suit your strategy
- Try simpler transforms
- Test on different time periods
- Remove transforms, focus on raw signals

---

## Transform Reference

| Transform | Type | Window | Output Range | Use Case |
|-----------|------|--------|--------------|----------|
| `zscore` | Statistical | Yes | Unbounded | Mean reversion |
| `percentile_rank` | Statistical | Yes | 0-100% | Relative strength |
| `smooth` | Smoothing | Yes | Same as input | Noise reduction |
| `ema_smooth` | Smoothing | No | Same as input | Responsive smoothing |
| `difference` | Time Series | No | Same as input | Stationarity, momentum |
| `clip` | Outlier | No | [min, max] | Outlier control |
| `normalize` | Statistical | No | 0-1 | Min-max scaling |
| `log_returns` | Time Series | No | Unbounded | Percentage change |
| `scale` | Simple | No | Scaled | Multiply by factor |

---

## Further Reading

- [Getting Started Guide](getting-started.md)
- [Indicator Reference](indicators.md)
- [Strategy Examples](../../strategies/examples/)
- [Phase 17 Documentation](../reports/phase17_complete.md)
- [Phase 18 Documentation](../reports/phase_18_complete.md)

---

**Master transforms for powerful preprocessing! 🔄**
