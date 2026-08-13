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
//	testdata/cases/<name>/links               one path per line, each an entry
//	                                          the walk is shown as a symbolic
//	                                          link, absent from most cases
//
// This file decides that layout. Nothing restates it, so there is nothing to
// drift against it.
//
// THE LAST ONE IS A DEPARTURE FROM THE PARAGRAPH ABOVE AND IT IS DELIBERATE.
// Every other case is bytes in the repository and nothing else. A link is not,
// because a checkout cannot be relied on to carry one: creating a symbolic
// link on Windows needs SeCreateSymbolicLinkPrivilege, which an ordinary
// account does not hold, and os.Symlink there returns windows error 1314. The
// measurement is on issue #61. A link stored as a tracked entry arrives on such
// a checkout as an ordinary file holding the target as text, so a case built
// that way would ask the runner about a link on one platform and about a text
// file on another while declaring one answer for both. Record 0012 runs the
// suite on windows/amd64 and #57 keeps every platform running the same suite
// with nothing skipped, so neither the tracked link nor a link built at run
// time is available.
//
// What a declared link costs and what it buys. The bytes under tree/ are still
// exactly what the walk read, and the only thing the harness supplies is the
// type of one directory entry. What that leaves unproved is that a link made
// by an operating system arrives as an entry of that type, which is the
// standard library's behaviour rather than this runner's. What it proves is
// the whole of what this runner decides: what it does when it meets one.
//
// The pair to read together is a-symbolic-link-under-experiments and its near
// neighbour an-ordinary-file-where-a-link-would-be. Their trees are identical
// byte for byte, and the only difference between them is the line declaring
// the entry a link, so the refusal is shown to be about the link and not about
// the name, the place or the bytes.
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
	"io/fs"
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
//
// It is deliberately not today. A fixed value that happened to be the day the
// suite ran would pass identically whether the code used the value it was given
// or reached for the clock a second time, and the whole point of the parameter
// is that those two are different. Any second read moves every number below.
var fixedNow = time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)

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

// walkCase walks one case's tree and returns what the runner found. Every test
// that runs a case goes through it, so the decision about how a case is
// presented to the walk is made once.
func walkCase(t *testing.T, name string) (Result, error) {
	t.Helper()

	tree := filepath.Join(casesDir, name, "tree")
	if _, err := os.Stat(tree); err != nil {
		t.Fatalf("case %s has no tree: %v", name, err)
	}

	var fsys fs.FS = os.DirFS(tree)
	if links := linksDeclaredBy(t, name); len(links) > 0 {
		fsys = treeWithLinks{FS: fsys, links: links}
	}
	return walk(fsys, tree, fixedNow)
}

// linksDeclaredBy reads the entries a case declares as symbolic links. A case
// declaring none is the ordinary case and reads as an empty set rather than as
// a missing file it has to apologise for.
func linksDeclaredBy(t *testing.T, name string) map[string]bool {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(casesDir, name, "links"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("case %s: %v", name, err)
	}

	links := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			links[line] = true
		}
	}
	return links
}

// treeWithLinks is a case's tree with some of its entries reported as symbolic
// links. Everything else is read from the tree on disk, unchanged, so the case
// is still the files a reader can open.
type treeWithLinks struct {
	fs.FS
	links map[string]bool
}

// ReadDir is the one thing this overrides, because a directory listing is
// where the walk learns what an entry is.
func (t treeWithLinks) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(t.FS, name)
	if err != nil {
		return nil, err
	}
	for i, entry := range entries {
		inside := entry.Name()
		if name != "." {
			inside = name + "/" + inside
		}
		if t.links[inside] {
			entries[i] = declaredLink{DirEntry: entry}
		}
	}
	return entries, nil
}

// declaredLink is one directory entry the case declares to be a symbolic link.
// The name and the underlying file are the tree's own; the type is what this
// supplies.
type declaredLink struct {
	fs.DirEntry
}

func (d declaredLink) Type() fs.FileMode { return fs.ModeSymlink }

func (d declaredLink) IsDir() bool { return false }

func (d declaredLink) Info() (fs.FileInfo, error) {
	info, err := d.DirEntry.Info()
	if err != nil {
		return nil, err
	}
	return linkInfo{FileInfo: info}, nil
}

// linkInfo carries the same answers as the file on disk apart from the one bit
// that says what it is.
type linkInfo struct {
	fs.FileInfo
}

func (l linkInfo) Mode() fs.FileMode { return l.FileInfo.Mode()&^fs.ModeType | fs.ModeSymlink }

func (l linkInfo) IsDir() bool { return false }

// TestADeclaredLinkIsReportedAsOne holds the harness's own half of the link
// case. The refusal it feeds is proved by the case; that the harness really
// puts a link in front of the walk is proved here, because a declaration this
// dropped in silence would leave the case green against an ordinary file and
// the refusal unproved while the suite said otherwise.
func TestADeclaredLinkIsReportedAsOne(t *testing.T) {
	const name = "a-symbolic-link-under-experiments"

	links := linksDeclaredBy(t, name)
	if len(links) == 0 {
		t.Fatalf("case %s declares no link, so the case for the link refusal is not the case it claims to be", name)
	}

	tree := filepath.Join(casesDir, name, "tree")
	plain, err := fs.ReadDir(os.DirFS(tree), ExperimentsDir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", tree, err)
	}
	declared, err := fs.ReadDir(treeWithLinks{FS: os.DirFS(tree), links: links}, ExperimentsDir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", tree, err)
	}

	for i := range plain {
		inside := ExperimentsDir + "/" + plain[i].Name()
		wantLink := links[inside]
		if got := declared[i].Type()&fs.ModeSymlink != 0; got != wantLink {
			t.Errorf("%s is reported as a link %v, want %v", inside, got, wantLink)
		}
		if wantLink && plain[i].Type()&fs.ModeSymlink != 0 {
			t.Errorf("%s is already a link on disk, so this case proves nothing about a declaration", inside)
		}
		if declared[i].Name() != plain[i].Name() {
			t.Errorf("the declaration renamed %s to %s", plain[i].Name(), declared[i].Name())
		}
	}
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
