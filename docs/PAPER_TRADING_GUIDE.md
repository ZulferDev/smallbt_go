# Paper Trading Guide

**Version:** 1.0  
**Date:** 2026-09-05  
**Phase:** 16 Week 4  

---

## Overview

Paper trading allows you to test strategies with simulated real-time execution without risking real capital. The smallbt_go paper trading system supports:

- Realistic order execution with latency (50-200ms)
- Portfolio tracking (cash, equity, positions)
- WebSocket real-time data feeds
- CSV historical data simulation
- Same strategy YAML as backtesting

---

## Quick Start

### Basic Paper Trading (Static Price)

```bash
trader paper --strategy strategies/examples/paper_ema_cross.yaml \
             --symbol BTCUSDT \
             --price 50000 \
             --balance 10000 \
             --duration 60
```

**Output:**
```
Starting paper trading...
Strategy: paper_ema_cross
Symbol: BTCUSDT
Initial Price: 50000.00
Initial Balance: 10000.00
Duration: 60 seconds

Press Ctrl+C to stop

[5s] Balance: 10000.00 | Equity: 10000.00 | Positions: 0
[10s] Balance: 10000.00 | Equity: 10000.00 | Positions: 0
...
```

### Paper Trading with WebSocket

```bash
trader paper --strategy strategies/examples/paper_ema_cross.yaml \
             --symbol BTCUSDT \
             --websocket ws://localhost:8080 \
             --balance 10000 \
             --duration 300
```

**Output:**
```
Starting paper trading...
Strategy: paper_ema_cross
Symbol: BTCUSDT
Initial Balance: 10000.00
Duration: 300 seconds
WebSocket: ws://localhost:8080

Connected to WebSocket: ws://localhost:8080
Subscribing to: BTCUSDT

[Candle 1] 15:26:25 | O:50000.00 H:50100.00 L:49900.00 C:50050.00 V:1500.00
[Candle 2] 15:26:30 | O:50050.00 H:50150.00 L:50000.00 C:50100.00 V:1200.00

[5s] Balance: 10000.00 | Equity: 10000.00 | Positions: 0 | Candles: 2

[Candle 3] 15:26:35 | O:50100.00 H:50200.00 L:50050.00 C:50150.00 V:1800.00
...
```

---

## CLI Flags

### Required Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--strategy` | Path to strategy YAML file | `--strategy ema.yaml` |

### Optional Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--symbol` | Trading symbol | BTCUSDT | `--symbol ETHUSDT` |
| `--price` | Initial price (static mode) | 50000.0 | `--price 45000` |
| `--balance` | Initial balance | 10000.0 | `--balance 50000` |
| `--duration` | Duration in seconds | 60 | `--duration 300` |
| `--websocket` | WebSocket URL (optional) | - | `--websocket ws://localhost:8080` |

---

## Data Sources

### 1. Static Price (Default)

Uses a fixed price for the entire session. Useful for testing order logic without price movement.

```bash
trader paper --strategy strategy.yaml --price 50000 --duration 60
```

**Use cases:**
- Testing order submission
- Testing position tracking
- Testing portfolio accounting
- Quick smoke tests

### 2. WebSocket Real-Time Data

Connects to a WebSocket server for live price updates.

```bash
trader paper --strategy strategy.yaml --websocket ws://localhost:8080 --duration 300
```

**Requirements:**
- WebSocket server must send JSON candle data
- Format: `{"timestamp": 1234567890, "open": 50000, "high": 50100, "low": 49900, "close": 50050, "volume": 1000}`

**Use cases:**
- Testing with realistic price movement
- Integration with exchange simulators
- Pre-live trading validation

---

## Strategy Configuration

Paper trading uses the same strategy YAML format as backtesting.

### Example Strategy

```yaml
strategy:
  name: paper_ema_cross
  version: "1.0"

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
    type: risk_percent
    value: 0.02

  stop_loss:
    type: atr
    period: 14
    multiplier: 2.0

  take_profit:
    type: risk_reward
    ratio: 3.0
```

**See:** `strategies/examples/paper_ema_cross.yaml`

---

## Order Execution

### Latency Simulation

Paper trading simulates realistic order execution latency:

- **Order submission:** 50-200ms (random)
- **Order processing:** Background goroutine (100ms ticker)
- **Order fill:** Market orders fill immediately after latency

**Configuration:**
```go
// Default latency config (internal)
DefaultLatencyConfig() LatencyConfig {
    Min: 50 * time.Millisecond,
    Max: 200 * time.Millisecond,
}
```

### Order Lifecycle

```
1. Strategy generates signal
   ↓
2. SubmitOrder(order) → Order ID returned
   ↓
3. Latency delay (50-200ms)
   ↓
4. Order queued (OrderStatusPending)
   ↓
5. Background processor picks up order (100ms ticker)
   ↓
6. Market order fills at current price
   ↓
7. Portfolio updated (cash, equity, positions)
   ↓
8. Order status → OrderStatusFilled
```

---

## Portfolio Tracking

### Balance Components

- **Cash:** Available funds for new positions
- **Equity:** Total portfolio value (cash + position value)
- **Margin:** Used for open positions (if applicable)

### Position Tracking

Each open position includes:
- Symbol
- Quantity
- Entry price
- Current price
- Unrealized PnL
- Entry time

### Real-Time Updates

Status printed every 5 seconds:

```
[25s] Balance: 9500.00 | Equity: 10050.00 | Positions: 1 | Candles: 5
  BTCUSDT: 0.1000 @ 50000.00 (PnL: 50.00)
```

**Components:**
- **Balance:** Cash remaining
- **Equity:** Total value
- **Positions:** Number of open positions
- **Candles:** Number received (WebSocket mode)

---

## Session Summary

At the end of the session, a summary is displayed:

```
==================================================
PAPER TRADING SUMMARY
==================================================

Final Balance: 9850.00
Final Equity:  10100.00
Positions:     1

Open Positions:
  BTCUSDT: 0.1000 @ 50000.00 (PnL: +250.00)

==================================================
```

---

## WebSocket Protocol

### Expected Message Format

Paper trading expects JSON messages with OHLCV candle data:

```json
{
  "timestamp": 1609459200,
  "open": 50000.0,
  "high": 50100.0,
  "low": 49900.0,
  "close": 50050.0,
  "volume": 1000.0
}
```

### Field Requirements

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `timestamp` | int64 | Unix timestamp (seconds) | ✅ |
| `open` | float64 | Open price | ✅ |
| `high` | float64 | High price | ✅ |
| `low` | float64 | Low price | ✅ |
| `close` | float64 | Close price | ✅ |
| `volume` | float64 | Volume | ✅ |

### Validation

Candles must pass validation:
- `high >= low`
- `high >= open`
- `high >= close`
- `low <= open`
- `low <= close`
- All prices >= 0
- Volume >= 0

Invalid candles are logged and skipped (connection remains active).

---

## Common Workflows

### 1. Test Strategy Before Live Trading

```bash
# Step 1: Backtest with historical data
trader backtest --strategy strategy.yaml --data BTCUSDT.csv

# Step 2: Paper trade with WebSocket
trader paper --strategy strategy.yaml --websocket ws://exchange-sim:8080 --duration 3600

# Step 3: Review results, iterate strategy

# Step 4: Deploy to live trading (Phase 17+)
```

### 2. Validate Order Logic

```bash
# Use static price to test order submission
trader paper --strategy strategy.yaml --price 50000 --duration 60
```

### 3. Test with Simulated Price Movement

```bash
# Connect to local WebSocket simulator
trader paper --strategy strategy.yaml --websocket ws://localhost:8080 --duration 300
```

---

## Troubleshooting

### Problem: WebSocket connection fails

**Symptoms:**
```
Error: connect to WebSocket: dial tcp: connection refused
```

**Solutions:**
- Verify WebSocket server is running: `curl --include --no-buffer --header "Connection: Upgrade" --header "Upgrade: websocket" http://localhost:8080`
- Check URL format: Must be `ws://` or `wss://`
- Check firewall/network: Ensure port is accessible

### Problem: No candles received

**Symptoms:**
```
Connected to WebSocket: ws://localhost:8080
Subscribing to: BTCUSDT

[5s] Balance: 10000.00 | Equity: 10000.00 | Positions: 0 | Candles: 0
```

**Solutions:**
- Verify server is sending data: Use WebSocket client to inspect messages
- Check message format: Must match expected JSON schema
- Check symbol: Server may not support the requested symbol

### Problem: Orders not filling

**Symptoms:**
```
Balance: 10000.00 | Equity: 10000.00 | Positions: 0
(No positions created despite signals)
```

**Solutions:**
- Check strategy logic: Verify entry conditions are being met
- Wait for latency: Orders take 50-200ms + processing time
- Check balance: Ensure sufficient funds for position size
- Check logs: Look for order rejection messages

### Problem: Unexpected PnL

**Symptoms:**
```
Position PnL doesn't match expected value
```

**Solutions:**
- Check fees: Paper trading simulates maker/taker fees
- Check price updates: Verify WebSocket is sending correct prices
- Check position size: Verify quantity calculation
- Allow for latency: Order fill price may differ from signal price

---

## Best Practices

### 1. Start with Short Sessions

```bash
# Start with 60-300 seconds
trader paper --strategy strategy.yaml --duration 60
```

### 2. Use Realistic Initial Balance

```bash
# Match your intended live trading capital
trader paper --strategy strategy.yaml --balance 10000
```

### 3. Test Both Data Sources

```bash
# Static for order logic
trader paper --strategy strategy.yaml --price 50000

# WebSocket for realistic price movement
trader paper --strategy strategy.yaml --websocket ws://localhost:8080
```

### 4. Monitor for Extended Periods

```bash
# Run for hours to test stability
trader paper --strategy strategy.yaml --websocket ws://localhost:8080 --duration 3600
```

### 5. Same Strategy for Backtest and Paper

```yaml
# Use identical YAML for both modes
# This ensures consistency
```

---

## Limitations

### Current Limitations (Phase 16)

1. **No Strategy Evaluation**
   - Price updates work
   - Manual signal generation only
   - Automatic strategy evaluation: Phase 17+

2. **No Order Types Beyond Market**
   - Market orders only in CLI
   - Limit/stop orders: Broker supports, CLI needs extension

3. **No Historical Data Replay**
   - Static price or WebSocket only
   - CSV replay: Phase 17+

4. **No Multi-Symbol**
   - Single symbol per session
   - Multi-symbol: Phase 17+

5. **No Advanced Risk Management**
   - Basic position sizing
   - Advanced rules: Phase 17+

### Future Enhancements (Phase 17+)

- Strategy evaluation with real-time data
- CSV historical data replay mode
- Multi-symbol paper trading
- Advanced order types (limit, stop, stop-limit)
- Risk management integration
- Performance analytics
- Trade journal export

---

## Examples

### Example 1: Basic EMA Cross

```bash
trader paper --strategy strategies/examples/paper_ema_cross.yaml \
             --symbol BTCUSDT \
             --price 50000 \
             --duration 120
```

### Example 2: WebSocket Real-Time

```bash
trader paper --strategy strategies/examples/paper_ema_cross.yaml \
             --symbol BTCUSDT \
             --websocket ws://localhost:8080 \
             --balance 50000 \
             --duration 600
```

### Example 3: Extended Session

```bash
trader paper --strategy strategies/examples/paper_ema_cross.yaml \
             --symbol ETHUSDT \
             --websocket ws://exchange-sim:8080 \
             --balance 20000 \
             --duration 3600
```

---

## Next Steps

1. **Create Your Strategy:** Start with `strategies/examples/paper_ema_cross.yaml`
2. **Run Paper Trading:** Test with static price first
3. **Set Up WebSocket:** Use local simulator or exchange testnet
4. **Iterate:** Refine strategy based on paper trading results
5. **Deploy Live:** Phase 17+ (with proper risk management)

---

## Additional Resources

- **Strategy YAML Reference:** See `AGENTS.md` Section 11-20
- **Indicator Documentation:** See `ARCHITECTURE.md`
- **Backtest Guide:** Run `trader backtest --help`
- **WebSocket Feed Architecture:** See `ARCHITECTURE.md` Real-Time Data Feed section

---

**Last Updated:** 2026-09-05  
**Version:** 1.0 (Phase 16 Week 4)
