package market

import (
	"testing"
	"time"
)

func TestNewTime(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
	}{
		{
			name: "UTC time",
			in:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "non-UTC time converts to UTC",
			in:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.FixedZone("EST", -5*3600)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTime(tt.in)
			if got.Location() != time.UTC {
				t.Errorf("expected UTC location, got %v", got.Location())
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantEqual time.Time
	}{
		{
			name:      "RFC3339 format",
			input:     "2024-01-15T10:30:00Z",
			wantErr:   false,
			wantEqual: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:      "simple date format",
			input:     "2024-01-15",
			wantErr:   false,
			wantEqual: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "datetime format",
			input:     "2024-01-15 14:30:45",
			wantErr:   false,
			wantEqual: time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(NewTime(tt.wantEqual)) {
				t.Errorf("ParseTime() = %v, want %v", got.Time, tt.wantEqual)
			}
		})
	}
}

func TestTimeComparisons(t *testing.T) {
	t1 := NewTime(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC))
	t2 := NewTime(time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC))
	t1Copy := NewTime(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC))

	if !t1.Before(t2) {
		t.Error("t1 should be before t2")
	}

	if !t2.After(t1) {
		t.Error("t2 should be after t1")
	}

	if !t1.Equal(t1Copy) {
		t.Error("t1 should equal t1Copy")
	}
}

func TestTimeArithmetic(t *testing.T) {
	t1 := NewTime(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC))
	t2 := t1.Add(time.Hour)

	if diff := t2.Sub(t1); diff != time.Hour {
		t.Errorf("expected 1 hour difference, got %v", diff)
	}
}
