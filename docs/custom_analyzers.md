# Custom Analyzers

Custom Analyzers allow you to extend smallbt_go with your own performance metrics without modifying the source code. This guide shows you how to create and register custom analyzers.

## Overview

The Custom Analyzer system provides:

- **Extensibility**: Add custom metrics without modifying core code
- **Type Safety**: Strongly-typed interface with error handling
- **Thread Safety**: Concurrent registration and calculation
- **Global Registry**: Package-level functions for easy access
- **Flexible Results**: Return any type (float64, string, map, struct, etc.)

## Interface

Custom analyzers implement the `CustomAnalyzer` interface:

```go
type CustomAnalyzer interface {
    // Name returns the unique name of this analyzer.
    Name() string

    // Calculate computes a custom metric from the analysis input.
    // The return value can be any type.
    Calculate(input AnalysisInput) (interface{}, error)
}
```

## AnalysisInput

The `AnalysisInput` struct provides all data needed for analysis:

```go
type AnalysisInput struct {
    InitialCash  float64            // Starting capital
    FinalEquity  float64            // Final equity
    StartTime    time.Time          // Backtest start time
    EndTime      time.Time          // Backtest end time
    TradeHistory []portfolio.Trade  // All completed trades
    EquityCurve  []EquityPoint      // Equity over time
    RiskFreeRate float64            // Annual risk-free rate
}
```

Each `portfolio.Trade` contains:

```go
type Trade struct {
    Symbol       string
    Side         Side
    EntryTime    time.Time
    EntryPrice   float64
    ExitTime     time.Time
    ExitPrice    float64
    Quantity     float64
    GrossPnL     float64
    EntryFee     float64
    ExitFee      float64
    NetPnL       float64
    Return       float64
    MAE          float64  // Maximum Adverse Excursion
    MFE          float64  // Maximum Favorable Excursion
    ExitReason   string
}
```

## Creating a Custom Analyzer

### Example 1: Winning Streak

Calculate the maximum number of consecutive winning trades:

```go
package main

import (
    "github.com/ZulferDev/smallbt_go/internal/analytics"
    "github.com/ZulferDev/smallbt_go/internal/portfolio"
)

type WinStreakAnalyzer struct{}

func (w *WinStreakAnalyzer) Name() string {
    return "max_win_streak"
}

func (w *WinStreakAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    if len(input.TradeHistory) == 0 {
        return 0, nil
    }

    maxStreak := 0
    currentStreak := 0

    for _, trade := range input.TradeHistory {
        if trade.NetPnL > 0 {
            currentStreak++
            if currentStreak > maxStreak {
                maxStreak = currentStreak
            }
        } else {
            currentStreak = 0
        }
    }

    return maxStreak, nil
}
```

### Example 2: Average Trade Duration

Calculate the average duration of all trades:

```go
type AvgDurationAnalyzer struct{}

func (a *AvgDurationAnalyzer) Name() string {
    return "avg_trade_duration"
}

func (a *AvgDurationAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    if len(input.TradeHistory) == 0 {
        return 0.0, nil
    }

    totalDuration := 0.0
    for _, trade := range input.TradeHistory {
        duration := trade.ExitTime.Sub(trade.EntryTime).Hours()
        totalDuration += duration
    }

    avgHours := totalDuration / float64(len(input.TradeHistory))
    return avgHours, nil
}
```

### Example 3: Risk-Reward Ratio

Calculate the average risk-reward ratio using MAE and MFE:

```go
type RiskRewardAnalyzer struct{}

func (r *RiskRewardAnalyzer) Name() string {
    return "avg_risk_reward"
}

func (r *RiskRewardAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    if len(input.TradeHistory) == 0 {
        return 0.0, nil
    }

    totalRR := 0.0
    count := 0

    for _, trade := range input.TradeHistory {
        if trade.MAE != 0 {
            rr := trade.MFE / trade.MAE
            totalRR += rr
            count++
        }
    }

    if count == 0 {
        return 0.0, nil
    }

    return totalRR / float64(count), nil
}
```

### Example 4: Complex Results (Struct)

Return a struct with multiple metrics:

```go
type TradeBreakdown struct {
    LongTrades  int     `json:"long_trades"`
    ShortTrades int     `json:"short_trades"`
    LongWinRate float64 `json:"long_win_rate"`
    ShortWinRate float64 `json:"short_win_rate"`
}

type SideAnalyzer struct{}

func (s *SideAnalyzer) Name() string {
    return "side_breakdown"
}

func (s *SideAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    if len(input.TradeHistory) == 0 {
        return TradeBreakdown{}, nil
    }

    breakdown := TradeBreakdown{}
    longWins := 0
    shortWins := 0

    for _, trade := range input.TradeHistory {
        if trade.Side == portfolio.Long {
            breakdown.LongTrades++
            if trade.NetPnL > 0 {
                longWins++
            }
        } else if trade.Side == portfolio.Short {
            breakdown.ShortTrades++
            if trade.NetPnL > 0 {
                shortWins++
            }
        }
    }

    if breakdown.LongTrades > 0 {
        breakdown.LongWinRate = float64(longWins) / float64(breakdown.LongTrades)
    }
    if breakdown.ShortTrades > 0 {
        breakdown.ShortWinRate = float64(shortWins) / float64(breakdown.ShortTrades)
    }

    return breakdown, nil
}
```

## Registering Analyzers

### Global Registry (Simple)

Use package-level functions for the global registry:

```go
package main

import "github.com/ZulferDev/smallbt_go/internal/analytics"

func init() {
    // Register analyzers on package initialization
    analytics.Register(&WinStreakAnalyzer{})
    analytics.Register(&AvgDurationAnalyzer{})
    analytics.Register(&RiskRewardAnalyzer{})
}
```

### Local Registry (Advanced)

Create a dedicated registry for specific use cases:

```go
registry := analytics.NewRegistry()
registry.Register(&WinStreakAnalyzer{})
registry.Register(&AvgDurationAnalyzer{})

// Calculate all custom metrics
results := registry.CalculateAll(input)
```

## Using Custom Analyzers

### Calculate Individual Analyzer

```go
result, err := analytics.Calculate("max_win_streak", input)
if err != nil {
    log.Fatalf("Analyzer failed: %v", err)
}

maxStreak := result.(int)
fmt.Printf("Maximum winning streak: %d trades\n", maxStreak)
```

### Calculate All Analyzers

```go
results := analytics.CalculateAll(input)
for name, result := range results {
    if err, isErr := result.(error); isErr {
        fmt.Printf("%s: ERROR - %v\n", name, err)
    } else {
        fmt.Printf("%s: %v\n", name, result)
    }
}
```

### Check if Analyzer Exists

```go
if analytics.Has("max_win_streak") {
    result, _ := analytics.Calculate("max_win_streak", input)
    // Use result
}
```

### List All Registered Analyzers

```go
names := analytics.List()
fmt.Println("Registered analyzers:", names)
```

## Integration with Backtests

You can calculate custom metrics after a backtest completes:

```go
// Run backtest
result := engine.Run(strategy, data)

// Prepare input for custom analyzers
input := analytics.AnalysisInput{
    InitialCash:  result.InitialCash,
    FinalEquity:  result.FinalEquity,
    StartTime:    result.StartTime,
    EndTime:      result.EndTime,
    TradeHistory: result.Trades,
    EquityCurve:  result.EquityCurve,
    RiskFreeRate: 0.02, // 2% annual risk-free rate
}

// Calculate all custom metrics
customMetrics := analytics.CalculateAll(input)

// Display results
for name, value := range customMetrics {
    fmt.Printf("%s: %v\n", name, value)
}
```

## Error Handling

Custom analyzers should return errors for invalid inputs:

```go
func (a *MyAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    if len(input.TradeHistory) == 0 {
        return nil, fmt.Errorf("no trades to analyze")
    }

    // Calculation logic
    result := doCalculation(input.TradeHistory)

    return result, nil
}
```

When using `CalculateAll()`, errors are captured in the results map:

```go
results := analytics.CalculateAll(input)
if err, isErr := results["my_analyzer"].(error); isErr {
    log.Printf("Analyzer failed: %v", err)
} else {
    // Use successful result
}
```

## Thread Safety

The registry is thread-safe and can be used concurrently:

```go
// Safe to call from multiple goroutines
go analytics.Register(&Analyzer1{})
go analytics.Register(&Analyzer2{})

// Safe to calculate concurrently
go analytics.Calculate("analyzer1", input)
go analytics.Calculate("analyzer2", input)
```

## Best Practices

### 1. Use Descriptive Names

```go
// Good
func (a *WinStreakAnalyzer) Name() string {
    return "max_win_streak"
}

// Bad
func (a *WinStreakAnalyzer) Name() string {
    return "ws"
}
```

### 2. Handle Edge Cases

```go
func (a *MyAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    // Always check for empty data
    if len(input.TradeHistory) == 0 {
        return 0, nil // Or return error
    }

    // Check for division by zero
    if denominator == 0 {
        return 0.0, nil
    }

    // Proceed with calculation
    return result, nil
}
```

### 3. Return Appropriate Types

```go
// Single value: return float64, int, string
func (a *SimpleAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    return 42.5, nil
}

// Multiple values: return struct or map
func (a *ComplexAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    return MyResult{Field1: 1, Field2: 2}, nil
}
```

### 4. Document Your Analyzers

```go
// WinStreakAnalyzer calculates the maximum number of consecutive winning trades.
// Returns 0 if there are no trades or no winning trades.
type WinStreakAnalyzer struct{}
```

### 5. Test Your Analyzers

```go
func TestWinStreakAnalyzer(t *testing.T) {
    analyzer := &WinStreakAnalyzer{}
    
    input := analytics.AnalysisInput{
        TradeHistory: []portfolio.Trade{
            {NetPnL: 100},  // Win
            {NetPnL: 50},   // Win
            {NetPnL: -20},  // Loss
            {NetPnL: 30},   // Win
            {NetPnL: 40},   // Win
            {NetPnL: 25},   // Win (streak = 3)
        },
    }

    result, err := analyzer.Calculate(input)
    if err != nil {
        t.Fatalf("Calculate failed: %v", err)
    }

    if result.(int) != 3 {
        t.Fatalf("Expected max streak 3, got %d", result.(int))
    }
}
```

## Advanced Examples

### Time-Based Analysis

```go
type MonthlyReturnAnalyzer struct{}

func (m *MonthlyReturnAnalyzer) Name() string {
    return "monthly_returns"
}

func (m *MonthlyReturnAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    monthlyReturns := make(map[string]float64)

    for _, trade := range input.TradeHistory {
        month := trade.ExitTime.Format("2006-01")
        monthlyReturns[month] += trade.NetPnL
    }

    return monthlyReturns, nil
}
```

### Statistical Analysis

```go
import "math"

type StdDevAnalyzer struct{}

func (s *StdDevAnalyzer) Name() string {
    return "return_std_dev"
}

func (s *StdDevAnalyzer) Calculate(input analytics.AnalysisInput) (interface{}, error) {
    if len(input.TradeHistory) == 0 {
        return 0.0, nil
    }

    // Calculate mean
    sum := 0.0
    for _, trade := range input.TradeHistory {
        sum += trade.Return
    }
    mean := sum / float64(len(input.TradeHistory))

    // Calculate variance
    variance := 0.0
    for _, trade := range input.TradeHistory {
        diff := trade.Return - mean
        variance += diff * diff
    }
    variance /= float64(len(input.TradeHistory))

    // Standard deviation
    stdDev := math.Sqrt(variance)
    return stdDev, nil
}
```

## API Reference

### Package Functions

```go
// Register adds an analyzer to the global registry
func Register(analyzer CustomAnalyzer) error

// Unregister removes an analyzer from the global registry
func Unregister(name string)

// Get retrieves an analyzer from the global registry
func Get(name string) CustomAnalyzer

// Has checks if an analyzer exists in the global registry
func Has(name string) bool

// List returns all analyzer names from the global registry
func List() []string

// Calculate runs a specific analyzer from the global registry
func Calculate(name string, input AnalysisInput) (interface{}, error)

// CalculateAll runs all analyzers from the global registry
func CalculateAll(input AnalysisInput) map[string]interface{}

// GlobalRegistry returns the global registry instance
func GlobalRegistry() *Registry
```

### Registry Methods

```go
// NewRegistry creates a new analyzer registry
func NewRegistry() *Registry

// Register adds an analyzer to the registry
func (r *Registry) Register(analyzer CustomAnalyzer) error

// Unregister removes an analyzer from the registry
func (r *Registry) Unregister(name string)

// Get retrieves an analyzer from the registry
func (r *Registry) Get(name string) CustomAnalyzer

// Has checks if an analyzer exists in the registry
func (r *Registry) Has(name string) bool

// List returns all analyzer names in the registry
func (r *Registry) List() []string

// Calculate runs a specific analyzer
func (r *Registry) Calculate(name string, input AnalysisInput) (interface{}, error)

// CalculateAll runs all analyzers in the registry
func (r *Registry) CalculateAll(input AnalysisInput) map[string]interface{}

// Clear removes all analyzers from the registry
func (r *Registry) Clear()
```

## Troubleshooting

### "Analyzer already registered" error

```go
// Problem: Trying to register the same analyzer twice
analytics.Register(&MyAnalyzer{})
analytics.Register(&MyAnalyzer{}) // Error!

// Solution: Check before registering
if !analytics.Has("my_analyzer") {
    analytics.Register(&MyAnalyzer{})
}
```

### Type assertion panics

```go
// Problem: Incorrect type assertion
result, _ := analytics.Calculate("my_analyzer", input)
value := result.(int) // Panics if result is not int

// Solution: Use safe type assertion
result, _ := analytics.Calculate("my_analyzer", input)
if value, ok := result.(int); ok {
    // Use value safely
} else {
    // Handle unexpected type
}
```

## See Also

- [Built-in Metrics](./analytics.md)
- [Backtest Results](./backtesting.md)
- [Trade Journal](./trade_journal.md)
