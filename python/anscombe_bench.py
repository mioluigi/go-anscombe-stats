"""Anscombe Quartet regressions in Python, for comparison with the Go program.

Adapted from Miller (2015) Exhibit 1.1 (reference/miller-mtpa-chapter-1-program.py):
  * the same statsmodels OLS fit (sm.OLS with sm.add_constant) for each set;
  * reads the shared data/anscombe.csv instead of hard-coding the data;
  * plotting removed (the assignment says to ignore it, and Miller's savefig
    call fails on current matplotlib);
  * adds numpy.polyfit as a second, lighter-weight Python method;
  * adds timing and memory measurement of the same unit of work as Go:
    fitting all four regressions once (plan section 8).

Usage (from the repository root, inside the virtual environment):
    python python/anscombe_bench.py [--repetitions 1000]
"""

import argparse
import platform
import statistics
import time
import tracemalloc
from pathlib import Path

import numpy as np
import pandas as pd
import statsmodels
import statsmodels.api as sm

DATA_FILE = Path(__file__).resolve().parent.parent / "data" / "anscombe.csv"
WARM_UP_RUNS = 10


def load_sets(path):
    """Return {set name: (x array, y array)} in file order."""
    frame = pd.read_csv(path)
    return {
        name: (group["x"].to_numpy(), group["y"].to_numpy())
        for name, group in frame.groupby("set", sort=False)
    }


def fit_statsmodels(x, y):
    """Miller's method: OLS on a design matrix with an intercept column."""
    result = sm.OLS(y, sm.add_constant(x)).fit()
    intercept, slope = result.params
    return intercept, slope, result.rsquared


def fit_numpy(x, y):
    """Least-squares line with numpy.polyfit; R² from the correlation."""
    slope, intercept = np.polyfit(x, y, 1)
    r_squared = np.corrcoef(x, y)[0, 1] ** 2
    return intercept, slope, r_squared


METHODS = {"statsmodels": fit_statsmodels, "numpy": fit_numpy}


def fit_all(fit, sets):
    """The unit of work timed in every language: fit all four sets once."""
    return [fit(x, y) for x, y in sets.values()]


def print_coefficients(method_name, fit, sets):
    print(f"Linear regression of y on x ({method_name})")
    print(f"{'Set':<5}{'n':>3}  {'Intercept':>9}  {'Slope':>7}  {'R²':>7}  {'Mean x':>7}  {'Mean y':>7}")
    for (name, (x, y)), (intercept, slope, r_squared) in zip(sets.items(), fit_all(fit, sets)):
        print(f"{name:<5}{len(x):>3}  {intercept:>9.4f}  {slope:>7.4f}  {r_squared:>7.4f}"
              f"  {x.mean():>7.4f}  {y.mean():>7.4f}")
    print()


def median_microseconds(fit, sets, repetitions):
    """Median wall time of one unit of work, after warm-up runs."""
    for _ in range(WARM_UP_RUNS):
        fit_all(fit, sets)
    durations = []
    for _ in range(repetitions):
        start = time.perf_counter_ns()
        fit_all(fit, sets)
        durations.append(time.perf_counter_ns() - start)
    return statistics.median(durations) / 1_000


def peak_kilobytes(fit, sets):
    """Peak memory allocated by Python during one unit of work.

    Measured in a separate run because tracemalloc slows the code it traces.
    """
    tracemalloc.start()
    fit_all(fit, sets)
    _, peak = tracemalloc.get_traced_memory()
    tracemalloc.stop()
    return peak / 1024


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--repetitions", type=int, default=1000,
                        help="timed runs of the unit of work (default: 1000)")
    args = parser.parse_args()
    if args.repetitions < 1:
        parser.error("--repetitions must be at least 1")

    sets = load_sets(DATA_FILE)
    print(f"Python {platform.python_version()} · numpy {np.__version__} · "
          f"pandas {pd.__version__} · statsmodels {statsmodels.__version__}")
    print(f"Data: {DATA_FILE.name} ({len(sets)} sets)\n")

    for method_name, fit in METHODS.items():
        print_coefficients(method_name, fit, sets)

    print(f"Unit of work: fit all {len(sets)} sets once; median of {args.repetitions} runs "
          f"after {WARM_UP_RUNS} warm-up runs")
    for method_name, fit in METHODS.items():
        median_us = median_microseconds(fit, sets, args.repetitions)
        peak_kb = peak_kilobytes(fit, sets)
        # One key=value line per method so scripts/bench_table.sh can parse it.
        print(f"benchmark language=python library={method_name} "
              f"runs={args.repetitions} median_us={median_us:.1f} peak_kb={peak_kb:.1f}")


if __name__ == "__main__":
    main()
