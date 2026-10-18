package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath is the program built once by TestMain and run by every test, so the
// tests exercise exactly what a user runs, including real exit codes.
var binaryPath string

func TestMain(m *testing.M) {
	buildDir, err := os.MkdirTemp("", "anscombe-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, "creating build directory:", err)
		os.Exit(1)
	}
	binaryPath = filepath.Join(buildDir, "anscombe")
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building program: %v\n%s", err, output)
		os.RemoveAll(buildDir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(buildDir)
	os.Exit(code)
}

// runResult is what a user would see after running the program once.
type runResult struct {
	exitCode       int
	stdout, stderr string
}

func runProgram(t *testing.T, args ...string) runResult {
	t.Helper()
	var stdout, stderr bytes.Buffer
	command := exec.Command(binaryPath, args...)
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()

	var exitErr *exec.ExitError
	exitCode := 0
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("could not run program: %v", err)
	}
	return runResult{exitCode, stdout.String(), stderr.String()}
}

// writeFile creates a CSV in a temporary directory and returns its path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestNoArgumentsPrintsAnscombeTable is the default run (FR-9, AC-4).
func TestNoArgumentsPrintsAnscombeTable(t *testing.T) {
	result := runProgram(t)
	if result.exitCode != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", result.exitCode, exitOK, result.stderr)
	}
	rows := rowsBySetName(result.stdout)
	// Each row: set, n, intercept, slope, R², mean x, mean y — values from spec section 5.
	want := map[string][]string{
		"I":   {"I", "11", "3.0001", "0.5001", "0.6665", "9.0000", "7.5009"},
		"II":  {"II", "11", "3.0009", "0.5000", "0.6662", "9.0000", "7.5009"},
		"III": {"III", "11", "3.0025", "0.4997", "0.6663", "9.0000", "7.5000"},
		"IV":  {"IV", "11", "3.0017", "0.4999", "0.6667", "9.0000", "7.5009"},
	}
	for name, wantFields := range want {
		if got := rows[name]; strings.Join(got, " ") != strings.Join(wantFields, " ") {
			t.Errorf("row %s = %v, want %v\nfull output:\n%s", name, got, wantFields, result.stdout)
		}
	}
}

// rowsBySetName splits the results table into fields, keyed by each row's first field.
func rowsBySetName(table string) map[string][]string {
	rows := map[string][]string{}
	for _, line := range strings.Split(table, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			rows[fields[0]] = fields
		}
	}
	return rows
}

// TestLogsGoToStandardError keeps the table clean for redirection (FR-12, AC-8).
func TestLogsGoToStandardError(t *testing.T) {
	result := runProgram(t)
	if !strings.Contains(result.stderr, "level=INFO") {
		t.Errorf("stderr has no INFO log lines:\n%s", result.stderr)
	}
	if strings.Contains(result.stdout, "level=") {
		t.Errorf("log lines leaked into stdout:\n%s", result.stdout)
	}
	if strings.Contains(result.stderr, "level=DEBUG") {
		t.Errorf("debug lines shown without -verbose:\n%s", result.stderr)
	}
}

func TestVerboseShowsDebugLogs(t *testing.T) {
	result := runProgram(t, "-verbose")
	if !strings.Contains(result.stderr, "level=DEBUG") {
		t.Errorf("-verbose produced no DEBUG lines:\n%s", result.stderr)
	}
}

// TestUserFile reads a user's own two-column CSV (FR-9).
func TestUserFile(t *testing.T) {
	path := writeFile(t, "line.csv", "x,y\n1,3\n2,5\n3,7\n")
	result := runProgram(t, "-data", path)
	if result.exitCode != exitOK {
		t.Fatalf("exit code = %d, want %d; stderr:\n%s", result.exitCode, exitOK, result.stderr)
	}
	for _, want := range []string{"user", "1.0000", "2.0000"} {
		if !strings.Contains(result.stdout, want) {
			t.Errorf("table does not contain %q:\n%s", want, result.stdout)
		}
	}
}

// TestDataProblemsExplainAndExit1 covers bad input files (FR-10, AC-5).
func TestDataProblemsExplainAndExit1(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantMessage string
	}{
		{"missing file", []string{"-data", "no-such-file.csv"}, "no-such-file.csv"},
		{"non-numeric value", []string{"-data", writeFile(t, "bad.csv", "x,y\n1,2\nabc,4\n")}, "line 3"},
		{"identical x values", []string{"-data", writeFile(t, "flat.csv", "x,y\n5,1\n5,2\n")}, "all x values are identical"},
		{"single point", []string{"-data", writeFile(t, "one.csv", "x,y\n1,2\n")}, "at least 2 points"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := runProgram(t, tc.args...)
			if result.exitCode != exitDataError {
				t.Errorf("exit code = %d, want %d", result.exitCode, exitDataError)
			}
			if !strings.Contains(result.stderr, tc.wantMessage) {
				t.Errorf("stderr does not mention %q:\n%s", tc.wantMessage, result.stderr)
			}
			if strings.Contains(result.stderr, "panic") || strings.Contains(result.stderr, "goroutine") {
				t.Errorf("program crashed instead of explaining:\n%s", result.stderr)
			}
		})
	}
}

// TestUsageProblemsExit2 covers command-line mistakes.
func TestUsageProblemsExit2(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"unknown flag", []string{"-bogus"}},
		{"unexpected argument", []string{"extra.csv"}},
		{"-data without a value", []string{"-data"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := runProgram(t, tc.args...)
			if result.exitCode != exitUsage {
				t.Errorf("exit code = %d, want %d; stderr:\n%s", result.exitCode, exitUsage, result.stderr)
			}
			if !strings.Contains(result.stderr, "Usage") {
				t.Errorf("stderr does not show usage:\n%s", result.stderr)
			}
		})
	}
}

func TestHelpExitsCleanly(t *testing.T) {
	result := runProgram(t, "-h")
	if result.exitCode != exitOK {
		t.Errorf("exit code = %d, want %d", result.exitCode, exitOK)
	}
	if !strings.Contains(result.stderr, "-data") {
		t.Errorf("help does not describe -data:\n%s", result.stderr)
	}
}
