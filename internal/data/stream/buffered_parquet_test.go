package stream

import (
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/data/parquet"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestNewBufferedParquetReader(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := generateTestCandles(100)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := DefaultBufferConfig()
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		t.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	if reader.config.ChunkSize != config.ChunkSize {
		t.Errorf("Expected chunk size %d, got %d", config.ChunkSize, reader.config.ChunkSize)
	}
}

func TestNewBufferedParquetReader_InvalidConfig(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := generateTestCandles(10)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Invalid config
	config := BufferConfig{ChunkSize: -1}
	_, err := NewBufferedParquetReader(testFile, config)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
}

func TestNewBufferedParquetReader_InvalidFile(t *testing.T) {
	config := DefaultBufferConfig()
	_, err := NewBufferedParquetReader("nonexistent.parquet", config)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestBufferedParquetReader_Next(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file with 100 candles
	candles := generateTestCandles(100)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 10}
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		t.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Read all candles
	count := 0
	for {
		candle, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next() failed: %v", err)
		}
		if candle == nil {
			t.Error("Expected non-nil candle")
		}
		count++
	}

	if count != 100 {
		t.Errorf("Expected 100 candles, got %d", count)
	}
}

func TestBufferedParquetReader_HasNext(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file with 5 candles
	candles := generateTestCandles(5)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 2}
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		t.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Check HasNext before reading
	if !reader.HasNext() {
		t.Error("Expected HasNext() to return true before reading")
	}

	// Read all candles
	for i := 0; i < 5; i++ {
		if !reader.HasNext() {
			t.Errorf("Expected HasNext() to return true at position %d", i)
		}
		_, err := reader.Next()
		if err != nil {
			t.Fatalf("Next() failed: %v", err)
		}
	}

	// Check HasNext after reading all
	if reader.HasNext() {
		t.Error("Expected HasNext() to return false after reading all")
	}
}

func TestBufferedParquetReader_NextChunk(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file with 100 candles
	candles := generateTestCandles(100)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 30}
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		t.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Read first chunk
	chunk1, err := reader.NextChunk()
	if err != nil {
		t.Fatalf("NextChunk() failed: %v", err)
	}
	if len(chunk1) != 30 {
		t.Errorf("Expected chunk size 30, got %d", len(chunk1))
	}

	// Read second chunk
	chunk2, err := reader.NextChunk()
	if err != nil {
		t.Fatalf("NextChunk() failed: %v", err)
	}
	if len(chunk2) != 30 {
		t.Errorf("Expected chunk size 30, got %d", len(chunk2))
	}

	// Read third chunk
	chunk3, err := reader.NextChunk()
	if err != nil {
		t.Fatalf("NextChunk() failed: %v", err)
	}
	if len(chunk3) != 30 {
		t.Errorf("Expected chunk size 30, got %d", len(chunk3))
	}

	// Read last chunk (smaller)
	chunk4, err := reader.NextChunk()
	if err != nil {
		t.Fatalf("NextChunk() failed: %v", err)
	}
	if len(chunk4) != 10 {
		t.Errorf("Expected last chunk size 10, got %d", len(chunk4))
	}

	// Read beyond end
	_, err = reader.NextChunk()
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestBufferedParquetReader_SetChunkSize(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := generateTestCandles(10)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 100}
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		t.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Set chunk size before reading
	err = reader.SetChunkSize(50)
	if err != nil {
		t.Errorf("SetChunkSize() before reading failed: %v", err)
	}

	// Read one candle
	_, err = reader.Next()
	if err != nil {
		t.Fatalf("Next() failed: %v", err)
	}

	// Try to set chunk size after reading
	err = reader.SetChunkSize(25)
	if err == nil {
		t.Error("Expected error when setting chunk size after reading")
	}
}

func TestBufferedParquetReader_Reset(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file with 10 candles
	candles := generateTestCandles(10)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 5}
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		t.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Read 5 candles
	for i := 0; i < 5; i++ {
		_, err := reader.Next()
		if err != nil {
			t.Fatalf("Next() failed: %v", err)
		}
	}

	// Reset
	err = reader.Reset()
	if err != nil {
		t.Fatalf("Reset() failed: %v", err)
	}

	// Read again - should start from beginning
	count := 0
	for {
		_, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next() after reset failed: %v", err)
		}
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 candles after reset, got %d", count)
	}
}

func TestBufferedParquetReader_Close(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := generateTestCandles(10)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := DefaultBufferConfig()
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		t.Fatalf("NewBufferedParquetReader failed: %v", err)
	}

	// Close
	err = reader.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Operations after close should fail
	_, err = reader.Next()
	if err != io.ErrClosedPipe {
		t.Errorf("Expected io.ErrClosedPipe after close, got %v", err)
	}

	// Close again should be safe
	err = reader.Close()
	if err != nil {
		t.Errorf("Second Close() failed: %v", err)
	}
}

func TestBufferedParquetReader_Stats(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file with 100 candles
	candles := generateTestCandles(100)
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := BufferConfig{ChunkSize: 10}
	reader, err := NewBufferedParquetReader(testFile, config)
	if err != nil {
		t.Fatalf("NewBufferedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Read 30 candles
	for i := 0; i < 30; i++ {
		_, err := reader.Next()
		if err != nil {
			t.Fatalf("Next() failed: %v", err)
		}
	}

	stats := reader.Stats()
	if stats.TotalCandles != 100 {
		t.Errorf("Expected TotalCandles 100, got %d", stats.TotalCandles)
	}
	if stats.CurrentIndex != 30 {
		t.Errorf("Expected CurrentIndex 30, got %d", stats.CurrentIndex)
	}
	if stats.ChunkSize != 10 {
		t.Errorf("Expected ChunkSize 10, got %d", stats.ChunkSize)
	}
	if stats.ChunksRead != 3 {
		t.Errorf("Expected ChunksRead 3, got %d", stats.ChunksRead)
	}
}

// Helper functions

func generateTestCandles(count int) []*market.Candle {
	candles := make([]*market.Candle, count)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		candles[i] = &market.Candle{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Open:      100.0 + float64(i),
			High:      105.0 + float64(i),
			Low:       95.0 + float64(i),
			Close:     102.0 + float64(i),
			Volume:    1000.0 + float64(i*10),
		}
	}

	return candles
}

func writeTestParquet(path string, candles []*market.Candle) error {
	writer, err := parquet.NewParquetWriter(path)
	if err != nil {
		return err
	}
	defer writer.Close()

	return writer.Write(candles)
}
