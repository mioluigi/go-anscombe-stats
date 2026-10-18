#!/usr/bin/env bash
# Whole-program measurement (spec FR-13, plan section 8).
#
# Runs each complete program -- Go, Python, R -- doing the analysis once (load
# data, fit all four sets, print the table) and reports the median wall-clock
# time and peak memory over RUNS runs (default 5). Unlike the component
# benchmarks, this includes start-up: launching an interpreter and importing
# libraries. The measuring itself is done by cmd/measure.
#
# Usage, from the repository root with the Python virtual environment active:
#   scripts/measure.sh > results/whole-program.txt
#   RUNS=11 scripts/measure.sh > results/whole-program.txt

set -euo pipefail

RUNS="${RUNS:-5}"
cd "$(dirname "$0")/.."

if ! [[ "$RUNS" =~ ^[1-9][0-9]*$ ]]; then
  echo "measure.sh: RUNS must be a positive whole number, got '$RUNS'" >&2
  exit 2
fi

machine="$(uname -sm)"
if [[ "$(uname)" == "Darwin" ]]; then
  machine="$machine · $(sysctl -n machdep.cpu.brand_string) · macOS $(sw_vers -productVersion)"
fi

echo "Whole-program measurement: analysis once per run, median of $RUNS runs"
echo "Machine: $machine"
echo "Go: $(go version | cut -d' ' -f3) · Python: $(python3 --version | cut -d' ' -f2) · R: $(Rscript -e 'cat(as.character(getRversion()))')"
echo

mkdir -p dist
go build -o dist/anscombe ./cmd/anscombe
go build -o dist/measure ./cmd/measure

dist/measure -runs "$RUNS" \
  "go=dist/anscombe" \
  "python=python3 python/anscombe_bench.py --fit-only" \
  "r=Rscript r/anscombe_bench.R --fit-only"
