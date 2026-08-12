// The harness these rules are proved with.
//
// A case is a directory under testdata/contexts/. It holds the workflow files
// the reader read, as files in the repository rather than as strings assembled
// at run time, so a reader can look at exactly the input the comparison saw. It
// holds the two lists the comparison is between, and it holds the whole set of
// refusals it should have produced.
//
// The layout of a case:
//
//	testdata/contexts/<name>/workflows/          what the reader reads
//	testdata/contexts/<name>/required            one context per line, may be
//	                                             empty
//	testdata/contexts/<name>/absences            one deliberately absent check
//	                                             name per line, may be empty
//	testdata/contexts/<name>/expected-refusals   one property per line, may be
//	                                             empty
//	testdata/contexts/<name>/near-neighbour      the case that differs by the
//	                                             smallest legal change, required
//	                                             of a case that refuses
//
// It is the shape the record checks and the invariants are proved with rather
// than a third one, because a reader who has understood one harness here should
// not have to learn another.
package contexts

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// casesDir is where the cases live. Record 0002 puts the runner's own fixtures
// at the root of the tree rather than beside the package, so the path climbs out
// of internal/contexts.
const casesDir = "../../testdata/contexts"

// workflowsInThisRepository is this repository's own workflow directory, read
// from inside the package directory.
const workflowsInThisRepository = "../../" + WorkflowsDir

// expectation is what a case says the comparison should have produced.
type expectation struct {
	required  []string
	absences  []Absence
	refusals  []string
	neighbour string
}

// loadCases reads every case directory. It fails rather than skipping: a harness
// that quietly runs no cases is a green suite that proves nothing.
func loadCases(t *testing.T) map[string]expectation {
	t.Helper()

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("cannot read %s: %v", casesDir, err)
	}
	cases := map[string]expectation{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(casesDir, entry.Name())
		exp := expectation{
			required:  lines(t, dir, "required"),
			refusals:  lines(t, dir, "expected-refusals"),
			neighbour: readOptional(t, dir, "near-neighbour"),
		}
		for _, name := range lines(t, dir, "absences") {
			exp.absences = append(exp.absences, Absence{
				Name:  name,
				Why:   "this case says so",
				Until: "",
			})
		}
		sort.Strings(exp.refusals)
		cases[entry.Name()] = exp
	}
	if len(cases) == 0 {
		t.Fatalf("no cases under %s", casesDir)
	}
	return cases
}

func lines(t *testing.T, dir, name string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("cannot read %s: %v", filepath.Join(dir, name), err)
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func readOptional(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("cannot read %s: %v", filepath.Join(dir, name), err)
	}
	return strings.TrimSpace(string(data))
}

// TestACaseRefusesExactlyWhatItDeclares is the leg that makes a fixture worth
// having. It compares the whole set of property ids, so a rule that refuses its
// own case and something else besides fails here rather than passing on the
// strength of the one line it was written for.
func TestACaseRefusesExactlyWhatItDeclares(t *testing.T) {
	for name, exp := range loadCases(t) {
		t.Run(name, func(t *testing.T) {
			declared, err := ReadWorkflows(filepath.Join(casesDir, name, "workflows"))
			if err != nil {
				t.Fatalf("reading the workflows of %s: %v", name, err)
			}
			verdict := Judge(declared, exp.required, exp.absences)
			got := verdict.Properties()
			if got == nil {
				got = []string{}
			}
			want := exp.refusals
			if want == nil {
				want = []string{}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("case %s refused %v and declares %v\n%s", name, got, want, verdict.Report(declared, exp.required, exp.absences))
			}
		})
	}
}

// TestEveryPropertyHasACaseThatTripsIt refuses a property nothing proves. A
// property with no case behind it is a rule nobody has watched bite, and the
// whole reason the fixtures exist is that a rule which cannot be shown to refuse
// is indistinguishable from one that never runs.
func TestEveryPropertyHasACaseThatTripsIt(t *testing.T) {
	all := []string{
		RequiredContextNothingReports,
		DeclaredNameOutsideTheRequiredSet,
		AbsenceNamesNothingDeclared,
		AbsenceIsRequired,
		NameCarriesAnExpression,
	}
	tripped := map[string]bool{}
	for _, exp := range loadCases(t) {
		for _, property := range exp.refusals {
			tripped[property] = true
		}
	}
	for _, property := range all {
		if !tripped[property] {
			t.Errorf("no case trips %s", property)
		}
	}
	for property := range tripped {
		found := false
		for _, known := range all {
			if property == known {
				found = true
			}
		}
		if !found {
			t.Errorf("a case declares %s and this package has no such property, so the case proves nothing", property)
		}
	}
}

// TestACaseThatRefusesNamesANearNeighbour is the leg that catches a rule
// refusing everything. A case proving a refusal names a case differing by the
// smallest legal change, and that case has to come out clean, so a rule that
// fires on anything at all fails here.
func TestACaseThatRefusesNamesANearNeighbour(t *testing.T) {
	cases := loadCases(t)
	for name, exp := range cases {
		if len(exp.refusals) == 0 {
			if exp.neighbour != "" {
				t.Errorf("case %s refuses nothing and declares a near neighbour", name)
			}
			continue
		}
		if exp.neighbour == "" {
			t.Errorf("case %s refuses %v and declares no near neighbour, so nothing here would notice a rule that refuses everything", name, exp.refusals)
			continue
		}
		neighbour, ok := cases[exp.neighbour]
		if !ok {
			t.Errorf("case %s names %s as its near neighbour and there is no such case", name, exp.neighbour)
			continue
		}
		if len(neighbour.refusals) != 0 {
			t.Errorf("case %s names %s as its near neighbour, and that case expects %v rather than nothing", name, exp.neighbour, neighbour.refusals)
		}
	}
}

// declarationsInThisRepository is every check name that arrives on a commit
// here: the ones the workflow files declare, and the ones a code-scanning upload
// reports under, which are written down in ReportedOutsideAWorkflowFile because
// no file in this tree carries them.
func declarationsInThisRepository(t *testing.T) []Declared {
	t.Helper()
	declared, err := ReadWorkflows(workflowsInThisRepository)
	if err != nil {
		t.Fatalf("reading %s: %v", workflowsInThisRepository, err)
	}
	return append(declared, ReportedOutsideAWorkflowFile...)
}

// TestNoDeliberateAbsenceNamesSomethingThisTreeDoesNotDeclare is the rename
// refusal, run against this repository rather than against a fixture.
//
// It is here as well as in the workflow because it needs no network. Every entry
// in the absence list is a literal check name, so a job renamed in passing
// leaves its entry pointing at nothing, and this fails in an ordinary suite run
// on the machine of whoever made the rename rather than waiting for the pull
// request. What it cannot do is the other direction, which needs the required
// set and therefore the API, and that half is the workflow's.
func TestNoDeliberateAbsenceNamesSomethingThisTreeDoesNotDeclare(t *testing.T) {
	declared := declarationsInThisRepository(t)
	names := map[string]bool{}
	for _, entry := range declared {
		names[entry.Name] = true
	}
	for _, absence := range Absences {
		if !names[absence.Name] {
			t.Errorf("the absence list names %q and no workflow in this tree declares it, so either a job was renamed and the entry was left behind or the entry never matched", absence.Name)
		}
	}
}

// TestEveryCheckNameThisTreeDeclaresIsWrittenDown is the same reading from the
// other side, and it holds only while the ruleset requires no status check.
//
// The command is what judges this against the live ruleset. This test exists
// because today the required set is empty, which makes the answer knowable
// without asking: every name has to be in the absence list, because there is no
// other place for it to be. When issue #26 assembles the set, a name moving into
// the ruleset is a name leaving this list, and this test is what makes that
// removal deliberate rather than silent.
func TestEveryCheckNameThisTreeDeclaresIsWrittenDown(t *testing.T) {
	declared := declarationsInThisRepository(t)
	written := map[string]bool{}
	for _, absence := range Absences {
		written[absence.Name] = true
	}
	for _, entry := range declared {
		if !written[entry.Name] {
			t.Errorf("%s declares the check name %q and nothing writes it down, so it is a check that runs and holds no merge", entry.Workflow, entry.Name)
		}
	}
}

// TestAnAbsenceCarriesAReason refuses an entry that says only that it is absent.
// The list is read by somebody deciding whether an absence is still right, and
// an entry with no reason gives them nothing to decide with.
func TestAnAbsenceCarriesAReason(t *testing.T) {
	for _, absence := range Absences {
		if strings.TrimSpace(absence.Why) == "" {
			t.Errorf("the absence %q carries no reason", absence.Name)
		}
		if absence.Until != "" && !strings.HasPrefix(absence.Until, "#") {
			t.Errorf("the absence %q says it is retired by %q, which is not an issue reference", absence.Name, absence.Until)
		}
	}
}

// TestAJobWithNoNameOfItsOwnIsNotedRatherThanRefused pins the one place this
// package answers with a note. The platform reports the job id where a job
// carries no name, so such a job is legal and its name is still a gate string, and
// a reader asking where that string came from is told.
func TestAJobWithNoNameOfItsOwnIsNotedRatherThanRefused(t *testing.T) {
	declared := []Declared{{Name: "dependency-review", Workflow: "dependency-review.yml", FromJobID: true}}
	verdict := Judge(declared, []string{"dependency-review"}, nil)
	if len(verdict.Refusals) != 0 {
		t.Fatalf("refused %v", verdict.Refusals)
	}
	if len(verdict.Notes) != 1 {
		t.Fatalf("produced %d note(s), and the fallback to the job id is meant to be said out loud", len(verdict.Notes))
	}
}

// TestANameReportedOutsideAWorkflowFileIsNotedRatherThanRefused pins the other
// place this package answers with a note. A code-scanning upload creates a check
// run named after the analysis rather than after the job that uploaded it, so
// the name is in no workflow file and a comparison that knew only about jobs
// would refuse a required context matching it as one nothing reports. That would
// be a false red on the change that assembles the required set.
func TestANameReportedOutsideAWorkflowFileIsNotedRatherThanRefused(t *testing.T) {
	if len(ReportedOutsideAWorkflowFile) == 0 {
		t.Fatal("nothing is written down as reported outside a workflow file, and the measurement in the list's comment says two names are")
	}
	var required []string
	for _, entry := range ReportedOutsideAWorkflowFile {
		if !entry.FromUpload {
			t.Errorf("%q is in this list and is not marked as coming from an upload, so the report would describe it as read out of a file", entry.Name)
		}
		required = append(required, entry.Name)
	}
	verdict := Judge(ReportedOutsideAWorkflowFile, required, nil)
	if len(verdict.Refusals) != 0 {
		t.Errorf("refused %v, and a required context these names match is reported rather than missing", verdict.Refusals)
	}
	if len(verdict.Notes) != len(ReportedOutsideAWorkflowFile) {
		t.Errorf("produced %d note(s) for %d name(s), and what a reader is holding for these is different from a name read out of a file", len(verdict.Notes), len(ReportedOutsideAWorkflowFile))
	}
}
