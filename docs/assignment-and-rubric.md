# Assignment Requirements and Rubric (summary)

> Summary of the Week 4 assignment, *Go for Statistics* (100 points), in our own words, so that every
> tool and reviewer can check the spec against it. The Canvas assignment page is the authoritative source.

## Scenario

A startup wants Go as its main language for both engineers and data scientists. Data scientists are
skeptical. They want evidence that the Go package `github.com/montanaflynn/stats` gives correct
answers compared with Python and R, starting with the Anscombe Quartet (Anscombe 1973; Miller 2015).

## Requirements

1. Acting as the company's data scientists, use the Anscombe Quartet data and Go's `testing` package to show that the Go package gives results **comparable to Python and R**, especially the **estimated linear regression coefficients**.
2. Use the Go `testing` package to measure **execution time**, and compare it with execution times of Python and R programs on the same data.
3. Be explicit about **which Python and R packages** are used (e.g., statsmodels vs. NumPy/SciPy vs. scikit-learn), since they can give different results.
4. The README.md must include a **recommendation to management** and the **concerns data scientists might have** about using the Go package instead of Python or R.

## Deliverables

- Cloneable GitHub repository URL **ending in `.git`**, submitted in the Canvas comments.
- `README.md` with complete documentation, including user instructions for running and testing the program.
- Repository includes complete **data and code** for Go and for R and/or Python, and shows **execution times and memory requirements**.

## Provided materials

- Miller (2015) Chapter 1 sample (Table 1.1 data; published result ŷ = 3 + 0.5x).
- `miller-mtpa-chapter-1-program.py` (statsmodels OLS; plotting can be ignored).
- `miller-mtpa-chapter-1-program.R` (base `lm()`).

## Rubric (5 criteria × 20 points)

Each criterion is scored: Outstanding 20 (all criteria met) · Meets Expectations 15 (most) · Needs Improvement 10 (many not met) · Partially Acceptable 5 (most not met) · Unacceptable 0.

| Criterion | What earns full marks |
|---|---|
| **Coding rules, organization, aesthetics** | Go modules and idiomatic Go. Readable code with meaningful, specific names (not too long or short). No useless temporary variables. Clean, consistent blocks and line breaks. Large blocks split into single-purpose pieces. Simple control flow (if/else, for). Distribute the *not* in complex Boolean expressions. Self-documenting code with comments explaining the logic (McConnell 2004, ch. 32). |
| **Testing and software metrics** | Unit tests of critical components, with synthetic test data where appropriate. Program logs and profiles where appropriate. Monitor memory and processing requirements of components and the whole program. Monte Carlo performance benchmark only if the requirements call for it. |
| **Design and development** | Clean, efficient, easy-to-understand design that meets all requirements and serves the use case. General-purpose, reusable modules where possible. |
| **Documentation** | Effective use of Git/GitHub. README.md that explains the roles of programs and data and how to test and use the application. |
| **Application** | Instructions for creating an executable (.exe on Windows or .app on macOS). Runs to completion without issues. Validates any user input and explains incorrect input. Clean UX/UI. |

## References named in the assignment

Anscombe (1973); Bates & LaNou (2023) ch. 7; Bodner (2024) ch. 15; James, *Learn Go with Tests*;
McConnell (2004) ch. 32; Miller (2015) ch. 1; Simion (2023) ch. 1–2; `testify` (optional).