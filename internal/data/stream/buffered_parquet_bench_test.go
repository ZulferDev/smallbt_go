package stream

import (
	"path/filepath"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/data/parquet"
)

// BenchmarkBufferedParquetReader_Next benchmarks Next() method
func BenchmarkBufferedParquetReader_Next(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.parquet")

	// Create test file with 10K candles
	candles := generateTestCandles(10000)
	if err := writeTestParquet(testFile, candles); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 1000}
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		b.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%10000 == 0 {
			_ = reader.Reset()
		}
		_, err := reader.Next()
		if err != nil {
			_ = reader.Reset()
		}
	}
}

// BenchmarkBufferedParquetReader_NextChunk benchmarks NextChunk() method
func BenchmarkBufferedParquetReader_NextChunk(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.parquet")

	// Create test file with 10K candles
	candles := generateTestCandles(10000)
	if err := writeTestParquet(testFile, candles); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 1000}
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		b.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := reader.NextChunk()
		if err != nil {
			_ = reader.Reset()
		}
	}
}

// BenchmarkBufferedParquetReader_ChunkSizes benchmarks different chunk sizes
func BenchmarkBufferedParquetReader_ChunkSizes(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.parquet")

	// Create test file with 100K candles
	candles := generateTestCandles(100000)
	if err := writeTestParquet(testFile, candles); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			config := BufferConfig{ChunkSize: size}
			reader, err := NewBufferedParquetReader(testFile, config)
			if err != nil {
				b.Fatalf("NewBufferedParquetReader failed: %v", err)
			}
			defer reader.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = reader.Reset()
				for {
					_, err := reader.NextChunk()
					if err != nil {
						break
					}
				}
			}
		})
	}
}

// BenchmarkBufferedParquetReader_MemoryUsage benchmarks memory usage
func BenchmarkBufferedParquetReader_MemoryUsage(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.parquet")

	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(formatSize(size), func(b *testing.B) {
			// Create test file
			candles := generateTestCandles(size)
			if err := writeTestParquet(testFile, candles); err != nil {
				b.Fatalf("Failed to create test file: %v", err)
			}

			config := BufferConfig{ChunkSize: 1000}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader, err := NewBufferedParquetReader(testFile, config)
				if err != nil {
					b.Fatalf("NewBufferedParquetReader failed: %v", err)
				}

				// Read all data
				for {
					_, err := reader.Next()
					if err != nil {
						break
					}
				}

				reader.Close()
			}
		})
	}
}

// BenchmarkBufferedParquetReader_StreamingVsFullLoad compares streaming vs full load
func BenchmarkBufferedParquetReader_StreamingVsFullLoad(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.parquet")

	// Create test file with 100K candles
	candles := generateTestCandles(100000)
	if err := writeTestParquet(testFile, candles); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.Run("Streaming", func(b *testing.B) {
		config := BufferConfig{ChunkSize: 10000}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reader, _ := NewBufferedParquetReader(testFile, config)
			count := 0
			for {
				_, err := reader.Next()
				if err != nil {
					break
				}
				count++
			}
			reader.Close()
		}
	})

	b.Run("FullLoad", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reader, _ := parquet.NewParquetReader(testFile)
			_, _ = reader.Read()
			reader.Close()
		}
	})
}

// BenchmarkBufferedParquetReader_LargeDataset benchmarks with large dataset
func BenchmarkBufferedParquetReader_LargeDataset(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large dataset benchmark in short mode")
	}

	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "large.parquet")

	// Create test file with 1M candles
	b.Log("Creating 1M candles test file...")
	candles := generateTestCandles(1000000)
	if err := writeTestParquet(testFile, candles); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 10000}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader, _ := NewBufferedParquetReader(testFile, config)
		count := 0
		for {
			_, err := reader.Next()
			if err != nil {
				break
			}
			count++
		}
		reader.Close()

		if count != 1000000 {
			b.Errorf("Expected 1M candles, got %d", count)
		}
	}
}

// BenchmarkBufferedParquetReader_Reset benchmarks Reset() performance
func BenchmarkBufferedParquetReader_Reset(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.parquet")

	// Create test file
	candles := generateTestCandles(10000)
	if err := writeTestParquet(testFile, candles); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 1000}
	reader, _ := NewBufferedParquetReader(testFile, config)
	defer reader.Close()

	// Read some data
	for i := 0; i < 5000; i++ {
		_, _ = reader.Next()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reader.Reset()
	}
}

// BenchmarkBufferedParquetReader_Stats benchmarks Stats() performance
func BenchmarkBufferedParquetReader_Stats(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.parquet")

	// Create test file
	candles := generateTestCandles(10000)
	if err := writeTestParquet(testFile, candles); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 1000}
	reader, _ := NewBufferedParquetReader(testFile, config)
	defer reader.Close()

	// Read some data
	for i := 0; i < 5000; i++ {
		_, _ = reader.Next()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reader.Stats()
	}
}

// Helper function to format size
func formatSize(size int) string {
	if size >= 1000000 {
		return "1M"
	} else if size >= 100000 {
		return "100K"
	} else if size >= 10000 {
		return "10K"
	} else if size >= 1000 {
		return "1K"
	}
	return "100"
}
