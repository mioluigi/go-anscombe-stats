# CLAUDE.md — project context

Go for Statistics (Week 4): validate `github.com/montanaflynn/stats` linear regression against Python and R
on the Anscombe Quartet, compare performance, and recommend to management whether data scientists can use Go.

## Read these first (in order)

1. `docs/constitution.md` — non-negotiable rules (rubric-derived)
2. `specs/001-anscombe-regression/spec.md` — what and why (APPROVED)
3. `specs/001-anscombe-regression/plan.md` — how (APPROVED)
4. `specs/001-anscombe-regression/tasks.md` — ordered checklist; work the next unchecked task

## Working rules

- One task per commit; message starts with the task ID (`T05: …`).
- Test first: write the failing test, run it, then implement.
- If implementation shows the spec or plan is wrong, update the document first, then the code.
- Record design changes in `docs/decisions.md`.
- Never hand-type results into the README; copy them from files in `results/`.

## Key commands

```bash
go test ./...                                         # all tests
go test -cover ./...                                  # coverage (≥ 90% for regression, dataset)
go vet ./... && gofmt -l .                            # lint (gofmt -l should print nothing)
go test -run='^$' -bench=. -benchmem ./regression/    # benchmarks
go run ./cmd/anscombe                                 # run the program on the bundled data
```

## Layout

`regression/` fitting (no I/O) · `dataset/` CSV reading · `data/` the shared Anscombe CSV, embedded ·
`cmd/anscombe/` the program · `python/`, `r/` comparison programs · `reference/` Miller's originals (do not edit) ·
`scripts/` measurement and packaging · `results/` captured outputs.
