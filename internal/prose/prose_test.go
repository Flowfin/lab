// The harness every rule in this package is proved with.
//
// A case is a directory under testdata/prose/. It holds the tree the scan read,
// as files in the repository rather than as strings assembled at run time, so a
// reader can look at exactly the bytes the scan saw. That matters more here than
// anywhere else in this repository: every rule in this package is about a byte
// nobody can see, and a fixture written as a string in a test file is a fixture
// whose trailing space the next reformat deletes.
//
//	testdata/prose/<name>/tree/               what the scan reads
//	testdata/prose/<name>/expected            one line, the count of files read
//	testdata/prose/<name>/expected-refusals   one property per line, may be empty
//	testdata/prose/<name>/near-neighbour      the case that differs by the
//	                                          smallest legal change, required of
//	                                          a case that refuses
//
// This file decides that layout. Nothing restates it, so there is nothing to
// drift against it.
package prose

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// casesDir is where the cases live. Record 0002 puts the runner's own fixtures
// at the root of the tree rather than beside the package, so the path climbs out
// of internal/prose.
const casesDir = "../../testdata/prose"

// expectation is what a case says the scan should have produced.
type expectation struct {
	files    int
	refusals []string
}

// loadCases reads every case directory. It fails rather than skipping: a harness
// that quietly runs no cases is a green suite that proves nothing.
func loadCases(t *testing.T) map[string]expectation {
	t.Helper()

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", casesDir, err)
	}

	cases := make(map[string]expectation)
	for _, entry := range entries {
		if !entry.IsDir() {
			t.Errorf("%s holds %s, which is not a case directory", casesDir, entry.Name())
			continue
		}
		cases[entry.Name()] = readExpectation(t, filepath.Join(casesDir, entry.Name()))
	}
	if len(cases) == 0 {
		t.Fatalf("no cases under %s", casesDir)
	}
	return cases
}

// readExpectation reads one case's expectation. A missing file is a failure,
// because a case that forgot to say what it expects would otherwise pass against
// whatever the scan happened to do.
func readExpectation(t *testing.T, dir string) expectation {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(dir, "expected"))
	if err != nil {
		t.Fatalf("case %s: %v", dir, err)
	}
	fields := strings.Fields(strings.TrimSpace(string(data)))
	if len(fields) != 2 || fields[0] != "files" {
		t.Fatalf("case %s: the expectation is %q, want one line reading files <n>", dir, string(data))
	}
	count, err := strconv.Atoi(fields[1])
	if err != nil {
		t.Fatalf("case %s: %v", dir, err)
	}

	exp := expectation{files: count}
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

// TestCases runs every case and compares what the scan found against what the
// case declares, including the whole refusal set. Comparing the whole set is
// what makes a case mean anything: a fixture that trips the rule it was written
// for and one other proves neither of them cleanly.
func TestCases(t *testing.T) {
	for name, want := range loadCases(t) {
		t.Run(name, func(t *testing.T) {
			got, err := Scan(filepath.Join(casesDir, name, "tree"))
			if err != nil {
				t.Fatalf("the scan failed: %v", err)
			}
			if got.Files != want.files {
				t.Errorf("read %d prose files, case expects %d", got.Files, want.files)
			}
			for _, diff := range diffRefusalSets(want.refusals, got.Properties()) {
				t.Error(diff)
			}
		})
	}
}

// TestEveryRefusalHasAFixtureThatProvesItBites is the standing requirement, and
// it reaches a rule nobody has written yet because the property list is read out
// of this package's source rather than out of a list somebody maintains.
//
// Three legs. A fixture whose whole refusal set is exactly that one property; a
// near neighbour of it differing by the smallest legal change and refusing
// nothing, without which a check that refuses everything passes its own test;
// and the refusal naming its subject, which TestARefusalNamesItsSubject holds.
//
// THE BOUND. Every leg compares which properties were refused and never which
// line inside the package refused them, so two sites producing one property are
// indistinguishable to all three. That is the state of this package rather than
// a hypothetical: the final-newline rule is produced at two sites, one for a
// missing newline and one for a blank line at the end. A case exists for each
// arm, and that is a fixture somebody remembered rather than something this test
// can require.
func TestEveryRefusalHasAFixtureThatProvesItBites(t *testing.T) {
	cases := loadCases(t)

	provers := make(map[string][]string)
	for name, exp := range cases {
		if len(exp.refusals) == 1 {
			provers[exp.refusals[0]] = append(provers[exp.refusals[0]], name)
		}
	}

	for _, property := range producibleProperties(t) {
		names := provers[property]
		if len(names) == 0 {
			t.Errorf("%s can be refused and no case declares exactly that refusal and no other", property)
			continue
		}
		sort.Strings(names)

		for _, name := range names {
			neighbour := readNearNeighbour(t, name)
			if neighbour == "" {
				t.Errorf("case %s proves %s and declares no near neighbour, so nothing here would notice a check that refuses everything", name, property)
				continue
			}
			exp, ok := cases[neighbour]
			if !ok {
				t.Errorf("case %s names %s as its near neighbour and there is no such case", name, neighbour)
				continue
			}
			if len(exp.refusals) != 0 {
				t.Errorf("case %s names %s as its near neighbour, and that case expects %v rather than nothing", name, neighbour, exp.refusals)
			}
		}
	}
}

// TestARefusalNamesItsSubject holds the half a property identifier cannot carry.
// A guard that fails without saying which file failed sends the reader to the
// source, and the byte these rules are about is invisible, so the file and the
// line are the whole of what makes a refusal actionable.
func TestARefusalNamesItsSubject(t *testing.T) {
	for name := range loadCases(t) {
		report, err := Scan(filepath.Join(casesDir, name, "tree"))
		if err != nil {
			t.Fatalf("the scan of %s failed: %v", name, err)
		}
		for _, refusal := range report.Refusals {
			if refusal.Subject == "" {
				t.Errorf("case %s: a %s refusal names no subject", name, refusal.Property)
				continue
			}
			if _, err := os.Stat(refusal.Subject); err != nil {
				t.Errorf("case %s: a %s refusal names %s, which cannot be opened: %v", name, refusal.Property, refusal.Subject, err)
			}
			if refusal.Detail == "" {
				t.Errorf("case %s: a %s refusal on %s says nothing about what was wrong", name, refusal.Property, refusal.Subject)
			}
			if !strings.Contains(refusal.String(), refusal.Subject) {
				t.Errorf("case %s: the message for a %s refusal does not name %s", name, refusal.Property, refusal.Subject)
			}
		}
	}
}

// TestTheFixtureBytesSurviveTheCheckout reads the fixtures that exist to carry a
// particular byte and fails if that byte is gone. Every rule here is about a
// byte a checkout can translate away, so a fixture that lost it is a case that
// passes while proving nothing.
//
// What it can prove is bounded. Removing testdata's exclusion from
// .gitattributes does not change a blob that is already stored, so this goes red
// when a fixture is next written through a translating checkout rather than on
// the commit that removes the rule.
func TestTheFixtureBytesSurviveTheCheckout(t *testing.T) {
	tests := []struct {
		fixture string
		holds   string
		check   func(string) bool
	}{
		{
			fixture: "prose-with-trailing-whitespace",
			holds:   "a trailing space",
			check:   func(s string) bool { return strings.Contains(s, " \n") },
		},
		{
			fixture: "prose-with-a-tab",
			holds:   "a tab",
			check:   func(s string) bool { return strings.Contains(s, "\t") },
		},
		{
			fixture: "prose-with-no-final-newline",
			holds:   "no newline at the end",
			check:   func(s string) bool { return !strings.HasSuffix(s, "\n") },
		},
		{
			fixture: "prose-with-a-blank-line-at-the-end",
			holds:   "a blank line at the end",
			check:   func(s string) bool { return strings.HasSuffix(s, "\n\n") },
		},
	}

	for _, tc := range tests {
		t.Run(tc.fixture, func(t *testing.T) {
			path := filepath.Join(casesDir, tc.fixture, "tree", "docs", "one.md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("cannot read %s: %v", path, err)
			}
			if !tc.check(string(data)) {
				t.Fatalf("%s no longer carries %s: the bytes were translated on the way into this checkout, and the fixture now proves nothing", path, tc.holds)
			}
		})
	}
}

// TestTheFixturesAreNotRead is the exclusion this package's comment argues for,
// held rather than trusted. Every case tree under testdata/ carries prose that
// departs from these rules on purpose, so a scan of this repository that read
// that directory would refuse the evidence that the rules work. It is one path
// entry away and it leaves no trace, which is why it is a test rather than a
// sentence.
func TestTheFixturesAreNotRead(t *testing.T) {
	report, err := Scan("../..")
	if err != nil {
		t.Fatalf("the scan failed: %v", err)
	}
	for _, refusal := range report.Refusals {
		if strings.Contains(filepath.ToSlash(refusal.Subject), "/"+FixturesDir+"/") {
			t.Errorf("the scan reached %s, which is a fixture", refusal.Subject)
		}
	}
}

// TestThisRepositorySatisfiesTheProseFormat is the leg the workflow runs. It is
// the only test here that reads this repository rather than a fixture, and it is
// what a red tick on the gate means.
func TestThisRepositorySatisfiesTheProseFormat(t *testing.T) {
	report, err := Scan("../..")
	if err != nil {
		t.Fatalf("the scan failed: %v", err)
	}
	t.Log("\n" + report.String())

	if report.Files == 0 {
		t.Fatal("the scan read no prose at all, which cannot be right in this repository, so a green result here would say nothing")
	}
	for _, refusal := range report.Refusals {
		t.Error(refusal)
	}
}

// diffRefusalSets compares two sets of refusals and returns one line per
// difference, empty when the sets are equal. Order does not matter and a
// repeated entry is one entry, because a verdict is a set.
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

// readNearNeighbour returns the case a refusing case declares as its near
// neighbour, or the empty string when it declares none.
func readNearNeighbour(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(casesDir, name, "near-neighbour"))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatalf("case %s: %v", name, err)
	}
	return strings.TrimSpace(string(data))
}

// producibleProperties reads this package's own source and returns every
// property a refusal site names. It is derived rather than listed, because a
// list written by hand stops being true the moment somebody adds a site and
// forgets it, and the whole point of the requirement above is that it reaches
// rules nobody has written yet.
func producibleProperties(t *testing.T) []string {
	t.Helper()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("cannot parse this package: %v", err)
	}

	constants := make(map[string]string)
	var sites []ast.Expr

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.ValueSpec:
					for i, name := range node.Names {
						if i >= len(node.Values) {
							continue
						}
						if lit, ok := node.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							if value, err := strconv.Unquote(lit.Value); err == nil {
								constants[name.Name] = value
							}
						}
					}
				case *ast.KeyValueExpr:
					if key, ok := node.Key.(*ast.Ident); ok && key.Name == "Property" {
						sites = append(sites, node.Value)
					}
				}
				return true
			})
		}
	}

	if len(sites) == 0 {
		t.Fatal("no refusal site in this package names a property, which cannot be right while the package refuses anything")
	}

	seen := make(map[string]bool)
	var props []string
	for _, site := range sites {
		ident, ok := site.(*ast.Ident)
		if !ok {
			t.Fatalf("a refusal site at %s names a property this cannot resolve", fset.Position(site.Pos()))
		}
		value, known := constants[ident.Name]
		if !known {
			t.Fatalf("a refusal site names %s and this cannot resolve it to a property", ident.Name)
		}
		if !seen[value] {
			seen[value] = true
			props = append(props, value)
		}
	}
	sort.Strings(props)
	return props
}
