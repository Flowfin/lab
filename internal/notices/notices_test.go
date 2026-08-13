// The harness these rules are proved with.
//
// A case is a directory under testdata/notices/. It holds the module set a
// binary recorded, as a file rather than as a struct assembled at run time, and
// the module cache the texts are read out of, as directories in the repository
// rather than as whichever modules this machine has downloaded. It holds the
// whole document the render should have produced and the whole set of
// properties it should have refused.
//
// The layout of a case:
//
//	testdata/notices/<name>/build              the module set, one line per fact
//	testdata/notices/<name>/cache/             the module cache, may be absent
//	testdata/notices/<name>/expected           the document, byte for byte
//	testdata/notices/<name>/expected-refusals  one property per line, may be
//	                                           empty
//	testdata/notices/<name>/near-neighbour     the case that differs by the
//	                                           smallest legal change, required
//	                                           of a case that refuses
//
// It is the shape internal/contexts and the record checks are proved with rather
// than a fourth one, because a reader who has understood one harness here should
// not have to learn another.
package notices

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// casesDir is where the cases live. Record 0002 puts the runner's own fixtures
// at the root of the tree rather than beside the package, so the path climbs out
// of internal/notices.
const casesDir = "../../testdata/notices"

// A caseInput is one case as its files declare it.
type caseInput struct {
	name      string
	build     Build
	cache     Cache
	expected  string
	refusals  []string
	neighbour string
}

// TestEveryCaseIsRenderedAsItsFilesDeclare holds the render to both halves of
// every case: the properties it refused and the bytes it produced.
//
// The document is compared in full rather than by a substring. A test asserting
// that the text contains a module path passes on a document that lost the
// licence underneath it, and losing the text while keeping the name is the exact
// failure this package exists to prevent.
func TestEveryCaseIsRenderedAsItsFilesDeclare(t *testing.T) {
	cases := readCases(t)
	if len(cases) == 0 {
		t.Fatalf("no cases under %s, so this suite proved nothing", casesDir)
	}

	for _, input := range cases {
		t.Run(input.name, func(t *testing.T) {
			document := Render(input.build, input.cache)

			got := document.Properties()
			sort.Strings(got)
			want := append([]string(nil), input.refusals...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("refused %v, and the case declares %v", got, want)
				for _, refusal := range document.Refusals {
					t.Logf("  %s", refusal)
				}
			}

			if text := document.Text(); text != input.expected {
				t.Errorf("the document does not match the case, byte for byte")
				t.Logf("produced:\n%s", text)
				t.Logf("declared:\n%s", input.expected)
			}
		})
	}
	t.Logf("%d case(s) read from %s", len(cases), casesDir)
}

// TestACaseThatRefusesNamesANeighbourThatDoesNot is the near-miss discipline.
//
// A case that refuses proves only that something in it was refused. What proves
// the rule bites for the reason it names is the neighbour: the same case with
// the one thing repaired, refusing nothing. Without it a rule that refused every
// module it read would pass every case here.
func TestACaseThatRefusesNamesANeighbourThatDoesNot(t *testing.T) {
	cases := readCases(t)
	byName := make(map[string]caseInput, len(cases))
	for _, input := range cases {
		byName[input.name] = input
	}

	checked := 0
	for _, input := range cases {
		if len(input.refusals) == 0 {
			continue
		}
		if input.neighbour == "" {
			t.Errorf("%s refuses %v and names no near neighbour", input.name, input.refusals)
			continue
		}
		neighbour, ok := byName[input.neighbour]
		if !ok {
			t.Errorf("%s names the near neighbour %s, which is not a case", input.name, input.neighbour)
			continue
		}
		if properties := Render(neighbour.build, neighbour.cache).Properties(); len(properties) != 0 {
			t.Errorf("%s is the near neighbour of %s and refuses %v, so it proves nothing about why %s was refused",
				neighbour.name, input.name, properties, input.name)
			continue
		}
		checked++
	}
	t.Logf("%d refusing case(s) proved against a neighbour that passes", checked)
}

// TestADependencyAddedToTheBuildIsListed is the leg issue #37 names in its own
// words: a run against a tree with a dependency added produces a notices file
// that lists it.
//
// It is written as one render of two builds rather than as two cases compared by
// eye, because what it has to hold is the difference. A document that named the
// module whatever the build said would pass a case that only ever saw a build
// with a dependency in it.
//
// THE BOUND, and it is why the command's own suite exists as well. The build
// here is a module set constructed in this test, not one read out of a compiled
// binary, so what this proves is the render. That the module table inside a real
// binary reaches this render is what cmd/notices proves, against a binary it
// builds.
func TestADependencyAddedToTheBuildIsListed(t *testing.T) {
	cases := readCases(t)
	byName := make(map[string]caseInput, len(cases))
	for _, input := range cases {
		byName[input.name] = input
	}

	without, ok := byName["a-build-with-no-dependencies"]
	if !ok {
		t.Fatalf("the case a-build-with-no-dependencies is not in %s", casesDir)
	}
	with, ok := byName["a-build-with-one-dependency"]
	if !ok {
		t.Fatalf("the case a-build-with-one-dependency is not in %s", casesDir)
	}
	if len(with.build.Deps) != 1 {
		t.Fatalf("a-build-with-one-dependency carries %d dependencies", len(with.build.Deps))
	}
	added := with.build.Deps[0]

	before := Render(without.build, without.cache).Text()
	if strings.Contains(before, added.Path) {
		t.Fatalf("the build with no dependencies already names %s, so the comparison below proves nothing", added.Path)
	}

	build := without.build
	build.Deps = []Module{added}
	after := Render(build, with.cache)

	if properties := after.Properties(); len(properties) != 0 {
		t.Fatalf("adding %s refused %v", added.Describe(), properties)
	}
	if !strings.Contains(after.Text(), added.Describe()) {
		t.Errorf("the document does not name %s", added.Describe())
	}

	licence, err := with.cache.Licence(added)
	if err != nil {
		t.Fatalf("the case's own cache cannot supply %s: %v", added.Describe(), err)
	}
	if !strings.Contains(after.Text(), strings.TrimSpace(licence.Text)) {
		t.Errorf("the document names %s and does not carry its licence text, which is the half a link would also have failed to supply", added.Path)
	}
	t.Logf("adding %s moved the document from %d to %d byte(s) and carried its licence text", added.Describe(), len(before), len(after.Text()))
}

// TestTheDocumentIsAFunctionOfTheBuildAlone holds the render to producing one
// file from one build.
//
// The release milestone asks for two runs from one tag to produce identical
// checksums, and a document carrying the time it was made would defeat that
// wherever it is attached. A clock is the easy thing to add to a generated file
// and the hard thing to notice afterwards, because a document with yesterday's
// date in it looks correct.
func TestTheDocumentIsAFunctionOfTheBuildAlone(t *testing.T) {
	for _, input := range readCases(t) {
		first := Render(input.build, input.cache).Text()
		second := Render(input.build, input.cache).Text()
		if first != second {
			t.Errorf("%s renders two different documents from one build", input.name)
		}
	}
}

// TestTheModuleCacheSpellingSurvivesACapital holds the escaping to the case the
// module path of this repository is an instance of.
//
// It is a unit rather than a case because the property is about a string, and a
// case proving it would need a directory whose name differs from the module path
// by exactly the substitution under test, which is the thing a reader would have
// to check by eye.
func TestTheModuleCacheSpellingSurvivesACapital(t *testing.T) {
	for _, pair := range []struct{ in, want string }{
		{"github.com/Flowfin/lab", "github.com/!flowfin/lab"},
		{"github.com/flowfin/lab", "github.com/flowfin/lab"},
		{"ALLCAPS", "!a!l!l!c!a!p!s"},
		{"", ""},
	} {
		if got := EscapePath(pair.in); got != pair.want {
			t.Errorf("EscapePath(%q) is %q, and the cache spells it %q", pair.in, got, pair.want)
		}
	}
}

// readCases reads every case directory, and fails rather than skipping when one
// is malformed. A case the harness could not read is not a case that passed.
func readCases(t *testing.T) []caseInput {
	t.Helper()

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", casesDir, err)
	}

	var cases []caseInput
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(casesDir, entry.Name())

		build, err := readBuild(filepath.Join(dir, "build"))
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		expected, err := os.ReadFile(filepath.Join(dir, "expected"))
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		cases = append(cases, caseInput{
			name:      entry.Name(),
			build:     build,
			cache:     Cache{Root: filepath.Join(dir, "cache")},
			expected:  string(expected),
			refusals:  readLines(t, filepath.Join(dir, "expected-refusals")),
			neighbour: readOptional(t, filepath.Join(dir, "near-neighbour")),
		})
	}
	return cases
}

// readBuild reads a case's module set. One fact per line, so that a case is
// read as a file rather than as a format somebody has to decode:
//
//	main <path> <version>
//	revision <sha>
//	dep <path> <version>
//	dep <path> <version> replaces <path> <version>
func readBuild(path string) (Build, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return Build{}, err
	}

	var build Build
	for number, line := range strings.Split(string(text), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		switch {
		case fields[0] == "main" && len(fields) == 3:
			build.Main = Module{Path: fields[1], Version: fields[2]}
		case fields[0] == "revision" && len(fields) == 2:
			build.Revision = fields[1]
		case fields[0] == "dep" && len(fields) == 3:
			build.Deps = append(build.Deps, Module{Path: fields[1], Version: fields[2]})
		case fields[0] == "dep" && len(fields) == 6 && fields[3] == "replaces":
			build.Deps = append(build.Deps, Module{
				Path: fields[1], Version: fields[2],
				ReplacedPath: fields[4], ReplacedVersion: fields[5],
			})
		default:
			// The message names the line rather than only the file, because
			// a harness that says a fixture is wrong and not where sends a
			// reader to look through it.
			return Build{}, fmt.Errorf("%s: line %d is not a fact this harness reads: %s", path, number+1, line)
		}
	}
	return build, nil
}

// readLines reads a file of one entry per line. An absent file is an empty list
// rather than a failure, because most cases refuse nothing and a file holding
// nothing is noise in every one of them.
func readLines(t *testing.T, path string) []string {
	t.Helper()

	text, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}

	var lines []string
	for _, line := range strings.Split(string(text), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// readOptional reads a one-line file, or the empty string where there is none.
func readOptional(t *testing.T, path string) string {
	t.Helper()

	lines := readLines(t, path)
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}
