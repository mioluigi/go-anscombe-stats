# Project Constitution — Go for Statistics (Anscombe Quartet)

> **Purpose of this file.** These are the non-negotiable rules for the project. Every spec, plan,
> task, and line of code is checked against them. If a later decision conflicts with this file,
> either the decision changes or this file is amended deliberately (with a note in `docs/decisions.md`).
>
> **Source of authority.** The rules below are derived directly from the Week 4 assignment and its
> **Programming Assignment Rubric** (5 criteria × 20 pts = 100 pts). Each rule is tagged with the
> rubric criterion it protects: **[C]** Coding, **[T]** Testing & metrics, **[D]** Design,
> **[Doc]** Documentation, **[A]** Application.

---

## Article 1 — Correctness before speed

1.1 The Go results are only trusted when they match an **independent reference** (Python and R) within a stated numeric tolerance. **[T]**
1.2 Every numeric claim in the README must be reproducible by a command a grader can run. **[Doc]**
1.3 No result is hand-typed into the README. Results are produced by programs and copied from their output. **[Doc]**

## Article 2 — Idiomatic, readable Go (McConnell, Ch. 32) **[C]**

2.1 Use **Go modules** (`go.mod`), a current stable Go toolchain, and `gofmt`/`go vet` clean code.
2.2 Names are meaningful and specific (e.g., `fitLine`, `anscombeSets`) — not abstract (`data`, `doIt`, `tmp`), not overly long.
2.3 No useless temporary variables or intermediate results.
2.4 Functions do **one thing**. Large blocks are split into small, named functions.
2.5 Control flow is plain `if/else` and `for`. In complex Boolean expressions, **distribute the NOT** (write `!a && !b`, not `!(a || b)`).
2.6 Code is self-documenting: every exported identifier has a doc comment; comments explain *why*, not *what*.
2.7 Errors are returned and handled, never silently ignored (no `_ =` on errors from the stats package).

## Article 3 — Test-first, measured **[T]**

3.1 Write the failing test before the implementation for every critical component (Learn Go with Tests style).
3.2 Use the standard library `testing` package as the foundation. (`testify` is permitted for assertion readability — see open question in spec.)
3.3 Prefer **table-driven tests**; include edge cases and **synthetic test data** with known answers (e.g., a perfect line `y = 2x + 1`).
3.4 Floating-point comparisons always use an explicit tolerance constant — never `==`.
3.5 Benchmarks use `testing.B` with `-benchmem` so **time and memory** are both reported.
3.6 Profiles (CPU/memory) are generated at least once and their location documented.

## Article 4 — Clean, reusable design **[D]**

4.1 Statistics logic lives in a **reusable package** separate from the command-line program (`main`).
4.2 The reusable package does not print, read files, or call `os.Exit`. Input/output belongs in `main`.
4.3 Data are stored as plain files (CSV) shared by the Go, Python, and R programs so all three analyze the **identical** inputs.
4.4 Python and R programs are kept minimal and comparable: same data, same model (simple OLS of y on x), same timing approach.

## Article 5 — Documentation and Git hygiene **[Doc]**

5.1 One `README.md` at the repository root is the complete documentation: purpose, roles of programs and data, how to build, run, test, and benchmark, results, and the **recommendation to management**.
5.2 Commit small and often, one task per commit, with clear messages (`feat:`, `test:`, `docs:`…).
5.3 Specs, plan, tasks, and decisions are versioned in the repo under `docs/` and `specs/`.

## Article 6 — A finished, friendly application **[A]**

6.1 The program builds into a standalone executable with documented commands (macOS binary; Windows `.exe` via cross-compilation).
6.2 It **runs to completion without issues** on the bundled data with no arguments.
6.3 Any user input is validated; bad input produces a clear, specific message and a non-zero exit code — never a panic or stack trace.
6.4 Output is clean and readable (aligned table), suitable for a non-programmer.

## Article 7 — How we work with AI tools

7.1 **Claude Code** is the only tool that writes files in this repo during a task.
7.2 **Gemini (web, gemini.google.com)** independently reviews specs, plans, and diffs. It does not edit the repo. (The Gemini Code Assist VS Code extension no longer supports individual accounts, so it is not used.)
7.3 Luis approves every spec, plan, and task list before implementation, and reads every diff before it is committed.
7.4 If implementation reveals the spec is wrong, **update the spec first**, then the code.
