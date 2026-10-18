# Plan 001 — How we will build Spec 001

| Field | Value |
|---|---|
| Status | **APPROVED v2 — 2026-10-18** (Gemini review: `docs/reviews/002-plan-review.md`) |
| Implements | `spec.md` (APPROVED v2) |
| Governing rules | `docs/constitution.md` |

> **Plan rule:** the spec says *what* and *why*; this plan says *how*. Every design choice below
> states its reason, so each one can be defended to a grader or changed deliberately.

---

## 1. Technical context

| Item | Choice | Why |
|---|---|---|
| Language | Go (current stable toolchain on Luis's Mac) | Assignment requirement. |
| Go statistics package | `github.com/montanaflynn/stats` **v0.13.0** | The package under evaluation (FR-3). Version pinned in `go.mod` so results are reproducible. |
| Go module path | `github.com/mioluigi/go-anscombe-stats` | Matches the GitHub repo, so `go install`/`go get` work. |
| Go test framework | Standard library `testing` only | Decision Q3. |
| Logging | Standard library `log/slog` | FR-12; no extra dependency. |
| Python | `statsmodels` + `numpy` + `pandas`, versions pinned in `python/requirements.txt`, installed in a virtual environment | Decision Q1; pinning avoids the matplotlib-style breakage found in Miller's program. |
| R | Base `lm()`; `bench` package for timing | Decision Q2. |
| Whole-program metrics | macOS `/usr/bin/time -l` | FR-13; built into macOS, works identically for all three languages. |

**Verified before planning (spike, see spec section 6):** coefficients derived from the fitted points
at the smallest and largest x match the reference table for all four sets.

## 2. Repository layout

```
go-anscombe-stats/
├── go.mod, go.sum                  Go module definition and pinned dependencies
├── cmd/anscombe/main.go            The command-line program: flags, logging, output table
├── regression/                     REUSABLE package — fitting only, no I/O
│   ├── regression.go
│   ├── regression_test.go          Correctness: Anscombe + synthetic + edge cases
│   └── regression_bench_test.go    Benchmarks: Anscombe unit of work + Monte Carlo sizes
├── dataset/                        REUSABLE package — reading and validating CSV data
│   ├── dataset.go
│   └── dataset_test.go
├── data/
│   ├── anscombe.csv                The one shared data file (Go, Python, R all read it)
│   └── embed.go                    Embeds anscombe.csv into the Go binary
├── python/
│   ├── anscombe_bench.py           Adapted from Miller Exhibit 1.1
│   └── requirements.txt            Pinned versions
├── r/anscombe_bench.R              Adapted from Miller Exhibit 1.2
├── reference/                      Miller's original programs, unmodified (FR-5a)
├── scripts/
│   ├── measure.sh                  Whole-program time + memory for Go, Python, R (FR-13)
│   └── build-macos-app.sh          Builds Anscombe.app (Q5)
├── results/                        Captured outputs used in the README
├── docs/                           constitution, rubric summary, reviews, decisions
├── specs/001-anscombe-regression/  spec.md, plan.md, tasks.md
├── CLAUDE.md                       Project context for Claude Code
└── README.md                       Complete documentation and recommendation (FR-11)
```

**Why two reusable packages instead of one file?** The rubric's Design criterion asks for
"general-purpose, reusable modules". `regression` knows nothing about files; `dataset` knows
nothing about statistics. Either can be imported by another project unchanged, and each can be
tested in isolation (Constitution 4.1, 4.2).

## 3. Data design

`data/anscombe.csv` uses **long ("tidy") format** — one row per observation:

```
set,x,y
I,10,8.04
I,8,6.95
...
IV,19,12.50
```

**Why long format, not Miller's wide format (x1, y1, …, x4, y4)?**
- Adding a fifth data set means adding rows, not changing the code.
- The same reader handles a user's own CSV (FR-9): a file with only `x,y` columns is treated as one set named "user".
- Python (`pandas.groupby("set")`) and R (`split(df, df$set)`) handle it in one line each.

**Embedding.** `data/embed.go` uses Go's `//go:embed anscombe.csv`, so the compiled program carries
its data. **Why:** the `.exe` and `.app` must "run to completion" with no arguments (FR-9, AC-4),
even when copied away from the repo.

**Why a tiny `data` package instead of embedding in `main.go`?** Go's `//go:embed` can only reach
files in the same directory as the Go file or below it — never `../`. Embedding from
`cmd/anscombe/main.go` would force the CSV to live inside `cmd/anscombe/`, which would either make
Python and R read data from a Go command folder or require a second copy of the data, breaking
FR-1's single source of truth. A two-line package beside the CSV keeps one data file for all three
languages. Embedding is a compile-time step, not runtime I/O, so Constitution 4.2 is respected.

## 4. Package `regression` — the core

```go
// Fit is the result of a simple linear regression of y on x.
type Fit struct {
    Intercept, Slope, RSquared, MeanX, MeanY float64
}

// FitLine estimates y = Intercept + Slope·x by least squares using montanaflynn/stats.
func FitLine(x, y []float64) (Fit, error)
```

**Algorithm (with reasons):**

1. **Validate first** — return typed errors for: unequal lengths, fewer than 2 points, NaN or ±Inf values.
   *Why:* clear messages for users (FR-10) and no reliance on library error text.
2. **Call `stats.LinearRegression`** on the points. It returns fitted points, not coefficients.
3. **Wrap `stats.ErrBounds`** (all x identical) with our own sentinel, keeping both in the error chain:
   `fmt.Errorf("%w: all x values are identical, so no line can be fitted: %w", ErrConstantX, err)`.
   *Why:* the user gets a plain-English message, and callers can test `errors.Is(err, ErrConstantX)`
   (or the library's `stats.ErrBounds`) programmatically. Multiple `%w` verbs are supported since Go 1.20.
4. **Derive coefficients from the fitted points at the minimum and maximum x.**
   `Slope = (ŷmax − ŷmin) / (xmax − xmin)`, `Intercept = ŷmin − Slope · xmin`.
   *Why min and max?* They are guaranteed distinct whenever any two x values differ (handles Set IV),
   and the widest spacing gives the most numerically stable slope.
   *Precision (measured):* back-deriving from fitted values differs from an independent closed-form OLS
   by at most **4.4 × 10⁻¹⁶** on the Anscombe data — rounding at the limit of float64. The source code
   carries a comment explaining *why* coefficients are back-derived (the library discards them), and a
   test guards precision against an independent closed-form calculation at tolerance 1e-9.
5. **R² = `stats.Correlation(x, y)`²; means = `stats.Mean`.** All statistics come from the package under test.

**Exported errors:** `ErrLengthMismatch`, `ErrTooFewPoints`, `ErrNotFinite`, `ErrConstantX`, testable with `errors.Is`.

## 5. Package `dataset`

```go
// Set is one named group of paired observations.
type Set struct { Name string; X, Y []float64 }

// Read parses CSV with header "set,x,y" or "x,y" and returns sets in first-seen order.
func Read(r io.Reader) ([]Set, error)
```

Errors include the **line number** and the problem (e.g., `line 7: y value "abc" is not a number`).
**Why `io.Reader` instead of a file name?** Tests can pass a string (`strings.NewReader`) — no temporary
files — and the same function reads the embedded data, a file, or standard input.

## 6. The program `cmd/anscombe`

```
anscombe                    # bundled Anscombe data
anscombe -data mine.csv     # user's own CSV
anscombe -verbose           # debug-level logs
anscombe -h                 # usage
```

- **Output:** results table on standard output via `text/tabwriter`: Set, n, Intercept, Slope, R², Mean x, Mean y.
- **Logs:** `slog` text handler on **standard error** (FR-12), so `anscombe > results.txt` keeps the table clean.
- **Exit codes:** `0` success · `1` data problem (message explains it) · `2` bad command-line usage.
- **`main` stays short:** parse flags → open data → `dataset.Read` → `regression.FitLine` per set → print.
  Each step is a small named function (Constitution 2.4).

## 7. Testing strategy (traceability to acceptance criteria)

| Test | Location | Covers |
|---|---|---|
| Anscombe table test: 4 sets × intercept, slope, R², mean x, mean y vs. spec section 5 (tolerance 1e-3; means vs. Miller's 9 and 7.5) | `regression_test.go` | AC-1, FR-6 |
| Synthetic known answers: y = 2x + 1; negative slope; horizontal line (slope 0; R² returned as 0 by the package — see note) | `regression_test.go` | AC-1a |
| Set IV pattern (repeated x, one differs) and all-identical x → `errors.Is(err, ErrConstantX)` | `regression_test.go` | AC-1b |
| Precision guard: `FitLine` vs. an independent closed-form OLS written in the test, tolerance 1e-9 | `regression_test.go` | FR-6, Constitution 1.1 |
| Invalid inputs: length mismatch, < 2 points, NaN, Inf | `regression_test.go` | AC-5 |
| CSV parsing: both headers, missing file, wrong columns, non-numeric value with line number, empty file | `dataset_test.go` | AC-5, FR-10 |
| Embedded data loads and has 4 sets × 11 rows | `data` package test | FR-1 |
| Program run end-to-end: exit code 0 and a 4-row table; bad file → exit 1 with message | `cmd/anscombe` test (builds and runs the binary) | AC-4, AC-5 |
| Coverage ≥ 90% on `regression` and `dataset` | `go test -cover ./...` | NFR-2 |

All tests are **table-driven** and written **before** the code they test (Constitution 3.1, 3.3).
Floating-point checks use a named tolerance constant (Constitution 3.4).

> **Note — horizontal line (verified by spike):** when every y is identical, correlation is mathematically
> undefined (0/0). `stats.Correlation` **silently returns 0 with no error**, whereas R reports `NA` and
> NumPy returns `nan` with a warning. `FitLine` passes the package's value through unchanged, because we
> are evaluating the package, not hiding its behavior. A test pins this behavior down, and the README lists
> it as a concern for data scientists: undefined results can look like real numbers.

## 8. Performance measurement design

**Unit of work (same in all three languages):** fit all four Anscombe regressions once.

| Language | Component timing | Memory | Repetitions |
|---|---|---|---|
| Go | `go test -bench=. -benchmem` → ns/op | B/op, allocs/op | Go chooses N automatically; `-count=10` for stability |
| Python (statsmodels) | `time.perf_counter_ns` | `tracemalloc` peak KB | warm-up, then N = 1,000; report median µs |
| Python (NumPy) | same | same | same |
| R (`lm`) | `bench::mark` | `mem_alloc` KB | 1,000 iterations; median µs |

**One unit for every table: microseconds (µs).** Go's benchmark reports nanoseconds per operation;
the README converts it (ns/op ÷ 1,000 = µs) so all three languages appear side by side in the same
unit. The conversion is done by a small script reading the saved benchmark output, not by hand
(Constitution 1.3), and the README states the conversion under the table. Memory is likewise shown in KB.

**Whole program (FR-13):** `scripts/measure.sh` runs each full program under `/usr/bin/time -l` and
records wall time and maximum resident set size into `results/whole-program.txt`.

**Monte Carlo stretch (Q4):** `BenchmarkFitLine/n=100 … n=1000000` on synthetic data
y = 3 + 0.5x + noise, generated with a **fixed random seed** so runs are repeatable. Shows how Go
scales with data size, which the 11-point Anscombe sets cannot.

**Profiles (NFR-3):** `go test -bench=FitLine -cpuprofile cpu.out -memprofile mem.out ./regression`,
viewed with `go tool pprof`. Commands documented in the README.

**Fairness rules:** printing stays outside timed code; same machine; laptop plugged in; results
reported with machine details (chip, OS, language versions).

## 9. Executables (Q5)

| Target | Command / script | Notes |
|---|---|---|
| macOS binary | `go build -o dist/anscombe ./cmd/anscombe` | Runs in Terminal. |
| Windows `.exe` | `GOOS=windows GOARCH=amd64 go build -o dist/anscombe.exe ./cmd/anscombe` | Go cross-compiles; no Windows machine needed. |
| macOS `.app` | `scripts/build-macos-app.sh` | Creates `dist/Anscombe.app` with an `Info.plist` and a launcher that opens Terminal and runs the program, so the output stays visible after a double-click. |

The README states plainly that the `.app` is a **thin wrapper around a command-line program**, built to
meet the rubric's packaging requirement; the program itself is a terminal application.

## 10. Documentation

- **`README.md`** — sections mapped to FR-11: purpose; repository map (roles of programs and data); prerequisites; build; run; test; benchmark; results (accuracy table, component performance, whole-program metrics, Monte Carlo); concerns data scientists may have; recommendation to management; references.
- **`CLAUDE.md`** — short project context for Claude Code: points to the constitution, spec, and plan; key commands; "one task per commit".
- **`docs/decisions.md`** — running log of design decisions that change after this plan.

## 11. Prerequisites on Luis's Mac (checked in the first task)

| Tool | Check command | Install if missing |
|---|---|---|
| Go | `go version` | `brew install go` |
| Python 3 | `python3 --version` | `brew install python` |
| R | `R --version` | `brew install r` |
| R `bench` package | `Rscript -e 'library(bench)'` | `Rscript -e 'install.packages("bench", repos="https://cloud.r-project.org")'` |

## 12. Risks and mitigations

| Risk | Mitigation |
|---|---|
| R or Python not installed / version drift | Prerequisite task first; pinned Python versions; versions printed with results. |
| Timing noise on a laptop | Repetitions + median; Go `-count=10`; record machine details. |
| Unsigned `.app` blocked by macOS Gatekeeper | Built locally from source, so not quarantined; README notes right-click → Open if needed. |
| Package hides undefined results (flat line → correlation 0, no error) | Pinned by a test; documented as a data-science concern in the README. |

## 13. Constitution check

| Article | How this plan complies |
|---|---|
| 1 Correctness first | Three-source reference values; automated tolerance tests. |
| 2 Idiomatic Go | Small packages, short functions, typed errors, `gofmt`/`go vet` in the done-checklist. |
| 3 Test-first, measured | Section 7 tests precede code; section 8 metrics; profiles. |
| 4 Reusable design | `regression` and `dataset` have no I/O coupling; data shared by all languages. |
| 5 Docs and Git | README plan in section 10; one task per commit. |
| 6 Finished application | Embedded data, validated input, exit codes, three executables. |
| 7 AI tool roles | Claude Code implements; Gemini (web) reviews this plan; Luis approves. |
