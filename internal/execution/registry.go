package execution

import (
	"fmt"
	"sync"

	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
)

// SlippageModelFactory is a function that creates a slippage model instance.
// Parameters can be passed to configure the model (e.g., percentage, factor).
type SlippageModelFactory func(params map[string]interface{}) (SlippageModel, error)

// Registry manages custom slippage model registration.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]SlippageModelFactory
}

// NewRegistry creates a new execution model registry.
func NewRegistry() *Registry {
	r := &Registry{
		factories: make(map[string]SlippageModelFactory),
	}
	// Register built-in models
	r.registerBuiltins()
	return r
}

// registerBuiltins registers the built-in slippage models.
func (r *Registry) registerBuiltins() {
	// Fixed slippage model
	r.factories["fixed"] = func(params map[string]interface{}) (SlippageModel, error) {
		amount, ok := params["amount"].(float64)
		if !ok {
			return nil, fmt.Errorf("fixed slippage model requires 'amount' parameter (float64)")
		}
		return NewFixedSlippageModel(amount), nil
	}

	// Percentage slippage model
	r.factories["percentage"] = func(params map[string]interface{}) (SlippageModel, error) {
		percentage, ok := params["percentage"].(float64)
		if !ok {
			return nil, fmt.Errorf("percentage slippage model requires 'percentage' parameter (float64)")
		}
		return NewPercentageSlippageModel(percentage), nil
	}

	// Volatility slippage model
	r.factories["volatility"] = func(params map[string]interface{}) (SlippageModel, error) {
		factor, ok := params["factor"].(float64)
		if !ok {
			return nil, fmt.Errorf("volatility slippage model requires 'factor' parameter (float64)")
		}
		maxSlippage := 0.0
		if max, ok := params["max_slippage"].(float64); ok {
			maxSlippage = max
		}
		return NewVolatilitySlippageModel(factor, maxSlippage), nil
	}

	// Volume slippage model
	r.factories["volume"] = func(params map[string]interface{}) (SlippageModel, error) {
		factor, ok := params["impact_factor"].(float64)
		if !ok {
			return nil, fmt.Errorf("volume slippage model requires 'impact_factor' parameter (float64)")
		}

		maxSlippage := 0.0
		if max, ok := params["max_slippage"].(float64); ok {
			maxSlippage = max
		}

		return NewVolumeSlippageModel(factor, maxSlippage), nil
	}

	// No slippage model
	r.factories["none"] = func(params map[string]interface{}) (SlippageModel, error) {
		return NewNoSlippageModel(), nil
	}
}

// Register adds a custom slippage model factory to the registry.
// Returns an error if a model with the same name already exists.
func (r *Registry) Register(name string, factory SlippageModelFactory) error {
	if name == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("slippage model %q already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Unregister removes a slippage model from the registry.
// Built-in models cannot be unregistered.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Protect built-in models
	builtins := map[string]bool{
		"fixed":      true,
		"percentage": true,
		"volatility": true,
		"volume":     true,
		"none":       true,
	}

	if builtins[name] {
		return fmt.Errorf("cannot unregister built-in model %q", name)
	}

	delete(r.factories, name)
	return nil
}

// Get retrieves a factory by name.
func (r *Registry) Get(name string) SlippageModelFactory {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.factories[name]
}

// Has checks if a model is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.factories[name]
	return exists
}

// List returns all registered model names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// Create instantiates a slippage model by name with given parameters.
func (r *Registry) Create(name string, params map[string]interface{}) (SlippageModel, error) {
	r.mu.RLock()
	factory := r.factories[name]
	r.mu.RUnlock()

	if factory == nil {
		return nil, fmt.Errorf("slippage model %q not found", name)
	}

	return factory(params)
}

// Clear removes all custom models (keeps built-ins).
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Save built-ins
	builtins := map[string]SlippageModelFactory{
		"fixed":      r.factories["fixed"],
		"percentage": r.factories["percentage"],
		"volatility": r.factories["volatility"],
		"volume":     r.factories["volume"],
		"none":       r.factories["none"],
	}

	// Clear all
	r.factories = builtins
}

// globalRegistry is the default global registry.
var globalRegistry = NewRegistry()

// Register adds a custom slippage model to the global registry.
func Register(name string, factory SlippageModelFactory) error {
	return globalRegistry.Register(name, factory)
}

// Unregister removes a custom slippage model from the global registry.
func Unregister(name string) error {
	return globalRegistry.Unregister(name)
}

// Get retrieves a factory from the global registry.
func Get(name string) SlippageModelFactory {
	return globalRegistry.Get(name)
}

// Has checks if a model is registered in the global registry.
func Has(name string) bool {
	return globalRegistry.Has(name)
}

// List returns all model names from the global registry.
func List() []string {
	return globalRegistry.List()
}

// Create instantiates a model from the global registry.
func Create(name string, params map[string]interface{}) (SlippageModel, error) {
	return globalRegistry.Create(name, params)
}

// GlobalRegistry returns the global registry instance.
func GlobalRegistry() *Registry {
	return globalRegistry
}

// Example custom slippage models for documentation.

// SpreadSlippageModel applies slippage based on bid-ask spread.
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

	if req.Side == order.OrderSideBuy {
		return slippage, nil
	}
	return -slippage, nil
}

// RegisterSpreadModel is a helper to register the spread slippage model.
func RegisterSpreadModel() error {
	return Register("spread", func(params map[string]interface{}) (SlippageModel, error) {
		multiplier, ok := params["multiplier"].(float64)
		if !ok {
			multiplier = 0.5 // Default 50% of range
		}
		return NewSpreadSlippageModel(multiplier), nil
	})
}
