package optimization

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZulferDev/smallbt_go/internal/strategy/ast"
	"github.com/ZulferDev/smallbt_go/internal/strategy/parser"
	"gopkg.in/yaml.v3"
)

// YAMLModifier implements StrategyModifier for YAML strategy files.
type YAMLModifier struct {
	basePath string
}

// NewYAMLModifier creates a new YAML strategy modifier.
func NewYAMLModifier() *YAMLModifier {
	return &YAMLModifier{}
}

// ModifyStrategy modifies a strategy AST with new parameter values.
func (m *YAMLModifier) ModifyStrategy(strategy *ast.Strategy, parameterSet ParameterSet) (*ast.Strategy, error) {
	// Create a deep copy of the strategy
	modified := m.copyStrategy(strategy)

	// Apply parameter values
	for paramName, paramValue := range parameterSet.Values {
		// Parse the path (e.g., "indicators.ema_fast.period")
		pathParts := strings.Split(paramName, ".")

		if len(pathParts) < 3 {
			return nil, fmt.Errorf("invalid parameter path format: %s (expected format: 'indicators.name.field')", paramName)
		}

		// Path structure: [category, indicatorName, field]
		category := pathParts[0]
		indicatorName := pathParts[1]
		field := pathParts[2]

		switch category {
		case "indicators":
			if modified.Indicators == nil {
				return nil, fmt.Errorf("no indicators defined in strategy")
			}

			indicator, exists := modified.Indicators[indicatorName]
			if !exists {
				return nil, fmt.Errorf("indicator %s not found in strategy", indicatorName)
			}

			// Set the field value
			switch field {
			case "period":
				indicator.Period = int(paramValue)
			case "source":
				// source is a string, but we might allow numeric references
				// For now, just convert to string
				return nil, fmt.Errorf("source field modification not yet supported")
			default:
				// For custom parameters in Params map
				if indicator.Params == nil {
					indicator.Params = make(map[string]interface{})
				}
				indicator.Params[field] = paramValue
			}

			modified.Indicators[indicatorName] = indicator

		case "risk":
			switch indicatorName {
			case "position_size":
				switch field {
				case "value":
					if modified.Risk.PositionSize.Value == 0 {
						modified.Risk.PositionSize.Value = paramValue
					} else {
						modified.Risk.PositionSize.Value = paramValue
					}
				}
			case "stop_loss":
				if modified.Risk.StopLoss != nil {
					switch field {
					case "value", "period", "multiplier":
						// These would go into stop loss config
						return nil, fmt.Errorf("stop loss modification not yet implemented")
					}
				}
			}

		default:
			return nil, fmt.Errorf("unsupported parameter category: %s", category)
		}
	}

	return modified, nil
}

// ModifyStrategyFile loads, modifies, and saves a YAML strategy file.
func (m *YAMLModifier) ModifyStrategyFile(yamlPath string, parameterSet ParameterSet) (string, error) {
	// Load the YAML file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return "", fmt.Errorf("read strategy file: %w", err)
	}

	// Parse into map for easier modification
	var strategyMap map[string]interface{}
	if err := yaml.Unmarshal(data, &strategyMap); err != nil {
		return "", fmt.Errorf("unmarshal YAML: %w", err)
	}

	// Apply parameter modifications
	modifiedMap := m.applyParameterModifications(strategyMap, parameterSet)

	// Convert back to YAML
	modifiedYAML, err := yaml.Marshal(modifiedMap)
	if err != nil {
		return "", fmt.Errorf("marshal modified YAML: %w", err)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "optimization_*.yaml")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.Write(modifiedYAML); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	return tempFile.Name(), nil
}

// applyParameterModifications applies parameter values to YAML map.
func (m *YAMLModifier) applyParameterModifications(strategyMap map[string]interface{}, parameterSet ParameterSet) map[string]interface{} {
	// Deep copy the map
	result := make(map[string]interface{})
	for k, v := range strategyMap {
		result[k] = v
	}

	// Apply each parameter
	for paramName, paramValue := range parameterSet.Values {
		// Parse path
		pathParts := strings.Split(paramName, ".")

		if len(pathParts) < 3 {
			continue // skip invalid paths
		}

		category := pathParts[0]
		indicatorName := pathParts[1]
		field := pathParts[2]

		// Navigate through the map
		current := result

		// Find category (e.g., "indicators")
		categoryMap, ok := current[category].(map[string]interface{})
		if !ok {
			continue
		}

		// Find indicator
		indicatorMap, ok := categoryMap[indicatorName].(map[string]interface{})
		if !ok {
			continue
		}

		// Set field value
		// For integer fields, convert to int if value is integer
		if field == "period" {
			indicatorMap[field] = int(paramValue)
		} else {
			indicatorMap[field] = paramValue
		}
	}

	return result
}

// copyStrategy creates a deep copy of a strategy.
func (m *YAMLModifier) copyStrategy(strategy *ast.Strategy) *ast.Strategy {
	// Basic fields
	copy := &ast.Strategy{
		Name:        strategy.Name,
		Version:     strategy.Version,
		Description: strategy.Description,
		Data: ast.DataConfig{
			Symbol:    strategy.Data.Symbol,
			Timeframe: strategy.Data.Timeframe,
		},
	}

	// Copy indicators
	if strategy.Indicators != nil {
		copy.Indicators = make(map[string]ast.IndicatorDef)
		for k, v := range strategy.Indicators {
			ind := ast.IndicatorDef{
				Name:      v.Name,
				Type:      v.Type,
				Source:    v.Source,
				Period:    v.Period,
				Timeframe: v.Timeframe,
				Left:      v.Left,
				Right:     v.Right,
				Op:        v.Op,
			}
			// Copy params if exists
			if v.Params != nil {
				ind.Params = make(map[string]interface{})
				for pk, pv := range v.Params {
					ind.Params[pk] = pv
				}
			}
			copy.Indicators[k] = ind
		}
	}

	// Copy entry rules (simplified)
	if strategy.Entry.Long != nil {
		// For optimization purposes, we just need structure, not deep copy
		copy.Entry = ast.EntryRules{
			Long:  &ast.Condition{Type: "placeholder"},
			Short: nil,
		}
	}

	// Copy risk config (simplified)
	copy.Risk = ast.RiskConfig{
		PositionSize: ast.PositionSizeConfig{
			Type:  strategy.Risk.PositionSize.Type,
			Value: strategy.Risk.PositionSize.Value,
		},
	}

	// Copy exit rules if exists
	if strategy.Exit.Long != nil {
		copy.Exit = ast.ExitRules{
			Long:  &ast.Condition{Type: "placeholder"},
			Short: nil,
		}
	}

	return copy
}

// ValidateParameterPath validates if a parameter path exists in the strategy.
func (m *YAMLModifier) ValidateParameterPath(yamlPath string, parameterPath string) error {
	// Load and parse strategy
	parser := parser.NewParser()
	strategy, err := parser.ParseFile(yamlPath)
	if err != nil {
		return fmt.Errorf("parse strategy: %w", err)
	}

	// Parse parameter path
	pathParts := strings.Split(parameterPath, ".")
	if len(pathParts) < 3 {
		return fmt.Errorf("invalid parameter path format: %s (expected: category.indicator.field)", parameterPath)
	}

	category := pathParts[0]
	indicatorName := pathParts[1]
	field := pathParts[2]

	switch category {
	case "indicators":
		// Check if indicator exists
		indicator, exists := strategy.Indicators[indicatorName]
		if !exists {
			return fmt.Errorf("indicator %s not found in strategy", indicatorName)
		}

		// Check if field is valid
		switch field {
		case "period":
			// Valid for all indicators
			return nil
		case "source":
			// Valid for some indicators
			if indicator.Source != "" {
				return nil
			}
			return fmt.Errorf("indicator %s does not have a source field", indicatorName)
		default:
			// Check if field exists in params
			if indicator.Params != nil {
				if _, exists := indicator.Params[field]; exists {
					return nil
				}
			}
			return fmt.Errorf("field %s not found in indicator %s", field, indicatorName)
		}

	case "risk":
		switch indicatorName {
		case "position_size":
			if field == "value" {
				return nil
			}
		case "stop_loss":
			if strategy.Risk.StopLoss != nil {
				if field == "value" || field == "multiplier" || field == "period" {
					return nil
				}
			}
		}
		return fmt.Errorf("risk field %s.%s not found", indicatorName, field)

	default:
		return fmt.Errorf("unsupported category: %s", category)
	}
}

// GetTempFileName generates a temporary file name for modified strategy.
func (m *YAMLModifier) GetTempFileName(baseName string, paramSet ParameterSet) string {
	// Create hash-based name
	hash := paramSet.Hash
	if hash == "" {
		hash = "default"
	}

	ext := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, ext)

	return fmt.Sprintf("%s_%s%s", name, hash[:8], ext)
}
