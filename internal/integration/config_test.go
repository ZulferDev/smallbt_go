package integration

import (
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/data/transform"
)

// TestGetIntParam tests getIntParam helper function.
func TestGetIntParam(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]interface{}
		key        string
		defaultVal int
		want       int
		wantErr    bool
	}{
		{
			name:       "nil params returns default",
			params:     nil,
			key:        "period",
			defaultVal: 10,
			want:       10,
			wantErr:    false,
		},
		{
			name:       "missing key returns default",
			params:     map[string]interface{}{"other": 5},
			key:        "period",
			defaultVal: 10,
			want:       10,
			wantErr:    false,
		},
		{
			name:       "valid int value",
			params:     map[string]interface{}{"period": 20},
			key:        "period",
			defaultVal: 10,
			want:       20,
			wantErr:    false,
		},
		{
			name:       "float64 converted to int",
			params:     map[string]interface{}{"period": 15.0},
			key:        "period",
			defaultVal: 10,
			want:       15,
			wantErr:    false,
		},
		{
			name:       "invalid type returns error",
			params:     map[string]interface{}{"period": "not_a_number"},
			key:        "period",
			defaultVal: 10,
			want:       0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIntParam(tt.params, tt.key, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIntParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIntParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetStringParam tests getStringParam helper function.
func TestGetStringParam(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]interface{}
		key        string
		defaultVal string
		want       string
		wantErr    bool
	}{
		{
			name:       "nil params returns default",
			params:     nil,
			key:        "field",
			defaultVal: "close",
			want:       "close",
			wantErr:    false,
		},
		{
			name:       "missing key returns default",
			params:     map[string]interface{}{"other": "value"},
			key:        "field",
			defaultVal: "close",
			want:       "close",
			wantErr:    false,
		},
		{
			name:       "valid string value",
			params:     map[string]interface{}{"field": "open"},
			key:        "field",
			defaultVal: "close",
			want:       "open",
			wantErr:    false,
		},
		{
			name:       "invalid type returns error",
			params:     map[string]interface{}{"field": 123},
			key:        "field",
			defaultVal: "close",
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getStringParam(tt.params, tt.key, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("getStringParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getStringParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetBoolParam tests getBoolParam helper function.
func TestGetBoolParam(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]interface{}
		key        string
		defaultVal bool
		want       bool
		wantErr    bool
	}{
		{
			name:       "nil params returns default",
			params:     nil,
			key:        "enabled",
			defaultVal: true,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "missing key returns default",
			params:     map[string]interface{}{"other": false},
			key:        "enabled",
			defaultVal: true,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "valid bool value true",
			params:     map[string]interface{}{"enabled": true},
			key:        "enabled",
			defaultVal: false,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "valid bool value false",
			params:     map[string]interface{}{"enabled": false},
			key:        "enabled",
			defaultVal: true,
			want:       false,
			wantErr:    false,
		},
		{
			name:       "invalid type returns error",
			params:     map[string]interface{}{"enabled": "yes"},
			key:        "enabled",
			defaultVal: false,
			want:       false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBoolParam(tt.params, tt.key, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBoolParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getBoolParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBuildTransform tests buildTransform for all transform types.
func TestBuildTransform(t *testing.T) {
	tests := []struct {
		name    string
		spec    TransformSpec
		wantErr bool
	}{
		{
			name: "scale transform",
			spec: TransformSpec{
				Type:  "scale",
				Field: "close",
				Params: map[string]interface{}{
					"factor": 2.0,
				},
			},
			wantErr: false,
		},
		{
			name: "normalize transform",
			spec: TransformSpec{
				Type:  "normalize",
				Field: "volume",
			},
			wantErr: false,
		},
		{
			name: "log_returns transform",
			spec: TransformSpec{
				Type:  "log_returns",
				Field: "close",
			},
			wantErr: false,
		},
		{
			name: "percentage_change transform",
			spec: TransformSpec{
				Type:  "percentage_change",
				Field: "close",
			},
			wantErr: false,
		},
		{
			name: "smooth transform",
			spec: TransformSpec{
				Type:  "smooth",
				Field: "close",
				Params: map[string]interface{}{
					"period": 5,
				},
			},
			wantErr: false,
		},
		{
			name: "zscore transform",
			spec: TransformSpec{
				Type:  "zscore",
				Field: "close",
				Params: map[string]interface{}{
					"window": 20,
				},
			},
			wantErr: false,
		},
		{
			name: "difference transform",
			spec: TransformSpec{
				Type:  "difference",
				Field: "close",
				Params: map[string]interface{}{
					"order": 1,
				},
			},
			wantErr: false,
		},
		{
			name: "ema_smooth transform",
			spec: TransformSpec{
				Type:  "ema_smooth",
				Field: "close",
				Params: map[string]interface{}{
					"period": 10,
				},
			},
			wantErr: false,
		},
		{
			name: "unknown transform type",
			spec: TransformSpec{
				Type:  "unknown_type",
				Field: "close",
			},
			wantErr: true,
		},
		{
			name: "scale with invalid factor type",
			spec: TransformSpec{
				Type:  "scale",
				Field: "close",
				Params: map[string]interface{}{
					"factor": "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "smooth with invalid period type",
			spec: TransformSpec{
				Type:  "smooth",
				Field: "close",
				Params: map[string]interface{}{
					"period": "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildTransform(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildTransform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("buildTransform() returned nil transform without error")
			}
		})
	}
}

// TestBuildTransformChain tests BuildTransformChain method.
func TestBuildTransformChain(t *testing.T) {
	tests := []struct {
		name    string
		config  TransformConfig
		wantNil bool
		wantErr bool
	}{
		{
			name: "disabled config returns nil",
			config: TransformConfig{
				Enabled: false,
				Transforms: []TransformSpec{
					{Type: "scale", Field: "close", Params: map[string]interface{}{"factor": 2.0}},
				},
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "empty transforms returns nil",
			config: TransformConfig{
				Enabled:    true,
				Transforms: []TransformSpec{},
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "valid single transform",
			config: TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{Type: "scale", Field: "close", Params: map[string]interface{}{"factor": 2.0}},
				},
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "valid multiple transforms",
			config: TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{Type: "normalize", Field: "close"},
					{Type: "smooth", Field: "close", Params: map[string]interface{}{"period": 5}},
					{Type: "scale", Field: "close", Params: map[string]interface{}{"factor": 1.5}},
				},
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "invalid transform in chain",
			config: TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{Type: "scale", Field: "close", Params: map[string]interface{}{"factor": 2.0}},
					{Type: "unknown_type", Field: "close"},
				},
			},
			wantNil: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.BuildTransformChain()
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildTransformChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil && got != nil {
				t.Error("BuildTransformChain() expected nil, got non-nil")
			}
			if !tt.wantNil && !tt.wantErr && got == nil {
				t.Error("BuildTransformChain() returned nil without error")
			}
		})
	}
}

// TestValidateTransformConfigExtended tests ValidateTransformConfig function with additional cases.
func TestValidateTransformConfigExtended(t *testing.T) {
	tests := []struct {
		name    string
		config  *TransformConfig
		wantErr bool
	}{
		{
			name:    "nil config returns error",
			config:  nil,
			wantErr: true,
		},
		{
			name: "disabled config is valid",
			config: &TransformConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "valid enabled config",
			config: &TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{Type: "scale", Field: "close", Params: map[string]interface{}{"factor": 2.0}},
				},
			},
			wantErr: false,
		},
		{
			name: "empty field is invalid",
			config: &TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{Type: "scale", Field: "", Params: map[string]interface{}{"factor": 2.0}},
				},
			},
			wantErr: true,
		},
		{
			name: "empty type is invalid",
			config: &TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{Type: "", Field: "close"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid transform type",
			config: &TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{Type: "invalid_type", Field: "close"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransformConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransformConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTransformChainIntegration tests end-to-end transform chain creation and usage.
func TestTransformChainIntegration(t *testing.T) {
	config := TransformConfig{
		Enabled: true,
		Transforms: []TransformSpec{
			{
				Type:  "scale",
				Field: "close",
				Params: map[string]interface{}{
					"factor": 2.0,
				},
			},
			{
				Type:  "normalize",
				Field: "volume",
			},
		},
		BatchSize: 50,
	}

	chain, err := config.BuildTransformChain()
	if err != nil {
		t.Fatalf("BuildTransformChain() failed: %v", err)
	}

	if chain == nil {
		t.Fatal("BuildTransformChain() returned nil chain")
	}

	// Verify chain is usable (implements Transform interface)
	var _ transform.Transform = chain
}
