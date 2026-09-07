package backtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/integration"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
)

// TestCreateOrderRequest tests createOrderRequest function for all order types.
func TestCreateOrderRequest(t *testing.T) {
	symbol := market.Symbol("BTCUSDT")
	quantity := 1.0
	price := 50000.0

	tests := []struct {
		name      string
		side      order.OrderSide
		orderType string
		wantType  order.OrderType
		wantPrice bool
		wantStop  bool
	}{
		{
			name:      "market buy order",
			side:      order.OrderSideBuy,
			orderType: "market",
			wantType:  order.OrderTypeMarket,
			wantPrice: false,
			wantStop:  false,
		},
		{
			name:      "market sell order",
			side:      order.OrderSideSell,
			orderType: "market",
			wantType:  order.OrderTypeMarket,
			wantPrice: false,
			wantStop:  false,
		},
		{
			name:      "limit buy order",
			side:      order.OrderSideBuy,
			orderType: "limit",
			wantType:  order.OrderTypeLimit,
			wantPrice: true,
			wantStop:  false,
		},
		{
			name:      "limit sell order",
			side:      order.OrderSideSell,
			orderType: "limit",
			wantType:  order.OrderTypeLimit,
			wantPrice: true,
			wantStop:  false,
		},
		{
			name:      "stop buy order",
			side:      order.OrderSideBuy,
			orderType: "stop",
			wantType:  order.OrderTypeStop,
			wantPrice: false,
			wantStop:  true,
		},
		{
			name:      "stop sell order",
			side:      order.OrderSideSell,
			orderType: "stop",
			wantType:  order.OrderTypeStop,
			wantPrice: false,
			wantStop:  true,
		},
		{
			name:      "stop_limit buy order",
			side:      order.OrderSideBuy,
			orderType: "stop_limit",
			wantType:  order.OrderTypeStopLimit,
			wantPrice: true,
			wantStop:  true,
		},
		{
			name:      "stop_limit sell order",
			side:      order.OrderSideSell,
			orderType: "stop_limit",
			wantType:  order.OrderTypeStopLimit,
			wantPrice: true,
			wantStop:  true,
		},
		{
			name:      "unknown type defaults to market",
			side:      order.OrderSideBuy,
			orderType: "unknown",
			wantType:  order.OrderTypeMarket,
			wantPrice: false,
			wantStop:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createOrderRequest(symbol, tt.side, tt.orderType, quantity, price)

			if req.Symbol != symbol {
				t.Errorf("Symbol = %v, want %v", req.Symbol, symbol)
			}
			if req.Side != tt.side {
				t.Errorf("Side = %v, want %v", req.Side, tt.side)
			}
			if req.Quantity != quantity {
				t.Errorf("Quantity = %v, want %v", req.Quantity, quantity)
			}
			if req.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", req.Type, tt.wantType)
			}

			if tt.wantPrice {
				if req.Price == nil {
					t.Error("Price should be set but is nil")
				} else if *req.Price != price {
					t.Errorf("Price = %v, want %v", *req.Price, price)
				}
			} else {
				if req.Price != nil {
					t.Errorf("Price should be nil but is %v", *req.Price)
				}
			}

			if tt.wantStop {
				if req.StopPrice == nil {
					t.Error("StopPrice should be set but is nil")
				} else if *req.StopPrice != price {
					t.Errorf("StopPrice = %v, want %v", *req.StopPrice, price)
				}
			} else {
				if req.StopPrice != nil {
					t.Errorf("StopPrice should be nil but is %v", *req.StopPrice)
				}
			}
		})
	}
}

// TestCreateStreamingPipeline tests createStreamingPipeline function.
func TestCreateStreamingPipeline(t *testing.T) {
	// Create temporary CSV file for testing
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test_data.csv")
	
	csvContent := `timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,100,110,90,105,1000
2024-01-01T01:00:00Z,105,115,95,110,1200
2024-01-01T02:00:00Z,110,120,100,115,1500`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		timeframe       market.Timeframe
		transformConfig *integration.TransformConfig
		wantErr         bool
		errContains     string
	}{
		{
			name:            "valid csv without transforms",
			path:            csvPath,
			timeframe:       market.Timeframe1h,
			transformConfig: nil,
			wantErr:         false,
		},
		{
			name:      "valid csv with disabled transforms",
			path:      csvPath,
			timeframe: market.Timeframe1h,
			transformConfig: &integration.TransformConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name:      "valid csv with enabled transforms",
			path:      csvPath,
			timeframe: market.Timeframe1h,
			transformConfig: &integration.TransformConfig{
				Enabled: true,
				Transforms: []integration.TransformSpec{
					{
						Type:  "scale",
						Field: "close",
						Params: map[string]interface{}{
							"factor": 2.0,
						},
					},
				},
				BatchSize: 50,
			},
			wantErr: false,
		},
		{
			name:            "non-existent file",
			path:            filepath.Join(tmpDir, "nonexistent.csv"),
			timeframe:       market.Timeframe1h,
			transformConfig: nil,
			wantErr:         true,
			errContains:     "load csv",
		},
		{
			name:            "non-csv file",
			path:            filepath.Join(tmpDir, "test.txt"),
			timeframe:       market.Timeframe1h,
			transformConfig: nil,
			wantErr:         true,
			errContains:     "only supports CSV",
		},
		{
			name:      "invalid transform config",
			path:      csvPath,
			timeframe: market.Timeframe1h,
			transformConfig: &integration.TransformConfig{
				Enabled: true,
				Transforms: []integration.TransformSpec{
					{
						Type:  "scale",
						Field: "", // Empty field is invalid
					},
				},
			},
			wantErr:     true,
			errContains: "invalid transform config",
		},
		{
			name:      "unknown transform type",
			path:      csvPath,
			timeframe: market.Timeframe1h,
			transformConfig: &integration.TransformConfig{
				Enabled: true,
				Transforms: []integration.TransformSpec{
					{
						Type:  "unknown_type",
						Field: "close",
					},
				},
			},
			wantErr:     true,
			errContains: "invalid transform config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := createStreamingPipeline(tt.path, tt.timeframe, tt.transformConfig)

			if tt.wantErr {
				if err == nil {
					t.Error("createStreamingPipeline() expected error but got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("createStreamingPipeline() error = %v, should contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("createStreamingPipeline() unexpected error = %v", err)
				return
			}

			if pipeline == nil {
				t.Error("createStreamingPipeline() returned nil pipeline without error")
				return
			}

			// Clean up pipeline
			if err := pipeline.Close(); err != nil {
				t.Logf("Warning: failed to close pipeline: %v", err)
			}
		})
	}
}

// TestCreateStreamingPipelineWithTransforms tests transform integration.
func TestCreateStreamingPipelineWithTransforms(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test_transforms.csv")
	
	csvContent := `timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,100,110,90,105,1000
2024-01-01T01:00:00Z,105,115,95,110,1200`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	config := &integration.TransformConfig{
		Enabled: true,
		Transforms: []integration.TransformSpec{
			{
				Type:  "normalize",
				Field: "close",
			},
			{
				Type:  "smooth",
				Field: "close",
				Params: map[string]interface{}{
					"period": 2,
				},
			},
		},
		BatchSize: 100,
	}

	pipeline, err := createStreamingPipeline(csvPath, market.Timeframe1h, config)
	if err != nil {
		t.Fatalf("createStreamingPipeline() with transforms failed: %v", err)
	}

	if pipeline == nil {
		t.Fatal("createStreamingPipeline() returned nil pipeline")
	}

	// Clean up
	if err := pipeline.Close(); err != nil {
		t.Logf("Warning: failed to close pipeline: %v", err)
	}
}

// Helper function to check if string contains substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
