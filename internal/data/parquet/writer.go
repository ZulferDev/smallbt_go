package parquet

import (
	"fmt"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// ParquetWriter writes OHLCV candle data to Parquet files.
type ParquetWriter struct {
	path       string
	fileWriter source.ParquetFile
	writer     *writer.ParquetWriter
}

// NewParquetWriter creates a new ParquetWriter for the given file path.
func NewParquetWriter(path string) (*ParquetWriter, error) {
	fileWriter, err := local.NewLocalFileWriter(path)
	if err != nil {
		return nil, fmt.Errorf("create file writer: %w", err)
	}

	pw, err := writer.NewParquetWriter(fileWriter, new(CandleParquet), 4)
	if err != nil {
		fileWriter.Close()
		return nil, fmt.Errorf("create parquet writer: %w", err)
	}

	// Set compression to SNAPPY
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	return &ParquetWriter{
		path:       path,
		fileWriter: fileWriter,
		writer:     pw,
	}, nil
}

// Write writes a slice of candles to the Parquet file.
// All candles are buffered and written when Close() is called.
func (w *ParquetWriter) Write(candles []*market.Candle) error {
	if w.writer == nil {
		return fmt.Errorf("writer is closed")
	}

	for _, candle := range candles {
		if candle == nil {
			return fmt.Errorf("cannot write nil candle")
		}

		parquetCandle := FromMarketCandle(candle)
		if err := w.writer.Write(*parquetCandle); err != nil {
			return fmt.Errorf("write candle: %w", err)
		}
	}

	return nil
}

// WriteOne writes a single candle to the Parquet file.
func (w *ParquetWriter) WriteOne(candle *market.Candle) error {
	if w.writer == nil {
		return fmt.Errorf("writer is closed")
	}

	if candle == nil {
		return fmt.Errorf("cannot write nil candle")
	}

	parquetCandle := FromMarketCandle(candle)
	if err := w.writer.Write(*parquetCandle); err != nil {
		return fmt.Errorf("write candle: %w", err)
	}

	return nil
}

// Close finalizes the Parquet file and releases resources.
// Must be called to ensure data is written to disk.
func (w *ParquetWriter) Close() error {
	if w.writer != nil {
		if err := w.writer.WriteStop(); err != nil {
			return fmt.Errorf("write stop: %w", err)
		}
		w.writer = nil
	}

	if w.fileWriter != nil {
		if err := w.fileWriter.Close(); err != nil {
			return fmt.Errorf("close file: %w", err)
		}
		w.fileWriter = nil
	}

	return nil
}
