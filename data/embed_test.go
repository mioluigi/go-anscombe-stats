package data

import (
	"bytes"
	"encoding/csv"
	"math"
	"slices"
	"strconv"
	"testing"
)

// Published summary for every Anscombe set (Miller 2015, p. 9; Anscombe 1973).
const (
	publishedMeanX = 9.0
	publishedMeanY = 7.5
	rowsPerSet     = 11
	// Miller rounds mean y to 7.5; the four sets' exact means are 7.4991 to 7.5009.
	meanTolerance = 0.01
)

// TestAnscombeCSV checks the embedded file against Miller's Table 1.1 summary.
// It parses the CSV with the standard library only, so it stays independent
// of the dataset package that will be tested against this same file.
func TestAnscombeCSV(t *testing.T) {
	records, err := csv.NewReader(bytes.NewReader(AnscombeCSV)).ReadAll()
	if err != nil {
		t.Fatalf("embedded CSV does not parse: %v", err)
	}

	wantHeader := []string{"set", "x", "y"}
	if got := records[0]; !slices.Equal(got, wantHeader) {
		t.Fatalf("header = %v, want %v", got, wantHeader)
	}

	sumsBySet := map[string]*setSums{}
	for line, record := range records[1:] {
		sums, seen := sumsBySet[record[0]]
		if !seen {
			sums = &setSums{}
			sumsBySet[record[0]] = sums
		}
		sums.add(t, line+2, record[1], record[2])
	}

	for _, name := range []string{"I", "II", "III", "IV"} {
		sums, seen := sumsBySet[name]
		if !seen {
			t.Errorf("set %s is missing", name)
			continue
		}
		if sums.count != rowsPerSet {
			t.Errorf("set %s has %d rows, want %d", name, sums.count, rowsPerSet)
		}
		checkMean(t, name, "x", sums.sumX/float64(sums.count), publishedMeanX)
		checkMean(t, name, "y", sums.sumY/float64(sums.count), publishedMeanY)
	}
	if len(sumsBySet) != 4 {
		t.Errorf("found %d sets, want 4", len(sumsBySet))
	}
}

// setSums accumulates the running totals needed for one set's means.
type setSums struct {
	count      int
	sumX, sumY float64
}

func (s *setSums) add(t *testing.T, line int, xText, yText string) {
	t.Helper()
	x, xErr := strconv.ParseFloat(xText, 64)
	y, yErr := strconv.ParseFloat(yText, 64)
	if xErr != nil || yErr != nil {
		t.Fatalf("line %d: x=%q y=%q are not both numbers", line, xText, yText)
	}
	s.count++
	s.sumX += x
	s.sumY += y
}

func checkMean(t *testing.T, set, variable string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > meanTolerance {
		t.Errorf("set %s: mean %s = %.4f, want %.1f ± %.2f", set, variable, got, want, meanTolerance)
	}
}
