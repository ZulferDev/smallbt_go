package portfolio

import "testing"

func TestNewPortfolio(t *testing.T) {
	initialCash := 10000.0
	p := NewPortfolio(initialCash)

	if p.InitialCash != initialCash {
		t.Errorf("expected initial cash %f, got %f", initialCash, p.InitialCash)
	}

	if p.Cash != initialCash {
		t.Errorf("expected cash %f, got %f", initialCash, p.Cash)
	}

	if p.Equity != initialCash {
		t.Errorf("expected equity %f, got %f", initialCash, p.Equity)
	}

	if len(p.Positions) != 0 {
		t.Errorf("expected 0 positions, got %d", len(p.Positions))
	}
}

func TestPosition_UnrealizedPnL(t *testing.T) {
	tests := []struct {
		name     string
		position Position
		want     float64
	}{
		{
			name: "long position profit",
			position: Position{
				Symbol:       "BTCUSDT",
				Side:         PositionSideLong,
				Quantity:     1.0,
				EntryPrice:   100.0,
				CurrentPrice: 110.0,
			},
			want: 10.0,
		},
		{
			name: "long position loss",
			position: Position{
				Symbol:       "BTCUSDT",
				Side:         PositionSideLong,
				Quantity:     1.0,
				EntryPrice:   100.0,
				CurrentPrice: 90.0,
			},
			want: -10.0,
		},
		{
			name: "short position profit",
			position: Position{
				Symbol:       "BTCUSDT",
				Side:         PositionSideShort,
				Quantity:     1.0,
				EntryPrice:   100.0,
				CurrentPrice: 90.0,
			},
			want: 10.0,
		},
		{
			name: "short position loss",
			position: Position{
				Symbol:       "BTCUSDT",
				Side:         PositionSideShort,
				Quantity:     1.0,
				EntryPrice:   100.0,
				CurrentPrice: 110.0,
			},
			want: -10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.position.UnrealizedPnL(); got != tt.want {
				t.Errorf("Position.UnrealizedPnL() = %v, want %v", got, tt.want)
			}
		})
	}
}
