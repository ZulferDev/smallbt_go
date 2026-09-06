package integration

import (
	"context"
	"io"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/data/transform"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// TestStrategyDataPipeline tests basic pipeline operations
func TestStrategyDataPipeline(t *testing.T) {
	feed := &mockFeed{candles: generateTestCandles(10)}
	chain := transform.NewTransformChain(
		transform.NewScaleTransform(2.0, "close"),
	)

	pipeline, err := NewStrategyDataPipeline(feed, chain, 5)
	if err != nil {
		t.Fatalf("NewStrategyDataPipeline failed: %v", err)
	}
	defer pipeline.Close()

	if !pipeline.HasTransform() {
		t.Error("Expected HasTransform() = true")
	}

	if pipeline.Chain() != chain {
		t.Error("Chain() returned wrong chain")
	}

	// Test Next()
	candle, err := pipeline.Next()
	if err != nil {
		t.Fatalf("Next() failed: %v", err)
	}

	// Should be scaled by 2
	if candle.Close != 202.0 { // 101 * 2
		t.Errorf("Expected scaled close 202, got %.2f", candle.Close)
	}
}

func TestStrategyDataPipeline_NoTransform(t *testing.T) {
	feed := &mockFeed{candles: generateTestCandles(10)}

	pipeline, err := NewStrategyDataPipeline(feed, nil, 5)
	if err != nil {
		t.Fatalf("NewStrategyDataPipeline failed: %v", err)
	}
	defer pipeline.Close()

	if pipeline.HasTransform() {
		t.Error("Expected HasTransform() = false")
	}

	// Test Next() returns raw data
	candle, err := pipeline.Next()
	if err != nil {
		t.Fatalf("Next() failed: %v", err)
	}

	if candle.Close != 101.0 {
		t.Errorf("Expected raw close 101, got %.2f", candle.Close)
	}
}

func TestStrategyDataPipeline_ReadAll(t *testing.T) {
	feed := &mockFeed{candles: generateTestCandles(10)}
	chain := transform.NewTransformChain(
		transform.NewNormalizeTransform("close"),
	)

	pipeline, err := NewStrategyDataPipeline(feed, chain, 5)
	if err != nil {
		t.Fatalf("NewStrategyDataPipeline failed: %v", err)
	}
	defer pipeline.Close()

	candles, err := pipeline.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() failed: %v", err)
	}

	if len(candles) != 10 {
		t.Errorf("Expected 10 candles, got %d", len(candles))
	}

	// Check normalized (should be between 0 and 1)
	for i, candle := range candles {
		if candle.Close < 0 || candle.Close > 1 {
			t.Errorf("Candle %d: close %.2f not normalized", i, candle.Close)
		}
	}
}

// TestMultiStrategyDataPipeline tests multi-strategy management
func TestMultiStrategyDataPipeline(t *testing.T) {
	msdp := NewMultiStrategyDataPipeline()

	// Add strategy A with scale transform
	feedA := &mockFeed{candles: generateTestCandles(10)}
	chainA := transform.NewTransformChain(transform.NewScaleTransform(2.0, "close"))
	pipelineA, _ := NewStrategyDataPipeline(feedA, chainA, 5)

	err := msdp.AddStrategy("strategyA", pipelineA)
	if err != nil {
		t.Fatalf("AddStrategy A failed: %v", err)
	}

	// Add strategy B with normalize transform
	feedB := &mockFeed{candles: generateTestCandles(10)}
	chainB := transform.NewTransformChain(transform.NewNormalizeTransform("close"))
	pipelineB, _ := NewStrategyDataPipeline(feedB, chainB, 5)

	err = msdp.AddStrategy("strategyB", pipelineB)
	if err != nil {
		t.Fatalf("AddStrategy B failed: %v", err)
	}

	// Check strategies registered
	strategies := msdp.Strategies()
	if len(strategies) != 2 {
		t.Errorf("Expected 2 strategies, got %d", len(strategies))
	}

	// Get pipeline A
	pA, err := msdp.GetPipeline("strategyA")
	if err != nil {
		t.Fatalf("GetPipeline A failed: %v", err)
	}
	if pA != pipelineA {
		t.Error("GetPipeline returned wrong pipeline")
	}

	// Test duplicate
	err = msdp.AddStrategy("strategyA", pipelineA)
	if err == nil {
		t.Error("Expected error when adding duplicate strategy")
	}

	// Close all
	if err := msdp.CloseAll(); err != nil {
		t.Errorf("CloseAll failed: %v", err)
	}
}

// TestPipelineBuilder tests builder pattern
func TestPipelineBuilder(t *testing.T) {
	feed := &mockFeed{candles: generateTestCandles(10)}
	chain := transform.NewTransformChain(transform.NewScaleTransform(2.0, "close"))

	pipeline, err := NewPipelineBuilder(feed).
		WithTransformChain(chain).
		WithBatchSize(50).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	defer pipeline.Close()

	if !pipeline.HasTransform() {
		t.Error("Expected transform")
	}
}

// TestTransformConfig_BuildChain tests config-based chain building
func TestTransformConfig_BuildChain(t *testing.T) {
	config := &TransformConfig{
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
				Field: "close",
			},
		},
		BatchSize: 100,
	}

	chain, err := config.BuildTransformChain()
	if err != nil {
		t.Fatalf("BuildTransformChain failed: %v", err)
	}

	if chain == nil {
		t.Fatal("Expected non-nil chain")
	}

	// Test applying chain
	candles := generateTestCandles(5)
	transformed, err := chain.Apply(candles)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(transformed) != 5 {
		t.Errorf("Expected 5 candles, got %d", len(transformed))
	}
}

func TestTransformConfig_Disabled(t *testing.T) {
	config := &TransformConfig{
		Enabled: false,
		Transforms: []TransformSpec{
			{Type: "scale", Field: "close"},
		},
	}

	chain, err := config.BuildTransformChain()
	if err != nil {
		t.Fatalf("BuildTransformChain failed: %v", err)
	}

	if chain != nil {
		t.Error("Expected nil chain when disabled")
	}
}

func TestTransformConfig_UnknownType(t *testing.T) {
	config := &TransformConfig{
		Enabled: true,
		Transforms: []TransformSpec{
			{Type: "unknown_transform", Field: "close"},
		},
	}

	_, err := config.BuildTransformChain()
	if err == nil {
		t.Error("Expected error for unknown transform type")
	}
}

func TestValidateTransformConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *TransformConfig
		wantError bool
	}{
		{
			name: "valid scale",
			config: &TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{
						Type:  "scale",
						Field: "close",
						Params: map[string]interface{}{
							"factor": 2.0,
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing field",
			config: &TransformConfig{
				Enabled: true,
				Transforms: []TransformSpec{
					{Type: "scale"},
				},
			},
			wantError: true,
		},
		{
			name: "disabled valid",
			config: &TransformConfig{
				Enabled:    false,
				Transforms: []TransformSpec{},
			},
			wantError: false,
		},
		{
			name: "enabled but no transforms",
			config: &TransformConfig{
				Enabled:    true,
				Transforms: []TransformSpec{},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransformConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTransformConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// Mock feed for testing

type mockFeed struct {
	candles []*market.Candle
	index   int
}

func (m *mockFeed) Next(ctx context.Context) (*market.Candle, error) {
	if m.index >= len(m.candles) {
		return nil, io.EOF
	}
	candle := m.candles[m.index]
	m.index++
	return candle, nil
}

func (m *mockFeed) Subscribe(ctx context.Context, symbols []string) error {
	return nil
}

func (m *mockFeed) Close() error {
	return nil
}

func generateTestCandles(count int) []*market.Candle {
	candles := make([]*market.Candle, count)
	for i := 0; i < count; i++ {
		price := 100.0 + float64(i)
		candles[i] = &market.Candle{
			Open:   price,
			High:   price + 5,
			Low:    price - 5,
			Close:  price + 1,
			Volume: 100000.0 + float64(i*1000),
		}
	}
	return candles
}
