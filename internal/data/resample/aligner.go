package resample

import (
	"fmt"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Aligner aligns candle data across multiple symbols to common timestamps.
type Aligner interface {
	// Align synchronizes candles from multiple symbols to a common timeline.
	// Returns aligned candles indexed by symbol, using forward-fill for missing data.
	Align(symbolData map[string][]*market.Candle) (map[string][]*market.Candle, error)
}

// DefaultAligner implements Aligner with forward-fill strategy.
type DefaultAligner struct {
	// FillStrategy determines how to handle missing data
	FillStrategy FillStrategy
}

// FillStrategy defines how to fill missing candles.
type FillStrategy int

const (
	// FillStrategyForward uses the last known candle for missing timestamps
	FillStrategyForward FillStrategy = iota
	// FillStrategyDrop skips timestamps where any symbol has missing data
	FillStrategyDrop
	// FillStrategyNone leaves gaps (returns error if misaligned)
	FillStrategyNone
)

// NewDefaultAligner creates a new DefaultAligner with forward-fill strategy.
func NewDefaultAligner() *DefaultAligner {
	return &DefaultAligner{
		FillStrategy: FillStrategyForward,
	}
}

// Align synchronizes candles from multiple symbols to a common timeline.
// 
// Algorithm:
// 1. Collect all unique timestamps across all symbols
// 2. Sort timestamps chronologically
// 3. For each timestamp, get candle from each symbol:
//    - If candle exists at timestamp: use it
//    - If missing + FillStrategyForward: use last known candle
//    - If missing + FillStrategyDrop: skip this timestamp
//    - If missing + FillStrategyNone: return error
// 4. Return aligned data
func (a *DefaultAligner) Align(symbolData map[string][]*market.Candle) (map[string][]*market.Candle, error) {
	if len(symbolData) == 0 {
		return map[string][]*market.Candle{}, nil
	}

	// Collect all unique timestamps
	timestampSet := make(map[time.Time]bool)
	for _, candles := range symbolData {
		for _, candle := range candles {
			timestampSet[candle.Timestamp] = true
		}
	}

	// Convert to sorted slice
	timestamps := make([]time.Time, 0, len(timestampSet))
	for ts := range timestampSet {
		timestamps = append(timestamps, ts)
	}
	sortTimestamps(timestamps)

	// Build indices for fast lookup: symbol -> (timestamp -> candle)
	indices := make(map[string]map[time.Time]*market.Candle)
	for symbol, candles := range symbolData {
		index := make(map[time.Time]*market.Candle)
		for _, candle := range candles {
			index[candle.Timestamp] = candle
		}
		indices[symbol] = index
	}

	// Align data
	result := make(map[string][]*market.Candle)
	lastKnown := make(map[string]*market.Candle) // For forward-fill

	for _, ts := range timestamps {
		skip := false
		rowCandles := make(map[string]*market.Candle)

		// Get candle for each symbol at this timestamp
		for symbol := range symbolData {
			candle, exists := indices[symbol][ts]

			if exists {
				rowCandles[symbol] = candle
				lastKnown[symbol] = candle
			} else {
				// Missing data
				switch a.FillStrategy {
				case FillStrategyForward:
					if last, ok := lastKnown[symbol]; ok {
						// Forward-fill: use last known candle with updated timestamp
						filled := *last
						filled.Timestamp = ts
						rowCandles[symbol] = &filled
					} else {
						// No prior data for forward-fill
						skip = true
						break
					}
				case FillStrategyDrop:
					skip = true
					break
				case FillStrategyNone:
					return nil, fmt.Errorf("missing data at %v for symbol %s (FillStrategyNone)", ts, symbol)
				}
			}
		}

		// Add row if not skipped
		if !skip {
			for symbol, candle := range rowCandles {
				result[symbol] = append(result[symbol], candle)
			}
		}
	}

	return result, nil
}

// AlignToReference aligns symbols to a reference symbol's timestamps.
// This is useful when you want one symbol to drive the timeline.
func (a *DefaultAligner) AlignToReference(
	referenceSymbol string,
	symbolData map[string][]*market.Candle,
) (map[string][]*market.Candle, error) {
	refCandles, ok := symbolData[referenceSymbol]
	if !ok {
		return nil, fmt.Errorf("reference symbol %s not found", referenceSymbol)
	}

	if len(refCandles) == 0 {
		return map[string][]*market.Candle{}, nil
	}

	// Build indices
	indices := make(map[string]map[time.Time]*market.Candle)
	for symbol, candles := range symbolData {
		if symbol == referenceSymbol {
			continue // Skip reference, we'll use it directly
		}
		index := make(map[time.Time]*market.Candle)
		for _, candle := range candles {
			index[candle.Timestamp] = candle
		}
		indices[symbol] = index
	}

	// Align to reference timestamps
	result := make(map[string][]*market.Candle)
	result[referenceSymbol] = refCandles // Reference always included

	lastKnown := make(map[string]*market.Candle)

	for _, refCandle := range refCandles {
		ts := refCandle.Timestamp

		for symbol := range symbolData {
			if symbol == referenceSymbol {
				continue
			}

			candle, exists := indices[symbol][ts]

			if exists {
				result[symbol] = append(result[symbol], candle)
				lastKnown[symbol] = candle
			} else {
				// Missing data
				switch a.FillStrategy {
				case FillStrategyForward:
					if last, ok := lastKnown[symbol]; ok {
						filled := *last
						filled.Timestamp = ts
						result[symbol] = append(result[symbol], &filled)
					} else {
						// No prior data - this row incomplete
						return nil, fmt.Errorf("no prior data for symbol %s at %v", symbol, ts)
					}
				case FillStrategyDrop:
					// Cannot drop with reference-based alignment
					return nil, fmt.Errorf("FillStrategyDrop not supported for reference-based alignment")
				case FillStrategyNone:
					return nil, fmt.Errorf("missing data at %v for symbol %s", ts, symbol)
				}
			}
		}
	}

	return result, nil
}

// sortTimestamps sorts timestamps in ascending order (in-place).
func sortTimestamps(timestamps []time.Time) {
	// Simple insertion sort (efficient for small-medium datasets)
	for i := 1; i < len(timestamps); i++ {
		key := timestamps[i]
		j := i - 1
		for j >= 0 && timestamps[j].After(key) {
			timestamps[j+1] = timestamps[j]
			j--
		}
		timestamps[j+1] = key
	}
}

// GetCommonTimeRange returns the time range where all symbols have data.
// Useful for finding the intersection period across multiple symbols.
func GetCommonTimeRange(symbolData map[string][]*market.Candle) (start, end time.Time, hasData bool) {
	if len(symbolData) == 0 {
		return time.Time{}, time.Time{}, false
	}

	var latestStart time.Time
	var earliestEnd time.Time
	first := true

	for _, candles := range symbolData {
		if len(candles) == 0 {
			return time.Time{}, time.Time{}, false
		}

		symbolStart := candles[0].Timestamp
		symbolEnd := candles[len(candles)-1].Timestamp

		if first {
			latestStart = symbolStart
			earliestEnd = symbolEnd
			first = false
		} else {
			if symbolStart.After(latestStart) {
				latestStart = symbolStart
			}
			if symbolEnd.Before(earliestEnd) {
				earliestEnd = symbolEnd
			}
		}
	}

	if latestStart.After(earliestEnd) {
		return time.Time{}, time.Time{}, false
	}

	return latestStart, earliestEnd, true
}

// FilterByTimeRange filters candles to only include those within the time range.
func FilterByTimeRange(candles []*market.Candle, start, end time.Time) []*market.Candle {
	result := make([]*market.Candle, 0)
	for _, candle := range candles {
		if !candle.Timestamp.Before(start) && !candle.Timestamp.After(end) {
			result = append(result, candle)
		}
	}
	return result
}
