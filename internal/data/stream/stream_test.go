package stream

import (
	"io"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestDefaultBufferConfig(t *testing.T) {
	config := DefaultBufferConfig()
	
	if config.ChunkSize != 10000 {
		t.Errorf("Expected default chunk size 10000, got %d", config.ChunkSize)
	}
	
	if config.Prefetch != false {
		t.Error("Expected default prefetch to be false")
	}
}

func TestBufferConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BufferConfig
		wantErr bool
	}{
		{
			name:    "valid default",
			config:  DefaultBufferConfig(),
			wantErr: false,
		},
		{
			name:    "valid custom",
			config:  BufferConfig{ChunkSize: 5000},
			wantErr: false,
		},
		{
			name:    "invalid zero",
			config:  BufferConfig{ChunkSize: 0},
			wantErr: true,
		},
		{
			name:    "invalid negative",
			config:  BufferConfig{ChunkSize: -100},
			wantErr: true,
		},
		{
			name:    "invalid too large",
			config:  BufferConfig{ChunkSize: MaxChunkSize + 1},
			wantErr: true,
		},
		{
			name:    "valid min",
			config:  BufferConfig{ChunkSize: MinChunkSize},
			wantErr: false,
		},
		{
			name:    "valid max",
			config:  BufferConfig{ChunkSize: MaxChunkSize},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if DefaultChunkSize != 10000 {
		t.Errorf("Expected DefaultChunkSize 10000, got %d", DefaultChunkSize)
	}
	
	if MinChunkSize != 100 {
		t.Errorf("Expected MinChunkSize 100, got %d", MinChunkSize)
	}
	
	if MaxChunkSize != 1000000 {
		t.Errorf("Expected MaxChunkSize 1000000, got %d", MaxChunkSize)
	}
}

func TestErrors(t *testing.T) {
	// Verify error constants are defined
	if ErrInvalidChunkSize == nil {
		t.Error("Expected ErrInvalidChunkSize to be defined")
	}
	
	if ErrChunkSizeTooLarge == nil {
		t.Error("Expected ErrChunkSizeTooLarge to be defined")
	}
	
	if ErrResetNotSupported == nil {
		t.Error("Expected ErrResetNotSupported to be defined")
	}
}

func TestBufferConfig_ValidateBoundaries(t *testing.T) {
	// Test boundary conditions
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"below_min", MinChunkSize - 1, false}, // Still valid, just small
		{"at_min", MinChunkSize, false},
		{"above_min", MinChunkSize + 1, false},
		{"at_max", MaxChunkSize, false},
		{"above_max", MaxChunkSize + 1, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := BufferConfig{ChunkSize: tt.size}
			err := config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("size=%d: error = %v, wantErr %v", tt.size, err, tt.wantErr)
			}
		})
	}
}

func TestBufferedReaderStats(t *testing.T) {
	stats := BufferedReaderStats{
		TotalCandles:  10000,
		CurrentIndex:  5000,
		ChunkSize:     1000,
		ChunksRead:    5,
		MemoryUsedMB:  0.625,
	}
	
	if stats.TotalCandles != 10000 {
		t.Errorf("Expected TotalCandles 10000, got %d", stats.TotalCandles)
	}
	
	if stats.CurrentIndex != 5000 {
		t.Errorf("Expected CurrentIndex 5000, got %d", stats.CurrentIndex)
	}
	
	if stats.ChunksRead != 5 {
		t.Errorf("Expected ChunksRead 5, got %d", stats.ChunksRead)
	}
}

func TestStreamFeed_InterfaceCompliance(t *testing.T) {
	// This test verifies that our interfaces are well-defined
	// Actual implementations will be tested in their own test files
	
	var _ StreamFeed = (*mockStreamFeed)(nil)
	var _ ChunkedFeed = (*mockChunkedFeed)(nil)
}

// Mock implementations for interface testing
type mockStreamFeed struct{}

func (m *mockStreamFeed) Next() (*market.Candle, error)   { return nil, io.EOF }
func (m *mockStreamFeed) HasNext() bool                   { return false }
func (m *mockStreamFeed) Close() error                    { return nil }
func (m *mockStreamFeed) Reset() error                    { return nil }

type mockChunkedFeed struct {
	mockStreamFeed
}

func (m *mockChunkedFeed) NextChunk() ([]*market.Candle, error) { return nil, io.EOF }
func (m *mockChunkedFeed) SetChunkSize(size int) error          { return nil }
