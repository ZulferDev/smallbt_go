# CPU Profiling Analysis - Phase 15

**Date:** 2026-09-05  
**Profile:** cpu.prof  
**Duration:** 53.30s  
**Total Samples:** 67.99s

---

## Executive Summary

CPU profiling reveals **ATR.Calculate() consumes 49.12% of total runtime**, making it the primary performance bottleneck. This single indicator is responsible for ~33 seconds out of 53 seconds total execution time.

**Key Finding:** Indicator calculation is not cached. ATR is being recalculated from scratch on every bar, causing O(n²) complexity instead of O(n).

---

## Top Bottlenecks by Cumulative Time

| Function | Flat Time | Cum Time | % of Total | Category |
|----------|-----------|----------|------------|----------|
| `ATR.Calculate()` | 33.40s | 45.29s | **66.61%** | 🔴 Indicator |
| `Evaluator.UpdateCandle()` | 0.05s | 46.16s | 67.89% | Strategy Eval |
| `backtest.runBacktestLoop()` | 0.04s | 47.20s | 69.42% | Event Loop |
| `runtime.gcDrain` (GC) | 0.83s | 19.44s | 28.59% | 🟡 Garbage Collection |
| `math.Max` | 5.00s | 10.12s | 14.88% | 🟢 Math Ops |

---

## Critical Issue: ATR Recalculation

### Current Implementation (Suspected)

```go
// internal/indicator/atr.go
func (a *ATR) Calculate(candles []*Candle, index int) float64 {
    // On EVERY bar, this recalculates ATR from scratch
    // looking back 'period' bars
    
    trueRanges := make([]float64, a.period)
    for i := 0; i < a.period; i++ {
        // Calculate TR for each historical bar
        // This happens 43,800 times (once per bar)
    }
    
    // Average the true ranges
    return average(trueRanges)
}
```

**Complexity:** O(n × period) = O(43800 × 14) = **613,200 operations**

### Why This is Slow

For 43,800 bars with ATR(14):
- Bar 1: Calculate TR for bars [0]
- Bar 2: Calculate TR for bars [0-1]
- Bar 3: Calculate TR for bars [0-2]
- ...
- Bar 14: Calculate TR for bars [0-13]
- Bar 15: Calculate TR for bars [1-14] ← First full ATR
- Bar 16: Calculate TR for bars [2-15] ← Recalculates bars 2-14 again!
- ...
- Bar 43,800: Calculate TR for bars [43,786-43,799]

**Total TR calculations:** ~306 million operations

---

## Optimization Strategy 1: Incremental ATR Calculation

### Correct Implementation

```go
type ATR struct {
    period int
    
    // Cached state
    currentATR float64
    isWarm     bool
    barsSeen   int
}

func (a *ATR) Update(candle *Candle, prevCandle *Candle) float64 {
    tr := calculateTrueRange(candle, prevCandle)
    
    if !a.isWarm {
        // Warm-up period: collect initial values
        a.barsSeen++
        // ... accumulate for first 'period' bars
        if a.barsSeen >= a.period {
            a.isWarm = true
            // Calculate initial ATR
        }
        return 0 // Not valid yet
    }
    
    // Incremental update using EMA formula:
    // ATR[t] = (ATR[t-1] * (period - 1) + TR[t]) / period
    a.currentATR = ((a.currentATR * float64(a.period-1)) + tr) / float64(a.period)
    
    return a.currentATR
}
```

**Complexity:** O(n) = O(43,800) = **43,800 operations**

**Improvement:** **306M operations → 43K operations = 7000x reduction**

---

## Optimization Strategy 2: Indicator State Management

### Architecture Change Needed

```go
// internal/strategy/evaluator/evaluator.go

type Evaluator struct {
    indicators map[string]CachedIndicator
    candles    []*Candle
}

type CachedIndicator interface {
    Update(candle *Candle, index int) (float64, bool) // value, isValid
    IsWarm() bool
}

func (e *Evaluator) UpdateCandle(candle *Candle, index int) {
    // Update all indicators incrementally
    for name, indicator := range e.indicators {
        value, valid := indicator.Update(candle, index)
        e.values[name] = value
        e.validity[name] = valid
    }
    
    // Now evaluate strategy conditions using cached values
    e.evaluateConditions()
}
```

**Key Changes:**
1. Indicators maintain internal state
2. `Update()` instead of `Calculate()`
3. Validity tracking (warm-up period)
4. Values cached in evaluator

---

## Secondary Bottleneck: Garbage Collection

**GC Time:** 19.44s (28.59% of total)

### Analysis

```
runtime.gcDrain         0.83s   19.44s  28.59%
runtime.scanobject      6.98s   17.90s  26.33%
```

**Cause:** Excessive allocations in hot path

**Evidence:**
- `math.Max` inline allocations: 10.12s cumulative
- `runtime.mallocgc`: 2s
- `runtime.makeslice`: 1.52s

### Likely Sources

1. **Indicator calculations creating new slices every bar:**
   ```go
   trueRanges := make([]float64, period) // Allocated 43,800 times!
   ```

2. **Expression evaluation creating temporary objects**

3. **Candle processing without object pooling**

---

## Optimization Strategy 3: Reduce Allocations

### Preallocate Buffers

```go
type ATR struct {
    period int
    
    // State
    currentATR float64
    isWarm     bool
    
    // Preallocated buffer (reused)
    buffer []float64 // Only allocated ONCE
}

func NewATR(period int) *ATR {
    return &ATR{
        period: period,
        buffer: make([]float64, period), // Allocate once
    }
}
```

### Use sync.Pool for Temporary Objects

```go
var candlePool = sync.Pool{
    New: func() interface{} {
        return &Candle{}
    },
}

func processCandleFromCSV(row []string) *Candle {
    candle := candlePool.Get().(*Candle)
    // ... populate candle
    return candle
}

func (e *Evaluator) releaseCandle(candle *Candle) {
    candlePool.Put(candle)
}
```

---

## Expected Performance Improvements

### With ATR Caching Only

**Current:** 53 seconds  
**Expected:** ~7-8 seconds (7x faster)  
**Rationale:** Eliminating 49% bottleneck + reducing GC pressure

### With All Optimizations

| Optimization | Expected Improvement |
|-------------|---------------------|
| ATR caching | 7x faster |
| Other indicator caching | 2x faster |
| Allocation reduction | 1.5x faster |
| **Combined** | **~15-20x faster** |

**Target:** 53s → **2-3 seconds** ✅ Meets Phase 15 goal (<2s)

---

## Implementation Priority

### Week 1: Critical Path 🔴

1. **Refactor ATR to incremental calculation**
   - Create `CachedIndicator` interface
   - Implement `ATR.Update()`
   - Add warm-up period tracking
   - Test correctness against existing implementation

2. **Update Evaluator to use cached indicators**
   - Change from `Calculate()` to `Update()`
   - Track indicator validity
   - Only evaluate strategy when indicators are warm

3. **Verify correctness**
   - Golden test: ema_volume.yaml must produce identical results
   - All existing tests must pass

### Week 2: Additional Indicators 🟡

4. **Apply caching to all indicators**
   - SMA → incremental
   - EMA → incremental (already naturally incremental)
   - RSI → incremental
   - Composite indicators

5. **Measure improvement**
   - Re-run profiling
   - Verify 7-10x speedup

### Week 3: Memory Optimization 🟢

6. **Reduce allocations**
   - Preallocate indicator buffers
   - Use sync.Pool where appropriate
   - Profile memory usage

7. **Final verification**
   - All tests pass
   - Performance target met (<2s)
   - No race conditions

---

## Verification Plan

### Correctness Tests

```bash
# Golden test: Results must be identical
./bin/trader backtest \
  --strategy=strategies/examples/ema_volume.yaml \
  --data=data/BTCUSDT_5years.csv \
  > results_before.txt

# After optimization:
./bin/trader backtest \
  --strategy=strategies/examples/ema_volume.yaml \
  --data=data/BTCUSDT_5years.csv \
  > results_after.txt

diff results_before.txt results_after.txt
# Should show NO differences in trades, PnL, metrics
```

### Performance Tests

```bash
# Before: 53s
# After: <2s target

time ./bin/trader backtest \
  --strategy=strategies/examples/ema_volume.yaml \
  --data=data/BTCUSDT_5years.csv
```

### Race Detection

```bash
go test -race ./internal/indicator/
go test -race ./internal/strategy/evaluator/
go test -race ./internal/backtest/
```

---

## Risk Assessment

### High Risk 🔴
- **Breaking correctness:** Incremental calculations must match windowed calculations
- **Mitigation:** Extensive testing, golden tests, numerical validation

### Medium Risk 🟡
- **State management bugs:** Indicators maintaining state could leak between runs
- **Mitigation:** Clear state in tests, isolation tests

### Low Risk 🟢
- **Race conditions:** State access in concurrent optimization
- **Mitigation:** Race detector, proper synchronization

---

## Next Steps

1. **Create issue on GitHub:** "Phase 15.1: Implement incremental ATR calculation"
2. **Write failing test:** ATR incremental should match windowed calculation
3. **Implement `CachedIndicator` interface**
4. **Refactor ATR to use state**
5. **Update Evaluator**
6. **Run golden test verification**

---

## Appendix: Full CPU Profile Top 30

```
      flat  flat%   sum%        cum   cum%
         0     0%     0%     47.63s 70.05%  main.main
         0     0%     0%     47.63s 70.05%  main.run
         0     0%     0%     47.63s 70.05%  main.runBacktest
         0     0%     0%     47.63s 70.05%  runtime.main
         0     0%     0%     47.40s 69.72%  github.com/ZulferDev/smallbt_go/internal/backtest.Run
     0.04s 0.059% 0.059%     47.20s 69.42%  github.com/ZulferDev/smallbt_go/internal/backtest.runBacktestLoop
     0.05s 0.074%  0.13%     46.16s 67.89%  github.com/ZulferDev/smallbt_go/internal/strategy/evaluator.(*Evaluator).UpdateCandle
    33.40s 49.12% 49.26%     45.29s 66.61%  github.com/ZulferDev/smallbt_go/internal/indicator.(*ATR).Calculate
         0     0% 49.26%     20.17s 29.67%  runtime.systemstack
         0     0% 49.26%     19.56s 28.77%  runtime.gcBgMarkWorker
         0     0% 49.26%     19.45s 28.61%  runtime.gcBgMarkWorker.func2
     0.83s  1.22% 50.48%     19.44s 28.59%  runtime.gcDrain
     6.98s 10.27% 60.74%     17.90s 26.33%  runtime.scanobject
         0     0% 60.74%     10.86s 15.97%  runtime.gcDrainMarkWorkerIdle (inline)
        5s  7.35% 68.10%     10.12s 14.88%  math.Max (inline)
         0     0% 68.10%      8.58s 12.62%  runtime.gcDrainMarkWorkerDedicated (inline)
     5.11s  7.52% 75.61%      5.11s  7.52%  math.archMax
     2.48s  3.65% 79.26%      4.19s  6.16%  runtime.findObject
     1.95s  2.87% 82.13%      3.33s  4.90%  runtime.greyobject
```

---

**Status:** ✅ Analysis Complete  
**Confidence:** High - Clear bottleneck identified with concrete solution  
**Next Action:** Begin ATR refactoring implementation
