package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/Flowfin/lab/internal/check"
)

// realWalk is the walk the command uses when it is not being tested. A test
// that wants the ordinary behaviour passes this, so a case exercising a
// contrived walk is visible as one at the call site.
var realWalk = check.Walk

// fixedNow is the value every test here supplies for the clock. It is a date
// rather than whatever the machine says, so the waiting column a test asserts
// is the same number next week.
var fixedNow = time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)

// ordinary is the set of edges a test that is not contriving one of them
// passes. The walk is named at the call site because several cases contrive
// it; the listing and the clock are the real ones unless a case says otherwise.
func ordinary(walk func(string, time.Time) (check.Result, error)) edges {
	return edges{walk: walk, list: check.List, now: fixedNow, buildInfo: stamped}
}

// stamped is the build information a test that is not contriving one gets. It
// is a value written out here rather than the real stamp, because the real one
// is a different string on a checkout, at a tag and on a modified tree, and a
// test asserting against it would assert whatever the machine running the suite
// happened to produce.
func stamped() (*debug.BuildInfo, bool) {
	return &debug.BuildInfo{
		Main: debug.Module{Path: "github.com/Flowfin/lab", Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
			{Key: "vcs.time", Value: "2026-08-09T11:00:00Z"},
			{Key: "vcs.modified", Value: "false"},
		},
	}, true
}

// TestExitCodes reaches every code this runner can return, and names the
// record that decides what each one means. Decision record 0011 requires a
// test per code, so that a code existing only in the document is caught here
// rather than by an operator.
func TestExitCodes(t *testing.T) {
	refusing := func(string, time.Time) (check.Result, error) {
		return check.Result{
			Root:               "somewhere",
			ExperimentsPresent: true,
			Directories:        1,
			Records:            1,
			Refusals: []check.Refusal{{
				Property: "a-property",
				Subject:  "somewhere/experiments/one/EXPERIMENT.md",
				Detail:   "what was wrong with it",
			}},
		}, nil
	}
	failing := func(string, time.Time) (check.Result, error) {
		return check.Result{}, errors.New("cannot read the tree")
	}

	tests := []struct {
		name string
		args []string
		walk func(string, time.Time) (check.Result, error)
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
		{
			name: "a listing of a tree with experiments in it",
			args: []string{"list", "../../testdata/listings/several-experiments/tree"},
			walk: realWalk,
			want: exitClean,
		},
		{
			// The listing is a report and not a gate. This tree carries a
			// record the checker refuses, and the listing still leaves with
			// the clean code, because nothing here fails for what it found.
			name: "a listing of a tree holding a record the checker refuses",
			args: []string{"list", "../../testdata/cases/record-answered-with-no-answer/tree"},
			walk: realWalk,
			want: exitClean,
		},
		{
			name: "a listing of a tree that is not there",
			args: []string{"list", "no-such-directory"},
			walk: realWalk,
			want: exitCannot,
		},
		{
			name: "more paths than list takes",
			args: []string{"list", "one", "two"},
			walk: realWalk,
			want: exitCannot,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			got := run(tc.args, &out, &errOut, ordinary(tc.walk))
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
	refusing := func(string, time.Time) (check.Result, error) {
		return check.Result{
			Root:               "somewhere",
			ExperimentsPresent: true,
			Refusals: []check.Refusal{{
				Property: "a-property",
				Subject:  "somewhere/experiments/one/EXPERIMENT.md",
				Detail:   "what was wrong with it",
			}},
		}, nil
	}

	var out, errOut bytes.Buffer
	if got := run([]string{"check", "."}, &out, &errOut, ordinary(refusing)); got != exitRefused {
		t.Fatalf("exit code %d, want %d", got, exitRefused)
	}
	if !strings.Contains(out.String(), "somewhere/experiments/one/EXPERIMENT.md") {
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

// TestHelpNamesTheDocumentsAnOperatorIsOwed holds the last paragraph of the
// usage text to the three files it points at. An operator who reads that
// paragraph and then looks for one of the files is the reader this asserts for,
// and both halves of the walk they make are here: that the text names the file,
// and that the file is in the tree to be found.
//
// The second half is the one that earns its place. A pointer in the runner is
// outside the subject of the invariants paths leg, which reads the files at the
// root and everything under docs/ and holds those to the paths they name. So a
// document deleted or moved under this paragraph reddens nothing anywhere else,
// and a binary would go on telling operators to read a file that is not there.
//
// What it does not judge is whether any of the three says what it should. That
// is a reading of prose and no test here makes it.
func TestHelpNamesTheDocumentsAnOperatorIsOwed(t *testing.T) {
	var out, errOut bytes.Buffer
	if got := run([]string{"help"}, &out, &errOut, ordinary(realWalk)); got != exitClean {
		t.Fatalf("exit code %d, want %d", got, exitClean)
	}

	for _, document := range documentsAnOperatorIsOwed {
		if !strings.Contains(out.String(), document) {
			t.Errorf("the help output does not name %s:\n%s", document, out.String())
		}
		if _, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(document))); err != nil {
			t.Errorf("the help output names %s and it is not in this tree: %v", document, err)
		}
	}
}

// TestVersionNamesTheDocumentsAnOperatorIsOwed is the version half of the rule
// the help test holds for the usage text. Both routes exist because an operator
// reaches for one or the other and not reliably for both, so a paragraph that
// only one of them prints reaches only half of them.
//
// It asserts the same two things that test does. That the output names each
// document, and that each document it names is in this tree to be found - a
// binary pointing at a file somebody moved is worse than one pointing nowhere,
// because the reader follows it.
func TestVersionNamesTheDocumentsAnOperatorIsOwed(t *testing.T) {
	var out, errOut bytes.Buffer
	if got := run([]string{"version"}, &out, &errOut, ordinary(realWalk)); got != exitClean {
		t.Fatalf("exit code %d, want %d", got, exitClean)
	}

	for _, document := range documentsAnOperatorIsOwed {
		if !strings.Contains(out.String(), document) {
			t.Errorf("the version output does not name %s:\n%s", document, out.String())
		}
		if _, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(document))); err != nil {
			t.Errorf("the version output names %s and it is not in this tree: %v", document, err)
		}
	}
}

// TestVersionReportsWhatTheToolchainStamped asserts that the version the verb
// prints is the one it was handed rather than anything written in this package.
// The failure it prevents is a constant in the source drifting from the tag a
// binary was built at, which is the defect issue #44 opens with: it disagrees
// silently, so every report from that build is misleading rather than wrong in
// a way somebody notices.
func TestVersionReportsWhatTheToolchainStamped(t *testing.T) {
	e := ordinary(realWalk)
	e.buildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Path: "github.com/Flowfin/lab", Version: "v9.9.9"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "feedfacefeedfacefeedfacefeedfacefeedface"},
				{Key: "vcs.time", Value: "2026-01-02T03:04:05Z"},
			},
		}, true
	}

	var out, errOut bytes.Buffer
	if got := run([]string{"version"}, &out, &errOut, e); got != exitClean {
		t.Fatalf("exit code %d, want %d", got, exitClean)
	}

	for _, want := range []string{"v9.9.9", "feedfacefeedfacefeedfacefeedfacefeedface", "2026-01-02T03:04:05Z"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the version output does not carry %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "v1.2.3") {
		t.Errorf("the version output carries a version nothing handed it:\n%s", out.String())
	}
}

// TestVersionSaysWhenTheTreeWasModified holds the disclosure that decides what
// the two lines above it are worth. A build from a tree carrying changes
// version control did not hold is described by neither the tag nor the commit,
// and a reader who takes the commit for a description of the bytes in front of
// them is the person this line exists for.
//
// The negative leg is the half worth having. A clean build must not print it,
// because a disclosure that appears on every run is one nobody reads.
func TestVersionSaysWhenTheTreeWasModified(t *testing.T) {
	const disclosure = "carried changes version control did not hold"

	withModified := func(modified string) string {
		e := ordinary(realWalk)
		e.buildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{Path: "github.com/Flowfin/lab", Version: "v1.2.3"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
					{Key: "vcs.modified", Value: modified},
				},
			}, true
		}
		var out, errOut bytes.Buffer
		if got := run([]string{"version"}, &out, &errOut, e); got != exitClean {
			t.Fatalf("exit code %d, want %d", got, exitClean)
		}
		return out.String()
	}

	if got := withModified("true"); !strings.Contains(got, disclosure) {
		t.Errorf("a build from a modified tree does not disclose it:\n%s", got)
	}
	if got := withModified("false"); strings.Contains(got, disclosure) {
		t.Errorf("a build from a clean tree discloses a modification that did not happen:\n%s", got)
	}
}

// TestVersionWithoutBuildInformationCannot covers the one route where the verb
// has nothing to answer with. The toolchain stamps build information into every
// binary it builds in module mode, so a binary carrying none is not one this
// repository's build produced, and saying that is worth more than printing an
// empty version the reader would take for a real one.
//
// The code is the one record 0011 gives a runner that could not do its job,
// rather than the one a refusal returns.
func TestVersionWithoutBuildInformationCannot(t *testing.T) {
	e := ordinary(realWalk)
	e.buildInfo = func() (*debug.BuildInfo, bool) { return nil, false }

	var out, errOut bytes.Buffer
	if got := run([]string{"version"}, &out, &errOut, e); got != exitCannot {
		t.Fatalf("exit code %d, want %d", got, exitCannot)
	}
	if out.Len() != 0 {
		t.Errorf("it wrote to standard output when it had no version to report:\n%s", out.String())
	}
	if !strings.Contains(errOut.String(), "no build information") {
		t.Errorf("the message does not say what was missing:\n%s", errOut.String())
	}
}

// TestVersionTakesNoArguments holds the verb to the shape the other three have.
// A verb that quietly ignored a word after it would let "lab version check" run
// as a version report, and the reader would have no way to see that the thing
// they asked for did not happen.
func TestVersionTakesNoArguments(t *testing.T) {
	var out, errOut bytes.Buffer
	if got := run([]string{"version", "somewhere"}, &out, &errOut, ordinary(realWalk)); got != exitCannot {
		t.Fatalf("exit code %d, want %d", got, exitCannot)
	}
	if out.Len() != 0 {
		t.Errorf("it printed a version for an invocation it refused:\n%s", out.String())
	}
}

// TestTheUsageTextNamesTheVersionVerb keeps the help text and the verbs the
// command answers from drifting apart. A verb missing from the usage text is a
// verb only somebody reading the source finds, and the usage text is the only
// listing of them in this repository.
func TestTheUsageTextNamesTheVersionVerb(t *testing.T) {
	if !strings.Contains(usage, "lab version") {
		t.Errorf("the usage text does not name the version verb:\n%s", usage)
	}
}
