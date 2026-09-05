package runtime

import "time"

// TimeProvider abstracts time (historical vs real-time)
// This allows backtest to use simulated time while paper/live use system time
type TimeProvider interface {
	// Now returns the current time
	Now() time.Time
	
	// Sleep pauses execution for duration d
	Sleep(d time.Duration)
	
	// After waits for duration d and then sends the current time on the returned channel
	After(d time.Duration) <-chan time.Time
}

// HistoricalTime uses simulated time for backtesting
// Time advances only when Advance() is called (controlled by backtest engine)
type HistoricalTime struct {
	current time.Time
}

// NewHistoricalTime creates a new historical time provider starting at t
func NewHistoricalTime(t time.Time) *HistoricalTime {
	return &HistoricalTime{current: t}
}

// Now returns the current simulated time
func (h *HistoricalTime) Now() time.Time {
	return h.current
}

// Sleep is a no-op in backtesting (time doesn't pass in real-time)
func (h *HistoricalTime) Sleep(d time.Duration) {
	// No-op: backtest doesn't wait in real-time
}

// After immediately returns a channel with the future time
// No actual waiting occurs in backtesting
func (h *HistoricalTime) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	ch <- h.current.Add(d)
	return ch
}

// Advance moves the simulated time forward to t
// This is called by the backtest engine as it processes historical candles
func (h *HistoricalTime) Advance(t time.Time) {
	h.current = t
}

// RealTime uses the system clock (for paper and live trading)
type RealTime struct{}

// NewRealTime creates a new real-time provider
func NewRealTime() *RealTime {
	return &RealTime{}
}

// Now returns the current system time
func (r *RealTime) Now() time.Time {
	return time.Now()
}

// Sleep pauses the goroutine for duration d
func (r *RealTime) Sleep(d time.Duration) {
	time.Sleep(d)
}

// After waits for duration d and then sends the current time
func (r *RealTime) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
