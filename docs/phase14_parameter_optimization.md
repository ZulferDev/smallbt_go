# Phase 14: Parameter Optimization - Specification

## Overview

Phase 14 implements automated parameter optimization for trading strategies, enabling users to find optimal configuration parameters without manually testing countless combinations.

## Requirements (AGENTS.md Lines 1536-1556)

### Core Requirements
1. **Grid search optimization** - Enumerate all parameter combinations
2. **Objective function** - Sharpe ratio, profit factor, expectancy, or custom metric
3. **Parameter ranges** - Support for integer and float ranges with step sizes
4. **Constraint validation** - Validate parameter constraints before optimization
5. **Optimization report** - Best parameters, performance metrics, parameter sensitivity

### Spec Requirements
- Parameter definition via YAML configuration
- Grid search algorithm implementation
- Support for multiple objective functions
- Deterministic optimization (seeded randomization if any)
- Export optimization results (JSON, CSV)

## Architecture Design

```
┌─────────────────────────────────────────────────────────────┐
│                      Parameter Optimizer                    │
├─────────────────────────────────────────────────────────────┤
│  1. Parse Strategy Configuration                            │
│  2. Extract Parameters to Optimize                          │
│  3. Generate Parameter Grid                                 │
│  4. Validate Constraints                                    │
│  5. Run Backtests for Each Combination                      │
│  6. Calculate Objective Metrics                             │
│  7. Find Optimal Parameters                                 │
│  8. Generate Optimization Report                            │
└─────────────────────────────────────────────────────────────┘
```

## Domain Model

### Optimizer Types

```go
type OptimizationType int

const (
    GridSearch OptimizationType = iota
    RandomSearch
    GeneticAlgorithm
    BayesianOptimization
)

type ObjectiveFunction int

const (
    SharpeRatio ObjectiveFunction = iota
    ProfitFactor
    Expectancy
    TotalReturn
    Custom
)
```

### Parameter Definition

```yaml
optimization:
  parameters:
    ema_fast:
      type: int
      min: 5
      max: 50
      step: 1
    ema_slow:
      type: int
      min: 20
      max: 200
      step: 5
    atr_multiplier:
      type: float
      min: 1.0
      max: 3.0
      step: 0.25
  objective:
    type: sharpe_ratio
    direction: max
  method: grid_search
  seed: 42
```

### Optimization Result

```go
type OptimizationResult struct {
    BestParameters map[string]interface{}
    BestObjective  float64
    AllResults     []OptimizationRun
    ObjectiveStats OptimizationStats
}

type OptimizationRun struct {
    Parameters map[string]interface{}
    Objective  float64
    Metrics    BacktestMetrics
}

type OptimizationStats struct {
    Mean    float64
    StdDev  float64
    Min     float64
    Max     float64
    Median  float64
}
```

## Implementation Plan

### Phase 14.0: Foundation
- [ ] Create `internal/optimize/` package
- [ ] Define core types and interfaces
- [ ] Implement parameter parser from YAML
- [ ] Implement grid generation

### Phase 14.1: Optimization Engine
- [ ] Implement grid search algorithm
- [ ] Implement objective function calculations
- [ ] Integrate with backtest framework
- [ ] Handle optimization state persistence

### Phase 14.2: Results and Reporting
- [ ] Implement optimization results aggregation
- [ ] Calculate statistics across runs
- [ ] Generate optimization reports (JSON, CSV)
- [ ] Export best parameters to YAML

### Phase 14.3: Validation and Testing
- [ ] Unit tests for all components
- [ ] Integration tests with sample strategies
- [ ] Regression tests for known optima
- [ ] Performance benchmarks

## CLI Commands

```bash
# Run parameter optimization
trader optimize \
  --strategy strategy.yaml \
  --data BTCUSDT.parquet \
  --output results.json

# View optimization results
trader report --optimize results.json

# Run optimization with specific parameters
trader optimize \
  --strategy strategy.yaml \
  --data BTCUSDT.parquet \
  --objective sharpe_ratio \
  --method grid_search
```

## API Design

### Optimizer Interface

```go
type Optimizer interface {
    Optimize() (*OptimizationResult, error)
    GetBestParameters() map[string]interface{}
    GetBestObjective() float64
    GetAllResults() []OptimizationRun
}
```

### GridSearch Implementation

```go
type GridSearch struct {
    config    OptimizationConfig
    strategy  *Strategy
    data      DataFeed
    seed      int64
}

func (gs *GridSearch) Optimize() (*OptimizationResult, error)
func (gs *GridSearch) generateGrid() []map[string]interface{}
```

## Validation Rules

### Parameter Constraints
- Minimum < Maximum
- Step > 0
- Integer steps must divide range appropriately
- Float precision handling

### Objective Function Constraints
- Must be valid metric from backtest
- Direction must be specified (max/min)
- Must be computable from backtest results

### Strategy Constraints
- All parameters must exist in strategy
- Parameter types must match strategy expectations
- No circular dependencies

## Research Applications

### 1. **Strategy Tuning**
Find optimal parameters for specific market conditions

### 2. **Parameter Sensitivity Analysis**
Understand how sensitive strategy is to parameter changes

### 3. **Overfitting Detection**
Compare in-sample vs out-of-sample performance

### 4. **Robustness Testing**
Verify optimal parameters work across different datasets

## Integration with Phase 13 (Monte Carlo)

After optimization, run Monte Carlo analysis:
1. Optimize parameters on training data
2. Validate with Monte Carlo on test data
3. Assess statistical significance of optimal parameters

## Performance Considerations

### Computational Complexity
- Grid search: O(n^k) where n = parameter values, k = number of parameters
- Grid search becomes impractical with many parameters
- Future: Random search, genetic algorithms for high-dimensional spaces

### Optimization Strategies
- **Grid Search**: Suitable for 1-3 parameters
- **Random Search**: For high-dimensional spaces
- **Evolutionary**: For non-convex, noisy landscapes

## Error Handling

```go
// Invalid parameter definition
ErrInvalidParameterDefinition
ErrParameterRangeInvalid
ErrStepSizeInvalid

// Optimization errors
ErrNoValidParameters
ErrObjectiveCalculationFailed
ErrConstraintViolation

// Runtime errors
ErrOptimizationTimeout
ErrMaxEvaluationsReached
```

## Testing Strategy

### Unit Tests
- [ ] Parameter parsing from YAML
- [ ] Grid generation with various ranges
- [ ] Objective function calculations
- [ ] Constraint validation
- [ ] Result aggregation

### Integration Tests
- [ ] End-to-end optimization workflow
- [ ] Grid search with sample strategy
- [ ] Multiple objective functions
- [ ] Large parameter grids
- [ ] Deterministic optimization

### Regression Tests
- [ ] Known optimal parameters for simple strategies
- [ ] Parameter sensitivity for known strategies
- [ ] Edge cases (single parameter, no optimization needed)

## Documentation Requirements

### User Guide
- How to define optimization parameters
- Choosing objective functions
- Interpreting optimization results
- Avoiding overfitting

### Developer Guide
- Extending with new optimization methods
- Adding new objective functions
- Custom constraint validation

## Future Enhancements

### Phase 14.1: Advanced Optimization
- Random search optimization
- Genetic algorithms
- Bayesian optimization

### Phase 14.2: Parallel Optimization
- Distributed grid search
- Multi-threaded optimization
- GPU acceleration for complex strategies

### Phase 14.3: Real-time Optimization
- Online parameter tuning
- Adaptive optimization
- Market regime detection

## Acceptance Criteria

Phase 14 is complete when:

1. **Core Functionality**
   - Grid search optimization works
   - Multiple objective functions supported
   - Parameter ranges parsed from YAML

2. **Output**
   - Best parameters identified
   - Statistics calculated
   - Results exported to JSON/CSV

3. **Testing**
   - All unit tests pass
   - Integration tests validate workflow
   - Example strategies optimized successfully

4. **Documentation**
   - User guide explains optimization
   - Example optimization configurations exist
   - API documentation complete

## References

- AGENTS.md lines 1536-1556: Optimization requirements
- AGENTS.md lines 1500-1513: Future optimization planning
- AGENTS.md line 1232: Seed 42 for reproducibility

---

**Status**: Specified
**Next Phase**: Phase 14 - Parameter Optimization
**Dependencies**: Phase 10 (Multi-Timeframe), Phase 5 (Backtest Core)
