// Package regression fits a simple linear regression (y on x) using the
// github.com/montanaflynn/stats package, which is the package being evaluated.
// It performs no input or output, so any Go program can reuse it.
package regression

import (
	"errors"
	"fmt"
	"math"

	"github.com/montanaflynn/stats"
)

// Sentinel errors, so callers can test the kind of problem with errors.Is.
var (
	ErrLengthMismatch = errors.New("x and y must have the same number of values")
	ErrTooFewPoints   = errors.New("at least 2 points are needed to fit a line")
	ErrNotFinite      = errors.New("values must be finite numbers (no NaN or infinity)")
	ErrConstantX      = errors.New("all x values are identical, so no line can be fitted")
)

// Fit is the result of a least-squares regression of y on x.
type Fit struct {
	Intercept float64 // estimated y when x = 0
	Slope     float64 // estimated change in y per unit change in x
	RSquared  float64 // share of the variance in y explained by x
	MeanX     float64
	MeanY     float64
}

// FitLine estimates y = Intercept + Slope·x by ordinary least squares.
// Every statistic comes from the montanaflynn/stats package.
func FitLine(x, y []float64) (Fit, error) {
	if err := validate(x, y); err != nil {
		return Fit{}, err
	}

	fitted, err := stats.LinearRegression(toSeries(x, y))
	if errors.Is(err, stats.ErrBounds) {
		// The package rejects data whose x values are all equal. Keep both errors in
		// the chain: ours explains the problem, the package's identifies the source.
		return Fit{}, fmt.Errorf("%w: %w", ErrConstantX, err)
	}
	if err != nil {
		return Fit{}, fmt.Errorf("stats.LinearRegression: %w", err)
	}
	intercept, slope := coefficientsFromFittedLine(fitted)

	// validate has already ruled out every condition these functions reject,
	// but their errors are still checked in case the package changes.
	correlation, errCorrelation := stats.Correlation(x, y)
	meanX, errMeanX := stats.Mean(x)
	meanY, errMeanY := stats.Mean(y)
	if err := errors.Join(errCorrelation, errMeanX, errMeanY); err != nil {
		return Fit{}, fmt.Errorf("summary statistics: %w", err)
	}

	return Fit{
		Intercept: intercept,
		Slope:     slope,
		RSquared:  correlation * correlation,
		MeanX:     meanX,
		MeanY:     meanY,
	}, nil
}

// validate rejects inputs that cannot produce a meaningful line.
func validate(x, y []float64) error {
	if len(x) != len(y) {
		return fmt.Errorf("%w: %d x values, %d y values", ErrLengthMismatch, len(x), len(y))
	}
	if len(x) < 2 {
		return fmt.Errorf("%w: got %d", ErrTooFewPoints, len(x))
	}
	for i := range x {
		if !isFinite(x[i]) || !isFinite(y[i]) {
			return fmt.Errorf("%w: point %d is (%v, %v)", ErrNotFinite, i+1, x[i], y[i])
		}
	}
	return nil
}

// coefficientsFromFittedLine recovers the intercept and slope from fitted points.
//
// Why this is needed: stats.LinearRegression computes the slope and intercept
// internally but returns only the fitted points (x, ŷ), discarding the
// coefficients. Two fitted points with different x determine the line exactly.
//
// Why the smallest and largest x: they always differ when any two x values
// differ (Anscombe Set IV has ten points at x = 8 and one at x = 19), and the
// widest spacing gives the most stable slope. Measured against closed-form OLS,
// the difference is about 1e-16, float64 rounding (see TestFitLinePrecision).
func coefficientsFromFittedLine(fitted stats.Series) (intercept, slope float64) {
	lowest, highest := fitted[0], fitted[0]
	for _, point := range fitted[1:] {
		if point.X < lowest.X {
			lowest = point
		}
		if point.X > highest.X {
			highest = point
		}
	}
	slope = (highest.Y - lowest.Y) / (highest.X - lowest.X)
	return lowest.Y - slope*lowest.X, slope
}

// toSeries pairs x and y values in the package's input format.
func toSeries(x, y []float64) stats.Series {
	series := make(stats.Series, len(x))
	for i := range x {
		series[i] = stats.Coordinate{X: x[i], Y: y[i]}
	}
	return series
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
