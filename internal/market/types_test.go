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
