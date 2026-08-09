// The harness every later refusal is proved with.
//
// A case is a directory under testdata/cases/. It holds the tree the runner
// walked, as files in the repository rather than as strings assembled at run
// time, so a reader can look at exactly the input the walk saw. It holds what
// that walk should have found, and it holds the whole set of refusals it
// should have produced.
//
// The set is the part that matters. A fixture that trips the refusal it was
// written for and one other proves less than it claims, and only comparing
// the whole set notices, which is why nothing here asks whether a particular
// refusal is present.
//
// The layout of a case:
//
//	testdata/cases/<name>/tree/               what the runner walks
//	testdata/cases/<name>/expected            four lines, see readExpectation
//	testdata/cases/<name>/expected-refusals   one property per line, may be empty
//	testdata/cases/<name>/near-neighbour      the case that differs by the
//	                                          smallest legal change, required
//	                                          of a case that refuses
//
// This file decides that layout. Nothing restates it, so there is nothing to
// drift against it.
//
// THE BOUND ON WHAT ANY OF THIS PROVES. Every comparison here is over which
// properties were refused, and never over which line inside the runner refused
// them. Two refusal sites producing one property are indistinguishable to
// every leg, so an operator holding several of them can lose one and stay
// green. That is the state of this tree rather than a hypothetical:
// record-is-not-text is produced at two sites, one for a null byte and one for
// an invalid sequence, and a fixture for either satisfies the standing
// requirement for both. What catches the loss of one arm is a case existing
// for that arm, which is a fixture somebody remembered and not something the
// harness can require. Read this before quoting a green run as proof that a
// refusal site is exercised.
package check

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// fixedNow is the one notion of now every test here runs against. The runner
// reads the time in exactly one place and everything downstream is given the
// value, so the suite supplies a date rather than whatever the machine says and
// an assertion about a date stays true next week.
var fixedNow = time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)

// casesDir is where the cases live. Decision record 0002 puts the runner's
// own fixtures at the root of the tree rather than beside the package, so the
// path climbs out of internal/check.
const casesDir = "../../testdata/cases"

// expectation is what a case says the walk should have produced.
type expectation struct {
	directories        int
	records            int
	experimentsPresent bool
	decisionsPresent   bool
	decisions          int
	refusals           []string
}

// loadCases reads every case directory. It fails rather than skipping: a
// harness that quietly runs no cases is a green suite that proves nothing,
// and that is the failure this function is written against.
func loadCases(t *testing.T) map[string]expectation {
	t.Helper()

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", casesDir, err)
	}

	cases := make(map[string]expectation)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cases[entry.Name()] = readExpectation(t, filepath.Join(casesDir, entry.Name()))
	}
	if len(cases) == 0 {
		t.Fatalf("no cases under %s", casesDir)
	}
	return cases
}

// readExpectation reads one case's expectation. Every field is required and a
// missing file is a failure, because a case that forgot to say what it
// expects would otherwise pass against whatever the runner happened to do.
func readExpectation(t *testing.T, dir string) expectation {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(dir, "expected"))
	if err != nil {
		t.Fatalf("case %s: %v", dir, err)
	}

	var exp expectation
	var seen []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("case %s: cannot read expectation line %q", dir, line)
		}
		key, value := fields[0], fields[1]
		seen = append(seen, key)
		switch key {
		case "directories":
			exp.directories = mustAtoi(t, dir, value)
		case "records":
			exp.records = mustAtoi(t, dir, value)
		case "experiments":
			switch value {
			case "present":
				exp.experimentsPresent = true
			case "absent":
				exp.experimentsPresent = false
			default:
				t.Fatalf("case %s: experiments is %q, want present or absent", dir, value)
			}
		case "decisions":
			// One key carrying both facts. A tree with no decisions
			// directory and one whose directory is empty both read zero
			// records, so a case saying zero would not separate them, and
			// absent is how a case says which of the two it is.
			if value == "absent" {
				exp.decisionsPresent = false
				exp.decisions = 0
				break
			}
			exp.decisionsPresent = true
			exp.decisions = mustAtoi(t, dir, value)
		default:
			t.Fatalf("case %s: unknown expectation %q", dir, key)
		}
	}
	sort.Strings(seen)
	if strings.Join(seen, ",") != "decisions,directories,experiments,records" {
		t.Fatalf("case %s: expectation names %v, want all four of directories, experiments, records and decisions", dir, seen)
	}

	refusals, err := os.ReadFile(filepath.Join(dir, "expected-refusals"))
	if err != nil {
		t.Fatalf("case %s: %v", dir, err)
	}
	for _, line := range strings.Split(string(refusals), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			exp.refusals = append(exp.refusals, line)
		}
	}

	return exp
}

func mustAtoi(t *testing.T, dir, value string) int {
	t.Helper()
	n, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("case %s: %v", dir, err)
	}
	return n
}

// diffRefusalSets compares two sets of refusals and returns one line per
// difference, empty when the sets are equal. Order does not matter and a
// repeated entry is one entry, because a verdict is a set: what a case
// declares is which refusals a tree produces, never how many times or in what
// order the runner happened to mention them.
//
// This is the comparison the harness rests on, so it is exercised directly
// rather than only through the cases, which today all expect nothing.
func diffRefusalSets(want, got []string) []string {
	wantSet := make(map[string]bool, len(want))
	for _, w := range want {
		wantSet[w] = true
	}
	gotSet := make(map[string]bool, len(got))
	for _, g := range got {
		gotSet[g] = true
	}

	var diffs []string
	for w := range wantSet {
		if !gotSet[w] {
			diffs = append(diffs, fmt.Sprintf("expected refusal not produced: %s", w))
		}
	}
	for g := range gotSet {
		if !wantSet[g] {
			diffs = append(diffs, fmt.Sprintf("refusal produced that no case expected: %s", g))
		}
	}
	sort.Strings(diffs)
	return diffs
}
