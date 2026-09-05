# Developer Guide

## Project Structure

```
backtest/
├── cmd/
│   └── trader/
│       └── main.go              # CLI entry point
├── internal/
│   ├── strategy/
│   │   ├── ast/                 # Abstract Syntax Tree
│   │   ├── parser/              # YAML → AST
│   │   ├── compiler/            # AST → runtime
│   │   ├── evaluator/           # Strategy evaluation
│   │   ├── state/               # State management
│   │   └── registry/            # Component registry
│   ├── expression/              # Expression evaluation
│   ├── indicator/
│   │   ├── builtin/             # SMA, EMA, RSI, ATR
│   │   └── registry/            # Indicator factory
│   ├── data/
│   │   ├── feed/                # Data feed interface
│   │   ├── csv/                 # CSV parser
│   │   └── resample/            # Timeframe resampling
│   ├── signal/                  # Signal generation
│   ├── order/                   # Order types
│   ├── execution/               # Execution simulation
│   ├── portfolio/               # Portfolio state
│   ├── risk/                    # Risk management
│   ├── backtest/                # Backtest engine
│   ├── analytics/               # Metrics calculation
│   ├── optimization/            # Parameter optimization
│   ├── walkforward/             # Walk Forward Analysis
│   └── montecarlo/              # Monte Carlo simulation
├── strategies/
│   ├── examples/                # Example strategies
│   └── tests/                   # Test strategies
├── data/                        # Sample data
├── tests/                       # Integration tests
├── docs/                        # Documentation
└── go.mod / go.sum              # Dependencies
```

## Adding a New Indicator

### Step 1: Define the Indicator

Create `internal/indicator/builtin/my_indicator.go`:

```go
package builtin

import (
    "github.com/ZulferDev/smallbt_go/internal/indicator"
)

type MyIndicator struct {
    period int
}

func (m *MyIndicator) Name() string {
    return "my_indicator"
}

func (m *MyIndicator) Calculate(ctx *indicator.EvaluationContext) indicator.Value {
    // Get required data
    closes := ctx.GetSeries("close")
    if len(closes) < m.period {
        return indicator.NoValue()
    }
    
    // Calculate indicator
    sum := 0.0
    for i := len(closes) - m.period; i < len(closes); i++ {
        sum += closes[i]
    }
    
    result := sum / float64(m.period)
    return indicator.FloatValue(result)
}
```

### Step 2: Register the Indicator

In `internal/indicator/registry/registry.go`:

```go
func RegisterBuiltinIndicators() {
    registry.Register("my_indicator", func(config map[string]interface{}) (Indicator, error) {
        period := int(config["period"].(float64))
        return &builtin.MyIndicator{period: period}, nil
    })
}
```

### Step 3: Add Tests

Create `internal/indicator/builtin/my_indicator_test.go`:

```go
package builtin

import (
    "testing"
    "github.com/ZulferDev/smallbt_go/internal/indicator"
)

func TestMyIndicator(t *testing.T) {
    ind := &MyIndicator{period: 3}
    
    ctx := indicator.NewEvaluationContext()
    ctx.SetSeries("close", []float64{1, 2, 3, 4, 5})
    
    result := ind.Calculate(ctx)
    
    expected := 4.0  // (3 + 4 + 5) / 3
    actual, _ := result.Float64()
    
    if actual != expected {
        t.Errorf("Expected %f, got %f", expected, actual)
    }
}
```

## Adding a New Condition

### Step 1: Define the Condition

In `internal/expression/conditions.go`:

```go
func EvaluateMyCondition(ctx *EvaluationContext, values ...Value) Value {
    if len(values) != 2 {
        return BoolValue(false)
    }
    
    left, ok1 := values[0].Float64()
    right, ok2 := values[1].Float64()
    
    if !ok1 || !ok2 {
        return BoolValue(false)
    }
    
    return BoolValue(left > right)
}
```

### Step 2: Register the Condition

In `internal/expression/registry.go`:

```go
func RegisterBuiltinConditions() {
    registry.RegisterCondition("my_condition", EvaluateMyCondition)
}
```

### Step 3: Update Parser

In `internal/strategy/parser/parser.go`, add parsing logic for the new condition.

## Adding a New Analyzer

### Step 1: Implement the Analyzer Interface

In `internal/analytics/my_analyzer.go`:

```go
package analytics

type MyAnalyzer struct{}

func (a *MyAnalyzer) Analyze(result *BacktestResult) Metric {
    // Calculate metric
    value := calculateMyMetric(result)
    
    return Metric{
        Name: "my_metric",
        Value: value,
        Description: "My custom metric",
    }
}
```

### Step 2: Register the Analyzer

In `internal/analytics/registry.go`:

```go
func RegisterBuiltinAnalyzers() {
    registry.RegisterAnalyzer("my_analyzer", &MyAnalyzer{})
}
```

## Testing Guidelines

### Unit Tests

Test individual components in isolation:

```go
func TestIndicatorCalculation(t *testing.T) {
    // Arrange
    ind := NewIndicator(period)
    ctx := setupContext()
    
    // Act
    result := ind.Calculate(ctx)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Integration Tests

Test complete workflows:

```go
func TestStrategyToBacktest(t *testing.T) {
    // Load strategy
    strategy := loadStrategy("test_strategy.yaml")
    
    // Load data
    data := loadTestData()
    
    // Run backtest
    results := runBacktest(strategy, data)
    
    // Verify results
    if results.TotalTrades != expectedTrades {
        t.Errorf("Expected %d trades, got %d", expectedTrades, results.TotalTrades)
    }
}
```

### Golden Tests

Create deterministic test datasets with known results:

```go
func TestGoldenBacktest(t *testing.T) {
    golden := loadGoldenData("golden_sma_cross.yaml")
    
    results := runBacktest(golden.Strategy, golden.Data)
    
    // Assert exact results
    assert.Equal(t, golden.ExpectedTrades, results.TotalTrades)
    assert.InDelta(t, golden.ExpectedReturn, results.TotalReturn, 0.0001)
}
```

## Code Style

### Naming Conventions

- **Interfaces**: `Indicator`, `DataFeed`, `Analyzer`
- **Structs**: `SMAIndicator`, `CSVDataFeed`, `BacktestResult`
- **Methods**: `Calculate()`, `GetNext()`, `Analyze()`
- **Constants**: `DEFAULT_PERIOD`, `MAX_LEVERAGE`

### Error Handling

Use error wrapping:

```go
func (p *Parser) Parse(yaml []byte) (*Strategy, error) {
    config := map[string]interface{}{}
    if err := unmarshaler.Unmarshal(yaml, &config); err != nil {
        return nil, fmt.Errorf("parse strategy YAML: %w", err)
    }
    // ...
}
```

### Comments

Document public functions:

```go
// Calculate computes the indicator value for the current context.
// Returns NoValue if insufficient data is available.
func (m *MyIndicator) Calculate(ctx *EvaluationContext) Value {
    // ...
}
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/indicator/builtin/...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./internal/indicator/builtin/
```

## Performance Profiling

### CPU Profiling

```bash
go test -cpuprofile=cpu.prof -bench=. ./internal/backtest/
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -memprofile=mem.prof -bench=. ./internal/backtest/
go tool pprof mem.prof
```

## CI/CD Integration

### Pre-commit Checks

```bash
#!/bin/bash
go fmt ./...
go vet ./...
go test ./...
```

### GitHub Actions

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go test -race ./...
```

## Debugging

### Debug Logging

Enable debug output:

```bash
BACKTEST_DEBUG=1 ./trader backtest --strategy strategy.yaml --data data.csv
```

### Pprof HTTP Server

In development, expose metrics:

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Then visit: `http://localhost:6060/debug/pprof/`

## Extending the Architecture

### Adding Custom Data Feeds

Implement the `DataFeed` interface:

```go
type MyDataFeed struct {
    // ...
}

func (d *MyDataFeed) Next() (*Candle, error) {
    // Return next candle or io.EOF
}

func (d *MyDataFeed) Reset() error {
    // Reset to beginning
}
```

### Adding Custom Execution Models

Implement the `Executor` interface:

```go
type MyExecutor struct {
    // ...
}

func (e *MyExecutor) ExecuteOrder(order *Order) (*Fill, error) {
    // Simulate execution
}
```

### Adding Custom Position Sizers

Implement the `PositionSizer` interface:

```go
type MyPositionSizer struct {
    // ...
}

func (p *MyPositionSizer) CalculateSize(ctx *SizingContext) (float64, error) {
    // Calculate position size
}
```

## Documentation Standards

- Keep docs in `docs/` directory
- Use Markdown format
- Include examples for complex features
- Link related documentation
- Update docs when changing architecture

## Release Process

1. Update version in `cmd/trader/main.go`
2. Update `CHANGELOG.md`
3. Tag release: `git tag v1.0.0`
4. Push tag: `git push origin v1.0.0`
5. Create release on GitHub with notes

## Common Gotchas

### 1. Look-Ahead Bias

Never access future data:

```go
// WRONG: Uses close[t+1]
signal[t] = close[t+1] > close[t]

// RIGHT: Uses only current or past data
signal[t] = close[t] > close[t-1]
```

### 2. Determinism

Avoid non-deterministic operations:

```go
// WRONG: Random number without seed
value := rand.Float64()

// RIGHT: Seeded random
rng := rand.New(rand.NewSource(seed))
value := rng.Float64()
```

### 3. Floating Point Precision

Use appropriate tolerances:

```go
// WRONG: Exact comparison
if value == expected {}

// RIGHT: With epsilon
epsilon := 1e-9
if math.Abs(value - expected) < epsilon {}
```

### 4. Timezone Issues

Always work in UTC:

```go
// WRONG: Implicit timezone
t := time.Parse("2006-01-02", "2024-01-01")

// RIGHT: Explicit UTC
t, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
```

## Getting Help

1. Read the AGENTS.md specification
2. Check existing code for patterns
3. Review tests for examples
4. Open an issue with reproduction steps