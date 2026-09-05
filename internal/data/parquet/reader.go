package parquet

import (
	"fmt"
	"time"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/source"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// ParquetReader reads OHLCV candle data from Parquet files.
type ParquetReader struct {
	path       string
	fileReader source.ParquetFile
	reader     *reader.ParquetReader
}

// NewParquetReader creates a new ParquetReader for the given file path.
func NewParquetReader(path string) (*ParquetReader, error) {
	fileReader, err := local.NewLocalFileReader(path)
	if err != nil {
		return nil, fmt.Errorf("open parquet file: %w", err)
	}

	pr, err := reader.NewParquetReader(fileReader, new(CandleParquet), 4)
	if err != nil {
		fileReader.Close()
		return nil, fmt.Errorf("create parquet reader: %w", err)
	}

	return &ParquetReader{
		path:       path,
		fileReader: fileReader,
		reader:     pr,
	}, nil
}

// Read reads all candles from the Parquet file.
func (r *ParquetReader) Read() ([]*market.Candle, error) {
	if r.reader == nil {
		return nil, fmt.Errorf("reader is closed")
	}

	numRows := int(r.reader.GetNumRows())
	if numRows == 0 {
		return []*market.Candle{}, nil
	}

	candles := make([]*market.Candle, 0, numRows)
	parquetCandles := make([]CandleParquet, numRows)

	if err := r.reader.Read(&parquetCandles); err != nil {
		return nil, fmt.Errorf("read parquet data: %w", err)
	}

	for i := range parquetCandles {
		candles = append(candles, parquetCandles[i].ToMarketCandle())
	}

	return candles, nil
}

// ReadRange reads candles within a specific time range.
// Returns candles where start <= timestamp <= end.
func (r *ParquetReader) ReadRange(start, end time.Time) ([]*market.Candle, error) {
	allCandles, err := r.Read()
	if err != nil {
		return nil, err
	}

	if len(allCandles) == 0 {
		return []*market.Candle{}, nil
	}

	startMillis := start.UnixMilli()
	endMillis := end.UnixMilli()

	filtered := make([]*market.Candle, 0, len(allCandles))
	for _, candle := range allCandles {
		candleMillis := candle.Timestamp.UnixMilli()
		if candleMillis >= startMillis && candleMillis <= endMillis {
			filtered = append(filtered, candle)
		}
	}

	return filtered, nil
}

// Close closes the Parquet reader and releases resources.
func (r *ParquetReader) Close() error {
	if r.reader != nil {
		r.reader.ReadStop()
		r.reader = nil
	}

	if r.fileReader != nil {
		if err := r.fileReader.Close(); err != nil {
			return fmt.Errorf("close file: %w", err)
		}
		r.fileReader = nil
	}

	return nil
}
