package validation

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestNewDefaultValidator(t *testing.T) {
	v := NewDefaultValidator()
	
	if v == nil {
		t.Fatal("NewDefaultValidator returned nil")
	}
	
	if v.AllowGaps != false {
		t.Errorf("expected AllowGaps=false, got %v", v.AllowGaps)
	}
	
	if v.MaxGapDuration != time.Hour {
		t.Errorf("expected MaxGapDuration=1h, got %v", v.MaxGapDuration)
	}
}

func TestValidator_ValidateNil(t *testing.T) {
	v := NewDefaultValidator()
	
	_, err := v.Validate(nil)
	if err == nil {
		t.Fatal("expected error for nil candles, got nil")
	}
}

func TestValidator_ValidateEmpty(t *testing.T) {
	v := NewDefaultValidator()
	
	report, err := v.Validate([]*market.Candle{})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	
	if !report.Valid {
		t.Error("expected Valid=true for empty slice")
	}
	
	if report.TotalCandles != 0 {
		t.Errorf("expected TotalCandles=0, got %d", report.TotalCandles)
	}
}

func TestValidator_ValidateValidCandles(t *testing.T) {
	v := NewDefaultValidator()
	
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := []*market.Candle{
		{
			Timestamp: baseTime,
			Open:      100.0,
			High:      110.0,
			Low:       90.0,
			Close:     105.0,
			Volume:    1000.0,
		},
		{
			Timestamp: baseTime.Add(time.Minute),
			Open:      105.0,
			High:      115.0,
			Low:       100.0,
			Close:     110.0,
			Volume:    1500.0,
		},
	}
	
	report, err := v.Validate(candles)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	
	if !report.Valid {
		t.Errorf("expected Valid=true, got false: %s", report.Summary)
	}
	
	if report.ErrorCount() != 0 {
		t.Errorf("expected 0 errors, got %d", report.ErrorCount())
	}
	
	if report.ValidCandles != 2 {
		t.Errorf("expected ValidCandles=2, got %d", report.ValidCandles)
	}
}

func TestValidator_ValidateOHLC(t *testing.T) {
	v := NewDefaultValidator()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name    string
		candle  *market.Candle
		wantErr bool
	}{
		{
			name: "high < low",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      100,
				High:      90,  // Invalid: high < low
				Low:       100,
				Close:     95,
				Volume:    1000,
			},
			wantErr: true,
		},
		{
			name: "high < open",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      110,
				High:      100,  // Invalid: high < open
				Low:       90,
				Close:     95,
				Volume:    1000,
			},
			wantErr: true,
		},
		{
			name: "high < close",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      100,
				High:      100,  // Invalid: high < close
				Low:       90,
				Close:     110,
				Volume:    1000,
			},
			wantErr: true,
		},
		{
			name: "low > open",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      90,
				High:      110,
				Low:       100,  // Invalid: low > open
				Close:     95,
				Volume:    1000,
			},
			wantErr: true,
		},
		{
			name: "low > close",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      100,
				High:      110,
				Low:       100,  // Invalid: low > close
				Close:     90,
				Volume:    1000,
			},
			wantErr: true,
		},
		{
			name: "valid OHLC",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      100,
				High:      110,
				Low:       90,
				Close:     105,
				Volume:    1000,
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := v.Validate([]*market.Candle{tt.candle})
			if err != nil {
				t.Fatalf("Validate failed: %v", err)
			}
			
			hasError := !report.Valid
			if hasError != tt.wantErr {
				t.Errorf("expected error=%v, got %v (report: %s)", 
					tt.wantErr, hasError, report.Summary)
			}
			
			if tt.wantErr && report.ErrorCount() == 0 {
				t.Error("expected errors in report, got none")
			}
			
			if tt.wantErr {
				found := false
				for _, e := range report.Errors {
					if e.Type == ErrorTypeOHLC {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected ErrorTypeOHLC in errors")
				}
			}
		})
	}
}

func TestValidator_ValidateInvalidValues(t *testing.T) {
	v := NewDefaultValidator()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name   string
		candle *market.Candle
	}{
		{
			name: "negative open",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      -100,
				High:      110,
				Low:       90,
				Close:     105,
				Volume:    1000,
			},
		},
		{
			name: "zero high",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      100,
				High:      0,
				Low:       90,
				Close:     105,
				Volume:    1000,
			},
		},
		{
			name: "negative low",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      100,
				High:      110,
				Low:       -90,
				Close:     105,
				Volume:    1000,
			},
		},
		{
			name: "zero close",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      100,
				High:      110,
				Low:       90,
				Close:     0,
				Volume:    1000,
			},
		},
		{
			name: "negative volume",
			candle: &market.Candle{
				Timestamp: baseTime,
				Open:      100,
				High:      110,
				Low:       90,
				Close:     105,
				Volume:    -1000,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := v.Validate([]*market.Candle{tt.candle})
			if err != nil {
				t.Fatalf("Validate failed: %v", err)
			}
			
			if report.Valid {
				t.Error("expected Valid=false for invalid value")
			}
			
			if report.ErrorCount() == 0 {
				t.Error("expected errors for invalid value")
			}
			
			// Accept either InvalidValue or OHLC error type
			// (negative/zero values may trigger OHLC checks first)
			found := false
			for _, e := range report.Errors {
				if e.Type == ErrorTypeInvalidValue || e.Type == ErrorTypeOHLC {
					found = true
					break
				}
			}
			if !found {
				t.Error("expected ErrorTypeInvalidValue or ErrorTypeOHLC in errors")
			}
		})
	}
}

func TestValidator_ValidateOrdering(t *testing.T) {
	v := NewDefaultValidator()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	t.Run("chronological order", func(t *testing.T) {
		candles := []*market.Candle{
			{Timestamp: baseTime, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: baseTime.Add(time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
			{Timestamp: baseTime.Add(2 * time.Minute), Open: 110, High: 120, Low: 105, Close: 115, Volume: 2000},
		}
		
		report, err := v.Validate(candles)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		
		if !report.Valid {
			t.Errorf("expected Valid=true for chronological order: %s", report.Summary)
		}
	})
	
	t.Run("out of order", func(t *testing.T) {
		candles := []*market.Candle{
			{Timestamp: baseTime.Add(time.Minute), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: baseTime, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},  // Goes backward
		}
		
		report, err := v.Validate(candles)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		
		if report.Valid {
			t.Error("expected Valid=false for out of order timestamps")
		}
		
		found := false
		for _, e := range report.Errors {
			if e.Type == ErrorTypeOrdering {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected ErrorTypeOrdering in errors")
		}
	})
}

func TestValidator_ValidateDuplicates(t *testing.T) {
	v := NewDefaultValidator()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	candles := []*market.Candle{
		{Timestamp: baseTime, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: baseTime, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},  // Duplicate timestamp
	}
	
	report, err := v.Validate(candles)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	
	if report.Valid {
		t.Error("expected Valid=false for duplicate timestamps")
	}
	
	if report.DuplicatesFound != 1 {
		t.Errorf("expected DuplicatesFound=1, got %d", report.DuplicatesFound)
	}
	
	found := false
	for _, e := range report.Errors {
		if e.Type == ErrorTypeDuplicate {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ErrorTypeDuplicate in errors")
	}
}

func TestValidator_ValidateGaps(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	t.Run("gaps not allowed", func(t *testing.T) {
		v := NewDefaultValidator()
		v.AllowGaps = false
		v.MaxGapDuration = 5 * time.Minute
		
		candles := []*market.Candle{
			{Timestamp: baseTime, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: baseTime.Add(10 * time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},  // 10min gap
		}
		
		report, err := v.Validate(candles)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		
		if report.Valid {
			t.Error("expected Valid=false when gaps not allowed")
		}
		
		if len(report.Gaps) != 1 {
			t.Errorf("expected 1 gap detected, got %d", len(report.Gaps))
		}
		
		found := false
		for _, e := range report.Errors {
			if e.Type == ErrorTypeGap {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected ErrorTypeGap in errors")
		}
	})
	
	t.Run("gaps allowed", func(t *testing.T) {
		v := NewDefaultValidator()
		v.AllowGaps = true
		v.MaxGapDuration = 5 * time.Minute
		
		candles := []*market.Candle{
			{Timestamp: baseTime, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: baseTime.Add(10 * time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},  // 10min gap
		}
		
		report, err := v.Validate(candles)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		
		if !report.Valid {
			t.Errorf("expected Valid=true when gaps allowed: %s", report.Summary)
		}
		
		if len(report.Gaps) != 1 {
			t.Errorf("expected 1 gap detected, got %d", len(report.Gaps))
		}
		
		if report.WarningCount() != 1 {
			t.Errorf("expected 1 warning for gap, got %d", report.WarningCount())
		}
	})
	
	t.Run("no gaps", func(t *testing.T) {
		v := NewDefaultValidator()
		v.MaxGapDuration = 5 * time.Minute
		
		candles := []*market.Candle{
			{Timestamp: baseTime, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: baseTime.Add(time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		}
		
		report, err := v.Validate(candles)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		
		if !report.Valid {
			t.Errorf("expected Valid=true with no gaps: %s", report.Summary)
		}
		
		if len(report.Gaps) != 0 {
			t.Errorf("expected 0 gaps, got %d", len(report.Gaps))
		}
	})
}

func TestValidator_ValidateNilCandle(t *testing.T) {
	v := NewDefaultValidator()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	candles := []*market.Candle{
		{Timestamp: baseTime, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		nil,  // Nil candle
		{Timestamp: baseTime.Add(2 * time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
	}
	
	report, err := v.Validate(candles)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	
	if report.Valid {
		t.Error("expected Valid=false for nil candle")
	}
	
	if report.ErrorCount() == 0 {
		t.Error("expected errors for nil candle")
	}
}

func TestValidationReport_Methods(t *testing.T) {
	report := &ValidationReport{
		Valid:        false,
		TotalCandles: 10,
		Errors: []ValidationError{
			{Type: ErrorTypeOHLC, Message: "test error 1"},
			{Type: ErrorTypeInvalidValue, Message: "test error 2"},
		},
		Warnings: []ValidationWarning{
			{Type: ErrorTypeGap, Message: "test warning"},
		},
	}
	
	t.Run("HasErrors", func(t *testing.T) {
		if !report.HasErrors() {
			t.Error("expected HasErrors=true")
		}
	})
	
	t.Run("HasWarnings", func(t *testing.T) {
		if !report.HasWarnings() {
			t.Error("expected HasWarnings=true")
		}
	})
	
	t.Run("ErrorCount", func(t *testing.T) {
		if report.ErrorCount() != 2 {
			t.Errorf("expected ErrorCount=2, got %d", report.ErrorCount())
		}
	})
	
	t.Run("WarningCount", func(t *testing.T) {
		if report.WarningCount() != 1 {
			t.Errorf("expected WarningCount=1, got %d", report.WarningCount())
		}
	})
	
	t.Run("String", func(t *testing.T) {
		s := report.String()
		if s == "" {
			t.Error("expected non-empty string")
		}
	})
}

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		want      string
	}{
		{ErrorTypeOHLC, "OHLC"},
		{ErrorTypeOrdering, "Ordering"},
		{ErrorTypeDuplicate, "Duplicate"},
		{ErrorTypeInvalidValue, "InvalidValue"},
		{ErrorTypeGap, "Gap"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.errorType.String(); got != tt.want {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityWarning, "Warning"},
		{SeverityError, "Error"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
