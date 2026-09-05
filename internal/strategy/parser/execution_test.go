package parser

import (
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/strategy/ast"
)

func TestParseExecutionConfig(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantErr  bool
		validate func(*testing.T, ast.ExecutionConfig)
	}{
		{
			name: "default execution config",
			yaml: `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
`,
			wantErr: false,
			validate: func(t *testing.T, cfg ast.ExecutionConfig) {
				if cfg.EntryOrderType != "market" {
					t.Errorf("expected default entry_order_type=market, got %s", cfg.EntryOrderType)
				}
				if cfg.ExitOrderType != "market" {
					t.Errorf("expected default exit_order_type=market, got %s", cfg.ExitOrderType)
				}
				if cfg.SlippageType != "percentage" {
					t.Errorf("expected default slippage_type=percentage, got %s", cfg.SlippageType)
				}
				if cfg.SlippageValue != 0.0005 {
					t.Errorf("expected default slippage_value=0.0005, got %f", cfg.SlippageValue)
				}
				if cfg.FeeMaker != 0.0002 {
					t.Errorf("expected default fee_maker=0.0002, got %f", cfg.FeeMaker)
				}
				if cfg.FeeTaker != 0.0005 {
					t.Errorf("expected default fee_taker=0.0005, got %f", cfg.FeeTaker)
				}
			},
		},
		{
			name: "custom limit orders",
			yaml: `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
execution:
  entry_order_type: limit
  exit_order_type: limit
`,
			wantErr: false,
			validate: func(t *testing.T, cfg ast.ExecutionConfig) {
				if cfg.EntryOrderType != "limit" {
					t.Errorf("expected entry_order_type=limit, got %s", cfg.EntryOrderType)
				}
				if cfg.ExitOrderType != "limit" {
					t.Errorf("expected exit_order_type=limit, got %s", cfg.ExitOrderType)
				}
			},
		},
		{
			name: "custom fees and slippage",
			yaml: `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
execution:
  slippage_type: percentage
  slippage_value: 0.001
  fee_maker: 0.0001
  fee_taker: 0.0003
`,
			wantErr: false,
			validate: func(t *testing.T, cfg ast.ExecutionConfig) {
				if cfg.SlippageValue != 0.001 {
					t.Errorf("expected slippage_value=0.001, got %f", cfg.SlippageValue)
				}
				if cfg.FeeMaker != 0.0001 {
					t.Errorf("expected fee_maker=0.0001, got %f", cfg.FeeMaker)
				}
				if cfg.FeeTaker != 0.0003 {
					t.Errorf("expected fee_taker=0.0003, got %f", cfg.FeeTaker)
				}
			},
		},
		{
			name: "stop orders",
			yaml: `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
execution:
  entry_order_type: stop
  exit_order_type: stop_limit
`,
			wantErr: false,
			validate: func(t *testing.T, cfg ast.ExecutionConfig) {
				if cfg.EntryOrderType != "stop" {
					t.Errorf("expected entry_order_type=stop, got %s", cfg.EntryOrderType)
				}
				if cfg.ExitOrderType != "stop_limit" {
					t.Errorf("expected exit_order_type=stop_limit, got %s", cfg.ExitOrderType)
				}
			},
		},
		{
			name: "intrabar policy",
			yaml: `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
execution:
  intrabar_policy: conservative
`,
			wantErr: false,
			validate: func(t *testing.T, cfg ast.ExecutionConfig) {
				if cfg.IntrabarPolicy != "conservative" {
					t.Errorf("expected intrabar_policy=conservative, got %s", cfg.IntrabarPolicy)
				}
			},
		},
		{
			name: "spread configuration",
			yaml: `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
execution:
  spread: 0.0002
`,
			wantErr: false,
			validate: func(t *testing.T, cfg ast.ExecutionConfig) {
				if cfg.Spread != 0.0002 {
					t.Errorf("expected spread=0.0002, got %f", cfg.Spread)
				}
			},
		},
		{
			name: "full custom execution config",
			yaml: `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
execution:
  entry_order_type: limit
  exit_order_type: market
  slippage_type: percentage
  slippage_value: 0.002
  fee_maker: 0.00015
  fee_taker: 0.00045
  spread: 0.0003
  intrabar_policy: conservative
`,
			wantErr: false,
			validate: func(t *testing.T, cfg ast.ExecutionConfig) {
				if cfg.EntryOrderType != "limit" {
					t.Errorf("expected entry_order_type=limit, got %s", cfg.EntryOrderType)
				}
				if cfg.ExitOrderType != "market" {
					t.Errorf("expected exit_order_type=market, got %s", cfg.ExitOrderType)
				}
				if cfg.SlippageType != "percentage" {
					t.Errorf("expected slippage_type=percentage, got %s", cfg.SlippageType)
				}
				if cfg.SlippageValue != 0.002 {
					t.Errorf("expected slippage_value=0.002, got %f", cfg.SlippageValue)
				}
				if cfg.FeeMaker != 0.00015 {
					t.Errorf("expected fee_maker=0.00015, got %f", cfg.FeeMaker)
				}
				if cfg.FeeTaker != 0.00045 {
					t.Errorf("expected fee_taker=0.00045, got %f", cfg.FeeTaker)
				}
				if cfg.Spread != 0.0003 {
					t.Errorf("expected spread=0.0003, got %f", cfg.Spread)
				}
				if cfg.IntrabarPolicy != "conservative" {
					t.Errorf("expected intrabar_policy=conservative, got %s", cfg.IntrabarPolicy)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			strategyAST, err := p.Parse([]byte(tt.yaml))

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, strategyAST.Execution)
			}
		})
	}
}

func TestExecutionConfigOrderTypes(t *testing.T) {
	validOrderTypes := []string{"market", "limit", "stop", "stop_limit"}

	for _, orderType := range validOrderTypes {
		t.Run("entry_"+orderType, func(t *testing.T) {
			yaml := `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
execution:
  entry_order_type: ` + orderType

			p := NewParser()
			strategyAST, err := p.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strategyAST.Execution.EntryOrderType != orderType {
				t.Errorf("expected entry_order_type=%s, got %s", orderType, strategyAST.Execution.EntryOrderType)
			}
		})

		t.Run("exit_"+orderType, func(t *testing.T) {
			yaml := `
strategy:
  name: test
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
execution:
  exit_order_type: ` + orderType

			p := NewParser()
			strategyAST, err := p.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strategyAST.Execution.ExitOrderType != orderType {
				t.Errorf("expected exit_order_type=%s, got %s", orderType, strategyAST.Execution.ExitOrderType)
			}
		})
	}
}

func TestExecutionConfigDefaults(t *testing.T) {
	yaml := `
strategy:
  name: minimal
data:
  symbol: BTCUSDT
  timeframe: 1h
indicators:
  ema9:
    type: ema
    period: 9
entry:
  long:
    - gt: [close, ema9]
`

	p := NewParser()
	strategyAST, err := p.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all defaults
	defaults := map[string]interface{}{
		"entry_order_type": "market",
		"exit_order_type":  "market",
		"slippage_type":    "percentage",
		"slippage_value":   0.0005,
		"fee_maker":        0.0002,
		"fee_taker":        0.0005,
		"spread":           0.0001,
		"intrabar_policy":  "conservative",
	}

	if strategyAST.Execution.EntryOrderType != defaults["entry_order_type"] {
		t.Errorf("entry_order_type: expected %v, got %v", defaults["entry_order_type"], strategyAST.Execution.EntryOrderType)
	}
	if strategyAST.Execution.ExitOrderType != defaults["exit_order_type"] {
		t.Errorf("exit_order_type: expected %v, got %v", defaults["exit_order_type"], strategyAST.Execution.ExitOrderType)
	}
	if strategyAST.Execution.SlippageType != defaults["slippage_type"] {
		t.Errorf("slippage_type: expected %v, got %v", defaults["slippage_type"], strategyAST.Execution.SlippageType)
	}
	if strategyAST.Execution.SlippageValue != defaults["slippage_value"] {
		t.Errorf("slippage_value: expected %v, got %v", defaults["slippage_value"], strategyAST.Execution.SlippageValue)
	}
	if strategyAST.Execution.FeeMaker != defaults["fee_maker"] {
		t.Errorf("fee_maker: expected %v, got %v", defaults["fee_maker"], strategyAST.Execution.FeeMaker)
	}
	if strategyAST.Execution.FeeTaker != defaults["fee_taker"] {
		t.Errorf("fee_taker: expected %v, got %v", defaults["fee_taker"], strategyAST.Execution.FeeTaker)
	}
	if strategyAST.Execution.Spread != defaults["spread"] {
		t.Errorf("spread: expected %v, got %v", defaults["spread"], strategyAST.Execution.Spread)
	}
	if strategyAST.Execution.IntrabarPolicy != defaults["intrabar_policy"] {
		t.Errorf("intrabar_policy: expected %v, got %v", defaults["intrabar_policy"], strategyAST.Execution.IntrabarPolicy)
	}
}
