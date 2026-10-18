// Package dataset reads paired (x, y) observations from CSV text and groups
// them into named sets. It knows nothing about statistics, so it can be
// reused by any program that needs validated two-column numeric data.
//
// Two layouts are accepted:
//
//	set,x,y   one row per observation; rows sharing a set name form one set
//	x,y       a single set, named UserSetName
//
// Header names are case-insensitive and surrounding spaces are ignored.
package dataset

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// UserSetName names the single set read from a file with an "x,y" header.
const UserSetName = "user"

// Sentinel errors, so callers can test the kind of problem with errors.Is.
// Every returned error also says which line caused it.
var (
	ErrNoData      = errors.New("no data rows found")
	ErrHeader      = errors.New(`unrecognized header; expected "set,x,y" or "x,y"`)
	ErrColumnCount = errors.New("wrong number of columns")
	ErrNotNumber   = errors.New("not a finite number")
	ErrSetName     = errors.New("set name is empty")
	ErrMalformed   = errors.New("malformed CSV")
)

// Set is one named group of paired observations; X[i] pairs with Y[i].
type Set struct {
	Name string
	X, Y []float64
}

// Read parses CSV from r and returns its sets in the order each set name
// first appears. It stops at the first problem and reports its line number.
func Read(r io.Reader) ([]Set, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 // column counts are checked below, with clearer messages

	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: the input is empty", ErrNoData)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	hasSetColumn, err := parseHeader(header)
	if err != nil {
		return nil, err
	}

	var sets []Set
	positionByName := map[string]int{}
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
		}
		line, _ := reader.FieldPos(0)
		name, x, y, err := parseRecord(record, hasSetColumn, line)
		if err != nil {
			return nil, err
		}

		position, seen := positionByName[name]
		if !seen {
			position = len(sets)
			positionByName[name] = position
			sets = append(sets, Set{Name: name})
		}
		sets[position].X = append(sets[position].X, x)
		sets[position].Y = append(sets[position].Y, y)
	}

	if len(sets) == 0 {
		return nil, fmt.Errorf("%w: the input has a header but no data rows", ErrNoData)
	}
	return sets, nil
}

// parseHeader reports whether the header includes a "set" column.
func parseHeader(header []string) (hasSetColumn bool, err error) {
	normalized := make([]string, len(header))
	for i, field := range header {
		normalized[i] = strings.ToLower(strings.TrimSpace(field))
	}
	switch strings.Join(normalized, ",") {
	case "set,x,y":
		return true, nil
	case "x,y":
		return false, nil
	}
	return false, fmt.Errorf("line 1: header %q: %w", strings.Join(header, ","), ErrHeader)
}

// parseRecord converts one CSV row into a set name and an (x, y) pair.
func parseRecord(record []string, hasSetColumn bool, line int) (name string, x, y float64, err error) {
	wantColumns, name := 2, UserSetName
	if hasSetColumn {
		wantColumns = 3
	}
	if len(record) != wantColumns {
		return "", 0, 0, fmt.Errorf("line %d: %w: expected %d, found %d",
			line, ErrColumnCount, wantColumns, len(record))
	}

	if hasSetColumn {
		name = strings.TrimSpace(record[0])
		if name == "" {
			return "", 0, 0, fmt.Errorf("line %d: %w", line, ErrSetName)
		}
		record = record[1:]
	}
	if x, err = parseFinite(record[0], "x", line); err != nil {
		return "", 0, 0, err
	}
	if y, err = parseFinite(record[1], "y", line); err != nil {
		return "", 0, 0, err
	}
	return name, x, y, nil
}

// parseFinite parses text as a float64, rejecting NaN and infinities because
// they cannot be meaningful observations.
func parseFinite(text, column string, line int) (float64, error) {
	trimmed := strings.TrimSpace(text)
	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("line %d: %s value %q is %w", line, column, trimmed, ErrNotNumber)
	}
	return value, nil
}
