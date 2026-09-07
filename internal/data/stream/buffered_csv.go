package stream

import (
	"io"

	"github.com/ZulferDev/smallbt_go/internal/data/csv"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// BufferedCSVReader implements ChunkedFeed for CSV files.
// Reads data in chunks to enable memory-efficient streaming.
type BufferedCSVReader struct {
	feed      *csv.CSVFeed
	config    BufferConfig
	buffer    []*market.Candle
	allData   []*market.Candle
	dataIndex int
	loaded    bool
	closed    bool
}

// NewBufferedCSVReader creates a new buffered CSV reader.
func NewBufferedCSVReader(path string, csvConfig csv.CSVConfig, bufConfig BufferConfig) (*BufferedCSVReader, error) {
	if err := bufConfig.Validate(); err != nil {
		return nil, err
	}

	feed, err := csv.NewCSVFeed(path, csvConfig)
	if err != nil {
		return nil, err
	}

	return &BufferedCSVReader{
		feed:   feed,
		config: bufConfig,
		buffer: make([]*market.Candle, 0, bufConfig.ChunkSize),
		loaded: false,
		closed: false,
	}, nil
}

// Next returns the next candle from the stream.
func (r *BufferedCSVReader) Next() (*market.Candle, error) {
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
func (r *BufferedCSVReader) HasNext() bool {
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
func (r *BufferedCSVReader) NextChunk() ([]*market.Candle, error) {
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
func (r *BufferedCSVReader) SetChunkSize(size int) error {
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
func (r *BufferedCSVReader) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	r.buffer = nil
	r.allData = nil

	// CSVFeed doesn't have Close method, so nothing to do
	return nil
}

// Reset resets the stream to the beginning.
func (r *BufferedCSVReader) Reset() error {
	if r.closed {
		return io.ErrClosedPipe
	}

	r.dataIndex = 0

	// Reset feed position if supported
	if r.feed != nil {
		r.feed.Reset()
	}

	return nil
}

// loadAll loads all data from the CSV feed.
// This is a temporary implementation - ideally would use true streaming.
func (r *BufferedCSVReader) loadAll() error {
	if r.loaded {
		return nil
	}

	var candles []*market.Candle
	for {
		data, err := r.feed.Next()
		if err != nil {
			break
		}

		// Convert MarketData.Candles to []*market.Candle
		for i := range data.Candles {
			candles = append(candles, &data.Candles[i])
		}
	}

	r.allData = candles
	r.dataIndex = 0
	r.loaded = true

	return nil
}

// Stats returns statistics about the buffered reader.
func (r *BufferedCSVReader) Stats() BufferedReaderStats {
	return BufferedReaderStats{
		TotalCandles: len(r.allData),
		CurrentIndex: r.dataIndex,
		ChunkSize:    r.config.ChunkSize,
		ChunksRead:   r.dataIndex / r.config.ChunkSize,
		MemoryUsedMB: float64(len(r.allData)*sizeOfCandle) / 1024 / 1024,
	}
}
