package feed

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestCandleBuffer_NewCandleBuffer(t *testing.T) {
	buffer := NewCandleBuffer(100)
	
	if buffer == nil {
		t.Fatal("expected buffer to be created")
	}
	
	if buffer.Len() != 0 {
		t.Errorf("expected empty buffer, got length %d", buffer.Len())
	}
	
	if buffer.Cap() != 100 {
		t.Errorf("expected capacity 100, got %d", buffer.Cap())
	}
}

func TestCandleBuffer_DefaultSize(t *testing.T) {
	buffer := NewCandleBuffer(0)
	
	if buffer.Cap() != 1000 {
		t.Errorf("expected default capacity 1000, got %d", buffer.Cap())
	}
	
	buffer2 := NewCandleBuffer(-1)
	if buffer2.Cap() != 1000 {
		t.Errorf("expected default capacity 1000 for negative size, got %d", buffer2.Cap())
	}
}

func TestCandleBuffer_Push(t *testing.T) {
	buffer := NewCandleBuffer(10)
	
	candle := &market.Candle{
		Timestamp: time.Now(),
		Open:      100,
		High:      105,
		Low:       95,
		Close:     102,
		Volume:    1000,
	}
	
	err := buffer.Push(candle)
	if err != nil {
		t.Errorf("Push() failed: %v", err)
	}
	
	if buffer.Len() != 1 {
		t.Errorf("expected length 1, got %d", buffer.Len())
	}
}

func TestCandleBuffer_PushNil(t *testing.T) {
	buffer := NewCandleBuffer(10)
	
	err := buffer.Push(nil)
	if err == nil {
		t.Error("expected error when pushing nil candle")
	}
}

func TestCandleBuffer_PushMultiple(t *testing.T) {
	buffer := NewCandleBuffer(10)
	
	for i := 0; i < 5; i++ {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Open:      100.0,
			High:      105.0,
			Low:       95.0,
			Close:     102.0,
			Volume:    1000.0,
		}
		
		err := buffer.Push(candle)
		if err != nil {
			t.Errorf("Push() failed at index %d: %v", i, err)
		}
	}
	
	if buffer.Len() != 5 {
		t.Errorf("expected length 5, got %d", buffer.Len())
	}
}

func TestCandleBuffer_Overflow(t *testing.T) {
	buffer := NewCandleBuffer(3)
	
	// Fill buffer to capacity
	for i := 0; i < 3; i++ {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Open:      100.0,
			High:      105.0,
			Low:       95.0,
			Close:     102.0,
			Volume:    1000.0,
		}
		
		err := buffer.Push(candle)
		if err != nil {
			t.Fatalf("Push() failed: %v", err)
		}
	}
	
	if buffer.Len() != 3 {
		t.Errorf("expected buffer length 3, got %d", buffer.Len())
	}
	
	// Push beyond capacity should go to overflow channel
	candle := &market.Candle{
		Timestamp: time.Now(),
		Open:      100.0,
		High:      105.0,
		Low:       95.0,
		Close:     102.0,
		Volume:    1000.0,
	}
	
	err := buffer.Push(candle)
	if err != nil {
		t.Errorf("Push() to overflow failed: %v", err)
	}
	
	// Check overflow channel
	select {
	case c := <-buffer.Overflow():
		if c == nil {
			t.Error("expected candle from overflow channel")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected candle in overflow channel")
	}
}

func TestCandleBuffer_Drain(t *testing.T) {
	buffer := NewCandleBuffer(10)
	
	// Push some candles
	for i := 0; i < 5; i++ {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Open:      float64(100 + i),
			High:      float64(105 + i),
			Low:       float64(95 + i),
			Close:     float64(102 + i),
			Volume:    1000.0,
		}
		buffer.Push(candle)
	}
	
	// Drain
	candles := buffer.Drain()
	
	if len(candles) != 5 {
		t.Errorf("expected 5 candles, got %d", len(candles))
	}
	
	if buffer.Len() != 0 {
		t.Errorf("expected buffer to be empty after drain, got length %d", buffer.Len())
	}
	
	// Verify candle values
	for i, candle := range candles {
		if candle.Open != float64(100+i) {
			t.Errorf("candle %d: expected open %f, got %f", i, float64(100+i), candle.Open)
		}
	}
}

func TestCandleBuffer_DrainEmpty(t *testing.T) {
	buffer := NewCandleBuffer(10)
	
	candles := buffer.Drain()
	
	if candles != nil {
		t.Errorf("expected nil from empty drain, got %d candles", len(candles))
	}
}

func TestCandleBuffer_Clear(t *testing.T) {
	buffer := NewCandleBuffer(10)
	
	// Push some candles
	for i := 0; i < 5; i++ {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Open:      100.0,
			High:      105.0,
			Low:       95.0,
			Close:     102.0,
			Volume:    1000.0,
		}
		buffer.Push(candle)
	}
	
	buffer.Clear()
	
	if buffer.Len() != 0 {
		t.Errorf("expected empty buffer after clear, got length %d", buffer.Len())
	}
}

func TestCandleBuffer_Concurrent(t *testing.T) {
	buffer := NewCandleBuffer(1000)
	
	done := make(chan bool)
	
	// Multiple goroutines pushing
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				candle := &market.Candle{
					Timestamp: time.Now(),
					Open:      100.0,
					High:      105.0,
					Low:       95.0,
					Close:     102.0,
					Volume:    1000.0,
				}
				buffer.Push(candle)
			}
			done <- true
		}()
	}
	
	// Wait for all pushes
	for i := 0; i < 10; i++ {
		<-done
	}
	
	if buffer.Len() != 1000 {
		t.Errorf("expected 1000 candles, got %d", buffer.Len())
	}
}
