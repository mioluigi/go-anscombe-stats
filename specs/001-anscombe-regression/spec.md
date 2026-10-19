# Feature Spec 001 — Validate the Go `stats` package against Python and R on the Anscombe Quartet

| Field | Value |
|---|---|
| Status | **APPROVED v2 — 2026-10-18** (Gemini review: `docs/reviews/001-spec-review.md`; decisions: section 11) |
| Owner | Luis Varela |
| Course deliverable | Week 4 Assignment: Go for Statistics (100 pts) |
| Governing rules | `docs/constitution.md` |

> **Spec rule:** this document says **what** and **why**. It deliberately avoids **how**
> (file layout, function names, libraries beyond those the assignment names). That belongs in `plan.md`.

---

## 1. Problem (why)

Management wants Go as the company's primary language. Data scientists will accept that only if the
proposed Go statistics package — `github.com/montanaflynn/stats` — gives **correct** answers and
**reasonable performance** compared with the Python and R tools they already trust.

The Anscombe Quartet is the agreed first test: four tiny data sets that look very different but
produce almost identical regression coefficients. If Go, Python, and R all reproduce those
coefficients, that is good initial evidence of correctness.

## 1a. Provided course materials (inputs to this spec)

| File | Role in this project |
|---|---|
| Week 4 assignment PDF + rubric | Requirements and grading criteria (source of the constitution). |
| `miller_mtpa_chapter_1_and_bibliography_sample.pdf` | **Authoritative data source** (Table 1.1, p. 9) and **published expected result** (p. 9): ŷ = 3 + 0.5x for all four sets, R² ≈ 0.67, mean of x = 9, mean of y = 7.5. |
| `miller-mtpa-chapter-1-program.py` (Exhibit 1.1) | **Python baseline.** Fits each set with `statsmodels` `sm.OLS` + `sm.add_constant`, prints `.summary()`. Uses pandas, NumPy, matplotlib. |
| `miller-mtpa-chapter-1-program.R` (Exhibit 1.2) | **R baseline.** Fits each set with base `lm(y ~ x)`, prints `summary()`. Writes a PDF plot. |

**Observed when running Miller's Python program as provided** (matplotlib 3.11, statsmodels 0.15):
the four regressions print correctly and match section 5, but the program then **exits with an error**
at `plt.savefig(...)`: `papertype` is no longer an accepted argument. This is one more reason the
adapted benchmark version drops plotting. It also supports a point for the recommendation: Python
analysis code breaks as library APIs change, so pinned dependency versions matter in either language.

Handling rules:
- The two provided programs are kept **unmodified** in the repo for traceability (they are the instructor's reference).
- Benchmark versions are **adapted copies**: plotting removed (the assignment says to ignore it), and console printing kept **outside** the timed region so Go, Python, and R time the same work — fitting four regressions.
- All three languages read the data from **one shared data file** transcribed from Table 1.1, rather than each hard-coding its own copy.

## 2. Users and what they need

| User | Needs |
|---|---|
| **Data scientists** (primary) | Proof that Go's regression coefficients match Python and R; an honest list of gaps vs. Python/R. |
| **Management** | A clear recommendation: adopt Go for data science, adopt with conditions, or don't. |
| **Grader / any engineer** | Clone the repo, follow the README, build, run, test, and benchmark without help. |

## 3. User stories

- **US-1 Correctness.** As a data scientist, I want the Go program to compute the intercept and slope of a simple linear regression (y on x) for each of the four Anscombe sets, so I can compare them with Python and R.
- **US-2 Automated proof.** As a data scientist, I want `go test` to fail if Go's coefficients differ from the Python/R reference values beyond a stated tolerance, so correctness is checked automatically, not by eye.
- **US-3 Performance.** As a data scientist, I want execution time and memory for Go, Python, and R on the same task, measured in a comparable way, so I can judge whether Go is fast enough.
- **US-4 Usable program.** As any user, I want to run one executable that prints a clean results table, and get a helpful message (not a crash) if I give it bad input.
- **US-5 Decision.** As a manager, I want a short recommendation in the README, including the concerns data scientists would raise.

## 4. Functional requirements

| ID | Requirement |
|---|---|
| FR-1 | The four Anscombe data sets, transcribed from Miller (2015) Table 1.1, are stored once as a data file and used by **all three** languages. |
| FR-2 | For each set, the Go program reports **intercept** and **slope** of the least-squares line of y on x. It also reports **R²** (correlation squared) and the **means of x and y** as supporting checks against Miller's published summary. |
| FR-3 | Go must use `github.com/montanaflynn/stats` for the regression (that package is what is being evaluated). |
| FR-4 | A Python program, adapted from Miller's Exhibit 1.1 (`statsmodels` OLS), fits the same models and reports the same quantities. |
| FR-5 | An R program, adapted from Miller's Exhibit 1.2 (base `lm()`), fits the same models and reports the same quantities. |
| FR-5a | Miller's original Python and R programs are kept unmodified in the repo; the README explains how the adapted versions differ and why. |
| FR-6 | Go unit tests assert Go coefficients equal the reference values within tolerance **1e-3** (reference values are reported to 4 decimals). |
| FR-7 | Go benchmarks report time per operation and memory/allocations per operation. |
| FR-8 | Python and R time the **same unit of work as Go** (fit all four regressions once), repeated **N = 1,000** times after a warm-up, and report the **median time per run in microseconds (µs)**. They also report memory allocated by that work: Python via `tracemalloc` (peak KB), R via `bench::mark` (`mem_alloc`, KB). Performance is **measured and reported, not pass/fail**. |
| FR-9 | The program runs with **no arguments** on the bundled data. It optionally accepts a path to a user-supplied CSV with `x,y` columns. |
| FR-10 | Invalid input — missing file, wrong columns, non-numeric values, unequal lengths, fewer than 2 points, all-identical x values — produces a clear message and a non-zero exit code. |
| FR-12 | **Program logging.** The Go program logs its main steps (data loaded, each set fitted, total elapsed time) with timestamps using the standard library `log/slog`, to standard error so they never mix with the results table. A `-verbose` flag shows debug-level detail. |
| FR-13 | **Whole-program metrics.** Each language's full program is run by a measurement tool (`cmd/measure`) that reads the operating system's per-process resource report (rusage — the same source `/usr/bin/time` uses), reporting **wall-clock time (ms)** and **maximum resident set size (MB)**, median of repeated runs. This compares complete programs, including interpreter/runtime start-up, not just the fitting code. *(Changed in T14: `/usr/bin/time` reports time only to 10 ms, too coarse for the Go program, which finishes in under 1 ms.)* |
| FR-11 | The README contains: purpose, roles of programs and data, build/run/test/benchmark instructions, results tables (accuracy and performance), and the recommendation to management. |

## 5. Reference values (the "known answers")

Three independent sources agree:

1. **Miller (2015), p. 9:** ŷ = 3 + 0.5x for every set; R² ≈ 0.67; mean x = 9; mean y = 7.5 (rounded, published).
2. **`statsmodels` OLS**, the method in Miller's Exhibit 1.1, run on Table 1.1 data while drafting this spec.
3. **NumPy least squares** (`numpy.polyfit`) as a cross-check — identical to 4 decimals.

| Set | Intercept | Slope | R² |
|---|---|---|---|
| I   | 3.0001 | 0.5001 | 0.6665 |
| II  | 3.0009 | 0.5000 | 0.6662 |
| III | 3.0025 | 0.4997 | 0.6663 |
| IV  | 3.0017 | 0.4999 | 0.6667 |

**R expectation:** R's `lm()` must reproduce this same table to 4 decimals (AC-3). The values are already defined; running R on Luis's Mac only *confirms* them, and the output is recorded in the README.

## 6. Key finding that shapes the plan ⚠️

`stats.LinearRegression(s Series) (regressions Series, err error)` **does not return slope and
intercept.** It returns the **fitted points** — one `Coordinate{X, Yhat}` per input point.

Consequences:
- Coefficients must be **derived** from two fitted points with distinct x values:
  `slope = (ŷ₂ − ŷ₁) / (x₂ − x₁)`, `intercept = ŷ₁ − slope · x₁`.
- Set IV has 10 of 11 points at x = 8, so the derivation must pick points with **different x**. This is an edge case to test explicitly.
- **Verified by spike (2026-10-18, stats v0.13.0):** deriving coefficients from the fitted points at the smallest and largest x reproduces section 5 exactly for all four sets, including Set IV. When every x is identical, the library itself returns `ErrBounds` ("Input is outside of range."); our code must translate that into a clear message for the user.
- This is itself a finding for the management recommendation: the Go package is less convenient than `statsmodels` or R's `lm()`, which return coefficients, standard errors, p-values, and diagnostics directly.

## 7. Non-functional requirements

| ID | Requirement | Rubric |
|---|---|---|
| NFR-1 | Code follows `docs/constitution.md` Article 2 (idiomatic Go, McConnell). | Coding |
| NFR-2 | Test coverage of the reusable statistics code ≥ 90% (`go test -cover`). | Testing |
| NFR-3 | CPU and memory profiles generated once; how to reproduce is documented. | Testing |
| NFR-4 | Statistics logic is a reusable package, separate from the program's input/output. | Design |
| NFR-5 | A grader can go from `git clone` to results in ≤ 5 documented commands. | Documentation / Application |
| NFR-6 | Documented build commands produce: a macOS command-line binary, a **Windows `.exe`** (Go cross-compilation), and — per Q5 — a macOS **`.app`** bundle. The program runs to completion on macOS. | Application |

## 8. Acceptance criteria (definition of done)

- [x] AC-1 `go test ./...` passes; the correctness tests compare all 4 sets × (intercept, slope, R², mean x, mean y) against section 5 and Miller's published summary.
- [x] AC-1a **Synthetic known-answer tests** pass: a perfect line (e.g., y = 2x + 1 → slope 2, intercept 1, R² 1), a negative slope, and a horizontal line (slope 0).
- [x] AC-1b **Distinct-x derivation tests** pass: coefficients are derived correctly when most x values repeat (Set IV pattern: one x differs from the rest), and an error — not a divide-by-zero — is returned when every x is identical.
- [x] AC-2 `go test -bench=. -benchmem ./...` prints ns/op, B/op, allocs/op.
- [x] AC-3 Python and R programs run and print coefficients matching section 5 to 4 decimals, plus median µs per run (N = 1,000) and allocated KB, as defined in FR-8.
- [x] AC-4 `go build` produces an executable that prints a 4-row results table with no arguments; the documented `.exe` (and `.app`, if chosen in Q5) build commands succeed.
- [x] AC-5 Each invalid-input case in FR-10 has a test and a friendly message.
- [x] AC-6 README has every item in FR-11, including a performance table across Go / Python / R.
- [ ] AC-7 The repo URL ending in `.git` is ready to paste into Canvas.
- [x] AC-8 Running the program writes timestamped log lines to standard error (FR-12); `-verbose` adds debug detail.
- [x] AC-9 README includes a whole-program table (wall time, max RSS) for Go, Python, and R, produced by `scripts/measure.sh` (FR-13).

## 9. Rubric traceability

| Rubric criterion (20 pts each) | Satisfied by |
|---|---|
| Coding rules, organization, aesthetics | NFR-1, Constitution Art. 2 |
| Testing and software metrics | FR-6, FR-7, FR-8, FR-12, FR-13, NFR-2, NFR-3, AC-1, AC-1a, AC-1b, AC-2, AC-5, AC-8, AC-9 |
| Design and development | FR-1, FR-3, NFR-4 |
| Documentation | FR-11, NFR-5, AC-6, AC-7 |
| Application | FR-9, FR-10, NFR-6, AC-4 |

## 10. Out of scope

- Plotting (the assignment says to ignore matplotlib).
- Multiple regression, standard errors, p-values (noted as a *gap* in the recommendation, not built).
- A GUI. The application is a command-line program.

## 11. Decisions — **approved by Luis, 2026-10-18**

| # | Question | Decision |
|---|---|---|
| Q1 | Python libraries | `statsmodels` OLS (Miller's Exhibit 1.1) as primary; `numpy.polyfit` timed as a second Python row. |
| Q2 | R | Base `lm()` (Miller's Exhibit 1.2); `bench` package for timing and memory. |
| Q3 | Test assertions | Go standard library `testing` only. |
| Q4 | Monte Carlo benchmark | Yes, as a stretch goal: fits on synthetic data of increasing size (100 → 1,000,000 points). |
| Q5 | Executables | Build both a Windows `.exe` and a macOS `.app` (wrapper that opens Terminal so output is visible). |
| Q6 | Repository | `github.com/mioluigi/go-anscombe-stats` |
