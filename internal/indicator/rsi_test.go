package indicator

import (
	"testing"

	"github.com/1jehuang/backtest/internal/market"
	"github.com/stretchr/testify/assert"
)

func createRSITestCandles() []market.Candle {
	// Create test data: 7 days with alternating gains and losses
	// Prices: 100, 101, 100.5, 102, 101, 103, 102.5
	// This should give us predictable RSI values
	candles := []market.Candle{
		{Close: 100.0}, // Day 0
		{Close: 101.0}, // Day 1: +1
		{Close: 100.5}, // Day 2: -0.5
		{Close: 102.0}, // Day 3: +1.5
		{Close: 101.0}, // Day 4: -1
		{Close: 103.0}, // Day 5: +2
		{Close: 102.5}, // Day 6: -0.5
	}
	return candles
}

func TestRSIFactory(t *testing.T) {
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
			config := Config{Type: "rsi", Period: tt.period, Source: "close"}
			rsi, err := RSIFactory(config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rsi)
				assert.Equal(t, "rsi", rsi.Name())
			}
		})
	}
}

func TestRSICalculate(t *testing.T) {
	tests := []struct {
		name        string
		period      int
		candleSize  int
		expectValid bool
	}{
		{
			name:        "minimum required candles",
			period:      5,
			candleSize:  6, // period + 1
			expectValid: true,
		},
		{
			name:        "insufficient candles",
			period:      5,
			candleSize:  4, // less than period + 1
			expectValid: false,
		},
		{
			name:        "more than required candles",
			period:      3,
			candleSize:  10,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Type: "rsi", Period: tt.period, Source: "close"}
			rsi, err := RSIFactory(config)
			assert.NoError(t, err)

			candles := createTestCandles(tt.candleSize)
			ctx := &Context{
				Current:         candles[len(candles)-1],
				Candles:         candles,
				BarIndex:        tt.candleSize - 1,
				IndicatorValues: make(map[string]Value),
			}

			result, err := rsi.Calculate(ctx)
			assert.NoError(t, err)

			if tt.expectValid {
				assert.True(t, result.Valid)
				// RSI should be between 0 and 100
				assert.GreaterOrEqual(t, result.Value, 0.0)
				assert.LessOrEqual(t, result.Value, 100.0)
			} else {
				assert.False(t, result.Valid)
			}
		})
	}
}

func TestRSIWithSpecificData(t *testing.T) {
	config := Config{Type: "rsi", Period: 6, Source: "close"}
	rsi, err := RSIFactory(config)
	assert.NoError(t, err)

	candles := createRSITestCandles()
	ctx := &Context{
		Current:         candles[len(candles)-1],
		Candles:         candles,
		BarIndex:        len(candles) - 1,
		IndicatorValues: make(map[string]Value),
	}

	result, err := rsi.Calculate(ctx)
	assert.NoError(t, err)
	assert.True(t, result.Valid)

	// For our test data, RSI should be calculated
	// We'll just verify it's in the valid range
	assert.GreaterOrEqual(t, result.Value, 0.0)
	assert.LessOrEqual(t, result.Value, 100.0)
}

func TestRSIAllGains(t *testing.T) {
	// Test RSI with all positive price changes (RSI should be 100)
	config := Config{Type: "rsi", Period: 3, Source: "close"}
	rsi, err := RSIFactory(config)
	assert.NoError(t, err)

	candles := []market.Candle{
		{Close: 100.0},
		{Close: 101.0},
		{Close: 102.0},
		{Close: 103.0},
		{Close: 104.0}, // Need period + 1 = 4 candles for period=3
	}

	ctx := &Context{
		Current:         candles[len(candles)-1],
		Candles:         candles,
		BarIndex:        len(candles) - 1,
		IndicatorValues: make(map[string]Value),
	}

	result, err := rsi.Calculate(ctx)
	assert.NoError(t, err)
	assert.True(t, result.Valid)

	// All gains should result in RSI close to 100
	assert.InDelta(t, 100.0, result.Value, 1.0)
}

func TestRSIWithZeroLoss(t *testing.T) {
	config := Config{Type: "rsi", Period: 3, Source: "close"}
	rsi, err := RSIFactory(config)
	assert.NoError(t, err)

	// Create candles with only gains (no losses)
	candles := []market.Candle{
		{Close: 100.0},
		{Close: 102.0},
		{Close: 104.0},
		{Close: 106.0},
	}

	ctx := &Context{
		Current:         candles[len(candles)-1],
		Candles:         candles,
		BarIndex:        len(candles) - 1,
		IndicatorValues: make(map[string]Value),
	}

	result, err := rsi.Calculate(ctx)
	assert.NoError(t, err)
	assert.True(t, result.Valid)

	// When average loss = 0, RSI = 100
	assert.InDelta(t, 100.0, result.Value, 0.001)
}

func TestRSIDeterministic(t *testing.T) {
	config := Config{Type: "rsi", Period: 7, Source: "close"}
	rsi, err := RSIFactory(config)
	assert.NoError(t, err)

	candles := createTestCandles(20)

	// Calculate RSI twice with same data
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

	result1, err := rsi.Calculate(ctx1)
	assert.NoError(t, err)

	result2, err := rsi.Calculate(ctx2)
	assert.NoError(t, err)

	// Results should be identical
	assert.Equal(t, result1.Valid, result2.Valid)
	assert.InDelta(t, result1.Value, result2.Value, 0.0001)
}
