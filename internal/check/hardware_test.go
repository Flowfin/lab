package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flowfin/lab/internal/hardware"
)

// TestTheHarnessNamesAreTheHarnessesOwn holds the copy in this package equal to
// the harness it is about. The runner cannot import internal/hardware, which
// imports testing, so the tag is written twice, and a string written twice is a
// string that drifts. This is the only place the two meet, and the suite is
// where they may meet because a test binary is allowed to import a test harness.
//
// The second leg is the one that would actually catch a rename. The harness
// finds its own files by that suffix, so a file in it carrying the suffix is
// evidence that the marker this package looks for in an experiment directory is
// the marker the harness uses.
func TestTheHarnessNamesAreTheHarnessesOwn(t *testing.T) {
	if HarnessBuildTag != hardware.BuildTag {
		t.Errorf("this package reads %q as the harness constraint and the harness calls it %q", HarnessBuildTag, hardware.BuildTag)
	}

	entries, err := os.ReadDir("../hardware")
	if err != nil {
		t.Fatalf("cannot read the harness package: %v", err)
	}
	found := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), HarnessTestSuffix) {
			found++
		}
	}
	if found == 0 {
		t.Errorf("no file in the harness is named %s, so the marker this package looks for names nothing", "*"+HarnessTestSuffix)
	}
}

// TestWhatCountsAsRegisteredWithTheHarness holds the read the refusal rests on.
// The rule is a comparison between a declaration and this answer, so a wrong
// answer here refuses an honest record and clears a dishonest one, and neither
// failure is visible in the refusal set a case declares.
func TestWhatCountsAsRegisteredWithTheHarness(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{name: "a-record-that-needs-hardware-and-registers-a-test", count: 1},
		{name: "a-record-that-needs-nothing-and-registers-a-test", count: 1},
		{name: "a-record-that-needs-hardware-and-registers-no-test", count: 0},
		{name: "a-record-that-needs-nothing-and-registers-no-test", count: 0},
		// An ordinary experiment directory that holds a record and nothing
		// else registers nothing, which is what almost every one of them is.
		{name: "one-experiment-with-a-record", count: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree := filepath.Join(casesDir, tc.name, "tree")
			registered, err := harnessTestsUnder(os.DirFS(tree), "experiments/one", filepath.Join(tree, "experiments", "one"))
			if err != nil {
				t.Fatalf("reading the directory failed: %v", err)
			}
			if len(registered) != tc.count {
				t.Fatalf("read %v as registered with the harness, want %d file(s)", registered, tc.count)
			}
		})
	}
}

// TestADirectoryThatIsNotThereRegistersNothing is the boundary between this
// read and the refusal that already covers its case. A directory under
// experiments/ with no record at all is refused by ExperimentHasNoRecord, and a
// second refusal about its harness files would name a repair nobody needs.
func TestADirectoryThatIsNotThereRegistersNothing(t *testing.T) {
	tree := filepath.Join(casesDir, "no-experiments-directory", "tree")
	registered, err := harnessTestsUnder(os.DirFS(tree), "experiments/nothing-here", filepath.Join(tree, "experiments", "nothing-here"))
	if err != nil {
		t.Fatalf("reading a directory that is not there failed: %v", err)
	}
	if len(registered) != 0 {
		t.Fatalf("read %v as registered, want nothing", registered)
	}
}
