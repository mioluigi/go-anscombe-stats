package regression

import (
	"bytes"
	"errors"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/mioluigi/go-anscombe-stats/data"
	"github.com/mioluigi/go-anscombe-stats/dataset"
	"github.com/montanaflynn/stats"
)

const (
	// referenceTolerance matches the 4-decimal precision of the spec's reference table (FR-6).
	referenceTolerance = 1e-3
	// exactTolerance is for answers known exactly, such as a perfect line.
	exactTolerance = 1e-12
	// precisionTolerance bounds the difference from independent closed-form OLS (plan §4).
	precisionTolerance = 1e-9
)

// TestFitLineAnscombe checks every set against spec section 5, which agrees with
// statsmodels, NumPy, and Miller's published summary (ŷ = 3 + 0.5x, x̄ = 9, ȳ = 7.5).
func TestFitLineAnscombe(t *testing.T) {
	want := map[string]Fit{
		"I":   {Intercept: 3.0001, Slope: 0.5001, RSquared: 0.6665, MeanX: 9, MeanY: 7.5},
		"II":  {Intercept: 3.0009, Slope: 0.5000, RSquared: 0.6662, MeanX: 9, MeanY: 7.5},
		"III": {Intercept: 3.0025, Slope: 0.4997, RSquared: 0.6663, MeanX: 9, MeanY: 7.5},
		"IV":  {Intercept: 3.0017, Slope: 0.4999, RSquared: 0.6667, MeanX: 9, MeanY: 7.5},
	}
	sets := anscombeSets(t)
	if len(sets) != len(want) {
		t.Fatalf("got %d Anscombe sets, want %d", len(sets), len(want))
	}
	for _, set := range sets {
		t.Run("set "+set.Name, func(t *testing.T) {
			got, err := FitLine(set.X, set.Y)
			if err != nil {
				t.Fatalf("FitLine: %v", err)
			}
			assertFit(t, got, want[set.Name], referenceTolerance)
		})
	}
}

// TestFitLineSyntheticKnownAnswers uses lines whose exact answers are known (AC-1a).
func TestFitLineSyntheticKnownAnswers(t *testing.T) {
	tests := []struct {
		name string
		x, y []float64
		want Fit
	}{
		{
			name: "perfect line y = 2x + 1",
			x:    []float64{1, 2, 3, 4, 5},
			y:    []float64{3, 5, 7, 9, 11},
			want: Fit{Intercept: 1, Slope: 2, RSquared: 1, MeanX: 3, MeanY: 7},
		},
		{
			name: "negative slope y = 10 - 3x, unsorted x",
			x:    []float64{4, 1, 3, 2},
			y:    []float64{-2, 7, 1, 4},
			want: Fit{Intercept: 10, Slope: -3, RSquared: 1, MeanX: 2.5, MeanY: 2.5},
		},
		{
			// Correlation is undefined (0/0) when every y is equal. The package returns 0
			// with no error, while R reports NA and NumPy returns nan. This test pins that
			// behavior so the README's warning to data scientists stays true.
			name: "horizontal line y = 4: package reports R² = 0",
			x:    []float64{1, 2, 3, 4},
			y:    []float64{4, 4, 4, 4},
			want: Fit{Intercept: 4, Slope: 0, RSquared: 0, MeanX: 2.5, MeanY: 4},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FitLine(tc.x, tc.y)
			if err != nil {
				t.Fatalf("FitLine: %v", err)
			}
			assertFit(t, got, tc.want, exactTolerance)
		})
	}
}

// TestFitLineRepeatedX covers the Set IV pattern: most x values are equal, so
// coefficients must be derived from points with different x (AC-1b).
func TestFitLineRepeatedX(t *testing.T) {
	x := []float64{8, 8, 8, 8, 19}
	y := []float64{6, 7, 8, 9, 12.5}
	got, err := FitLine(x, y)
	if err != nil {
		t.Fatalf("FitLine: %v", err)
	}
	wantIntercept, wantSlope := closedFormOLS(x, y)
	assertClose(t, "intercept", got.Intercept, wantIntercept, precisionTolerance)
	assertClose(t, "slope", got.Slope, wantSlope, precisionTolerance)
}

// TestFitLineInvalidInput checks that every bad input is rejected with the right
// sentinel error, never a panic or a silently wrong answer (AC-1b, AC-5).
func TestFitLineInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		x, y    []float64
		wantErr error
	}{
		{"all x identical", []float64{5, 5, 5}, []float64{1, 2, 3}, ErrConstantX},
		{"different lengths", []float64{1, 2, 3}, []float64{1, 2}, ErrLengthMismatch},
		{"no points", nil, nil, ErrTooFewPoints},
		{"one point", []float64{1}, []float64{2}, ErrTooFewPoints},
		{"NaN in x", []float64{1, math.NaN(), 3}, []float64{1, 2, 3}, ErrNotFinite},
		{"+Inf in y", []float64{1, 2, 3}, []float64{1, math.Inf(1), 3}, ErrNotFinite},
		{"-Inf in x", []float64{math.Inf(-1), 2, 3}, []float64{1, 2, 3}, ErrNotFinite},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FitLine(tc.x, tc.y)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("FitLine error = %v, want errors.Is(%v); fit = %+v", err, tc.wantErr, got)
			}
		})
	}
}

// TestFitLineConstantXKeepsLibraryError checks the wrapped error chain: callers
// can match our sentinel or the library's own error (plan §4).
func TestFitLineConstantXKeepsLibraryError(t *testing.T) {
	_, err := FitLine([]float64{2, 2}, []float64{1, 3})
	if !errors.Is(err, stats.ErrBounds) {
		t.Errorf("error %v does not wrap stats.ErrBounds", err)
	}
}

// TestFitLinePrecision compares FitLine with an independent textbook formula.
// FitLine back-derives coefficients from the package's fitted values; this test
// proves that costs no meaningful precision, on Anscombe and on 10,000 random points.
func TestFitLinePrecision(t *testing.T) {
	samples := map[string][2][]float64{}
	for _, set := range anscombeSets(t) {
		samples["Anscombe "+set.Name] = [2][]float64{set.X, set.Y}
	}
	x, y := noisyLine(10_000, 42)
	samples["10,000 random points"] = [2][]float64{x, y}

	for name, sample := range samples {
		t.Run(name, func(t *testing.T) {
			got, err := FitLine(sample[0], sample[1])
			if err != nil {
				t.Fatalf("FitLine: %v", err)
			}
			wantIntercept, wantSlope := closedFormOLS(sample[0], sample[1])
			assertClose(t, "intercept", got.Intercept, wantIntercept, precisionTolerance)
			assertClose(t, "slope", got.Slope, wantSlope, precisionTolerance)
		})
	}
}

// --- test helpers ---

// anscombeSets loads the shared data file, so tests use the same data as every program.
func anscombeSets(t testing.TB) []dataset.Set {
	t.Helper()
	sets, err := dataset.Read(bytes.NewReader(data.AnscombeCSV))
	if err != nil {
		t.Fatalf("reading Anscombe data: %v", err)
	}
	return sets
}

// closedFormOLS is the textbook least-squares estimate, written independently of
// the stats package: slope = Σ(x−x̄)(y−ȳ) / Σ(x−x̄)², intercept = ȳ − slope·x̄.
func closedFormOLS(x, y []float64) (intercept, slope float64) {
	n := float64(len(x))
	var sumX, sumY float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX, meanY := sumX/n, sumY/n
	var sumXY, sumXX float64
	for i := range x {
		sumXY += (x[i] - meanX) * (y[i] - meanY)
		sumXX += (x[i] - meanX) * (x[i] - meanX)
	}
	slope = sumXY / sumXX
	return meanY - slope*meanX, slope
}

// noisyLine generates y = 3 + 0.5x + noise with a fixed seed, so runs are repeatable.
func noisyLine(n int, seed uint64) (x, y []float64) {
	random := rand.New(rand.NewPCG(seed, seed))
	x = make([]float64, n)
	y = make([]float64, n)
	for i := range x {
		x[i] = random.Float64() * 20
		y[i] = 3 + 0.5*x[i] + random.NormFloat64()
	}
	return x, y
}

func assertFit(t *testing.T, got, want Fit, tolerance float64) {
	t.Helper()
	assertClose(t, "intercept", got.Intercept, want.Intercept, tolerance)
	assertClose(t, "slope", got.Slope, want.Slope, tolerance)
	assertClose(t, "R²", got.RSquared, want.RSquared, tolerance)
	assertClose(t, "mean x", got.MeanX, want.MeanX, tolerance)
	assertClose(t, "mean y", got.MeanY, want.MeanY, tolerance)
}

func assertClose(t *testing.T, label string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("%s = %.10g, want %.10g (tolerance %g)", label, got, want, tolerance)
	}
}
