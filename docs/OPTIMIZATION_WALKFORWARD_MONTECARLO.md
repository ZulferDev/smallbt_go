# Optimization, Walk Forward Analysis, and Monte Carlo

## Parameter Optimization

### Overview

Parameter optimization searches for the best strategy parameters using grid search:

```bash
./trader optimize \
  --strategy strategy.yaml \
  --data data.csv \
  --output optimization_results.json
```

### Configuration

Define optimizable parameters in strategy YAML:

```yaml
optimization:
  parameters:
    ema_fast_period:
      range: [5, 20]
      step: 1
    
    ema_slow_period:
      range: [20, 100]
      step: 5
    
    atr_multiplier:
      range: [1.0, 3.0]
      step: 0.25
  
  objective:
    type: sharpe  # Maximize Sharpe ratio
```

### Objective Functions

Supported objectives:

- `sharpe`: Maximize Sharpe ratio (annualized)
- `sortino`: Maximize Sortino ratio
- `total_return`: Maximize total return
- `profit_factor`: Maximize profit factor
- `calmar`: Maximize Calmar ratio
- `win_rate`: Maximize win rate
- `expectancy`: Maximize expectancy

### Grid Search

The optimizer performs exhaustive grid search:

```
Parameter 1: [5, 6, 7, ..., 20] → 16 values
Parameter 2: [20, 25, 30, ..., 100] → 17 values
Total combinations: 16 × 17 = 272 backtests
```

For more than 3 parameters, consider step sizes to reduce combinations.

### Output

Optimization results include:

```json
{
  "best_parameters": {
    "ema_fast_period": 12,
    "ema_slow_period": 55,
    "atr_multiplier": 1.5
  },
  "best_result": {
    "sharpe": 1.82,
    "total_return": 0.3542,
    "max_drawdown": -0.182
  },
  "parameter_sensitivity": {
    "ema_fast_period": [
      {"value": 5, "sharpe": 1.23},
      {"value": 6, "sharpe": 1.34}
    ]
  }
}
```

## Walk Forward Analysis

### Overview

Walk Forward Analysis validates strategy robustness through multiple train/test windows:

```bash
./trader walkforward \
  --strategy strategy.yaml \
  --data data.csv \
  --output wfa_results.json
```

### Configuration

Configure window sizes:

```yaml
walk_forward:
  train:
    bars: 2000    # Training window size (bars)
  test:
    bars: 500     # Testing window size (bars)
  step:
    bars: 500     # Window step size (bars)
```

### Process

1. **Training Phase**: Optimize parameters on training data
2. **Testing Phase**: Test optimized parameters on unseen data
3. **Roll Forward**: Slide window forward
4. **Repeat**: Until data exhausted

```
Initial:   [Train: 0-2000] [Test: 2000-2500]
Step 1:    [Train: 500-2500] [Test: 2500-3000]
Step 2:    [Train: 1000-3000] [Test: 3000-3500]
...
```

### Output

WFA results include out-of-sample performance:

```json
{
  "windows": [
    {
      "window": 1,
      "train_start": "2020-01-01",
      "train_end": "2021-06-01",
      "test_start": "2021-06-01",
      "test_end": "2021-09-01",
      "train_metrics": {
        "total_return": 0.45,
        "sharpe": 1.92
      },
      "test_metrics": {
        "total_return": 0.12,
        "sharpe": 0.85
      },
      "parameters": {
        "ema_fast": 12,
        "ema_slow": 55
      }
    }
  ],
  "aggregate_metrics": {
    "oos_return": 0.103,
    "oos_sharpe": 0.78,
    "parameter_stability": 0.65,
    "performance_decay": -0.28
  }
}
```

### Metrics

#### Out-of-Sample Performance
- **OOS Return**: Average return across test windows
- **OOS Sharpe**: Average Sharpe ratio across test windows
- **Performance Decay**: (Test Sharpe - Train Sharpe) / Train Sharpe

#### Parameter Stability
- **Parameter Variance**: How much optimized parameters vary between windows
- **Correlation**: Correlation between train and test performance

## Monte Carlo Analysis

### Overview

Monte Carlo simulation analyzes statistical properties through random sampling:

```bash
./trader montecarlo \
  --strategy strategy.yaml \
  --data data.csv \
  --simulations 10000 \
  --seed 42 \
  --output montecarlo_results.json
```

### Simulation Types

#### 1. Trade Reshuffling
Randomly reorder trades while preserving trade statistics:

```
Original: [Win, Loss, Win, Win, Loss]
Shuffled: [Loss, Win, Win, Loss, Win]
```

#### 2. Return Reshuffling
Randomly reorder returns while preserving distribution.

#### 3. Bootstrap Sampling
Randomly sample trades with replacement.

### Configuration

```yaml
monte_carlo:
  simulations: 10000
  seed: 42        # For reproducibility
  method: trade_reshuffle
  confidence_level: 0.95
```

### Output

Monte Carlo results include probability distributions:

```json
{
  "simulations": 10000,
  "distribution": {
    "total_return": {
      "mean": 0.25,
      "median": 0.23,
      "std": 0.18,
      "percentile_5": -0.08,
      "percentile_95": 0.52,
      "min": -0.45,
      "max": 0.89
    },
    "max_drawdown": {
      "mean": -0.18,
      "percentile_5": -0.32,
      "percentile_95": -0.08
    }
  },
  "probability_metrics": {
    "probability_of_ruin": 0.12,
    "probability_positive_return": 0.85,
    "confidence_interval_return": [-0.05, 0.55]
  }
}
```

### Risk Analysis

#### Probability of Ruin
Probability that equity drops below a threshold (e.g., 20% drawdown).

#### Confidence Intervals
Range within which metrics will fall with X% confidence.

#### Worst-Case Scenarios
Extreme outcomes from simulation (e.g., 1st percentile).

## Integrated Workflow

### Complete Research Pipeline

```bash
# 1. Parameter optimization
./trader optimize \
  --strategy base_strategy.yaml \
  --data train_data.csv \
  --output optimization.json

# 2. Apply optimized parameters
generate_strategy.py optimization.json base_strategy.yaml optimized_strategy.yaml

# 3. Walk Forward Analysis
./trader walkforward \
  --strategy optimized_strategy.yaml \
  --data full_data.csv \
  --output wfa_results.json

# 4. Monte Carlo validation
./trader montecarlo \
  --strategy optimized_strategy.yaml \
  --data test_data.csv \
  --simulations 10000 \
  --output montecarlo_results.json

# 5. Generate report
./trader report \
  --strategy optimized_strategy.yaml \
  --backtest backtest_results.json \
  --optimization optimization.json \
  --walkforward wfa_results.json \
  --montecarlo montecarlo_results.json \
  --output research_report.md
```

## Best Practices

### 1. Prevent Overfitting

- Use Walk Forward Analysis to validate out-of-sample performance
- Limit parameter space to avoid excessive optimization
- Monitor performance decay between train/test

### 2. Statistical Significance

- Run sufficient Monte Carlo simulations (10,000+)
- Ensure enough trades for meaningful statistics (100+)
- Check confidence intervals

### 3. Parameter Stability

Optimized parameters should:
- Vary minimally between walk-forward windows
- Produce consistent results across different time periods
- Not be overly sensitive to small changes

### 4. Robustness Testing

Test across:
- Multiple assets
- Different market conditions
- Various timeframes
- With and without fees/slippage

### 5. Interpret Results Correctly

- Out-of-sample results > in-sample results
- Statistical significance > point estimates
- Risk metrics > return metrics
- Consistency > maximum performance

## Performance Considerations

### Optimization

Grid search complexity grows exponentially:
```
2 parameters × 10 values each = 100 backtests
3 parameters × 10 values each = 1,000 backtests
4 parameters × 10 values each = 10,000 backtests
```

Strategies:
- Use step sizes to reduce parameter space
- Prioritize important parameters
- Use pre-screening to eliminate bad regions

### Monte Carlo

10,000 simulations with 100 trades each = 1,000,000 trade evaluations.

Optimize:
- Parallel execution
- Memory-efficient trade storage
- Incremental statistics calculation

## Common Pitfalls

### 1. Data Snooping

Optimizing on the same data used for final testing leads to inflated results.

**Solution**: Always reserve separate test data.

### 2. Multiple Comparison Problem

Testing many parameter combinations increases chance of finding lucky results.

**Solution**: Adjust confidence levels for multiple comparisons.

### 3. Survivorship Bias

Testing only successful assets ignores failed ones.

**Solution**: Test across multiple assets.

### 4. Look-Ahead Bias in Optimization

If optimization accesses future test data, results are invalid.

**Solution**: Strict chronological separation of train/test data.

## Example: Complete Strategy Validation

```yaml
# strategy_with_optimization.yaml
strategy:
  name: ema_cross_with_optimization

optimization:
  parameters:
    ema_fast:
      range: [5, 15]
      step: 1
    
    ema_slow:
      range: [20, 50]
      step: 2
    
    atr_multiplier:
      range: [1.0, 3.0]
      step: 0.25
  
  objective:
    type: sharpe

walk_forward:
  train:
    bars: 1000
  test:
    bars: 250
  step:
    bars: 250

monte_carlo:
  simulations: 5000
  method: trade_reshuffle
```

Run comprehensive validation:

```bash
# Full validation pipeline
./trader optimize --strategy strategy.yaml --data train.csv
./trader walkforward --strategy strategy.yaml --data full.csv
./trader montecarlo --strategy strategy.yaml --data test.csv

# Combined report
./trader report --strategy strategy.yaml \
  --backtest backtest.json \
  --optimization optimize.json \
  --walkforward wfa.json \
  --montecarlo mc.json
```

## Advanced Topics

### Bayesian Optimization

Future enhancement: More efficient parameter search for expensive backtests.

### Genetic Algorithms

Future: Evolutionary optimization for complex parameter spaces.

### Ensemble Methods

Combine multiple optimized strategies for improved robustness.

### Multi-Objective Optimization

Optimize multiple metrics simultaneously (e.g., Sharpe + drawdown).