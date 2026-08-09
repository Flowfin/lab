package check

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
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
			for _, diff := range diffRefusalSets(want.refusals, got.Properties()) {
				t.Error(diff)
			}
		})
	}
}

// TestRefusalSetsAreComparedAsSets proves the comparison the harness rests on
// bites, and bites for the reason it names. The cases exercise it over the
// refusals that exist; this exercises it over the combinations they do not
// reach, including a fixture that trips its own refusal and one other.
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

// TestFixtureBytesSurviveTheCheckout reads every fixture that exists to carry
// a particular byte and fails if that byte is gone.
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
	tests := []struct {
		fixture string
		holds   string
		check   func([]byte) bool
	}{
		{
			fixture: "record-with-carriage-returns",
			holds:   "a carriage return",
			check:   func(b []byte) bool { return bytes.Contains(b, []byte("\r\n")) },
		},
		{
			fixture: "record-with-a-byte-order-mark",
			holds:   "a byte-order mark",
			check:   func(b []byte) bool { return bytes.HasPrefix(b, byteOrderMark) },
		},
		{
			fixture: "record-with-an-invalid-sequence",
			holds:   "an invalid encoding sequence",
			check:   func(b []byte) bool { return !utf8.Valid(b) },
		},
		{
			fixture: "record-with-a-null-byte",
			holds:   "a null byte",
			check:   func(b []byte) bool { return bytes.IndexByte(b, 0) >= 0 },
		},
	}

	for _, tc := range tests {
		t.Run(tc.fixture, func(t *testing.T) {
			path := filepath.Join(casesDir, tc.fixture, "tree", "experiments", "one", RecordName)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("cannot read %s: %v", path, err)
			}
			if !tc.check(data) {
				t.Fatalf("%s no longer carries %s: the bytes were translated on the way into this checkout, and the fixture now proves nothing", path, tc.holds)
			}
		})
	}
}

// TestARefusalNamesItsSubject holds the half a property identifier cannot
// carry. A guard that fails without saying what failed sends the reader to
// the source, and the reader is usually somebody who has just arrived.
func TestARefusalNamesItsSubject(t *testing.T) {
	for name := range loadCases(t) {
		result, err := Walk(filepath.Join(casesDir, name, "tree"))
		if err != nil {
			t.Fatalf("walk of %s failed: %v", name, err)
		}
		for _, refusal := range result.Refusals {
			if refusal.Subject == "" {
				t.Errorf("case %s: a %s refusal names no subject", name, refusal.Property)
				continue
			}
			if _, err := os.Stat(refusal.Subject); err != nil {
				t.Errorf("case %s: a %s refusal names %s, which cannot be opened: %v",
					name, refusal.Property, refusal.Subject, err)
			}
			if refusal.Detail == "" {
				t.Errorf("case %s: a %s refusal on %s says nothing about what was wrong",
					name, refusal.Property, refusal.Subject)
			}
			if !strings.Contains(refusal.String(), refusal.Subject) {
				t.Errorf("case %s: the message for a %s refusal does not name %s",
					name, refusal.Property, refusal.Subject)
			}
		}
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

// TestWhatCountsAsAPathInProse holds the half that decides whether the path
// refusal is useful or annoying. It has to read a path written in a sentence,
// which is where paths rot, and it has to leave alone the things that look
// like paths and are not.
func TestWhatCountsAsAPathInProse(t *testing.T) {
	tests := []struct {
		text string
		want []string
	}{
		{"the script in experiments/one/measure.sh produced it", []string{"experiments/one/measure.sh"}},
		{"see [the layout](docs/decisions/0002-repository-layout.md)", []string{"docs/decisions/0002-repository-layout.md"}},
		{"the walk is in `internal/check/check.go`", []string{"internal/check/check.go"}},
		{"the workflow is .github/workflows/dco.yml", []string{".github/workflows/dco.yml"}},
		{"relative, as ./docs/privacy.md", []string{"./docs/privacy.md"}},
		{"published at https://example.invalid/docs/thing.md", nil},
		{"either one and/or the other", nil},
		{"two thirds, or 2/3 of them", nil},
		{"a bare word, experiments, with no path after it", nil},
		{"somebody else's tree at /usr/share/doc", nil},
	}

	for _, tc := range tests {
		t.Run(tc.text, func(t *testing.T) {
			got := pathsNamedInProse(tc.text)
			if len(got) != len(tc.want) {
				t.Fatalf("read %v as paths, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("read %v as paths, want %v", got, tc.want)
				}
			}
		})
	}
}
