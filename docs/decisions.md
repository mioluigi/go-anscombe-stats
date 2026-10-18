# Decision log

Design decisions made **after** the plan was approved, newest first. Each entry: date, decision,
reason, and which document was updated. Decisions made during planning are recorded in
`specs/001-anscombe-regression/plan.md` and `docs/reviews/`.

| Date | Decision | Reason | Updated |
|---|---|---|---|
| 2026-10-18 | Whole-program max RSS (FR-13) is the primary cross-language memory comparison; per-component memory figures are reported with their definitions but not ranked against each other. | Found in T12: Go, Python, and R measure component memory differently, and R's profiler misses small allocations (reported ~2 KB for four `lm()` fits). | plan.md §8 |
| 2026-10-18 | Whole-program metrics measured by a Go tool (`cmd/measure`) reading rusage, instead of `/usr/bin/time` directly; Python and R gain a `--fit-only` mode. | Found in T14: `/usr/bin/time` reports time in 10 ms steps and showed the Go program as 0.000 s. Rusage is the same memory source `/usr/bin/time` uses. `--fit-only` makes the measured run the analysis, not the 1,000-run benchmark. | spec.md FR-13, AC-9; plan.md §1, §8; tasks.md T14 |
