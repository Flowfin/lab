package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Flowfin/lab/internal/check"
)

// realWalk is the walk the command uses when it is not being tested. A test
// that wants the ordinary behaviour passes this, so a case exercising a
// contrived walk is visible as one at the call site.
var realWalk = check.Walk

// TestExitCodes reaches every code this runner can return, and names the
// record that decides what each one means. Decision record 0011 requires a
// test per code, so that a code existing only in the document is caught here
// rather than by an operator.
func TestExitCodes(t *testing.T) {
	refusing := func(string) (check.Result, error) {
		return check.Result{
			Root:               "somewhere",
			ExperimentsPresent: true,
			Directories:        1,
			Records:            1,
			Refusals:           []string{"a refusal"},
		}, nil
	}
	failing := func(string) (check.Result, error) {
		return check.Result{}, errors.New("cannot read the tree")
	}

	tests := []struct {
		name string
		args []string
		walk func(string) (check.Result, error)
		want int
	}{
		{
			name: "a walk that refused nothing",
			args: []string{"check", "../../testdata/cases/one-experiment-with-a-record/tree"},
			walk: realWalk,
			want: exitClean,
		},
		{
			name: "help",
			args: []string{"help"},
			walk: realWalk,
			want: exitClean,
		},
		{
			// Nothing refuses anything yet, so the walk is the parameter
			// here. What is proved is the mapping: a result carrying a
			// refusal leaves this command with the code a refusal returns,
			// and that is settled before the first refusal is written
			// rather than after it.
			name: "a walk that refused something",
			args: []string{"check", "."},
			walk: refusing,
			want: exitRefused,
		},
		{
			name: "a walk that could not read the tree",
			args: []string{"check", "."},
			walk: failing,
			want: exitCannot,
		},
		{
			name: "a path that is not a directory",
			args: []string{"check", "main.go"},
			walk: realWalk,
			want: exitCannot,
		},
		{
			name: "a path that is not there",
			args: []string{"check", "no-such-directory"},
			walk: realWalk,
			want: exitCannot,
		},
		{
			name: "a verb the runner does not have",
			args: []string{"inspect"},
			walk: realWalk,
			want: exitCannot,
		},
		{
			name: "no verb at all",
			args: nil,
			walk: realWalk,
			want: exitCannot,
		},
		{
			name: "more paths than check takes",
			args: []string{"check", "one", "two"},
			walk: realWalk,
			want: exitCannot,
		},
		{
			name: "help with an argument",
			args: []string{"help", "check"},
			walk: realWalk,
			want: exitCannot,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			got := run(tc.args, &out, &errOut, tc.walk)
			if got != tc.want {
				t.Fatalf("exit code %d, want %d\nstdout: %s\nstderr: %s",
					got, tc.want, out.String(), errOut.String())
			}
		})
	}
}

// TestRefusalsAreNamedInTheOutput holds the half of record 0011 that the exit
// code cannot carry. A caller reading only the number learns that something
// was refused; a caller that wants to know what was refused reads the output,
// and that only works if the output names it.
func TestRefusalsAreNamedInTheOutput(t *testing.T) {
	refusing := func(string) (check.Result, error) {
		return check.Result{
			Root:               "somewhere",
			ExperimentsPresent: true,
			Refusals:           []string{"the refusal a reader has to see"},
		}, nil
	}

	var out, errOut bytes.Buffer
	if got := run([]string{"check", "."}, &out, &errOut, refusing); got != exitRefused {
		t.Fatalf("exit code %d, want %d", got, exitRefused)
	}
	if !strings.Contains(out.String(), "the refusal a reader has to see") {
		t.Fatalf("the output does not name the refusal:\n%s", out.String())
	}
}

// TestTheDocumentedCodesAreTheNumbersTheRecordFixes pins the three numbers
// this command uses to the three record 0011 states. Renumbering any of them
// reddens the suite, which is what a workflow keyed on one of these numbers
// needs, since such a workflow is a reader of that record and would otherwise
// change meaning without its own file being edited.
//
// It does not notice a fifth constant added beside these three. Nothing here
// reads the record, so what would catch that is the review, and record 0011
// says so in its own text rather than leaving this test to imply otherwise.
func TestTheDocumentedCodesAreTheNumbersTheRecordFixes(t *testing.T) {
	codes := map[string]int{
		"clean":   exitClean,
		"refused": exitRefused,
		"cannot":  exitCannot,
	}
	want := map[string]int{"clean": 0, "refused": 1, "cannot": 2}

	for name, code := range codes {
		if code != want[name] {
			t.Errorf("%s is %d, record 0011 says %d", name, code, want[name])
		}
	}
	for name := range want {
		if _, ok := codes[name]; !ok {
			t.Errorf("record 0011 names %s and this command has no constant for it", name)
		}
	}
}
