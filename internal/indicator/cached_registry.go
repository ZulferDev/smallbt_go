package indicator

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CachedFactory is a function that creates a cached indicator instance.
type CachedFactory func(config Config) (CachedIndicator, error)

// CachedRegistry manages both stateless and cached indicator factories.
// It prefers cached versions when available for better performance.
type CachedRegistry struct {
	statelessFactories map[string]Factory
	cachedFactories    map[string]CachedFactory
}

// NewCachedRegistry creates a new cached registry.
func NewCachedRegistry() *CachedRegistry {
	return &CachedRegistry{
		statelessFactories: make(map[string]Factory),
		cachedFactories:    make(map[string]CachedFactory),
	}
}

// RegisterStateless adds a stateless indicator factory.
func (r *CachedRegistry) RegisterStateless(indicatorType string, factory Factory) error {
	if _, exists := r.statelessFactories[indicatorType]; exists {
		return fmt.Errorf("stateless indicator type %q already registered", indicatorType)
	}
	r.statelessFactories[indicatorType] = factory
	return nil
}

// RegisterCached adds a cached indicator factory.
// This will be preferred over stateless when creating indicators.
func (r *CachedRegistry) RegisterCached(indicatorType string, factory CachedFactory) error {
	if _, exists := r.cachedFactories[indicatorType]; exists {
		return fmt.Errorf("cached indicator type %q already registered", indicatorType)
	}
	r.cachedFactories[indicatorType] = factory
	return nil
}

// CreateCached creates a cached indicator if available, otherwise wraps a stateless one.
func (r *CachedRegistry) CreateCached(config Config) (CachedIndicator, error) {
	// Try cached factory first
	if cachedFactory, exists := r.cachedFactories[config.Type]; exists {
		return cachedFactory(config)
	}

	// Fall back to stateless and wrap it
	if statelessFactory, exists := r.statelessFactories[config.Type]; exists {
		stateless, err := statelessFactory(config)
		if err != nil {
			return nil, err
		}
		return &statelessWrapper{
			indicator: stateless,
			period:    config.Period,
		}, nil
	}

	return nil, fmt.Errorf("unknown indicator type: %q", config.Type)
}

// HasCached returns true if a cached version exists for this indicator type.
func (r *CachedRegistry) HasCached(indicatorType string) bool {
	_, exists := r.cachedFactories[indicatorType]
	return exists
}

// Types returns all registered indicator types.
func (r *CachedRegistry) Types() []string {
	typeMap := make(map[string]bool)
	for t := range r.statelessFactories {
		typeMap[t] = true
	}
	for t := range r.cachedFactories {
		typeMap[t] = true
	}

	types := make([]string, 0, len(typeMap))
	for t := range typeMap {
		types = append(types, t)
	}
	return types
}

// statelessWrapper wraps a stateless indicator to implement CachedIndicator interface.
// This provides a fallback for indicators that don't have optimized cached versions yet.
type statelessWrapper struct {
	indicator Indicator
	period    int
	candles   []market.Candle
	isWarm    bool
}

// Name returns the indicator name.
func (w *statelessWrapper) Name() string {
	return w.indicator.Name()
}

// Calculate implements the stateless interface.
func (w *statelessWrapper) Calculate(ctx *Context) (Value, error) {
	return w.indicator.Calculate(ctx)
}

// Update implements CachedIndicator by buffering candles and calling Calculate.
// This is not optimized but provides compatibility.
func (w *statelessWrapper) Update(candle market.Candle, prevCandle *market.Candle) (Value, error) {
	w.candles = append(w.candles, candle)

	// Check warmup
	if !w.isWarm && len(w.candles) >= w.period+1 {
		w.isWarm = true
	}

	// Call the stateless Calculate with buffered candles
	ctx := &Context{
		Current:  candle,
		Candles:  w.candles,
		BarIndex: len(w.candles) - 1,
	}

	return w.indicator.Calculate(ctx)
}

// IsWarm returns true when enough candles have been collected.
func (w *statelessWrapper) IsWarm() bool {
	return w.isWarm
}

// Reset clears the buffered candles.
func (w *statelessWrapper) Reset() {
	w.candles = nil
	w.isWarm = false
}

// WarmupPeriod returns the warmup period.
func (w *statelessWrapper) WarmupPeriod() int {
	if w.period > 0 {
		return w.period + 1
	}
	return 1
}

// BuiltinCachedRegistry returns a registry with all built-in indicators.
// Cached versions are registered when available for better performance.
func BuiltinCachedRegistry() *CachedRegistry {
	reg := NewCachedRegistry()

	// Register stateless versions
	reg.RegisterStateless("sma", SMAFactory)
	reg.RegisterStateless("ema", EMAFactory)
	reg.RegisterStateless("rsi", RSIFactory)
	reg.RegisterStateless("atr", ATRFactory)

	// Register cached versions (preferred)
	reg.RegisterCached("atr", CachedATRFactory)

	// TODO: Add cached versions for other indicators as they are implemented
	// reg.RegisterCached("sma", CachedSMAFactory)
	// reg.RegisterCached("ema", CachedEMAFactory)
	// reg.RegisterCached("rsi", CachedRSIFactory)

	return reg
}
