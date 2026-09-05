package indicator

import (
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/stretchr/testify/assert"
)

func createTestCandles(count int) []market.Candle {
	candles := make([]market.Candle, count)
	for i := 0; i < count; i++ {
		// Create candles with increasing close prices: 100, 101, 102, ...
		price := 100.0 + float64(i)
		candles[i] = market.Candle{
			Open:   price,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1000.0,
		}
	}
	return candles
}

func TestSMAFactory(t *testing.T) {
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
			config := Config{Type: "sma", Period: tt.period, Source: "close"}
			sma, err := SMAFactory(config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sma)
				assert.Equal(t, "sma", sma.Name())
			}
		})
	}
}

func TestSMACalculate(t *testing.T) {
	tests := []struct {
		name        string
		period      int
		candleSize  int
		expectValid bool
		expectValue float64
	}{
		{
			name:        "exact period",
			period:      5,
			candleSize:  5,
			expectValid: true,
			expectValue: 102.0, // avg of 100,101,102,103,104
		},
		{
			name:        "more than period",
			period:      3,
			candleSize:  5,
			expectValid: true,
			expectValue: 103.0, // avg of 102,103,104 (last 3)
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
			config := Config{Type: "sma", Period: tt.period, Source: "close"}
			sma, err := SMAFactory(config)
			assert.NoError(t, err)

			candles := createTestCandles(tt.candleSize)
			ctx := &Context{
				Current:         candles[len(candles)-1],
				Candles:         candles,
				BarIndex:        tt.candleSize - 1,
				IndicatorValues: make(map[string]Value),
			}

			result, err := sma.Calculate(ctx)
			assert.NoError(t, err)

			if tt.expectValid {
				assert.True(t, result.Valid)
				assert.InDelta(t, tt.expectValue, result.Value, 0.001)
			} else {
				assert.False(t, result.Valid)
			}
		})
	}
}

func TestSMAWithVolume(t *testing.T) {
	config := Config{Type: "sma", Period: 3, Source: "volume"}
	sma, err := SMAFactory(config)
	assert.NoError(t, err)

	candles := make([]market.Candle, 5)
	for i := 0; i < 5; i++ {
		candles[i] = market.Candle{
			Open:   100.0,
			High:   101.0,
			Low:    99.0,
			Close:  100.0,
			Volume: 1000.0 + float64(i*100),
		}
	}

	ctx := &Context{
		Current:         candles[4],
		Candles:         candles,
		BarIndex:        4,
		IndicatorValues: make(map[string]Value),
	}

	result, err := sma.Calculate(ctx)
	assert.NoError(t, err)
	assert.True(t, result.Valid)
	// Last 3 volumes: 1200, 1300, 1400 -> avg = 1300
	assert.InDelta(t, 1300.0, result.Value, 0.001)
}
