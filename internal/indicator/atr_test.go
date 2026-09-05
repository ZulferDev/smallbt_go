package indicator

import (
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/stretchr/testify/assert"
)

func createATRTestCandles() []market.Candle {
	// Create test data with known true ranges
	candles := []market.Candle{
		{Open: 100, High: 101, Low: 99, Close: 100.5},      // Day 0
		{Open: 100.5, High: 102, Low: 100, Close: 101.5},   // Day 1: TR = max(2, 1.5, 1.5) = 2
		{Open: 101.5, High: 103, Low: 101, Close: 102.0},   // Day 2: TR = max(2, 1.5, 1.5) = 2
		{Open: 102.0, High: 104, Low: 101.5, Close: 103.0}, // Day 3: TR = max(2.5, 2.5, 2) = 2.5
		{Open: 103.0, High: 105, Low: 102, Close: 104.0},   // Day 4: TR = max(3, 2, 2) = 3
	}
	return candles
}

func TestATRFactory(t *testing.T) {
	tests := []struct {
		name      string
		period    int
		expectErr bool
	}{
		{"valid period", 14, false},
		{"negative period", -5, true},
		{"zero period", 0, true},
		{"period of 1", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Type: "atr", Period: tt.period, Source: ""}
			atr, err := ATRFactory(config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, atr)
				assert.Equal(t, "atr", atr.Name())
			}
		})
	}
}

func TestATRCalculate(t *testing.T) {
	tests := []struct {
		name        string
		period      int
		candleSize  int
		expectValid bool
	}{
		{
			name:        "minimum required candles",
			period:      3,
			candleSize:  4, // period + 1
			expectValid: true,
		},
		{
			name:        "insufficient candles",
			period:      3,
			candleSize:  3, // less than period + 1
			expectValid: false,
		},
		{
			name:        "more than required candles",
			period:      2,
			candleSize:  10,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Type: "atr", Period: tt.period, Source: ""}
			atr, err := ATRFactory(config)
			assert.NoError(t, err)

			candles := make([]market.Candle, tt.candleSize)
			for i := 0; i < tt.candleSize; i++ {
				candles[i] = market.Candle{
					Open:  100.0 + float64(i),
					High:  102.0 + float64(i),
					Low:   99.0 + float64(i),
					Close: 101.0 + float64(i),
				}
			}

			ctx := &Context{
				Current:         candles[len(candles)-1],
				Candles:         candles,
				BarIndex:        tt.candleSize - 1,
				IndicatorValues: make(map[string]Value),
			}

			result, err := atr.Calculate(ctx)
			assert.NoError(t, err)

			if tt.expectValid {
				assert.True(t, result.Valid)
				// ATR should be positive
				assert.Greater(t, result.Value, 0.0)
			} else {
				assert.False(t, result.Valid)
			}
		})
	}
}

func TestATRWithSpecificData(t *testing.T) {
	config := Config{Type: "atr", Period: 3, Source: ""}
	atr, err := ATRFactory(config)
	assert.NoError(t, err)

	candles := createATRTestCandles()
	ctx := &Context{
		Current:         candles[len(candles)-1],
		Candles:         candles,
		BarIndex:        len(candles) - 1,
		IndicatorValues: make(map[string]Value),
	}

	result, err := atr.Calculate(ctx)
	assert.NoError(t, err)
	assert.True(t, result.Valid)

	// For our test data:
	// Day 1 TR = 2, Day 2 TR = 2, Day 3 TR = 2.5, Day 4 TR = 3
	// First ATR = (2 + 2 + 2.5) / 3 = 2.166...
	assert.Greater(t, result.Value, 2.0)
	assert.Less(t, result.Value, 3.0)
}

func TestATRDeterministic(t *testing.T) {
	config := Config{Type: "atr", Period: 5, Source: ""}
	atr, err := ATRFactory(config)
	assert.NoError(t, err)

	candles := make([]market.Candle, 20)
	for i := 0; i < 20; i++ {
		candles[i] = market.Candle{
			Open:  100.0 + float64(i),
			High:  102.0 + float64(i),
			Low:   99.0 + float64(i),
			Close: 101.0 + float64(i),
		}
	}

	// Calculate ATR twice with same data
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

	result1, err := atr.Calculate(ctx1)
	assert.NoError(t, err)

	result2, err := atr.Calculate(ctx2)
	assert.NoError(t, err)

	// Results should be identical
	assert.Equal(t, result1.Valid, result2.Valid)
	assert.InDelta(t, result1.Value, result2.Value, 0.0001)
}

func TestATRSimpleRange(t *testing.T) {
	// Test case where high-low is the dominant component
	candles := []market.Candle{
		{Open: 100, High: 101, Low: 99, Close: 100.5},
		{Open: 100.5, High: 101, Low: 100, Close: 100.5}, // Small range
	}

	config := Config{Type: "atr", Period: 1, Source: ""}
	atr, err := ATRFactory(config)
	assert.NoError(t, err)

	ctx := &Context{
		Current:         candles[1],
		Candles:         candles,
		BarIndex:        1,
		IndicatorValues: make(map[string]Value),
	}

	result, err := atr.Calculate(ctx)
	assert.NoError(t, err)
	assert.True(t, result.Valid)

	// First ATR = (1 + 1) / 2 = 1
	assert.InDelta(t, 1.0, result.Value, 0.001)
}
