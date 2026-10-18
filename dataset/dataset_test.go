package dataset

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/mioluigi/go-anscombe-stats/data"
)

// TestReadValid covers every accepted input shape.
func TestReadValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Set
	}{
		{
			name:  "set,x,y keeps sets in first-seen order",
			input: "set,x,y\nB,1,2\nA,3,4\nB,5,6\n",
			want: []Set{
				{Name: "B", X: []float64{1, 5}, Y: []float64{2, 6}},
				{Name: "A", X: []float64{3}, Y: []float64{4}},
			},
		},
		{
			name:  "x,y becomes one set named user",
			input: "x,y\n1,2\n3,4\n",
			want:  []Set{{Name: UserSetName, X: []float64{1, 3}, Y: []float64{2, 4}}},
		},
		{
			name:  "header is case-insensitive and spaces around values are ignored",
			input: " Set , X , Y \n I , 10 , 8.04 \n",
			want:  []Set{{Name: "I", X: []float64{10}, Y: []float64{8.04}}},
		},
		{
			name:  "Windows line endings and blank lines are accepted",
			input: "x,y\r\n1,2\r\n\r\n3,4\r\n",
			want:  []Set{{Name: UserSetName, X: []float64{1, 3}, Y: []float64{2, 4}}},
		},
		{
			name:  "negative and scientific-notation numbers parse",
			input: "x,y\n-1.5,2e3\n",
			want:  []Set{{Name: UserSetName, X: []float64{-1.5}, Y: []float64{2000}}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Read(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Read returned error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Read = %+v\nwant   %+v", got, tc.want)
			}
		})
	}
}

// TestReadInvalid checks that each problem returns the right sentinel error
// and that the message points the user to the offending line.
func TestReadInvalid(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     error
		wantMessage string // must appear in err.Error()
	}{
		{"empty input", "", ErrNoData, "no data"},
		{"header only", "x,y\n", ErrNoData, "no data"},
		{"unknown header", "a,b\n1,2\n", ErrHeader, `"a,b"`},
		{"too few columns", "x,y\n1,2\n3\n", ErrColumnCount, "line 3"},
		{"too many columns", "set,x,y\nI,1,2,9\n", ErrColumnCount, "line 2"},
		{"x is not a number", "x,y\n1,2\nabc,4\n", ErrNotNumber, `line 3: x value "abc"`},
		{"y is not a number", "set,x,y\nI,1,\n", ErrNotNumber, `line 2: y value ""`},
		{"NaN is rejected", "x,y\nNaN,1\n", ErrNotNumber, "line 2"},
		{"infinity is rejected", "x,y\n1,Inf\n", ErrNotNumber, "line 2"},
		{"empty set name", "set,x,y\n,1,2\n", ErrSetName, "line 2"},
		{"malformed quoting", "x,y\n\"1,2\n", ErrMalformed, "line"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Read(strings.NewReader(tc.input))
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Read error = %v, want errors.Is(%v); sets = %+v", err, tc.wantErr, got)
			}
			if !strings.Contains(err.Error(), tc.wantMessage) {
				t.Errorf("error %q does not mention %q", err.Error(), tc.wantMessage)
			}
		})
	}
}

// TestReadAnscombe reads the real shared data file end to end.
func TestReadAnscombe(t *testing.T) {
	sets, err := Read(bytes.NewReader(data.AnscombeCSV))
	if err != nil {
		t.Fatalf("Read(anscombe.csv): %v", err)
	}
	var names []string
	for _, set := range sets {
		names = append(names, set.Name)
		if len(set.X) != 11 || len(set.Y) != 11 {
			t.Errorf("set %s has %d x and %d y values, want 11 each", set.Name, len(set.X), len(set.Y))
		}
	}
	if want := []string{"I", "II", "III", "IV"}; !reflect.DeepEqual(names, want) {
		t.Errorf("set names = %v, want %v", names, want)
	}
}
