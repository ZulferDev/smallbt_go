package csv

import (
	"os"
	"testing"
)

func TestAutoDetectColumns(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "standard order",
			headers: []string{"timestamp", "open", "high", "low", "close", "volume"},
			wantErr: false,
		},
		{
			name:    "different order",
			headers: []string{"open", "high", "low", "close", "volume", "timestamp"},
			wantErr: false,
		},
		{
			name:    "case insensitive",
			headers: []string{"TIMESTAMP", "Open", "HIGH", "low", "Close", "VOLUME"},
			wantErr: false,
		},
		{
			name:    "alternative names",
			headers: []string{"time", "open", "high", "low", "close", "vol"},
			wantErr: false,
		},
		{
			name:        "missing volume",
			headers:     []string{"timestamp", "open", "high", "low", "close"},
			wantErr:     true,
			errContains: "missing required columns: [volume]",
		},
		{
			name:        "missing multiple",
			headers:     []string{"timestamp", "open", "high"},
			wantErr:     true,
			errContains: "missing required columns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultCSVConfig("BTCUSDT", "1h")
			err := autoDetectColumns(tt.headers, &config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want to contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCSVFeedWithDifferentColumnOrder(t *testing.T) {
	// Create temporary CSV with different column order
	tmpfile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := `open,high,low,close,volume,timestamp
100.0,105.0,99.0,102.0,1000.0,2020-01-01 00:00:00
102.0,108.0,101.0,107.0,1200.0,2020-01-01 01:00:00
107.0,112.0,106.0,110.0,1500.0,2020-01-01 02:00:00
`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config := DefaultCSVConfig("BTCUSDT", "1h")
	feed, err := NewCSVFeed(tmpfile.Name(), config)
	if err != nil {
		t.Fatalf("NewCSVFeed failed: %v", err)
	}

	if feed.Length() != 3 {
		t.Errorf("expected 3 candles, got %d", feed.Length())
	}

	// Verify first candle
	md, err := feed.Next()
	if err != nil {
		t.Fatalf("Next() failed: %v", err)
	}

	candle := md.Candles[0]
	if candle.Open != 100.0 {
		t.Errorf("expected open=100.0, got %v", candle.Open)
	}
	if candle.Close != 102.0 {
		t.Errorf("expected close=102.0, got %v", candle.Close)
	}
}

func TestCSVFeedWithMilliseconds(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := `timestamp,open,high,low,close,volume
2017-08-17 04:00:00.000,100.0,105.0,99.0,102.0,1000.0
2017-08-17 05:00:00.123,102.0,108.0,101.0,107.0,1200.0
2017-08-17 06:00:00.999,107.0,112.0,106.0,110.0,1500.0
`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config := DefaultCSVConfig("BTCUSDT", "1h")
	feed, err := NewCSVFeed(tmpfile.Name(), config)
	if err != nil {
		t.Fatalf("NewCSVFeed failed: %v", err)
	}

	if feed.Length() != 3 {
		t.Errorf("expected 3 candles, got %d", feed.Length())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
