# Indicator Reference

**Complete guide to built-in technical indicators**

---

## Table of Contents

1. [Overview](#overview)
2. [Trend Indicators](#trend-indicators)
3. [Momentum Indicators](#momentum-indicators)
4. [Volatility Indicators](#volatility-indicators)
5. [Volume Indicators](#volume-indicators)
6. [Composite Indicators](#composite-indicators)
7. [Custom Indicators](#custom-indicators)

---

## Overview

Indicators are technical analysis tools that derive values from OHLCV data. smallbt_go provides:

- **Built-in indicators** - SMA, EMA, RSI, ATR, etc.
- **Composite indicators** - Combine multiple indicators
- **Custom indicators** - Extend with your own

### Basic Configuration

```yaml
indicators:
  indicator_name:
    type: sma
    source: close
    period: 20
```

### Common Parameters

- **type**: Indicator type (required)
- **source**: Data field (open/high/low/close/volume)
- **period**: Lookback window (most indicators)

---

## Trend Indicators

### Simple Moving Average (SMA)

Average price over a period.

**Formula:** `SMA = (sum of prices over period) / period`

**Configuration:**
```yaml
indicators:
  sma_20:
    type: sma
    source: close
    period: 20
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source` | string | Yes | Data field (close, open, etc.) |
| `period` | int | Yes | Lookback period |

**Use Cases:**
- Trend identification
- Support/resistance levels
- Entry/exit signals
- Crossover strategies

**Example Strategy:**
```yaml
indicators:
  sma_fast:
    type: sma
    source: close
    period: 10
  
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
```

**Properties:**
- Warm-up period: period - 1 candles
- Lag: High (equal weight to all prices)
- Smoothness: High

---

### Exponential Moving Average (EMA)

Weighted average emphasizing recent prices.

**Formula:** `EMA[t] = α * price[t] + (1-α) * EMA[t-1]`  
where `α = 2 / (period + 1)`

**Configuration:**
```yaml
indicators:
  ema_12:
    type: ema
    source: close
    period: 12
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source` | string | Yes | Data field |
| `period` | int | Yes | Lookback period |

**Use Cases:**
- Trend following with less lag
- Dynamic support/resistance
- Faster crossover signals
- Short-term trend detection

**Example Strategy:**
```yaml
indicators:
  ema_9:
    type: ema
    source: close
    period: 9
  
  ema_21:
    type: ema
    source: close
    period: 21
  
  ema_200:
    type: ema
    source: close
    period: 200

entry:
  long:
    all:
      - cross_above: [ema_9, ema_21]
      - gt: [close, ema_200]  # Above long-term trend
```

**Properties:**
- Warm-up period: None (can start immediately)
- Lag: Lower than SMA
- Smoothness: Medium
- Responsiveness: Higher than SMA

**Comparison: SMA vs EMA**

| Aspect | SMA | EMA |
|--------|-----|-----|
| Weights | Equal | Exponential |
| Lag | More | Less |
| Responsiveness | Lower | Higher |
| Best For | Smooth trends | Quick signals |

---

## Momentum Indicators

### Relative Strength Index (RSI)

Measures momentum on a 0-100 scale.

**Formula:**
```
RS = Average Gain / Average Loss
RSI = 100 - (100 / (1 + RS))
```

**Configuration:**
```yaml
indicators:
  rsi_14:
    type: rsi
    source: close
    period: 14
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source` | string | Yes | Data field (typically close) |
| `period` | int | Yes | Lookback period (14 is standard) |

**Interpretation:**

| Range | Condition | Meaning |
|-------|-----------|---------|
| 0-30 | Oversold | Potential buy signal |
| 30-70 | Neutral | Normal range |
| 70-100 | Overbought | Potential sell signal |

**Use Cases:**
- Overbought/oversold conditions
- Divergence trading
- Trend strength confirmation
- Mean reversion strategies

**Example Strategy:**
```yaml
indicators:
  rsi:
    type: rsi
    source: close
    period: 14
  
  ema_50:
    type: ema
    source: close
    period: 50

entry:
  long:
    all:
      - lt: [rsi, 30]          # Oversold
      - gt: [close, ema_50]    # Above trend

exit:
  long:
    any:
      - gt: [rsi, 70]          # Overbought
```

**Advanced Strategy (Divergence):**
```yaml
# Note: Divergence requires state tracking (advanced)
indicators:
  rsi:
    type: rsi
    period: 14

entry:
  long:
    all:
      - lt: [rsi, 30]                    # Oversold
      - rising: [rsi]                    # RSI rising
      # price making lower low, RSI making higher low = bullish divergence
```

**Properties:**
- Warm-up period: period candles
- Range: 0-100 (bounded)
- Oscillator: Yes
- Best period: 14 (standard), 7 (fast), 21 (slow)

---

### MACD (Moving Average Convergence Divergence)

Shows relationship between two moving averages.

**Formula:**
```
MACD Line = EMA(12) - EMA(26)
Signal Line = EMA(9) of MACD Line
Histogram = MACD Line - Signal Line
```

**Configuration:**
```yaml
indicators:
  macd:
    type: macd
    source: close
    fast_period: 12
    slow_period: 26
    signal_period: 9
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source` | string | Yes | Data field |
| `fast_period` | int | No | Fast EMA (default: 12) |
| `slow_period` | int | No | Slow EMA (default: 26) |
| `signal_period` | int | No | Signal line (default: 9) |

**Output Fields:**
- `macd.line` - MACD line
- `macd.signal` - Signal line
- `macd.histogram` - Histogram

**Use Cases:**
- Trend direction
- Momentum strength
- Crossover signals
- Divergence detection

**Example Strategy:**
```yaml
indicators:
  macd:
    type: macd
    source: close

entry:
  long:
    all:
      - cross_above: [macd.line, macd.signal]
      - gt: [macd.histogram, 0]

exit:
  long:
    any:
      - cross_below: [macd.line, macd.signal]
```

**Properties:**
- Warm-up period: slow_period candles
- Unbounded: Yes
- Lagging: Medium

---

## Volatility Indicators

### Average True Range (ATR)

Measures market volatility.

**Formula:**
```
True Range = max(
  high - low,
  abs(high - previous_close),
  abs(low - previous_close)
)
ATR = Average of True Range over period
```

**Configuration:**
```yaml
indicators:
  atr_14:
    type: atr
    period: 14
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `period` | int | Yes | Lookback period (14 is standard) |

**Use Cases:**
- Position sizing
- Stop loss placement
- Volatility filtering
- Breakout confirmation

**Example Strategy (ATR Stop Loss):**
```yaml
indicators:
  ema_20:
    type: ema
    source: close
    period: 20
  
  atr:
    type: atr
    period: 14

entry:
  long:
    all:
      - cross_above: [close, ema_20]

risk:
  stop_loss:
    type: atr
    indicator: atr
    multiplier: 1.5  # 1.5x ATR stop

  take_profit:
    type: risk_reward
    ratio: 2
```

**Example Strategy (Volatility Filter):**
```yaml
indicators:
  atr:
    type: atr
    period: 14
  
  atr_avg:
    type: sma
    source: atr  # Can use indicators as source
    period: 20

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [atr, atr_avg]  # Only trade when volatility is high
```

**Properties:**
- Warm-up period: period candles
- Always positive
- Not directional
- Unit: Same as price

---

### Bollinger Bands

Volatility bands around moving average.

**Formula:**
```
Middle Band = SMA(period)
Upper Band = Middle + (std_dev * multiplier)
Lower Band = Middle - (std_dev * multiplier)
```

**Configuration:**
```yaml
indicators:
  bollinger:
    type: bollinger
    source: close
    period: 20
    std_dev: 2
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source` | string | Yes | Data field |
| `period` | int | Yes | SMA period (20 is standard) |
| `std_dev` | float | No | Standard deviations (default: 2) |

**Output Fields:**
- `bollinger.upper` - Upper band
- `bollinger.middle` - Middle band (SMA)
- `bollinger.lower` - Lower band

**Use Cases:**
- Overbought/oversold
- Volatility breakouts
- Mean reversion
- Squeeze patterns

**Example Strategy (Mean Reversion):**
```yaml
indicators:
  bollinger:
    type: bollinger
    source: close
    period: 20
    std_dev: 2
  
  rsi:
    type: rsi
    period: 14

entry:
  long:
    all:
      - lt: [close, bollinger.lower]  # Below lower band
      - lt: [rsi, 30]                 # Oversold

exit:
  long:
    any:
      - gt: [close, bollinger.middle]  # Return to middle
```

**Example Strategy (Breakout):**
```yaml
indicators:
  bollinger:
    type: bollinger
    period: 20
    std_dev: 2
  
  volume_avg:
    type: sma
    source: volume
    period: 20

entry:
  long:
    all:
      - gt: [close, bollinger.upper]      # Break above
      - gt: [volume, volume_avg * 1.5]    # Volume confirmation
```

**Properties:**
- Warm-up period: period candles
- Adaptive: Bands widen in volatile markets
- Bounded: Contains ~95% of prices (2 std dev)

---

## Volume Indicators

### Volume SMA

Average volume over period.

**Configuration:**
```yaml
indicators:
  volume_avg:
    type: sma
    source: volume
    period: 20
```

**Use Cases:**
- Volume confirmation
- Breakout validation
- Liquidity assessment

**Example Strategy:**
```yaml
indicators:
  ema_fast:
    type: ema
    period: 9
  
  ema_slow:
    type: ema
    period: 21
  
  volume_avg:
    type: sma
    source: volume
    period: 20

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]
      - gt: [volume, volume_avg * 1.2]  # 20% above average volume
```

---

### Volume Ratio

Compare current volume to average.

**Configuration:**
```yaml
indicators:
  volume_avg:
    type: sma
    source: volume
    period: 20
  
  volume_ratio:
    type: divide
    left: volume
    right: volume_avg
```

**Use Cases:**
- Relative volume strength
- Breakout confirmation
- Liquidity filtering

**Example Strategy:**
```yaml
indicators:
  volume_avg:
    type: sma
    source: volume
    period: 20
  
  volume_ratio:
    type: divide
    left: volume
    right: volume_avg
  
  breakout_high:
    type: highest
    source: high
    period: 20

entry:
  long:
    all:
      - gt: [close, breakout_high]    # New high
      - gt: [volume_ratio, 1.5]       # 50% above average
```

---

## Composite Indicators

### Custom Calculations

Combine indicators with arithmetic operations.

**Available Operations:**
- `add` - Addition
- `subtract` - Subtraction
- `multiply` - Multiplication
- `divide` - Division

**Example: EMA Distance**
```yaml
indicators:
  ema_fast:
    type: ema
    period: 12
  
  ema_slow:
    type: ema
    period: 26
  
  # Distance between EMAs
  ema_distance:
    type: subtract
    left: ema_fast
    right: ema_slow
  
  # Percentage distance
  ema_distance_pct:
    type: divide
    left: ema_distance
    right: ema_slow
```

**Example: Volatility-Adjusted Signal**
```yaml
indicators:
  rsi:
    type: rsi
    period: 14
  
  atr:
    type: atr
    period: 14
  
  atr_avg:
    type: sma
    source: atr
    period: 20
  
  # Volatility ratio
  vol_ratio:
    type: divide
    left: atr
    right: atr_avg

entry:
  long:
    all:
      - lt: [rsi, 30]           # Oversold
      - gt: [vol_ratio, 1.2]    # Above-average volatility
```

---

## Custom Indicators

### Creating Custom Indicators

You can extend smallbt_go with custom indicators in Go.

**Example: Custom Indicator Structure**

```go
package indicator

type MyCustomIndicator struct {
    period int
    values []float64
}

func NewMyCustomIndicator(period int) *MyCustomIndicator {
    return &MyCustomIndicator{
        period: period,
        values: make([]float64, 0),
    }
}

func (m *MyCustomIndicator) Calculate(candle *market.Candle) float64 {
    // Your calculation logic here
    m.values = append(m.values, candle.Close)
    
    if len(m.values) < m.period {
        return 0 // Not enough data
    }
    
    // Example: custom calculation
    sum := 0.0
    for i := len(m.values) - m.period; i < len(m.values); i++ {
        sum += m.values[i]
    }
    
    return sum / float64(m.period)
}

func (m *MyCustomIndicator) IsValid() bool {
    return len(m.values) >= m.period
}
```

**Register Custom Indicator:**

```go
// internal/config/config.go
func init() {
    RegisterIndicator("my_custom", func(params map[string]interface{}) Indicator {
        period := params["period"].(int)
        return indicator.NewMyCustomIndicator(period)
    })
}
```

**Use in Strategy:**
```yaml
indicators:
  custom:
    type: my_custom
    period: 14
```

---

## Indicator Best Practices

### 1. Start with Standard Periods

Use commonly accepted periods first:
- SMA/EMA: 10, 20, 50, 100, 200
- RSI: 14
- ATR: 14
- MACD: 12/26/9
- Bollinger: 20

### 2. Avoid Over-Optimization

```yaml
# ❌ Too many indicators
indicators:
  sma_5, sma_10, sma_15, sma_20, sma_25...
  ema_5, ema_10, ema_15...
  rsi_7, rsi_14, rsi_21...

# ✅ Focused set
indicators:
  ema_fast:
    type: ema
    period: 12
  
  ema_slow:
    type: ema
    period: 26
  
  rsi:
    type: rsi
    period: 14
```

### 3. Consider Warm-up Period

```yaml
# This strategy needs 200 candles before valid signals
indicators:
  ema_200:
    type: ema
    period: 200
```

Ensure your data has sufficient history.

### 4. Use Appropriate Source

```yaml
# ✅ Good: Close for most indicators
indicators:
  sma:
    type: sma
    source: close
    period: 20

# ⚠️ Careful: Volume has different characteristics
indicators:
  volume_sma:
    type: sma
    source: volume  # Make sure this makes sense
    period: 20
```

### 5. Validate Indicator Logic

Test indicator calculations:

```bash
# Enable verbose to see indicator values
trader backtest --strategy s.yaml --data d.csv --verbose
```

### 6. Combine Complementary Indicators

```yaml
# ✅ Good combination: Trend + Momentum + Volume
indicators:
  ema_20:      # Trend
    type: ema
    period: 20
  
  rsi:         # Momentum
    type: rsi
    period: 14
  
  volume_avg:  # Volume
    type: sma
    source: volume
    period: 20

entry:
  long:
    all:
      - gt: [close, ema_20]              # Above trend
      - lt: [rsi, 30]                    # Oversold
      - gt: [volume, volume_avg * 1.2]   # Volume confirmation
```

---

## Common Indicator Combinations

### Combination 1: Dual Moving Average

```yaml
indicators:
  ma_fast:
    type: ema
    period: 12
  
  ma_slow:
    type: ema
    period: 26

entry:
  long:
    all:
      - cross_above: [ma_fast, ma_slow]
```

**Use Case:** Trend following

---

### Combination 2: Triple Moving Average

```yaml
indicators:
  ma_short:
    type: ema
    period: 9
  
  ma_medium:
    type: ema
    period: 21
  
  ma_long:
    type: ema
    period: 200

entry:
  long:
    all:
      - cross_above: [ma_short, ma_medium]
      - gt: [close, ma_long]  # Above long-term trend
```

**Use Case:** Trend with filter

---

### Combination 3: RSI + Bollinger

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
      - lt: [rsi, 30]
      - lt: [close, bollinger.lower]
```

**Use Case:** Mean reversion

---

### Combination 4: MACD + Volume

```yaml
indicators:
  macd:
    type: macd
  
  volume_avg:
    type: sma
    source: volume
    period: 20

entry:
  long:
    all:
      - cross_above: [macd.line, macd.signal]
      - gt: [volume, volume_avg * 1.5]
```

**Use Case:** Momentum with confirmation

---

## Troubleshooting

### "Insufficient data" Error

**Problem:** Not enough candles for indicator warm-up.

**Solution:**
- Use more historical data
- Reduce indicator periods
- Check indicator requirements

### Unexpected Indicator Values

**Problem:** Indicator producing strange values.

**Solution:**
- Verify source field is correct
- Check for data quality issues
- Enable verbose logging
- Compare with reference implementation

### Slow Backtest Performance

**Problem:** Too many indicators slowing execution.

**Solution:**
- Reduce number of indicators
- Use shorter periods where appropriate
- Optimize composite indicators
- Consider caching

---

## Further Reading

- [Getting Started Guide](getting-started.md)
- [Transform Guide](transforms.md)
- [CLI Reference](cli.md)
- [Strategy Examples](../../strategies/examples/)

---

**Master indicators for powerful strategies! 📊**
