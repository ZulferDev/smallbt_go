# Data Validation Guide

**smallbt_go Data Validation Documentation**

---

## Overview

Data quality is critical for accurate backtesting. Invalid or corrupted data can produce misleading results, false signals, and incorrect performance metrics. This guide explains the validation system, validation rules, and how to use them effectively.

---

## Why Validate Data?

### Common Data Issues

1. **OHLC Inconsistencies**
   - High < Low (impossible)
   - High < Open or Close (impossible)
   - Low > Open or Close (impossible)

2. **Invalid Values**
   - Negative prices (impossible)
   - Zero prices (invalid)
   - Negative volume (impossible)

3. **Timestamp Issues**
   - Out-of-order timestamps
   - Duplicate timestamps
   - Time gaps (missing data)

4. **Data Corruption**
   - Null/nil values
   - Malformed records
   - Incomplete candles

### Impact of Bad Data

❌ **Without Validation:**
- False positive signals (bad data triggers entry)
- Incorrect performance metrics (unrealistic PnL)
- Look-ahead bias (time ordering issues)
- Strategy failures in production

✅ **With Validation:**
- Reliable backtest results
- Accurate performance metrics
- Clean data for production
- Early detection of data issues

---

## Validation System Architecture

```
Data Source
    ↓
Reader (CSV/Parquet)
    ↓
[]*market.Candle
    ↓
Validator ← Configuration
    ↓
ValidationReport
    ↓
✅ Valid → Continue
❌ Invalid → Fix Data
```

---

## Validation Rules

### 1. OHLC Consistency (5 checks)

**Rules:**
```
High >= Low
High >= Open
High >= Close
Low <= Open
Low <= Close
```

**Example - Valid:**
```
Open:  100.0
High:  110.0  ✓ High >= Open, Close, Low
Low:    90.0  ✓ Low <= Open, Close
Close: 105.0
```

**Example - Invalid:**
```
Open:  100.0
High:   90.0  ✗ High < Open (impossible!)
Low:   100.0
Close:  95.0
```

**Error Message:**
```
[0] OHLC: high (90.00) < open (100.00)
```

### 2. Value Validation (5 checks)

**Rules:**
```
Open > 0
High > 0
Low > 0
Close > 0
Volume >= 0
```

**Example - Valid:**
```
Open:   42000.0  ✓ > 0
High:   42500.0  ✓ > 0
Low:    41800.0  ✓ > 0
Close:  42300.0  ✓ > 0
Volume:   123.45  ✓ >= 0
```

**Example - Invalid:**
```
Open:   -100.0  ✗ Negative price
High:      0.0  ✗ Zero price
Low:      90.0  ✓
Close:   105.0  ✓
Volume:  -50.0  ✗ Negative volume
```

**Error Messages:**
```
[0] InvalidValue: open must be positive, got -100.00
[0] InvalidValue: high must be positive, got 0.00
[0] InvalidValue: volume cannot be negative, got -50.00
```

### 3. Chronological Ordering

**Rule:**
```
candles[i].Timestamp < candles[i+1].Timestamp
```

**Example - Valid:**
```
[0] 2024-01-01T00:00:00Z  ✓
[1] 2024-01-01T00:01:00Z  ✓ After previous
[2] 2024-01-01T00:02:00Z  ✓ After previous
```

**Example - Invalid:**
```
[0] 2024-01-01T00:00:00Z  ✓
[1] 2024-01-01T00:02:00Z  ✓
[2] 2024-01-01T00:01:00Z  ✗ Before previous (out of order!)
```

**Error Message:**
```
[2] Ordering: timestamp 2024-01-01T00:01:00Z not after previous 2024-01-01T00:02:00Z
```

### 4. Duplicate Detection

**Rule:**
```
All timestamps must be unique
```

**Example - Valid:**
```
[0] 2024-01-01T00:00:00Z  ✓
[1] 2024-01-01T00:01:00Z  ✓ Unique
[2] 2024-01-01T00:02:00Z  ✓ Unique
```

**Example - Invalid:**
```
[0] 2024-01-01T00:00:00Z  ✓
[1] 2024-01-01T00:01:00Z  ✓
[2] 2024-01-01T00:01:00Z  ✗ Duplicate!
```

**Error Message:**
```
[2] Duplicate: timestamp 2024-01-01T00:01:00Z already seen
```

### 5. Gap Detection (Configurable)

**Rule:**
```
Time gap between consecutive candles <= MaxGapDuration
```

**Example - No Gap (1-minute data):**
```
[0] 2024-01-01T00:00:00Z
[1] 2024-01-01T00:01:00Z  ✓ 1 minute gap (expected)
[2] 2024-01-01T00:02:00Z  ✓ 1 minute gap (expected)
```

**Example - Gap Detected (1-minute data):**
```
[0] 2024-01-01T00:00:00Z
[1] 2024-01-01T00:10:00Z  ⚠ 10 minute gap (missing data?)
[2] 2024-01-01T00:11:00Z
```

**Warning/Error:**
```
⚠ Gap: from 2024-01-01T00:00:00Z to 2024-01-01T00:10:00Z (duration: 10m0s)
```

**Behavior:**
- If `AllowGaps = false`: **Error** (validation fails)
- If `AllowGaps = true`: **Warning** (validation succeeds)

---

## Validator Configuration

### DefaultValidator

```go
type DefaultValidator struct {
    AllowGaps      bool          // Allow time gaps?
    MaxGapDuration time.Duration // Max allowed gap
}
```

### Default Settings

```go
validator := validation.NewDefaultValidator()
// AllowGaps = false
// MaxGapDuration = 1 hour
```

### Custom Settings

```go
validator := validation.NewDefaultValidator()

// For 24/7 crypto markets (strict)
validator.AllowGaps = false
validator.MaxGapDuration = 5 * time.Minute

// For stock markets (weekends OK)
validator.AllowGaps = true
validator.MaxGapDuration = 24 * time.Hour
```

---

## ValidationReport Structure

```go
type ValidationReport struct {
    Valid            bool                  // Overall result
    TotalCandles     int                   // Total candles checked
    ValidCandles     int                   // Valid candles
    Errors           []ValidationError     // Critical errors
    Warnings         []ValidationWarning   // Non-critical warnings
    Gaps             []Gap                 // Detected gaps
    DuplicatesFound  int                   // Count of duplicates
    Summary          string                // Human-readable summary
}
```

### Helper Methods

```go
report.HasErrors()    // bool - any errors?
report.HasWarnings()  // bool - any warnings?
report.ErrorCount()   // int - count of errors
report.WarningCount() // int - count of warnings
report.String()       // string - formatted report
```

---

## Usage Examples

### Basic Usage

```go
import "github.com/ZulferDev/smallbt_go/internal/data/validation"

// Read data
reader, _ := parquet.NewParquetReader("data.parquet")
candles, _ := reader.Read()
reader.Close()

// Validate with defaults
validator := validation.NewDefaultValidator()
report, err := validator.Validate(candles)
if err != nil {
    log.Fatal(err)
}

// Check result
if !report.Valid {
    log.Printf("Validation failed: %s", report.Summary)
    for _, e := range report.Errors {
        log.Printf("  [%d] %s: %s", e.Index, e.Type, e.Message)
    }
    os.Exit(1)
}

log.Println("✓ Data validation passed")
```

### Custom Configuration

```go
// Create validator for crypto data (24/7 markets)
validator := validation.NewDefaultValidator()
validator.AllowGaps = false              // Strict: no gaps allowed
validator.MaxGapDuration = 5 * time.Minute  // Max 5 minutes between candles

report, _ := validator.Validate(candles)
```

### Stock Market Configuration

```go
// Create validator for stock data (weekends/holidays)
validator := validation.NewDefaultValidator()
validator.AllowGaps = true               // Lenient: gaps OK (weekends)
validator.MaxGapDuration = 72 * time.Hour   // Max 3 days (weekend)

report, _ := validator.Validate(candles)
```

### Detailed Error Handling

```go
report, _ := validator.Validate(candles)

if report.HasErrors() {
    fmt.Printf("Found %d errors:\n", report.ErrorCount())
    for _, e := range report.Errors {
        fmt.Printf("  Index %d: %s - %s\n", e.Index, e.Type, e.Message)
    }
}

if report.HasWarnings() {
    fmt.Printf("Found %d warnings:\n", report.WarningCount())
    for _, w := range report.Warnings {
        fmt.Printf("  Index %d: %s\n", w.Index, w.Message)
    }
}

if len(report.Gaps) > 0 {
    fmt.Printf("Found %d gaps:\n", len(report.Gaps))
    for _, g := range report.Gaps {
        fmt.Printf("  %v to %v (duration: %v)\n", 
            g.StartTime, g.EndTime, g.Duration)
    }
}
```

---

## Validation Workflow

### Recommended Workflow

```
1. Read Data
   ↓
2. Validate Data
   ↓
3. If invalid → Fix & Retry
   ↓
4. If valid → Continue to Backtest
```

### Example Pipeline

```go
func validateAndBacktest(dataPath, strategyPath string) error {
    // Step 1: Read data
    reader, err := parquet.NewParquetReader(dataPath)
    if err != nil {
        return fmt.Errorf("read data: %w", err)
    }
    defer reader.Close()
    
    candles, err := reader.Read()
    if err != nil {
        return fmt.Errorf("parse candles: %w", err)
    }
    
    // Step 2: Validate data
    validator := validation.NewDefaultValidator()
    report, err := validator.Validate(candles)
    if err != nil {
        return fmt.Errorf("validation error: %w", err)
    }
    
    // Step 3: Check result
    if !report.Valid {
        return fmt.Errorf("validation failed:\n%s", report.String())
    }
    
    log.Printf("✓ Validated %d candles", report.TotalCandles)
    
    // Step 4: Run backtest
    return runBacktest(candles, strategyPath)
}
```

---

## Error Types

### ErrorType Enum

```go
const (
    ErrorTypeOHLC         // OHLC consistency violations
    ErrorTypeOrdering     // Timestamp ordering issues
    ErrorTypeDuplicate    // Duplicate timestamps
    ErrorTypeInvalidValue // Negative/zero values
    ErrorTypeGap          // Time gaps (when not allowed)
)
```

### Filtering by Type

```go
// Get only OHLC errors
var ohlcErrors []validation.ValidationError
for _, e := range report.Errors {
    if e.Type == validation.ErrorTypeOHLC {
        ohlcErrors = append(ohlcErrors, e)
    }
}

fmt.Printf("Found %d OHLC errors\n", len(ohlcErrors))
```

---

## Real-World Examples

### Example 1: Clean Data

**Input:**
```go
candles := []*market.Candle{
    {Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
    {Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
    {Timestamp: t2, Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
}
```

**Result:**
```
✓ Validation passed
  Total candles: 3
  Valid candles: 3
  Errors: 0
  Warnings: 0
```

### Example 2: OHLC Error

**Input:**
```go
candles := []*market.Candle{
    {Timestamp: t0, Open: 100, High: 90, Low: 100, Close: 95, Volume: 1000},
    // High (90) < Low (100) - IMPOSSIBLE!
}
```

**Result:**
```
✗ Validation failed
  Total candles: 1
  Valid candles: 0
  Errors: 1
  
  [0] OHLC: high (90.00) < low (100.00)
```

### Example 3: Out of Order

**Input:**
```go
candles := []*market.Candle{
    {Timestamp: t0, ...},  // 2024-01-01 00:00
    {Timestamp: t2, ...},  // 2024-01-01 00:02
    {Timestamp: t1, ...},  // 2024-01-01 00:01 (OUT OF ORDER!)
}
```

**Result:**
```
✗ Validation failed
  Total candles: 3
  Valid candles: 2
  Errors: 1
  
  [2] Ordering: timestamp 2024-01-01T00:01:00Z not after previous 2024-01-01T00:02:00Z
```

### Example 4: Gap Warning

**Input:**
```go
validator := validation.NewDefaultValidator()
validator.AllowGaps = true
validator.MaxGapDuration = 5 * time.Minute

candles := []*market.Candle{
    {Timestamp: t0, ...},              // 00:00
    {Timestamp: t0.Add(10*time.Minute), ...},  // 00:10 (10min gap)
}
```

**Result:**
```
✓ Validation passed (with warnings)
  Total candles: 2
  Valid candles: 2
  Errors: 0
  Warnings: 1
  Gaps: 1
  
  ⚠ Gap: from 2024-01-01T00:00:00Z to 2024-01-01T00:10:00Z (duration: 10m0s)
```

---

## Best Practices

### 1. Always Validate Before Backtesting

```go
// ✗ Bad: Skip validation
candles, _ := reader.Read()
runBacktest(candles)  // May produce incorrect results!

// ✓ Good: Validate first
candles, _ := reader.Read()
report, _ := validator.Validate(candles)
if !report.Valid {
    log.Fatal("Invalid data")
}
runBacktest(candles)
```

### 2. Configure for Your Market

```go
// Crypto (24/7)
validator.AllowGaps = false
validator.MaxGapDuration = 5 * time.Minute

// Stocks (business days only)
validator.AllowGaps = true
validator.MaxGapDuration = 72 * time.Hour
```

### 3. Log Validation Results

```go
report, _ := validator.Validate(candles)

log.Printf("Validation: %s", report.Summary)
if report.HasWarnings() {
    log.Printf("  Warnings: %d", report.WarningCount())
}
if report.HasErrors() {
    log.Printf("  Errors: %d", report.ErrorCount())
}
```

### 4. Handle Gaps Appropriately

```go
// For production: strict validation
validator.AllowGaps = false

// For research: lenient validation
validator.AllowGaps = true
if report.HasWarnings() {
    log.Printf("Note: data has %d gaps", len(report.Gaps))
}
```

### 5. Archive Validation Reports

```go
// Save validation report with backtest results
type BacktestMetadata struct {
    Strategy        string
    DataFile        string
    ValidationValid bool
    ValidationErrors int
    ValidationWarnings int
    // ... backtest results
}
```

---

## Troubleshooting

### Problem: "high < low"

**Cause:** Data corruption or incorrect parsing

**Solution:**
1. Check data source
2. Verify file integrity
3. Re-download data
4. Check parser configuration

### Problem: "out of order timestamps"

**Cause:** Data not sorted or merged incorrectly

**Solution:**
1. Sort data by timestamp before validation
2. Check data source sorting
3. Fix merge logic

```go
// Sort candles before validation
sort.Slice(candles, func(i, j int) bool {
    return candles[i].Timestamp.Before(candles[j].Timestamp)
})
```

### Problem: "many gaps detected"

**Cause:** Missing data from source

**Solution:**
1. If expected (weekends): set `AllowGaps = true`
2. If unexpected: fill gaps or re-download data
3. Consider different data source

### Problem: "validation too slow"

**Cause:** Large dataset

**Solution:**
1. Validation is O(n) - fast enough for most use cases
2. For very large datasets: sample validation
3. Validate once, cache results

---

## Integration with CLI (Future)

```bash
# Validate data file
trader validate-data --file data.parquet \
                     --allow-gaps \
                     --max-gap 1h

# Backtest with automatic validation
trader backtest --strategy strategy.yaml \
                --data data.parquet \
                --validate \
                --strict
```

---

## Testing Your Data

### Quick Check Script

```go
package main

import (
    "log"
    "github.com/ZulferDev/smallbt_go/internal/data/parquet"
    "github.com/ZulferDev/smallbt_go/internal/data/validation"
)

func main() {
    // Read data
    reader, _ := parquet.NewParquetReader("data.parquet")
    candles, _ := reader.Read()
    reader.Close()
    
    // Validate
    validator := validation.NewDefaultValidator()
    report, _ := validator.Validate(candles)
    
    // Print report
    log.Println(report.String())
    
    // Exit with error if invalid
    if !report.Valid {
        log.Fatal("Validation failed")
    }
}
```

---

## Performance

### Validation Speed

| Dataset Size | Validation Time |
|--------------|-----------------|
| 1K candles | ~1ms |
| 10K candles | ~10ms |
| 100K candles | ~100ms |
| 1M candles | ~1s |

**Note:** Validation is O(n) and very fast. Always validate!

---

## Summary

### Key Points

✅ **Always validate data before backtesting**
✅ **Configure validator for your market type**
✅ **Check ValidationReport.Valid before continuing**
✅ **Log validation errors for debugging**
✅ **Handle gaps appropriately (strict vs lenient)**

### Validation Checklist

- [ ] OHLC consistency checked
- [ ] Value validity checked (no negatives/zeros)
- [ ] Chronological ordering verified
- [ ] Duplicates detected
- [ ] Gaps handled (error or warning)
- [ ] ValidationReport inspected
- [ ] Errors logged/fixed

### When to Validate

✅ After reading from CSV/Parquet
✅ After downloading from exchange
✅ After data transformation
✅ Before every backtest
✅ In production pipeline

❌ Not needed after validation (cache results)
❌ Not needed for trusted, pre-validated data

---

**Version:** Phase 17 Week 1  
**Last Updated:** 2026-09-05  
**Status:** Production Ready
