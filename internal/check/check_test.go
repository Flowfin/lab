package check

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCases runs every case under testdata/cases and compares what the walk
// found against what the case declares, including the whole refusal set.
//
// The smallest case is the one that matters most today: a tree with nothing
// in it, walked, reporting zero. That is the smallest true statement this
// runner can make, and it is worth having because a walk that was broken
// would print the same thing.
func TestCases(t *testing.T) {
	for name, want := range loadCases(t) {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(casesDir, name, "tree")
			if _, err := os.Stat(root); err != nil {
				t.Fatalf("case %s has no tree: %v", name, err)
			}

			got, err := Walk(root)
			if err != nil {
				t.Fatalf("walk failed: %v", err)
			}

			if got.Directories != want.directories {
				t.Errorf("walked %d experiment directories, case expects %d", got.Directories, want.directories)
			}
			if got.Records != want.records {
				t.Errorf("read %d records, case expects %d", got.Records, want.records)
			}
			if got.ExperimentsPresent != want.experimentsPresent {
				t.Errorf("experiments directory present is %v, case expects %v", got.ExperimentsPresent, want.experimentsPresent)
			}
			for _, diff := range diffRefusalSets(want.refusals, got.Refusals) {
				t.Error(diff)
			}
		})
	}
}

// TestRefusalSetsAreComparedAsSets proves the comparison the harness rests on
// bites, and bites for the reason it names. Every case in the tree today
// expects no refusals, so nothing else in this suite could tell a working
// comparison from one that always agrees.
func TestRefusalSetsAreComparedAsSets(t *testing.T) {
	tests := []struct {
		name      string
		want, got []string
		diffs     int
	}{
		{
			name:  "nothing expected and nothing produced",
			diffs: 0,
		},
		{
			// The case #14 names: a fixture that trips the refusal it was
			// written for and one other. Asking only whether the intended
			// refusal is present would pass this.
			name:  "the expected refusal and one more",
			want:  []string{"a"},
			got:   []string{"a", "b"},
			diffs: 1,
		},
		{
			name:  "the expected refusal is missing",
			want:  []string{"a"},
			got:   nil,
			diffs: 1,
		},
		{
			name:  "a refusal nobody expected, against an empty expectation",
			want:  nil,
			got:   []string{"a"},
			diffs: 1,
		},
		{
			name:  "the same set in another order",
			want:  []string{"a", "b"},
			got:   []string{"b", "a"},
			diffs: 0,
		},
		{
			name:  "one expected, a different one produced",
			want:  []string{"a"},
			got:   []string{"b"},
			diffs: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diffs := diffRefusalSets(tc.want, tc.got)
			if len(diffs) != tc.diffs {
				t.Fatalf("got %d differences %v, want %d", len(diffs), diffs, tc.diffs)
			}
		})
	}
}

// TestEveryCaseIsRunByTestCases refuses a case directory that the harness
// would walk past. A fixture nobody runs is the quietest way a suite loses
// coverage while staying green, and the failure it produces is a green board
// over a tree the runner never read.
func TestEveryCaseIsRunByTestCases(t *testing.T) {
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", casesDir, err)
	}

	loaded := loadCases(t)
	for _, entry := range entries {
		if !entry.IsDir() {
			t.Errorf("%s holds %s, which is not a case directory", casesDir, entry.Name())
			continue
		}
		if _, ok := loaded[entry.Name()]; !ok {
			t.Errorf("case %s is in the tree and is not loaded", entry.Name())
		}
	}
}

// TestFixtureBytesSurviveTheCheckout reads the fixture that exists to carry a
// carriage return and fails if the carriage return is gone.
//
// The bytes reaching the runner have to be exact. This repository's
// .gitattributes stores every tracked text file with LF, which would delete
// the byte this fixture exists to preserve, and testdata is excluded from
// that translation for exactly this reason. Nothing tells a reader that the
// exclusion is still in place, so this test is what tells them.
//
// What it can prove is bounded and the bound is worth stating. Removing the
// exclusion does not change a blob that is already stored, so this test goes
// red when the fixture is next written through a checkout that translates,
// not on the commit that removes the rule.
func TestFixtureBytesSurviveTheCheckout(t *testing.T) {
	const fixture = casesDir + "/record-with-carriage-returns/tree/experiments/one/EXPERIMENT.md"

	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatalf("cannot read %s: %v", fixture, err)
	}
	if !strings.Contains(string(data), "\r\n") {
		t.Fatalf("%s carries no carriage return: the bytes were translated on the way into this checkout", fixture)
	}
}

// TestWalkWritesNothing proves the property the runner was built around. A
// checker that edits what it judges cannot be trusted about what it found, so
// the claim is worth more than the grep that would otherwise stand for it.
//
// The whole fixture tree is fingerprinted before and after a walk of every
// case, by path and by content, so a file created, changed or removed
// anywhere under it is a difference this test reports.
func TestWalkWritesNothing(t *testing.T) {
	before := fingerprint(t, casesDir)

	for name := range loadCases(t) {
		if _, err := Walk(filepath.Join(casesDir, name, "tree")); err != nil {
			t.Fatalf("walk of %s failed: %v", name, err)
		}
	}

	after := fingerprint(t, casesDir)

	for path, sum := range before {
		switch other, ok := after[path]; {
		case !ok:
			t.Errorf("%s was removed by a walk", path)
		case other != sum:
			t.Errorf("%s was changed by a walk", path)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			t.Errorf("%s was created by a walk", path)
		}
	}
}

// fingerprint maps every path under root to a hash of its contents.
// Directories are recorded too, so a walk that creates an empty one is a
// difference rather than something this test cannot see.
func fingerprint(t *testing.T, root string) map[string]string {
	t.Helper()

	sums := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			sums[filepath.ToSlash(path)] = "directory"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sums[filepath.ToSlash(path)] = fmt.Sprintf("%x", sha256.Sum256(data))
		return nil
	})
	if err != nil {
		t.Fatalf("cannot fingerprint %s: %v", root, err)
	}
	return sums
}
