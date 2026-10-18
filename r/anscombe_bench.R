# Anscombe Quartet regressions in R, for comparison with the Go program.
#
# Adapted from Miller (2015) Exhibit 1.2 (reference/miller-mtpa-chapter-1-program.R):
#   * the same base-R lm(y ~ x) fit for each set;
#   * reads the shared data/anscombe.csv instead of hard-coding the data;
#   * plotting removed (not needed for the comparison);
#   * adds timing and memory measurement with the bench package, using the same
#     unit of work as Go and Python: fitting all four regressions once (plan section 8).
#
# Usage (from the repository root):
#   Rscript r/anscombe_bench.R [repetitions]      default repetitions: 1000

library(bench)

warm_up_runs <- 10

# Locate data/anscombe.csv relative to this script, so the program works from any
# working directory; fall back to the repository root when run interactively.
script_dir <- function() {
  file_arg <- grep("^--file=", commandArgs(trailingOnly = FALSE), value = TRUE)
  if (length(file_arg) == 0) {
    return("r")
  }
  dirname(normalizePath(sub("^--file=", "", file_arg)))
}

# Read the optional repetitions argument, stopping with a clear message if invalid.
read_repetitions <- function() {
  args <- commandArgs(trailingOnly = TRUE)
  if (length(args) == 0) {
    return(1000L)
  }
  repetitions <- suppressWarnings(as.integer(args[1]))
  if (is.na(repetitions) || repetitions < 1) {
    stop("repetitions must be a whole number of at least 1, got \"", args[1], "\"", call. = FALSE)
  }
  repetitions
}

# Return a list of data frames (x, y), one per set, in file order.
load_sets <- function(path) {
  anscombe <- read.csv(path)
  in_file_order <- factor(anscombe$set, levels = unique(anscombe$set))
  split(anscombe[c("x", "y")], in_file_order)
}

# Miller's method: ordinary least squares with lm(); R-squared from its summary.
fit_lm <- function(set) {
  model <- lm(y ~ x, data = set)
  coefficients <- coef(model)
  c(intercept = coefficients[[1]], slope = coefficients[[2]],
    r_squared = summary(model)$r.squared)
}

# The unit of work timed in every language: fit all four sets once.
fit_all <- function(sets) {
  lapply(sets, fit_lm)
}

print_coefficients <- function(sets) {
  cat("Linear regression of y on x (lm)\n")
  # Written literally: sprintf pads by bytes, and "R²" is 3 bytes but 2 characters wide.
  cat("Set    n  Intercept    Slope       R²   Mean x   Mean y\n")
  fits <- fit_all(sets)
  for (name in names(sets)) {
    set <- sets[[name]]
    fit <- fits[[name]]
    cat(sprintf("%-5s%3d  %9.4f  %7.4f  %7.4f  %7.4f  %7.4f\n",
                name, nrow(set), fit[["intercept"]], fit[["slope"]], fit[["r_squared"]],
                mean(set$x), mean(set$y)))
  }
  cat("\n")
}

main <- function() {
  repetitions <- read_repetitions()
  data_file <- file.path(script_dir(), "..", "data", "anscombe.csv")
  sets <- load_sets(data_file)

  cat(sprintf("%s · bench %s\n", R.version.string, packageVersion("bench")))
  cat(sprintf("Data: anscombe.csv (%d sets)\n\n", length(sets)))
  print_coefficients(sets)

  for (run in seq_len(warm_up_runs)) {
    fit_all(sets)
  }
  timing <- bench::mark(fit_all(sets), iterations = repetitions, check = FALSE)
  median_us <- as.numeric(timing$median) * 1e6
  alloc_kb <- as.numeric(timing$mem_alloc) / 1024

  cat(sprintf("Unit of work: fit all %d sets once; median of %d runs after %d warm-up runs\n",
              length(sets), repetitions, warm_up_runs))
  # One key=value line so scripts/bench_table.sh can parse it. alloc_kb comes from R's
  # memory profiler, which records large allocations but misses many small ones, so it
  # understates lm()'s true use; compare languages with whole-program memory (FR-13).
  cat(sprintf("benchmark language=r library=lm runs=%d median_us=%.1f alloc_kb=%.1f\n",
              repetitions, median_us, alloc_kb))
}

main()
