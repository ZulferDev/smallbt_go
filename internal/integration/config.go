package integration

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/data/transform"
)

// TransformConfig defines transform configuration for strategy YAML.
type TransformConfig struct {
	// Enabled determines if transforms are applied
	Enabled bool `yaml:"enabled"`

	// Transforms lists the transform pipeline
	Transforms []TransformSpec `yaml:"transforms"`

	// BatchSize for batch processing (default: 100)
	BatchSize int `yaml:"batch_size,omitempty"`
}

// TransformSpec specifies a single transform in the pipeline.
type TransformSpec struct {
	// Type of transform (scale, normalize, clip_outliers, etc.)
	Type string `yaml:"type"`

	// Field to transform (close, open, high, low, volume)
	Field string `yaml:"field,omitempty"`

	// Parameters specific to transform type
	Params map[string]interface{} `yaml:"params,omitempty"`
}

// BuildTransformChain constructs a transform chain from config.
func (tc *TransformConfig) BuildTransformChain() (*transform.TransformChain, error) {
	if !tc.Enabled || len(tc.Transforms) == 0 {
		return nil, nil
	}

	transforms := make([]transform.Transform, 0, len(tc.Transforms))

	for i, spec := range tc.Transforms {
		t, err := buildTransform(spec)
		if err != nil {
			return nil, fmt.Errorf("transform %d (%s): %w", i, spec.Type, err)
		}
		transforms = append(transforms, t)
	}

	return transform.NewTransformChain(transforms...), nil
}

// buildTransform creates a transform from spec.
func buildTransform(spec TransformSpec) (transform.Transform, error) {
	switch spec.Type {
	case "scale":
		factor, err := getFloatParam(spec.Params, "factor", 1.0)
		if err != nil {
			return nil, err
		}
		return transform.NewScaleTransform(factor, spec.Field), nil

	case "normalize":
		return transform.NewNormalizeTransform(spec.Field), nil

	case "log_returns":
		return transform.NewLogReturnsTransform(spec.Field), nil

	case "percentage_change":
		return transform.NewPercentageChangeTransform(spec.Field), nil

	case "smooth":
		period, err := getIntParam(spec.Params, "period", 5)
		if err != nil {
			return nil, err
		}
		return transform.NewMovingAverageSmoothTransform(period, spec.Field), nil

	case "zscore":
		window, err := getIntParam(spec.Params, "window", 20)
		if err != nil {
			return nil, err
		}
		return &transform.ZScoreTransform{
			Field:  spec.Field,
			Window: window,
		}, nil

	case "difference":
		order, err := getIntParam(spec.Params, "order", 1)
		if err != nil {
			return nil, err
		}
		return &transform.DifferencingTransform{
			Field: spec.Field,
			Order: order,
		}, nil

	case "ema_smooth":
		period, err := getIntParam(spec.Params, "period", 10)
		if err != nil {
			return nil, err
		}
		return &transform.EMASmooth{
			Field:  spec.Field,
			Period: period,
		}, nil

	case "percentile_rank":
		window, err := getIntParam(spec.Params, "window", 20)
		if err != nil {
			return nil, err
		}
		return &transform.PercentileRankTransform{
			Field:  spec.Field,
			Window: window,
		}, nil

	case "clip":
		min, err := getFloatParam(spec.Params, "min", 0)
		if err != nil {
			return nil, err
		}
		max, err := getFloatParam(spec.Params, "max", 100)
		if err != nil {
			return nil, err
		}
		return &transform.ClipTransform{
			Field: spec.Field,
			Min:   min,
			Max:   max,
		}, nil

	default:
		return nil, fmt.Errorf("unknown transform type: %s", spec.Type)
	}
}

// Helper functions to extract typed parameters

func getFloatParam(params map[string]interface{}, key string, defaultVal float64) (float64, error) {
	if params == nil {
		return defaultVal, nil
	}

	val, exists := params[key]
	if !exists {
		return defaultVal, nil
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("param %s: expected number, got %T", key, val)
	}
}

func getIntParam(params map[string]interface{}, key string, defaultVal int) (int, error) {
	if params == nil {
		return defaultVal, nil
	}

	val, exists := params[key]
	if !exists {
		return defaultVal, nil
	}

	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("param %s: expected integer, got %T", key, val)
	}
}

func getStringParam(params map[string]interface{}, key string, defaultVal string) (string, error) {
	if params == nil {
		return defaultVal, nil
	}

	val, exists := params[key]
	if !exists {
		return defaultVal, nil
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("param %s: expected string, got %T", key, val)
	}

	return str, nil
}

func getBoolParam(params map[string]interface{}, key string, defaultVal bool) (bool, error) {
	if params == nil {
		return defaultVal, nil
	}

	val, exists := params[key]
	if !exists {
		return defaultVal, nil
	}

	b, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("param %s: expected bool, got %T", key, val)
	}

	return b, nil
}

// ValidateTransformConfig validates transform configuration.
func ValidateTransformConfig(config *TransformConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if !config.Enabled {
		return nil // No validation needed if disabled
	}

	if len(config.Transforms) == 0 {
		return fmt.Errorf("enabled but no transforms defined")
	}

	for i, spec := range config.Transforms {
		if spec.Type == "" {
			return fmt.Errorf("transform %d: type is required", i)
		}

		// Validate transform-specific requirements
		switch spec.Type {
		case "scale":
			if spec.Field == "" {
				return fmt.Errorf("transform %d (scale): field is required", i)
			}
			if _, err := getFloatParam(spec.Params, "factor", 1.0); err != nil {
				return fmt.Errorf("transform %d (scale): %w", i, err)
			}

		case "normalize", "log_returns", "percentage_change":
			if spec.Field == "" {
				return fmt.Errorf("transform %d (%s): field is required", i, spec.Type)
			}

		case "smooth":
			if spec.Field == "" {
				return fmt.Errorf("transform %d (smooth): field is required", i)
			}
			period, err := getIntParam(spec.Params, "period", 5)
			if err != nil {
				return fmt.Errorf("transform %d (smooth): %w", i, err)
			}
			if period < 2 {
				return fmt.Errorf("transform %d (smooth): period must be >= 2", i)
			}

		case "zscore":
			if spec.Field == "" {
				return fmt.Errorf("transform %d (zscore): field is required", i)
			}
			window, err := getIntParam(spec.Params, "window", 20)
			if err != nil {
				return fmt.Errorf("transform %d (zscore): %w", i, err)
			}
			if window < 2 {
				return fmt.Errorf("transform %d (zscore): window must be >= 2", i)
			}

		case "difference":
			if spec.Field == "" {
				return fmt.Errorf("transform %d (difference): field is required", i)
			}
			order, err := getIntParam(spec.Params, "order", 1)
			if err != nil {
				return fmt.Errorf("transform %d (difference): %w", i, err)
			}
			if order < 1 || order > 3 {
				return fmt.Errorf("transform %d (difference): order must be between 1 and 3", i)
			}

		case "ema_smooth":
			if spec.Field == "" {
				return fmt.Errorf("transform %d (ema_smooth): field is required", i)
			}
			period, err := getIntParam(spec.Params, "period", 10)
			if err != nil {
				return fmt.Errorf("transform %d (ema_smooth): %w", i, err)
			}
			if period < 2 {
				return fmt.Errorf("transform %d (ema_smooth): period must be >= 2", i)
			}

		case "percentile_rank":
			if spec.Field == "" {
				return fmt.Errorf("transform %d (percentile_rank): field is required", i)
			}
			window, err := getIntParam(spec.Params, "window", 20)
			if err != nil {
				return fmt.Errorf("transform %d (percentile_rank): %w", i, err)
			}
			if window < 2 {
				return fmt.Errorf("transform %d (percentile_rank): window must be >= 2", i)
			}

		case "clip":
			if spec.Field == "" {
				return fmt.Errorf("transform %d (clip): field is required", i)
			}
			min, err := getFloatParam(spec.Params, "min", 0)
			if err != nil {
				return fmt.Errorf("transform %d (clip): %w", i, err)
			}
			max, err := getFloatParam(spec.Params, "max", 100)
			if err != nil {
				return fmt.Errorf("transform %d (clip): %w", i, err)
			}
			if min >= max {
				return fmt.Errorf("transform %d (clip): min must be < max", i)
			}

		default:
			return fmt.Errorf("transform %d: unknown type: %s", i, spec.Type)
		}
	}

	if config.BatchSize < 0 {
		return fmt.Errorf("batch_size must be >= 0")
	}

	return nil
}
