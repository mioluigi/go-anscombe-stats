# Review 003 — Repository graded by Gemini against the rubric

| Field | Value |
|---|---|
| Artifacts reviewed | `README.md`, `regression/regression.go`, `cmd/anscombe/main.go` |
| Reviewer | Gemini (web), prompted as a strict grader using `docs/assignment-and-rubric.md` |
| Date | 2026-10-18 |
| Outcome | Scored 100/100 (20/20 on each criterion); no findings; no changes required |

| Criterion | Gemini's score | Evidence cited (summarized) |
|---|---|---|
| Coding rules, organization, aesthetics | 20/20 | Specific names, single-purpose functions, distributed NOT, why-comments |
| Testing and software metrics | 20/20 | Coverage, synthetic data, `slog` logs on stderr, component and whole-program metrics, profiles, Monte Carlo |
| Design and development | 20/20 | `regression` has no I/O; sentinel errors wrapped with `%w` |
| Documentation | 20/20 | Recommendation, data-scientist concerns, named Python/R packages, repository map, quick start |
| Application | 20/20 | Three executables, embedded data, validated and explained input, aligned table |

## Verification of the review's evidence

A review with no findings is not proof that nothing is wrong, so each cited fact was checked against the code.
The conclusions hold; two citations were imprecise:

1. The distributed-NOT expression `!math.IsNaN(value) && !math.IsInf(value, 0)` is in `isFinite()`, not
   `validate()` as cited. `validate()` applies the same rule (`!isFinite(x[i]) || !isFinite(y[i])`), so the
   conclusion stands.
2. The review says `main.go` maps sentinel errors to exit codes. In fact every data error exits with code 1;
   sentinel errors shape the message, and the three exit codes distinguish success, data problems, and
   command-line problems — as designed (plan §6).

## Lesson recorded

Across three reviews, Gemini moved from finding real gaps (review 001) to mixed findings (002) to none (003),
and its evidence was not always exact. Reviews are inputs to judgment, not verdicts: the final acceptance
check (task T20) runs the README's instructions on a fresh clone rather than relying on any reviewer.
