package parser

// YAMLStrategy represents the raw YAML structure before AST conversion.
type YAMLStrategy struct {
	Strategy   YAMLStrategyDef         `yaml:"strategy"`
	Data       YAMLDataConfig          `yaml:"data"`
	Indicators map[string]interface{}  `yaml:"indicators"`
	Entry      YAMLEntryRules          `yaml:"entry"`
	Exit       YAMLExitRules           `yaml:"exit"`
	Risk       YAMLRiskConfig          `yaml:"risk"`
	Execution  YAMLExecutionConfig     `yaml:"execution"`
	State      map[string]YAMLStateDef `yaml:"state"`
}

// YAMLStrategyDef defines strategy metadata.
type YAMLStrategyDef struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// YAMLDataConfig defines data source.
type YAMLDataConfig struct {
	Symbol    string `yaml:"symbol"`
	Timeframe string `yaml:"timeframe"`
}

// YAMLEntryRules defines entry conditions.
type YAMLEntryRules struct {
	Long  interface{} `yaml:"long"`
	Short interface{} `yaml:"short"`
}

// YAMLExitRules defines exit conditions.
type YAMLExitRules struct {
	Long  interface{} `yaml:"long"`
	Short interface{} `yaml:"short"`
}

// YAMLCondition represents a condition (can be nested).
type YAMLCondition struct {
	All  []interface{}
	Any  []interface{}
	Not  interface{}
	Expr interface{}
	Func string
	Args []interface{}
	// For direct operators like cross_above, gt, etc.
	// The key is the function name, value is the args
}

// YAMLRiskConfig defines risk management.
type YAMLRiskConfig struct {
	PositionSize     map[string]interface{} `yaml:"position_size"`
	StopLoss         map[string]interface{} `yaml:"stop_loss"`
	TakeProfit       map[string]interface{} `yaml:"take_profit"`
	TrailingStop     map[string]interface{} `yaml:"trailing_stop"`
	MaxPositions     int                    `yaml:"max_positions"`
	MaxPortfolioRisk float64                `yaml:"max_portfolio_risk"`
	MaxDailyLoss     float64                `yaml:"max_daily_loss"`
	MaxDrawdown      float64                `yaml:"max_drawdown"`
}

// YAMLStateDef defines a state variable.
type YAMLStateDef struct {
	Type    string      `yaml:"type"`
	Default interface{} `yaml:"default"`
}

// YAMLExecutionConfig defines execution behavior.
type YAMLExecutionConfig struct {
	EntryOrderType  string  `yaml:"entry_order_type"`
	ExitOrderType   string  `yaml:"exit_order_type"`
	SlippageType    string  `yaml:"slippage_type"`
	SlippageValue   float64 `yaml:"slippage_value"`
	FeeMaker        float64 `yaml:"fee_maker"`
	FeeTaker        float64 `yaml:"fee_taker"`
	Spread          float64 `yaml:"spread"`
	IntrabarPolicy  string  `yaml:"intrabar_policy"`
}
