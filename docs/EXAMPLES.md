# SmallBT Examples

Complete step-by-step examples for using SmallBT from beginner to advanced scenarios.

## Table of Contents

1. [Beginner Examples](#beginner-examples)
2. [Intermediate Examples](#intermediate-examples)
3. [Advanced Examples](#advanced-examples)
4. [Real-World Scenarios](#real-world-scenarios)

---

## Beginner Examples

### Example 1: Simple Moving Average Crossover

**Goal:** Create and backtest a basic SMA crossover strategy.

#### Step 1: Create Strategy File

Create `sma_cross.yaml`:

```yaml
strategy:
  name: sma_crossover_basic
  version: "1"
  description: "Simple SMA 50/200 crossover - Golden Cross strategy"

data:
  symbol: BTCUSDT
  timeframe: 1d

indicators:
  sma50:
    type: sma
    source: close
    period: 50

  sma200:
    type: sma
    source: close
    period: 200

entry:
  long:
    all:
      - cross_above: [sma50, sma200]  # Golden Cross

exit:
  long:
    any:
      - cross_below: [sma50, sma200]  # Death Cross

risk:
  position_size:
    type: percent_equity
    value: 0.95  # Use 95% of equity (full allocation)
  
  max_positions: 1
```

#### Step 2: Validate Strategy

```bash
./trader validate sma_cross.yaml
```

**Expected Output:**
```
✅ Strategy 'sma_crossover_basic' (v1) validated successfully
   Symbol: BTCUSDT, Timeframe: 1d
   Indicators: 2
   Entry rules: 1
   - sma50: sma (period: 50)
   - sma200: sma (period: 200)
```

#### Step 3: Run Backtest

```bash
./trader backtest \
  --strategy sma_cross.yaml \
  --data data/BTCUSDT.csv \
  --cash 10000 \
  --output results/sma_cross.json \
  --csv results/sma_cross_trades.csv
```

#### Step 4: Analyze Results

```bash
./trader analyze-trades --result results/sma_cross.json
```

**What to look for:**
- Total Return: Should be positive for trending markets
- Max Drawdown: Typically high (20-30%) for long-term trend following
- Win Rate: Usually low (30-40%) but large winners
- Number of Trades: Should be low (5-15 per year for daily timeframe)

---

### Example 2: RSI Oversold/Overbought

**Goal:** Trade RSI extremes with simple rules.

#### Create Strategy

`rsi_simple.yaml`:

```yaml
strategy:
  name: rsi_simple
  version: "1"
  description: "Buy oversold, sell overbought"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  rsi:
    type: rsi
    source: close
    period: 14

entry:
  long:
    all:
      - lt: [rsi, 30]  # RSI below 30 (oversold)

exit:
  long:
    any:
      - gt: [rsi, 70]  # RSI above 70 (overbought)

risk:
  position_size:
    type: percent_equity
    value: 0.2  # 20% per trade
  
  max_positions: 3  # Allow up to 3 concurrent positions
```

#### Run Complete Analysis

```bash
# 1. Validate
./trader validate rsi_simple.yaml

# 2. Backtest
./trader backtest \
  --strategy rsi_simple.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-01-01 \
  --end 2023-12-31 \
  --output results/rsi_simple.json

# 3. Check trade details
./trader export-trades \
  --result results/rsi_simple.json \
  --output results/rsi_trades.csv

# 4. Analyze patterns
./trader analyze-trades \
  --result results/rsi_simple.json \
  --group-by month
```

**Key Learning Points:**
- Mean reversion strategies work best in ranging markets
- Multiple positions help catch various opportunities
- RSI alone may generate many false signals in trending markets

---

### Example 3: Adding Stop Loss and Take Profit

**Goal:** Improve the RSI strategy with risk management.

#### Enhanced Strategy

`rsi_with_risk.yaml`:

```yaml
strategy:
  name: rsi_with_risk
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  rsi:
    type: rsi
    source: close
    period: 14
    
  atr:
    type: atr
    period: 14

entry:
  long:
    all:
      - lt: [rsi, 30]

exit:
  long:
    any:
      - gt: [rsi, 70]

risk:
  position_size:
    type: risk_percent
    value: 0.01  # Risk 1% per trade (much safer!)
    
  stop_loss:
    type: atr
    indicator: atr
    multiplier: 2.0  # Stop at 2x ATR
    
  take_profit:
    type: risk_reward
    ratio: 3  # Take profit at 3:1 risk/reward
  
  max_positions: 3
  max_daily_loss: 0.03  # Stop trading if lose 3% in one day
```

#### Compare With and Without Risk Management

```bash
# Run both
./trader backtest --strategy rsi_simple.yaml --data data/BTCUSDT.csv --output results/rsi_simple.json
./trader backtest --strategy rsi_with_risk.yaml --data data/BTCUSDT.csv --output results/rsi_risk.json

# Compare
echo "=== Simple RSI ==="
./trader report --result results/rsi_simple.json | grep -E "Return|Drawdown|Trades"

echo "=== RSI with Risk Management ==="
./trader report --result results/rsi_risk.json | grep -E "Return|Drawdown|Trades"
```

**Expected Differences:**
- With risk management: Lower return but much lower drawdown
- More trades (stops and targets exit earlier)
- More consistent performance
- Better Sharpe ratio

---

## Intermediate Examples

### Example 4: EMA + Volume Confirmation

**Goal:** Build a trend-following strategy with volume filter.

#### Strategy File

`ema_volume.yaml`:

```yaml
strategy:
  name: ema_volume_trend
  version: "1"
  description: "EMA crossover with volume confirmation"

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
    
  volume_avg:
    type: sma
    source: volume
    period: 20
    
  # Calculate volume ratio
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
      - cross_above: [ema_fast, ema_slow]  # Bullish crossover
      - gt: [volume_ratio, 1.2]  # Volume 20% above average

exit:
  long:
    any:
      - cross_below: [ema_fast, ema_slow]  # Bearish crossover

risk:
  position_size:
    type: risk_percent
    value: 0.01
    
  stop_loss:
    type: atr
    indicator: atr
    multiplier: 1.5
    
  take_profit:
    type: risk_reward
    ratio: 2
    
  trailing_stop:
    type: atr
    indicator: atr
    multiplier: 1.0  # Trail stop at 1x ATR
  
  max_positions: 1
```

#### Testing Workflow

```bash
# 1. Validate
./trader validate ema_volume.yaml

# 2. Initial backtest on training data
./trader backtest \
  --strategy ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-01-01 \
  --end 2023-06-30 \
  --output results/ema_vol_train.json

# 3. Test on out-of-sample data
./trader backtest \
  --strategy ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-07-01 \
  --end 2023-12-31 \
  --output results/ema_vol_test.json

# 4. Compare performance
echo "=== Training Period ==="
./trader report --result results/ema_vol_train.json

echo "=== Testing Period ==="
./trader report --result results/ema_vol_test.json
```

**What to Check:**
- Test period should have similar (not necessarily better) metrics than training
- If test period is much worse, may be overfitted
- If test period is much better, may have gotten lucky (run more tests)

---

### Example 5: Parameter Optimization

**Goal:** Find optimal parameters for the EMA volume strategy.

#### Step 1: Define Parameter Ranges

Based on market characteristics:
- Fast EMA: 5-15 (too fast = noise, too slow = late entries)
- Slow EMA: 20-30 (keep reasonable range)
- Volume threshold: 1.0-2.0 (1.0 = no filter, 2.0 = very strict)

#### Step 2: Run Optimization

```bash
./trader optimize \
  --strategy ema_volume.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-01-01 \
  --end 2023-06-30 \
  --parameters "ema_fast.period:5:15:1,ema_slow.period:20:30:2,volume_ratio:1.0:2.0:0.2" \
  --objective sharpe \
  --parallel 8 \
  --output results/optimization.json \
  --csv results/all_combos.csv \
  --sensitivity results/sensitivity.csv \
  --top 20
```

This tests:
- Fast EMA: 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 (11 values)
- Slow EMA: 20, 22, 24, 26, 28, 30 (6 values)
- Volume: 1.0, 1.2, 1.4, 1.6, 1.8, 2.0 (6 values)
- Total: 11 × 6 × 6 = 396 combinations

#### Step 3: Analyze Results

```bash
# View best parameters
head -20 results/all_combos.csv

# Check sensitivity
cat results/sensitivity.csv
```

**Sensitivity Analysis Interpretation:**

```csv
parameter,min,max,mean,std_dev,impact
ema_fast.period,5,15,1.45,0.32,high
ema_slow.period,20,30,1.48,0.15,medium
volume_ratio,1.0,2.0,1.42,0.08,low
```

- **High impact**: Performance varies significantly with this parameter
- **Low impact**: Parameter doesn't matter much (may remove it)

#### Step 4: Validate Best Parameters

```bash
# Update strategy with best parameters (e.g., fast=9, slow=24, volume=1.4)
# Save as ema_volume_optimized.yaml

# Test on out-of-sample data
./trader backtest \
  --strategy ema_volume_optimized.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-07-01 \
  --end 2023-12-31 \
  --output results/ema_vol_optimized_test.json

# Compare with non-optimized
./trader report --result results/ema_vol_test.json > results/original.txt
./trader report --result results/ema_vol_optimized_test.json > results/optimized.txt
diff results/original.txt results/optimized.txt
```

---

### Example 6: Walk Forward Analysis

**Goal:** Test strategy robustness over time.

#### Run Walk Forward

```bash
./trader walkforward \
  --strategy ema_volume_optimized.yaml \
  --data data/BTCUSDT.csv \
  --train 2000 \
  --test 500 \
  --step 250 \
  --output results/walkforward.json \
  --csv-windows results/wf_windows.csv \
  --csv-aggregate results/wf_aggregate.csv \
  --csv-stability results/wf_stability.csv
```

#### Analyze Walk Forward Results

```bash
# 1. Overall summary
./trader report --result results/walkforward.json

# 2. Check window consistency
cat results/wf_windows.csv | awk -F',' '{print $1,$4}' | column -t

# 3. Check stability metrics
cat results/wf_stability.csv
```

**Red Flags to Watch:**
- **Degrading performance**: Each window worse than previous
- **High variance**: Some windows +50%, others -30%
- **Low profitable window %**: < 60% windows profitable
- **Negative out-of-sample**: Aggregate out-of-sample returns negative

**Good Signs:**
- **Stable returns**: Windows consistently positive
- **Similar in/out-of-sample**: Training and testing perform similarly
- **High stability scores**: > 0.70 for return and Sharpe stability

---

### Example 7: Monte Carlo Risk Analysis

**Goal:** Understand worst-case scenarios and confidence intervals.

#### Run Monte Carlo

```bash
./trader montecarlo \
  --result results/ema_vol_optimized_test.json \
  --simulations 10000 \
  --seed 42 \
  --method trade_reshuffle \
  --output results/montecarlo.json
```

#### Interpret Results

```bash
./trader report --result results/montecarlo.json
```

**Key Questions to Answer:**

1. **What's the worst realistic outcome?**
   - Check 5th percentile return
   - If 5th percentile is -10%, you could lose 10% even with "good" strategy

2. **What's the probability of success?**
   - Check P(Return > 0)
   - If 85%, you have 85% chance of profit (15% chance of loss)

3. **What's the maximum expected drawdown?**
   - Check 95th percentile max drawdown
   - This is the "worst case" you should prepare for

4. **Is the strategy robust?**
   - Tight confidence intervals = robust
   - Wide intervals = high uncertainty

**Example Interpretation:**

```
P(Return > 0): 87%           ← Good! High probability of profit
P(Return > 10%): 65%         ← Reasonable chance of good returns
P(Drawdown < -20%): 12%      ← 12% chance of severe drawdown

5th percentile return: +2%   ← Even in bad scenarios, still profitable
95th percentile drawdown: -18% ← Worst expected drawdown
```

---

## Advanced Examples

### Example 8: Multi-Timeframe Strategy

**Goal:** Use higher timeframe for trend, lower for entries.

#### Strategy File

`mtf_strategy.yaml`:

```yaml
strategy:
  name: multi_timeframe_trend
  version: "1"
  description: "Daily trend, 4H entries"

data:
  symbol: BTCUSDT
  timeframe: 4h
  
  # Resample to daily for trend filter
  transforms:
    - type: resample
      timeframe: 1d
      output: daily_data

indicators:
  # 4H indicators
  ema9_4h:
    type: ema
    source: close
    period: 9
    
  ema21_4h:
    type: ema
    source: close
    period: 21
    
  rsi_4h:
    type: rsi
    source: close
    period: 14
    
  atr_4h:
    type: atr
    period: 14
    
  # Daily trend filter
  ema200_1d:
    type: ema
    source: daily_data.close
    period: 200

entry:
  long:
    all:
      - gt: [close, ema200_1d]  # Above daily trend
      - cross_above: [ema9_4h, ema21_4h]  # 4H crossover
      - gt: [rsi_4h, 50]  # RSI confirmation

exit:
  long:
    any:
      - cross_below: [ema9_4h, ema21_4h]
      - lt: [close, ema200_1d]  # Price broke below daily trend

risk:
  position_size:
    type: risk_percent
    value: 0.015
    
  stop_loss:
    type: atr
    indicator: atr_4h
    multiplier: 2.0
    
  take_profit:
    type: risk_reward
    ratio: 3
    
  max_positions: 2
```

#### Test the Strategy

```bash
# Validate transforms work correctly
./trader validate-transforms \
  --strategy mtf_strategy.yaml \
  --data data/BTCUSDT.csv \
  --rows 200

# Run backtest
./trader backtest \
  --strategy mtf_strategy.yaml \
  --data data/BTCUSDT.csv \
  --output results/mtf.json
```

---

### Example 9: Stateful Strategy (Setup Pattern)

**Goal:** Wait for specific setup, then trade on trigger.

#### Strategy with State

`stateful_setup.yaml`:

```yaml
strategy:
  name: stateful_breakout
  version: "1"
  description: "Wait for consolidation, trade breakout"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  sma20:
    type: sma
    source: close
    period: 20
    
  bb_upper:
    type: bb
    source: close
    period: 20
    stddev: 2
    band: upper
    
  bb_lower:
    type: bb
    source: close
    period: 20
    stddev: 2
    band: lower
    
  bb_width:
    type: subtract
    left: bb_upper
    right: bb_lower
    
  bb_width_pct:
    type: divide
    left: bb_width
    right: sma20
    
  volume_avg:
    type: sma
    source: volume
    period: 20
    
  atr:
    type: atr
    period: 14

# State variables
state:
  consolidation_setup:
    type: bool
    default: false
    
  setup_price:
    type: float
    default: 0

# Setup detection: Narrow Bollinger Bands
rules:
  - when:
      lt: [bb_width_pct, 0.02]  # BB width < 2% (tight consolidation)
    set:
      consolidation_setup: true
      setup_price: close

  # Reset setup after 10 bars
  - when:
      bars_since:
        event: consolidation_setup
        greater_than: 10
    set:
      consolidation_setup: false

entry:
  long:
    all:
      - eq: [consolidation_setup, true]
      - gt: [close, bb_upper]  # Breakout above upper band
      - gt: [volume, volume_avg]  # Volume confirmation

exit:
  long:
    any:
      - lt: [close, sma20]  # Price below 20 SMA

risk:
  position_size:
    type: risk_percent
    value: 0.02
    
  stop_loss:
    type: atr
    indicator: atr
    multiplier: 1.5
    
  take_profit:
    type: risk_reward
    ratio: 2.5
    
  max_positions: 1
```

**Key Concepts:**
- `state`: Defines variables that persist between bars
- `rules`: Update state based on conditions
- `bars_since`: Track time since event occurred
- State allows complex multi-bar patterns

---

### Example 10: Complete Research Workflow

**Goal:** Full research process from idea to validation.

#### Step 1: Initial Idea Testing

```bash
# Create basic version
cat > idea_v1.yaml << 'EOF'
strategy:
  name: idea_v1
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  rsi:
    type: rsi
    period: 14

entry:
  long:
    all:
      - lt: [rsi, 30]

exit:
  long:
    any:
      - gt: [rsi, 70]

risk:
  position_size:
    type: percent_equity
    value: 0.1
  max_positions: 1
EOF

# Quick test
./trader backtest \
  --strategy idea_v1.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-01-01 \
  --end 2023-03-31 \
  --output results/idea_v1_quick.json
```

#### Step 2: Add Risk Management

```bash
# Create v2 with stops and targets
cat > idea_v2.yaml << 'EOF'
# ... (add ATR stops and targets)
EOF

./trader backtest \
  --strategy idea_v2.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-01-01 \
  --end 2023-06-30 \
  --output results/idea_v2_train.json
```

#### Step 3: Parameter Optimization

```bash
./trader optimize \
  --strategy idea_v2.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-01-01 \
  --end 2023-06-30 \
  --parameters "rsi.period:10:20:2,stop_loss.multiplier:1.0:3.0:0.5,take_profit.ratio:1.5:3.0:0.5" \
  --objective sharpe \
  --parallel 8 \
  --output results/optimization.json \
  --top 10
```

#### Step 4: Walk Forward Validation

```bash
# Use best parameters in idea_v3.yaml
./trader walkforward \
  --strategy idea_v3.yaml \
  --data data/BTCUSDT.csv \
  --train 2000 \
  --test 500 \
  --step 250 \
  --output results/walkforward.json
```

#### Step 5: Monte Carlo Assessment

```bash
./trader montecarlo \
  --result results/walkforward.json \
  --simulations 10000 \
  --output results/montecarlo.json
```

#### Step 6: Out-of-Sample Test

```bash
./trader backtest \
  --strategy idea_v3.yaml \
  --data data/BTCUSDT.csv \
  --start 2023-07-01 \
  --end 2023-12-31 \
  --output results/final_test.json
```

#### Step 7: Decision Matrix

Create `decision.md`:

```markdown
# Strategy Decision: idea_v3

## Metrics

| Metric | Training | Walk Forward | Out-of-Sample | Target | Pass? |
|--------|----------|--------------|---------------|--------|-------|
| Return | +23% | +18% | +21% | >15% | ✅ |
| Sharpe | 1.85 | 1.62 | 1.71 | >1.5 | ✅ |
| Max DD | -12% | -15% | -13% | <20% | ✅ |
| Win Rate | 54% | 51% | 52% | >45% | ✅ |
| Trades | 142 | 98 | 67 | >50 | ✅ |

## Monte Carlo (10,000 sims)

- P(Return > 0): 87%
- P(Return > 10%): 68%
- 5th percentile return: +3%
- 95th percentile drawdown: -22%

## Walk Forward

- Profitable windows: 72%
- Return stability: 0.74
- Sharpe stability: 0.69

## Decision: **GO TO PAPER TRADING**

Rationale:
- All metrics meet targets
- Consistent across validation methods
- Monte Carlo shows acceptable risk
- Walk Forward demonstrates stability
```

#### Step 8: Paper Trading

```bash
./trader paper \
  --strategy idea_v3.yaml \
  --symbol BTCUSDT \
  --price 45000 \
  --duration 7d \
  --interval 1m \
  --output results/paper_trading.json
```

**After 2 weeks of paper trading:**
- Compare paper vs backtest results
- Check execution assumptions (fills, slippage)
- Monitor for unexpected behaviors
- If paper results match backtest → Consider live trading (with small size)

---

## Real-World Scenarios

### Scenario A: "My Strategy Worked in Backtest but Failed Live"

**Problem:** Backtest showed 30% annual return, live trading lost 10% in 3 months.

**Diagnosis Checklist:**

```bash
# 1. Check for look-ahead bias
./trader validate strategy.yaml

# 2. Did you optimize on all data?
# Re-run with proper train/test split

# 3. Was Walk Forward consistent?
./trader walkforward \
  --strategy strategy.yaml \
  --data data.csv \
  --train 2000 --test 500

# 4. Check Monte Carlo probabilities
./trader montecarlo \
  --result backtest.json \
  --simulations 10000

# 5. Compare paper trading first
./trader paper \
  --strategy strategy.yaml \
  --symbol SYMBOL \
  --duration 14d
```

**Common Causes:**
1. **Overfitting**: Optimized parameters work only on specific data
2. **Look-ahead bias**: Using future data in indicators
3. **Unrealistic fills**: Assumed perfect execution
4. **Insufficient testing**: Didn't use Walk Forward or Monte Carlo
5. **Market regime change**: Strategy works in trending markets, but market became ranging

---

### Scenario B: "How Much Data Do I Need?"

**Answer depends on:**

1. **Timeframe**: More bars needed for lower timeframes
   - 1 minute: Need 6-12 months (250,000+ bars)
   - 1 hour: Need 2-3 years (17,500+ bars)
   - 1 day: Need 5-10 years (1,800+ bars)

2. **Indicator periods**: Need at least 2x max period for warmup
   - EMA 200: Need 400+ bars before first valid signal

3. **Statistical significance**: Need enough trades
   - Minimum: 30 trades for Monte Carlo
   - Good: 100+ trades
   - Excellent: 300+ trades

**Test Data Requirements:**

```bash
# Check your data
wc -l data/BTCUSDT.csv

# Calculate minimum needed
# If using EMA 200 and want 100 trades:
# Assume 1 trade per 20 bars = need 2000 bars
# Plus warmup: 2000 + 400 = 2400 bars minimum
```

---

### Scenario C: "My Optimization Shows Amazing Results"

**Red Flags:**

```bash
# Run optimization
./trader optimize \
  --strategy strategy.yaml \
  --data data.csv \
  --parameters "param:1:100:1" \
  --output results/opt.json \
  --sensitivity results/sensitivity.csv

# Check sensitivity
cat results/sensitivity.csv
```

**Warning Signs:**

1. **Too good to be true**: Sharpe > 5, Win Rate > 80%
2. **Single optimal parameter**: Only one parameter value works well
3. **Parameter cliff**: Small change = huge performance drop
4. **Too many parameters**: Optimizing 5+ parameters
5. **Narrow ranges**: param:9:11 (curve fitting to specific value)

**Solution:**

```bash
# 1. Use Walk Forward instead
./trader walkforward --strategy strategy.yaml --data data.csv

# 2. Check parameter stability
# Look for broad plateau in sensitivity analysis

# 3. Validate on new data
# Never trade based on optimization alone
```

---

### Scenario D: "Should I Use This Strategy?"

**Decision Framework:**

```bash
# 1. Full validation pipeline
./trader backtest --strategy strategy.yaml --data train.csv --output train.json
./trader walkforward --strategy strategy.yaml --data full.csv --output wf.json
./trader montecarlo --result wf.json --simulations 10000 --output mc.json
./trader backtest --strategy strategy.yaml --data test.csv --output test.json
```

**Minimum Requirements:**

| Metric | Minimum | Good | Excellent |
|--------|---------|------|-----------|
| Sharpe Ratio | 1.0 | 1.5 | 2.0+ |
| Win Rate | 40% | 50% | 60%+ |
| Profit Factor | 1.3 | 1.5 | 2.0+ |
| Max Drawdown | <25% | <15% | <10% |
| Number of Trades | 30 | 100 | 300+ |
| WF Stability (Return) | 0.60 | 0.70 | 0.80+ |
| MC P(Return > 0) | 75% | 85% | 90%+ |

**Go/No-Go Checklist:**

- [ ] All metrics meet minimum requirements
- [ ] Walk Forward shows consistency (>60% profitable windows)
- [ ] Monte Carlo 5th percentile is acceptable
- [ ] Out-of-sample test confirms in-sample
- [ ] Paper trading for 2+ weeks matches backtest
- [ ] Understand why strategy works (not black box)
- [ ] Comfortable with maximum expected drawdown
- [ ] Have risk capital to withstand worst-case scenario

**If 8/8 checked → Proceed to live (small size)**  
**If 6-7/8 → More testing needed**  
**If <6/8 → Don't trade this strategy**

---

## Tips and Tricks

### Quick Testing

```bash
# Test on recent data only (faster)
./trader backtest \
  --strategy strategy.yaml \
  --data data.csv \
  --start 2023-11-01 \
  --output quick_test.json

# Test with smaller data file
head -5000 data/BTCUSDT.csv > data/sample.csv
./trader backtest --strategy strategy.yaml --data data/sample.csv
```

### Batch Processing

```bash
# Test multiple strategies
for strategy in strategies/*.yaml; do
  echo "Testing $strategy"
  ./trader backtest \
    --strategy "$strategy" \
    --data data/BTCUSDT.csv \
    --output "results/$(basename $strategy .yaml).json"
done

# Compare results
for result in results/*.json; do
  echo "=== $(basename $result) ==="
  ./trader report --result "$result" | grep -E "Return|Sharpe|Drawdown"
done
```

### Strategy Comparison

```bash
# Create comparison report
cat > compare_strategies.sh << 'EOF'
#!/bin/bash
echo "Strategy,Return,Sharpe,Max DD,Win Rate,Trades"
for result in results/*.json; do
  name=$(basename $result .json)
  metrics=$(./trader report --result "$result" --format json | jq -r '"\(.total_return),\(.sharpe),\(.max_drawdown),\(.win_rate),\(.total_trades)"')
  echo "$name,$metrics"
done
EOF

chmod +x compare_strategies.sh
./compare_strategies.sh > strategy_comparison.csv
```

### Version Control for Strategies

```bash
# Track strategy evolution
git init strategy_research
cd strategy_research

# Commit each version
git add strategy_v1.yaml results/v1_backtest.json
git commit -m "v1: Initial RSI strategy - Sharpe 1.2"

git add strategy_v2.yaml results/v2_backtest.json
git commit -m "v2: Added ATR stops - Sharpe 1.5"

# View history
git log --oneline
```

---

## Common Patterns

### Pattern 1: The Starter Template

```yaml
strategy:
  name: STRATEGY_NAME
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  # Your indicators

entry:
  long:
    all:
      - # Your conditions

exit:
  long:
    any:
      - # Your conditions

risk:
  position_size:
    type: risk_percent
    value: 0.01
  stop_loss:
    type: atr
    indicator: atr
    multiplier: 2.0
  take_profit:
    type: risk_reward
    ratio: 2.0
  max_positions: 1
```

### Pattern 2: The Research Script

```bash
#!/bin/bash
# research.sh - Full validation pipeline

STRATEGY=$1
DATA=$2
NAME=$(basename $STRATEGY .yaml)

echo "=== Researching $NAME ==="

# 1. Validate
./trader validate $STRATEGY || exit 1

# 2. Train
./trader backtest \
  --strategy $STRATEGY \
  --data $DATA \
  --start 2023-01-01 \
  --end 2023-06-30 \
  --output results/${NAME}_train.json

# 3. Test
./trader backtest \
  --strategy $STRATEGY \
  --data $DATA \
  --start 2023-07-01 \
  --end 2023-12-31 \
  --output results/${NAME}_test.json

# 4. Walk Forward
./trader walkforward \
  --strategy $STRATEGY \
  --data $DATA \
  --train 2000 \
  --test 500 \
  --output results/${NAME}_wf.json

# 5. Monte Carlo
./trader montecarlo \
  --result results/${NAME}_wf.json \
  --simulations 10000 \
  --output results/${NAME}_mc.json

# 6. Reports
./trader report --result results/${NAME}_train.json > reports/${NAME}_train.txt
./trader report --result results/${NAME}_test.json > reports/${NAME}_test.txt
./trader report --result results/${NAME}_wf.json > reports/${NAME}_wf.txt
./trader report --result results/${NAME}_mc.json > reports/${NAME}_mc.txt

echo "=== Research complete for $NAME ==="
echo "Check reports/ directory for results"
```

Usage:
```bash
chmod +x research.sh
./research.sh strategies/my_strategy.yaml data/BTCUSDT.csv
```

---

**For more examples, check:**
- `strategies/examples/` directory
- `docs/USER_GUIDE.md` for command reference
- `docs/STRATEGY_DSL.md` for complete DSL documentation

---

**Version:** 1.0  
**Last Updated:** 2026-09-07
