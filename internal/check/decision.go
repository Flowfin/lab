package check

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// DecisionsDir is where record 0000 puts the decision records, relative to the
// root of the tree. Record 0002 fixes the directory and record 0000 fixes what
// a file in it looks like, so the two rules below read one directory rather
// than searching for records that might be anywhere.
const DecisionsDir = "docs/decisions"

// The properties the decision records can be refused for. They are separate
// from the properties about an experiment record because the artefact is
// different, and they are in the same walk and the same harness because the
// mechanism is the same one.
const (
	// DecisionRecordIsMissingASection refuses a decision record that does not
	// carry one of the four headings record 0000 fixes. The one that will be
	// missing is the last, because it is the one that takes work, and it is
	// the one the format exists for: a record listing no rejected option and
	// no cost is an announcement rather than a decision, and the rest of this
	// plan is derived from these records.
	DecisionRecordIsMissingASection = "decision-record-is-missing-a-section"

	// TwoDecisionRecordsShareANumber refuses a number written twice. A number
	// is taken from the highest one already landed, so two changes started
	// from the same base take the same number, both are green on their own
	// branches, and nothing notices until they meet. Afterwards no rule has
	// been broken and a reference to that number stops resolving to one
	// thing. The refusal fires on the pair, which is the last moment
	// renumbering is still cheap.
	TwoDecisionRecordsShareANumber = "two-decision-records-share-a-number"

	// SupersessionNamesARecordThatIsNotThere refuses a record that says it
	// supersedes a number no record carries. Records in this plan supersede
	// rather than get edited, so this route is walked before the first
	// release, and a supersession pointing at nothing leaves the superseded
	// decision reading as current.
	SupersessionNamesARecordThatIsNotThere = "supersession-names-a-record-that-is-not-there"
)

// decisionSections are the four headings record 0000 fixes, written here as
// well as there because a check that reads its list out of a document goes
// quiet the day the document is reworded.
//
// Only presence is judged. Record 0000 also fixes the order they appear in and
// says what each one is for, and neither of those is read here: a record
// carrying all four in the wrong order passes, and so does one whose fourth
// section names an option nobody seriously considered. The first is worth a
// separate rule and does not have one; the second is a judgement about meaning
// that no reading of the tree makes, and the review is where it is caught.
var decisionSections = []string{
	"What was decided",
	"What it applies to",
	"What else was considered",
	"What each rejected option would have cost",
}

// decisionFileName is what record 0000 names a decision record: four digits, a
// hyphen, a slug. A file in the directory that does not match is not read at
// all, which is a hole and is named rather than hidden: a record misnamed on
// the way in escapes all three rules below.
var decisionFileName = regexp.MustCompile(`^(\d{4})-[A-Za-z0-9-]+\.md$`)

// supersession is the active form only. A record saying it supersedes 0008 is
// asserting that 0008 is there, and that assertion is what the third rule
// reads. The passive form is deliberately not matched: several records here
// say that a later record will supersede them one day, and a rule that read
// those would refuse a record for naming a number that is not supposed to
// exist yet.
var supersession = regexp.MustCompile(`(?i)supersedes\s+(?:record\s+)?(\d{4})`)

// refuseDecisions reads the decision records and returns how many it read
// along with what it refused. A tree with no decisions directory reads none,
// which is an ordinary state for a fixture tree and is reported rather than
// treated as an error.
func refuseDecisions(root string) (int, []Refusal, error) {
	dir := filepath.Join(root, filepath.FromSlash(DecisionsDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil, nil
		}
		return 0, nil, fmt.Errorf("cannot read %s: %w", dir, err)
	}

	var refusals []Refusal
	read := 0
	numbers := make(map[string]string)
	present := make(map[string]bool)
	var records []string

	// os.ReadDir sorts by filename, so the record that reports a shared number
	// is the later one of the pair every time this runs rather than whichever
	// the filesystem happened to hand over first.
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := decisionFileName.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		read++
		records = append(records, entry.Name())
		present[match[1]] = true
	}

	for _, name := range records {
		path := filepath.Join(dir, name)
		number := decisionFileName.FindStringSubmatch(name)[1]

		if first, taken := numbers[number]; taken {
			refusals = append(refusals, Refusal{
				Property: TwoDecisionRecordsShareANumber,
				Subject:  path,
				Detail:   fmt.Sprintf("it is numbered %s and so is %s, so a reference to %s resolves to two records", number, first, number),
			})
		} else {
			numbers[number] = name
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return read, nil, fmt.Errorf("cannot read %s: %w", path, err)
		}
		text := string(data)

		for _, section := range headingsMissingFrom(text) {
			refusals = append(refusals, Refusal{
				Property: DecisionRecordIsMissingASection,
				Subject:  path,
				Detail:   fmt.Sprintf("it carries no %q section, and record 0000 fixes four: %s", section, strings.Join(decisionSections, "; ")),
			})
		}

		for _, named := range numbersSuperseded(text) {
			if present[named] {
				continue
			}
			refusals = append(refusals, Refusal{
				Property: SupersessionNamesARecordThatIsNotThere,
				Subject:  path,
				Detail:   fmt.Sprintf("it says it supersedes %s and no record in %s carries that number, so whatever it replaced still reads as current", named, DecisionsDir),
			})
		}
	}

	return read, refusals, nil
}

// headingsMissingFrom returns the fixed sections a record does not carry, in
// the order record 0000 writes them.
func headingsMissingFrom(text string) []string {
	carried := make(map[string]bool)
	for _, line := range strings.Split(text, "\n") {
		if heading, isHeading := strings.CutPrefix(strings.TrimSpace(line), "## "); isHeading {
			carried[strings.TrimSpace(heading)] = true
		}
	}

	var missing []string
	for _, section := range decisionSections {
		if !carried[section] {
			missing = append(missing, section)
		}
	}
	return missing
}

// numbersSuperseded returns the record numbers a text claims to supersede,
// each once, sorted so that a record naming two of them refuses in an order a
// reader can predict.
func numbersSuperseded(text string) []string {
	seen := make(map[string]bool)
	var named []string
	for _, match := range supersession.FindAllStringSubmatch(text, -1) {
		if seen[match[1]] {
			continue
		}
		seen[match[1]] = true
		named = append(named, match[1])
	}
	sort.Strings(named)
	return named
}
