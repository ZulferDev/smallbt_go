package parquet

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestNewParquetWriter(t *testing.T) {
	t.Run("valid path", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.parquet")

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}
		defer writer.Close()

		if writer.path != filePath {
			t.Errorf("expected path %s, got %s", filePath, writer.path)
		}
	})

	t.Run("invalid directory", func(t *testing.T) {
		_, err := NewParquetWriter("/nonexistent/directory/test.parquet")
		if err == nil {
			t.Fatal("expected error for invalid directory, got nil")
		}
	})
}

func TestParquetWriter_Write(t *testing.T) {
	t.Run("write valid candles", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.parquet")

		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		candles := []*market.Candle{
			{
				Timestamp: baseTime,
				Open:      100.0,
				High:      110.0,
				Low:       90.0,
				Close:     105.0,
				Volume:    1000.0,
			},
			{
				Timestamp: baseTime.Add(time.Hour),
				Open:      105.0,
				High:      115.0,
				Low:       100.0,
				Close:     110.0,
				Volume:    1500.0,
			},
		}

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}

		err = writer.Write(candles)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatal("Parquet file was not created")
		}

		// Read back and verify
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		readCandles, err := reader.Read()
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		if len(readCandles) != len(candles) {
			t.Fatalf("expected %d candles, got %d", len(candles), len(readCandles))
		}

		for i, candle := range readCandles {
			if candle.Timestamp.UnixMilli() != candles[i].Timestamp.UnixMilli() {
				t.Errorf("candle[%d] timestamp mismatch", i)
			}
			if candle.Open != candles[i].Open {
				t.Errorf("candle[%d] open mismatch", i)
			}
			if candle.High != candles[i].High {
				t.Errorf("candle[%d] high mismatch", i)
			}
			if candle.Low != candles[i].Low {
				t.Errorf("candle[%d] low mismatch", i)
			}
			if candle.Close != candles[i].Close {
				t.Errorf("candle[%d] close mismatch", i)
			}
			if candle.Volume != candles[i].Volume {
				t.Errorf("candle[%d] volume mismatch", i)
			}
		}
	})

	t.Run("write empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "empty.parquet")

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}

		err = writer.Write([]*market.Candle{})
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Read back and verify
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		readCandles, err := reader.Read()
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		if len(readCandles) != 0 {
			t.Errorf("expected 0 candles, got %d", len(readCandles))
		}
	})

	t.Run("write nil candle", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.parquet")

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}
		defer writer.Close()

		candles := []*market.Candle{nil}
		err = writer.Write(candles)
		if err == nil {
			t.Fatal("expected error for nil candle, got nil")
		}
	})

	t.Run("write after close", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.parquet")

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}

		writer.Close()

		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		candles := []*market.Candle{
			{Timestamp: baseTime, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		}

		err = writer.Write(candles)
		if err == nil {
			t.Fatal("expected error writing after close, got nil")
		}
	})
}

func TestParquetWriter_WriteOne(t *testing.T) {
	t.Run("write single candle", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "single.parquet")

		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		candle := &market.Candle{
			Timestamp: baseTime,
			Open:      100.0,
			High:      110.0,
			Low:       90.0,
			Close:     105.0,
			Volume:    1000.0,
		}

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}

		err = writer.WriteOne(candle)
		if err != nil {
			t.Fatalf("WriteOne failed: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Read back and verify
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		readCandles, err := reader.Read()
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		if len(readCandles) != 1 {
			t.Fatalf("expected 1 candle, got %d", len(readCandles))
		}

		if readCandles[0].Timestamp.UnixMilli() != candle.Timestamp.UnixMilli() {
			t.Errorf("timestamp mismatch")
		}
		if readCandles[0].Open != candle.Open {
			t.Errorf("open mismatch")
		}
	})

	t.Run("write multiple with WriteOne", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "multiple.parquet")

		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}

		for i := 0; i < 5; i++ {
			candle := &market.Candle{
				Timestamp: baseTime.Add(time.Duration(i) * time.Hour),
				Open:      100.0 + float64(i),
				High:      110.0 + float64(i),
				Low:       90.0 + float64(i),
				Close:     105.0 + float64(i),
				Volume:    1000.0 + float64(i*100),
			}

			err = writer.WriteOne(candle)
			if err != nil {
				t.Fatalf("WriteOne[%d] failed: %v", i, err)
			}
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Read back and verify
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		readCandles, err := reader.Read()
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		if len(readCandles) != 5 {
			t.Fatalf("expected 5 candles, got %d", len(readCandles))
		}
	})

	t.Run("write nil candle with WriteOne", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.parquet")

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}
		defer writer.Close()

		err = writer.WriteOne(nil)
		if err == nil {
			t.Fatal("expected error for nil candle, got nil")
		}
	})
}

func TestParquetWriter_Close(t *testing.T) {
	t.Run("close once", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.parquet")

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	})

	t.Run("close twice", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.parquet")

		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}

		writer.Close()
		err = writer.Close()
		if err != nil {
			t.Fatalf("second Close failed: %v", err)
		}
	})
}

func TestParquetWriter_RoundTrip(t *testing.T) {
	t.Run("large dataset round trip", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "large.parquet")

		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		
		// Create 1000 candles
		candles := make([]*market.Candle, 1000)
		for i := 0; i < 1000; i++ {
			candles[i] = &market.Candle{
				Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
				Open:      100.0 + float64(i)*0.1,
				High:      110.0 + float64(i)*0.1,
				Low:       90.0 + float64(i)*0.1,
				Close:     105.0 + float64(i)*0.1,
				Volume:    1000.0 + float64(i),
			}
		}

		// Write
		writer, err := NewParquetWriter(filePath)
		if err != nil {
			t.Fatalf("NewParquetWriter failed: %v", err)
		}

		err = writer.Write(candles)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Read
		reader, err := NewParquetReader(filePath)
		if err != nil {
			t.Fatalf("NewParquetReader failed: %v", err)
		}
		defer reader.Close()

		readCandles, err := reader.Read()
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		if len(readCandles) != len(candles) {
			t.Fatalf("expected %d candles, got %d", len(candles), len(readCandles))
		}

		// Verify a few samples
		samples := []int{0, 100, 500, 999}
		for _, i := range samples {
			if readCandles[i].Timestamp.UnixMilli() != candles[i].Timestamp.UnixMilli() {
				t.Errorf("candle[%d] timestamp mismatch", i)
			}
			if readCandles[i].Close != candles[i].Close {
				t.Errorf("candle[%d] close mismatch: expected %.2f, got %.2f",
					i, candles[i].Close, readCandles[i].Close)
			}
		}
	})
}
