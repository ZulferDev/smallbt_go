package market

import (
	"testing"
	"time"
)

func TestCandle_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		candle Candle
		want   bool
	}{
		{
			name: "valid candle",
			candle: Candle{
				Timestamp: time.Now(),
				Open:      100,
				High:      110,
				Low:       95,
				Close:     105,
				Volume:    1000,
			},
			want: true,
		},
		{
			name: "high lower than low",
			candle: Candle{
				Timestamp: time.Now(),
				Open:      100,
				High:      90,
				Low:       95,
				Close:     105,
				Volume:    1000,
			},
			want: false,
		},
		{
			name: "negative volume",
			candle: Candle{
				Timestamp: time.Now(),
				Open:      100,
				High:      110,
				Low:       95,
				Close:     105,
				Volume:    -100,
			},
			want: false,
		},
		{
			name: "negative price",
			candle: Candle{
				Timestamp: time.Now(),
				Open:      -100,
				High:      110,
				Low:       95,
				Close:     105,
				Volume:    1000,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.candle.IsValid(); got != tt.want {
				t.Errorf("Candle.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarketData_AddCandle(t *testing.T) {
	md := NewMarketData("BTCUSDT", Timeframe1h)

	if md.Length() != 0 {
		t.Errorf("expected initial length 0, got %d", md.Length())
	}

	candle := Candle{
		Timestamp: time.Now(),
		Open:      100,
		High:      110,
		Low:       95,
		Close:     105,
		Volume:    1000,
	}

	md.AddCandle(candle)

	if md.Length() != 1 {
		t.Errorf("expected length 1, got %d", md.Length())
	}

	latest := md.GetLatest()
	if latest == nil {
		t.Fatal("expected latest candle, got nil")
	}

	if latest.Close != 105 {
		t.Errorf("expected close 105, got %f", latest.Close)
	}
}

func TestMarketData_GetLatest(t *testing.T) {
	md := NewMarketData(Symbol("BTCUSDT"), Timeframe("1h"))

	// Test empty data
	latest := md.GetLatest()
	if latest != nil {
		t.Error("GetLatest() on empty data should return nil")
	}

	// Add one candle
	candle1 := Candle{
		Timestamp: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Open:      100,
		High:      110,
		Low:       90,
		Close:     105,
		Volume:    1000,
	}
	md.AddCandle(candle1)

	latest = md.GetLatest()
	if latest == nil {
		t.Fatal("GetLatest() should return non-nil after adding candle")
	}
	if latest.Close != 105 {
		t.Errorf("GetLatest() Close = %f, want 105", latest.Close)
	}

	// Add second candle (newer)
	candle2 := Candle{
		Timestamp: time.Date(2021, 1, 1, 1, 0, 0, 0, time.UTC),
		Open:      105,
		High:      120,
		Low:       95,
		Close:     115,
		Volume:    1500,
	}
	md.AddCandle(candle2)

	latest = md.GetLatest()
	if latest == nil {
		t.Fatal("GetLatest() should return non-nil")
	}
	if latest.Close != 115 {
		t.Errorf("GetLatest() Close = %f, want 115 (latest candle)", latest.Close)
	}
}
