# Go for Statistics: validating `montanaflynn/stats` on the Anscombe Quartet

Can the company's data scientists trust Go for statistical work? This repository tests the pure-Go
statistics package [`github.com/montanaflynn/stats`](https://pkg.go.dev/github.com/montanaflynn/stats)
against Python and R on the Anscombe Quartet (Anscombe 1973; Miller 2015), compares speed and memory,
and makes a recommendation to management.

**Short answer:** the Go package gives the **same regression coefficients** as Python (statsmodels and
NumPy) and R (`lm`) to four decimal places on all four data sets, and it is **hundreds of times faster**.
But it offers far less statistical functionality than Python or R. Go is ready for *production
statistics in services*, not yet for *exploratory data science*. See [Recommendation](#recommendation-to-management).

**Contents:** [Recommendation](#recommendation-to-management) ·
[Results](#results) · [Concerns for data scientists](#concerns-data-scientists-may-have) ·
[Quick start](#quick-start) · [Using the program](#using-the-program) · [Testing](#testing) ·
[Reproducing the measurements](#reproducing-the-measurements) · [Building executables](#building-the-executables) ·
[Repository map](#repository-map) · [How this was built](#how-this-project-was-built) · [References](#references)

---

## Recommendation to management

> **Adopt Go for production statistical computation; keep Python and R for exploratory analysis and
> statistical inference — for now.**

1. **Correctness is not a barrier.** On the agreed first test, Go's linear regression matched statsmodels,
   NumPy, and R exactly to four decimals for every Anscombe set, and matched an independent textbook
   formula to floating-point precision (differences of about 10⁻¹⁶ on the Anscombe data, and within
   10⁻⁹ on 10,000 random points).
2. **Performance strongly favors Go** where code runs repeatedly — APIs, pipelines, batch scoring. Fitting
   the four regressions took under 1 µs in Go versus hundreds of µs in Python and R, and a complete Go
   run used a small fraction of the memory and start-up time of Python or R (tables below).
3. **Functionality favors Python and R.** The Go package returns no standard errors, p-values, confidence
   intervals, or diagnostics, supports only one predictor, and has no data-frame, formula, or plotting
   tools. Data scientists would have to rebuild what statsmodels and R provide in one line.
4. **Suggested path:** (a) build production services and pipelines in Go, as management intends;
   (b) let data scientists continue research in Python/R, and port finished, validated models to Go for
   deployment, using tests like the ones in this repository to prove the port matches; (c) evaluate a
   broader Go numerical library such as Gonum before any move to Go-only data science; and (d) pin
   package versions in every language.

*Evidence for each point is in the sections below; every number comes from files in `results/`.*

---

## Results

All results were measured on an **Apple M1 Max, macOS 27.0.1**, with **Go 1.27.2**, **Python 3.12.8**
(NumPy 2.5.3, pandas 3.0.5, statsmodels 0.15.0) and **R 4.4.2** (bench 1.1.4). Raw outputs are in
[`results/`](results/).

### Correctness: regression coefficients agree in all four languages/libraries

| Set | Intercept | Slope | R² | Mean x | Mean y | All 4 methods identical to 4 decimals? |
|---|---:|---:|---:|---:|---:|---|
| I | 3.0001 | 0.5001 | 0.6665 | 9.0000 | 7.5009 | ✅ yes |
| II | 3.0009 | 0.5000 | 0.6662 | 9.0000 | 7.5009 | ✅ yes |
| III | 3.0025 | 0.4997 | 0.6663 | 9.0000 | 7.5000 | ✅ yes |
| IV | 3.0017 | 0.4999 | 0.6667 | 9.0000 | 7.5009 | ✅ yes |

Methods compared: Go (montanaflynn/stats), statsmodels, numpy, lm.

Every set yields ŷ ≈ 3 + 0.5x with R² ≈ 0.67, matching Miller's published summary (2015, p. 9) — the very
point of the Anscombe Quartet: four very different data sets share the same statistics, so **summary
statistics alone are not enough; the data must be plotted.** Table built by `scripts/accuracy_table.sh`
from `results/go.txt`, `results/python.txt`, and `results/r.txt`. (Go's coefficients do not depend on the
machine; the automated tests check them on every run.)

### Performance

#### Component benchmark: fit all four Anscombe sets once

| Language | Library | Median time (µs) | Relative to Go | Memory measure | Memory (KB) |
|---|---|---:|---:|---|---:|
| Go | montanaflynn/stats | 0.9 | 1× | allocated per run (`-benchmem`) | 2.5 |
| Python | statsmodels | 252.0 | 273× | peak traced (`tracemalloc`) | 8.4 |
| Python | numpy | 123.1 | 133× | peak traced (`tracemalloc`) | 10.5 |
| R | lm | 838.9 | 908× | profiled allocations (`bench`) ¹ | 1.8 |

Go: median of 10 `go test -bench` runs; ns/op ÷ 1,000 = µs and B/op ÷ 1,024 = KB. Python and R: median of 1,000 timed runs after 10 warm-up runs.
The three memory measures count different things and are not ranked against each other; 
see the whole-program table for a like-for-like memory comparison.
¹ R's memory profiler misses many small allocations, so this understates `lm()`'s true use.

#### Monte Carlo benchmark (Go): one fit at increasing data size

| Points (n) | Median time (µs) | Allocated per fit (MB) | Allocations per fit |
|---:|---:|---:|---:|
| 100 | 1.8 | 0.01 | 8 |
| 1000 | 20.9 | 0.06 | 12 |
| 10000 | 209.2 | 0.81 | 19 |
| 100000 | 2154.5 | 10.06 | 29 |
| 1000000 | 19631.0 | 99.22 | 39 |

Synthetic data y = 3 + 0.5x + noise with a fixed random seed; median over the `-count` runs.

#### Whole program: analysis run once, including start-up

| Language | Median wall time (ms) | Peak memory, max RSS (MB) |
|---|---:|---:|
| Go | 11.80 | 5.0 |
| Python (pandas + statsmodels + NumPy) | 1873.69 | 162.2 |
| R (base `lm`) | 154.28 | 76.0 |

Median of 5 runs measured by `cmd/measure` (wall clock from Go's monotonic timer; peak memory from the operating system's rusage report).

**How to read these tables.** The *component benchmark* times only the regression fitting — the work
a long-running service repeats. The *whole-program* table is what a user waits for when running each
program once, including starting the interpreter and importing libraries; Python's time there is
dominated by importing pandas and statsmodels, not by fitting. The *Monte Carlo* table shows Go scales
linearly: about 20 ns and 100 bytes per data point from 100 to 1,000,000 points.

**Profiles** of the 1,000,000-point fit (`results/cpu-profile.txt`, `results/memory-profile.txt`): **84 %**
of the memory allocated is inside `stats.LinearRegression` itself, and most CPU time is spent summing
values inside the package (`stats.Sum`, `Float64Data.Get`) and returning memory to the operating system
(`runtime.madvise`).

---

## Concerns data scientists may have

Each point below was observed in this project, not assumed.

| Concern | What we found | Why it matters |
|---|---|---|
| **Coefficients are not returned directly** | `stats.LinearRegression` returns fitted points (x, ŷ), not the intercept and slope. We recover them from the fitted points at the smallest and largest x ([`regression/regression.go`](regression/regression.go)). | Every user must know this workaround; a naive version (using the first two points) returns NaN on Anscombe Set IV, where ten of eleven x values are equal. |
| **No statistical inference** | No standard errors, t-statistics, p-values, confidence intervals, or residual diagnostics — all of which statsmodels' `summary()` and R's `summary(lm())` print in Miller's original programs. | Data scientists cannot test hypotheses or judge fit quality with this package alone. |
| **Only simple regression** | One predictor only; no multiple regression, formulas, or categorical variables. | Most real models have several predictors. |
| **Undefined results look like real numbers** | For a perfectly flat line (all y equal), correlation is mathematically undefined. The package silently returns **0** with no error, where R reports `NA` and NumPy returns `nan` with a warning (pinned by a test in [`regression/regression_test.go`](regression/regression_test.go)). | Silent wrong answers are more dangerous than errors. |
| **Memory-hungry at scale** | About 100 bytes allocated per data point (~99 MB for 1,000,000 points), 84 % of it inside the package. | Fine for small and medium data; large jobs need care. |
| **Pre-1.0 package** | Version v0.13.0; pkg.go.dev does not mark it as a stable (v1+) module. | Its API may change between versions — pin the version, as `go.mod` does. |
| **No visualization or notebooks** | Go has no equivalent of matplotlib/ggplot2 or Jupyter in everyday use for this team. | The Anscombe Quartet exists to show that plots catch what statistics miss. |
| **Python and R code also break** | Miller's own Python program now **fails** at its plotting step on current matplotlib (`papertype` is no longer accepted), after printing correct regressions. | Version pinning matters in every language; Go's `go.mod`/`go.sum` make it the default. |

---

## Quick start

Requires **Go 1.22 or later** (tested with Go 1.27.2). Five commands from clone to results:

```bash
git clone https://github.com/mioluigi/go-anscombe-stats.git
cd go-anscombe-stats
go test ./...
go run ./cmd/anscombe
scripts/build.sh
```

The fourth command prints the results table; the last builds the executables into `dist/`.

---

## Using the program

```text
Usage: anscombe [-data file.csv] [-verbose]

Fits y = intercept + slope·x for each data set with github.com/montanaflynn/stats
and prints intercept, slope, R², and the means of x and y.

The CSV header must be "set,x,y" (one or more named sets) or "x,y" (one set).

Flags:
  -data file
    	CSV file to analyze (default: the built-in Anscombe Quartet)
  -verbose
    	show detailed (debug) log messages
```

- **With no arguments** it analyzes the built-in Anscombe Quartet; the data is embedded in the program,
  so the executable runs on its own.
- **Your own data:** a CSV whose header is `set,x,y` (one or more named sets) or `x,y` (one set), e.g.
  `go run ./cmd/anscombe -data mydata.csv`.
- **Output:** the results table goes to standard output; timestamped progress logs go to standard error,
  so `anscombe > results.txt` saves only the table.
- **Bad input is explained, never a crash.** Examples:
  ```text
  anscombe: reading mydata.csv: line 3: x value "abc" is not a finite number
  anscombe: set "user": all x values are identical, so no line can be fitted: Input is outside of range.
  ```
- **Exit codes:** `0` success · `1` the data could not be read or fitted · `2` incorrect command line.

---

## Testing

```bash
go test ./...                 # all tests
go test -cover ./...          # with coverage
go vet ./... && gofmt -l .    # static checks (gofmt prints nothing when formatting is correct)
```

| Package | What the tests check | Coverage |
|---|---|---|
| `regression` | All four Anscombe sets vs. the reference values; synthetic lines with exact answers (perfect, negative, horizontal); the Set IV repeated-x case; invalid input (NaN, ±Inf, too few points, mismatched lengths, identical x); a precision check against an independent closed-form formula at 1e-9 | 94 % |
| `dataset` | Both CSV layouts, set order, spacing, Windows line endings; every kind of bad input with its line number | 98 % |
| `data` | The embedded file: 4 sets × 11 rows and Miller's published means | — |
| `cmd/anscombe` | Builds and runs the real program: the full results table, exit codes, error messages, logs kept off standard output | end-to-end |
| `cmd/measure` | Median calculation, argument parsing, measuring a real process, the exact output format | 93 % |

Tests were written **before** the code they test, and were themselves checked by deliberately breaking
the code (*mutation testing*). This found two real gaps that passing tests had hidden: an untested
negative-infinity case, and a test helper that accepted NaN as correct — which meant the tests could not
have caught the Set IV bug they were written for. Both were fixed.

---

## Reproducing the measurements

From the repository root, on macOS or Linux, with Go, Python 3.12+, and R (plus the `bench` package):

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r python/requirements.txt
Rscript -e 'install.packages("bench", repos="https://cloud.r-project.org")'
```

```bash
python python/anscombe_bench.py > results/python.txt
Rscript r/anscombe_bench.R > results/r.txt
go run ./cmd/anscombe 2>/dev/null > results/go.txt
go test -run='^$' -bench=. -benchmem -count=10 ./regression/ > results/go-bench.txt
scripts/measure.sh > results/whole-program.txt
scripts/accuracy_table.sh > results/accuracy.md
scripts/bench_table.sh > results/tables.md
```

**Profiles** (CPU and memory) of the largest Monte Carlo fit:

```bash
go test -run='^$' -bench='FitLine/n=1000000' -cpuprofile results/cpu.out -memprofile results/mem.out ./regression/
go tool pprof -top -nodecount=15 results/cpu.out > results/cpu-profile.txt
go tool pprof -sample_index=alloc_space -top -nodecount=15 results/mem.out > results/memory-profile.txt
```

**Fairness rules used:** every language fits the same data from the same file (`data/anscombe.csv`);
printing is kept outside timed code; medians of repeated runs are reported; all runs on one machine, on
power, in one session. Because each language measures component memory differently, those figures are
shown with their definitions, and **whole-program peak memory** is the like-for-like comparison.

---

## Building the executables

```bash
scripts/build.sh
```

| Output | What it is | How to run it |
|---|---|---|
| `dist/anscombe` | Command-line program for this computer | `dist/anscombe` |
| `dist/anscombe.exe` | Windows program, cross-compiled (`GOOS=windows GOARCH=amd64`) — no Windows machine needed | In Command Prompt: `anscombe.exe` |
| `dist/Anscombe.app` | macOS application (built on a Mac only); universal binary for Apple Silicon and Intel | Double-click in Finder, or `open dist/Anscombe.app` |

**About the `.app`:** this is a *thin wrapper* around the command-line program, built to meet the
assignment's packaging requirement. Double-clicking opens Terminal and shows the results table; progress
logs are written to `~/Library/Logs/Anscombe.log`. If macOS says the app "can't be opened", right-click it
in Finder and choose **Open** (macOS asks once for apps not from the App Store).

---

## Repository map

| Path | Role |
|---|---|
| [`regression/`](regression/) | Reusable Go package: fits y = a + b·x with `montanaflynn/stats`. No input/output. Tests and benchmarks alongside. |
| [`dataset/`](dataset/) | Reusable Go package: reads and validates `set,x,y` / `x,y` CSV data. No statistics. |
| [`data/`](data/) | `anscombe.csv` — the **single** data file read by Go, Python, and R (transcribed from Miller 2015, Table 1.1) — and the Go code that embeds it in the program. |
| [`cmd/anscombe/`](cmd/anscombe/) | The program: command-line flags, logging, results table, exit codes. |
| [`cmd/measure/`](cmd/measure/) | Tool that measures whole-program time and peak memory (macOS/Linux). |
| [`python/`](python/) | Python comparison (statsmodels and NumPy), adapted from Miller's Exhibit 1.1; pinned `requirements.txt`. |
| [`r/`](r/) | R comparison (`lm`), adapted from Miller's Exhibit 1.2. |
| [`reference/`](reference/) | Miller's original Python and R programs, **unmodified**, for traceability. |
| [`scripts/`](scripts/) | Build (`build.sh`, `build-macos-app.sh`) and measurement (`measure.sh`, `bench_table.sh`, `accuracy_table.sh`). |
| [`results/`](results/) | Every measured output used in this README. |
| [`docs/`](docs/), [`specs/`](specs/) | How the project was specified, planned, reviewed, and decided (next section). |

---

## How this project was built

The project followed **spec-driven development**: requirements were written and reviewed before any code.

1. **Constitution** ([`docs/constitution.md`](docs/constitution.md)) — non-negotiable rules derived from the
   assignment's grading rubric ([`docs/assignment-and-rubric.md`](docs/assignment-and-rubric.md)).
2. **Specification** ([`spec.md`](specs/001-anscombe-regression/spec.md)) — what and why: user stories,
   requirements, reference values, acceptance criteria, and a table tracing every rubric criterion to them.
3. **Plan** ([`plan.md`](specs/001-anscombe-regression/plan.md)) — how: design choices, each with its reason.
4. **Tasks** ([`tasks.md`](specs/001-anscombe-regression/tasks.md)) — 21 small test-first steps, one commit each.
5. **Independent reviews** ([`docs/reviews/`](docs/reviews/)) and a **decision log**
   ([`docs/decisions.md`](docs/decisions.md)) recording every finding accepted or rejected, and why.

**AI assistance, disclosed.** Claude (Anthropic) drafted the specification, plan, code, and tests under the
author's direction; Gemini (Google) reviewed the specification and plan independently; the author approved
every document, ran all measurements on his own Mac, and reviewed every change. The commit history shows
each step.

---

## References

- Anscombe, F. J. 1973. "Graphs in Statistical Analysis." *The American Statistician* 27 (1): 17–21.
- Bates, Mark, and Cory LaNou. 2023. *Go Fundamentals*. Addison-Wesley. Chapter 7, Testing.
- Bodner, Jon. 2024. *Learning Go: An Idiomatic Approach to Real-World Go Programming*, 2nd ed. O'Reilly. Chapter 15, Writing Tests.
- James, Chris. 2023. *Learn Go with Tests*. https://quii.gitbook.io/learn-go-with-tests/
- McConnell, Steve. 2004. *Code Complete*, 2nd ed. Microsoft Press. Chapter 32, Self-Documenting Code.
- Miller, Thomas W. 2015. *Modeling Techniques in Predictive Analytics with Python and R*. Pearson. Chapter 1.
- Simion, Adelina. 2023. *Test-Driven Development in Go*. Packt. Chapters 1–2.
- Flynn, Montana. `stats` package, v0.13.0. https://pkg.go.dev/github.com/montanaflynn/stats
