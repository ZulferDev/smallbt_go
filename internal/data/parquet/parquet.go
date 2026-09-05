package parquet

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Reader reads market data from Parquet files.
type Reader struct {
	path string
}

// NewReader creates a new Parquet reader.
func NewReader(path string) *Reader {
	return &Reader{path: path}
}

// Read reads candles from a Parquet file.
func (r *Reader) Read() ([]market.Candle, error) {
	return nil, fmt.Errorf("parquet reader not implemented yet")
}
