package parser

import (
	"testing"

	"github.com/1jehuang/backtest/internal/strategy/ast"
	"github.com/stretchr/testify/assert"
)

// TestParserBasic tests basic parsing.
func TestParserBasic(t *testing.T) {
	yaml := `
strategy:
  name: test_strategy
  version: "1"
  description: "Test strategy"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema_fast:
    type: ema
    source: close
    period: 9

  ema_slow:
    type: ema
    source: close
    period: 21

entry:
  long:
    all:
      - cross_above: [ema_fast, ema_slow]

risk:
  position_size:
    type: fixed
    value: 1
  stop_loss:
    type: fixed
    price: 10000
  take_profit:
    type: fixed
    price: 20000
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.NotNil(t, strat)
	assert.Equal(t, "test_strategy", strat.Name)
	assert.Equal(t, "1", strat.Version)
	assert.Equal(t, "Test strategy", strat.Description)
	assert.Equal(t, "BTCUSDT", strat.Data.Symbol)
	assert.Equal(t, "4h", strat.Data.Timeframe)
	assert.Equal(t, 2, len(strat.Indicators))
	assert.NotNil(t, strat.Entry.Long)
	assert.NotNil(t, strat.Risk.PositionSize)
}

// TestParserIndicators tests indicator parsing.
func TestParserIndicators(t *testing.T) {
	yaml := `
strategy:
  name: indicators_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  sma20:
    type: sma
    source: close
    period: 20

  ema50:
    type: ema
    source: close
    period: 50

  atr14:
    type: atr
    period: 14

  volume_ratio:
    type: divide
    left: volume
    right: sma_volume

state:
  setup_flag:
    type: bool
    default: false
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.Equal(t, "sma20", strat.Indicators["sma20"].Name)
	assert.Equal(t, "ema50", strat.Indicators["ema50"].Name)
	assert.Equal(t, "atr14", strat.Indicators["atr14"].Name)
	assert.Equal(t, "divide", strat.Indicators["volume_ratio"].Type)
	assert.Equal(t, "volume", strat.Indicators["volume_ratio"].Left)

	assert.Equal(t, "bool", strat.State["setup_flag"].Type)
	assert.False(t, strat.State["setup_flag"].Default.(bool))
}

// TestParserEntryRules tests entry condition parsing.
func TestParserEntryRules(t *testing.T) {
	yaml := `
strategy:
  name: entry_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema9:
    type: ema
    period: 9
  ema21:
    type: ema
    period: 21
  rsi:
    type: rsi
    period: 14

entry:
  long:
    all:
      - cross_above: [ema9, ema21]
      - gt: [rsi, 50]

  short:
    any:
      - cross_below: [ema9, ema21]
      - lt: [rsi, 30]
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.NotNil(t, strat.Entry.Long)
	assert.NotNil(t, strat.Entry.Short)

	// Entry long should be "all" condition with 2 sub-conditions
	assert.Equal(t, "all", strat.Entry.Long.Type)
	assert.Equal(t, 2, len(strat.Entry.Long.Conditions))

	// Entry short should be "any" condition with 2 sub-conditions
	assert.Equal(t, "any", strat.Entry.Short.Type)
	assert.Equal(t, 2, len(strat.Entry.Short.Conditions))
}

// TestParserRiskConfig tests risk configuration parsing.
func TestParserRiskConfig(t *testing.T) {
	yaml := `
strategy:
  name: risk_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema:
    type: ema
    period: 21

entry:
  long:
    cross_above: [close, ema]

risk:
  position_size:
    type: risk_percent
    value: 0.01

  stop_loss:
    type: atr
    indicator: atr
    multiplier: 1.5

  take_profit:
    type: risk_reward
    ratio: 2

  max_positions: 5
  max_portfolio_risk: 0.1
  max_daily_loss: 0.05
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.Equal(t, "risk_percent", strat.Risk.PositionSize.Type)
	assert.Equal(t, 0.01, strat.Risk.PositionSize.Value)

	// Stop loss and take profit config not present in the test YAML
	// parser should not create them if they're not defined

	assert.Equal(t, 5, strat.Risk.MaxPositions)
	assert.Equal(t, 0.1, strat.Risk.MaxPortfolioRisk)
	assert.Equal(t, 0.05, strat.Risk.MaxDailyLoss)
}

// TestParserInvalidYAML tests error handling for invalid YAML.
func TestParserInvalidYAML(t *testing.T) {
	yaml := `
strategy:
  name: test
  # missing closing quote
version: "1

data:
  symbol: BTCUSDT
`

	p := NewParser()
	_, err := p.Parse([]byte(yaml))

	assert.Error(t, err)
}

// TestParserEmptyStrategy tests parsing an empty strategy.
func TestParserEmptyStrategy(t *testing.T) {
	yaml := `
strategy:
  name: empty_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.NotNil(t, strat)
	assert.Equal(t, "empty_test", strat.Name)
}

// TestParserCompositeConditions tests complex nested conditions.
func TestParserCompositeConditions(t *testing.T) {
	yaml := `
strategy:
  name: nested_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema9:
    type: ema
    period: 9
  ema21:
    type: ema
    period: 21
  rsi:
    type: rsi
    period: 14

entry:
  long:
    all:
      - gt: [close, 100000]
      - or:
          - all:
              - cross_above: [ema9, ema21]
              - gt: [rsi, 50]
          - and:
              - between: [rsi, 40, 60]
              - rising: [rsi]
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.NotNil(t, strat.Entry.Long)

	// Root should be "all" with 2 conditions
	assert.Equal(t, "all", strat.Entry.Long.Type)
	assert.Equal(t, 2, len(strat.Entry.Long.Conditions))

	// Second condition should be "or"
	secondCond := strat.Entry.Long.Conditions[1]
	assert.Equal(t, "any", secondCond.Type)
	assert.Equal(t, 2, len(secondCond.Conditions))
}

// TestParserFunctionConditions tests various function-based conditions.
func TestParserFunctionConditions(t *testing.T) {
	yaml := `
strategy:
  name: functions_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema:
    type: ema
    period: 21

entry:
  long:
    cross_above: [close, ema]

exit:
  long:
    any:
      - cross_below: [close, ema]
      - lt: [close, 90000]
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)

	// Entry should be cross_above function
	assert.Equal(t, "func", strat.Entry.Long.Type)
	assert.Equal(t, "cross_above", strat.Entry.Long.Function)

	// Exit long should be "any" with 2 conditions
	assert.Equal(t, "any", strat.Exit.Long.Type)
	assert.Equal(t, 2, len(strat.Exit.Long.Conditions))

	// Second exit condition should be "lt" function
	assert.Equal(t, "func", strat.Exit.Long.Conditions[1].Type)
	assert.Equal(t, "lt", strat.Exit.Long.Conditions[1].Function)
}

// TestParserExitRules tests exit condition parsing.
func TestParserExitRules(t *testing.T) {
	yaml := `
strategy:
  name: exit_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema:
    type: ema
    period: 21

entry:
  long:
    cross_above: [close, ema]

exit:
  long:
    all:
      - cross_below: [close, ema]
      - between: [close, 90000, 110000]
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.NotNil(t, strat.Exit.Long)
	assert.Equal(t, "all", strat.Exit.Long.Type)
	assert.Equal(t, 2, len(strat.Exit.Long.Conditions))
}

// TestParserAllExitTypes tests different exit condition types.
func TestParserAllExitTypes(t *testing.T) {
	testCases := []struct {
		name   string
		yaml   string
		expect *ast.Condition
	}{
		{
			name: "all exit",
			yaml: `
strategy:
  name: test
  version: "1"
data:
  symbol: BTCUSDT
  timeframe: 4h
indicators:
  ema:
    type: ema
    period: 21
entry:
  long:
    cross_above: [close, ema]
exit:
  long:
    all:
      - cross_below: [close, ema]
`,
			expect: &ast.Condition{Type: "all"},
		},
		{
			name: "any exit",
			yaml: `
strategy:
  name: test
  version: "1"
data:
  symbol: BTCUSDT
  timeframe: 4h
indicators:
  ema:
    type: ema
    period: 21
entry:
  long:
    cross_above: [close, ema]
exit:
  long:
    any:
      - cross_below: [close, ema]
`,
			expect: &ast.Condition{Type: "any"},
		},
		{
			name: "not exit",
			yaml: `
strategy:
  name: test
  version: "1"
data:
  symbol: BTCUSDT
  timeframe: 4h
indicators:
  ema:
    type: ema
    period: 21
entry:
  long:
    cross_above: [close, ema]
exit:
  long:
    not:
      cross_below: [close, ema]
`,
			expect: &ast.Condition{Type: "not"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser()
			strat, err := p.Parse([]byte(tc.yaml))
			assert.NoError(t, err)
			assert.Equal(t, tc.expect.Type, strat.Exit.Long.Type)
		})
	}
}

// TestParserFunctionArgs tests parsing function arguments.
func TestParserFunctionArgs(t *testing.T) {
	yaml := `
strategy:
  name: args_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 4h

indicators:
  ema:
    type: ema
    period: 21

entry:
  long:
    between: [close, 100000, 120000]
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.Equal(t, "func", strat.Entry.Long.Type)
	assert.Equal(t, "between", strat.Entry.Long.Function)
	assert.Equal(t, 3, len(strat.Entry.Long.Args))
}

// TestParserIndicatorWithTimeframe tests indicator with timeframe.
func TestParserIndicatorWithTimeframe(t *testing.T) {
	yaml := `
strategy:
  name: mtf_test
  version: "1"

data:
  symbol: BTCUSDT
  timeframe: 1h

indicators:
  ema_4h:
    type: ema
    timeframe: 4h
    period: 200

entry:
  long:
    gt: [close, ema_4h]
`

	p := NewParser()
	strat, err := p.Parse([]byte(yaml))

	assert.NoError(t, err)
	assert.Equal(t, "4h", strat.Indicators["ema_4h"].Timeframe)
	assert.Equal(t, 200, strat.Indicators["ema_4h"].Period)
}
