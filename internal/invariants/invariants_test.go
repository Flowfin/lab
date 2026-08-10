// The harness these legs are proved with.
//
// A case is a directory under testdata/invariants/. It holds the tree the scan
// read, as files in the repository rather than as strings assembled at run
// time, so a reader can look at exactly the input the scan saw. It holds what
// the scan should have found, and it holds the whole set of refusals it should
// have produced.
//
// The layout of a case:
//
//	testdata/invariants/<name>/tree/               what the scan reads
//	testdata/invariants/<name>/expected            one line, see readExpectation
//	testdata/invariants/<name>/expected-refusals   one property per line, may be
//	                                               empty
//	testdata/invariants/<name>/licence             the licence the case declares,
//	                                               absent where it declares none
//	testdata/invariants/<name>/near-neighbour      the case that differs by the
//	                                               smallest legal change,
//	                                               required of a case that
//	                                               refuses
//
// It is the shape the record checks are proved with rather than a second one,
// because a reader who has understood one harness here should not have to learn
// another. The one field it adds is the licence, which the record cases have no
// use for.
package invariants

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
// at the root of the tree rather than beside the package, so the path climbs
// out of internal/invariants.
const casesDir = "../../testdata/invariants"

// repositoryRoot is this repository, read from inside the package directory.
const repositoryRoot = "../.."

// expectation is what a case says the scan should have produced.
type expectation struct {
	textFiles int
	licence   string
	refusals  []string
}

// loadCases reads every case directory. It fails rather than skipping: a
// harness that quietly runs no cases is a green suite that proves nothing.
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

// readExpectation reads one case's expectation. The count is required and a
// missing file is a failure, because a case that forgot to say what it expects
// would otherwise pass against whatever the scan happened to do, including
// reading nothing at all.
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
		seen = append(seen, fields[0])
		switch fields[0] {
		case "text-files":
			n, err := strconv.Atoi(fields[1])
			if err != nil {
				t.Fatalf("case %s: %v", dir, err)
			}
			exp.textFiles = n
		default:
			t.Fatalf("case %s: unknown expectation %q", dir, fields[0])
		}
	}
	if strings.Join(seen, ",") != "text-files" {
		t.Fatalf("case %s: expectation names %v, want text-files", dir, seen)
	}

	exp.licence = readOptional(t, dir, "licence")

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

// readOptional reads a one-line file a case may or may not carry.
func readOptional(t *testing.T, dir, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(dir, name))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatalf("case %s: %v", dir, err)
	}
	return strings.TrimSpace(string(data))
}

// TestCases runs every case and compares what the scan found against what the
// case declares, including the whole refusal set. Comparing the whole set is
// what makes a case mean anything: a fixture that also trips a second refusal
// proves neither of them cleanly.
func TestCases(t *testing.T) {
	for name, want := range loadCases(t) {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(casesDir, name, "tree")
			got, err := Scan(root, want.licence)
			if err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			if got.TextFiles != want.textFiles {
				t.Errorf("read %d text files, case expects %d", got.TextFiles, want.textFiles)
			}
			for _, diff := range diffRefusalSets(want.refusals, got.Properties()) {
				t.Error(diff)
			}
		})
	}
}

// TestARefusalNamesItsSubject holds the half a property identifier cannot
// carry. A guard that fails without saying what failed sends the reader to the
// source, and the reader is usually somebody who has just arrived.
func TestARefusalNamesItsSubject(t *testing.T) {
	for name, want := range loadCases(t) {
		report, err := Scan(filepath.Join(casesDir, name, "tree"), want.licence)
		if err != nil {
			t.Fatalf("scan of %s failed: %v", name, err)
		}
		for _, refusal := range report.Refusals {
			if refusal.Subject == "" {
				t.Errorf("case %s: a %s refusal names no subject", name, refusal.Property)
				continue
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

// TestEveryRefusalHasACaseThatProvesItBites is the standing requirement, and it
// reaches refusals nobody has written yet, because the property list is read
// out of this package's source rather than out of a list somebody maintains.
//
// Two legs. A case whose whole refusal set is exactly that one property, and a
// near neighbour of that case differing by the smallest legal change and
// refusing nothing. Without the second, a leg that refuses everything passes
// its own test.
//
// THE BOUND. Both legs compare which properties were refused and never which
// line inside the package refused them, so two refusal sites producing one
// property are indistinguishable to them and an operator holding several of
// them can lose one and stay green. That is the state of this package rather
// than a hypothetical: the notice leg produces
// readme-does-not-link-the-intended-use-notice at two sites, one for a readme
// that is absent and one for a readme that never names the notice.
func TestEveryRefusalHasACaseThatProvesItBites(t *testing.T) {
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
			neighbour := readOptional(t, filepath.Join(casesDir, name), "near-neighbour")
			if neighbour == "" {
				t.Errorf("case %s proves %s and declares no near neighbour, so nothing here would notice a leg that refuses everything", name, property)
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

// TestANearNeighbourIsDeclaredOnlyWhereItMeansSomething refuses a near
// neighbour declared by a case that refuses nothing. Such a declaration reads
// as a proof and is not one, since a case that refuses nothing has no refusal
// for a neighbour to be near to.
func TestANearNeighbourIsDeclaredOnlyWhereItMeansSomething(t *testing.T) {
	for name, exp := range loadCases(t) {
		if len(exp.refusals) == 0 && readOptional(t, filepath.Join(casesDir, name), "near-neighbour") != "" {
			t.Errorf("case %s refuses nothing and declares a near neighbour", name)
		}
	}
}

// producibleProperties reads this package's own source and returns every
// property a refusal site names. It is derived rather than listed, because a
// list written by hand stops being true the moment somebody adds a refusal site
// and forgets it, and the requirement above exists to reach exactly that case.
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
		value, ok := constants[ident.Name]
		if !ok {
			t.Fatalf("a refusal site at %s names %s and this cannot resolve it to a property, so nothing here can require a case for it",
				fset.Position(site.Pos()), ident.Name)
		}
		if !seen[value] {
			seen[value] = true
			props = append(props, value)
		}
	}

	sort.Strings(props)
	return props
}

// TestTheVocabularyDoesNotRefuseItself holds the property the vocabulary file
// is written against. Every pattern there opens with a character class so that
// its own source text does not match it, and the whole reason that matters is
// that this package scans tracked text and the vocabulary is tracked text.
//
// Without this test the repair for a pattern written as a plain literal is an
// exclusion for the file holding the vocabulary, and a scanner that skips the
// file defining its own rules has a hole nobody can see from outside.
func TestTheVocabularyDoesNotRefuseItself(t *testing.T) {
	const vocabulary = "vocabulary.go"

	data, err := os.ReadFile(vocabulary)
	if err != nil {
		t.Fatalf("cannot read %s: %v", vocabulary, err)
	}
	if len(credentialShapes) == 0 || len(attributionMarkers) == 0 {
		t.Fatal("a vocabulary is empty, so this test would pass over nothing")
	}

	file := textFile{path: vocabulary, relative: vocabulary, text: string(data)}
	for _, tc := range []struct {
		name string
		leg  func([]textFile) (Leg, []Refusal)
	}{
		{"credential shapes", credentialLeg},
		{"attribution markers", attributionLeg},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, refusals := tc.leg([]textFile{file})
			for _, refusal := range refusals {
				t.Errorf("%s refuses the file that defines it: %s", tc.name, refusal)
			}
		})
	}
}

// TestThisRepositorySatisfiesTheInvariants is the run the workflow makes. It is
// a test rather than a verb of the runner because these are properties of this
// repository rather than of a tree the runner is pointed at, and a verb that
// only ever has one correct argument is a verb that will be pointed at the
// wrong one.
//
// It prints the report whatever the outcome, so a green run says what it
// examined instead of saying nothing.
func TestThisRepositorySatisfiesTheInvariants(t *testing.T) {
	report, err := Scan(repositoryRoot, DeclaredLicence)
	if err != nil {
		t.Fatalf("cannot scan this repository: %v", err)
	}
	t.Log("\n" + report.String())

	if report.TextFiles == 0 {
		t.Fatal("the scan read no text files in this repository, which cannot be right")
	}
	for _, refusal := range report.Refusals {
		t.Errorf("%s", refusal)
	}
}

// TestTheLicenceLegSaysWhetherItWasAsked holds the sentence the licence
// question owes a reader. While no licence is declared the leg has to report
// that it was not asked and what asking would cost, so that a run which could
// not check the licence is never read as one that checked and was satisfied.
// Once the question is answered the same test requires the opposite, so
// answering it does not leave a stale assertion behind.
func TestTheLicenceLegSaysWhetherItWasAsked(t *testing.T) {
	report, err := Scan(repositoryRoot, DeclaredLicence)
	if err != nil {
		t.Fatalf("cannot scan this repository: %v", err)
	}

	var licence *Leg
	for i := range report.Legs {
		if report.Legs[i].Name == "the licence" {
			licence = &report.Legs[i]
		}
	}
	if licence == nil {
		t.Fatal("the report carries no licence leg at all")
	}

	if DeclaredLicence == "" {
		if licence.Asked {
			t.Fatal("no licence is declared and the leg reports that it was asked")
		}
		if !strings.Contains(licence.NotAsked, "#47") {
			t.Errorf("the licence leg says it was not asked and does not say what asking would cost: %q", licence.NotAsked)
		}
		if !strings.Contains(report.String(), "NOT ASKED") {
			t.Error("the report does not say that a leg was not asked")
		}
		return
	}
	if !licence.Asked {
		t.Fatalf("%s is declared and the licence leg reports that it was not asked", DeclaredLicence)
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
