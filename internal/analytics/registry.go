package analytics

import (
	"fmt"
	"sync"
)

// CustomAnalyzer is the interface for user-defined analyzers.
// Users can implement this interface to create custom metrics.
type CustomAnalyzer interface {
	// Name returns the unique name of this analyzer.
	Name() string

	// Calculate computes a custom metric from the analysis input.
	// The return value can be any type (float64, string, map, struct, etc).
	Calculate(input AnalysisInput) (interface{}, error)
}

// Registry manages custom analyzer registration.
type Registry struct {
	mu        sync.RWMutex
	analyzers map[string]CustomAnalyzer
}

// NewRegistry creates a new analyzer registry.
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make(map[string]CustomAnalyzer),
	}
}

// Register adds a custom analyzer to the registry.
// Returns an error if an analyzer with the same name already exists.
func (r *Registry) Register(analyzer CustomAnalyzer) error {
	if analyzer == nil {
		return fmt.Errorf("analyzer cannot be nil")
	}

	name := analyzer.Name()
	if name == "" {
		return fmt.Errorf("analyzer name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.analyzers[name]; exists {
		return fmt.Errorf("analyzer %q already registered", name)
	}

	r.analyzers[name] = analyzer
	return nil
}

// Unregister removes a custom analyzer from the registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.analyzers, name)
}

// Get retrieves a custom analyzer by name.
// Returns nil if the analyzer is not found.
func (r *Registry) Get(name string) CustomAnalyzer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.analyzers[name]
}

// Has checks if an analyzer is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.analyzers[name]
	return exists
}

// List returns all registered analyzer names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.analyzers))
	for name := range r.analyzers {
		names = append(names, name)
	}
	return names
}

// Clear removes all registered analyzers.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.analyzers = make(map[string]CustomAnalyzer)
}

// CalculateAll runs all registered analyzers and returns a map of results.
// If an analyzer fails, the error is captured in the results map as an error value.
func (r *Registry) CalculateAll(input AnalysisInput) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]interface{}, len(r.analyzers))
	for name, analyzer := range r.analyzers {
		result, err := analyzer.Calculate(input)
		if err != nil {
			results[name] = fmt.Errorf("analyzer %q failed: %w", name, err)
		} else {
			results[name] = result
		}
	}
	return results
}

// Calculate runs a specific analyzer by name.
func (r *Registry) Calculate(name string, input AnalysisInput) (interface{}, error) {
	r.mu.RLock()
	analyzer := r.analyzers[name]
	r.mu.RUnlock()

	if analyzer == nil {
		return nil, fmt.Errorf("analyzer %q not found", name)
	}

	return analyzer.Calculate(input)
}

// globalRegistry is the default global registry for custom analyzers.
var globalRegistry = NewRegistry()

// Register adds a custom analyzer to the global registry.
// This is a convenience function for package-level registration.
func Register(analyzer CustomAnalyzer) error {
	return globalRegistry.Register(analyzer)
}

// Unregister removes a custom analyzer from the global registry.
func Unregister(name string) {
	globalRegistry.Unregister(name)
}

// Get retrieves a custom analyzer from the global registry.
func Get(name string) CustomAnalyzer {
	return globalRegistry.Get(name)
}

// Has checks if an analyzer is registered in the global registry.
func Has(name string) bool {
	return globalRegistry.Has(name)
}

// List returns all analyzer names from the global registry.
func List() []string {
	return globalRegistry.List()
}

// CalculateAll runs all analyzers from the global registry.
func CalculateAll(input AnalysisInput) map[string]interface{} {
	return globalRegistry.CalculateAll(input)
}

// Calculate runs a specific analyzer from the global registry.
func Calculate(name string, input AnalysisInput) (interface{}, error) {
	return globalRegistry.Calculate(name, input)
}

// GlobalRegistry returns the global registry instance.
// This is useful for advanced use cases that need direct access.
func GlobalRegistry() *Registry {
	return globalRegistry
}
