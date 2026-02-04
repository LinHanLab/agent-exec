# Display Package Refactor Plan

## Overview

The `pkg/display` package provides event-driven console output formatting for the agent-exec CLI tool. This document outlines a refactoring plan to improve code organization, testability, maintainability, and robustness while **preserving all existing features**.

---

## Current State Analysis

### File Structure

| File | Size | Purpose |
|------|------|---------|
| `console.go` | ~75 lines | `JSONFormatter` (main formatter), `NewConsoleFormatter`, constants |
| `formatter.go` | ~53 lines | `Formatter` interface, `Display` coordinator, `EventFormatter` type, `FormatContext`, `eventFormatters` map |
| `handlers.go` | ~362 lines | All 24 individual `format*` functions + `GetColorForEventType` + `formatPrettyJSON` |
| `content.go` | ~95 lines | `ContentFilter`, `ToolInputFilter`, `defaultToolInputFilters` |
| `text.go` | ~276 lines | `TextFormatter`, `FormatContentWithFrame` (~190 lines complex function) |
| `color.go` | ~25 lines | ANSI color code constants |

### Test Coverage

| File | Test File | Coverage |
|------|-----------|----------|
| console.go | console_test.go | Partial (5 tests) |
| formatter.go | - | None |
| handlers.go | - | None (implicit via console tests) |
| content.go | content_test.go | Good |
| text.go | text_test.go | Good |

---

## Key Architectural Decisions

### Split Rather Than Merge

Unlike the original plan, this refactor **splits** `handlers.go` into logical files rather than merging. This follows the modularity principle:

**Original plan (problematic)**: Merge handlers.go (362 lines) into formatter.go (53 lines) → ~415 line file
**Refined approach**: Split handlers.go into category files → each ~50-80 lines, focused responsibility

### Justification

1. **handlers.go contains two distinct responsibilities**:
   - Event registry and color mapping (should stay together)
   - Individual formatter implementations (can be split by category)

2. **Large files hurt maintainability**:
   - Harder to review changes
   - Higher likelihood of merge conflicts
   - Violates single responsibility principle

3. **Testability improves with smaller modules**:
   - Can test formatter categories independently
   - Easier to add new formatters in the right place

4. **Note on formatter function size**: Each `format*` function is 5-15 lines. Splitting into 4+ files may create overhead. **Recommended: 2 files** (`events.go` for all formatters with section markers, or `events.go` + `events_git.go` for git-specific ones).

---

## Issues Identified (Prioritized)

### Critical (Must Fix)

#### 1. Implicit Type Assertions (Panic Risk)

**Problem**: All event formatters use implicit type assertions that panic on mismatch:
```go
func formatRunPromptStarted(event events.Event, ctx *FormatContext) (string, error) {
    data := event.Data.(events.RunPromptStartedData)  // PANICS on wrong type
```

**Impact**: Runtime panics if event data types mismatch. Should use the safe variant with error return.

**Fix**: Add `mustGetEventData` helper that panics with clear message (better for debugging than Go's default).

#### 2. Error Handling in Display.Start()

**Problem**: `Display.Start()` silently exits the processing loop on format errors:
```go
if err := d.formatter.Format(event); err != nil {
    return  // Silent failure, exits loop - LOST EVENTS!
}
```

**Impact**: If formatting fails, all subsequent events are lost. No notification to caller.

**Fix**: Don't exit loop on error; log to stderr and continue processing.

#### 3. FormatContentWithFrame Is Overly Complex

**Problem**: Single function ~190 lines handling:
- Empty content
- Frame borders (box drawing vs whitespace)
- Line wrapping with natural break points
- Color application
- Padding/spacing

**Impact**: Difficult to maintain, test, and modify.

**Fix**: Extract `FrameBuilder` to encapsulate frame construction logic.

### Important (Should Fix)

#### 4. Missing TextFormatter Interface

**Problem**: `TextFormatter` is a concrete struct with no interface:
```go
type TextFormatter struct {
    terminalWidth int
}
```

**Impact**: Cannot mock for testing or swap implementations (e.g., fixed-width for testing).

**Fix**: Add `TextFormatter` interface with existing struct implementing it.

#### 5. Global State Limits Testability

**Problem**: `eventFormatters` map and `defaultToolInputFilters` are global:
- Cannot test formatters in isolation with custom event mappings
- Cannot test content filtering with different filter rules
- Potential race conditions if modified at runtime

**Impact**: Makes isolated testing harder.

**Fix**: Extract registry and filters as injectable dependencies.

#### 6. Inconsistent Error Handling in Formatters

**Problem**: Some formatters return errors but the errors are not actionable:
- `formatClaudeToolUse` returns error from `formatPrettyJSON`
- Most other formatters don't return errors

**Impact**: Inconsistent API for formatters.

**Fix**: Either standardize error returns or document why some errors are possible and others aren't.

### Nice to Have (Consider)

#### 7. Color Constants Are Global

**Problem**: All color constants are exported package-level constants.

**Impact**: Cannot customize colors for different terminals/themes.

**Decision**: Low priority - colors are a theme concern, not functionality.

#### 8. Verbose Mode Passed Implicitly

**Problem**: `verbose` flag is passed via `FormatContext` but is a global concept.

**Impact**: Makes it harder to reason about verbose mode's effect.

**Decision**: Low priority - current design is acceptable for this use case.

---

## Refactoring Plan

### Phase 0: Baseline (Before Any Changes)

**Steps**:
1. Run existing tests: `go test ./pkg/display/... -v`
2. Record test coverage: `go test ./pkg/display/... -coverprofile=before.txt`
3. Verify all tests pass
4. Note any flaky tests

**Exit criteria**: Clean baseline established.

---

### Phase 1: Stabilization (No Behavior Changes)

**Goal**: Fix critical bugs and add safety checks without restructuring.

#### 1.1 Add Type-Safe Event Data Extraction Helper

**New file**: `events.go` (in display package)

```go
// mustGetEventData safely extracts typed data from an event.
// Panics with a clear message if type assertion fails.
func mustGetEventData[T any](event events.Event, expectedType string) T {
    data, ok := event.Data.(T)
    if !ok {
        panic(fmt.Sprintf("event data for %s must be %T, got %T",
            expectedType, data, event.Data))
    }
    return data
}
```

**Update formatters** to use the helper in handlers.go.

**Rationale**: Makes type mismatches obvious during development/debugging. Better panic messages than Go's default.

#### 1.2 Fix Error Handling in Display.Start()

**File**: `formatter.go`

**Current**:
```go
if err := d.formatter.Format(event); err != nil {
    return
}
```

**Refactored**:
```go
if err := d.formatter.Format(event); err != nil {
    fmt.Fprintf(os.Stderr, "[display] format error: %v\n", err)
}
```

**Rationale**: At least surfaces errors for debugging. Exits only on channel close, not on format errors.

#### 1.3 Add Error Field Documentation

**File**: `handlers.go` - Document which formatters can return errors and why.

**Rationale**: Clarifies API contract for future maintainers.

**Exit criteria**: All Phase 1 changes pass tests.

---

### Phase 2: Improved Testability

**Goal**: Add interfaces and test utilities without changing behavior.

#### 2.1 Create TextFormatter Interface

**File**: `text.go`

**Add interface**:
```go
// TextFormatter defines the interface for text formatting operations.
type TextFormatter interface {
    IndentContent(content string) string
    FormatContentWithFrame(content string, useBorder ...bool) string
    FormatContentWithFrameAndColor(content string, color string, useBorder ...bool) string
    FormatDuration(d time.Duration) string
    FormatTime() string
    ApplyReverseVideo(text string, color string) string
    TerminalWidth() int
}

// Ensure existing struct implements interface
var _ TextFormatter = (*TextFormatter)(nil)
```

**Rationale**: Enables mocking for tests, potential future implementations (e.g., fixed-width for testing).

#### 2.2 Extract FrameBuilder for Complex Formatting

**New file**: `text_frame.go`

**Extract from `FormatContentWithFrame`**:
```go
// FrameBuilder constructs framed content with borders and wrapping.
type FrameBuilder struct {
    contentWidth  int
    borderChar    string
    indent        string
    textColor     string
    useBoxDrawing bool
}

func NewFrameBuilder(opts ...FrameOption) *FrameBuilder
type FrameOption func(*FrameBuilder)

func (fb *FrameBuilder) Build(content string) string
func (fb *FrameBuilder) wrapLine(line string) string
```

**Rationale**:
- Breaks down complex function (~200 lines with duplicated wrapping logic) into smaller, testable pieces
- Makes line wrapping logic independently verifiable
- Enables different border styles via options
- **Opportunity**: Consolidate duplicated wrapping code during extraction

#### 2.3 Add MockFormatter for Testing

**New file**: `mock.go`

```go
// MockFormatter captures events for testing.
type MockFormatter struct {
    mu     sync.Mutex
    Events []events.Event
}

func (m *MockFormatter) Format(event events.Event) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.Events = append(m.Events, event)
    return nil
}

func (m *MockFormatter) Flush() error { return nil }
```

**Rationale**: Enables integration testing of the display pipeline without stdout. Thread-safe for concurrent use.

#### 2.4 Extract Test Utilities

**New file**: `test_util_test.go` (test file)

```go
// stripANSI removes ANSI color codes from a string for testing
func stripANSI(s string) string { ... }
```

**Rationale**: Removes duplication across test files. Single source of truth for test helpers.

#### 2.5 Add FrameBuilder Tests

**New file**: `text_frame_test.go`

```go
func TestFrameBuilder_EmptyContent(t *testing.T)
func TestFrameBuilder_NoBorder(t *testing.T)
func TestFrameBuilder_WithBorder(t *testing.T)
func TestFrameBuilder_LineWrapping(t *testing.T)
func TestFrameBuilder_WrappingAtWhitespace(t *testing.T)
func TestFrameBuilder_ColorApplied(t *testing.T)
```

**Rationale**: Complex logic deserves dedicated tests. Line wrapping has edge cases worth covering.

**Exit criteria**: Tests pass, coverage maintained or improved.

---

### Phase 3: Modularization (File Organization)

**Goal**: Split handlers.go into logical category files while preserving behavior.

#### 3.1 Move Formatter Registration to Formatter Registry

**File**: `formatter_registry.go` (new)

Move from handlers.go:
- `EventFormatter` type
- `FormatContext` struct
- `eventFormatters` map
- `GetColorForEventType` function

**Keep in handlers.go**:
- All individual `format*` functions
- `formatPrettyJSON` helper

**Rationale**: Registry (configuration) is separate from implementations (formatters).

#### 3.2 Split Formatters by Category

**New files** to create:

| New File | Move From handlers.go | Expected Size | Rationale |
|----------|----------------------|---------------|-----------|
| `events.go` | 20 prompt/loop/evolve formatters | ~200 lines | Core workflow events, grouped by section markers |
| `events_git.go` | 4 git formatters | ~50 lines | Git operations are distinct from workflow events |

**Alternative** (if formatters grow): Split into more files as needed.

**DELETE**: `handlers.go` (after moving all functions)

**Rationale**:
- Each file has focused responsibility
- Small files are easier to review and maintain
- Avoid over-engineering: 5-15 line functions don't each need their own file
- Follows principle of compositional design

#### 3.3 Update Imports Across Package

**Action**: Ensure all new files have correct imports. The `formatPrettyJSON` helper should be in a shared location (e.g., `formatter_util.go`) or duplicated if it stays local to each file.

**Rationale**: Splitting files requires updating import paths.

**Exit criteria**: All tests pass after file reorganization.

---

### Phase 4: Advanced Improvements (Optional)

These changes add value but are not required for stability or testability.

#### 4.1 Extract ColorScheme Type

**New file**: `color_scheme.go`

```go
// ColorScheme defines ANSI color mappings for event types.
type ColorScheme struct {
    PromptStarted  string
    Started        string
    Completed      string
    Failed         string
    Info           string
    Tool           string
}

func DefaultColorScheme() ColorScheme { ... }
```

**Rationale**:
- Enables theme customization
- Makes color logic independently testable
- Decouples color definition from formatting logic

#### 4.2 Extract ContentFilter to Interface

**File**: `content.go`

```go
type ContentFilter interface {
    ApplyToolInputFilters(toolName string, input map[string]interface{}) map[string]interface{}
    LimitCodeBlock(content string) string
    Verbose() bool
}
```

**Rationale**:
- Enables custom filters for testing
- Makes verbose mode a method rather than a field

#### 4.3 ConsoleFormatter Configuration Struct

**File**: `console.go`

```go
type ConsoleFormatterConfig struct {
    Writer        io.Writer
    Verbose       bool
    TextFormatter TextFormatter
    ContentFilter ContentFilter
    ColorScheme   ColorScheme
}

func NewConsoleFormatter(cfg ConsoleFormatterConfig) *JSONFormatter
```

**Rationale**: Cleaner constructor, easier to add options, enables dependency injection.

---

## Implementation Order (Refined)

### Priority 1: Safety & Stability (30-45 min)
1. **Phase 1.1**: Add `mustGetEventData` helper for type-safe extraction
2. **Phase 1.2**: Fix error handling in `Display.Start()` - don't exit loop on error
3. **Phase 1.3**: Document error-returning formatters

### Priority 2: Testability (1-2 hours)
1. **Phase 2.1**: Add `TextFormatter` interface
2. **Phase 2.2**: Extract `FrameBuilder` from `FormatContentWithFrame`
3. **Phase 2.3**: Add `MockFormatter` for testing
4. **Phase 2.4**: Extract test utilities
5. **Phase 2.5**: Add FrameBuilder tests (`text_frame_test.go`)

### Priority 3: Modularization (1-2 hours)
1. **Phase 3.1**: Create `formatter_registry.go` (move registry, color mapping)
2. **Phase 3.2**: Create `events.go` (move 20 formatters with section headers)
3. **Phase 3.3**: Create `events_git.go` (move 4 git formatters)
4. **Phase 3.4**: Delete `handlers.go`
5. **Phase 3.5**: **Validation step** - run `make quality` to confirm all tests pass

### Priority 4: Optional Improvements (1-2 hours)
1. Extract `ColorScheme` type
2. Extract `ContentFilter` interface
3. Add configuration struct for ConsoleFormatter

---

## Git Commit Strategy

| Commit | Changes |
|--------|---------|
| 1 | Phase 0: Baseline test run |
| 2 | Phase 1.1: Add `mustGetEventData` helper |
| 3 | Phase 1.2: Fix error handling in `Display.Start()` |
| 4 | Phase 1.3: Document formatter errors |
| 5 | Phase 2.1: Add `TextFormatter` interface |
| 6 | Phase 2.2: Extract `FrameBuilder` |
| 7 | Phase 2.3: Add `MockFormatter` |
| 8 | Phase 2.4: Extract test utilities + FrameBuilder tests |
| 9 | Phase 3.1: Create `formatter_registry.go` (registry + color mapping) |
| 10 | Phase 3.2: Create `events.go` (20 formatters with section headers) |
| 11 | Phase 3.3: Create `events_git.go` (4 git formatters) |
| 12 | Phase 3.4: Delete `handlers.go` |
| 13 | Phase 4+: Optional improvements |

**Rationale**: Small, atomic commits make rollback easier and review faster. Consolidated formatters into 2 files instead of 4 to reduce overhead.

---

## Backward Compatibility

**All exported symbols must be preserved**:

| Original | Preserved As |
|----------|--------------|
| `NewConsoleFormatter` | `NewConsoleFormatter` (in console.go) |
| `NewDisplay` | `NewDisplay` (in formatter.go) |
| `NewTextFormatter` | `NewTextFormatter` |
| `NewContentFilter` | `NewContentFilter` |
| `Display.Format` | Via interface |
| `Display.Flush` | Via interface |
| `Display.Start` | `Display.Start` |
| `Display.Wait` | `Display.Wait` |
| `TextFormatter.*` | Methods preserved |
| `ContentFilter.*` | Methods preserved |
| All `Format*` functions | Preserved (in events.go/events_git.go) |
| `GetColorForEventType` | `GetColorForEventType` (in formatter_registry.go) |
| All color constants | Preserved (in color.go) |
| All constants (`DefaultTerminalWidth`, `ContentIndent`, `Max*`) | Preserved |
| `Formatter` interface | `Formatter` interface |
| `EventFormatter` type | `EventFormatter` type |
| `FormatContext` struct | `FormatContext` struct |
| `eventFormatters` map | `eventFormatters` map |

**Files deleted**: `handlers.go` (split into category files)

**Files added**:
- `events.go` (mustGetEventData helper + 20 formatters)
- `text_frame.go` (FrameBuilder)
- `mock.go` (MockFormatter)
- `test_util_test.go` (test utilities)
- `formatter_registry.go` (registry + color mapping)
- `events_git.go` (4 git formatters)

---

## Files After Refactoring

```
pkg/display/
├── console.go              # JSONFormatter, constants
├── formatter.go            # Formatter interface, Display
├── formatter_registry.go   # EventFormatter type, FormatContext, eventFormatters map, GetColorForEventType
├── text.go                 # TextFormatter with interface
├── text_frame.go           # FrameBuilder for complex formatting
├── color.go                # ANSI constants
├── events.go               # Prompt/loop/evolve event formatters (20 functions)
├── events_git.go           # Git event formatters (4 functions)
├── content.go              # ContentFilter
├── mock.go                 # MockFormatter for testing
├── test_util_test.go       # Test utilities
├── console_test.go         # Tests for Display
├── content_test.go         # Tests for ContentFilter
└── text_test.go            # Tests for TextFormatter
```

**Deleted**: `handlers.go`

**Note**: FrameBuilder tests should be added in `text_frame_test.go` when extracting Phase 2.2.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking existing tests | Low | High | Run tests after every commit; no test file modifications planned |
| Runtime panics from type assertions | Medium | High | Replace with `mustGetEventData` helper in Phase 1 |
| Loss of functionality | Low | High | Comprehensive backward compatibility table above |
| Increased complexity from new types | Medium | Medium | Phase carefully; only extract when complexity warrants |
| Merge conflicts during file split | Low | Medium | Split in atomic commits; run tests immediately |
| Error handling regression | Medium | High | Verify Display.Start() loop doesn't exit on format errors |
| Over-modularization (YAGNI) | Low | Medium | Consolidate formatters into 2 files instead of 4+ to avoid file overhead |

---

## Rollback Plan

If any phase introduces issues:

1. **Before Phase 3 (Modularization)**:
   - `git checkout HEAD -- pkg/display/handlers.go` restores original
   - Revert specific commits from Phase 1 or 2

2. **After Phase 3**:
   - `git checkout HEAD -- pkg/display/handlers.go` restores the pre-split file
   - Delete the new category files
   - This is safe because category files contain only moved code

3. **FrameBuilder extraction issues**:
   - If consolidation reveals bugs, stop and document the issue
   - FrameBuilder should preserve existing behavior - refactor, don't rewrite

---

## Estimated Effort

| Phase | Estimated Time | Risk Level |
|-------|----------------|------------|
| Phase 1 (Stabilization) | 30-45 min | Low |
| Phase 2 (Testability) | 1-2 hours | Medium |
| Phase 3 (Modularization) | 1-2 hours | Medium |
| Phase 4 (Optional) | 1-2 hours | Low |

**Note**: Reduced estimates from simplification - fewer files to create/delete.

---

## Success Criteria

1. [ ] All existing tests pass without modification
2. [ ] No behavioral changes to output formatting
3. [ ] File organization matches target structure (9 files, none >200 lines)
4. [ ] New interfaces enable easier testing (`TextFormatter`, `MockFormatter`)
5. [ ] Code complexity reduced (smaller files, extracted `FrameBuilder`)
6. [ ] FrameBuilder has dedicated tests (`text_frame_test.go`)
7. [ ] MockFormatter enables integration testing without stdout

---

## Testing Strategy

### New Tests to Write

| Test File | What to Test | Why |
|-----------|--------------|-----|
| `text_frame_test.go` | FrameBuilder: empty content, borders, line wrapping, colors | Complex logic needs coverage |
| `mock_test.go` (in mock.go) | MockFormatter: event capture, thread safety | Validates testing infrastructure |

### Validation After Each Phase

1. **After Phase 1**: `go test ./pkg/display/...` - verify no regressions
2. **After Phase 2**: `go test ./pkg/display/... -cover` - confirm coverage maintained/improved
3. **After Phase 3**: `make quality` - full validation including linter

---

## Notes

### What This Refactor Does NOT Change

- **Output formatting**: Colors, frames, wrapping, timing formats all preserved
- **Event types**: All 24 event formatters continue to work identically
- **Public API**: All exported functions and types remain available
- **Dependencies**: No new external dependencies added

### What This Refactor Improves

- **Debuggability**: Clearer panic messages (`mustGetEventData`), error logging in Display.Start()
- **Testability**: Interfaces enable mocking, FrameBuilder extraction enables unit testing of complex logic
- **Maintainability**: Registry separated from implementations, formatters grouped by section headers
- **Extensibility**: Clear location for new formatters (events.go or events_git.go)
- **Robustness**: Error handling no longer silently drops events on format failures

### Go Version Note

The `max` builtin mentioned in the original plan is not an issue. Go 1.21+ includes `max` as a builtin that doesn't require import. Since go.mod requires Go 1.25.3, this is not a concern.
