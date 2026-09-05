package validation

import (
	"fmt"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// ErrorType represents the type of validation error.
type ErrorType int

const (
	ErrorTypeOHLC ErrorType = iota
	ErrorTypeOrdering
	ErrorTypeDuplicate
	ErrorTypeInvalidValue
	ErrorTypeGap
)

func (e ErrorType) String() string {
	switch e {
	case ErrorTypeOHLC:
		return "OHLC"
	case ErrorTypeOrdering:
		return "Ordering"
	case ErrorTypeDuplicate:
		return "Duplicate"
	case ErrorTypeInvalidValue:
		return "InvalidValue"
	case ErrorTypeGap:
		return "Gap"
	default:
		return "Unknown"
	}
}

// Severity represents the severity of a validation issue.
type Severity int

const (
	SeverityWarning Severity = iota
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "Warning"
	case SeverityError:
		return "Error"
	default:
		return "Unknown"
	}
}

// ValidationError represents a single validation error.
type ValidationError struct {
	Index     int
	Timestamp time.Time
	Type      ErrorType
	Message   string
	Severity  Severity
}

// ValidationWarning represents a non-critical validation issue.
type ValidationWarning struct {
	Index     int
	Timestamp time.Time
	Type      ErrorType
	Message   string
}

// Gap represents a detected gap in the data.
type Gap struct {
	StartIndex int
	EndIndex   int
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
}

// ValidationReport contains the results of data validation.
type ValidationReport struct {
	Valid            bool
	TotalCandles     int
	ValidCandles     int
	Errors           []ValidationError
	Warnings         []ValidationWarning
	Gaps             []Gap
	DuplicatesFound  int
	Summary          string
}

// HasErrors returns true if the report contains any errors.
func (r *ValidationReport) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if the report contains any warnings.
func (r *ValidationReport) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ErrorCount returns the number of errors.
func (r *ValidationReport) ErrorCount() int {
	return len(r.Errors)
}

// WarningCount returns the number of warnings.
func (r *ValidationReport) WarningCount() int {
	return len(r.Warnings)
}

// String returns a human-readable summary of the validation report.
func (r *ValidationReport) String() string {
	if r.Valid {
		return fmt.Sprintf("Validation passed: %d candles validated", r.TotalCandles)
	}
	return fmt.Sprintf("Validation failed: %d errors, %d warnings in %d candles",
		r.ErrorCount(), r.WarningCount(), r.TotalCandles)
}

// Validator validates candle data.
type Validator interface {
	Validate(candles []*market.Candle) (*ValidationReport, error)
}

// DefaultValidator implements comprehensive candle validation.
type DefaultValidator struct {
	// Configuration
	AllowGaps      bool
	MaxGapDuration time.Duration
}

// NewDefaultValidator creates a new DefaultValidator with default settings.
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{
		AllowGaps:      false,
		MaxGapDuration: time.Hour,
	}
}

// Validate performs comprehensive validation on candle data.
func (v *DefaultValidator) Validate(candles []*market.Candle) (*ValidationReport, error) {
	if candles == nil {
		return nil, fmt.Errorf("candles slice is nil")
	}

	report := &ValidationReport{
		Valid:        true,
		TotalCandles: len(candles),
		Errors:       make([]ValidationError, 0),
		Warnings:     make([]ValidationWarning, 0),
		Gaps:         make([]Gap, 0),
	}

	if len(candles) == 0 {
		report.Summary = "No candles to validate"
		return report, nil
	}

	validCount := 0

	for i, candle := range candles {
		if candle == nil {
			report.Errors = append(report.Errors, ValidationError{
				Index:    i,
				Type:     ErrorTypeInvalidValue,
				Message:  "candle is nil",
				Severity: SeverityError,
			})
			report.Valid = false
			continue
		}

		// Validate OHLC consistency
		if err := v.validateOHLC(candle, i, report); err != nil {
			report.Valid = false
			continue
		}

		// Validate values are not negative or zero
		if err := v.validateValues(candle, i, report); err != nil {
			report.Valid = false
			continue
		}

		// Validate chronological ordering (if not first candle)
		if i > 0 && candles[i-1] != nil {
			if err := v.validateOrdering(candles[i-1], candle, i, report); err != nil {
				report.Valid = false
				continue
			}

			// Detect duplicates
			if candle.Timestamp.Equal(candles[i-1].Timestamp) {
				report.Errors = append(report.Errors, ValidationError{
					Index:     i,
					Timestamp: candle.Timestamp,
					Type:      ErrorTypeDuplicate,
					Message:   "duplicate timestamp",
					Severity:  SeverityError,
				})
				report.DuplicatesFound++
				report.Valid = false
				continue
			}

			// Detect gaps
			gap := candle.Timestamp.Sub(candles[i-1].Timestamp)
			if gap > v.MaxGapDuration {
				gapInfo := Gap{
					StartIndex: i - 1,
					EndIndex:   i,
					StartTime:  candles[i-1].Timestamp,
					EndTime:    candle.Timestamp,
					Duration:   gap,
				}
				report.Gaps = append(report.Gaps, gapInfo)

				if !v.AllowGaps {
					report.Errors = append(report.Errors, ValidationError{
						Index:     i,
						Timestamp: candle.Timestamp,
						Type:      ErrorTypeGap,
						Message:   fmt.Sprintf("gap of %v exceeds max %v", gap, v.MaxGapDuration),
						Severity:  SeverityError,
					})
					report.Valid = false
					continue
				} else {
					report.Warnings = append(report.Warnings, ValidationWarning{
						Index:     i,
						Timestamp: candle.Timestamp,
						Type:      ErrorTypeGap,
						Message:   fmt.Sprintf("gap of %v detected", gap),
					})
				}
			}
		}

		validCount++
	}

	report.ValidCandles = validCount

	// Generate summary
	if report.Valid {
		report.Summary = fmt.Sprintf("All %d candles valid", report.TotalCandles)
	} else {
		report.Summary = fmt.Sprintf("%d errors, %d warnings found in %d candles",
			report.ErrorCount(), report.WarningCount(), report.TotalCandles)
	}

	return report, nil
}

// validateOHLC checks OHLC consistency rules.
func (v *DefaultValidator) validateOHLC(candle *market.Candle, index int, report *ValidationReport) error {
	// High must be >= Low
	if candle.High < candle.Low {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeOHLC,
			Message:   fmt.Sprintf("high (%.2f) < low (%.2f)", candle.High, candle.Low),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid OHLC")
	}

	// High must be >= Open
	if candle.High < candle.Open {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeOHLC,
			Message:   fmt.Sprintf("high (%.2f) < open (%.2f)", candle.High, candle.Open),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid OHLC")
	}

	// High must be >= Close
	if candle.High < candle.Close {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeOHLC,
			Message:   fmt.Sprintf("high (%.2f) < close (%.2f)", candle.High, candle.Close),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid OHLC")
	}

	// Low must be <= Open
	if candle.Low > candle.Open {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeOHLC,
			Message:   fmt.Sprintf("low (%.2f) > open (%.2f)", candle.Low, candle.Open),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid OHLC")
	}

	// Low must be <= Close
	if candle.Low > candle.Close {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeOHLC,
			Message:   fmt.Sprintf("low (%.2f) > close (%.2f)", candle.Low, candle.Close),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid OHLC")
	}

	return nil
}

// validateValues checks that values are positive and valid.
func (v *DefaultValidator) validateValues(candle *market.Candle, index int, report *ValidationReport) error {
	// Check for negative or zero prices
	if candle.Open <= 0 {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeInvalidValue,
			Message:   fmt.Sprintf("open price <= 0: %.2f", candle.Open),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid value")
	}

	if candle.High <= 0 {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeInvalidValue,
			Message:   fmt.Sprintf("high price <= 0: %.2f", candle.High),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid value")
	}

	if candle.Low <= 0 {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeInvalidValue,
			Message:   fmt.Sprintf("low price <= 0: %.2f", candle.Low),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid value")
	}

	if candle.Close <= 0 {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeInvalidValue,
			Message:   fmt.Sprintf("close price <= 0: %.2f", candle.Close),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid value")
	}

	// Volume can be zero but not negative
	if candle.Volume < 0 {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: candle.Timestamp,
			Type:      ErrorTypeInvalidValue,
			Message:   fmt.Sprintf("volume < 0: %.2f", candle.Volume),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid value")
	}

	return nil
}

// validateOrdering checks chronological ordering.
func (v *DefaultValidator) validateOrdering(prev, current *market.Candle, index int, report *ValidationReport) error {
	if current.Timestamp.Before(prev.Timestamp) {
		report.Errors = append(report.Errors, ValidationError{
			Index:     index,
			Timestamp: current.Timestamp,
			Type:      ErrorTypeOrdering,
			Message:   fmt.Sprintf("timestamp goes backward: %v -> %v", prev.Timestamp, current.Timestamp),
			Severity:  SeverityError,
		})
		return fmt.Errorf("invalid ordering")
	}
	return nil
}
