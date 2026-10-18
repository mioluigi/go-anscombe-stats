// Command anscombe fits a simple linear regression of y on x for each data set
// using github.com/montanaflynn/stats and prints a results table.
//
// With no arguments it analyzes the built-in Anscombe Quartet (Anscombe 1973).
// Results go to standard output; progress logs go to standard error, so
// "anscombe > results.txt" saves a clean table.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"

	"github.com/mioluigi/go-anscombe-stats/data"
	"github.com/mioluigi/go-anscombe-stats/dataset"
	"github.com/mioluigi/go-anscombe-stats/regression"
)

// Exit codes tell scripts what happened without parsing messages.
const (
	exitOK        = 0 // results printed
	exitDataError = 1 // the data could not be read or fitted; the message explains why
	exitUsage     = 2 // the command line was wrong; usage is printed
)

const builtInSource = "built-in Anscombe Quartet"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// options holds the parsed command line.
type options struct {
	dataPath string
	verbose  bool
}

// setResult pairs a data set's name and size with its fitted line.
type setResult struct {
	name string
	n    int
	fit  regression.Fit
}

// run carries out one invocation and returns the exit code. Keeping it apart
// from main makes the whole program testable without terminating the test.
func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}
	if err != nil {
		return exitUsage
	}

	logger := newLogger(stderr, opts.verbose)
	started := time.Now()
	logger.Debug("options", "data", opts.dataPath, "verbose", opts.verbose)

	sets, source, err := loadSets(opts.dataPath)
	if err != nil {
		fmt.Fprintf(stderr, "anscombe: %v\n", err)
		return exitDataError
	}
	logger.Info("data loaded", "source", source, "sets", len(sets))

	results, err := fitAll(sets, logger)
	if err != nil {
		fmt.Fprintf(stderr, "anscombe: %v\n", err)
		return exitDataError
	}

	if err := printTable(stdout, source, results); err != nil {
		fmt.Fprintf(stderr, "anscombe: writing results: %v\n", err)
		return exitDataError
	}
	logger.Info("finished", "elapsed", time.Since(started))
	return exitOK
}

// parseOptions reads the command line. The flag package prints its own error
// message and the usage text before returning an error.
func parseOptions(args []string, stderr io.Writer) (options, error) {
	var opts options
	flags := flag.NewFlagSet("anscombe", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&opts.dataPath, "data", "", "CSV `file` to analyze (default: the "+builtInSource+")")
	flags.BoolVar(&opts.verbose, "verbose", false, "show detailed (debug) log messages")
	flags.Usage = func() {
		fmt.Fprint(flags.Output(), `Usage: anscombe [-data file.csv] [-verbose]

Fits y = intercept + slope·x for each data set with github.com/montanaflynn/stats
and prints intercept, slope, R², and the means of x and y.

The CSV header must be "set,x,y" (one or more named sets) or "x,y" (one set).

Flags:
`)
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected argument %q (use -data to name a file)\n", flags.Arg(0))
		flags.Usage()
		return options{}, errors.New("unexpected argument")
	}
	return opts, nil
}

// newLogger writes timestamped key=value logs to stderr (spec FR-12).
func newLogger(stderr io.Writer, verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level}))
}

// loadSets reads the user's file, or the embedded Anscombe data when no file
// is given, and reports which source was used.
func loadSets(path string) (sets []dataset.Set, source string, err error) {
	if path == "" {
		sets, err = dataset.Read(bytes.NewReader(data.AnscombeCSV))
		return sets, builtInSource, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("opening data file: %w", err)
	}
	defer file.Close()

	sets, err = dataset.Read(file)
	if err != nil {
		return nil, "", fmt.Errorf("reading %s: %w", path, err)
	}
	return sets, path, nil
}

// fitAll fits every set, stopping at the first set that cannot be fitted.
func fitAll(sets []dataset.Set, logger *slog.Logger) ([]setResult, error) {
	results := make([]setResult, 0, len(sets))
	for _, set := range sets {
		fit, err := regression.FitLine(set.X, set.Y)
		if err != nil {
			return nil, fmt.Errorf("set %q: %w", set.Name, err)
		}
		logger.Info("set fitted", "set", set.Name, "n", len(set.X))
		logger.Debug("coefficients", "set", set.Name,
			"intercept", fit.Intercept, "slope", fit.Slope, "r_squared", fit.RSquared)
		results = append(results, setResult{name: set.Name, n: len(set.X), fit: fit})
	}
	return results, nil
}

// printTable writes the results as aligned columns, rounded to 4 decimals to
// match the precision of the published reference values.
func printTable(stdout io.Writer, source string, results []setResult) error {
	fmt.Fprintf(stdout, "Linear regression of y on x (github.com/montanaflynn/stats)\nData: %s\n\n", source)

	table := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "Set\tn\tIntercept\tSlope\tR²\tMean x\tMean y")
	for _, result := range results {
		fit := result.fit
		fmt.Fprintf(table, "%s\t%d\t%.4f\t%.4f\t%.4f\t%.4f\t%.4f\n",
			result.name, result.n, fit.Intercept, fit.Slope, fit.RSquared, fit.MeanX, fit.MeanY)
	}
	return table.Flush()
}
