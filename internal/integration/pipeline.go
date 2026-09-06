package integration

import (
	"context"
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/data"
	"github.com/ZulferDev/smallbt_go/internal/data/transform"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// StrategyDataPipeline combines data feed with transform pipeline for strategy execution.
// This allows strategies to receive preprocessed data (normalized, scaled, filtered, etc.)
// before indicator calculation and signal generation.
type StrategyDataPipeline struct {
	feed      data.DataFeed
	tfeed     *transform.TransformedFeed
	chain     *transform.TransformChain
	batchSize int
	hasChain  bool
}

// NewStrategyDataPipeline creates a pipeline with optional transform chain.
func NewStrategyDataPipeline(feed data.DataFeed, chain *transform.TransformChain, batchSize int) (*StrategyDataPipeline, error) {
	if feed == nil {
		return nil, fmt.Errorf("feed cannot be nil")
	}

	if batchSize <= 0 {
		batchSize = 100 // Default
	}

	sdp := &StrategyDataPipeline{
		feed:      feed,
		chain:     chain,
		batchSize: batchSize,
		hasChain:  chain != nil,
	}

	// Create transformed feed if chain provided
	if sdp.hasChain {
		tfeed, err := transform.NewTransformedFeed(feed, chain, batchSize)
		if err != nil {
			return nil, fmt.Errorf("create transformed feed: %w", err)
		}
		sdp.tfeed = tfeed
	}

	return sdp, nil
}

// Next returns the next candle (transformed if chain provided, raw otherwise).
func (sdp *StrategyDataPipeline) Next() (*market.Candle, error) {
	if sdp.hasChain {
		return sdp.tfeed.Next()
	}
	// DataFeed.Next requires context, use background
	ctx := context.Background()
	return sdp.feed.Next(ctx)
}

// ReadAll reads all candles (transformed if chain provided, raw otherwise).
func (sdp *StrategyDataPipeline) ReadAll() ([]*market.Candle, error) {
	if sdp.hasChain {
		return sdp.tfeed.ReadAll()
	}

	// Read all from raw feed
	var candles []*market.Candle
	ctx := context.Background()
	for {
		candle, err := sdp.feed.Next(ctx)
		if err != nil {
			break
		}
		candles = append(candles, candle)
	}
	return candles, nil
}

// HasTransform returns true if pipeline has transform chain.
func (sdp *StrategyDataPipeline) HasTransform() bool {
	return sdp.hasChain
}

// Chain returns the transform chain (nil if none).
func (sdp *StrategyDataPipeline) Chain() *transform.TransformChain {
	return sdp.chain
}

// Close closes the underlying feed.
func (sdp *StrategyDataPipeline) Close() error {
	if sdp.hasChain && sdp.tfeed != nil {
		return sdp.tfeed.Close()
	}
	// DataFeed interface doesn't have Close, but TransformedFeed does
	return nil
}

// MultiStrategyDataPipeline manages pipelines for multiple strategies.
// Each strategy can have its own transform pipeline.
type MultiStrategyDataPipeline struct {
	pipelines map[string]*StrategyDataPipeline
}

// NewMultiStrategyDataPipeline creates a multi-strategy pipeline manager.
func NewMultiStrategyDataPipeline() *MultiStrategyDataPipeline {
	return &MultiStrategyDataPipeline{
		pipelines: make(map[string]*StrategyDataPipeline),
	}
}

// AddStrategy adds a strategy with its data pipeline.
func (msdp *MultiStrategyDataPipeline) AddStrategy(strategyName string, pipeline *StrategyDataPipeline) error {
	if strategyName == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}
	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be nil")
	}

	if _, exists := msdp.pipelines[strategyName]; exists {
		return fmt.Errorf("strategy %s already exists", strategyName)
	}

	msdp.pipelines[strategyName] = pipeline
	return nil
}

// GetPipeline returns the pipeline for a strategy.
func (msdp *MultiStrategyDataPipeline) GetPipeline(strategyName string) (*StrategyDataPipeline, error) {
	pipeline, exists := msdp.pipelines[strategyName]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", strategyName)
	}
	return pipeline, nil
}

// Strategies returns list of registered strategy names.
func (msdp *MultiStrategyDataPipeline) Strategies() []string {
	names := make([]string, 0, len(msdp.pipelines))
	for name := range msdp.pipelines {
		names = append(names, name)
	}
	return names
}

// CloseAll closes all pipelines.
func (msdp *MultiStrategyDataPipeline) CloseAll() error {
	var errs []error
	for name, pipeline := range msdp.pipelines {
		if err := pipeline.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// PipelineBuilder helps construct strategy data pipelines.
type PipelineBuilder struct {
	feed      data.DataFeed
	chain     *transform.TransformChain
	batchSize int
}

// NewPipelineBuilder creates a builder for strategy data pipeline.
func NewPipelineBuilder(feed data.DataFeed) *PipelineBuilder {
	return &PipelineBuilder{
		feed:      feed,
		batchSize: 100,
	}
}

// WithTransformChain sets the transform chain.
func (pb *PipelineBuilder) WithTransformChain(chain *transform.TransformChain) *PipelineBuilder {
	pb.chain = chain
	return pb
}

// WithBatchSize sets the batch size.
func (pb *PipelineBuilder) WithBatchSize(size int) *PipelineBuilder {
	pb.batchSize = size
	return pb
}

// Build creates the strategy data pipeline.
func (pb *PipelineBuilder) Build() (*StrategyDataPipeline, error) {
	return NewStrategyDataPipeline(pb.feed, pb.chain, pb.batchSize)
}
