### Component benchmark: fit all four Anscombe sets once

| Language | Library | Median time (µs) | Relative to Go | Memory measure | Memory (KB) |
|---|---|---:|---:|---|---:|
| Go | montanaflynn/stats | 0.9 | 1× | allocated per run (`-benchmem`) | 2.5 |
| Python | statsmodels | 252.0 | 273× | peak traced (`tracemalloc`) | 8.4 |
| Python | numpy | 123.1 | 133× | peak traced (`tracemalloc`) | 10.5 |
| R | lm | 838.9 | 908× | profiled allocations (`bench`) ¹ | 1.8 |

Go: median of 10 `go test -bench` runs; ns/op ÷ 1,000 = µs and B/op ÷ 1,024 = KB. Python and R: median of 1,000 timed runs after 10 warm-up runs.
The three memory measures count different things and are not ranked against each other; 
see the whole-program table for a like-for-like memory comparison.
¹ R's memory profiler misses many small allocations, so this understates `lm()`'s true use.

### Monte Carlo benchmark (Go): one fit at increasing data size

| Points (n) | Median time (µs) | Allocated per fit (MB) | Allocations per fit |
|---:|---:|---:|---:|
| 100 | 1.8 | 0.01 | 8 |
| 1000 | 20.9 | 0.06 | 12 |
| 10000 | 209.2 | 0.81 | 19 |
| 100000 | 2154.5 | 10.06 | 29 |
| 1000000 | 19631.0 | 99.22 | 39 |

Synthetic data y = 3 + 0.5x + noise with a fixed random seed; median over the `-count` runs.

### Whole program: analysis run once, including start-up

| Language | Median wall time (ms) | Peak memory, max RSS (MB) |
|---|---:|---:|
| Go | 11.80 | 5.0 |
| Python (pandas + statsmodels + NumPy) | 1873.69 | 162.2 |
| R (base `lm`) | 154.28 | 76.0 |

Median of 5 runs measured by `cmd/measure` (wall clock from Go's monotonic timer; peak memory from the operating system's rusage report).
