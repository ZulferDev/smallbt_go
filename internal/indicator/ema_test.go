package indicator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEMAFactory(t *testing.T) {
	tests := []struct {
		name      string
		period    int
		expectErr bool
	}{
		{"valid period", 10, false},
		{"negative period", -5, true},
		{"zero period", 0, true},
		{"period of 1", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Type: "ema", Period: tt.period, Source: "close"}
			ema, err := EMAFactory(config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ema)
				assert.Equal(t, "ema", ema.Name())
			}
		})
	}
}

func TestEMACalculate(t *testing.T) {
	tests := []struct {
		name        string
		period      int
		candleSize  int
		expectValid bool
	}{
		{
			name:        "exact period",
			period:      5,
			candleSize:  5,
			expectValid: true,
		},
		{
			name:        "more than period",
			period:      3,
			candleSize:  10,
			expectValid: true,
		},
		{
			name:        "less than period",
			period:      10,
			candleSize:  5,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Type: "ema", Period: tt.period, Source: "close"}
			ema, err := EMAFactory(config)
			assert.NoError(t, err)

			candles := createTestCandles(tt.candleSize)
			ctx := &Context{
				Current:         candles[len(candles)-1],
				Candles:         candles,
				BarIndex:        tt.candleSize - 1,
				IndicatorValues: make(map[string]Value),
			}

			result, err := ema.Calculate(ctx)
			assert.NoError(t, err)

			if tt.expectValid {
				assert.True(t, result.Valid)
				// EMA should be a valid number
				assert.NotZero(t, result.Value)
			} else {
				assert.False(t, result.Valid)
			}
		})
	}
}

func TestEMAFirstValue(t *testing.T) {
	// First EMA value should be SMA
	config := Config{Type: "ema", Period: 5, Source: "close"}
	ema, err := EMAFactory(config)
	assert.NoError(t, err)

	candles := createTestCandles(5)
	ctx := &Context{
		Current:         candles[4],
		Candles:         candles,
		BarIndex:        4,
		IndicatorValues: make(map[string]Value),
	}

	result, err := ema.Calculate(ctx)
	assert.NoError(t, err)
	assert.True(t, result.Valid)

	// First EMA = SMA of the 5 values: (100+101+102+103+104)/5 = 102
	assert.InDelta(t, 102.0, result.Value, 0.001)
}

func TestEMAMultiplier(t *testing.T) {
	period := 10
	config := Config{Type: "ema", Period: period, Source: "close"}
	ema, err := EMAFactory(config)
	assert.NoError(t, err)

	emaImpl := ema.(*EMA)
	expectedMultiplier := 2.0 / (float64(period) + 1.0)
	assert.InDelta(t, expectedMultiplier, emaImpl.multiplier, 0.0001)
}

func TestEMADeterministic(t *testing.T) {
	config := Config{Type: "ema", Period: 5, Source: "close"}
	ema, err := EMAFactory(config)
	assert.NoError(t, err)

	candles := createTestCandles(15)

	// Calculate EMA twice with same data
	ctx1 := &Context{
		Current:         candles[len(candles)-1],
		Candles:         candles,
		BarIndex:        len(candles) - 1,
		IndicatorValues: make(map[string]Value),
	}

	ctx2 := &Context{
		Current:         candles[len(candles)-1],
		Candles:         candles,
		BarIndex:        len(candles) - 1,
		IndicatorValues: make(map[string]Value),
	}

	result1, err := ema.Calculate(ctx1)
	assert.NoError(t, err)

	result2, err := ema.Calculate(ctx2)
	assert.NoError(t, err)

	// Results should be identical
	assert.Equal(t, result1.Valid, result2.Valid)
	assert.InDelta(t, result1.Value, result2.Value, 0.0001)
}
