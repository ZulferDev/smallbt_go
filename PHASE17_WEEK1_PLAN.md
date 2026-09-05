# Phase 17 Week 1 - Plan

**Date:** 2026-09-05  
**Phase:** 17 (Enhanced Data Handling)  
**Week:** 1 (Parquet Support + Data Validation)  
**Duration:** 4 days (8 hours total)  
**Status:** Planning  

---

## Week 1 Objectives

Implement Parquet file format support and comprehensive data validation pipeline.

**Key Deliverables:**
- ✅ ParquetReader with schema enforcement
- ✅ ParquetWriter for efficient storage
- ✅ Data validation pipeline (OHLC, ordering, duplicates)
- ✅ CLI integration (--format parquet, --validate flags)
- ✅ 25+ unit tests
- ✅ Documentation

---

## Daily Breakdown

### Day 1 (2h): Parquet Reader Implementation

**Objective:** Implement ParquetReader to read OHLCV data from Parquet files.

**Tasks:**
1. Add Parquet dependencies to go.mod
2. Define Parquet schema for OHLCV data
3. Implement ParquetReader struct and methods
4. Add unit tests (10+ tests)
5. Create daily report

**Deliverables:**
- `internal/data/parquet/reader.go` (~150 lines)
- `internal/data/parquet/schema.go` (~50 lines)
- `internal/data/parquet/reader_test.go` (~200 lines)
- Day 1 report (~300 lines)

**Tests:**
- Read valid Parquet file
- Read empty file
- Read file with invalid schema
- ReadRange with valid range
- ReadRange with out-of-bounds
- Error handling (file not found, corrupt file)
- Close after read
- Multiple reads

**Success Criteria:**
- ✅ ParquetReader reads OHLCV data correctly
- ✅ 10+ tests passing
- ✅ No regressions (20/20 packages)

---

### Day 2 (2h): Parquet Writer Implementation

**Objective:** Implement ParquetWriter to write candle data to Parquet files.

**Tasks:**
1. Implement ParquetWriter struct and methods
2. Schema validation on write
3. Integration with existing CSV data
4. Round-trip tests (CSV → Parquet → Candles)
5. Performance benchmarks
6. Create daily report

**Deliverables:**
- `internal/data/parquet/writer.go` (~100 lines)
- `internal/data/parquet/writer_test.go` (~150 lines)
- Day 2 report (~300 lines)

**Tests:**
- Write valid candle data
- Write empty candle list
- Write with schema validation
- Round-trip: Write → Read → Compare
- Performance benchmark vs CSV
- Error handling (write failures)

**Success Criteria:**
- ✅ ParquetWriter writes valid Parquet files
- ✅ Round-trip preserves data accuracy
- ✅ 10x performance improvement over CSV
- ✅ No regressions (20/20 packages)

---

### Day 3 (2h): Data Validation Pipeline

**Objective:** Implement comprehensive data validation pipeline.

**Tasks:**
1. Implement Validator interface
2. Validation rules: OHLC, ordering, duplicates, invalid values
3. ValidationReport structure
4. Unit tests (15+ tests)
5. Create daily report

**Deliverables:**
- `internal/data/validation/validator.go` (~150 lines)
- `internal/data/validation/rules.go` (~100 lines)
- `internal/data/validation/report.go` (~50 lines)
- `internal/data/validation/validator_test.go` (~300 lines)
- Day 3 report (~300 lines)

**Validation Rules:**
- OHLC consistency (high >= open, high >= close, low <= open, low <= close)
- Chronological ordering (timestamps increasing)
- No duplicate timestamps
- Volume >= 0
- Price > 0
- No NaN or Inf values

**Tests:**
- Valid OHLC candles
- Invalid OHLC (high < low, etc.)
- Out-of-order timestamps
- Duplicate timestamps
- Negative volume
- Zero/negative prices
- NaN/Inf values
- Empty candle list
- Single candle validation
- Large dataset validation

**Success Criteria:**
- ✅ All validation rules implemented
- ✅ ValidationReport provides actionable feedback
- ✅ 15+ tests passing
- ✅ No regressions (20/20 packages)

---

### Day 4 (2h): CLI Integration & Documentation

**Objective:** Integrate Parquet and validation into CLI, complete documentation.

**Tasks:**
1. Update `trader backtest` to support --format parquet
2. Add --validate flag for data validation
3. Validation report output formatting
4. Create data_formats.md guide
5. Create data_validation.md guide
6. Create daily report

**Deliverables:**
- `cmd/trader/main.go` updates (~50 lines)
- `docs/DATA_FORMATS.md` (~200 lines)
- `docs/DATA_VALIDATION.md` (~200 lines)
- Day 4 report (~300 lines)

**CLI Changes:**
```bash
# Read Parquet file
trader backtest --strategy strategy.yaml \
                --data data.parquet \
                --format parquet

# Validate data before backtest
trader backtest --strategy strategy.yaml \
                --data data.csv \
                --validate

# Standalone validation
trader validate --data data.csv
trader validate --data data.parquet --format parquet
```

**Documentation:**
- DATA_FORMATS.md: CSV vs Parquet comparison, migration guide
- DATA_VALIDATION.md: Validation rules, troubleshooting

**Success Criteria:**
- ✅ CLI supports Parquet files
- ✅ Validation integrated into workflow
- ✅ Complete documentation
- ✅ No regressions (20/20 packages)

---

## Week 1 Metrics

### Code Deliverables

| Component | Lines | Tests | Total |
|-----------|-------|-------|-------|
| ParquetReader | 200 | 200 | 400 |
| ParquetWriter | 100 | 150 | 250 |
| Validator | 300 | 300 | 600 |
| CLI Integration | 50 | - | 50 |
| **Total** | **650** | **650** | **1,300** |

### Documentation Deliverables

| Document | Lines |
|----------|-------|
| Daily reports (4) | 1,200 |
| DATA_FORMATS.md | 200 |
| DATA_VALIDATION.md | 200 |
| **Total** | **1,600** |

### Quality Metrics

| Metric | Target |
|--------|--------|
| New tests | 35+ |
| Test pass rate | 100% |
| Packages passing | 20/20 |
| Regressions | 0 |
| Performance | Parquet 10x faster |

---

## Technical Specifications

### Parquet Schema

```go
// internal/data/parquet/schema.go
type CandleParquet struct {
    Timestamp int64   `parquet:"name=timestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
    Open      float64 `parquet:"name=open, type=DOUBLE"`
    High      float64 `parquet:"name=high, type=DOUBLE"`
    Low       float64 `parquet:"name=low, type=DOUBLE"`
    Close     float64 `parquet:"name=close, type=DOUBLE"`
    Volume    float64 `parquet:"name=volume, type=DOUBLE"`
    Symbol    string  `parquet:"name=symbol, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY, repetitiontype=OPTIONAL"`
}
```

### ParquetReader Interface

```go
// internal/data/parquet/reader.go
type ParquetReader struct {
    path       string
    fileReader source.ParquetFile
    reader     *reader.ParquetReader
}

func NewParquetReader(path string) (*ParquetReader, error)
func (r *ParquetReader) Read() ([]*market.Candle, error)
func (r *ParquetReader) ReadRange(start, end time.Time) ([]*market.Candle, error)
func (r *ParquetReader) Close() error
```

### Validator Interface

```go
// internal/data/validation/validator.go
type Validator interface {
    Validate(candles []*market.Candle) (*ValidationReport, error)
}

type ValidationReport struct {
    Valid            bool
    TotalCandles     int
    ValidCandles     int
    Errors           []ValidationError
    Warnings         []ValidationWarning
    GapsDetected     int
    DuplicatesFound  int
    Summary          string
}

type ValidationError struct {
    Index      int
    Timestamp  time.Time
    Type       ErrorType
    Message    string
    Severity   Severity
}

type ErrorType int
const (
    ErrorTypeOHLC ErrorType = iota
    ErrorTypeOrdering
    ErrorTypeDuplicate
    ErrorTypeInvalidValue
    ErrorTypeGap
)
```

---

## Dependencies

### External Libraries

**Parquet:**
```bash
go get github.com/xitongsys/parquet-go@v1.6.2
go get github.com/xitongsys/parquet-go-source@v0.0.0-20220315005136-aec0fe3e777c
```

**Existing:**
- Standard Go library
- internal/market (Candle structure)
- internal/data/csv (existing CSV reader)

---

## Risk Management

### Risk 1: Parquet Library Complexity
**Probability:** Medium  
**Impact:** Medium  
**Mitigation:** Start with simple schema, comprehensive tests  
**Fallback:** CSV remains as backup format  

### Risk 2: Performance Requirements
**Probability:** Low  
**Impact:** Medium  
**Mitigation:** Benchmark early (Day 2), columnar storage naturally fast  
**Target:** 10x improvement (realistic for Parquet vs CSV)  

### Risk 3: Schema Evolution
**Probability:** Low  
**Impact:** Low  
**Mitigation:** Version schema, support backward compatibility  
**Note:** Week 1 uses simple schema, extensions in future phases  

---

## Success Criteria

### Functional ✅
- ✅ ParquetReader reads OHLCV files correctly
- ✅ ParquetWriter writes valid Parquet files
- ✅ Round-trip preserves data
- ✅ Validator detects all rule violations
- ✅ CLI supports Parquet format

### Quality ✅
- ✅ 35+ tests passing
- ✅ Zero regressions
- ✅ Code formatted and linted
- ✅ Documentation complete

### Performance ✅
- ✅ Parquet read 10x faster than CSV
- ✅ Validation overhead <5% of read time
- ✅ No memory leaks

---

## Next Week Preview

### Week 2: Gap Handling + Multi-Timeframe

**Day 1:** Gap detection implementation  
**Day 2:** Gap handling strategies  
**Day 3:** Multi-timeframe synchronization  
**Day 4:** Integration and completion  

---

## Daily Schedule

| Day | Date | Focus | Hours | Deliverables |
|-----|------|-------|-------|--------------|
| 1 | 2026-09-05 | ParquetReader | 2h | Reader (200 lines) + Tests (200) |
| 2 | 2026-09-06 | ParquetWriter | 2h | Writer (100 lines) + Tests (150) |
| 3 | 2026-09-09 | Validator | 2h | Validator (300 lines) + Tests (300) |
| 4 | 2026-09-10 | CLI + Docs | 2h | CLI (50) + Docs (400) |

**Total:** 8 hours over 4 days

---

## Verification Plan

### Daily Verification
- Run full test suite after each day
- Verify zero regressions
- Check code formatting
- Review daily report

### Week-End Verification
- Comprehensive test run
- Performance benchmarks
- Documentation review
- Week completion report
- Week verification report

---

## Conclusion

Week 1 focuses on foundational data handling capabilities:
- Parquet format support for efficient storage
- Comprehensive data validation pipeline
- CLI integration for seamless user experience

These features enable production-ready quantitative research workflows and prepare for Week 2 (gap handling + multi-timeframe).

---

**Status:** ✅ PLANNED  
**Ready to start:** Yes  
**Start date:** 2026-09-05 (Today)  
**Target completion:** 2026-09-10
