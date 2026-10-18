#!/usr/bin/env bash
# Builds the README's performance tables from the saved results, so no number is
# ever typed by hand (Constitution 1.3). All times are shown in microseconds or
# milliseconds and all memory in KB or MB, with conversions stated under each table.
#
# Reads (from the repository root):
#   results/go-bench.txt       go test -bench output, run with -count for repeats
#   results/python.txt         "benchmark ..." lines from python/anscombe_bench.py
#   results/r.txt              "benchmark ..." lines from r/anscombe_bench.R
#   results/whole-program.txt  "wholeprogram ..." lines from scripts/measure.sh
#
# Usage:
#   scripts/bench_table.sh > results/tables.md
#
# Uses only POSIX awk features, so it runs with the awk shipped with macOS.

set -euo pipefail
cd "$(dirname "$0")/.."

files=(results/go-bench.txt results/python.txt results/r.txt results/whole-program.txt)
for file in "${files[@]}"; do
  if [[ ! -s "$file" ]]; then
    echo "bench_table.sh: missing or empty $file (see tasks T13 and T15)" >&2
    exit 1
  fi
done

awk '
# median of values[1..count], sorted in place (insertion sort; counts are small)
function median(values, count,    i, j, held) {
  for (i = 2; i <= count; i++) {
    held = values[i]
    for (j = i - 1; j >= 1 && values[j] > held; j--) values[j + 1] = values[j]
    values[j + 1] = held
  }
  if (count % 2) return values[(count + 1) / 2]
  return (values[count / 2] + values[count / 2 + 1]) / 2
}

# field returns the value of key=value in the current line
function field(key,    i, pair) {
  for (i = 1; i <= NF; i++) {
    split($i, pair, "=")
    if (pair[1] == key) return pair[2]
  }
  return ""
}

# Go benchmark lines: Name-PROCS  iterations  ns/op  B/op  allocs/op
FILENAME ~ /go-bench/ && /^Benchmark/ {
  name = $1
  sub(/-[0-9]+$/, "", name)
  if (!(name in runs)) order[++names] = name
  count = ++runs[name]
  nanoseconds[name, count] = $3
  bytes[name, count] = $5
  allocations[name, count] = $7
}

/^benchmark / { component[field("language") "," field("library")] = $0; componentOrder[++components] = field("language") "," field("library") }
/^wholeprogram / { whole[++wholes] = $0 }

END {
  for (n = 1; n <= names; n++) {
    name = order[n]
    for (i = 1; i <= runs[name]; i++) { t[i] = nanoseconds[name, i]; b[i] = bytes[name, i]; a[i] = allocations[name, i] }
    medianNs[name] = median(t, runs[name]); medianB[name] = median(b, runs[name]); medianAllocs[name] = median(a, runs[name])
  }
  goMicroseconds = medianNs["BenchmarkFitAnscombe"] / 1000

  print "### Component benchmark: fit all four Anscombe sets once"
  print ""
  print "| Language | Library | Median time (µs) | Relative to Go | Memory measure | Memory (KB) |"
  print "|---|---|---:|---:|---|---:|"
  printf "| Go | montanaflynn/stats | %.1f | 1× | allocated per run (`-benchmem`) | %.1f |\n", goMicroseconds, medianB["BenchmarkFitAnscombe"] / 1024
  for (c = 1; c <= components; c++) {
    $0 = component[componentOrder[c]]
    language = field("language") == "python" ? "Python" : "R"
    microseconds = field("median_us")
    if (field("peak_kb") != "") { measure = "peak traced (`tracemalloc`)"; kilobytes = field("peak_kb") }
    else { measure = "profiled allocations (`bench`) ¹"; kilobytes = field("alloc_kb") }
    printf "| %s | %s | %.1f | %.0f× | %s | %.1f |\n", language, field("library"), microseconds, microseconds / goMicroseconds, measure, kilobytes
  }
  print ""
  printf "Go: median of %d `go test -bench` runs; ns/op ÷ 1,000 = µs and B/op ÷ 1,024 = KB. ", runs["BenchmarkFitAnscombe"]
  print "Python and R: median of 1,000 timed runs after 10 warm-up runs."
  print "The three memory measures count different things and are not ranked against each other; "
  print "see the whole-program table for a like-for-like memory comparison."
  print "¹ R'\''s memory profiler misses many small allocations, so this understates `lm()`'\''s true use."
  print ""

  print "### Monte Carlo benchmark (Go): one fit at increasing data size"
  print ""
  print "| Points (n) | Median time (µs) | Allocated per fit (MB) | Allocations per fit |"
  print "|---:|---:|---:|---:|"
  for (n = 1; n <= names; n++) {
    name = order[n]
    if (name !~ /^BenchmarkFitLine\/n=/) continue
    size = name; sub(/^BenchmarkFitLine\/n=/, "", size)
    printf "| %d | %.1f | %.2f | %d |\n", size, medianNs[name] / 1000, medianB[name] / 1048576, medianAllocs[name]
  }
  print ""
  print "Synthetic data y = 3 + 0.5x + noise with a fixed random seed; median over the `-count` runs."
  print ""

  print "### Whole program: analysis run once, including start-up"
  print ""
  print "| Language | Median wall time (ms) | Peak memory, max RSS (MB) |"
  print "|---|---:|---:|"
  for (w = 1; w <= wholes; w++) {
    $0 = whole[w]
    language = field("language")
    label = language == "go" ? "Go" : (language == "python" ? "Python (pandas + statsmodels + NumPy)" : "R (base `lm`)")
    printf "| %s | %.2f | %.1f |\n", label, field("median_wall_ms"), field("median_maxrss_mb")
    wholeRuns = field("runs")
  }
  print ""
  printf "Median of %d runs measured by `cmd/measure` (wall clock from Go'\''s monotonic timer; peak memory from the operating system'\''s rusage report).\n", wholeRuns
}
' "${files[@]}"
