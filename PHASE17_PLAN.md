# Phase 17 - Enhanced Data Handling Plan

**Date:** 2026-09-05  
**Phase:** 17 (Enhanced Data Handling)  
**Priority:** Critical Path (v1.0.0)  
**Estimated Duration:** 2-3 weeks (16-24 hours)  
**Status:** Planning  

---

## Executive Summary

Phase 17 focuses on enhanced data handling capabilities to support production-ready quantitative trading. This includes Parquet file support for efficient storage, market gap handling, data validation pipeline, and multi-timeframe synchronization.

**Key Objectives:**
- ✅ Parquet file format support (efficient storage)
- ✅ Comprehensive data validation pipeline
- ✅ Market gap detection and handling
- ✅ Multi-timeframe data synchronization
- ✅ Data integrity guarantees

---

## Context

### Phase 16 Achievements
- Production-ready paper trading (PaperBroker)
- Real-time WebSocket data feed
- CLI integration with WebSocket
- 20/20 packages passing, zero regressions

### Phase 17 Focus
Move from CSV-only data to production-grade data handling with multiple formats, validation, and synchronization capabilities.

---

## Technical Requirements

### 1. Parquet Support

**Why Parquet:**
- Columnar storage (efficient queries)
- Built-in compression (smaller files)
- Type safety (schema enforcement)
- Industry standard (compatible with data pipelines)
- Fast aggregations (perfect for backtesting)

**Implementation:**
```go
// internal/data/parquet/reader.go
type ParquetReader struct {
    path   string
    file   *parquet.File
    schema *parquet.Schema
}

func NewParquetReader(path string) (*ParquetReader, error)
func (r *ParquetReader) Read() ([]*market.Candle, error)
func (r *ParquetReader) ReadRange(start, end time.Time) ([]*market.Candle, error)
func (r *ParquetReader) Close() error
```

**Schema:**
```
message Candle {
    required int64  timestamp;  // Unix milliseconds
    required double open;
    required double high;
    required double low;
    required double close;
    required double volume;
    optional string symbol;
}
```

### 2. Data Validation Pipeline

**Validation Rules:**
- ✅ Chronological ordering (timestamps increasing)
- ✅ OHLC consistency (high >= open, high >= close, low <= open, low <= close)
- ✅ No duplicate timestamps
- ✅ Volume >= 0
- ✅ Price > 0
- ✅ Gap detection (configurable threshold)

**Implementation:**
```go
// internal/data/validation/validator.go
type Validator interface {
    Validate(candles []*market.Candle) (*ValidationReport, error)
}

type ValidationReport struct {
    Valid            bool
    Errors           []ValidationError
    Warnings         []ValidationWarning
    TotalCandles     int
    ValidCandles     int
    GapsDetected     int
    DuplicatesFound  int
}

type ValidationError struct {
    Index   int
    Candle  *market.Candle
    Type    ErrorType
    Message string
}
```

### 3. Gap Handling

**Gap Types:**
- **Small gaps** (<1 hour): Warning only
- **Medium gaps** (1-24 hours): Configurable (warn/error/fill)
- **Large gaps** (>24 hours): Always error unless explicitly allowed

**Handling Strategies:**
```yaml
data:
  validation:
    allow_gaps: true
    max_gap_minutes: 60
    
  gap_handling:
    strategy: forward_fill  # forward_fill | error | skip | interpolate
    max_fill_candles: 5
```

**Implementation:**
```go
// internal/data/gap/handler.go
type GapHandler interface {
    DetectGaps(candles []*market.Candle, threshold time.Duration) []Gap
    FillGaps(candles []*market.Candle, strategy FillStrategy) ([]*market.Candle, error)
}

type Gap struct {
    StartIndex int
    EndIndex   int
    StartTime  time.Time
    EndTime    time.Time
    Duration   time.Duration
}

type FillStrategy int
const (
    FillStrategyError FillStrategy = iota
    FillStrategyForwardFill
    FillStrategySkip
    FillStrategyInterpolate
)
```

### 4. Multi-Timeframe Synchronization

**Use Case:**
Strategy uses:
- 1-hour candles for signals
- 4-hour candles for trend confirmation
- Daily candles for overall market context

**Synchronization Requirements:**
- Align candles to common timestamps
- Prevent look-ahead bias (4H candle[t] available only after 4H completes)
- Handle missing data in higher timeframes
- Efficient storage (don't duplicate data)

**Implementation:**
```go
// internal/data/multiframe/synchronizer.go
type Synchronizer struct {
    primaryTF   Timeframe
    higherTFs   []Timeframe
    alignment   AlignmentStrategy
}

func (s *Synchronizer) Sync(feeds map[Timeframe][]*Candle) (*SyncedData, error)

type SyncedData struct {
    Timestamps []time.Time
    Data       map[Timeframe][]*Candle  // Aligned by timestamp
}
```

---

## Architecture

### Package Structure

```
internal/data/
├── parquet/
│   ├── reader.go          (Parquet file reader)
│   ├── writer.go          (Parquet file writer)
│   ├── schema.go          (Parquet schema definition)
│   └── reader_test.go
│
├── validation/
│   ├── validator.go       (Validation interface + implementation)
│   ├── rules.go           (Validation rules: OHLC, ordering, etc.)
│   ├── report.go          (ValidationReport structure)
│   └── validator_test.go
│
├── gap/
│   ├── detector.go        (Gap detection)
│   ├── handler.go         (Gap handling strategies)
│   └── gap_test.go
│
└── multiframe/
    ├── synchronizer.go    (Multi-timeframe sync)
    ├── alignment.go       (Alignment strategies)
    └── synchronizer_test.go
```

### Data Flow

```
Parquet File → ParquetReader
    ↓
Raw Candles
    ↓
Validator (rules check)
    ↓
ValidationReport (errors/warnings)
    ↓
GapHandler (if gaps found)
    ↓
Validated Candles
    ↓
Synchronizer (if multi-timeframe)
    ↓
SyncedData (aligned by timestamp)
    ↓
Backtest Engine / Strategy
```

---

## Weekly Breakdown

### Week 1: Parquet Support + Data Validation (8 hours)

**Day 1 (2h): Parquet Reader**
- Implement ParquetReader
- Schema definition
- Read() and ReadRange() methods
- Unit tests (10+ tests)

**Day 2 (2h): Parquet Writer**
- Implement ParquetWriter
- Write() method with schema validation
- Integration with existing CSV data
- Round-trip tests (CSV → Parquet → Candles)

**Day 3 (2h): Data Validation Pipeline**
- Validator interface
- Validation rules (OHLC, ordering, duplicates)
- ValidationReport structure
- Unit tests (15+ tests)

**Day 4 (2h): CLI Integration**
- Update `trader backtest` to support Parquet
- Add `--validate` flag for data validation
- Validation report output
- Documentation update

**Deliverables:**
- ParquetReader + ParquetWriter (200-300 lines)
- Validator (150-200 lines)
- Tests (400+ lines)
- CLI integration (50 lines)
- Documentation updates

### Week 2: Gap Handling + Multi-Timeframe (8 hours)

**Day 1 (2h): Gap Detection**
- GapDetector implementation
- DetectGaps() with configurable threshold
- Gap reporting
- Unit tests (10+ tests)

**Day 2 (2h): Gap Handling**
- GapHandler with multiple strategies
- Forward fill implementation
- Skip and error strategies
- Integration tests

**Day 3 (2h): Multi-Timeframe Synchronizer**
- Synchronizer implementation
- Alignment strategies
- Look-ahead bias prevention
- Unit tests (12+ tests)

**Day 4 (2h): Integration + Documentation**
- CLI integration for gap handling
- Multi-timeframe strategy example
- Complete user guide
- Verification tests

**Deliverables:**
- GapDetector + GapHandler (150-200 lines)
- Synchronizer (200-250 lines)
- Tests (500+ lines)
- Example multi-timeframe strategy
- Documentation

---

## Success Criteria

### Functional Requirements ✅

1. **Parquet Support:**
   - ✅ Read Parquet files with OHLCV schema
   - ✅ Write Parquet files from candle data
   - ✅ Performance: 10x faster than CSV for large files
   - ✅ Round-trip: CSV → Parquet → Candles preserves data

2. **Data Validation:**
   - ✅ Detect OHLC inconsistencies
   - ✅ Detect chronological issues
   - ✅ Detect duplicates
   - ✅ Detect invalid values (negative, zero)
   - ✅ Produce actionable ValidationReport

3. **Gap Handling:**
   - ✅ Detect gaps with configurable threshold
   - ✅ Support multiple handling strategies
   - ✅ Forward fill without look-ahead bias
   - ✅ Clear gap reporting

4. **Multi-Timeframe:**
   - ✅ Synchronize multiple timeframes
   - ✅ Prevent look-ahead bias
   - ✅ Handle missing data gracefully
   - ✅ Efficient storage

### Quality Requirements ✅

- ✅ All tests passing (30+ new tests)
- ✅ Zero regressions (20/20 packages)
- ✅ Comprehensive documentation
- ✅ Example strategies for each feature

### Performance Requirements ✅

- ✅ Parquet reads 10x faster than CSV
- ✅ Validation overhead <5% of read time
- ✅ Gap handling <1ms per gap
- ✅ Multi-timeframe sync <10ms for 10k candles

---

## Risk Assessment

### Technical Risks

**1. Parquet Library Dependency**
- Risk: External dependency (github.com/xitongsys/parquet-go)
- Mitigation: Well-maintained library, Apache Parquet standard
- Fallback: Can always use CSV as backup

**2. Look-Ahead Bias in Multi-Timeframe**
- Risk: Accidentally using future data from higher timeframes
- Mitigation: Strict timestamp alignment tests
- Validation: Comprehensive test suite with known edge cases

**3. Gap Handling Correctness**
- Risk: Forward fill introduces artificial data
- Mitigation: Clear documentation of limitations
- Alternative: Provide multiple strategies (skip, error)

### Schedule Risks

**1. Parquet Integration Complexity**
- Risk: Schema mapping may be complex
- Mitigation: Start simple (OHLCV only), extend later
- Buffer: Week 2 can absorb overflow

**2. Multi-Timeframe Testing**
- Risk: Edge cases may be hard to test
- Mitigation: Use golden test data with known results
- Buffer: 2-day buffer built into schedule

---

## Testing Strategy

### Unit Tests (30+ tests)

**ParquetReader (10 tests):**
- Read valid Parquet file
- Read empty file
- Read file with invalid schema
- ReadRange with valid range
- ReadRange with out-of-bounds
- Close after read
- Error handling

**Validator (15 tests):**
- Valid OHLC candles
- Invalid OHLC (high < low)
- Chronological ordering
- Duplicate timestamps
- Negative volume
- Zero/negative prices
- Empty candle list
- Single candle validation

**GapHandler (10 tests):**
- Detect small gaps
- Detect large gaps
- Forward fill strategy
- Skip strategy
- Error strategy
- No gaps detected
- Multiple gaps handling

**Synchronizer (12 tests):**
- Sync two timeframes
- Sync three timeframes
- Handle missing higher-TF data
- Prevent look-ahead bias
- Empty data handling
- Misaligned timestamps

### Integration Tests (5+ tests)

**End-to-End:**
- CSV → Parquet → Backtest
- Multi-file Parquet read
- Validation → Gap handling → Backtest
- Multi-timeframe strategy execution

---

## Documentation Plan

### User Documentation

**1. Data Formats Guide (data_formats.md)**
- Supported formats (CSV, Parquet)
- Format comparison table
- When to use each format
- Migration guide (CSV → Parquet)

**2. Data Validation Guide (data_validation.md)**
- Why validate data
- Validation rules explained
- Interpreting ValidationReport
- Common issues and fixes

**3. Gap Handling Guide (gap_handling.md)**
- What are gaps
- Gap detection configuration
- Handling strategies comparison
- Best practices

**4. Multi-Timeframe Guide (multi_timeframe.md)**
- Multi-timeframe concepts
- Strategy example (1H + 4H + 1D)
- Look-ahead bias prevention
- Performance considerations

### Technical Documentation

**Daily Reports (8 documents):**
- Week 1: Day 1-4 reports
- Week 2: Day 1-4 reports

**Completion Reports (2 documents):**
- Week 1 completion report
- Week 2 completion report

**Verification Reports (2 documents):**
- Week 1 verification report
- Week 2 verification report

**Handoff Document:**
- Phase 17 complete handoff

---

## Dependencies

### External Libraries

**Parquet:**
```go
github.com/xitongsys/parquet-go v1.6.2
github.com/xitongsys/parquet-go-source v0.0.0-20220315005136-aec0fe3e777c
```

**Testing:**
- Standard Go testing
- Existing test infrastructure

### Internal Dependencies

**Existing Modules:**
- internal/market (Candle structure)
- internal/data/csv (existing CSV reader)
- internal/backtest (backtest engine)

**No Modifications Required:**
- Phase 16 code (PaperBroker, WebSocketFeed)
- Phase 1-15 code (indicators, analytics, etc.)

---

## Success Metrics

### Code Metrics

| Metric | Target |
|--------|--------|
| Production code | 700-900 lines |
| Test code | 900-1,200 lines |
| Documentation | 1,500+ lines |
| Test coverage | >80% |

### Quality Metrics

| Metric | Target |
|--------|--------|
| Tests passing | 20/20 packages |
| New tests | 30+ |
| Regressions | 0 |
| Performance | Parquet 10x faster than CSV |

### Delivery Metrics

| Metric | Target |
|--------|--------|
| Duration | 16 hours (2 weeks) |
| Schedule variance | <10% |
| Documentation | Complete |
| Commits | Clean history |

---

## Next Steps After Phase 17

### Phase 18: Exchange Integration (Binance)
- Binance REST API client
- Binance WebSocket client
- Authentication (API key, signature)
- Order execution on real exchange
- Live trading mode

### Phase 19: Risk Management Enhancement
- Position size limits
- Daily loss limits
- Kill switches
- Manual approval for large orders
- Portfolio-level risk

### Phase 20: Strategy Evaluation Integration
- Real-time indicator calculation with WebSocket
- Automatic signal generation
- Entry/exit automation
- Strategy state management

---

## Conclusion

Phase 17 delivers production-ready data handling capabilities:
- Parquet support for efficient storage
- Comprehensive data validation
- Market gap detection and handling
- Multi-timeframe synchronization

These features are critical for v1.0.0 production release and enable advanced quantitative research workflows.

---

**Status:** ✅ PLANNED  
**Ready to start:** Yes  
**Estimated duration:** 2 weeks (16 hours)  
**Target completion:** 2026-09-19
