package resample

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestNewDefaultResampler(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1m)
	if r == nil {
		t.Fatal("expected non-nil resampler")
	}
	if r.SourceTimeframe != market.Timeframe1m {
		t.Errorf("expected source timeframe 1m, got %s", r.SourceTimeframe)
	}
}

func TestParseTimeframe(t *testing.T) {
	tests := []struct {
		name     string
		tf       market.Timeframe
		expected time.Duration
		wantErr  bool
	}{
		{"1m", market.Timeframe1m, time.Minute, false},
		{"5m", market.Timeframe5m, 5 * time.Minute, false},
		{"15m", market.Timeframe15m, 15 * time.Minute, false},
		{"30m", market.Timeframe30m, 30 * time.Minute, false},
		{"1h", market.Timeframe1h, time.Hour, false},
		{"4h", market.Timeframe4h, 4 * time.Hour, false},
		{"1d", market.Timeframe1d, 24 * time.Hour, false},
		{"1w", market.Timeframe1w, 7 * 24 * time.Hour, false},
		{"1mo", market.Timeframe1mo, 30 * 24 * time.Hour, false},
		{"invalid", market.Timeframe("invalid"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimeframe(tt.tf)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %s", tt.tf)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestAlignTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		period   time.Duration
		expected time.Time
	}{
		{
			name:     "1m alignment",
			time:     time.Date(2024, 1, 1, 14, 23, 45, 0, time.UTC),
			period:   time.Minute,
			expected: time.Date(2024, 1, 1, 14, 23, 0, 0, time.UTC),
		},
		{
			name:     "5m alignment",
			time:     time.Date(2024, 1, 1, 14, 23, 45, 0, time.UTC),
			period:   5 * time.Minute,
			expected: time.Date(2024, 1, 1, 14, 20, 0, 0, time.UTC),
		},
		{
			name:     "15m alignment",
			time:     time.Date(2024, 1, 1, 14, 23, 45, 0, time.UTC),
			period:   15 * time.Minute,
			expected: time.Date(2024, 1, 1, 14, 15, 0, 0, time.UTC),
		},
		{
			name:     "1h alignment",
			time:     time.Date(2024, 1, 1, 14, 23, 45, 0, time.UTC),
			period:   time.Hour,
			expected: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		},
		{
			name:     "4h alignment",
			time:     time.Date(2024, 1, 1, 14, 23, 45, 0, time.UTC),
			period:   4 * time.Hour,
			expected: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "1d alignment",
			time:     time.Date(2024, 1, 1, 14, 23, 45, 0, time.UTC),
			period:   24 * time.Hour,
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlignTimestamp(tt.time, tt.period)
			if !got.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestResampler_EmptyCandles(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1m)
	result, err := r.Resample([]*market.Candle{}, market.Timeframe5m)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d candles", len(result))
	}
}

func TestResampler_SameTimeframe(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1m)
	input := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
	}

	result, err := r.Resample(input, market.Timeframe1m)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != len(input) {
		t.Errorf("expected %d candles, got %d", len(input), len(result))
	}

	// Verify values match
	for i := range result {
		if result[i].Open != input[i].Open || result[i].Close != input[i].Close {
			t.Errorf("candle %d mismatch", i)
		}
	}
}

func TestResampler_InvalidTargetTimeframe(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe5m)
	input := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}

	_, err := r.Resample(input, market.Timeframe1m)
	if err == nil {
		t.Error("expected error when target timeframe < source timeframe")
	}
}

func TestResampler_1mTo5m(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Create 5 x 1-minute candles that should aggregate into 1 x 5-minute candle
	input := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		{Timestamp: time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC), Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
		{Timestamp: time.Date(2024, 1, 1, 0, 3, 0, 0, time.UTC), Open: 115, High: 125, Low: 110, Close: 120, Volume: 1300},
		{Timestamp: time.Date(2024, 1, 1, 0, 4, 0, 0, time.UTC), Open: 120, High: 130, Low: 115, Close: 125, Volume: 1400},
	}

	result, err := r.Resample(input, market.Timeframe5m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(result))
	}

	candle := result[0]

	// Verify OHLCV aggregation
	if candle.Open != 100 { // First open
		t.Errorf("expected open 100, got %f", candle.Open)
	}
	if candle.High != 130 { // Highest high
		t.Errorf("expected high 130, got %f", candle.High)
	}
	if candle.Low != 95 { // Lowest low
		t.Errorf("expected low 95, got %f", candle.Low)
	}
	if candle.Close != 125 { // Last close
		t.Errorf("expected close 125, got %f", candle.Close)
	}
	if candle.Volume != 6400 { // Sum of volumes
		t.Errorf("expected volume 6400, got %f", candle.Volume)
	}

	// Verify timestamp is aligned to 5m period start
	expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !candle.Timestamp.Equal(expectedTime) {
		t.Errorf("expected timestamp %v, got %v", expectedTime, candle.Timestamp)
	}
}

func TestResampler_1mTo15m(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Create 15 x 1-minute candles
	var input []*market.Candle
	for i := 0; i < 15; i++ {
		input = append(input, &market.Candle{
			Timestamp: time.Date(2024, 1, 1, 0, i, 0, 0, time.UTC),
			Open:      float64(100 + i),
			High:      float64(110 + i),
			Low:       float64(90 + i),
			Close:     float64(105 + i),
			Volume:    1000,
		})
	}

	result, err := r.Resample(input, market.Timeframe15m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(result))
	}

	candle := result[0]

	// Verify aggregation
	if candle.Open != 100 { // First open
		t.Errorf("expected open 100, got %f", candle.Open)
	}
	if candle.High != 124 { // Highest high (110 + 14)
		t.Errorf("expected high 124, got %f", candle.High)
	}
	if candle.Low != 90 { // Lowest low
		t.Errorf("expected low 90, got %f", candle.Low)
	}
	if candle.Close != 119 { // Last close (105 + 14)
		t.Errorf("expected close 119, got %f", candle.Close)
	}
	if candle.Volume != 15000 { // Sum of volumes
		t.Errorf("expected volume 15000, got %f", candle.Volume)
	}
}

func TestResampler_1mTo1h(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Create 60 x 1-minute candles
	var input []*market.Candle
	for i := 0; i < 60; i++ {
		input = append(input, &market.Candle{
			Timestamp: time.Date(2024, 1, 1, 0, i, 0, 0, time.UTC),
			Open:      100,
			High:      110,
			Low:       90,
			Close:     105,
			Volume:    100,
		})
	}

	result, err := r.Resample(input, market.Timeframe1h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(result))
	}

	candle := result[0]

	if candle.Volume != 6000 { // 60 * 100
		t.Errorf("expected volume 6000, got %f", candle.Volume)
	}

	expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !candle.Timestamp.Equal(expectedTime) {
		t.Errorf("expected timestamp %v, got %v", expectedTime, candle.Timestamp)
	}
}

func TestResampler_MultiplePeriods(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Create 10 x 1-minute candles (2 x 5-minute periods)
	var input []*market.Candle
	for i := 0; i < 10; i++ {
		input = append(input, &market.Candle{
			Timestamp: time.Date(2024, 1, 1, 0, i, 0, 0, time.UTC),
			Open:      float64(100 + i),
			High:      float64(110 + i),
			Low:       float64(90 + i),
			Close:     float64(105 + i),
			Volume:    1000,
		})
	}

	result, err := r.Resample(input, market.Timeframe5m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 candles, got %d", len(result))
	}

	// First period (0:00 - 0:04)
	c1 := result[0]
	if c1.Open != 100 {
		t.Errorf("period 1: expected open 100, got %f", c1.Open)
	}
	if c1.High != 114 { // 110 + 4
		t.Errorf("period 1: expected high 114, got %f", c1.High)
	}
	if c1.Low != 90 {
		t.Errorf("period 1: expected low 90, got %f", c1.Low)
	}
	if c1.Close != 109 { // 105 + 4
		t.Errorf("period 1: expected close 109, got %f", c1.Close)
	}
	if c1.Volume != 5000 {
		t.Errorf("period 1: expected volume 5000, got %f", c1.Volume)
	}

	// Second period (0:05 - 0:09)
	c2 := result[1]
	if c2.Open != 105 { // 100 + 5
		t.Errorf("period 2: expected open 105, got %f", c2.Open)
	}
	if c2.High != 119 { // 110 + 9
		t.Errorf("period 2: expected high 119, got %f", c2.High)
	}
	if c2.Low != 95 { // 90 + 5
		t.Errorf("period 2: expected low 95, got %f", c2.Low)
	}
	if c2.Close != 114 { // 105 + 9
		t.Errorf("period 2: expected close 114, got %f", c2.Close)
	}
	if c2.Volume != 5000 {
		t.Errorf("period 2: expected volume 5000, got %f", c2.Volume)
	}
}

func TestResampler_PartialPeriod(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Create 7 x 1-minute candles (1 full 5m period + 2 candles)
	var input []*market.Candle
	for i := 0; i < 7; i++ {
		input = append(input, &market.Candle{
			Timestamp: time.Date(2024, 1, 1, 0, i, 0, 0, time.UTC),
			Open:      100,
			High:      110,
			Low:       90,
			Close:     105,
			Volume:    1000,
		})
	}

	result, err := r.Resample(input, market.Timeframe5m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce 2 candles: 1 full + 1 partial
	if len(result) != 2 {
		t.Fatalf("expected 2 candles, got %d", len(result))
	}

	// Full period
	if result[0].Volume != 5000 {
		t.Errorf("full period: expected volume 5000, got %f", result[0].Volume)
	}

	// Partial period (only 2 candles)
	if result[1].Volume != 2000 {
		t.Errorf("partial period: expected volume 2000, got %f", result[1].Volume)
	}
}

func TestResampler_5mTo1h(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe5m)

	// Create 12 x 5-minute candles (1 hour)
	var input []*market.Candle
	for i := 0; i < 12; i++ {
		input = append(input, &market.Candle{
			Timestamp: time.Date(2024, 1, 1, 0, i*5, 0, 0, time.UTC),
			Open:      float64(100 + i),
			High:      float64(110 + i),
			Low:       float64(90 + i),
			Close:     float64(105 + i),
			Volume:    1000,
		})
	}

	result, err := r.Resample(input, market.Timeframe1h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(result))
	}

	candle := result[0]
	if candle.Open != 100 {
		t.Errorf("expected open 100, got %f", candle.Open)
	}
	if candle.High != 121 { // 110 + 11
		t.Errorf("expected high 121, got %f", candle.High)
	}
	if candle.Low != 90 {
		t.Errorf("expected low 90, got %f", candle.Low)
	}
	if candle.Close != 116 { // 105 + 11
		t.Errorf("expected close 116, got %f", candle.Close)
	}
	if candle.Volume != 12000 {
		t.Errorf("expected volume 12000, got %f", candle.Volume)
	}
}

func TestResampler_1hTo4h(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1h)

	// Create 8 x 1-hour candles (2 x 4-hour periods)
	var input []*market.Candle
	for i := 0; i < 8; i++ {
		input = append(input, &market.Candle{
			Timestamp: time.Date(2024, 1, 1, i, 0, 0, 0, time.UTC),
			Open:      float64(100 + i*10),
			High:      float64(110 + i*10),
			Low:       float64(90 + i*10),
			Close:     float64(105 + i*10),
			Volume:    1000,
		})
	}

	result, err := r.Resample(input, market.Timeframe4h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 candles, got %d", len(result))
	}

	// First 4h period
	c1 := result[0]
	if c1.Open != 100 {
		t.Errorf("period 1: expected open 100, got %f", c1.Open)
	}
	if c1.Close != 135 { // 105 + 3*10
		t.Errorf("period 1: expected close 135, got %f", c1.Close)
	}

	// Second 4h period
	c2 := result[1]
	if c2.Open != 140 { // 100 + 4*10
		t.Errorf("period 2: expected open 140, got %f", c2.Open)
	}
	if c2.Close != 175 { // 105 + 7*10
		t.Errorf("period 2: expected close 175, got %f", c2.Close)
	}
}

func TestResampler_1hTo1d(t *testing.T) {
	r := NewDefaultResampler(market.Timeframe1h)

	// Create 24 x 1-hour candles (1 day)
	var input []*market.Candle
	for i := 0; i < 24; i++ {
		input = append(input, &market.Candle{
			Timestamp: time.Date(2024, 1, 1, i, 0, 0, 0, time.UTC),
			Open:      100,
			High:      110,
			Low:       90,
			Close:     105,
			Volume:    100,
		})
	}

	result, err := r.Resample(input, market.Timeframe1d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(result))
	}

	candle := result[0]
	if candle.Volume != 2400 { // 24 * 100
		t.Errorf("expected volume 2400, got %f", candle.Volume)
	}

	expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !candle.Timestamp.Equal(expectedTime) {
		t.Errorf("expected timestamp %v, got %v", expectedTime, candle.Timestamp)
	}
}
