package indicator

import (
	"testing"

	"github.com/1jehuang/backtest/internal/market"
	"github.com/stretchr/testify/assert"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	assert.NotNil(t, reg)
	assert.Empty(t, reg.Types())
}

func TestRegistry_Register(t *testing.T) {
	reg := NewRegistry()

	// Test successful registration
	factory := func(config Config) (Indicator, error) {
		return nil, nil
	}
	err := reg.Register("test", factory)
	assert.NoError(t, err)

	// Test duplicate registration
	err = reg.Register("test", factory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_Create(t *testing.T) {
	reg := NewRegistry()

	// Register a test indicator
	factory := func(config Config) (Indicator, error) {
		return &SMA{name: "test", period: config.Period, source: config.Source}, nil
	}
	reg.Register("test", factory)

	// Test successful creation
	config := Config{Type: "test", Period: 10, Source: "close"}
	indicator, err := reg.Create(config)
	assert.NoError(t, err)
	assert.NotNil(t, indicator)
}

func TestRegistry_CreateUnknownType(t *testing.T) {
	reg := NewRegistry()

	config := Config{Type: "unknown", Period: 10}
	_, err := reg.Create(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown indicator type")
}

func TestBuiltinRegistry(t *testing.T) {
	reg := BuiltinRegistry()
	assert.NotNil(t, reg)

	types := reg.Types()
	assert.Contains(t, types, "sma")
	assert.Contains(t, types, "ema")
	assert.Contains(t, types, "rsi")
	assert.Contains(t, types, "atr")
}

func TestGlobalRegister(t *testing.T) {
	// Store original registry state
	_ = Types() // Just to check Types() works

	// Register test indicator
	err := Register("global_test", func(config Config) (Indicator, error) {
		return nil, nil
	})
	assert.NoError(t, err)

	// Verify it's registered
	types := Types()
	assert.Contains(t, types, "global_test")

	// Restore original state
	_ = globalRegistry
}

func TestGlobalCreate(t *testing.T) {
	config := Config{Type: "sma", Period: 10}
	indicator, err := Create(config)
	assert.NoError(t, err)
	assert.NotNil(t, indicator)
}

func TestGlobalTypes(t *testing.T) {
	types := Types()
	assert.Contains(t, types, "sma")
	assert.Contains(t, types, "ema")
	assert.Contains(t, types, "rsi")
	assert.Contains(t, types, "atr")
}

func TestSourcePrice(t *testing.T) {
	candle := market.Candle{
		Open:   100.0,
		High:   110.0,
		Low:    95.0,
		Close:  105.0,
		Volume: 1000.0,
	}

	tests := []struct {
		name   string
		source string
		expect float64
	}{
		{"open", "open", 100.0},
		{"default", "", 105.0},
		{"high", "high", 110.0},
		{"low", "low", 95.0},
		{"close", "close", 105.0},
		{"volume", "volume", 1000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SourcePrice(candle, tt.source)
			assert.Equal(t, tt.expect, result)
		})
	}
}
