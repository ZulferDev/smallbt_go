package parser

import (
	"fmt"
	"os"

	"github.com/ZulferDev/smallbt_go/internal/strategy/ast"
	"gopkg.in/yaml.v3"
)

// Parser parses YAML strategy files into AST.
type Parser struct {
	// Future: schema validation, version checking
}

// NewParser creates a new strategy parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a YAML file into a Strategy AST.
func (p *Parser) ParseFile(path string) (*ast.Strategy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read strategy file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses YAML bytes into a Strategy AST.
func (p *Parser) Parse(data []byte) (*ast.Strategy, error) {
	var yamlStrat YAMLStrategy
	if err := yaml.Unmarshal(data, &yamlStrat); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	strategy := &ast.Strategy{
		Name:        yamlStrat.Strategy.Name,
		Version:     yamlStrat.Strategy.Version,
		Description: yamlStrat.Strategy.Description,
		Data: ast.DataConfig{
			Symbol:    yamlStrat.Data.Symbol,
			Timeframe: yamlStrat.Data.Timeframe,
		},
		Indicators: make(map[string]ast.IndicatorDef),
		State:      make(map[string]ast.StateDef),
	}

	// Parse indicators
	if err := p.parseIndicators(yamlStrat.Indicators, strategy); err != nil {
		return nil, fmt.Errorf("parse indicators: %w", err)
	}

	// Parse entry rules
	if err := p.parseEntryRules(yamlStrat.Entry, strategy); err != nil {
		return nil, fmt.Errorf("parse entry rules: %w", err)
	}

	// Parse exit rules
	if err := p.parseExitRules(yamlStrat.Exit, strategy); err != nil {
		return nil, fmt.Errorf("parse exit rules: %w", err)
	}

	// Parse risk config
	if err := p.parseRiskConfig(yamlStrat.Risk, strategy); err != nil {
		return nil, fmt.Errorf("parse risk config: %w", err)
	}

	// Parse state definitions
	for name, stateDef := range yamlStrat.State {
		strategy.State[name] = ast.StateDef{
			Name:    name,
			Type:    stateDef.Type,
			Default: stateDef.Default,
		}
	}

	// Parse execution config
	if err := p.parseExecutionConfig(yamlStrat.Execution, strategy); err != nil {
		return nil, fmt.Errorf("parse execution config: %w", err)
	}

	return strategy, nil
}

// parseIndicators parses indicator definitions.
func (p *Parser) parseIndicators(indicators map[string]interface{}, strategy *ast.Strategy) error {
	for name, def := range indicators {
		indicatorDef, err := p.parseIndicator(name, def)
		if err != nil {
			return fmt.Errorf("indicator %q: %w", name, err)
		}
		strategy.Indicators[name] = indicatorDef
	}
	return nil
}

// parseIndicator parses a single indicator definition.
func (p *Parser) parseIndicator(name string, def interface{}) (ast.IndicatorDef, error) {
	indicator := ast.IndicatorDef{
		Name:   name,
		Params: make(map[string]interface{}),
	}

	// def can be a map or a simple type
	switch d := def.(type) {
	case map[string]interface{}:
		// Extract known fields
		if v, ok := d["type"]; ok {
			indicator.Type = fmt.Sprint(v)
		}
		if v, ok := d["source"]; ok {
			indicator.Source = fmt.Sprint(v)
		}
		if v, ok := d["period"]; ok {
			switch period := v.(type) {
			case int:
				indicator.Period = period
			case float64:
				indicator.Period = int(period)
			default:
				indicator.Period = 0
			}
		}
		if v, ok := d["timeframe"]; ok {
			indicator.Timeframe = fmt.Sprint(v)
		}
		if v, ok := d["left"]; ok {
			indicator.Left = fmt.Sprint(v)
		}
		if v, ok := d["right"]; ok {
			indicator.Right = fmt.Sprint(v)
		}
		if v, ok := d["op"]; ok {
			indicator.Op = fmt.Sprint(v)
		}

		// Store all params for custom indicators
		for key, val := range d {
			indicator.Params[key] = val
		}

	default:
		return indicator, fmt.Errorf("invalid indicator definition type: %T", def)
	}

	return indicator, nil
}

// parseEntryRules parses entry conditions.
func (p *Parser) parseEntryRules(entry YAMLEntryRules, strategy *ast.Strategy) error {
	if entry.Long != nil {
		cond, err := p.parseCondition(entry.Long)
		if err != nil {
			return fmt.Errorf("long entry: %w", err)
		}
		strategy.Entry.Long = cond
	}

	if entry.Short != nil {
		cond, err := p.parseCondition(entry.Short)
		if err != nil {
			return fmt.Errorf("short entry: %w", err)
		}
		strategy.Entry.Short = cond
	}

	return nil
}

// parseExitRules parses exit conditions.
func (p *Parser) parseExitRules(exit YAMLExitRules, strategy *ast.Strategy) error {
	if exit.Long != nil {
		cond, err := p.parseCondition(exit.Long)
		if err != nil {
			return fmt.Errorf("long exit: %w", err)
		}
		strategy.Exit.Long = cond
	}

	if exit.Short != nil {
		cond, err := p.parseCondition(exit.Short)
		if err != nil {
			return fmt.Errorf("short exit: %w", err)
		}
		strategy.Exit.Short = cond
	}

	return nil
}

// parseCondition parses a condition (can be nested with all/any/not).
func (p *Parser) parseCondition(def interface{}) (*ast.Condition, error) {
	switch d := def.(type) {
	case map[string]interface{}:
		cond := &ast.Condition{}

		// Check for composite conditions
		if all, ok := d["all"]; ok {
			cond.Type = "all"
			conditions, err := p.parseConditionList(all)
			if err != nil {
				return nil, fmt.Errorf("parse 'all': %w", err)
			}
			cond.Conditions = conditions
			return cond, nil
		}

		if any, ok := d["any"]; ok {
			cond.Type = "any"
			conditions, err := p.parseConditionList(any)
			if err != nil {
				return nil, fmt.Errorf("parse 'any': %w", err)
			}
			cond.Conditions = conditions
			return cond, nil
		}

		if not, ok := d["not"]; ok {
			cond.Type = "not"
			inner, err := p.parseCondition(not)
			if err != nil {
				return nil, fmt.Errorf("parse 'not': %w", err)
			}
			cond.Conditions = []*ast.Condition{inner}
			return cond, nil
		}

		if or, ok := d["or"]; ok {
			cond.Type = "any"
			conditions, err := p.parseConditionList(or)
			if err != nil {
				return nil, fmt.Errorf("parse 'or': %w", err)
			}
			cond.Conditions = conditions
			return cond, nil
		}

		if and, ok := d["and"]; ok {
			cond.Type = "all"
			conditions, err := p.parseConditionList(and)
			if err != nil {
				return nil, fmt.Errorf("parse 'and': %w", err)
			}
			cond.Conditions = conditions
			return cond, nil
		}

		// Check for function-based conditions
		// The key is the function name (e.g., "cross_above", "gt")
		for key, args := range d {
			switch key {
			case "cross_above", "cross_below", "rising", "falling", "between",
				"gt", "lt", "ge", "le", "eq", "ne":
				cond.Type = "func"
				cond.Function = key
				argsList, err := p.parseArgs(args)
				if err != nil {
					return nil, fmt.Errorf("parse %s args: %w", key, err)
				}
				cond.Args = argsList
				return cond, nil
			}
		}

		return nil, fmt.Errorf("unknown condition type: %v", d)

	case []interface{}:
		// Inline array is treated as 'all' by default
		cond := &ast.Condition{Type: "all"}
		conditions, err := p.parseConditionList(d)
		if err != nil {
			return nil, err
		}
		cond.Conditions = conditions
		return cond, nil

	case string:
		// Simple string condition (e.g., just an indicator name)
		return &ast.Condition{
			Type:     "expr",
			Function: "eq",
			Args:     []interface{}{d, true},
		}, nil

	default:
		return nil, fmt.Errorf("invalid condition type: %T", def)
	}
}

// parseConditionList parses a list of conditions.
func (p *Parser) parseConditionList(def interface{}) ([]*ast.Condition, error) {
	list, ok := def.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected list, got %T", def)
	}

	conditions := make([]*ast.Condition, 0, len(list))
	for i, item := range list {
		cond, err := p.parseCondition(item)
		if err != nil {
			return nil, fmt.Errorf("condition %d: %w", i, err)
		}
		conditions = append(conditions, cond)
	}

	return conditions, nil
}

// parseArgs parses function arguments.
func (p *Parser) parseArgs(args interface{}) ([]interface{}, error) {
	switch a := args.(type) {
	case []interface{}:
		return a, nil
	case string, float64, int, bool:
		return []interface{}{a}, nil
	default:
		return nil, fmt.Errorf("invalid args type: %T", args)
	}
}

// parseExecutionConfig parses execution configuration.
func (p *Parser) parseExecutionConfig(config YAMLExecutionConfig, strategy *ast.Strategy) error {
	strategy.Execution.EntryOrderType = "market"
	if config.EntryOrderType != "" {
		strategy.Execution.EntryOrderType = config.EntryOrderType
	}

	strategy.Execution.ExitOrderType = "market"
	if config.ExitOrderType != "" {
		strategy.Execution.ExitOrderType = config.ExitOrderType
	}

	strategy.Execution.SlippageType = "percentage"
	if config.SlippageType != "" {
		strategy.Execution.SlippageType = config.SlippageType
	}
	strategy.Execution.SlippageValue = 0.0005 // 0.05% default
	if config.SlippageValue != 0 {
		strategy.Execution.SlippageValue = config.SlippageValue
	}

	strategy.Execution.FeeMaker = 0.0002 // 0.02% default
	if config.FeeMaker != 0 {
		strategy.Execution.FeeMaker = config.FeeMaker
	}

	strategy.Execution.FeeTaker = 0.0005 // 0.05% default
	if config.FeeTaker != 0 {
		strategy.Execution.FeeTaker = config.FeeTaker
	}

	strategy.Execution.Spread = 0.0001 // 0.01% default
	if config.Spread != 0 {
		strategy.Execution.Spread = config.Spread
	}

	strategy.Execution.IntrabarPolicy = "conservative"
	if config.IntrabarPolicy != "" {
		strategy.Execution.IntrabarPolicy = config.IntrabarPolicy
	}

	return nil
}

// parseRiskConfig parses risk management configuration.
func (p *Parser) parseRiskConfig(risk YAMLRiskConfig, strategy *ast.Strategy) error {
	// Parse position sizing
	if risk.PositionSize != nil {
		ps := ast.PositionSizeConfig{}
		if v, ok := risk.PositionSize["type"]; ok {
			ps.Type = fmt.Sprint(v)
		}
		if v, ok := risk.PositionSize["value"]; ok {
			switch val := v.(type) {
			case float64:
				ps.Value = val
				// For risk_percent, also set RiskPercent field
				if ps.Type == "risk_percent" {
					ps.RiskPercent = val
				}
			case int:
				ps.Value = float64(val)
				if ps.Type == "risk_percent" {
					ps.RiskPercent = float64(val)
				}
			}
		}
		strategy.Risk.PositionSize = ps
	}

	// Parse stop loss
	if risk.StopLoss != nil {
		sl := &ast.StopLossConfig{}
		if v, ok := risk.StopLoss["type"]; ok {
			sl.Type = fmt.Sprint(v)
		}
		if v, ok := risk.StopLoss["price"]; ok {
			if val, ok := v.(float64); ok {
				sl.Price = val
			}
		}
		if v, ok := risk.StopLoss["percentage"]; ok {
			if val, ok := v.(float64); ok {
				sl.Percentage = val
			}
		}
		// Support "value" as alias for "percentage"
		if v, ok := risk.StopLoss["value"]; ok {
			if val, ok := v.(float64); ok {
				sl.Percentage = val * 100 // Convert 0.02 to 2.0
			}
		}
		if v, ok := risk.StopLoss["indicator"]; ok {
			sl.Indicator = fmt.Sprint(v)
		}
		if v, ok := risk.StopLoss["multiplier"]; ok {
			if val, ok := v.(float64); ok {
				sl.Multiplier = val
			}
		}
		strategy.Risk.StopLoss = sl
	}

	// Parse take profit
	if risk.TakeProfit != nil {
		tp := &ast.TakeProfitConfig{}
		if v, ok := risk.TakeProfit["type"]; ok {
			tp.Type = fmt.Sprint(v)
		}
		if v, ok := risk.TakeProfit["ratio"]; ok {
			if val, ok := v.(float64); ok {
				tp.Ratio = val
			}
		}
		strategy.Risk.TakeProfit = tp
	}

	// Portfolio-level risk
	strategy.Risk.MaxPositions = risk.MaxPositions
	strategy.Risk.MaxPortfolioRisk = risk.MaxPortfolioRisk
	strategy.Risk.MaxDailyLoss = risk.MaxDailyLoss
	strategy.Risk.MaxDrawdown = risk.MaxDrawdown

	return nil
}
