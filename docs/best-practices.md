# Best Practices Guide

**Building robust and reliable trading strategies**

---

## Table of Contents

1. [Strategy Development](#strategy-development)
2. [Data Quality](#data-quality)
3. [Risk Management](#risk-management)
4. [Testing & Validation](#testing--validation)
5. [Performance Analysis](#performance-analysis)
6. [Common Pitfalls](#common-pitfalls)
7. [Production Readiness](#production-readiness)

---

## Strategy Development

### Start Simple, Add Complexity Gradually

```yaml
# ✅ Phase 1: Simple baseline
indicators:
  sma_fast:
    type: sma
    period: 10
  
  sma_slow:
    type: sma
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

Once baseline works:

```yaml
# ✅ Phase 2: Add volume confirmation
indicators:
  sma_fast:
    type: sma
    period: 10
  
  sma_slow:
    type: sma
    period: 30
  
  volume_avg:
    type: sma
    source: volume
    period: 20

entry:
  long:
    all:
      - cross_above: [sma_fast, sma_slow]
      - gt: [volume, volume_avg * 1.2]  # NEW
```

Then add more refinements incrementally.

---

### Use Meaningful Indicator Names

```yaml
# ❌ Bad: Unclear purpose
indicators:
  ind1:
    type: ema
    period: 12
  
  ind2:
    type: ema
    period: 26

# ✅ Good: Self-documenting
indicators:
  ema_fast:
    type: ema
    period: 12
    
  ema_slow:
    type: ema
    period: 26
  
  trend_filter:
    type: ema
    period: 200
```

---

### Keep Entry Conditions Focused

```yaml
# ❌ Too many conditions (overfitting risk)
entry:
  long:
    all:
      - cross_above: [ema_9, ema_21]
      - gt: [rsi, 50]
      - lt: [rsi, 70]
      - gt: [volume, volume_avg]
      - lt: [atr, atr_high]
      - gt: [close, bollinger.middle]
      - rising: [macd.histogram]

# ✅ Focused conditions
entry:
  long:
    all:
      - cross_above: [ema_9, ema_21]     # Trend signal
      - gt: [volume, volume_avg * 1.2]   # Volume confirmation
      - gt: [close, ema_200]             # Long-term trend filter
```

**Rule of Thumb:** 2-4 conditions is usually optimal.

---

### Document Your Strategy Logic

```yaml
strategy:
  name: ema_volume_breakout
  version: "1.2"
  description: |
    Trend-following strategy using EMA crossover with volume confirmation.
    
    Entry Logic:
    - Fast EMA crosses above slow EMA (trend change)
    - Volume 20% above average (conviction)
    - Price above 200 EMA (bull market filter)
    
    Exit Logic:
    - Fast EMA crosses below slow EMA
    - Or 2% stop loss hit
    - Or 4% take profit reached
    
    Risk: 1% per trade, 2:1 R/R
```

---

### Version Your Strategies

```yaml
strategy:
  name: my_strategy
  version: "2.1"  # Always version
  
  # Keep changelog in description
  description: |
    v2.1 (2024-12-01): Added volume filter
    v2.0 (2024-11-15): Changed to EMA from SMA
    v1.0 (2024-11-01): Initial version
```

---

## Data Quality

### Always Validate Data First

```bash
# Before backtesting, check data quality
head -20 data/BTCUSDT_1h.csv
tail -20 data/BTCUSDT_1h.csv

# Look for:
# - Proper timestamp format
# - No missing values
# - Reasonable OHLC relationships
# - No negative values
# - Consistent timeframe
```

---

### Check for Data Issues

**Common Problems:**

1. **Missing Candles**
```csv
# ❌ Gap in data
2024-01-01 00:00:00,42000,42500,41800,42300,150.5
2024-01-01 01:00:00,42300,42800,42200,42600,180.2
# Missing 02:00:00
2024-01-01 03:00:00,42600,43000,42500,42900,200.1
```

**Solution:** Fill gaps or split into separate backtests.

2. **Invalid OHLC Relationships**
```csv
# ❌ Close > High (impossible)
2024-01-01 00:00:00,42000,42500,41800,42600,150.5
                                    high ^ ^ close
```

**Solution:** Clean data before backtesting.

3. **Extreme Outliers**
```csv
# ❌ Flash crash / erroneous data
2024-01-01 00:00:00,42000,42500,41800,42300,150.5
2024-01-01 01:00:00,42300,42800,100,42600,180.2  # Low = 100?!
```

**Solution:** Use clipping transform or clean data.

---

### Use Sufficient Historical Data

**Minimum Recommendations:**

| Timeframe | Minimum Candles | Recommended |
|-----------|----------------|-------------|
| 1m | 10,000 | 50,000+ |
| 5m | 5,000 | 20,000+ |
| 15m | 3,000 | 10,000+ |
| 1h | 2,000 | 5,000+ |
| 4h | 1,000 | 2,500+ |
| 1d | 500 | 1,000+ |

**Why:** More data = better statistical significance.

---

### Account for Warm-up Period

```yaml
# This needs 200 candles before valid signals
indicators:
  ema_200:
    type: ema
    period: 200
```

**Ensure:** `total_candles > warm_up_period + minimum_trades * 50`

---

## Risk Management

### Always Use Stop Losses

```yaml
# ❌ No protection
risk:
  position_size:
    type: percent_equity
    value: 0.5

# ✅ Protected
risk:
  position_size:
    type: percent_equity
    value: 0.5
  
  stop_loss:
    type: percentage
    value: 0.02  # 2% stop
  
  take_profit:
    type: risk_reward
    ratio: 2
```

**Never trade without stops in backtesting** - it hides strategy weakness.

---

### Size Positions Appropriately

**Conservative Sizing:**

```yaml
risk:
  position_size:
    type: risk_percent
    value: 0.01  # Risk 1% per trade
  
  stop_loss:
    type: atr
    period: 14
    multiplier: 1.5
```

**Position Size = (Account Risk / Stop Distance)**

**Risk Per Trade Guidelines:**

| Risk Level | % Per Trade | Suitable For |
|------------|-------------|--------------|
| Conservative | 0.5-1% | Most traders |
| Moderate | 1-2% | Experienced |
| Aggressive | 2-5% | Very experienced |
| Dangerous | >5% | Not recommended |

---

### Use Risk-Reward Ratios

```yaml
# ✅ Positive expectancy structure
risk:
  stop_loss:
    type: percentage
    value: 0.02  # Risk 2%
  
  take_profit:
    type: risk_reward
    ratio: 2  # Target 4% (2x risk)
```

**Minimum R:R:** 1.5:1 (target 1.5x your risk)  
**Recommended:** 2:1 or better

---

### Limit Maximum Exposure

```yaml
risk:
  # Per-trade risk
  position_size:
    type: risk_percent
    value: 0.01
  
  # Portfolio-level risk (future feature)
  max_positions: 3
  max_exposure: 0.50  # Max 50% of capital deployed
```

---

## Testing & Validation

### The Testing Hierarchy

1. **Unit Test**: Does the indicator calculate correctly?
2. **Backtest**: Does the strategy work in-sample?
3. **Walk-Forward**: Does it work out-of-sample?
4. **Monte Carlo**: What's the confidence interval?
5. **Paper Trade**: Does it work in real-time?
6. **Live Trade**: Does it actually make money?

**Never skip steps!**

---

### Always Test Out-of-Sample

```bash
# ❌ Bad: Only in-sample
trader backtest --strategy s.yaml --data all_data.csv

# ✅ Good: Train/test split
trader backtest --strategy s.yaml --data train_2020_2022.csv
trader backtest --strategy s.yaml --data test_2023_2024.csv

# ✅ Better: Walk-forward
trader walkforward \
  --strategy s.yaml \
  --data all_data.csv \
  --train 2000 \
  --test 500 \
  --step 250
```

---

### Check for Overfitting

**Warning Signs:**

1. **Too many parameters optimized**
```bash
# ❌ Overfitting risk
trader optimize \
  --param ind1.period:5,50,1 \
  --param ind2.period:5,50,1 \
  --param ind3.period:5,50,1 \
  --param ind4.period:5,50,1
  # 46^4 = 4.5 million combinations!

# ✅ Focused optimization
trader optimize \
  --param sma_fast.period:5,30,5 \
  --param sma_slow.period:20,100,10
  # 6 * 9 = 54 combinations
```

2. **Perfect in-sample, terrible out-of-sample**
```
In-sample:  +150% return, Sharpe 3.5
Out-sample: -20% return, Sharpe 0.3
→ Overfitted!
```

3. **Win rate > 80%**
```
Win rate: 92%
→ Suspicious! Check for look-ahead bias or overfitting
```

---

### Use Walk-Forward Analysis

```bash
trader walkforward \
  --strategy my_strategy.yaml \
  --data BTCUSDT_2020_2024.csv \
  --train 2000 \
  --test 500 \
  --step 250
```

**Good Strategy:** Out-of-sample performance within 20-30% of in-sample.

**Example:**
```
Window 1: Train +45% → Test +32% (degradation: -29%)
Window 2: Train +48% → Test +35% (degradation: -27%)
Window 3: Train +52% → Test +38% (degradation: -27%)
Average degradation: -28% ✓ Acceptable
```

---

### Run Monte Carlo Analysis

```bash
trader montecarlo \
  --result backtest.json \
  --simulations 10000 \
  --confidence 95
```

**Check:**
- Probability of positive return > 70%
- 95% CI doesn't include catastrophic drawdown
- Mean expectancy > 0

---

### Test on Multiple Time Periods

```bash
# Bull market
trader backtest --strategy s.yaml --data 2020_2021_bull.csv

# Bear market
trader backtest --strategy s.yaml --data 2022_bear.csv

# Sideways market
trader backtest --strategy s.yaml --data 2019_sideways.csv
```

**Good Strategy:** Works in multiple market regimes.

---

## Performance Analysis

### Focus on Risk-Adjusted Returns

```yaml
# ❌ Focusing only on return
Return: +150%
# Could be one lucky trade!

# ✅ Comprehensive analysis
Return:        +150%
Sharpe:        2.1   # Good risk-adjusted return
Max Drawdown:  -18%  # Acceptable risk
Trades:        87    # Sufficient sample
Win Rate:      51%   # Realistic
Profit Factor: 2.3   # Solid edge
```

---

### Key Metrics to Monitor

**Must-Have Metrics:**

| Metric | Good | Acceptable | Warning |
|--------|------|------------|---------|
| Sharpe Ratio | >2.0 | >1.0 | <0.5 |
| Sortino Ratio | >2.5 | >1.5 | <0.75 |
| Max Drawdown | <15% | <25% | >40% |
| Win Rate | 45-65% | 35-75% | <35% or >80% |
| Profit Factor | >2.0 | >1.5 | <1.2 |
| Expectancy | >0.5R | >0.2R | <0.1R |
| Trades | >50 | >30 | <20 |

---

### Understand Trade Distribution

```bash
trader backtest --strategy s.yaml --data d.csv --format json > result.json

# Analyze trade distribution
python scripts/analyze_trades.py result.json
```

**Check:**
- Is profit from many trades or one big win?
- Are losses consistent or one catastrophic loss?
- Is there serial correlation?

---

### Compare to Buy-and-Hold

```bash
# Your strategy
trader backtest --strategy active.yaml --data d.csv

# Buy-and-hold baseline
trader backtest --strategy buy_hold.yaml --data d.csv
```

**Your strategy should:**
- Beat buy-and-hold returns, OR
- Have better Sharpe ratio, OR
- Have lower drawdown

Otherwise, why bother with active trading?

---

## Common Pitfalls

### Pitfall 1: Look-Ahead Bias

```yaml
# ❌ Using future information
indicators:
  sma_20:
    type: sma
    source: close
    period: 20

entry:
  long:
    all:
      - gt: [close[t], sma_20[t+1]]  # Using FUTURE SMA!
```

**Prevention:**
- Always use `[t]` or `[t-n]`, never `[t+n]`
- Test with walk-forward
- Manual spot-checks

---

### Pitfall 2: Survivorship Bias

```
# Testing only on BTCUSDT (winner)
# Ignoring 100 failed coins
```

**Prevention:**
- Test on multiple symbols
- Include delisted/failed assets if possible
- Be realistic about selection bias

---

### Pitfall 3: Ignoring Transaction Costs

```bash
# ❌ Unrealistic
trader backtest --strategy s.yaml --data d.csv
# Uses 0% fees!

# ✅ Realistic
trader backtest \
  --strategy s.yaml \
  --data d.csv \
  --commission 0.001 \
  --slippage 0.0005
```

**Use realistic fees:**
- Crypto: 0.1-0.2% per trade
- Stocks: $5-10 per trade or 0.01%
- Futures: Variable by contract

---

### Pitfall 4: Overfitting to Noise

```yaml
# ❌ Too specific, likely noise
entry:
  long:
    all:
      - eq: [rsi, 31.4]  # Exactly 31.4?!
      - eq: [sma_fast.period, 17]  # Why 17?
```

**Prevention:**
- Use round numbers
- Test parameter robustness
- Avoid excessive precision

---

### Pitfall 5: Data Snooping

```
Researcher tests 100 strategies on same data.
Publishes the 1 that worked (by chance).
```

**Prevention:**
- Test on fresh data
- Use walk-forward analysis
- Adjust for multiple testing

---

### Pitfall 6: Ignoring Market Regimes

```yaml
# Strategy works great 2020-2021 (bull)
# Fails catastrophically 2022 (bear)
```

**Prevention:**
- Test multiple time periods
- Add regime detection
- Use volatility filters

---

### Pitfall 7: Unrealistic Order Execution

```yaml
# ❌ Assumes instant fills at exact price
# Reality: slippage, partial fills, rejected orders
```

**Prevention:**
- Use slippage models
- Account for liquidity constraints
- Conservative execution assumptions

---

## Production Readiness

### Checklist Before Live Trading

- [ ] Strategy validated out-of-sample
- [ ] Walk-forward analysis completed
- [ ] Monte Carlo risk assessment done
- [ ] Transaction costs included
- [ ] Slippage modeled
- [ ] Stop losses implemented
- [ ] Position sizing appropriate (≤2% risk)
- [ ] Maximum drawdown acceptable
- [ ] Trade frequency manageable
- [ ] Sufficient capital for strategy
- [ ] Paper traded successfully for 1+ month
- [ ] Execution plan documented
- [ ] Risk limits defined
- [ ] Monitoring system in place

---

### Paper Trading

Before live capital:

1. **Run strategy in real-time** (simulated orders)
2. **Monitor for 30+ days**
3. **Check:**
   - Do signals match backtest?
   - Is execution realistic?
   - Any unexpected behaviors?
   - Performance close to backtest?

---

### Risk Limits

Define hard limits:

```yaml
# Example risk limits
max_daily_loss: -2%      # Stop trading if daily loss > 2%
max_weekly_loss: -5%     # Stop if weekly loss > 5%
max_monthly_loss: -10%   # Stop if monthly loss > 10%
max_drawdown: -20%       # Stop if drawdown > 20%

max_position_size: 50%   # Never exceed 50% capital in one trade
max_positions: 3         # Max 3 concurrent positions
```

**Stick to limits!** Discipline beats discretion.

---

### Monitoring

Monitor these metrics:

- **Performance vs backtest**: Tracking error
- **Sharpe ratio**: Rolling 30-day
- **Drawdown**: Current and max
- **Win rate**: Recent performance
- **Slippage**: Actual vs expected
- **Latency**: Order execution time

---

### When to Stop a Strategy

**Red Flags:**

1. Drawdown exceeds backtest max by 50%
2. Sharpe ratio < 0 for 3+ months
3. Execution slippage much worse than backtest
4. Market regime permanently changed
5. Hit predefined loss limits

**Don't:**
- Keep trading hoping to "get back to even"
- Increase position size to recover losses
- Abandon risk management

---

## Strategy Development Workflow

### Recommended Process

```
1. Idea Generation
   ↓
2. Simple Implementation
   ↓
3. Backtest Baseline
   ↓
4. Add Risk Management
   ↓
5. Test Out-of-Sample
   ↓
6. Walk-Forward Analysis
   ↓
7. Parameter Robustness Check
   ↓
8. Monte Carlo Analysis
   ↓
9. Multi-Regime Testing
   ↓
10. Paper Trade
    ↓
11. Live Trade (small size)
    ↓
12. Scale Up Gradually
```

---

### Documentation Template

```yaml
strategy:
  name: my_strategy
  version: "1.0"
  description: |
    ## Overview
    Brief description of strategy logic.
    
    ## Hypothesis
    Why this should work theoretically.
    
    ## Entry Logic
    - Condition 1
    - Condition 2
    
    ## Exit Logic
    - Exit condition
    
    ## Risk Management
    - Position sizing
    - Stop loss
    - Take profit
    
    ## Backtest Results
    - Period: 2020-2024
    - Return: +45%
    - Sharpe: 1.8
    - Max DD: -15%
    - Trades: 87
    
    ## Walk-Forward
    - 5 windows
    - Avg degradation: -25%
    - Out-sample Sharpe: 1.5
    
    ## Known Limitations
    - Performs poorly in low volatility
    - Requires liquid markets
    
    ## Changelog
    v1.0 (2024-12-01): Initial version
```

---

## Further Reading

- [Getting Started Guide](getting-started.md)
- [Transform Guide](transforms.md)
- [CLI Reference](cli.md)
- [Indicator Reference](indicators.md)
- [AGENTS.md](../AGENTS.md) - Project philosophy

---

**Build robust strategies with discipline! 🎯**
