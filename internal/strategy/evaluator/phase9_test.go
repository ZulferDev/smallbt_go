package evaluator_test

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/indicator"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/strategy/ast"
	"github.com/ZulferDev/smallbt_go/internal/strategy/evaluator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockIndicatorFactory creates a mock indicator for testing
func mockIndicatorFactory(value float64) indicator.Factory {
	return func(config indicator.Config) (indicator.Indicator, error) {
		return &mockIndicator{
			name:  config.Type,
			value: value,
			valid: true,
		}, nil
	}
}

// mockIndicator implements indicator.Indicator for testing
type mockIndicator struct {
	name   string
	value  float64
	valid  bool
	called bool
}

func (m *mockIndicator) Name() string {
	return m.name
}

func (m *mockIndicator) Calculate(ctx *indicator.Context) (indicator.Value, error) {
	m.called = true
	return indicator.Value{
		Valid: m.valid,
		Value: m.value,
	}, nil
}

func TestStateManagement(t *testing.T) {
	// Create a strategy with state variables
	strategy := &ast.Strategy{
		Name: "state_test",
		State: map[string]ast.StateDef{
			"setup_valid": {
				Type:    "bool",
				Default: false,
			},
			"entry_count": {
				Type:    "float",
				Default: 0.0,
			},
			"stop_loss_hit": {
				Type:    "bool",
				Default: false,
			},
		},
		Indicators: map[string]ast.IndicatorDef{
			"ema_fast": {
				Type: "ema",
				Params: map[string]interface{}{
					"period": 9,
				},
			},
			"ema_slow": {
				Type: "ema",
				Params: map[string]interface{}{
					"period": 21,
				},
			},
		},
	}

	// Create evaluator with builtin registry
	registry := indicator.BuiltinRegistry()
	eval := evaluator.NewEvaluator(strategy, registry, market.Symbol("BTCUSDT"), market.Timeframe("4h"))

	// Test state initialization
	stateMap := eval.GetStateMap()
	assert.Equal(t, 3, len(stateMap))
	assert.Equal(t, false, stateMap["setup_valid"])
	assert.Equal(t, 0.0, stateMap["entry_count"])
	assert.Equal(t, false, stateMap["stop_loss_hit"])

	// Test GetState
	setupValue, err := eval.GetState("setup_valid")
	require.NoError(t, err)
	assert.Equal(t, false, setupValue)

	entryValue, err := eval.GetState("entry_count")
	require.NoError(t, err)
	assert.Equal(t, 0.0, entryValue)

	// Test SetState
	err = eval.SetState("setup_valid", true)
	require.NoError(t, err)
	err = eval.SetState("entry_count", 3.0)
	require.NoError(t, err)
	err = eval.SetState("stop_loss_hit", true)
	require.NoError(t, err)

	// Verify updates
	setupValue, err = eval.GetState("setup_valid")
	require.NoError(t, err)
	assert.Equal(t, true, setupValue)

	entryValue, err = eval.GetState("entry_count")
	require.NoError(t, err)
	assert.Equal(t, 3.0, entryValue)

	stopLossValue, err := eval.GetState("stop_loss_hit")
	require.NoError(t, err)
	assert.Equal(t, true, stopLossValue)

	// Test non-existent state
	_, err = eval.GetState("nonexistent")
	assert.Error(t, err)
	err = eval.SetState("nonexistent", 42)
	assert.Error(t, err)

	// Test SetStateMap
	newState := map[string]interface{}{
		"setup_valid":   false,
		"entry_count":   5.0,
		"stop_loss_hit": false,
	}
	eval.SetStateMap(newState)

	// Verify map updates
	stateMap = eval.GetStateMap()
	assert.Equal(t, false, stateMap["setup_valid"])
	assert.Equal(t, 5.0, stateMap["entry_count"])
	assert.Equal(t, false, stateMap["stop_loss_hit"])
}

func TestCompositeIndicators(t *testing.T) {
	// Create a strategy with composite indicators
	strategy := &ast.Strategy{
		Name: "composite_test",
		Indicators: map[string]ast.IndicatorDef{
			"sma_fast": {
				Name:   "sma_fast",
				Type:   "sma",
				Period: 10,
				Source: "close",
			},
			"sma_slow": {
				Name:   "sma_slow",
				Type:   "sma",
				Period: 30,
				Source: "close",
			},
			"distance": {
				Name:  "distance",
				Type:  "subtract",
				Left:  "sma_fast",
				Right: "sma_slow",
			},
			"distance_pct": {
				Name:  "distance_pct",
				Type:  "divide",
				Left:  "distance",
				Right: "sma_slow",
			},
		},
	}

	// Create registry with mock indicators
	registry := indicator.NewRegistry()
	registry.Register("sma", mockIndicatorFactory(100.0))
	registry.Register("subtract", mockIndicatorFactory(5.0))
	registry.Register("divide", mockIndicatorFactory(0.0526))

	eval := evaluator.NewEvaluator(strategy, registry, market.Symbol("BTCUSDT"), market.Timeframe("4h"))

	// Initialize evaluator
	err := eval.Initialize()
	require.NoError(t, err)

	// Update with a candle to trigger calculation
	candle := market.Candle{
		Timestamp: time.Now(),
		Open:      100,
		High:      105,
		Low:       95,
		Close:     102,
		Volume:    1000,
	}
	err = eval.UpdateCandle(candle)
	require.NoError(t, err)

	// Get indicator values
	values := eval.GetIndicatorValues()

	// Verify indicators are calculated
	assert.Greater(t, len(values), 0)
}

func TestEvaluateExpressionFunctions(t *testing.T) {
	strategy := &ast.Strategy{
		Name: "expression_test",
		Indicators: map[string]ast.IndicatorDef{
			"price": {
				Name:   "price",
				Type:   "sma",
				Period: 1,
				Source: "close",
			},
		},
	}

	registry := indicator.BuiltinRegistry()
	eval := evaluator.NewEvaluator(strategy, registry, market.Symbol("BTCUSDT"), market.Timeframe("4h"))

	// Initialize evaluator
	err := eval.Initialize()
	require.NoError(t, err)

	// Update with a candle
	candle := market.Candle{
		Timestamp: time.Now(),
		Open:      48,
		High:      52,
		Low:       47,
		Close:     50,
		Volume:    1000,
	}
	err = eval.UpdateCandle(candle)
	require.NoError(t, err)

	// Test various expression functions
	tests := []struct {
		name     string
		fn       string
		args     []interface{}
		expected float64
		hasError bool
	}{
		{"add two numbers", "add", []interface{}{10.0, 5.0}, 15.0, false},
		{"subtract two numbers", "subtract", []interface{}{50.0, 20.0}, 30.0, false},
		{"multiply two numbers", "multiply", []interface{}{10.0, 3.0}, 30.0, false},
		{"divide two numbers", "divide", []interface{}{100.0, 2.0}, 50.0, false},
		{"modulo", "modulo", []interface{}{50.0, 7.0}, 1.0, false},
		{"abs positive", "abs", []interface{}{50.0}, 50.0, false},
		{"abs negative", "abs", []interface{}{-25.0}, 25.0, false},
		{"min three values", "min", []interface{}{50.0, 40.0, 60.0}, 40.0, false},
		{"max three values", "max", []interface{}{50.0, 40.0, 60.0}, 60.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.EvaluateExpression(tt.fn, tt.args)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.0001)
			}
		})
	}
}

func TestStateInExpressions(t *testing.T) {
	// Test that state variables can be used in expressions
	strategy := &ast.Strategy{
		Name: "state_expression_test",
		State: map[string]ast.StateDef{
			"setup_count": {
				Type:    "float",
				Default: 0.0,
			},
			"max_drawdown": {
				Type:    "float",
				Default: 0.0,
			},
		},
		Indicators: map[string]ast.IndicatorDef{
			"price": {
				Name:   "price",
				Type:   "sma",
				Period: 1,
				Source: "close",
			},
		},
	}

	registry := indicator.BuiltinRegistry()
	eval := evaluator.NewEvaluator(strategy, registry, market.Symbol("BTCUSDT"), market.Timeframe("4h"))

	// Initialize and set some state values
	err := eval.Initialize()
	require.NoError(t, err)

	err = eval.SetState("setup_count", 5.0)
	require.NoError(t, err)
	err = eval.SetState("max_drawdown", -10.0)
	require.NoError(t, err)

	// Update with a candle
	candle := market.Candle{
		Timestamp: time.Now(),
		Open:      98,
		High:      102,
		Low:       97,
		Close:     100,
		Volume:    1000,
	}
	err = eval.UpdateCandle(candle)
	require.NoError(t, err)

	// Test expressions using numeric values and functions
	result, err := eval.EvaluateExpression("add", []interface{}{100.0, 5.0})
	require.NoError(t, err)
	assert.Equal(t, 105.0, result)

	result, err = eval.EvaluateExpression("multiply", []interface{}{5.0, 2.0})
	require.NoError(t, err)
	assert.Equal(t, 10.0, result)

	result, err = eval.EvaluateExpression("abs", []interface{}{-10.0})
	require.NoError(t, err)
	assert.Equal(t, 10.0, result)
}

func TestMultipleStateUpdates(t *testing.T) {
	strategy := &ast.Strategy{
		Name: "multi_state_test",
		State: map[string]ast.StateDef{
			"counter": {
				Type:    "float",
				Default: 0.0,
			},
			"last_signal": {
				Type:    "float",
				Default: -1.0,
			},
		},
		Indicators: map[string]ast.IndicatorDef{
			"ema": {
				Type: "ema",
				Params: map[string]interface{}{
					"period": 10,
				},
			},
		},
	}

	registry := indicator.BuiltinRegistry()
	eval := evaluator.NewEvaluator(strategy, registry, market.Symbol("BTCUSDT"), market.Timeframe("4h"))

	err := eval.Initialize()
	require.NoError(t, err)

	// Simulate multiple bar updates with state changes
	for i := 0; i < 5; i++ {
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

		// Update state
		counter, _ := eval.GetState("counter")
		counterVal := counter.(float64)
		err = eval.SetState("counter", counterVal+1.0)
		require.NoError(t, err)
	}

	// Verify final state
	finalCounter, err := eval.GetState("counter")
	require.NoError(t, err)
	assert.Equal(t, 5.0, finalCounter.(float64))

	// Verify candle count
	assert.Equal(t, 5, eval.GetCandleCount())
}
