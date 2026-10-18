package regression

import (
	"fmt"
	"testing"
)

// benchmarkSink keeps results alive so the compiler cannot optimize away the
// work being measured.
var benchmarkSink Fit

// BenchmarkFitAnscombe times the unit of work shared with the Python and R
// programs: fitting all four Anscombe regressions once (plan §8). Loading the
// data happens before the timer starts, so only fitting is measured.
//
//	go test -run='^$' -bench=FitAnscombe -benchmem -count=10 ./regression/
func BenchmarkFitAnscombe(b *testing.B) {
	sets := anscombeSets(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		for _, set := range sets {
			fit, err := FitLine(set.X, set.Y)
			if err != nil {
				b.Fatal(err)
			}
			benchmarkSink = fit
		}
	}
}

// BenchmarkFitLine is the Monte Carlo benchmark (decision Q4): one fit on
// y = 3 + 0.5x + noise at increasing sizes, showing how time and memory grow
// with the data. A fixed seed makes every run use identical data.
//
//	go test -run='^$' -bench=FitLine -benchmem ./regression/
func BenchmarkFitLine(b *testing.B) {
	for _, size := range []int{100, 1_000, 10_000, 100_000, 1_000_000} {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			x, y := noisyLine(size, 42)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				fit, err := FitLine(x, y)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkSink = fit
			}
		})
	}
}
