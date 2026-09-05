# Performance Baseline Report

**Date:** 2026-09-05  
**Version:** v0.1.0 (MVP Complete)  
**Purpose:** Establish performance baseline before Phase 15 optimization

---

## Test Environment

- **CPU:** (to be profiled)
- **Memory:** (to be profiled)
- **Go Version:** 1.24.4
- **OS:** Linux

---

## Baseline Performance Tests

### Test 1: Simple Strategy (ema_cross.yaml)

| Dataset Size | Rows | Runtime | Result |
|-------------|------|---------|--------|
| Small (500h) | 500 | 10.85ms | 0 trades |
| Medium (2000h) | 2000 | 13.44ms | 0 trades |
| Large (5000h) | 5000 | 15.13ms | 0 trades |
| 5 years (43800h) | 43800 | 380.13ms | 0 trades |

**Observation:** Simple strategy scales well - ~0.009ms per candle

---

### Test 2: Complex Strategy (ema_volume.yaml)

| Dataset Size | Rows | Runtime | Trades | Result |
|-------------|------|---------|--------|--------|
| 5 years (43800h) | 43800 | **50.40s** | 5 | +298.71% |

**Observation:** Complex strategy with multiple indicators shows **132x slowdown**

**Performance Issue Identified:** ~1.15ms per candle (vs 0.009ms for simple strategy)

---

## Performance Analysis

### Current State
- ✅ Simple strategies: Fast (380ms for 5 years)
- ❌ Complex strategies: **Very slow (50s for 5 years)**
- 🎯 Target: <2 seconds for 5 years hourly data

### Performance Gap
- **Current:** 50.40 seconds
- **Target:** 2 seconds
- **Required improvement:** **25x faster**

---

## Suspected Bottlenecks

Based on the 132x slowdown with complex strategy:

1. **Indicator Recalculation** 🔴 HIGH PRIORITY
   - ema_volume.yaml has multiple indicators (EMA fast, EMA slow, volume SMA, volume ratio)
   - Likely recalculating indicators multiple times per candle
   - No caching mechanism visible

2. **Expression Evaluation** 🟡 MEDIUM PRIORITY
   - Complex conditions evaluated on every bar
   - Expression tree traversal overhead

3. **Memory Allocation** 🟡 MEDIUM PRIORITY
   - Possible excessive allocations in hot path
   - Need to profile heap allocations

4. **Data Structure Inefficiency** 🟢 LOW PRIORITY
   - Simple strategy is fast, so data loading is efficient
   - Bottleneck is in strategy evaluation, not data layer

---

## Phase 15 Optimization Targets

### Target 1: Indicator Caching 🔴 CRITICAL
**Expected improvement:** 15-20x faster

```go
// Current (suspected):
for each bar:
    calculate ema_fast
    calculate ema_slow
    calculate volume_sma
    calculate volume_ratio
    evaluate conditions

// Target:
for each bar:
    update ema_fast (O(1) incremental)
    update ema_slow (O(1) incremental)
    update volume_sma (O(1) incremental)
    update volume_ratio (O(1) calculation)
    evaluate conditions
```

### Target 2: Expression Optimization 🟡
**Expected improvement:** 2-3x faster

- Cache expression AST evaluation results
- Avoid redundant condition checks
- Short-circuit boolean operations

### Target 3: Memory Optimization 🟢
**Expected improvement:** 1.5-2x faster

- Reduce allocations in hot path
- Preallocate buffers where possible
- Use sync.Pool for temporary objects

---

## Profiling Plan

### Step 1: CPU Profiling
```bash
go build -o ./bin/trader ./cmd/trader
./bin/trader backtest \
  --strategy=strategies/examples/ema_volume.yaml \
  --data=data/BTCUSDT_5years.csv \
  --cpuprofile=cpu.prof

go tool pprof -http=:8080 cpu.prof
```

### Step 2: Memory Profiling
```bash
./bin/trader backtest \
  --strategy=strategies/examples/ema_volume.yaml \
  --data=data/BTCUSDT_5years.csv \
  --memprofile=mem.prof

go tool pprof -http=:8080 mem.prof
```

### Step 3: Benchmark Suite
```bash
go test -bench=. -benchmem ./internal/backtest/
go test -bench=. -benchmem ./internal/indicator/
go test -bench=. -benchmem ./internal/expression/
```

---

## Success Criteria for Phase 15

### Performance Targets
- [ ] 5 years hourly data: <2 seconds (currently 50s)
- [ ] 100 parameter optimization combinations: <10 seconds
- [ ] Memory usage: <500MB (need to measure baseline)

### Correctness Requirements
- [ ] All existing tests pass
- [ ] No race conditions (`go test -race`)
- [ ] Identical backtest results before/after optimization
- [ ] Golden test verification for determinism

---

## Next Steps

1. **This Week:**
   - [ ] Add CPU profiling flag to CLI
   - [ ] Run profiling on ema_volume strategy
   - [ ] Identify top 3 bottlenecks
   - [ ] Design indicator caching architecture

2. **Week 2:**
   - [ ] Implement indicator caching
   - [ ] Add cache invalidation logic
   - [ ] Verify correctness with tests
   - [ ] Measure performance improvement

3. **Week 3:**
   - [ ] Expression optimization
   - [ ] Memory optimization
   - [ ] Benchmark suite
   - [ ] Final verification

---

## Baseline Preservation

This document preserves the pre-optimization baseline for comparison.

**Golden Test Strategy:** ema_volume.yaml on BTCUSDT_5years.csv
- Runtime: 50.40s
- Trades: 5
- Return: +298.71%
- Win Rate: 60.00%

After optimization, these results must remain **identical** (determinism requirement).

---

**Status:** ✅ Baseline Established  
**Next:** Begin CPU profiling
