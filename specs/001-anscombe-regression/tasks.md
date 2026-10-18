# Tasks 001 — Ordered checklist to implement Plan 001

| Field | Value |
|---|---|
| Status | **APPROVED — 2026-10-18** (Luis: Claude implements all code tasks, including the former 🎓 tasks) |
| Implements | `plan.md` (APPROVED v2) → `spec.md` (APPROVED v2) |

## How to work this list

- **One task = one commit.** Commit message starts with the task ID, e.g. `T05: implement dataset.Read`.
- **Test first.** Tasks marked 🔴 write failing tests; the next task 🟢 makes them pass. Run the tests in
  between and *see them fail* — a test that has never failed has never proven anything.
- **Who:** 🤖 Claude writes it and pushes · 👤 Luis runs it on his Mac (timings and R must come from
  his machine).
- **After every task:** `git pull`, read the diff, run the "Verify" command. Tick the box when it passes.
- **Done when** = the observable result that proves the task is finished.

---

## Phase 0 — Setup

- [x] **T00 👤 Check tools on the Mac** (plan §11)
  Run: `go version; python3 --version; R --version | head -1; Rscript -e 'library(bench)' && echo "bench OK"`
  Done when: all four print a version / "bench OK"; anything missing installed per plan §11.

- [x] **T01 🤖 Project scaffolding**
  `CLAUDE.md` (project context and key commands), `.gitignore` (`dist/`, `*.out`, `.venv/`, `__pycache__/`), `docs/decisions.md` (empty log).
  Done when: files committed. Verify: `cat CLAUDE.md`

- [x] **T02 🤖 Go module**
  `go mod init github.com/mioluigi/go-anscombe-stats`; require `github.com/montanaflynn/stats v0.13.0`.
  Done when: `go.mod` and `go.sum` committed. Verify: `go mod verify`

## Phase 1 — Data (FR-1)

- [x] **T03 🤖 Shared data file + embed package + test**
  `data/anscombe.csv` (44 rows, `set,x,y`, transcribed from Miller Table 1.1), `data/embed.go`, `data/embed_test.go` (4 sets × 11 rows; overall mean x = 9, mean y ≈ 7.5).
  Done when: test passes. Verify: `go test ./data/`

## Phase 2 — Package `dataset` (FR-9, FR-10)

- [x] **T04 🔴🤖 Tests for `dataset.Read`** (table-driven)
  Cases: valid `set,x,y`; valid `x,y` → one set named "user"; sets returned in first-seen order; errors with **line numbers** for non-numeric value, wrong column count, empty input, unknown header.
  Done when: `go test ./dataset/` **fails** (no implementation yet).

- [x] **T05 🟢 Implement `dataset.Read`**
  Done when: `go test -cover ./dataset/` passes with ≥ 90% coverage.

## Phase 3 — Package `regression` (FR-2, FR-3, FR-6)

- [x] **T06 🔴 Tests for `regression.FitLine`** (table-driven)
  - Anscombe: 4 sets × intercept, slope, R², mean x, mean y vs. spec §5 (tolerance 1e-3)
  - Synthetic (AC-1a): y = 2x + 1; a negative slope; a horizontal line (R² pinned at 0 — package behavior)
  - Set IV pattern (AC-1b); all-identical x → `errors.Is(err, ErrConstantX)` **and** `errors.Is(err, stats.ErrBounds)`
  - Invalid input (AC-5): length mismatch, < 2 points, NaN, ±Inf
  - Precision guard: vs. independent closed-form OLS written in the test, tolerance 1e-9
  Done when: `go test ./regression/` **fails**.

- [x] **T07 🟢 Implement `regression.FitLine`**
  Including the *why* comment on back-deriving coefficients (plan §4) and wrapped errors.
  Done when: `go test -cover ./regression/` passes with ≥ 90% coverage; `go vet ./...` clean.

- [x] **T08 🤖 Benchmarks** (FR-7, Q4)
  `BenchmarkFitAnscombe` (the shared unit of work: all four fits) and `BenchmarkFitLine/n=100 … n=1000000` (synthetic, fixed seed).
  Done when: `go test -run=^$ -bench=. -benchmem ./regression/` prints ns/op, B/op, allocs/op.

## Phase 4 — The program `cmd/anscombe` (FR-9, FR-10, FR-12)

- [x] **T09 🔴 End-to-end tests**
  Build the binary once in `TestMain`, then run it: no arguments → exit 0 and a 4-row table; `-data missing.csv` → exit 1 with a clear message; bad flag → exit 2; logs appear on stderr, never stdout.
  Done when: `go test ./cmd/anscombe/` **fails**.

- [x] **T10 🟢🤖 Implement `main.go`**
  Flags (`-data`, `-verbose`), `slog` to stderr, `tabwriter` table, exit codes 0/1/2, short named functions.
  Done when: `go test ./...` passes and `go run ./cmd/anscombe` prints the table.

## Phase 5 — Python and R comparisons (FR-4, FR-5, FR-8)

- [x] **T11 🤖 Python program**
  `python/anscombe_bench.py` (statsmodels primary, NumPy second row; reads `data/anscombe.csv`; coefficients, then timing N = 1,000 with median µs and `tracemalloc` KB) + pinned `python/requirements.txt`.
  Done when: runs in Claude's workspace and prints coefficients matching spec §5.

- [x] **T12 🤖 R program**
  `r/anscombe_bench.R` (`lm` per set; `bench::mark` 1,000 iterations, median µs, `mem_alloc` KB).
  Done when: committed (R runs only on Luis's Mac — next task).

- [x] **T13 👤 Run Python and R on the Mac; save results**
  `python3 -m venv .venv && source .venv/bin/activate && pip install -r python/requirements.txt`
  `python python/anscombe_bench.py > results/python.txt` · `Rscript r/anscombe_bench.R > results/r.txt`
  Done when: both files committed; **R coefficients match spec §5 to 4 decimals** (AC-3).

## Phase 6 — Measurement (FR-7, FR-8, FR-13, NFR-3)

- [x] **T14 🤖 Measurement scripts**
  `cmd/measure` (Go tool, test-first: runs a program N times; median wall ms and max RSS MB from rusage), `scripts/measure.sh` (builds and measures Go/Python/R; `--fit-only` added to the Python and R programs), and `scripts/bench_table.sh` (converts Go ns/op → µs, B → KB for the README).
  Done when: scripts committed with usage comments.

- [ ] **T15 👤 Run all measurements on the Mac**
  `go test -run=^$ -bench=. -benchmem -count=10 ./regression/ > results/go-bench.txt`, then `scripts/measure.sh`, then the profile commands (plan §8). Record machine details.
  Done when: `results/` committed with all outputs.

## Phase 7 — Executables (NFR-6, AC-4, Q5)

- [ ] **T16 🤖 Build scripts**
  `scripts/build-macos-app.sh` (creates `dist/Anscombe.app`) and documented `.exe` cross-compile command.
  Done when: committed.

- [ ] **T17 👤 Build and try all three**
  Build the binary, the `.exe`, and the `.app`; double-click `Anscombe.app` and confirm the table appears.
  Done when: all three build; `.app` shows the results.

## Phase 8 — Documentation and submission (FR-11, US-5)

- [ ] **T18 🤖 README**
  Every FR-11 section, results tables in µs/KB (from `results/`), data-scientist concerns (incl. flat-line R² = 0, coefficients not returned directly, no standard errors/p-values), recommendation to management.
  Done when: every item in AC-6 and AC-9 is present.

- [ ] **T19 👤 Gemini review of README and code**
  Upload README, `regression.go`, `main.go`, and the rubric summary; triage into `docs/reviews/003-…`.
  Done when: review recorded and accepted changes made.

- [ ] **T20 👤🤖 Final acceptance check and submission**
  Walk spec §8 (every AC) and the rubric table; fresh `git clone` into a temporary folder and follow the README from scratch.
  Done when: every box ticked; `https://github.com/mioluigi/go-anscombe-stats.git` pasted into Canvas.

---

## Traceability

| Acceptance criterion | Tasks |
|---|---|
| AC-1, AC-1a, AC-1b | T06, T07 |
| AC-2 | T08, T15 |
| AC-3 | T11, T12, T13 |
| AC-4 | T09, T10, T16, T17 |
| AC-5 | T04, T05, T06, T09 |
| AC-6, AC-9 | T14, T15, T18 |
| AC-7 | T20 |
| AC-8 | T09, T10 |
