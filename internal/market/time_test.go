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

func TestNow(t *testing.T) {
	now := Now()
	if now.IsZero() {
		t.Error("Now() returned zero time")
	}
	if now.Location() != time.UTC {
		t.Error("Now() should return UTC time")
	}
}

func TestUnix(t *testing.T) {
	timestamp := int64(1609459200)
	mt := Unix(timestamp, 0)
	if mt.Unix() != timestamp {
		t.Errorf("Unix(%d, 0) returned %d", timestamp, mt.Unix())
	}
	if mt.Location() != time.UTC {
		t.Error("Unix() should return UTC time")
	}
}

func TestTime_IsZero(t *testing.T) {
	zero := Time{}
	if !zero.IsZero() {
		t.Error("Zero Time should report IsZero() = true")
	}

	nonZero := Now()
	if nonZero.IsZero() {
		t.Error("Non-zero Time should report IsZero() = false")
	}
}

func TestTime_String(t *testing.T) {
	mt := Unix(1609459200, 0)
	str := mt.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

func TestTime_Format(t *testing.T) {
	mt := Unix(1609459200, 0)
	formatted := mt.Format("2006-01-02")
	if formatted != "2021-01-01" {
		t.Errorf("Format() = %q, want '2021-01-01'", formatted)
	}
}

func TestTime_MarshalJSON(t *testing.T) {
	mt := Unix(1609459200, 0)
	data, err := mt.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalJSON() returned empty data")
	}
}

func TestTime_UnmarshalJSON(t *testing.T) {
	var mt Time
	err := mt.UnmarshalJSON([]byte(`"2021-01-01T00:00:00Z"`))
	if err != nil {
		t.Errorf("UnmarshalJSON() error: %v", err)
	}
	if mt.IsZero() {
		t.Error("UnmarshalJSON() resulted in zero time")
	}
}

func TestTime_MarshalText(t *testing.T) {
	mt := Unix(1609459200, 0)
	data, err := mt.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error: %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalText() returned empty data")
	}
}

func TestTime_UnmarshalText(t *testing.T) {
	var mt Time
	err := mt.UnmarshalText([]byte("2021-01-01T00:00:00Z"))
	if err != nil {
		t.Errorf("UnmarshalText() error: %v", err)
	}
}

func TestTime_Local(t *testing.T) {
	mt := Unix(1609459200, 0).UTC()
	local := mt.Local()
	if mt.Unix() != local.Unix() {
		t.Error("Local() changed the instant in time")
	}
}

func TestTime_UTC(t *testing.T) {
	mt := Unix(1609459200, 0)
	utc := mt.UTC()
	if utc.Location() != time.UTC {
		t.Error("UTC() did not return UTC time")
	}
}

func TestTime_In(t *testing.T) {
	mt := Unix(1609459200, 0)
	utc := mt.In(time.UTC)
	if utc.Location() != time.UTC {
		t.Error("In(UTC) did not change location")
	}
}
