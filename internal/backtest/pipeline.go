package backtest

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ZulferDev/smallbt_go/internal/data/csv"
	"github.com/ZulferDev/smallbt_go/internal/data/parquet"
	"github.com/ZulferDev/smallbt_go/internal/data/transform"
	"github.com/ZulferDev/smallbt_go/internal/integration"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// loadCandlesWithPipeline loads candles with optional transform pipeline.
// If transformConfig is nil, returns raw candles.
func loadCandlesWithPipeline(path string, tf market.Timeframe, transformConfig *integration.TransformConfig) ([]*market.Candle, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Determine file type and create feed
	ext := filepath.Ext(absPath)

	switch ext {
	case ".csv":
		csvFeed, err := csv.NewCSVDataFeed(absPath, csv.DefaultCSVConfig(market.Symbol(""), tf))
		if err != nil {
			return nil, fmt.Errorf("load csv: %w", err)
		}

		// Build transform chain if config provided
		var chain *transform.TransformChain
		batchSize := 100

		if transformConfig != nil && transformConfig.Enabled {
			if err := integration.ValidateTransformConfig(transformConfig); err != nil {
				return nil, fmt.Errorf("invalid transform config: %w", err)
			}

			chain, err = transformConfig.BuildTransformChain()
			if err != nil {
				return nil, fmt.Errorf("build transform chain: %w", err)
			}

			if transformConfig.BatchSize > 0 {
				batchSize = transformConfig.BatchSize
			}
		}

		// Create pipeline
		pipeline, err := integration.NewStrategyDataPipeline(csvFeed, chain, batchSize)
		if err != nil {
			return nil, fmt.Errorf("create pipeline: %w", err)
		}
		defer pipeline.Close()

		return pipeline.ReadAll()

	case ".parquet":
		// Parquet: load directly and apply transforms
		reader, err := parquet.NewParquetReader(absPath)
		if err != nil {
			return nil, fmt.Errorf("load parquet: %w", err)
		}

		candles, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("read parquet: %w", err)
		}

		// Apply transforms if configured
		if transformConfig != nil && transformConfig.Enabled {
			if err := integration.ValidateTransformConfig(transformConfig); err != nil {
				return nil, fmt.Errorf("invalid transform config: %w", err)
			}

			chain, err := transformConfig.BuildTransformChain()
			if err != nil {
				return nil, fmt.Errorf("build transform chain: %w", err)
			}

			candles, err = chain.Apply(candles)
			if err != nil {
				return nil, fmt.Errorf("apply transforms: %w", err)
			}
		}

		return candles, nil

	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// createStreamingPipeline creates a streaming pipeline for incremental processing.
// Returns the pipeline (caller must close it).
// Only supports CSV for now (parquet doesn't have streaming feed).
func createStreamingPipeline(path string, tf market.Timeframe, transformConfig *integration.TransformConfig) (*integration.StrategyDataPipeline, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(absPath)
	if ext != ".csv" {
		return nil, fmt.Errorf("streaming pipeline only supports CSV, got: %s", ext)
	}

	// Create CSV feed
	csvFeed, err := csv.NewCSVDataFeed(absPath, csv.DefaultCSVConfig(market.Symbol(""), tf))
	if err != nil {
		return nil, fmt.Errorf("load csv: %w", err)
	}

	// Subscribe to initialize
	ctx := context.Background()
	if err := csvFeed.Subscribe(ctx, []string{string(market.Symbol(""))}); err != nil {
		return nil, fmt.Errorf("subscribe csv: %w", err)
	}

	// Build transform chain
	var chain *transform.TransformChain
	batchSize := 100

	if transformConfig != nil && transformConfig.Enabled {
		if err := integration.ValidateTransformConfig(transformConfig); err != nil {
			return nil, fmt.Errorf("invalid transform config: %w", err)
		}

		chain, err = transformConfig.BuildTransformChain()
		if err != nil {
			return nil, fmt.Errorf("build transform chain: %w", err)
		}

		if transformConfig.BatchSize > 0 {
			batchSize = transformConfig.BatchSize
		}
	}

	// Create pipeline
	pipeline, err := integration.NewStrategyDataPipeline(csvFeed, chain, batchSize)
	if err != nil {
		return nil, fmt.Errorf("create pipeline: %w", err)
	}

	return pipeline, nil
}
