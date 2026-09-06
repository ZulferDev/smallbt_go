# CLI Guide: Data Transforms

## Overview

The `trader` CLI provides built-in support for data preprocessing transforms. Transforms are automatically detected from your strategy YAML and applied during backtest execution.

**Key Features:**
- ✅ Auto-detection from strategy YAML
- ✅ Validation before execution
- ✅ Progress reporting
- ✅ Backward compatible

---

## Commands

### validate-transforms

Validate transform configuration in a strategy YAML file.

**Syntax:**
```bash
trader validate-transforms --strategy <path>
```

**Example:**
```bash
trader validate-transforms --strategy strategies/examples/sma_cross_normalized.yaml
```

**Output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TRANSFORM VALIDATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy: strategies/examples/sma_cross_normalized.yaml
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy Name: sma_cross_normalized (v1)

Transforms: 2 in chain

1. Transform: normalize
   ✅ Valid configuration

2. Transform: smooth
   ✅ Valid configuration
   Parameters:
     - period: 5

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Transform validation complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Use Cases:**
- Validate strategy before running backtest
- Check transform parameters are correct
- Debug transform configuration issues
- Verify transforms are enabled

---

### backtest (with transforms)

Run backtest with automatic transform detection.

**Syntax:**
```bash
trader backtest --strategy <path> --data <path> [flags]
```

**Example:**
```bash
trader backtest \
  --strategy strategies/examples/sma_cross_normalized.yaml \
  --data data/BTCUSDT_4h.csv \
  --cash 10000
```

**Output with Transforms:**
```
📊 Transforms enabled: 2 in chain
   1. normalize
   2. smooth

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Running backtest...
Strategy: strategies/examples/sma_cross_normalized.yaml
Data:     data/BTCUSDT_4h.csv
Symbol:   BTCUSDT
Cash:     $10000.00
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
...
```

**Features:**
- Auto-detects transforms from strategy YAML
- Shows transform count and types before execution
- Applies transforms transparently
- No additional flags needed

---

## Strategy YAML Format

### Basic Structure

```yaml
strategy:
  name: my_strategy
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h
  
  # Optional transforms section
  transforms:
    enabled: true
    transforms:
      - type: normalize
        field: close
      
      - type: smooth
        field: volume
        params:
          period: 5

indicators:
  # ... your indicators
```

### Transform Configuration

**Required Fields:**
- `enabled` (bool): Enable/disable transforms
- `transforms` (array): List of transform specs

**Transform Spec:**
- `type` (string): Transform type (required)
- `field` (string): Target field (required)
- `params` (map): Transform parameters (optional)

---

## Available Transforms

### 1. scale

Multiply field values by a constant factor.

```yaml
- type: scale
  field: close
  params:
    factor: 0.001
```

**Parameters:**
- `factor` (float): Multiplication factor

**Use Case:** Unit conversion, magnitude reduction

---

### 2. normalize

Normalize field to standard range.

```yaml
- type: normalize
  field: close
```

**Parameters:** None (uses default normalization)

**Use Case:** Cross-asset strategies, price-agnostic indicators

---

### 3. log_returns

Convert prices to logarithmic returns.

```yaml
- type: log_returns
  field: close
```

**Parameters:** None

**Use Case:** Momentum analysis, statistical modeling

**Note:** First row will be lost (return not defined)

---

### 4. percentage_change

Calculate percentage change.

```yaml
- type: percentage_change
  field: close
```

**Parameters:** None

**Use Case:** Simple returns, rate of change

**Note:** First row will be lost

---

### 5. smooth

Apply moving average smoothing.

```yaml
- type: smooth
  field: volume
  params:
    period: 5
```

**Parameters:**
- `period` (int): Smoothing window size (required, >= 2)

**Use Case:** Noise reduction, data smoothing

**Note:** First `period-1` rows will be lost

---

## Enabling/Disabling Transforms

### Enabled

```yaml
transforms:
  enabled: true
  transforms:
    - type: normalize
      field: close
```

**Behavior:** Transforms are applied to data before strategy evaluation.

---

### Disabled

```yaml
transforms:
  enabled: false
  transforms:
    - type: normalize
      field: close
```

**Behavior:** Strategy receives raw, untransformed data. Transform definitions are ignored.

**Use Case:**
- Temporarily disable transforms without deleting config
- A/B testing (with vs without transforms)

---

### No Transforms

```yaml
data:
  symbol: BTCUSDT
  timeframe: 4h
  # No transforms section
```

**Behavior:** Strategy receives raw data. No overhead.

---

## Examples

### Example 1: Normalized SMA Crossover

**File:** `strategies/examples/sma_cross_normalized.yaml`

```yaml
data:
  symbol: BTCUSDT
  timeframe: 4h
  
  transforms:
    enabled: true
    transforms:
      - type: normalize
        field: close
      
      - type: smooth
        field: volume
        params:
          period: 5
```

**Run:**
```bash
trader backtest \
  --strategy strategies/examples/sma_cross_normalized.yaml \
  --data data/BTCUSDT_4h.csv
```

**Output:**
```
📊 Transforms enabled: 2 in chain
   1. normalize
   2. smooth
```

---

### Example 2: Log Returns Momentum

**File:** `strategies/examples/momentum_log_returns.yaml`

```yaml
data:
  symbol: BTCUSDT
  timeframe: 1h
  
  transforms:
    enabled: true
    transforms:
      - type: log_returns
        field: close
```

**Run:**
```bash
trader backtest \
  --strategy strategies/examples/momentum_log_returns.yaml \
  --data data/BTCUSDT_1h.csv
```

**Output:**
```
📊 Transforms enabled: 1 in chain
   1. log_returns
```

---

### Example 3: Validate Before Run

**Workflow:**
```bash
# Step 1: Validate transforms
trader validate-transforms --strategy my_strategy.yaml

# Step 2: If valid, run backtest
trader backtest --strategy my_strategy.yaml --data data.csv
```

**Why Validate First:**
- Catch configuration errors early
- Verify transform parameters
- Ensure transforms are enabled
- See what will be applied

---

## Troubleshooting

### Issue: Transform not applied

**Symptom:** No 📊 icon shown during backtest

**Possible Causes:**
1. `enabled: false` in YAML
2. No `transforms:` section in YAML
3. Empty `transforms:` array

**Solution:**
```bash
# Check configuration
trader validate-transforms --strategy my_strategy.yaml
```

---

### Issue: Invalid transform type

**Symptom:**
```
Error: unknown transform type: invalid_type
```

**Solution:** Check transform type spelling. Valid types:
- `scale`
- `normalize`
- `log_returns`
- `percentage_change`
- `smooth`

---

### Issue: Missing required field

**Symptom:**
```
Error: transform 1 (normalize): field is required
```

**Solution:** Add `field:` to transform spec:
```yaml
- type: normalize
  field: close  # Required!
```

---

### Issue: Invalid parameters

**Symptom:**
```
Error: transform 2 (smooth): period must be >= 2
```

**Solution:** Check parameter constraints:
```yaml
- type: smooth
  field: volume
  params:
    period: 5  # Must be >= 2
```

---

### Issue: All data filtered out

**Symptom:** Backtest reports "no candles after filtering"

**Possible Cause:** Transforms removed too many rows

**Example:**
```yaml
transforms:
  - type: log_returns
    field: close
  # Loses 1 row
  
  - type: smooth
    field: close
    params:
      period: 100
  # Loses 99 more rows
```

**Solution:**
- Reduce transform window sizes
- Use fewer transforms
- Ensure sufficient input data

---

## Best Practices

### 1. Validate First

Always validate before running long backtests:

```bash
trader validate-transforms --strategy strategy.yaml
```

### 2. Start Simple

Begin with one transform:

```yaml
transforms:
  enabled: true
  transforms:
    - type: normalize
      field: close
```

Add more transforms incrementally after validation.

### 3. Document Intent

Use YAML comments to explain why each transform is used:

```yaml
transforms:
  enabled: true
  transforms:
    # Normalize for cross-asset compatibility
    - type: normalize
      field: close
    
    # Smooth volume to reduce noise
    - type: smooth
      field: volume
      params:
        period: 5
```

### 4. Test Both Modes

Compare results with and without transforms:

```bash
# With transforms
trader backtest --strategy strategy.yaml --data data.csv

# Without transforms (set enabled: false in YAML)
trader backtest --strategy strategy.yaml --data data.csv
```

### 5. Check Data Loss

Transforms can reduce row count:

```yaml
# log_returns: loses 1 row
# smooth (period=10): loses 9 rows
# Total: loses 10 rows from start
```

Ensure sufficient data remains after transforms.

---

## Integration with Other Commands

### optimize

Transforms are applied during optimization:

```bash
trader optimize \
  --strategy strategy_with_transforms.yaml \
  --data data.csv \
  --parameters "indicators.sma_fast.period:5:20:1"
```

Each optimization iteration uses transformed data.

---

### walkforward

Transforms are applied to each window:

```bash
trader walkforward \
  --strategy strategy_with_transforms.yaml \
  --data data.csv \
  --train 1000 \
  --test 200
```

Training and testing windows both receive transformed data.

---

### paper

Future: Paper trading will use transforms:

```bash
trader paper \
  --strategy strategy_with_transforms.yaml \
  --symbol BTCUSDT
```

Real-time data will be transformed before strategy evaluation.

---

## Advanced Usage

### Multiple Fields

Transform different fields:

```yaml
transforms:
  enabled: true
  transforms:
    - type: normalize
      field: close
    
    - type: normalize
      field: volume
    
    - type: smooth
      field: high
      params:
        period: 3
```

### Chained Transforms

Apply multiple transforms to same field:

```yaml
transforms:
  enabled: true
  transforms:
    # Step 1: Convert to returns
    - type: log_returns
      field: close
    
    # Step 2: Smooth returns
    - type: smooth
      field: close
      params:
        period: 5
```

**Order matters!** Transforms are applied sequentially.

---

## Performance

### Overhead

**No Transforms:**
- Zero overhead
- Original data flow

**With Transforms:**
- Negligible overhead (~2-5%)
- One-time application at data load
- Cached for strategy evaluation

### Memory

**CSV with transforms:**
- Streaming mode: O(window_size)
- Typical: 1-2 MB

**Parquet with transforms:**
- Batch mode: O(dataset_size)
- Example: 100k candles = ~20 MB

---

## Migration Guide

### From Week 3 Format

**Old Format (Week 3):**
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

**New Format (CLI/Integration):**
```yaml
transforms:
  enabled: true
  transforms:
    - type: normalize
      field: close
```

**Changes:**
- `chain:` → `transforms:`
- `params.columns:` → `field:`
- Simplified parameters
- Single field per transform

---

## See Also

- [Transform Guide](transforms.md) - Detailed transform reference
- [Strategy Examples](../strategies/examples/) - Example strategies with transforms
- [Week 4 Day 2 Report](reports/phase17_week4_day2_backtest_integration.md) - Integration details

---

**Document Version:** 1.0  
**Last Updated:** 2026-09-06  
**Phase:** 17 Week 4 Day 3
