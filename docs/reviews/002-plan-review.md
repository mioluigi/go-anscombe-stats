# Review 002 — Plan 001 reviewed by Gemini

| Field | Value |
|---|---|
| Artifact reviewed | `specs/001-anscombe-regression/plan.md` (draft v1) |
| Reviewer | Gemini (web), prompted as a strict grader and senior Go engineer |
| Checked against | `spec.md` (approved v2), `docs/constitution.md` |
| Date | 2026-10-18 |
| Outcome | 2 accepted, 2 accepted with changes, 1 rejected → plan draft v2 |

| # | Finding (summarized) | Decision | Reason / change made |
|---|---|---|---|
| 1 | Back-deriving slope and intercept from fitted values risks **floating-point precision loss**; add a *why* comment. | 🔶 Accept with change | **Measured, not assumed:** derived coefficients differ from an independent closed-form OLS by at most 4.4 × 10⁻¹⁶ — float64 rounding, ~10¹² times below the 1e-3 tolerance. Risk is negligible, but the *why* comment is good practice. Plan §4 now requires the comment **and** a precision-guard test at 1e-9 against independent closed-form OLS. |
| 2 | A separate `data/` package just for `//go:embed` is **unidiomatic**; embed in `main.go` instead. | ❌ Reject | `//go:embed` cannot reference parent directories (`..`). Embedding from `cmd/anscombe/` would force the CSV into that folder, making Python and R read from a Go command directory or duplicating the data — breaking FR-1's single source of truth. Rationale recorded in plan §3. |
| 3 | Go reports **ns/op** while Python/R report **µs**; the README needs one unit. | ✅ Accept | Plan §8: all README tables in µs (and KB); Go converted by script (ns ÷ 1,000), conversion stated under the table. |
| 4 | The `.app` Terminal wrapper is irregular but justified by the rubric; **document it as such**. | ✅ Accept | Plan §9: README states the `.app` is a thin wrapper around a command-line program. |
| 5 | Use idiomatic **error wrapping** (`%w`) rather than a bare sentinel. | 🔶 Accept with change | Plan §4: wrap with **both** our sentinel and the library error — `fmt.Errorf("%w: …: %w", ErrConstantX, err)` (Go ≥ 1.20) — so `errors.Is` works for either. Test updated to use `errors.Is`. |

## Lesson recorded

Finding 1 shows the value of **measuring before mitigating**: a plausible-sounding risk turned out to be
twelve orders of magnitude below the tolerance. Finding 2 shows that an idiomatic-sounding suggestion can
conflict with a language constraint and an existing requirement — check suggestions against both.
