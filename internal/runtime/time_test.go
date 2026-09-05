package runtime

import (
	"testing"
	"time"
)

func TestHistoricalTime(t *testing.T) {
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ht := NewHistoricalTime(start)

	// Test Now()
	if got := ht.Now(); !got.Equal(start) {
		t.Errorf("Now() = %v, want %v", got, start)
	}

	// Test Advance()
	newTime := start.Add(24 * time.Hour)
	ht.Advance(newTime)
	if got := ht.Now(); !got.Equal(newTime) {
		t.Errorf("After Advance(), Now() = %v, want %v", got, newTime)
	}

	// Test Sleep() - should be no-op
	before := ht.Now()
	ht.Sleep(1 * time.Hour)
	after := ht.Now()
	if !before.Equal(after) {
		t.Errorf("Sleep() changed time: before=%v, after=%v", before, after)
	}

	// Test After()
	ch := ht.After(2 * time.Hour)
	select {
	case got := <-ch:
		expected := newTime.Add(2 * time.Hour)
		if !got.Equal(expected) {
			t.Errorf("After() sent %v, want %v", got, expected)
		}
	default:
		t.Error("After() channel should have immediate value")
	}
}

func TestRealTime(t *testing.T) {
	rt := NewRealTime()

	// Test Now() returns current system time (approximately)
	before := time.Now()
	got := rt.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("Now() = %v, expected between %v and %v", got, before, after)
	}

	// Test Sleep() actually waits
	start := time.Now()
	rt.Sleep(50 * time.Millisecond)
	elapsed := time.Since(start)
	if elapsed < 50*time.Millisecond {
		t.Errorf("Sleep(50ms) only waited %v", elapsed)
	}

	// Test After() waits
	ch := rt.After(10 * time.Millisecond)
	select {
	case <-ch:
		// Expected - timer fired
	case <-time.After(100 * time.Millisecond):
		t.Error("After() did not fire within timeout")
	}
}

func TestTimeProvider_Interface(t *testing.T) {
	// Verify both implementations satisfy TimeProvider interface
	var _ TimeProvider = (*HistoricalTime)(nil)
	var _ TimeProvider = (*RealTime)(nil)
}
