# Indicators and Conditions

## Indicators

### Built-in Indicators

All indicators implement the `Indicator` interface:

```go
type Indicator interface {
    Name() string
    Calculate(ctx *EvaluationContext) Value
}
```

### SMA - Simple Moving Average

**Type**: `sma`

**Parameters**:
- `source`: `open`, `high`, `low`, `close`, `volume`
- `period`: positive integer

**Formula**:
```
SMA = (sum of N periods) / N
```

**Example**:
```yaml
indicators:
  sma20:
    type: sma
    source: close
    period: 20
```

### EMA - Exponential Moving Average

**Type**: `ema`

**Parameters**:
- `source`: `open`, `high`, `low`, `close`, `volume`
- `period`: positive integer

**Formula**:
```
Multiplier = 2 / (period + 1)
EMA = (Close - EMA_prev) * Multiplier + EMA_prev
```

**Example**:
```yaml
indicators:
  ema9:
    type: ema
    source: close
    period: 9
```

### RSI - Relative Strength Index

**Type**: `rsi`

**Parameters**:
- `source`: `close`
- `period`: positive integer

**Formula**:
```
Average Gain = sum of gains over period / period
Average Loss = sum of losses over period / period
RS = Average Gain / Average Loss
RSI = 100 - (100 / (1 + RS))
```

**Example**:
```yaml
indicators:
  rsi14:
    type: rsi
    source: close
    period: 14
```

### ATR - Average True Range

**Type**: `atr`

**Parameters**:
- `period`: positive integer

**Formula**:
```
True Range = max(high - low, |high - prev_close|, |low - prev_close|)
ATR = (sum of TR over period) / period
```

**Example**:
```yaml
indicators:
  atr14:
    type: atr
    period: 14
```

## Custom Indicators

Register custom indicators via Go:

```go
type MyIndicator struct {
    period int
}

func (m *MyIndicator) Name() string { return "my_indicator" }

func (m *MyIndicator) Calculate(ctx *indicator.EvaluationContext) indicator.Value {
    // Implementation
    return indicator.FloatValue(0)
}

func init() {
    registry.Register("my_indicator", func(config map[string]interface{}) (indicator.Indicator, error) {
        period := int(config["period"].(float64))
        return &MyIndicator{period: period}, nil
    })
}
```

## Conditions

### Comparison Operators

All comparison operators work with indicator values and constants:

| Operator | Description | Example |
|----------|-------------|---------|
| `gt` | Greater than | `gt: [close, sma20]` |
| `lt` | Less than | `lt: [rsi, 70]` |
| `gte` | Greater than or equal | `gte: [volume, 1000]` |
| `lte` | Less than or equal | `lte: [atr, 5]` |
| `eq` | Equal | `eq: [state, true]` |

### Cross Detection

#### cross_above

True if left value crosses above right value:

```yaml
- cross_above: [ema9, ema21]
```

**Conditions**:
- `ema9[t-1] <= ema21[t-1]` AND `ema9[t] > ema21[t]`

#### cross_below

True if left value crosses below right value:

```yaml
- cross_below: [ema9, ema21]
```

**Conditions**:
- `ema9[t-1] >= ema21[t-1]` AND `ema9[t] < ema21[t]`

### Trend Detection

#### rising

True if value is higher than previous bar:

```yaml
- rising: [rsi]
```

**Conditions**:
- `rsi[t] > rsi[t-1]`

#### falling

True if value is lower than previous bar:

```yaml
- falling: [atr]
```

**Conditions**:
- `atr[t] < atr[t-1]`

### Range Detection

#### between

True if value is within range:

```yaml
- between: [rsi, 40, 60]
```

**Conditions**:
- `rsi[t] >= 40` AND `rsi[t] <= 60`

### Logical Operators

#### all (AND)

All conditions must be true:

```yaml
entry:
  long:
    all:
      - cross_above: [ema9, ema21]
      - gt: [volume, volume_avg]
```

#### any (OR)

At least one condition must be true:

```yaml
entry:
  long:
    any:
      - gt: [rsi, 70]
      - lt: [rsi, 30]
```

#### not (NOT)

Condition must be false:

```yaml
entry:
  long:
    not:
      - gt: [atr, max_volatility]
```

## Composite Indicators

### Arithmetic Operators

Supported operators: `add`, `subtract`, `multiply`, `divide`

```yaml
indicators:
  ema20:
    type: ema
    period: 20
  
  ema50:
    type: ema
    period: 50
  
  ema_distance:
    type: subtract
    left: ema20
    right: ema50
  
  ema_distance_pct:
    type: divide
    left: ema_distance
    right: ema50
```

## Historical References

### previous / shift

Access historical values:

```yaml
indicators:
  ema21_prev:
    type: previous
    value: ema21
```

Or with explicit bars:

```yaml
indicators:
  ema21_3bars_ago:
    type: shift
    value: ema21
    bars: 3
```

## Condition Examples

### Complex Entry Rule

```yaml
entry:
  long:
    all:
      - gt: [close, ema200]  # Price above 200 EMA
      - or:
          - all:
              - cross_above: [ema9, ema21]
              - gt: [volume, volume_avg]
          - all:
              - between: [rsi, 40, 50]
              - rising: [rsi]
```

### Stateful Exit

```yaml
state:
  setup_valid:
    default: false

rules:
  - when:
      cross_above: [ema9, ema21]
    set:
      setup_valid: true
  
  - when:
      all:
        - eq: [setup_valid, true]
        - cross_below: [ema9, ema21]
    action:
      exit: long
```