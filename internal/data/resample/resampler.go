package resample

import (
	"fmt"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Resampler resamples candle data from one timeframe to another.
type Resampler interface {
	// Resample converts candles from source timeframe to target timeframe.
	// Returns error if target timeframe is smaller than source.
	Resample(candles []*market.Candle, targetTimeframe market.Timeframe) ([]*market.Candle, error)
}

// DefaultResampler implements Resampler with OHLCV aggregation.
type DefaultResampler struct {
	// SourceTimeframe is the timeframe of input candles (e.g., 1m)
	SourceTimeframe market.Timeframe
}

// NewDefaultResampler creates a new DefaultResampler.
func NewDefaultResampler(sourceTimeframe market.Timeframe) *DefaultResampler {
	return &DefaultResampler{
		SourceTimeframe: sourceTimeframe,
	}
}

// Resample converts candles to target timeframe using OHLCV aggregation rules:
// - Open: first candle's open in the period
// - High: highest high in the period
// - Low: lowest low in the period
// - Close: last candle's close in the period
// - Volume: sum of volumes in the period
// - Timestamp: start of the period (aligned to target timeframe)
func (r *DefaultResampler) Resample(candles []*market.Candle, targetTimeframe market.Timeframe) ([]*market.Candle, error) {
	if len(candles) == 0 {
		return []*market.Candle{}, nil
	}

	// Get durations
	sourceDuration, err := ParseTimeframe(r.SourceTimeframe)
	if err != nil {
		return nil, fmt.Errorf("parse source timeframe: %w", err)
	}

	targetDuration, err := ParseTimeframe(targetTimeframe)
	if err != nil {
		return nil, fmt.Errorf("parse target timeframe: %w", err)
	}

	// Validate: target must be >= source
	if targetDuration < sourceDuration {
		return nil, fmt.Errorf("target timeframe (%s) must be >= source timeframe (%s)", targetTimeframe, r.SourceTimeframe)
	}

	// If same timeframe, return copy
	if targetDuration == sourceDuration {
		result := make([]*market.Candle, len(candles))
		for i, c := range candles {
			candle := *c
			result[i] = &candle
		}
		return result, nil
	}

	// Resample
	var resampled []*market.Candle
	var currentBucket *candleBucket

	for _, candle := range candles {
		periodStart := AlignTimestamp(candle.Timestamp, targetDuration)

		// Start new bucket if needed
		if currentBucket == nil || !currentBucket.start.Equal(periodStart) {
			// Finalize previous bucket
			if currentBucket != nil {
				resampled = append(resampled, currentBucket.ToCandle())
			}

			// Start new bucket
			currentBucket = &candleBucket{
				start:  periodStart,
				open:   candle.Open,
				high:   candle.High,
				low:    candle.Low,
				close:  candle.Close,
				volume: candle.Volume,
			}
		} else {
			// Update bucket
			currentBucket.Update(candle)
		}
	}

	// Finalize last bucket
	if currentBucket != nil {
		resampled = append(resampled, currentBucket.ToCandle())
	}

	return resampled, nil
}

// candleBucket accumulates candles for a time period.
type candleBucket struct {
	start  time.Time
	open   float64
	high   float64
	low    float64
	close  float64
	volume float64
}

// Update updates the bucket with a new candle.
func (b *candleBucket) Update(c *market.Candle) {
	// Open: keep first (already set)
	// High: max
	if c.High > b.high {
		b.high = c.High
	}
	// Low: min
	if c.Low < b.low {
		b.low = c.Low
	}
	// Close: use latest
	b.close = c.Close
	// Volume: sum
	b.volume += c.Volume
}

// ToCandle converts the bucket to a Candle.
func (b *candleBucket) ToCandle() *market.Candle {
	return &market.Candle{
		Timestamp: b.start,
		Open:      b.open,
		High:      b.high,
		Low:       b.low,
		Close:     b.close,
		Volume:    b.volume,
	}
}

// ParseTimeframe converts a Timeframe to time.Duration.
func ParseTimeframe(tf market.Timeframe) (time.Duration, error) {
	switch tf {
	case market.Timeframe1m:
		return time.Minute, nil
	case market.Timeframe5m:
		return 5 * time.Minute, nil
	case market.Timeframe15m:
		return 15 * time.Minute, nil
	case market.Timeframe30m:
		return 30 * time.Minute, nil
	case market.Timeframe1h:
		return time.Hour, nil
	case market.Timeframe4h:
		return 4 * time.Hour, nil
	case market.Timeframe1d:
		return 24 * time.Hour, nil
	case market.Timeframe1w:
		return 7 * 24 * time.Hour, nil
	case market.Timeframe1mo:
		// Approximate: 30 days
		return 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported timeframe: %s", tf)
	}
}

// AlignTimestamp aligns a timestamp to the start of a period.
// For example, with 1h period, 14:23:45 becomes 14:00:00.
func AlignTimestamp(t time.Time, period time.Duration) time.Time {
	// For periods < 1 day: align to period boundary
	if period < 24*time.Hour {
		unix := t.Unix()
		periodSeconds := int64(period.Seconds())
		aligned := (unix / periodSeconds) * periodSeconds
		return time.Unix(aligned, 0).UTC()
	}

	// For 1 day: align to start of day (00:00:00)
	if period == 24*time.Hour {
		year, month, day := t.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	}

	// For 1 week: align to start of week (Monday 00:00:00)
	if period == 7*24*time.Hour {
		year, month, day := t.Date()
		weekday := t.Weekday()
		// Calculate days to subtract to get to Monday
		daysToMonday := int(weekday) - int(time.Monday)
		if daysToMonday < 0 {
			daysToMonday += 7
		}
		return time.Date(year, month, day-daysToMonday, 0, 0, 0, 0, time.UTC)
	}

	// For 1 month: align to start of month (1st day, 00:00:00)
	if period == 30*24*time.Hour {
		year, month, _ := t.Date()
		return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	}

	// Fallback: simple division-based alignment
	unix := t.Unix()
	periodSeconds := int64(period.Seconds())
	aligned := (unix / periodSeconds) * periodSeconds
	return time.Unix(aligned, 0).UTC()
}
