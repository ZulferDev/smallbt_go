package indicator

import (
	"fmt"
)

// Registry manages indicator factories and provides indicator creation.
type Registry struct {
	factories map[string]Factory
}

// Factory is a function that creates a new indicator instance from config.
type Factory func(config Config) (Indicator, error)

// NewRegistry creates a new indicator registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

// Register adds an indicator factory to the registry.
func (r *Registry) Register(indicatorType string, factory Factory) error {
	if _, exists := r.factories[indicatorType]; exists {
		return fmt.Errorf("indicator type %q already registered", indicatorType)
	}
	r.factories[indicatorType] = factory
	return nil
}

// Create creates a new indicator instance from configuration.
func (r *Registry) Create(config Config) (Indicator, error) {
	factory, exists := r.factories[config.Type]
	if !exists {
		return nil, fmt.Errorf("unknown indicator type: %q", config.Type)
	}
	return factory(config)
}

// Types returns all registered indicator types.
func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

// BuiltinRegistry returns the registry with all built-in indicators registered.
func BuiltinRegistry() *Registry {
	reg := NewRegistry()

	// Register built-in indicators
	reg.Register("sma", SMAFactory)
	reg.Register("ema", EMAFactory)
	reg.Register("rsi", RSIFactory)
	reg.Register("atr", ATRFactory)

	return reg
}

// Global registry instance for convenience
var globalRegistry = BuiltinRegistry()

// Register adds an indicator factory to the global registry.
func Register(indicatorType string, factory Factory) error {
	return globalRegistry.Register(indicatorType, factory)
}

// Create creates a new indicator instance from configuration using the global registry.
func Create(config Config) (Indicator, error) {
	return globalRegistry.Create(config)
}

// Types returns all registered indicator types in the global registry.
func Types() []string {
	return globalRegistry.Types()
}
