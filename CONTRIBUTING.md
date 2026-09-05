# Contributing to Jcode Backtest Engine

Thank you for your interest in contributing to this quantitative trading research platform. This guide will help you understand the project architecture and contribution process.

## Getting Started

### Prerequisites
- Go 1.22 or later
- Git
- Basic understanding of quantitative trading concepts

### Development Setup

```bash
git clone https://github.com/1jehuang/jcode.git
cd jcode
go mod download
go test ./...
```

### Building the CLI

```bash
go build -o ./bin/trader ./cmd/trader
./bin/trader --help
```

## Project Architecture

Read [AGENTS.md](AGENTS.md) first. It defines:

- **15 completed development phases** (MVP)
- **11 planned future phases** (post-MVP roadmap in [ROADMAP.md](ROADMAP.md))
- **10 architectural invariants** that must never be violated
- **Core design principle**: "YAML is an interface, not the engine"

Key packages:

```
internal/
├── strategy/      # Strategy DSL parsing, AST, compilation
├── indicator/     # Indicator registry and implementations
├── expression/    # Expression AST and evaluation
├── backtest/      # Backtest engine event loop
├── order/         # Order model and lifecycle
├── execution/     # Execution simulation
├── portfolio/     # Portfolio accounting
├── risk/          # Risk management
├── analytics/     # Metrics and analysis
├── optimization/  # Parameter optimization
├── walkforward/   # Walk-forward analysis
├── montecarlo/    # Monte Carlo simulation
└── data/          # Data feeds and OHLCV handling
```

## Contribution Guidelines

### Before You Code

1. **Read AGENTS.md** - Understand the architecture first
2. **Check existing code** - Search for similar functionality
3. **Identify domain** - Which package does this belong in?
4. **Design first** - Sketch the change on paper before coding

### Code Style

Follow Go conventions:

```bash
go fmt ./...      # Automatic formatting
go vet ./...      # Static analysis
golangci-lint     # Additional linting
```

### Testing Requirements

Every meaningful behavior must have tests:

- **Unit tests**: Test individual functions in isolation
- **Integration tests**: Test cross-package interactions
- **Regression tests**: Every bug fix becomes a test
- **Golden tests**: For deterministic behavior verification

Minimal test template:

```go
func TestFeatureName(t *testing.T) {
    // Arrange
    input := setupTestData()
    
    // Act
    result := FunctionUnderTest(input)
    
    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

Run all tests:

```bash
go test -race -cover ./...
```

### Commit Messages

Use descriptive commits following conventional commits:

```
feat(package): Short description

Longer explanation if needed.

- Bullet point 1
- Bullet point 2

Fixes #123
```

Examples:
- `feat(indicator): Add MACD indicator`
- `fix(backtest): Correct look-ahead in cross detection`
- `docs(strategy_dsl): Update expression syntax examples`
- `refactor(portfolio): Simplify PnL calculation`

### Pull Request Process

1. **Create feature branch**: `git checkout -b feat/description`
2. **Make changes** with tests
3. **Verify**: 
   ```bash
   go fmt ./...
   go vet ./...
   go test -race ./...
   ```
4. **Push and create PR**: Describe what and why
5. **Respond to review**: Address feedback promptly

### What NOT to Do

❌ **Never violate architectural invariants**:
- Strategy logic must not require Go code
- No look-ahead bias ever
- Execution and signal generation must remain separate
- Risk management must be independent
- Always maintain determinism

❌ **Never introduce**:
- Hardcoded indicator lists in core engine
- YAML-specific logic in domain layers
- Global mutable state
- Silent error handling
- Implicit timezone conversions
- Unbounded memory growth

❌ **Never merge without**:
- All tests passing
- Code formatted
- Documentation updated
- Architecture review approval

## Common Contribution Types

### Adding a New Indicator

1. Create implementation in `internal/indicator/builtin/`
2. Register in indicator registry
3. Add unit tests
4. Update strategy examples
5. Document in `docs/INDICATORS_CONDITIONS.md`

Example:

```go
// internal/indicator/builtin/macd.go
type MACD struct {
    fastPeriod   int
    slowPeriod   int
    signalPeriod int
}

func (m *MACD) Calculate(ctx *EvaluationContext) ([]float64, error) {
    // Implementation
}
```

### Fixing a Bug

1. Write a test that reproduces the bug (fails)
2. Fix the bug (test passes)
3. Verify no regressions: `go test -race ./...`
4. Commit with `Fixes #issue-number`

### Improving Performance

1. Profile first: `go test -bench=. -benchmem ./...`
2. Identify bottleneck
3. Implement optimization
4. Verify improvement with before/after benchmarks
5. Ensure no semantic changes

### Adding Documentation

- Keep docs in `docs/` directory
- Use markdown format
- Include code examples
- Update table of contents in README
- Link from relevant sections

## Testing for Look-Ahead Bias

Critical for this project. Every indicator and expression must verify:

```go
func TestNoLookAhead(t *testing.T) {
    // Create two datasets where t+1 differs
    dataA := setupData(candles1)
    dataB := setupData(candles1) // Same except candle[t+1]
    dataB[len(dataB)-1] = differentCandle
    
    // Signal at t must be identical
    signalA := Evaluate(dataA, timeT)
    signalB := Evaluate(dataB, timeT)
    
    if signalA != signalB {
        t.Error("future data influenced signal")
    }
}
```

## Determinism Testing

All backtests must be deterministic:

```go
func TestDeterminism(t *testing.T) {
    result1 := RunBacktest(strategy, data)
    result2 := RunBacktest(strategy, data)
    
    if result1.Trades != result2.Trades {
        t.Error("backtest not deterministic")
    }
}
```

## Documentation Requirements

Update these when applicable:

- **AGENTS.md**: Architecture changes
- **README.md**: High-level changes
- **ROADMAP.md**: Future features
- **docs/**: Detailed guides
- **Code comments**: Complex logic
- **Examples**: New capabilities

## Release Process

1. Update version in `cmd/trader/main.go`
2. Update `CHANGELOG.md` (when created)
3. Commit: `chore: Bump version to vX.Y.Z`
4. Tag: `git tag vX.Y.Z`
5. Push: `git push origin vX.Y.Z`
6. GitHub Actions automatically builds and releases

## Getting Help

- **Architecture questions**: Read AGENTS.md
- **Code questions**: Check examples in `strategies/examples/`
- **Design discussions**: Open an issue
- **Bug reports**: Include minimal reproduction case

## Code of Conduct

Be respectful and constructive. Value:

- Technical accuracy
- Clear communication
- Learning from mistakes
- Helping others understand

## Compensation

This is an open source project. Contributions are voluntary.

However, significant contributors may be eligible for:
- Recognition in README
- Commit history credit
- Attribution in documentation

## Questions?

- Check existing [GitHub Issues](https://github.com/1jehuang/jcode/issues)
- Review pull request discussions
- Consult AGENTS.md for architectural questions
- Read example strategies for usage patterns

---

**Thank you for contributing to better quantitative trading research tools!**
