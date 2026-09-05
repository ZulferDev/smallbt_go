package parquet

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// generateBenchmarkCandles creates test candles for benchmarking
func generateBenchmarkCandles(n int) []*market.Candle {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]*market.Candle, n)
	
	for i := 0; i < n; i++ {
		candles[i] = &market.Candle{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Open:      100.0 + float64(i)*0.1,
			High:      110.0 + float64(i)*0.1,
			Low:       90.0 + float64(i)*0.1,
			Close:     105.0 + float64(i)*0.1,
			Volume:    1000.0 + float64(i),
		}
	}
	
	return candles
}

func BenchmarkParquetWriter_Write(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	
	for _, size := range sizes {
		b.Run(string(rune(size)), func(b *testing.B) {
			candles := generateBenchmarkCandles(size)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				tmpDir := b.TempDir()
				filePath := filepath.Join(tmpDir, "bench.parquet")
				b.StartTimer()
				
				writer, err := NewParquetWriter(filePath)
				if err != nil {
					b.Fatalf("NewParquetWriter failed: %v", err)
				}
				
				err = writer.Write(candles)
				if err != nil {
					b.Fatalf("Write failed: %v", err)
				}
				
				err = writer.Close()
				if err != nil {
					b.Fatalf("Close failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkParquetReader_Read(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	
	for _, size := range sizes {
		b.Run(string(rune(size)), func(b *testing.B) {
			// Setup: create file once
			tmpDir := b.TempDir()
			filePath := filepath.Join(tmpDir, "bench.parquet")
			candles := generateBenchmarkCandles(size)
			
			writer, err := NewParquetWriter(filePath)
			if err != nil {
				b.Fatalf("NewParquetWriter failed: %v", err)
			}
			writer.Write(candles)
			writer.Close()
			
			// Benchmark reading
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reader, err := NewParquetReader(filePath)
				if err != nil {
					b.Fatalf("NewParquetReader failed: %v", err)
				}
				
				_, err = reader.Read()
				if err != nil {
					b.Fatalf("Read failed: %v", err)
				}
				
				reader.Close()
			}
		})
	}
}

func BenchmarkParquetRoundTrip(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	
	for _, size := range sizes {
		b.Run(string(rune(size)), func(b *testing.B) {
			candles := generateBenchmarkCandles(size)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				tmpDir := b.TempDir()
				filePath := filepath.Join(tmpDir, "bench.parquet")
				b.StartTimer()
				
				// Write
				writer, err := NewParquetWriter(filePath)
				if err != nil {
					b.Fatalf("NewParquetWriter failed: %v", err)
				}
				
				err = writer.Write(candles)
				if err != nil {
					b.Fatalf("Write failed: %v", err)
				}
				
				err = writer.Close()
				if err != nil {
					b.Fatalf("Close failed: %v", err)
				}
				
				// Read
				reader, err := NewParquetReader(filePath)
				if err != nil {
					b.Fatalf("NewParquetReader failed: %v", err)
				}
				
				_, err = reader.Read()
				if err != nil {
					b.Fatalf("Read failed: %v", err)
				}
				
				reader.Close()
			}
		})
	}
}

func BenchmarkParquetFileSize(b *testing.B) {
	// This benchmark measures file size (run once)
	if b.N > 1 {
		b.N = 1
	}
	
	sizes := []int{1000, 10000, 100000}
	
	for _, size := range sizes {
		candles := generateBenchmarkCandles(size)
		
		tmpDir := b.TempDir()
		filePath := filepath.Join(tmpDir, "size_test.parquet")
		
		writer, err := NewParquetWriter(filePath)
		if err != nil {
			b.Fatalf("NewParquetWriter failed: %v", err)
		}
		
		err = writer.Write(candles)
		if err != nil {
			b.Fatalf("Write failed: %v", err)
		}
		
		err = writer.Close()
		if err != nil {
			b.Fatalf("Close failed: %v", err)
		}
		
		stat, err := os.Stat(filePath)
		if err != nil {
			b.Fatalf("Stat failed: %v", err)
		}
		
		bytesPerCandle := float64(stat.Size()) / float64(size)
		b.Logf("Size=%d candles: file=%d bytes, per_candle=%.2f bytes", 
			size, stat.Size(), bytesPerCandle)
	}
}
