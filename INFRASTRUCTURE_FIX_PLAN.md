# Infrastructure Issues Fix Plan

## Issue 1: Security Audit - gosec@v2 action not found

### Problem
```
##[error]Unable to resolve action `securego/gosec@v2`, unable to find version `v2`
```

### Root Cause
The `securego/gosec@v2` action version doesn't exist or has been deprecated. GitHub Actions marketplace shows the action but `v2` tag is invalid.

### Solutions (Priority Order)

#### Option 1: Use Latest Stable Version ⭐ RECOMMENDED
```yaml
- name: Run GoSec
  uses: securego/gosec@v2.21.4  # Use specific stable version
  with:
    args: ./...
```

**Pros:**
- Specific version, reproducible builds
- Known to work
- No breaking changes

**Cons:**
- Need to update manually for new versions

---

#### Option 2: Use Master Branch
```yaml
- name: Run GoSec
  uses: securego/gosec@master
  with:
    args: ./...
```

**Pros:**
- Always latest features
- No version tracking needed

**Cons:**
- Potential breaking changes
- Less stable

---

#### Option 3: Install gosec Binary
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: ${{ env.GO_VERSION }}

- name: Install gosec
  run: go install github.com/securego/gosec/v2/cmd/gosec@latest

- name: Run gosec
  run: gosec -fmt=sarif -out=gosec.sarif ./...
```

**Pros:**
- Full control over version
- Can customize format and output
- No dependency on GitHub Action

**Cons:**
- Slightly more complex
- Need to manage output format manually

---

#### Option 4: Disable Security Audit (NOT RECOMMENDED)
```yaml
# Comment out or remove Security Audit job
```

**Pros:**
- Immediate fix

**Cons:**
- Lose security scanning
- Not production-ready

---

### Recommended Action Plan

**Step 1:** Try Option 1 (specific version v2.21.4)
- If fails, check GitHub marketplace for latest version tag

**Step 2:** If Option 1 fails, use Option 3 (install binary)
- More reliable and flexible

**Step 3:** Add gosec config file `.gosec.json` to exclude false positives:
```json
{
  "exclude": [],
  "severity": "medium",
  "confidence": "medium",
  "exclude-generated": true
}
```

---

## Issue 2: MVP Validation - Go tar errors

### Problem
```
/usr/bin/tar: ../../../go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.4.linux-amd64/src/encoding/gob/error.go: Cannot open: File exists
```

### Root Cause
GitHub Actions Go cache corruption. Multiple concurrent jobs or previous failed jobs left cache in inconsistent state.

### Solutions (Priority Order)

#### Option 1: Disable Go Cache for MVP Validation ⭐ RECOMMENDED
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: ${{ env.GO_VERSION }}
    cache: false  # Disable cache to avoid tar errors
```

**Pros:**
- Immediate fix
- No cache corruption issues
- MVP Validation is fast enough without cache

**Cons:**
- Slightly slower (but negligible for this job)

---

#### Option 2: Use Different Cache Key
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: ${{ env.GO_VERSION }}
    cache: true
    cache-dependency-path: |
      go.sum
      go.mod
```

**Pros:**
- Still use caching
- Different key might avoid collision

**Cons:**
- May still have cache issues

---

#### Option 3: Clean Cache Before Setup
```yaml
- name: Clean Go cache
  run: |
    rm -rf ~/.cache/go-build
    rm -rf ~/go/pkg/mod

- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: ${{ env.GO_VERSION }}
    cache: true
```

**Pros:**
- Fresh start every time
- No cache corruption

**Cons:**
- Defeats purpose of caching
- Slower builds

---

#### Option 4: Use go-version from Matrix
```yaml
strategy:
  matrix:
    go-version: ['1.22', '1.23']
    
steps:
  - name: Set up Go
    uses: actions/setup-go@v5
    with:
      go-version: ${{ matrix.go-version }}
      cache: true
```

**Pros:**
- Test multiple Go versions
- Cache per version

**Cons:**
- More complex
- More CI minutes used

---

### Recommended Action Plan

**Step 1:** Implement Option 1 (disable cache for MVP Validation)
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: ${{ env.GO_VERSION }}
    cache: false  # MVP validation is fast, don't need cache
```

**Step 2:** Monitor for 3-5 runs to confirm fix

**Step 3:** If still fails, try Option 3 (clean cache)

**Step 4:** Keep cache enabled for other jobs (Test Coverage, Build) since they benefit more

---

## Implementation Priority

### High Priority (Fix Immediately)
1. ✅ Security Audit - Use gosec@v2.21.4 or install binary
2. ✅ MVP Validation - Disable cache

### Medium Priority (Monitor)
3. Static Analysis - Minor gosimple warnings (already mostly fixed)

### Low Priority (Optional)
4. Update Node.js actions to v24 (deprecation warnings)

---

## Testing Strategy

### Phase 1: Fix Security Audit
1. Update gosec action version
2. Commit and push
3. Monitor 1 CI run
4. If fails, switch to binary installation method

### Phase 2: Fix MVP Validation
1. Disable cache in MVP Validation job only
2. Commit and push
3. Monitor 2-3 CI runs to confirm stability
4. Document that cache is intentionally disabled

### Phase 3: Verify
1. Check 3 consecutive CI runs
2. All should show 7/8 or 8/8 jobs passing
3. Document any remaining minor issues

---

## Expected Outcome

**Before:** 6/8 jobs passing (75%)
- ✅ Test Coverage
- ✅ Build
- ✅ Integration Tests
- ✅ Performance Regression
- ✅ Documentation Check
- ⚠️ Static Analysis (minor warnings)
- ❌ Security Audit
- ❌ MVP Validation

**After Fixes:** 7-8/8 jobs passing (87.5-100%)
- ✅ Test Coverage
- ✅ Build
- ✅ Integration Tests
- ✅ Performance Regression
- ✅ Documentation Check
- ✅ Security Audit (FIXED)
- ✅ MVP Validation (FIXED)
- ⚠️ Static Analysis (acceptable minor warnings)

---

## Rollback Plan

If fixes cause new issues:

1. **Security Audit:** Revert to no security scan temporarily
2. **MVP Validation:** Keep cache disabled (no downside)
3. **Static Analysis:** Already in acceptable state

---

## Success Criteria

✅ Security Audit job passes  
✅ MVP Validation job passes  
✅ No new failures introduced  
✅ CI runtime not significantly increased (< 30 seconds)  
✅ All tests still pass

---

*Plan created: 2026-09-07*  
*Priority: HIGH - Blocks production readiness*  
*Estimated time: 30-45 minutes*
