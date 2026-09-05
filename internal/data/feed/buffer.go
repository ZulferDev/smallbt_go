package feed

import (
	"fmt"
	"sync"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CandleBuffer provides thread-safe buffering for candles.
type CandleBuffer struct {
	mu       sync.RWMutex
	candles  []*market.Candle
	maxSize  int
	overflow chan *market.Candle
}

// NewCandleBuffer creates a new candle buffer.
func NewCandleBuffer(maxSize int) *CandleBuffer {
	if maxSize <= 0 {
		maxSize = 1000
	}
	
	return &CandleBuffer{
		candles:  make([]*market.Candle, 0, maxSize),
		maxSize:  maxSize,
		overflow: make(chan *market.Candle, 100),
	}
}

// Push adds a candle to the buffer.
// Returns error if buffer is full and overflow channel is also full.
func (b *CandleBuffer) Push(candle *market.Candle) error {
	if candle == nil {
		return fmt.Errorf("cannot push nil candle")
	}
	
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if len(b.candles) < b.maxSize {
		b.candles = append(b.candles, candle)
		return nil
	}
	
	// Buffer is full, try overflow channel
	select {
	case b.overflow <- candle:
		return nil
	default:
		return fmt.Errorf("buffer overflow: max size %d reached", b.maxSize)
	}
}

// Drain removes and returns all candles from the buffer.
func (b *CandleBuffer) Drain() []*market.Candle {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if len(b.candles) == 0 {
		return nil
	}
	
	result := make([]*market.Candle, len(b.candles))
	copy(result, b.candles)
	b.candles = b.candles[:0]
	
	return result
}

// Len returns the current buffer size.
func (b *CandleBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.candles)
}

// Cap returns the buffer capacity.
func (b *CandleBuffer) Cap() int {
	return b.maxSize
}

// Clear empties the buffer.
func (b *CandleBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.candles = b.candles[:0]
}

// Overflow returns the overflow channel.
func (b *CandleBuffer) Overflow() <-chan *market.Candle {
	return b.overflow
}
