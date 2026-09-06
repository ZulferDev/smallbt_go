package stream

import (
	"io"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// StreamFeed represents a streaming data feed that provides candles one at a time.
// Implementations should be memory-efficient for large datasets.
type StreamFeed interface {
	// Next returns the next candle from the stream.
	// Returns io.EOF when no more data is available.
	Next() (*market.Candle, error)

	// HasNext returns true if more data is available.
	// This is a non-blocking check that doesn't advance the stream.
	HasNext() bool

	// Close releases any resources held by the stream.
	Close() error

	// Reset resets the stream to the beginning if supported.
	// Returns an error if reset is not supported.
	Reset() error
}

// ChunkedFeed represents a feed that reads data in chunks for efficiency.
type ChunkedFeed interface {
	StreamFeed

	// NextChunk returns the next chunk of candles.
	// The chunk size is implementation-defined.
	// Returns io.EOF when no more data is available.
	NextChunk() ([]*market.Candle, error)

	// SetChunkSize sets the chunk size for reading.
	// Must be called before any reading operations.
	SetChunkSize(size int) error
}

// BufferConfig contains configuration for buffered readers.
type BufferConfig struct {
	// ChunkSize is the number of candles to read per chunk.
	// Default: 10000
	ChunkSize int

	// Prefetch enables prefetching of next chunk while processing current.
	// Default: false
	Prefetch bool
}

// DefaultBufferConfig returns a default buffer configuration.
func DefaultBufferConfig() BufferConfig {
	return BufferConfig{
		ChunkSize: 10000,
		Prefetch:  false,
	}
}

// Validate validates the buffer configuration.
func (c BufferConfig) Validate() error {
	if c.ChunkSize <= 0 {
		return ErrInvalidChunkSize
	}
	if c.ChunkSize > MaxChunkSize {
		return ErrChunkSizeTooLarge
	}
	return nil
}

// Constants
const (
	// DefaultChunkSize is the default number of candles per chunk.
	DefaultChunkSize = 10000

	// MinChunkSize is the minimum allowed chunk size.
	MinChunkSize = 100

	// MaxChunkSize is the maximum allowed chunk size.
	MaxChunkSize = 1000000
)

// Errors
var (
	// ErrInvalidChunkSize is returned when chunk size is invalid.
	ErrInvalidChunkSize = io.ErrUnexpectedEOF

	// ErrChunkSizeTooLarge is returned when chunk size exceeds limit.
	ErrChunkSizeTooLarge = io.ErrUnexpectedEOF

	// ErrResetNotSupported is returned when Reset is not supported.
	ErrResetNotSupported = io.ErrNoProgress
)
