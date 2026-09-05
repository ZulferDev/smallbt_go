# Strategy DSL Reference

## Overview

The Strategy DSL allows you to define trading strategies declaratively in YAML without writing Go code.

## Strategy Structure

```yaml
strategy:
  name: string          # Required
  version: string       # Required

data:
  symbol: string        # Required
  timeframe: string     # Required (1m, 5m, 15m, 1h, 4h, 1d)

indicators:
  # Indicator definitions

entry:
  long: # Entry conditions for long positions
  short: # Entry conditions for short positions (optional)

exit:
  long: # Exit conditions for long positions
  short: # Exit conditions for short positions (optional)

risk:
  position_size: # Position sizing rules
  stop_loss: # Stop loss configuration (optional)
  take_profit: # Take profit configuration (optional)

execution: # Optional execution parameters
  fees:
    maker: float
    taker: float
  slippage:
    type: percentage
    value: float
```

## Indicators

### Built-in Indicators

#### SMA (Simple Moving Average)
```yaml
indicators:
  sma20:
    type: sma
    source: close  # open, high, low, close, volume
    period: 20
```

#### EMA (Exponential Moving Average)
```yaml
indicators:
  ema9:
    type: ema
    source: close
    period: 9
```

#### RSI (Relative Strength Index)
```yaml
indicators:
  rsi14:
    type: rsi
    source: close
    period: 14
```

#### ATR (Average True Range)
```yaml
indicators:
  atr14:
    type: atr
    period: 14
```

### Composite Indicators

You can compose indicators using arithmetic operators:

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

Supported operators: `add`, `subtract`, `multiply`, `divide`

## Conditions

### Comparison Operators

```yaml
entry:
  long:
    all:
      - gt: [close, sma20]      # Greater than
      - lt: [rsi, 70]           # Less than
      - gte: [volume, 1000]     # Greater than or equal
      - lte: [atr, 5]           # Less than or equal
      - eq: [state_var, true]   # Equal
```

### Logical Operators

```yaml
entry:
  long:
    all:  # AND - all conditions must be true
      - cross_above: [ema9, ema21]
      - gt: [volume, volume_avg]
    
    any:  # OR - at least one condition must be true
      - gt: [rsi, 70]
      - lt: [rsi, 30]
    
    not:  # NOT - condition must be false
      - gt: [atr, max_volatility]
```

### Trading Conditions

#### Cross Above / Cross Below
```yaml
entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]

exit:
  long:
    any:
      - cross_below: [ema_fast, ema_slow]
```

#### Rising / Falling
```yaml
entry:
  long:
    all:
      - rising: [rsi]   # Value increased from previous bar
      - falling: [atr]  # Value decreased from previous bar
```

#### Between
```yaml
entry:
  long:
    all:
      - between: [rsi, 40, 60]  # RSI between 40 and 60
```

## Entry and Exit Rules

### Long Entry
```yaml
entry:
  long:
    all:
      - cross_above: [ema9, ema21]
      - gt: [volume, volume_avg]
      - between: [rsi, 40, 60]
```

### Short Entry
```yaml
entry:
  short:
    all:
      - cross_below: [ema9, ema21]
      - gt: [volume, volume_avg]
```

### Exit Rules
```yaml
exit:
  long:
    any:  # Exit if ANY condition is true
      - cross_below: [ema9, ema21]
      - gt: [rsi, 80]
  
  short:
    any:
      - cross_above: [ema9, ema21]
      - lt: [rsi, 20]
```

## Risk Management

### Position Sizing

#### Fixed Size
```yaml
risk:
  position_size:
    type: fixed
    value: 100  # Fixed quantity
```

#### Percent of Equity
```yaml
risk:
  position_size:
    type: percent_equity
    value: 0.1  # 10% of equity
```

#### Risk Percent (with stop loss)
```yaml
risk:
  position_size:
    type: risk_percent
    value: 0.01  # Risk 1% of equity
```

### Stop Loss

#### Fixed Percentage
```yaml
risk:
  stop_loss:
    type: fixed
    value: 0.95  # 5% stop loss (price * 0.95)
```

#### ATR-based
```yaml
risk:
  stop_loss:
    type: atr
    period: 14
    multiplier: 1.5  # 1.5 * ATR
```

### Take Profit

#### Fixed Percentage
```yaml
risk:
  take_profit:
    type: fixed
    value: 1.05  # 5% take profit
```

#### Risk/Reward Ratio
```yaml
risk:
  take_profit:
    type: risk_reward
    ratio: 2  # 2x the risk (stop loss distance)
```

### Trailing Stop
```yaml
risk:
  trailing_stop:
    type: percentage
    value: 0.02  # Trail 2% below highest price
```

## Stateful Strategies

Define state variables and rules:

```yaml
state:
  breakout_setup:
    default: false
  
  bars_since_setup:
    default: 0

rules:
  - when:
      gt: [close, resistance]
    set:
      breakout_setup: true
      bars_since_setup: 0
  
  - when:
      eq: [breakout_setup, true]
    set:
      bars_since_setup:
        type: increment  # bars_since_setup += 1
  
  - when:
      all:
        - eq: [breakout_setup, true]
        - gt: [volume, volume_threshold]
        - lt: [bars_since_setup, 10]
    action:
      enter: long
  
  - when:
      gte: [bars_since_setup, 10]
    set:
      breakout_setup: false
```

## Multi-Timeframe

```yaml
data:
  primary:
    timeframe: 1h
  
  higher:
    timeframe: 4h

indicators:
  ema200_4h:
    type: ema
    timeframe: 4h
    period: 200
  
  rsi_1h:
    type: rsi
    timeframe: 1h
    period: 14

entry:
  long:
    all:
      - gt: [close, ema200_4h]  # Price above 4H EMA200
      - lt: [rsi_1h, 70]        # 1H RSI below 70
```

## Execution Parameters

```yaml
execution:
  fees:
    maker: 0.0002  # 0.02%
    taker: 0.0005  # 0.05%
  
  slippage:
    type: percentage
    value: 0.0005  # 0.05% slippage
```

## Complete Example

```yaml
strategy:
  name: ema_volume_trend
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema_fast:
    type: ema
    source: close
    period: 9
  
  ema_slow:
    type: ema
    source: close
    period: 21
  
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
      - gt: [volume_ratio, 1.2]

exit:
  long:
    any:
      - cross_below: [ema_fast, ema_slow]

risk:
  position_size:
    type: risk_percent
    value: 0.01
  
  stop_loss:
    type: atr
    period: 14
    multiplier: 1.5
  
  take_profit:
    type: risk_reward
    ratio: 2

execution:
  fees:
    maker: 0.0002
    taker: 0.0005
  slippage:
    type: percentage
    value: 0.0005
```

## Validation

Validate your strategy before backtesting:

```bash
./trader validate --strategy your_strategy.yaml
```

Common validation errors:
- Unknown indicator type
- Circular indicator dependencies
- Invalid parameter values
- Missing required fields
- Invalid condition syntax