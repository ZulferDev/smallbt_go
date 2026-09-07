# CSV Data Format

## Overview

The backtest engine supports flexible CSV formats for OHLCV (Open, High, Low, Close, Volume) market data. The CSV parser automatically detects column ordering from headers.

## Supported Formats

### Standard Format
```csv
timestamp,open,high,low,close,volume
2020-01-01 00:00:00,100.0,105.0,99.0,102.0,1000.0
2020-01-01 01:00:00,102.0,108.0,101.0,107.0,1200.0
```

### Flexible Column Ordering
Columns can appear in any order:
```csv
open,high,low,close,volume,timestamp
100.0,105.0,99.0,102.0,1000.0,2020-01-01 00:00:00
```

### Case-Insensitive Headers
```csv
TIMESTAMP,Open,HIGH,low,Close,VOLUME
2020-01-01 00:00:00,100.0,105.0,99.0,102.0,1000.0
```

### Alternative Header Names
Supported aliases:
- **Timestamp**: `timestamp`, `time`, `date`, `datetime`
- **Volume**: `volume`, `vol`

```csv
time,open,high,low,close,vol
2020-01-01 00:00:00,100.0,105.0,99.0,102.0,1000.0
```

## Timestamp Formats

The parser automatically detects various timestamp formats:

### With Milliseconds
```csv
timestamp,open,high,low,close,volume
2017-08-17 04:00:00.000,100.0,105.0,99.0,102.0,1000.0
2017-08-17 05:00:00.123,102.0,108.0,101.0,107.0,1200.0
```

### With Microseconds
```csv
2017-08-17 04:00:00.000000,100.0,105.0,99.0,102.0,1000.0
```

### Standard Formats
```csv
2020-01-01 00:00:00
2020-01-01T00:00:00
2020-01-01T00:00:00Z
2020-01-01
```

### Unix Timestamps
Both seconds and milliseconds are supported:
```csv
1609459200      # Seconds
1609459200000   # Milliseconds
```

## Required Columns

All CSV files must include these columns (in any order):
- `timestamp` (or aliases: time, date, datetime)
- `open`
- `high`
- `low`
- `close`
- `volume` (or alias: vol)

## Error Messages

### Missing Columns
```
Error: auto-detect columns: missing required columns: [volume]. 
Expected format: timestamp,open,high,low,close,volume. 
Got headers: [timestamp open high low close]
```

### Invalid OHLC
```
Error: row 5: invalid OHLC relationships
```
This occurs when:
- `high` < `low`
- `open` or `close` outside `[low, high]` range
- Negative prices or volume

### Chronological Order
```
Error: chronological validation failed: candle at index 10 is not after previous candle
```
Timestamps must be in ascending order.

## Best Practices

1. **Always include headers** - The first row should contain column names
2. **Use consistent timestamps** - All timestamps should be in the same timezone (UTC recommended)
3. **Validate data quality** - Ensure no missing values, invalid prices, or gaps
4. **Sort chronologically** - Data must be ordered by timestamp (oldest first)
5. **Precision** - Use appropriate decimal places for prices (typically 2-8 digits)

## Examples

### Binance-style Export
```csv
timestamp,open,high,low,close,volume
2020-01-01 00:00:00,30058.06,30125.59,29998.17,30043.17,962.38
2020-01-01 01:00:00,29910.75,30036.80,29775.11,29830.78,394.91
```

### Exchange with Different Order
```csv
date,vol,close,low,high,open
2020-01-01,1000.0,102.0,99.0,105.0,100.0
2020-01-02,1200.0,107.0,101.0,108.0,102.0
```

### High-Resolution Data
```csv
datetime,open,high,low,close,volume
2020-01-01 00:00:00.000,100.0,100.5,99.8,100.2,500.0
2020-01-01 00:01:00.000,100.2,100.7,100.0,100.5,600.0
```

## Technical Details

### Auto-Detection Algorithm

1. Parse CSV header row
2. Normalize header names (lowercase, trim whitespace)
3. Map each header to its column index
4. Validate all required columns are present
5. Use detected indices for parsing data rows

### Validation Checks

For each data row:
- Parse timestamp in detected format
- Parse OHLC values as float64
- Validate OHLC relationships: `low <= open <= high` and `low <= close <= high`
- Ensure positive volume
- Verify chronological ordering

### Performance

- CSV files are loaded entirely into memory
- Column detection happens once during initialization
- No runtime overhead from flexible formats
- Typical loading speed: ~100,000 rows/second

## Troubleshooting

### Problem: "unrecognized timestamp format"
**Solution**: Check that timestamps follow one of the supported formats. Consider using ISO 8601 format: `2020-01-01T00:00:00Z`

### Problem: "insufficient columns"
**Solution**: Ensure all data rows have the same number of columns as the header

### Problem: "invalid OHLC relationships"
**Solution**: Verify that high >= low, and open/close are within [low, high] range

### Problem: Parser treating timestamp as float
**Solution**: Ensure the CSV has a header row. If not, set `HasHeaders: false` in config (requires API usage)

## API Usage (Advanced)

For custom CSV formats, use the CSVConfig API:

```go
config := csv.CSVConfig{
    Symbol:       "BTCUSDT",
    Timeframe:    "1h",
    HasHeaders:   true,
    TimestampCol: 0,
    OpenCol:      1,
    HighCol:      2,
    LowCol:       3,
    CloseCol:     4,
    VolumeCol:    5,
    TimeFormat:   "", // Auto-detect
}

feed, err := csv.NewCSVFeed("data.csv", config)
```

For files without headers or non-standard formats, manually specify column indices and time format.
