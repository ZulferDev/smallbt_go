package market

import (
	"fmt"
	"time"
)

// Time represents a timestamp with explicit UTC timezone.
// All internal timestamps should use UTC to avoid timezone confusion.
type Time struct {
	time.Time
}

// NewTime creates a new Time from a time.Time value.
// The input time is converted to UTC if it has a different timezone.
func NewTime(t time.Time) Time {
	return Time{t.UTC()}
}

// ParseTime parses a string into a Time.
// Supported formats: RFC3339, RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02".
// All times are parsed and stored as UTC.
func ParseTime(s string) (Time, error) {
	// Try common formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return NewTime(t), nil
		}
		lastErr = err
	}

	return Time{}, fmt.Errorf("failed to parse time %q: %w", s, lastErr)
}

// Now returns the current time as a Time in UTC.
func Now() Time {
	return NewTime(time.Now().UTC())
}

// Unix returns the local Time corresponding to the given Unix time.
func Unix(sec int64, nsec int64) Time {
	return NewTime(time.Unix(sec, nsec).UTC())
}

// IsZero reports whether t represents the zero time instant.
func (t Time) IsZero() bool {
	return t.Time.IsZero()
}

// Before reports whether t is before u.
func (t Time) Before(u Time) bool {
	return t.Time.Before(u.Time)
}

// After reports whether t is after u.
func (t Time) After(u Time) bool {
	return t.Time.After(u.Time)
}

// Equal reports whether t equals u.
func (t Time) Equal(u Time) bool {
	return t.Time.Equal(u.Time)
}

// Add returns t + duration.
func (t Time) Add(d time.Duration) Time {
	return NewTime(t.Time.Add(d))
}

// Sub returns t - u.
func (t Time) Sub(u Time) time.Duration {
	return t.Time.Sub(u.Time)
}

// String returns the time formatted using RFC3339.
func (t Time) String() string {
	return t.Time.Format(time.RFC3339)
}

// Format returns the time formatted using the specified format string.
func (t Time) Format(format string) string {
	return t.Time.Format(format)
}

// MarshalJSON implements json.Marshaler.
func (t Time) MarshalJSON() ([]byte, error) {
	return t.Time.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Time) UnmarshalJSON(data []byte) error {
	var parsed time.Time
	if err := parsed.UnmarshalJSON(data); err != nil {
		return err
	}
	t.Time = parsed.UTC()
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (t Time) MarshalText() ([]byte, error) {
	return t.Time.MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Time) UnmarshalText(data []byte) error {
	var parsed time.Time
	if err := parsed.UnmarshalText(data); err != nil {
		return err
	}
	t.Time = parsed.UTC()
	return nil
}

// Local returns t with the location set to local time.
func (t Time) Local() time.Time {
	return t.Time.Local()
}

// In returns t with the location set to loc.
func (t Time) In(loc *time.Location) time.Time {
	return t.Time.In(loc)
}

// UTC returns t with the location set to UTC.
func (t Time) UTC() time.Time {
	return t.Time.UTC()
}

// Unix returns t as a Unix time, the number of seconds elapsed since January 1, 1970 UTC.
func (t Time) Unix() int64 {
	return t.Time.Unix()
}

// UnixNano returns t as a Unix time, the number of nanoseconds elapsed since January 1, 1970 UTC.
func (t Time) UnixNano() int64 {
	return t.Time.UnixNano()
}
