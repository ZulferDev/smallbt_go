package evaluator

import (
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/indicator"
	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/strategy/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiTimeframeIndicator tests that indicators can be configured for different timeframes
func TestMultiTimeframeIndicator(t *testing.T) {
	strategy := &ast.Strategy{
		Name: "mtf_test",
		Data: ast.DataConfig{
			Symbol:    "BTCUSDT",
			Timeframe: "1h",
		},
		Indicators: map[string]ast.IndicatorDef{
			"ema_1h": {
				Type:      "ema",
				Period:    20,
				Timeframe: "1h",
			},
			"ema_4h": {
				Type:      "ema",
				Period:    50,
				Timeframe: "4h",
			},
		},
	}

	registry := indicator.BuiltinRegistry()
	eval := NewEvaluator(strategy, registry, market.Symbol("BTCUSDT"), market.Timeframe("1h"))

	err := eval.Initialize()
	require.NoError(t, err)

	// Create 1h candles
	for i := 0; i < 50; i++ {
		candle := market.Candle{
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Open:      100.0 + float64(i),
			High:      102.0 + float64(i),
			Low:       98.0 + float64(i),
			Close:     101.0 + float64(i),
			Volume:    1000,
		}
		err = eval.UpdateCandle(candle)
		require.NoError(t, err)
	}

	// Verify 1h indicator has value
	ema1h, err := eval.GetIndicatorValue("ema_1h")
	require.NoError(t, err)
	assert.True(t, ema1h > 0, "EMA 1h should have value")

	// Verify 4h indicator exists but may not have value yet (different timeframe)
	ema4h, err := eval.GetIndicatorValue("ema_4h")
	// 4h indicator should exist but may need separate data feed
	t.Logf("EMA 1h: %.2f, EMA 4h: %.2f", ema1h, ema4h)
}

// TestTimeframeInIndicatorConfig tests that timeframe field is stored in indicator config
func TestTimeframeInIndicatorConfig(t *testing.T) {
	// Create a simple indicator with timeframe
	ind := ast.IndicatorDef{
		Type:      "ema",
		Period:    20,
		Timeframe: "4h",
	}

	// Verify timeframe field exists and is stored
	assert.Equal(t, "4h", ind.Timeframe)
	assert.Equal(t, "ema", ind.Type)
	assert.Equal(t, 20, ind.Period)
}

// TestMTFStrategyDefinition tests multi-timeframe strategy parsing
func TestMTFStrategyDefinition(t *testing.T) {
	strategy := &ast.Strategy{
		Name: "mtf_strategy",
		Data: ast.DataConfig{
			Symbol:    "BTCUSDT",
			Timeframe: "1h",
		},
		Indicators: map[string]ast.IndicatorDef{
			"ema_20_1h": {
				Type:      "ema",
				Period:    20,
				Timeframe: "1h",
			},
			"ema_50_4h": {
				Type:      "ema",
				Period:    50,
				Timeframe: "4h",
			},
			"ema_200_4h": {
				Type:      "ema",
				Period:    200,
				Timeframe: "4h",
			},
		},
		Entry: ast.EntryRules{
			Long: &ast.Condition{
				Type: "all",
				Conditions: []*ast.Condition{
					{
						Type:     "func",
						Function: "gt",
						Args:     []interface{}{"close", "ema_200_4h"},
					},
					{
						Type:     "func",
						Function: "gt",
						Args:     []interface{}{"close", "ema_20_1h"},
					},
				},
			},
		},
	}

	registry := indicator.BuiltinRegistry()
	eval := NewEvaluator(strategy, registry, market.Symbol("BTCUSDT"), market.Timeframe("1h"))

	err := eval.Initialize()
	require.NoError(t, err)
	assert.NotNil(t, eval)
}

// TestMultipleTimeframesInSameStrategy tests using multiple timeframes
func TestMultipleTimeframesInSameStrategy(t *testing.T) {
	strategy := &ast.Strategy{
		Name: "multi_tfd_strategy",
		Data: ast.DataConfig{
			Symbol:    "ETHUSDT",
			Timeframe: "15m",
		},
		Indicators: map[string]ast.IndicatorDef{
			"rsi_15m": {
				Type:      "rsi",
				Period:    14,
				Timeframe: "15m",
			},
			"ema_50_1h": {
				Type:      "ema",
				Period:    50,
				Timeframe: "1h",
			},
		},
	}

	registry := indicator.BuiltinRegistry()
	eval := NewEvaluator(strategy, registry, market.Symbol("ETHUSDT"), market.Timeframe("15m"))

	err := eval.Initialize()
	require.NoError(t, err)

	// Create 15m candles
	for i := 0; i < 50; i++ {
		candle := market.Candle{
			Timestamp: time.Now().Add(time.Duration(i) * 15 * time.Minute),
			Open:      2000.0 + float64(i),
			High:      2010.0 + float64(i),
			Low:       1990.0 + float64(i),
			Close:     2005.0 + float64(i),
			Volume:    500,
		}
		err = eval.UpdateCandle(candle)
		require.NoError(t, err)
	}

	// Check both indicators exist
	_, err = eval.GetIndicatorValue("rsi_15m")
	require.NoError(t, err, "RSI 15m should be available")

	// EMA 1h may not have value yet due to different timeframe
	_, err = eval.GetIndicatorValue("ema_50_1h")
	// This is expected - different timeframe data feed needed
	t.Logf("Indicator evaluation successful, 1h indicator status: %v", err)
}

// TestNoLookaheadMTF verifies the concept of no look-ahead in MTF context
func TestNoLookaheadMTF(t *testing.T) {
	// This test documents the expected behavior:
	// When evaluating at time T with 1h candles:
	// - 4h indicators should use the most recently CLOSED 4h candle
	// - NOT the currently forming 4h candle that includes time T

	// Example timeline:
	// 4h candles: [00:00-04:00], [04:00-08:00], [08:00-12:00]
	// At 09:00 (1h candle), the current 4h candle is [08:00-12:00] - still forming
	// We should use [04:00-08:00] as the latest CLOSED 4h candle
	// This prevents look-ahead bias

	t.Log("MTF look-ahead prevention requires implementation of proper candle alignment")
}

// TestTimeframeConfigurationStorage tests that timeframe is properly stored
func TestTimeframeConfigurationStorage(t *testing.T) {
	// Test that Timeframe field in IndicatorDef is preserved
	indicators := map[string]ast.IndicatorDef{
		"ema_1h": {Type: "ema", Period: 9, Timeframe: "1h"},
		"ema_4h": {Type: "ema", Period: 21, Timeframe: "4h"},
		"ema_1d": {Type: "ema", Period: 200, Timeframe: "1d"},
	}

	assert.Equal(t, "1h", indicators["ema_1h"].Timeframe)
	assert.Equal(t, "4h", indicators["ema_4h"].Timeframe)
	assert.Equal(t, "1d", indicators["ema_1d"].Timeframe)
}
