package stream

import (
	"io"

	"github.com/ZulferDev/smallbt_go/internal/data/parquet"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// BufferedParquetReader implements ChunkedFeed for Parquet files.
// Reads data in chunks to enable memory-efficient streaming.
type BufferedParquetReader struct {
	reader    *parquet.ParquetReader
	config    BufferConfig
	buffer    []*market.Candle
	position  int
	allData   []*market.Candle
	dataIndex int
	loaded    bool
	closed    bool
}

// NewBufferedParquetReader creates a new buffered Parquet reader.
func NewBufferedParquetReader(path string, config BufferConfig) (*BufferedParquetReader, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	reader, err := parquet.NewParquetReader(path)
	if err != nil {
		return nil, err
	}

	return &BufferedParquetReader{
		reader:   reader,
		config:   config,
		buffer:   make([]*market.Candle, 0, config.ChunkSize),
		position: 0,
		loaded:   false,
		closed:   false,
	}, nil
}

// Next returns the next candle from the stream.
func (r *BufferedParquetReader) Next() (*market.Candle, error) {
	if r.closed {
		return nil, io.ErrClosedPipe
	}

	// Load all data on first call
	if !r.loaded {
		if err := r.loadAll(); err != nil {
			return nil, err
		}
	}

	// Check if we've reached the end
	if r.dataIndex >= len(r.allData) {
		return nil, io.EOF
	}

	candle := r.allData[r.dataIndex]
	r.dataIndex++
	return candle, nil
}

// HasNext returns true if more data is available.
func (r *BufferedParquetReader) HasNext() bool {
	if r.closed {
		return false
	}

	if !r.loaded {
		// Need to load to know
		if err := r.loadAll(); err != nil {
			return false
		}
	}

	return r.dataIndex < len(r.allData)
}

// NextChunk returns the next chunk of candles.
func (r *BufferedParquetReader) NextChunk() ([]*market.Candle, error) {
	if r.closed {
		return nil, io.ErrClosedPipe
	}

	// Load all data on first call
	if !r.loaded {
		if err := r.loadAll(); err != nil {
			return nil, err
		}
	}

	// Check if we've reached the end
	if r.dataIndex >= len(r.allData) {
		return nil, io.EOF
	}

	// Calculate chunk size
	remaining := len(r.allData) - r.dataIndex
	chunkSize := r.config.ChunkSize
	if remaining < chunkSize {
		chunkSize = remaining
	}

	// Return chunk
	chunk := r.allData[r.dataIndex : r.dataIndex+chunkSize]
	r.dataIndex += chunkSize

	return chunk, nil
}

// SetChunkSize sets the chunk size for reading.
func (r *BufferedParquetReader) SetChunkSize(size int) error {
	if r.loaded {
		return io.ErrUnexpectedEOF // Already started reading
	}

	if size <= 0 || size > MaxChunkSize {
		return ErrInvalidChunkSize
	}

	r.config.ChunkSize = size
	r.buffer = make([]*market.Candle, 0, size)
	return nil
}

// Close releases resources held by the reader.
func (r *BufferedParquetReader) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	r.buffer = nil
	r.allData = nil

	if r.reader != nil {
		return r.reader.Close()
	}

	return nil
}

// Reset resets the stream to the beginning.
func (r *BufferedParquetReader) Reset() error {
	if r.closed {
		return io.ErrClosedPipe
	}

	r.dataIndex = 0
	return nil
}

// loadAll loads all data from the Parquet file.
// This is a temporary implementation - ideally would use true streaming.
func (r *BufferedParquetReader) loadAll() error {
	if r.loaded {
		return nil
	}

	candles, err := r.reader.Read()
	if err != nil {
		return err
	}

	r.allData = candles
	r.dataIndex = 0
	r.loaded = true

	return nil
}

// Stats returns statistics about the buffered reader.
func (r *BufferedParquetReader) Stats() BufferedReaderStats {
	return BufferedReaderStats{
		TotalCandles:  len(r.allData),
		CurrentIndex:  r.dataIndex,
		ChunkSize:     r.config.ChunkSize,
		ChunksRead:    r.dataIndex / r.config.ChunkSize,
		MemoryUsedMB:  float64(len(r.allData)*sizeOfCandle) / 1024 / 1024,
	}
}

// BufferedReaderStats contains statistics about buffered reader.
type BufferedReaderStats struct {
	TotalCandles  int
	CurrentIndex  int
	ChunkSize     int
	ChunksRead    int
	MemoryUsedMB  float64
}

// Approximate size of one candle in bytes
const sizeOfCandle = 64
