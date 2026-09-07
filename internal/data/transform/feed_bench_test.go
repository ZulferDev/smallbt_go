package transform

import (
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/data/csv"
)

// Benchmark helpers

func createBenchCSV(b *testing.B, count int) string {
	b.Helper()

	// Create temp CSV using testing.T wrapper
	t := &testing.T{}
	return createTestCSV(t, count)
}

// TransformedFeed benchmarks

func BenchmarkTransformedFeed_Next_100(b *testing.B) {
	csvPath := createBenchCSV(b, 100)
	config := csv.DefaultCSVConfig("TEST", "1h")

	feed, err := csv.NewCSVDataFeed(csvPath, config)
	if err != nil {
		b.Fatalf("Failed to create feed: %v", err)
	}
	defer feed.Close()

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))
	_, _ = NewTransformedFeed(feed, chain, 10) // Discard initial setup
	var tfeed *TransformedFeed

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed.Reset()
		tfeed = &TransformedFeed{
			feed:      feed,
			chain:     chain,
			batchSize: 10,
		}

		for j := 0; j < 100; j++ {
			_, err := tfeed.Next()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkTransformedFeed_Next_1K(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	feed, err := csv.NewCSVDataFeed(csvPath, config)
	if err != nil {
		b.Fatalf("Failed to create feed: %v", err)
	}
	defer feed.Close()

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed.Reset()
		tfeed := &TransformedFeed{
			feed:      feed,
			chain:     chain,
			batchSize: 100,
		}

		for j := 0; j < 1000; j++ {
			_, err := tfeed.Next()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkTransformedFeed_ReadAll_100(b *testing.B) {
	csvPath := createBenchCSV(b, 100)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 10)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}

func BenchmarkTransformedFeed_ReadAll_1K(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 100)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}

func BenchmarkTransformedFeed_ReadAll_10K(b *testing.B) {
	csvPath := createBenchCSV(b, 10000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 100)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}

// Complex chain benchmarks

func BenchmarkTransformedFeed_ComplexChain_1K(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewMovingAverageSmoothTransform(5, "close"),
		NewScaleTransform(100.0, "close"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 100)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}

// Batch size comparison

func BenchmarkTransformedFeed_BatchSize10_1K(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 10)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}

func BenchmarkTransformedFeed_BatchSize100_1K(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 100)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}

func BenchmarkTransformedFeed_BatchSize1000_1K(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 1000)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}

// StatsTracker benchmarks

func BenchmarkStatsTracker_1K(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 100)
		tracker := NewStatsTracker(tfeed)

		for j := 0; j < 1000; j++ {
			_, err := tracker.Next()
			if err != nil {
				break
			}
		}

		feed.Close()
	}
}

// Memory allocation benchmarks

func BenchmarkTransformedFeed_ReadAll_1K_Allocs(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 100)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}

func BenchmarkTransformedFeed_ComplexChain_1K_Allocs(b *testing.B) {
	csvPath := createBenchCSV(b, 1000)
	config := csv.DefaultCSVConfig("TEST", "1h")

	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewMovingAverageSmoothTransform(5, "close"),
		NewScaleTransform(100.0, "close"),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		tfeed, _ := NewTransformedFeed(feed, chain, 100)
		_, _ = tfeed.ReadAll()
		feed.Close()
	}
}
