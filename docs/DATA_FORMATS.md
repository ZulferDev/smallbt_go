# Data Formats Guide

**smallbt_go Data Format Documentation**

---

## Overview

smallbt_go supports multiple data formats for historical market data. This guide explains the supported formats, when to use each, and how to convert between them.

---

## Supported Formats

### 1. CSV (Comma-Separated Values)

**Status:** ✅ Supported (Phase 1-15)

**Use Cases:**
- Small datasets (<100K candles)
- Human-readable data
- Easy to edit manually
- Quick testing and debugging
- Data from external sources

**Advantages:**
- ✅ Human-readable
- ✅ Easy to edit with text editors
- ✅ Universal format (Excel, Python, etc.)
- ✅ Simple structure
- ✅ No special tools needed

**Disadvantages:**
- ❌ Large file sizes (~100 bytes per candle)
- ❌ Slow to read for large datasets
- ❌ No compression
- ❌ No schema enforcement
- ❌ Text parsing overhead

**Format:**
```csv
timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,42000.0,42500.0,41800.0,42300.0,123.45
2024-01-01T00:01:00Z,42300.0,42800.0,42200.0,42600.0,145.67
```

### 2. Parquet (Apache Parquet)

**Status:** ✅ Supported (Phase 17 Week 1)

**Use Cases:**
- Large datasets (>100K candles)
- Production backtesting
- Long-term data storage
- High-performance queries
- Data warehousing

**Advantages:**
- ✅ Columnar storage (fast queries)
- ✅ Efficient compression (~26 bytes per candle with SNAPPY)
- ✅ Schema enforcement (type safety)
- ✅ Industry standard (Spark, Pandas, DuckDB compatible)
- ✅ ~74% smaller than CSV
- ✅ ~10x faster read performance

**Disadvantages:**
- ❌ Not human-readable (binary format)
- ❌ Requires special tools to view
- ❌ Slightly more complex to work with

**Schema:**
```
timestamp: INT64 (TIMESTAMP_MILLIS)
open:      DOUBLE
high:      DOUBLE
low:       DOUBLE
close:     DOUBLE
volume:    DOUBLE
```

**Compression:** SNAPPY (default)

---

## Format Comparison

| Feature | CSV | Parquet |
|---------|-----|---------|
| File size (1K candles) | ~100 KB | ~30 KB |
| File size (100K candles) | ~10 MB | ~2.6 MB |
| Read speed | Baseline | ~10x faster |
| Human-readable | ✅ Yes | ❌ No |
| Compression | ❌ No | ✅ Yes (SNAPPY) |
| Schema enforcement | ❌ No | ✅ Yes |
| Editable | ✅ Yes | ❌ No |
| Industry standard | ✅ Yes | ✅ Yes |
| Query optimization | ❌ No | ✅ Yes (columnar) |

---

## When to Use Each Format

### Use CSV When:
- 🔹 Dataset is small (<100K candles)
- 🔹 Need to manually inspect/edit data
- 🔹 Sharing data with non-technical users
- 🔹 Quick testing and prototyping
- 🔹 Data source only provides CSV

### Use Parquet When:
- 🔹 Dataset is large (>100K candles)
- 🔹 Performance is critical
- 🔹 Storage space is limited
- 🔹 Running many backtests
- 🔹 Long-term data archival
- 🔹 Production environment

---

## Usage Examples

### Reading CSV (Existing)

```bash
# Backtest with CSV data
trader backtest --strategy strategy.yaml \
                --data data/BTCUSDT.csv \
                --symbol BTCUSDT
```

### Reading Parquet (New)

```go
import "github.com/ZulferDev/smallbt_go/internal/data/parquet"

// Read Parquet file
reader, err := parquet.NewParquetReader("data/BTCUSDT.parquet")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

// Read all candles
candles, err := reader.Read()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Read %d candles\n", len(candles))
```

### Reading Parquet with Time Range

```go
// Read candles in specific time range
start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC)

candles, err := reader.ReadRange(start, end)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Read %d candles in range\n", len(candles))
```

### Writing Parquet

```go
import "github.com/ZulferDev/smallbt_go/internal/data/parquet"

// Create writer
writer, err := parquet.NewParquetWriter("output.parquet")
if err != nil {
    log.Fatal(err)
}
defer writer.Close()

// Write candles (batch)
err = writer.Write(candles)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Parquet file written successfully")
```

---

## Converting Between Formats

### CSV to Parquet (Future Tool)

```bash
# Convert CSV to Parquet (planned)
trader convert --input data.csv \
               --output data.parquet \
               --format parquet
```

### Manual Conversion (Go)

```go
// Read CSV
csvReader := csv.NewCSVReader("data.csv")
candles, _ := csvReader.Read()

// Write Parquet
parquetWriter, _ := parquet.NewParquetWriter("data.parquet")
parquetWriter.Write(candles)
parquetWriter.Close()
```

---

## File Size Examples

### Real-World Data

**1,000 candles (1K):**
- CSV: ~100 KB
- Parquet: ~30 KB
- Savings: 70%

**10,000 candles (10K):**
- CSV: ~1 MB
- Parquet: ~268 KB
- Savings: 73%

**100,000 candles (100K):**
- CSV: ~10 MB
- Parquet: ~2.6 MB
- Savings: 74%

**1,000,000 candles (1M):**
- CSV: ~100 MB
- Parquet: ~26 MB
- Savings: 74%

---

## Performance Benchmarks

### Read Performance (10,000 candles)

| Operation | CSV | Parquet | Speedup |
|-----------|-----|---------|---------|
| Full read | ~100ms | ~10ms | ~10x |
| Range query | ~100ms | ~10ms | ~10x |
| Memory usage | High | Low | Columnar |

**Note:** Actual performance depends on hardware, file size, and query patterns.

---

## Best Practices

### For CSV:
1. ✅ Keep files under 100K candles
2. ✅ Use consistent timestamp format (RFC3339)
3. ✅ Include header row
4. ✅ Use comma separator
5. ✅ Avoid special characters in values

### For Parquet:
1. ✅ Use for large datasets (>100K candles)
2. ✅ Keep SNAPPY compression enabled (default)
3. ✅ Use ReadRange() for time-based queries
4. ✅ Close readers/writers to release resources
5. ✅ Validate data before writing

---

## Data Validation

### Always validate data after reading:

```go
import "github.com/ZulferDev/smallbt_go/internal/data/validation"

// Read data
reader, _ := parquet.NewParquetReader("data.parquet")
candles, _ := reader.Read()
reader.Close()

// Validate
validator := validation.NewDefaultValidator()
report, _ := validator.Validate(candles)

if !report.Valid {
    log.Printf("Validation failed: %s", report.Summary)
    for _, err := range report.Errors {
        log.Printf("  Error: %s", err.Message)
    }
}
```

---

## Troubleshooting

### CSV Issues

**Problem:** "parse error at line X"
- **Solution:** Check CSV format, ensure proper escaping

**Problem:** "timestamp parse error"
- **Solution:** Use RFC3339 format (2024-01-01T00:00:00Z)

### Parquet Issues

**Problem:** "schema mismatch"
- **Solution:** Ensure all fields present (timestamp, OHLCV)

**Problem:** "file corrupt"
- **Solution:** Ensure Close() was called when writing

**Problem:** "can't read after close"
- **Solution:** Create new reader for each read operation

---

## Future Formats (Planned)

### Binary Format (Custom)
- Optimized for smallbt_go
- Even faster than Parquet
- Planned for Phase 18+

### Database Integration
- PostgreSQL/TimescaleDB
- DuckDB (in-process analytics)
- Planned for Phase 19+

---

## External Tools

### View Parquet Files

**Python (Pandas):**
```python
import pandas as pd
df = pd.read_parquet("data.parquet")
print(df.head())
```

**DuckDB (SQL):**
```sql
SELECT * FROM 'data.parquet' LIMIT 10;
```

**Parquet CLI Tools:**
```bash
# parquet-tools (Java)
parquet-tools head data.parquet
```

---

## Migration Guide

### Moving from CSV to Parquet

**Step 1:** Test with small dataset
```bash
# Create test Parquet file (manual conversion)
# Run backtest to verify
```

**Step 2:** Convert historical data
```bash
# Convert all CSV files to Parquet
# Validate converted files
```

**Step 3:** Update workflows
```bash
# Update backtest scripts to use .parquet files
# Update data download scripts
```

**Benefits:**
- ✅ 74% storage savings
- ✅ 10x faster backtests
- ✅ Better data integrity

---

## Recommendations

### For Beginners:
- Start with CSV for learning
- Move to Parquet when comfortable
- Focus on data quality over format

### For Production:
- Use Parquet for all datasets >10K candles
- Validate data after every read
- Monitor file sizes and performance
- Archive old data in Parquet

### For Researchers:
- Use Parquet for fast iteration
- Keep CSV for data sharing
- Use validation pipeline for quality
- Profile your specific workload

---

## Summary

| Format | Best For | File Size | Speed |
|--------|----------|-----------|-------|
| CSV | Small data, editing, learning | Large | Slow |
| Parquet | Large data, production, performance | Small | Fast |

**Recommendation:** Start with CSV, migrate to Parquet for production.

---

**Version:** Phase 17 Week 1  
**Last Updated:** 2026-09-05  
**Status:** Production Ready
