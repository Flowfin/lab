// The harness these rules are proved with.
//
// A case is a directory under testdata/bom/. It holds the module set a binary
// recorded, as a file rather than as a struct assembled at run time, the whole
// document the render should have produced, and the whole set of properties it
// should have refused.
//
// The layout of a case:
//
//	testdata/bom/<name>/build              the module set, one line per fact
//	testdata/bom/<name>/expected           the document, byte for byte
//	testdata/bom/<name>/expected-refusals  one property per line, may be absent
//	testdata/bom/<name>/near-neighbour     the case that differs by the
//	                                       smallest legal change, required of a
//	                                       case that refuses
//
// It is the shape internal/notices, internal/contexts and the record checks are
// proved with rather than a fourth one, because a reader who has understood one
// harness here should not have to learn another. The case file has one fact
// this repository's other harnesses have no spelling for: a dependency recorded
// with no version at all, written as a dep line carrying only a path.
package bom

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Flowfin/lab/internal/notices"
)

// casesDir is where the cases live. Record 0002 puts the runner's own fixtures
// at the root of the tree rather than beside the package, so the path climbs
// out of internal/bom.
const casesDir = "../../testdata/bom"

// A caseInput is one case as its files declare it.
type caseInput struct {
	name      string
	build     notices.Build
	expected  string
	refusals  []string
	neighbour string
}

// TestEveryCaseIsRenderedAsItsFilesDeclare holds the render to both halves of
// every case: the properties it refused and the bytes it produced.
//
// The document is compared in full rather than by a substring. A test asserting
// that the document contains a module path passes on one that lost the version
// beside it or wrote the wrong reference, and a bill of materials is read by a
// program, which will not notice either.
func TestEveryCaseIsRenderedAsItsFilesDeclare(t *testing.T) {
	cases := readCases(t)
	if len(cases) == 0 {
		t.Fatalf("no cases under %s, so this suite proved nothing", casesDir)
	}

	for _, input := range cases {
		t.Run(input.name, func(t *testing.T) {
			document := Render(input.build)

			got := document.Properties()
			want := append([]string(nil), input.refusals...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("refused %v, and the case declares %v", got, want)
				for _, refusal := range document.Refusals {
					t.Logf("  %s", refusal)
				}
			}

			text, err := document.JSON()
			if err != nil {
				t.Fatalf("the document could not be rendered: %v", err)
			}
			if text != input.expected {
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
// the one thing repaired, refusing nothing. Without it a rule that refused
// every build it read would pass every case here.
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
		if properties := Render(neighbour.build).Properties(); len(properties) != 0 {
			t.Errorf("%s is the near neighbour of %s and refuses %v, so it proves nothing about why %s was refused",
				neighbour.name, input.name, properties, input.name)
			continue
		}
		checked++
	}
	t.Logf("%d refusing case(s) proved against a neighbour that passes", checked)
}

// TestAPreReleaseVersionIsAReleaseVersion is the near-miss the version rule
// will actually meet.
//
// The refusal is written against a version derived from a commit, and the
// obvious way to write that rule is to refuse a version carrying a hyphen. An
// ordinary release candidate carries one too, so that rule would refuse the
// first tag anybody cuts under it, at the moment the release is being made and
// with the message saying the build names no release. The case list carries
// a release candidate for this reason and this test says so out loud.
func TestAPreReleaseVersionIsAReleaseVersion(t *testing.T) {
	for _, version := range []string{"v1.0.0", "v1.0.0-rc.1", "v0.1.0-alpha", "v2.3.4+incompatible"} {
		if !isReleaseVersion(version) {
			t.Errorf("%q is a version somebody tags a release with and this refuses it", version)
		}
	}
	for _, version := range []string{
		"",
		"(devel)",
		"v0.0.0-20260827021325-db0a15471e88",
		"v1.2.3-rc.1-20260827021325-db0a15471e88",
		// The one this repository actually produced. It is written out
		// rather than described, because the entry that was missing is the
		// suffix and a reader has to be able to see it.
		"v0.0.0-20260827021325-db0a15471e88+dirty",
	} {
		if isReleaseVersion(version) {
			t.Errorf("%q names no release a reader can resolve and this accepts it", version)
		}
	}
}

// TestTheDocumentIsAFunctionOfTheBuildAlone holds the render to producing one
// file from one build.
//
// The release milestone asks for two runs from one tag to produce identical
// checksums, and this document is published beside the artefact those checksums
// cover. CycloneDX offers both a timestamp and a serial number, both optional,
// and either of them would break that while looking like metadata: a document
// with yesterday's date in it looks correct.
func TestTheDocumentIsAFunctionOfTheBuildAlone(t *testing.T) {
	for _, input := range readCases(t) {
		first, err := Render(input.build).JSON()
		if err != nil {
			t.Fatalf("%s: %v", input.name, err)
		}
		second, err := Render(input.build).JSON()
		if err != nil {
			t.Fatalf("%s: %v", input.name, err)
		}
		if first != second {
			t.Errorf("%s renders two different documents from one build", input.name)
		}
	}
}

// TestADependencyAddedToTheBuildIsListed holds the render to the difference
// rather than to a document that happened to name a module.
//
// THE BOUND, and it is why the command's own suite exists as well. The build
// here is a module set constructed in this test, not one read out of a compiled
// binary, so what this proves is the render. That the module table inside a
// real binary reaches this render is what cmd/bom proves, against a binary it
// builds.
func TestADependencyAddedToTheBuildIsListed(t *testing.T) {
	cases := readCases(t)
	byName := make(map[string]caseInput, len(cases))
	for _, input := range cases {
		byName[input.name] = input
	}

	without, ok := byName["a-release-with-no-dependencies"]
	if !ok {
		t.Fatalf("the case a-release-with-no-dependencies is not in %s", casesDir)
	}
	with, ok := byName["a-release-with-one-dependency"]
	if !ok {
		t.Fatalf("the case a-release-with-one-dependency is not in %s", casesDir)
	}
	if len(with.build.Deps) != 1 {
		t.Fatalf("a-release-with-one-dependency carries %d dependencies", len(with.build.Deps))
	}
	added := with.build.Deps[0]

	before := Render(without.build)
	if len(before.Components) != 0 {
		t.Fatalf("the build with no dependencies already lists %d component(s), so the comparison below proves nothing", len(before.Components))
	}

	build := without.build
	build.Deps = []notices.Module{added}
	after := Render(build)

	if properties := after.Properties(); len(properties) != 0 {
		t.Fatalf("adding %s refused %v", added.Describe(), properties)
	}
	if len(after.Components) != 1 {
		t.Fatalf("adding one dependency produced %d component(s)", len(after.Components))
	}
	component := after.Components[0]
	if component.Name != added.Path || component.Version != added.Version {
		t.Errorf("the component is %s@%s and the module added was %s", component.Name, component.Version, added.Describe())
	}
	if component.PURL != PURL(added) {
		t.Errorf("the component reference is %q and the module resolves to %q", component.PURL, PURL(added))
	}
	t.Logf("adding %s moved the document from %d to %d component(s)", added.Describe(), len(before.Components), len(after.Components))
}

// TestThePackageURLSurvivesTheCharactersAModulePathCanHold is a unit rather
// than a case, because the property is about a string and a case proving it
// would put the answer in a file a reader has to compare by eye.
//
// The capital is the entry that matters here. The module path of this very
// repository carries one, so the spelling is not an edge somebody invented for
// a test, and the temptation to lower-case it is real: two module paths
// differing only in case are different modules, and a reference that folded
// them would name a module that was never in the binary.
func TestThePackageURLSurvivesTheCharactersAModulePathCanHold(t *testing.T) {
	for _, pair := range []struct {
		module notices.Module
		want   string
	}{
		{notices.Module{Path: "github.com/Flowfin/lab", Version: "v1.0.0"}, "pkg:golang/github.com/Flowfin/lab@v1.0.0"},
		{notices.Module{Path: "example.com/widget", Version: "v1.2.3"}, "pkg:golang/example.com/widget@v1.2.3"},
		{notices.Module{Path: "example.com/wid get", Version: "v1.0.0"}, "pkg:golang/example.com/wid%20get@v1.0.0"},
		{notices.Module{Path: "example.com/widget/v2", Version: "v2.0.0+incompatible"}, "pkg:golang/example.com/widget/v2@v2.0.0+incompatible"},
		{notices.Module{Path: "example.com/wid?get", Version: "v1.0.0"}, "pkg:golang/example.com/wid%3Fget@v1.0.0"},
	} {
		if got := PURL(pair.module); got != pair.want {
			t.Errorf("PURL(%s) is %q, and the reference is %q", pair.module.Describe(), got, pair.want)
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
//	modified true                           built from a tree carrying changes
//	dep <path> <version>
//	dep <path>                              recorded with no version at all
//	dep <path> <version> replaces <path> <version>
func readBuild(path string) (notices.Build, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return notices.Build{}, err
	}

	var build notices.Build
	for number, line := range strings.Split(string(text), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		switch {
		case fields[0] == "main" && len(fields) == 3:
			build.Main = notices.Module{Path: fields[1], Version: fields[2]}
		case fields[0] == "revision" && len(fields) == 2:
			build.Revision = fields[1]
		case fields[0] == "modified" && len(fields) == 2:
			build.Modified = fields[1] == "true"
		case fields[0] == "dep" && len(fields) == 2:
			build.Deps = append(build.Deps, notices.Module{Path: fields[1]})
		case fields[0] == "dep" && len(fields) == 3:
			build.Deps = append(build.Deps, notices.Module{Path: fields[1], Version: fields[2]})
		case fields[0] == "dep" && len(fields) == 6 && fields[3] == "replaces":
			build.Deps = append(build.Deps, notices.Module{
				Path: fields[1], Version: fields[2],
				ReplacedPath: fields[4], ReplacedVersion: fields[5],
			})
		default:
			// The message names the line rather than only the file, because
			// a harness that says a fixture is wrong and not where sends a
			// reader to look through it.
			return notices.Build{}, fmt.Errorf("%s: line %d is not a fact this harness reads: %s", path, number+1, line)
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
