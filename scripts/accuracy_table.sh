#!/usr/bin/env bash
# Builds the README's accuracy table from the saved results: for each Anscombe
# set, the Go coefficients and whether every method (Go, Python statsmodels,
# Python NumPy, R lm) printed exactly the same values to 4 decimals. The
# agreement is computed from the files, never typed by hand (Constitution 1.3).
#
# Reads results/go.txt, results/python.txt, results/r.txt. Usage:
#   scripts/accuracy_table.sh > results/accuracy.md

set -euo pipefail
cd "$(dirname "$0")/.."

for file in results/go.txt results/python.txt results/r.txt; do
  if [[ ! -s "$file" ]]; then
    echo "accuracy_table.sh: missing or empty $file" >&2
    exit 1
  fi
done

awk '
# Each results table has a title line naming the method, then rows:
# Set n Intercept Slope R² MeanX MeanY
/^Linear regression of y on x/ {
  method = $0
  sub(/.*\(/, "", method); sub(/\).*/, "", method)
  if (method ~ /montanaflynn/) method = "Go (montanaflynn/stats)"
  if (!(method in seen)) { seen[method] = 1; methods[++methodCount] = method }
}
$1 ~ /^(I|II|III|IV)$/ && $2 ~ /^[0-9]+$/ {
  set = $1
  if (!(set in setSeen)) { setSeen[set] = 1; sets[++setCount] = set }
  values[set, method] = $3 " " $4 " " $5 " " $6 " " $7
}
END {
  print "| Set | Intercept | Slope | R² | Mean x | Mean y | All " methodCount " methods identical to 4 decimals? |"
  print "|---|---:|---:|---:|---:|---:|---|"
  for (s = 1; s <= setCount; s++) {
    set = sets[s]
    reference = values[set, methods[1]]
    verdict = "✅ yes"
    for (m = 2; m <= methodCount; m++) {
      if (values[set, methods[m]] != reference) verdict = "❌ no: " methods[m] " differs"
    }
    split(reference, number, " ")
    printf "| %s | %s | %s | %s | %s | %s | %s |\n", set, number[1], number[2], number[3], number[4], number[5], verdict
  }
  printf "\nMethods compared:"
  for (m = 1; m <= methodCount; m++) printf "%s %s", (m > 1 ? "," : ""), methods[m]
  print "."
}
' results/go.txt results/python.txt results/r.txt
