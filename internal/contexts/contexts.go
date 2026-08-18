// Package contexts compares the two lists that have to agree for the merge
// gate on the default branch to mean what a reader thinks it means: the
// contexts the ruleset requires, and the check names the workflows in this
// repository declare.
//
// Five issues in the plan carry the same warning in different words. The check
// name is what the required set refers to, and a job renamed in passing removes
// itself from that set while the tab still looks green. Two other pieces of work
// come close and neither closes it. One assembles the required set from the
// names a completed run reported, which settles what the contexts are called on
// the day it runs. One walks the pull requests those contexts have to arrive on,
// which settles that they arrive on the day it walks. Both are true when they
// are done and neither notices the rename afterwards.
//
// THE JUDGEMENT IS A FUNCTION OVER VALUES AND IT READS NOTHING. Judge takes two
// lists and returns a verdict. Nothing here opens a file, runs a command or
// reads an environment variable, so every rule below is proved against a value a
// test writes out in full. Where the declared names come from is workflows.go's
// half, and the command under cmd/contexts is the only place that asks the API
// for the required set.
//
// # What a green verdict here does not say
//
// IT CANNOT JUDGE BEHAVIOUR. A job that reports the right name having verified
// nothing passes this comparison exactly like one that did its work. Nothing in
// a name says what ran under it, and the proof that a check bites is the fixture
// its own suite carries rather than the string it reports under. A green result
// here says the gate refers to checks that exist and that every check the tree
// declares is accounted for, and it says nothing whatever about what any of them
// found.
//
// IT READS THE RULESET AS THE API ANSWERED ON THE DAY IT RAN. The required set
// is a live setting rather than a tracked file, so a context added or removed by
// hand is caught on the next pull request rather than at the moment it changes.
// Between those two moments the gate and this comparison disagree and nothing
// says so. That is the ordinary failure of anything that describes a setting it
// does not own, and it is stated here because a green result is otherwise read
// as the gate being intact right now.
package contexts

import (
	"fmt"
	"sort"
	"strings"
)

// The properties this package can refuse. A property is the rule, named once
// here and named nowhere else, so a fixture declaring what it expects and a
// refusal the judgement produced are the same string or they are not equal.
const (
	// RequiredContextNothingReports refuses a context the ruleset requires
	// that no workflow in this tree declares. It blocks every merge and gives
	// no reason on the pull request, because a context that never arrives is
	// indistinguishable from one that has not finished. The first response is
	// to wait and the second is to drop the context from the required set,
	// which repairs a blocked board by making the gate smaller.
	RequiredContextNothingReports = "required-context-nothing-reports"

	// DeclaredNameOutsideTheRequiredSet refuses a check name the workflows
	// declare that is neither in the required set nor written down as a
	// deliberate absence. This is the quieter of the two directions: a gate
	// somebody believes in and nothing enforces. It survives indefinitely, and
	// a green board is what hides it.
	DeclaredNameOutsideTheRequiredSet = "declared-name-outside-the-required-set"

	// AbsenceNamesNothingDeclared refuses a deliberate absence naming a check
	// name no workflow in this tree declares. It is the direction that catches
	// the rename while the required set is still empty: an absence is written
	// against a literal string, so a job renamed in passing leaves its absence
	// pointing at nothing and this refuses rather than passing in silence.
	AbsenceNamesNothingDeclared = "absence-names-nothing-declared"

	// AbsenceIsRequired refuses a deliberate absence naming a context the
	// ruleset requires. The two statements contradict each other, and the
	// contradiction is what makes the absence list load-bearing rather than
	// decorative: an entry that stayed behind after its context became
	// required would silently exempt that context from the direction above.
	AbsenceIsRequired = "absence-is-required"

	// NameCarriesAnExpression refuses a declared check name still carrying a
	// workflow expression the reader could not resolve. A name reaching the
	// comparison half-expanded would be compared against a string no gate can
	// ever hold, so this refuses rather than reporting a drift that is really
	// a reader that did not understand the file.
	NameCarriesAnExpression = "name-carries-an-expression"
)

// Absence is one check name this tree declares and deliberately keeps out of
// the required set, with the reason and with what retires it.
//
// THE LIST IS DATA AND NOT A JUDGEMENT. A comparison that cannot tell an
// intended absence from a drift produces a red board nobody can act on, and one
// that infers the intent produces a green board nobody can check. So every
// absence is written down as a literal string beside the reason for it, and an
// entry that stops matching the tree is a refusal rather than a silent pass.
type Absence struct {
	// Name is the check name, exactly as the workflow declares it.
	Name string

	// Why is the reason this name is outside the required set.
	Why string

	// Until names the issue that retires this entry, or is empty where the
	// absence is permanent. The difference is the whole value of the field: a
	// permanent absence is a decision and a pending one is a debt, and a list
	// that collapsed them would read as settled everywhere.
	Until string
}

// Absences is the deliberate-absence list, and it is here rather than in a
// document because this is where the comparison is defined and a list written
// somewhere else drifts against the thing that reads it.
//
// MOST OF THIS LIST IS ONE FACT REPEATED. The ruleset on the default branch
// requires no status check at all today, so every check name this tree declares
// is outside the required set, and each of them is written down rather than the
// whole comparison being switched off while the set is empty. What that buys is
// the rename: an absence names a literal string, so a job renamed while the set
// is still empty leaves its entry pointing at nothing and reddens this check.
// A comparison that special-cased the empty set would notice none of it.
//
// Issue #26 is where the set is assembled, and each pending entry below is
// retired by moving its name into the ruleset rather than by editing a document.
var Absences = []Absence{
	{
		Name:  "Scorecard analysis",
		Why:   "the supply-chain self-audit publishes from the default branch and has no pull-request trigger, so requiring it would require a context that never arrives on the thing being gated",
		Until: "",
	},
	{Name: "build (linux/amd64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "build (linux/arm64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "build (darwin/amd64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "build (darwin/arm64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "build (windows/amd64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "build (windows/arm64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "test (linux/amd64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "test (windows/amd64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "test (darwin/arm64)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "vet", Why: theSetIsEmpty, Until: "#26"},
	{Name: "format", Why: theSetIsEmpty, Until: "#26"},
	{Name: "CodeQL (go)", Why: theSetIsEmpty, Until: "#26"},
	{Name: "CodeQL", Why: theSetIsEmpty, Until: "#26"},
	{Name: "zizmor", Why: theSetIsEmpty, Until: "#26"},
	{Name: "DCO sign-off", Why: theSetIsEmpty, Until: "#26"},
	{Name: "dependency-review", Why: theSetIsEmpty, Until: "#26"},
	{Name: "headless and unelevated", Why: theSetIsEmpty, Until: "#26"},
	{Name: "invariants", Why: theSetIsEmpty, Until: "#26"},
	{Name: "prose format", Why: theSetIsEmpty, Until: "#26"},
	{Name: "pull request", Why: theSetIsEmpty, Until: "#26"},
	{Name: "records", Why: theSetIsEmpty, Until: "#26"},
	{Name: "required contexts", Why: theSetIsEmpty, Until: "#26"},
	{Name: "Reject Trojan Source Unicode", Why: theSetIsEmpty, Until: "#26"},
	{Name: "Audit workflows (zizmor)", Why: theSetIsEmpty, Until: "#26"},
}

// theSetIsEmpty is the reason every pending entry above carries, written once so
// two dozen rows cannot drift into two dozen slightly different sentences.
const theSetIsEmpty = "the ruleset on the default branch requires no status check at all today, so no name this tree declares can be in a set that has no members"

// ReportedOutsideAWorkflowFile is every check name that arrives on a commit here
// and is written in no workflow file, so the reader of those files can never
// find it.
//
// NOT EVERY CHECK RUN COMES FROM A JOB. A code-scanning upload creates a check
// run of its own, named after the analysis rather than after the job that
// uploaded it, and it is a name a ruleset can require exactly like any other. A
// comparison that knew only about jobs would refuse such a context as one
// nothing reports, which is a false red on the change that assembles the
// required set, and a false red is the failure this comparison exists to avoid
// rather than to cause.
//
// Measured rather than assumed. Every check run on one commit, with the
// application that created it:
//
//	gh api "repos/Flowfin/lab/commits/$(git rev-parse HEAD)/check-runs" --paginate \
//	  --jq '.check_runs[] | "\(.name)\t\(.app.slug)"' | sort -u
//
// On 2026-08-12 that printed twenty-four names, of which twenty-two came from
// github-actions and the two below came from github-advanced-security. Re-run it
// before trusting this list.
//
// WHAT THIS LIST COSTS. These two names are held here rather than read out of a
// file, so the rename this whole check is about is not caught for them: the
// string a code-scanning upload reports under is decided by the analysis rather
// than by a line somebody can change in this tree, and nothing here reads it.
// For these two the comparison is a statement that the name is expected, and for
// every other name it is a statement about what the tree says. The two are
// different claims and the report says which one a reader is holding.
var ReportedOutsideAWorkflowFile = []Declared{
	{
		Name:       "CodeQL",
		Workflow:   "codeql.yml",
		FromUpload: true,
	},
	{
		Name:       "zizmor",
		Workflow:   "zizmor.yml",
		FromUpload: true,
	},
}

// Declared is one check name a workflow in this tree declares.
type Declared struct {
	// Name is the check name as it will be reported.
	Name string

	// Workflow is the file it was read from, so a refusal names the file a
	// reader has to open rather than only the string that was wrong.
	Workflow string

	// FromJobID says the name was not written in the file and was taken from
	// the job id, which is what the platform reports when a job carries no
	// name. It is a note rather than a refusal: the fallback is the platform's
	// own rule and a job relying on it is legal, but a reader asking where a
	// gate's string comes from is entitled to know it came from an identifier
	// rather than from a line somebody chose.
	FromJobID bool

	// FromUpload says this name was not read out of a workflow file at all and
	// comes from ReportedOutsideAWorkflowFile. Workflow then names the file
	// whose run causes the check to appear rather than the file the string was
	// read from, and a rename of the string is not something this comparison
	// can see. It is a note for the same reason as the field above: what a
	// reader is holding is different for these names and the report says so.
	FromUpload bool
}

// Refusal is one violation, named by its property so a fixture and a run compare
// the same strings.
type Refusal struct {
	Property string
	Subject  string
	Detail   string
}

func (r Refusal) String() string {
	return fmt.Sprintf("%s: %s (%s)", r.Subject, r.Detail, r.Property)
}

// Note is something worth printing that is not a violation.
type Note struct {
	Subject string
	Detail  string
}

// Verdict is what one comparison produced.
type Verdict struct {
	Refusals []Refusal
	Notes    []Note
}

// Properties is the set of property ids this verdict refused, sorted and
// deduplicated, which is what a fixture compares against.
func (v Verdict) Properties() []string {
	seen := map[string]bool{}
	var props []string
	for _, refusal := range v.Refusals {
		if !seen[refusal.Property] {
			seen[refusal.Property] = true
			props = append(props, refusal.Property)
		}
	}
	sort.Strings(props)
	return props
}

// Judge compares what the ruleset requires against what the workflows declare,
// with the deliberate absences it is given.
//
// The absences are a parameter rather than a package-level read, so a fixture
// proves the rules against a list it wrote out in full and never against this
// repository's own. The list this repository runs with is Absences above, and
// the command passes it in.
func Judge(declared []Declared, required []string, absences []Absence) Verdict {
	var verdict Verdict

	declaredByName := map[string]Declared{}
	for _, entry := range declared {
		if strings.Contains(entry.Name, "${{") {
			verdict.Refusals = append(verdict.Refusals, Refusal{
				Property: NameCarriesAnExpression,
				Subject:  entry.Workflow,
				Detail:   fmt.Sprintf("the check name %q still carries a workflow expression, so what this job reports under was not resolved and nothing here can be compared against it", entry.Name),
			})
			continue
		}
		declaredByName[entry.Name] = entry
		if entry.FromJobID {
			verdict.Notes = append(verdict.Notes, Note{
				Subject: entry.Workflow,
				Detail:  fmt.Sprintf("the check name %q comes from the job id, because the job carries no name of its own", entry.Name),
			})
		}
		if entry.FromUpload {
			verdict.Notes = append(verdict.Notes, Note{
				Subject: entry.Workflow,
				Detail:  fmt.Sprintf("the check name %q is written in no workflow file and is expected here rather than read, so a change to the string it reports under is not caught by this comparison", entry.Name),
			})
		}
	}

	requiredSet := map[string]bool{}
	for _, context := range required {
		requiredSet[context] = true
	}

	absent := map[string]bool{}
	for _, absence := range absences {
		absent[absence.Name] = true
		if _, ok := declaredByName[absence.Name]; !ok {
			verdict.Refusals = append(verdict.Refusals, Refusal{
				Property: AbsenceNamesNothingDeclared,
				Subject:  absence.Name,
				Detail:   fmt.Sprintf("this name is written down as deliberately outside the required set, and no workflow in this tree declares it. Either a job was renamed and this entry was left behind, or the entry never matched. The reason it carries is %q", absence.Why),
			})
		}
		if requiredSet[absence.Name] {
			verdict.Refusals = append(verdict.Refusals, Refusal{
				Property: AbsenceIsRequired,
				Subject:  absence.Name,
				Detail:   fmt.Sprintf("the ruleset requires this context and this list says it is deliberately outside the required set, so the two disagree. The reason the list carries is %q", absence.Why),
			})
		}
	}

	for _, context := range sortedKeys(requiredSet) {
		if _, ok := declaredByName[context]; !ok {
			verdict.Refusals = append(verdict.Refusals, Refusal{
				Property: RequiredContextNothingReports,
				Subject:  context,
				Detail:   "the ruleset on the default branch requires this context and no workflow in this tree declares a check name matching it, so every merge waits for a tick that never arrives",
			})
		}
	}

	for _, name := range sortedKeys(declaredByName) {
		if requiredSet[name] || absent[name] {
			continue
		}
		verdict.Refusals = append(verdict.Refusals, Refusal{
			Property: DeclaredNameOutsideTheRequiredSet,
			Subject:  name,
			Detail:   fmt.Sprintf("%s declares this check name, the ruleset does not require it, and nothing writes it down as a deliberate absence, so it is a check that runs and holds no merge", declaredByName[name].Workflow),
		})
	}

	return verdict
}

// Report is what the run printed. It says what it examined before it says what
// it found, because a run that compared two empty lists and a run that compared
// two full ones both produce no refusals and are otherwise indistinguishable.
func (v Verdict) Report(declared []Declared, required []string, absences []Absence) string {
	out := "compared the required contexts against the check names this tree declares\n"
	out += fmt.Sprintf("  required by the ruleset: %d\n", len(required))
	out += fmt.Sprintf("  declared by the workflows: %d\n", len(declared))
	out += fmt.Sprintf("  written down as deliberately absent: %d\n", len(absences))

	for _, note := range v.Notes {
		out += fmt.Sprintf("note: %s: %s\n", note.Subject, note.Detail)
	}
	for _, refusal := range v.Refusals {
		out += "refused: " + refusal.String() + "\n"
	}
	out += fmt.Sprintf("%d refusal(s), %d note(s)\n", len(v.Refusals), len(v.Notes))
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
