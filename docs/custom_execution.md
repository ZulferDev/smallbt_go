# Custom Execution Models

Custom Execution Models allow you to extend smallbt_go with your own slippage calculation logic without modifying the source code. This guide shows you how to create and register custom slippage models.

## Overview

The Custom Execution Model system provides:

- **Extensibility**: Add custom slippage models without modifying core code
- **Factory Pattern**: Parameterized model creation
- **Thread Safety**: Concurrent registration and model creation
- **Built-in Models**: 5 pre-registered models (fixed, percentage, volatility, volume, none)
- **Global Registry**: Package-level functions for easy access

## Interface

Custom slippage models implement the `SlippageModel` interface:

```go
type SlippageModel interface {
    // CalculateSlippage computes slippage for an order.
    // Returns the slippage amount (positive for adverse, negative for favorable).
    CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error)
    
    // Name returns the model name for logging/debugging.
    Name() string
}
```

## Factory Pattern

Models are registered via factory functions:

```go
type SlippageModelFactory func(params map[string]interface{}) (SlippageModel, error)
```

This allows parameterized model creation at runtime.

## Built-in Models

### 1. Fixed Slippage

Applies a constant slippage amount:

```go
model, _ := execution.Create("fixed", map[string]interface{}{
    "amount": 1.0,  // Fixed slippage: $1.00
})
```

### 2. Percentage Slippage

Applies slippage as a percentage of fill price:

```go
model, _ := execution.Create("percentage", map[string]interface{}{
    "percentage": 0.001,  // 0.1% slippage
})
```

### 3. Volatility-Based Slippage

Uses candle range (high - low) as volatility proxy:

```go
model, _ := execution.Create("volatility", map[string]interface{}{
    "factor":       0.1,   // 10% of candle range
    "max_slippage": 10.0,  // Cap at $10
})
```

### 4. Volume-Based Slippage

Calculates slippage based on order size relative to volume:

```go
model, _ := execution.Create("volume", map[string]interface{}{
    "impact_factor": 1.0,   // Market impact multiplier
    "max_slippage":  5.0,   // Cap at $5
})
```

### 5. No Slippage

Perfect execution (for testing):

```go
model, _ := execution.Create("none", map[string]interface{}{})
```

## Creating a Custom Model

### Example 1: Spread-Based Slippage

Apply slippage based on estimated bid-ask spread:

```go
package main

import (
    "fmt"
    "github.com/ZulferDev/smallbt_go/internal/execution"
    "github.com/ZulferDev/smallbt_go/internal/market"
    "github.com/ZulferDev/smallbt_go/internal/order"
)

type SpreadSlippageModel struct {
    SpreadMultiplier float64
}

func NewSpreadSlippageModel(multiplier float64) *SpreadSlippageModel {
    return &SpreadSlippageModel{SpreadMultiplier: multiplier}
}

func (m *SpreadSlippageModel) Name() string {
    return "spread"
}

func (m *SpreadSlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
    if candle == nil {
        return 0, fmt.Errorf("candle required for spread slippage")
    }

    // Estimate spread from candle range
    spread := (candle.High - candle.Low) * m.SpreadMultiplier
    
    // Apply half spread as slippage
    slippage := spread / 2.0
    
    // Direction: buyers pay more, sellers receive less
    if req.Side == order.OrderSideBuy {
        return slippage, nil
    }
    return -slippage, nil
}
```

### Example 2: Time-Based Slippage

Higher slippage during market open/close:

```go
type TimeBasedSlippageModel struct {
    BaseSlippage   float64
    PeakMultiplier float64
}

func NewTimeBasedSlippageModel(base, peak float64) *TimeBasedSlippageModel {
    return &TimeBasedSlippageModel{
        BaseSlippage:   base,
        PeakMultiplier: peak,
    }
}

func (m *TimeBasedSlippageModel) Name() string {
    return "time_based"
}

func (m *TimeBasedSlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
    if candle == nil {
        return m.BaseSlippage, nil
    }

    hour := candle.Timestamp.Hour()
    
    // Higher slippage during first/last hour
    multiplier := 1.0
    if hour == 9 || hour == 15 { // Market open/close hours (example)
        multiplier = m.PeakMultiplier
    }
    
    slippage := m.BaseSlippage * multiplier
    
    if req.Side == order.OrderSideBuy {
        return slippage, nil
    }
    return -slippage, nil
}
```

### Example 3: Momentum-Based Slippage

Slippage increases with price momentum:

```go
type MomentumSlippageModel struct {
    BasePercentage float64
    MomentumFactor float64
}

func NewMomentumSlippageModel(base, factor float64) *MomentumSlippageModel {
    return &MomentumSlippageModel{
        BasePercentage: base,
        MomentumFactor: factor,
    }
}

func (m *MomentumSlippageModel) Name() string {
    return "momentum"
}

func (m *MomentumSlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
    if candle == nil {
        return fillPrice * m.BasePercentage, nil
    }

    // Calculate momentum: (close - open) / open
    momentum := 0.0
    if candle.Open != 0 {
        momentum = (candle.Close - candle.Open) / candle.Open
    }
    
    // Absolute momentum increases slippage
    momentumImpact := 1.0 + (math.Abs(momentum) * m.MomentumFactor)
    
    slippage := fillPrice * m.BasePercentage * momentumImpact
    
    if req.Side == order.OrderSideBuy {
        return slippage, nil
    }
    return -slippage, nil
}
```

## Registering Custom Models

### Global Registry (Simple)

Register at package initialization:

```go
package main

import "github.com/ZulferDev/smallbt_go/internal/execution"

func init() {
    // Register spread model
    execution.Register("spread", func(params map[string]interface{}) (execution.SlippageModel, error) {
        multiplier, ok := params["multiplier"].(float64)
        if !ok {
            multiplier = 0.5 // Default
        }
        return NewSpreadSlippageModel(multiplier), nil
    })
    
    // Register time-based model
    execution.Register("time_based", func(params map[string]interface{}) (execution.SlippageModel, error) {
        base := params["base"].(float64)
        peak := params["peak"].(float64)
        return NewTimeBasedSlippageModel(base, peak), nil
    })
}
```

### Local Registry (Advanced)

Create a dedicated registry:

```go
registry := execution.NewRegistry()
registry.Register("custom", myFactory)

model, err := registry.Create("custom", params)
```

## Using Custom Models

### Create Model Instance

```go
// Using global registry
model, err := execution.Create("spread", map[string]interface{}{
    "multiplier": 0.3,
})
if err != nil {
    log.Fatalf("Failed to create model: %v", err)
}

// Calculate slippage
req := order.OrderRequest{
    Side:     order.OrderSideBuy,
    Quantity: 1.0,
}

candle := &market.Candle{
    High: 105.0,
    Low:  95.0,
}

slippage, err := model.CalculateSlippage(req, 100.0, candle)
fmt.Printf("Slippage: $%.2f\n", slippage)
```

### Check Available Models

```go
models := execution.List()
for _, name := range models {
    fmt.Println("Available model:", name)
}

if execution.Has("spread") {
    fmt.Println("Spread model is registered")
}
```

## Parameter Validation

Always validate parameters in your factory:

```go
execution.Register("my_model", func(params map[string]interface{}) (execution.SlippageModel, error) {
    // Required parameter
    factor, ok := params["factor"].(float64)
    if !ok {
        return nil, fmt.Errorf("'factor' parameter (float64) is required")
    }
    
    // Validate range
    if factor < 0 || factor > 1.0 {
        return nil, fmt.Errorf("'factor' must be between 0 and 1")
    }
    
    // Optional parameter with default
    maxSlippage := 10.0
    if max, ok := params["max_slippage"].(float64); ok {
        maxSlippage = max
    }
    
    return NewMyModel(factor, maxSlippage), nil
})
```

## Slippage Direction Convention

**Important**: Slippage should be applied consistently:

- **Buy orders**: Positive slippage = pay MORE (adverse)
- **Sell orders**: Negative slippage = receive LESS (adverse)

Example:

```go
func (m *MyModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
    slippage := calculateSlippage() // Always positive
    
    if req.Side == order.OrderSideBuy {
        return slippage, nil  // Pay more
    }
    return -slippage, nil // Receive less
}
```

## Error Handling

Models should return errors for invalid conditions:

```go
func (m *MyModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
    if candle == nil {
        return 0, fmt.Errorf("candle data required for this model")
    }
    
    if fillPrice <= 0 {
        return 0, fmt.Errorf("invalid fill price: %f", fillPrice)
    }
    
    // Calculate slippage
    return slippage, nil
}
```

## Integration with Backtests

Custom models can be used in backtest configurations:

```go
// Create and register custom model
execution.Register("my_model", myFactory)

// Use in backtest
config := backtest.Config{
    Slippage: execution.Create("my_model", map[string]interface{}{
        "param1": 1.0,
        "param2": 2.0,
    }),
}
```

## Best Practices

### 1. Name Models Clearly

```go
// Good
func (m *SpreadSlippageModel) Name() string {
    return "spread"
}

// Bad
func (m *SpreadSlippageModel) Name() string {
    return "s"
}
```

### 2. Handle Edge Cases

```go
func (m *MyModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
    // Check for nil candle
    if candle == nil {
        return m.defaultSlippage, nil
    }
    
    // Check for zero/negative values
    if candle.Volume <= 0 {
        return m.minSlippage, nil
    }
    
    // Check for division by zero
    if candle.Open == 0 {
        return m.defaultSlippage, nil
    }
    
    // Normal calculation
    return calculateSlippage(), nil
}
```

### 3. Implement Reasonable Limits

```go
type MyModel struct {
    MinSlippage float64
    MaxSlippage float64
}

func (m *MyModel) CalculateSlippage(...) (float64, error) {
    slippage := calculateRawSlippage()
    
    // Clamp to reasonable bounds
    if slippage < m.MinSlippage {
        slippage = m.MinSlippage
    }
    if slippage > m.MaxSlippage {
        slippage = m.MaxSlippage
    }
    
    return applyDirection(slippage), nil
}
```

### 4. Document Parameters

```go
// SpreadSlippageModel calculates slippage based on estimated bid-ask spread.
// 
// Parameters:
//   - multiplier: Fraction of candle range to use as spread estimate (0.0-1.0)
//     Default: 0.5 (50% of high-low range)
//
// Behavior:
//   - Uses (High - Low) * multiplier as spread estimate
//   - Applies half-spread as slippage
//   - Requires candle data (returns error if nil)
type SpreadSlippageModel struct {
    SpreadMultiplier float64
}
```

### 5. Test Your Models

```go
func TestMyCustomModel(t *testing.T) {
    model := NewMyModel(1.0)
    
    req := order.OrderRequest{
        Side: order.OrderSideBuy,
        Quantity: 1.0,
    }
    
    candle := &market.Candle{
        High: 110.0,
        Low: 90.0,
    }
    
    slippage, err := model.CalculateSlippage(req, 100.0, candle)
    if err != nil {
        t.Fatalf("CalculateSlippage failed: %v", err)
    }
    
    // Verify slippage is reasonable
    if slippage < 0 || slippage > 20 {
        t.Fatalf("Unexpected slippage: %f", slippage)
    }
}
```

## Registry API Reference

### Package Functions

```go
// Register adds a custom model to the global registry
func Register(name string, factory SlippageModelFactory) error

// Unregister removes a custom model from the global registry
// Built-in models cannot be unregistered
func Unregister(name string) error

// Get retrieves a factory from the global registry
func Get(name string) SlippageModelFactory

// Has checks if a model is registered in the global registry
func Has(name string) bool

// List returns all model names from the global registry
func List() []string

// Create instantiates a model from the global registry
func Create(name string, params map[string]interface{}) (SlippageModel, error)

// GlobalRegistry returns the global registry instance
func GlobalRegistry() *Registry
```

### Registry Methods

```go
// NewRegistry creates a new registry (includes built-in models)
func NewRegistry() *Registry

// Register adds a custom model to the registry
func (r *Registry) Register(name string, factory SlippageModelFactory) error

// Unregister removes a custom model (protects built-ins)
func (r *Registry) Unregister(name string) error

// Get retrieves a factory by name
func (r *Registry) Get(name string) SlippageModelFactory

// Has checks if a model is registered
func (r *Registry) Has(name string) bool

// List returns all registered model names
func (r *Registry) List() []string

// Create instantiates a model with parameters
func (r *Registry) Create(name string, params map[string]interface{}) (SlippageModel, error)

// Clear removes all custom models (keeps built-ins)
func (r *Registry) Clear()
```

## Troubleshooting

### "Model already registered" error

```go
// Problem: Trying to register the same name twice
execution.Register("my_model", factory1)
execution.Register("my_model", factory2) // Error!

// Solution: Check before registering
if !execution.Has("my_model") {
    execution.Register("my_model", factory)
}
```

### "Cannot unregister built-in model" error

```go
// Problem: Trying to unregister built-in
execution.Unregister("fixed") // Error!

// Solution: Only unregister custom models
execution.Unregister("my_custom_model") // OK
```

### Parameter type assertion panics

```go
// Problem: Unsafe type assertion
factor := params["factor"].(float64) // Panics if not float64

// Solution: Safe type assertion
factor, ok := params["factor"].(float64)
if !ok {
    return nil, fmt.Errorf("'factor' parameter required (float64)")
}
```

## See Also

- [Slippage Models](./slippage.md)
- [Execution Simulation](./execution.md)
- [Backtesting](./backtesting.md)
