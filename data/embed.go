// Package data holds the Anscombe Quartet as the single shared data file for
// the Go, Python, and R programs, and embeds it into the Go binary.
//
// The CSV lives beside this file because //go:embed cannot reach parent
// directories; keeping it here gives all three languages one source of truth
// (spec FR-1) while letting the compiled program run with no data files.
package data

import _ "embed"

// AnscombeCSV is data/anscombe.csv: header "set,x,y", then 11 rows for each of
// sets I-IV, transcribed from Miller (2015) Table 1.1.
//
//go:embed anscombe.csv
var AnscombeCSV []byte
