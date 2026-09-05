package parquet

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Helper function to create a temporary Parquet file with test data
func createTestParquetFile(t *testing.T, candles []CandleParquet) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.parquet")

	fw, err := local.NewLocalFileWriter(filePath)
	if err != nil {
		t.Fatalf("create file writer: %v", err)
	}

	pw, err := writer.NewParquetWriter(fw, new(CandleParquet), 4)
	if err != nil {
		t.Fatalf("create parquet writer: %v", err)
	}

	for i := range candles {
		if err := pw.Write(candles[i]); err != nil {
			t.Fatalf("write candle: %v", err)
		}
	}

	if err := pw.WriteStop(); err != nil {
		t.Fatalf("write stop: %v", err)
	}

	fw.Close()

	return filePath
}

func TestNewParquetReader(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		candles := []CandleParquet{
			{Timestamp: 1000, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		}
		filePath := createTestParquetFile(t, candles)

		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		if reader.path != filePath {
			t.Errorf("expected path %s, got %s", filePath, reader.path)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewParquetReader("/nonexistent/file.parquet")
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
	})

	t.Run("invalid file format", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "invalid.parquet")
		
		// Create a non-Parquet file
		if err := os.WriteFile(filePath, []byte("not a parquet file"), 0644); err != nil {
			t.Fatalf("write invalid file: %v", err)
		}

		_, err := NewParquetReader(filePath)
		if err == nil {
			t.Fatal("expected error for invalid Parquet file, got nil")
		}
	})
}

func TestParquetReader_Read(t *testing.T) {
	t.Run("read valid candles", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		
		expected := []CandleParquet{
			{
				Timestamp: baseTime.UnixMilli(),
				Open:      100.0,
				High:      110.0,
				Low:       90.0,
				Close:     105.0,
				Volume:    1000.0,
			},
			{
				Timestamp: baseTime.Add(time.Hour).UnixMilli(),
				Open:      105.0,
				High:      115.0,
				Low:       100.0,
				Close:     110.0,
				Volume:    1500.0,
			},
			{
				Timestamp: baseTime.Add(2 * time.Hour).UnixMilli(),
				Open:      110.0,
				High:      120.0,
				Low:       105.0,
				Close:     115.0,
				Volume:    2000.0,
			},
		}

		filePath := createTestParquetFile(t, expected)
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		candles, err := reader.Read()
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		if len(candles) != len(expected) {
			t.Fatalf("expected %d candles, got %d", len(expected), len(candles))
		}

		for i, candle := range candles {
			if candle.Timestamp.UnixMilli() != expected[i].Timestamp {
				t.Errorf("candle[%d] timestamp: expected %d, got %d", 
					i, expected[i].Timestamp, candle.Timestamp.UnixMilli())
			}
			if candle.Open != expected[i].Open {
				t.Errorf("candle[%d] open: expected %.2f, got %.2f", 
					i, expected[i].Open, candle.Open)
			}
			if candle.High != expected[i].High {
				t.Errorf("candle[%d] high: expected %.2f, got %.2f", 
					i, expected[i].High, candle.High)
			}
			if candle.Low != expected[i].Low {
				t.Errorf("candle[%d] low: expected %.2f, got %.2f", 
					i, expected[i].Low, candle.Low)
			}
			if candle.Close != expected[i].Close {
				t.Errorf("candle[%d] close: expected %.2f, got %.2f", 
					i, expected[i].Close, candle.Close)
			}
			if candle.Volume != expected[i].Volume {
				t.Errorf("candle[%d] volume: expected %.2f, got %.2f", 
					i, expected[i].Volume, candle.Volume)
			}
		}
	})

	t.Run("read empty file", func(t *testing.T) {
		filePath := createTestParquetFile(t, []CandleParquet{})
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		candles, err := reader.Read()
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		if len(candles) != 0 {
			t.Errorf("expected 0 candles, got %d", len(candles))
		}
	})

	t.Run("read after close", func(t *testing.T) {
		candles := []CandleParquet{
			{Timestamp: 1000, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		}
		filePath := createTestParquetFile(t, candles)

		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}

		reader.Close()

		_, err = reader.Read()
		if err == nil {
			t.Fatal("expected error reading after close, got nil")
		}
	})
}

func TestParquetReader_ReadRange(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	candles := []CandleParquet{
		{Timestamp: baseTime.UnixMilli(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: baseTime.Add(1 * time.Hour).UnixMilli(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		{Timestamp: baseTime.Add(2 * time.Hour).UnixMilli(), Open: 110, High: 120, Low: 105, Close: 115, Volume: 2000},
		{Timestamp: baseTime.Add(3 * time.Hour).UnixMilli(), Open: 115, High: 125, Low: 110, Close: 120, Volume: 2500},
		{Timestamp: baseTime.Add(4 * time.Hour).UnixMilli(), Open: 120, High: 130, Low: 115, Close: 125, Volume: 3000},
	}

	filePath := createTestParquetFile(t, candles)

	t.Run("range in middle", func(t *testing.T) {
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()
		start := baseTime.Add(1 * time.Hour)
		end := baseTime.Add(3 * time.Hour)

		result, err := reader.ReadRange(start, end)
		if err != nil {
			t.Fatalf("ReadRange failed: %v", err)
		}

		if len(result) != 3 {
			t.Fatalf("expected 3 candles, got %d", len(result))
		}

		// Check timestamps are within range
		for i, candle := range result {
			if candle.Timestamp.Before(start) || candle.Timestamp.After(end) {
				t.Errorf("candle[%d] timestamp %v is outside range [%v, %v]",
					i, candle.Timestamp, start, end)
			}
		}
	})

	t.Run("range at start", func(t *testing.T) {
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		start := baseTime
		end := baseTime.Add(1 * time.Hour)

		result, err := reader.ReadRange(start, end)
		if err != nil {
			t.Fatalf("ReadRange failed: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 2 candles, got %d", len(result))
		}
	})

	t.Run("range at end", func(t *testing.T) {
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		start := baseTime.Add(3 * time.Hour)
		end := baseTime.Add(4 * time.Hour)

		result, err := reader.ReadRange(start, end)
		if err != nil {
			t.Fatalf("ReadRange failed: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 2 candles, got %d", len(result))
		}
	})

	t.Run("range outside data", func(t *testing.T) {
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		start := baseTime.Add(10 * time.Hour)
		end := baseTime.Add(20 * time.Hour)

		result, err := reader.ReadRange(start, end)
		if err != nil {
			t.Fatalf("ReadRange failed: %v", err)
		}

		if len(result) != 0 {
			t.Fatalf("expected 0 candles, got %d", len(result))
		}
	})

	t.Run("range covers all data", func(t *testing.T) {
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		start := baseTime.Add(-1 * time.Hour)
		end := baseTime.Add(10 * time.Hour)

		result, err := reader.ReadRange(start, end)
		if err != nil {
			t.Fatalf("ReadRange failed: %v", err)
		}

		if len(result) != 5 {
			t.Fatalf("expected 5 candles, got %d", len(result))
		}
	})
}

func TestParquetReader_Close(t *testing.T) {
	t.Run("close once", func(t *testing.T) {
		candles := []CandleParquet{
			{Timestamp: 1000, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		}
		filePath := createTestParquetFile(t, candles)

		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}

		err = reader.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	})

	t.Run("close twice", func(t *testing.T) {
		candles := []CandleParquet{
			{Timestamp: 1000, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		}
		filePath := createTestParquetFile(t, candles)

		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}

		reader.Close()
		err = reader.Close()
		if err != nil {
			t.Fatalf("second Close failed: %v", err)
		}
	})
}

func TestCandleParquet_Conversions(t *testing.T) {
	t.Run("ToMarketCandle", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		
		parquetCandle := CandleParquet{
			Timestamp: baseTime.UnixMilli(),
			Open:      100.0,
			High:      110.0,
			Low:       90.0,
			Close:     105.0,
			Volume:    1000.0,
		}

		marketCandle := parquetCandle.ToMarketCandle()

		if marketCandle.Timestamp.UnixMilli() != parquetCandle.Timestamp {
			t.Errorf("timestamp mismatch: expected %d, got %d",
				parquetCandle.Timestamp, marketCandle.Timestamp.UnixMilli())
		}
		if marketCandle.Open != parquetCandle.Open {
			t.Errorf("open mismatch: expected %.2f, got %.2f",
				parquetCandle.Open, marketCandle.Open)
		}
		if marketCandle.High != parquetCandle.High {
			t.Errorf("high mismatch: expected %.2f, got %.2f",
				parquetCandle.High, marketCandle.High)
		}
		if marketCandle.Low != parquetCandle.Low {
			t.Errorf("low mismatch: expected %.2f, got %.2f",
				parquetCandle.Low, marketCandle.Low)
		}
		if marketCandle.Close != parquetCandle.Close {
			t.Errorf("close mismatch: expected %.2f, got %.2f",
				parquetCandle.Close, marketCandle.Close)
		}
		if marketCandle.Volume != parquetCandle.Volume {
			t.Errorf("volume mismatch: expected %.2f, got %.2f",
				parquetCandle.Volume, marketCandle.Volume)
		}
	})

	t.Run("FromMarketCandle", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		
		marketCandle := &market.Candle{
			Timestamp: baseTime,
			Open:      100.0,
			High:      110.0,
			Low:       90.0,
			Close:     105.0,
			Volume:    1000.0,
		}

		parquetCandle := FromMarketCandle(marketCandle)

		if parquetCandle.Timestamp != marketCandle.Timestamp.UnixMilli() {
			t.Errorf("timestamp mismatch: expected %d, got %d",
				marketCandle.Timestamp.UnixMilli(), parquetCandle.Timestamp)
		}
		if parquetCandle.Open != marketCandle.Open {
			t.Errorf("open mismatch: expected %.2f, got %.2f",
				marketCandle.Open, parquetCandle.Open)
		}
		if parquetCandle.High != marketCandle.High {
			t.Errorf("high mismatch: expected %.2f, got %.2f",
				marketCandle.High, parquetCandle.High)
		}
		if parquetCandle.Low != marketCandle.Low {
			t.Errorf("low mismatch: expected %.2f, got %.2f",
				marketCandle.Low, parquetCandle.Low)
		}
		if parquetCandle.Close != marketCandle.Close {
			t.Errorf("close mismatch: expected %.2f, got %.2f",
				marketCandle.Close, parquetCandle.Close)
		}
		if parquetCandle.Volume != marketCandle.Volume {
			t.Errorf("volume mismatch: expected %.2f, got %.2f",
				marketCandle.Volume, parquetCandle.Volume)
		}
	})

	t.Run("round trip conversion", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		
		original := &market.Candle{
			Timestamp: baseTime,
			Open:      100.0,
			High:      110.0,
			Low:       90.0,
			Close:     105.0,
			Volume:    1000.0,
		}

		// market.Candle -> CandleParquet -> market.Candle
		parquet := FromMarketCandle(original)
		result := parquet.ToMarketCandle()

		if result.Timestamp.UnixMilli() != original.Timestamp.UnixMilli() {
			t.Errorf("timestamp mismatch after round trip")
		}
		if result.Open != original.Open {
			t.Errorf("open mismatch after round trip")
		}
		if result.High != original.High {
			t.Errorf("high mismatch after round trip")
		}
		if result.Low != original.Low {
			t.Errorf("low mismatch after round trip")
		}
		if result.Close != original.Close {
			t.Errorf("close mismatch after round trip")
		}
		if result.Volume != original.Volume {
			t.Errorf("volume mismatch after round trip")
		}
	})
}
