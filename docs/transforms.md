# Data Transforms in Strategy Configuration

## Overview

The backtest engine supports optional data preprocessing transforms that are applied to market data before strategy evaluation. Transforms enable normalization, scaling, smoothing, and statistical transformations to improve strategy robustness and cross-asset applicability.

## Configuration

Add transforms to your strategy YAML under `data.transforms`:

```yaml
data:
  symbol: BTCUSDT
  timeframe: 4h
  
  transforms:
    enabled: true
    chain:
      - type: normalize
        params:
          columns: [close]
          method: minmax
          window: 100
      
      - type: smooth
        params:
          columns: [volume]
          window: 5
          method: sma
```

## Available Transforms

### 1. Scale

Multiply column values by a constant factor.

```yaml
- type: scale
  params:
    columns: [close, high, low]
    factor: 0.001
```

**Use cases:**
- Convert between units (e.g., cents to dollars)
- Reduce numerical magnitude for stability
- Prepare data for ML models

### 2. Normalize

Normalize data to standard range or distribution.

**Min-Max Normalization** (0 to 1 range):
```yaml
- type: normalize
  params:
    columns: [close]
    method: minmax
    window: 100
```

**Z-Score Normalization** (mean=0, std=1):
```yaml
- type: normalize
  params:
    columns: [close]
    method: zscore
    window: 50
```

**Use cases:**
- Cross-asset strategies (BTC vs ETH at different price levels)
- Momentum comparison across instruments
- Statistical analysis requiring standard distributions
- Indicator stability across price ranges

**Parameters:**
- `columns`: List of columns to normalize
- `method`: `minmax` or `zscore`
- `window`: Rolling window size for statistics

### 3. Log Returns

Convert prices to logarithmic returns.

```yaml
- type: log_returns
  params:
    columns: [close]
    periods: 1
```

**Use cases:**
- Momentum and trend analysis
- Better statistical properties (stationarity)
- Correct handling of compounding
- Time series modeling

**Parameters:**
- `columns`: Price columns to convert
- `periods`: Number of periods for return calculation (default: 1)

**Note:** First `periods` rows become NaN and are filtered out.

### 4. Percentage Change

Calculate percentage change over periods.

```yaml
- type: percentage_change
  params:
    columns: [close]
    periods: 1
```

**Use cases:**
- Simple return calculation
- Momentum indicators
- Rate of change analysis

**Parameters:**
- `columns`: Columns to compute percentage change
- `periods`: Number of periods (default: 1)

**Note:** First `periods` rows become NaN and are filtered out.

### 5. Smooth

Apply smoothing to reduce noise.

```yaml
- type: smooth
  params:
    columns: [volume, close]
    window: 5
    method: sma
```

**Use cases:**
- Reduce volume noise
- Smooth price data for cleaner signals
- Prepare data for indicator calculation

**Parameters:**
- `columns`: Columns to smooth
- `window`: Smoothing window size
- `method`: Currently only `sma` supported

**Note:** First `window-1` rows become NaN and are filtered out.

## Transform Chain Execution

Transforms are applied sequentially in the order specified:

```yaml
transforms:
  enabled: true
  chain:
    - type: log_returns      # Step 1: Convert to returns
      params:
        columns: [close]
        periods: 1
    
    - type: normalize        # Step 2: Normalize returns
      params:
        columns: [close]
        method: zscore
        window: 50
    
    - type: smooth           # Step 3: Smooth normalized returns
      params:
        columns: [close]
        window: 3
        method: sma
```

**Execution:**
1. Load raw candles
2. Apply log_returns → close becomes log returns
3. Apply normalize → close becomes normalized z-scores
4. Apply smooth → close becomes smoothed z-scores
5. Strategy receives final transformed data

## Important Considerations

### Data Loss

Some transforms remove initial rows:
- `log_returns` with `periods: n` → lose first `n` rows
- `percentage_change` with `periods: n` → lose first `n` rows  
- `smooth` with `window: n` → lose first `n-1` rows
- `normalize` with `window: n` → lose first `n-1` rows

**Impact:** Your backtest will start later than the raw data start time.

### Indicator Warm-Up

Transforms are applied **before** indicator calculation:
- Indicators see transformed data
- Indicator warm-up periods apply to transformed data
- Total warm-up = transform data loss + indicator periods

**Example:**
```yaml
transforms:
  chain:
    - type: log_returns
      params:
        periods: 1        # Lose 1 row

indicators:
  sma_20:
    type: sma
    period: 20          # Need 20 rows for first value

# Total: Need 21 rows before first signal
```

### Column Semantics

After transforms, column meanings change:

**Original:**
- `close` = closing price (e.g., 50000.0)
- `volume` = trade volume (e.g., 123.45)

**After log_returns:**
- `close` = log return (e.g., 0.002)
- `volume` = unchanged

**After normalize (minmax):**
- `close` = normalized value in [0, 1] (e.g., 0.745)
- `volume` = unchanged unless specified

**Impact:** Be mindful when configuring:
- Stop losses (percentage vs absolute)
- Position sizing
- Risk management
- Exit conditions

## Performance

### Streaming Mode (CSV)

CSV data feeds use lazy evaluation:
```
CSV Reader → Transform Pipeline → Strategy
```
- Memory efficient
- Processes one candle at a time
- Suitable for large datasets

### Batch Mode (Parquet)

Parquet loads entire dataset first:
```
Parquet Reader → Load All → Apply Transforms → Strategy
```
- Faster for small/medium datasets
- Higher memory usage
- Future: Streaming parquet support planned

## Disabling Transforms

Set `enabled: false` to disable without removing config:

```yaml
transforms:
  enabled: false
  chain:
    - type: normalize
      params:
        columns: [close]
        method: minmax
        window: 100
```

Strategy will receive raw, untransformed data.

## Examples

### Example 1: Cross-Asset Strategy

Normalize prices for strategies that work across BTC, ETH, SOL:

```yaml
data:
  transforms:
    enabled: true
    chain:
      - type: normalize
        params:
          columns: [close, high, low]
          method: minmax
          window: 100
```

Indicators now work on normalized [0,1] range regardless of asset price.

### Example 2: Statistical Momentum

Use z-scores for mean reversion:

```yaml
data:
  transforms:
    enabled: true
    chain:
      - type: normalize
        params:
          columns: [close]
          method: zscore
          window: 50

entry:
  long:
    all:
      - lt: [close, -2]  # Buy when 2 std devs below mean
```

### Example 3: Returns-Based Strategy

Work with log returns instead of prices:

```yaml
data:
  transforms:
    enabled: true
    chain:
      - type: log_returns
        params:
          columns: [close]
          periods: 1
      
      - type: normalize
        params:
          columns: [close]
          method: zscore
          window: 20

indicators:
  momentum:
    type: sma
    source: close
    period: 10

entry:
  long:
    all:
      - gt: [momentum, 0.5]  # Positive normalized momentum
```

## Validation

The engine validates transforms before execution:

**Valid:**
```yaml
transforms:
  enabled: true
  chain:
    - type: normalize
      params:
        columns: [close]
        method: minmax
        window: 100
```

**Invalid - Will Error:**
```yaml
transforms:
  enabled: true
  chain:
    - type: normalize
      params:
        columns: [close]
        method: invalid_method  # ❌ Unknown method
        window: 100
```

**Invalid - Will Error:**
```yaml
transforms:
  enabled: true
  chain:
    - type: unknown_transform  # ❌ Unknown transform type
      params:
        columns: [close]
```

Clear error messages guide you to fix configuration issues.

## Testing Your Transforms

1. **Validate First:**
```bash
trader validate --strategy your_strategy.yaml
```

2. **Small Backtest:**
```bash
trader backtest --strategy your_strategy.yaml --data small_sample.csv
```

3. **Check Output:**
- Verify trade count is reasonable
- Check if transforms caused excessive data loss
- Validate strategy logic still makes sense on transformed data

4. **Compare:**
Run same strategy with and without transforms to understand impact.

## Best Practices

1. **Start Simple:** Use one transform at a time initially
2. **Understand Data Loss:** Account for warm-up period reduction
3. **Validate Logic:** Ensure strategy conditions make sense on transformed data
4. **Document Intent:** Add comments explaining why each transform is used
5. **Test Both Modes:** Verify strategy works with and without transforms
6. **Monitor Warm-Up:** Ensure sufficient data after transform data loss

## Future Enhancements

Planned features:
- Custom transform functions
- Multi-symbol transforms (cointegration, correlation)
- Conditional transforms (apply only when conditions met)
- Transform parameter optimization
- Streaming parquet support
- More smoothing methods (EMA, Kalman)

## See Also

- Strategy examples: `strategies/examples/sma_cross_normalized.yaml`
- Strategy examples: `strategies/examples/momentum_log_returns.yaml`
- Transform implementation: `internal/data/transform/`
- Integration guide: Week 4 Day 2 report
