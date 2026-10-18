# Review 001 — Spec 001 reviewed by Gemini

| Field | Value |
|---|---|
| Artifact reviewed | `specs/001-anscombe-regression/spec.md` (draft v1) |
| Reviewer | Gemini (web), prompted as a strict grader and senior engineer |
| Checked against | `docs/constitution.md`, `docs/assignment-and-rubric.md` |
| Date | 2026-10-18 |
| Outcome | 6 accepted, 2 accepted with changes, 1 rejected → spec draft v2 |

Every review finding gets a decision and a reason. AI reviewers are useful but not always right;
recording *why* each point was accepted or rejected is part of the engineering record.

| # | Finding (summarized) | Decision | Reason / change made |
|---|---|---|---|
| 1 | No requirement for **program logs**, although the rubric asks for logs and profiles. | ✅ Accept | Added **FR-12** (`log/slog` to stderr, `-verbose` flag) and **AC-8**. |
| 2 | No **whole-program** memory monitoring; `-benchmem` only covers components. | ✅ Accept | Added **FR-13** (`/usr/bin/time -l`: wall time, max RSS for Go, Python, R) and **AC-9**. Also makes the cross-language comparison fairer, since it includes interpreter start-up. |
| 3 | Constitution 3.3 requires **synthetic test data**, but no AC enforces it. | ✅ Accept | Added **AC-1a**: perfect line, negative slope, horizontal line. |
| 4 | The **Set IV edge case** is identified in section 6 but no AC verifies it. | ✅ Accept | Added **AC-1b**: repeated-x derivation and the all-identical-x error case. |
| 5 | Python/R **performance metrics** have no defined units, tools, or tolerance. | 🔶 Accept with change | Units and tools added to **FR-8** (median µs over N = 1,000 runs; `tracemalloc` / `bench::mark` KB). *Rejected* the idea of a performance tolerance: performance is measured and reported, not pass/fail. |
| 6 | Rubric asks for **.exe or .app**; a plain macOS binary may lose points. | 🔶 Accept with change | Strictly, the rubric says ".exe **or** .app", so a Windows `.exe` satisfies it. Still cheap insurance: **Q5** now recommends building both, and **NFR-6 / AC-4** updated. |
| 7 | The testing rubric block would fail on synthetic data. | ✅ Accept | Same root cause as #3; resolved by AC-1a. |
| 8 | **R reference values** are "still to confirm", so AC-3 is untestable. | ❌ Reject (clarified wording) | The expected values are already defined in section 5; R must reproduce them. Running R only confirms them. Reworded section 5 so this is unambiguous. |
| 9 | The spec is still a **draft with open questions**, so it can't be signed off. | ✅ Accept (by design) | Correct, and intentional: section 11 is where the owner decides. The spec is signed off once Luis answers Q1–Q6. |

## Process change from this review

The Gemini Code Assist VS Code extension no longer supports individual accounts. Reviews now
use Gemini on the web. Constitution Article 7.2 was updated to match.
