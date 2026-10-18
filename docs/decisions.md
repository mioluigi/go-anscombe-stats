# Decision log

Design decisions made **after** the plan was approved, newest first. Each entry: date, decision,
reason, and which document was updated. Decisions made during planning are recorded in
`specs/001-anscombe-regression/plan.md` and `docs/reviews/`.

| Date | Decision | Reason | Updated |
|---|---|---|---|
| 2026-10-18 | Whole-program max RSS (FR-13) is the primary cross-language memory comparison; per-component memory figures are reported with their definitions but not ranked against each other. | Found in T12: Go, Python, and R measure component memory differently, and R's profiler misses small allocations (reported ~2 KB for four `lm()` fits). | plan.md §8 |
