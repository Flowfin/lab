// Package check walks a checkout of this repository and reports what it
// examined. It opens files for reading and creates, modifies or removes
// nothing: a checker that repairs the tree it is judging cannot be trusted
// about what it found, because the reader can no longer tell a tree that was
// right from one that was corrected on the way past.
package check

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// ExperimentsDir is the one directory an experiment may live in. Decision
// record 0002 fixes the layout, and the walk reads that single directory
// rather than searching the tree, so a stray file with the right name
// elsewhere is not an experiment nobody meant to declare.
const ExperimentsDir = "experiments"

// RecordName is the file that holds an experiment's record. Decision record
// 0002 fixes its location; its shape is decided separately.
const RecordName = "EXPERIMENT.md"

// FixturesDir holds the runner's own fixtures. Record 0002 says they are
// fixtures and are not experiments, which is why the two checks below walk
// past it: a fixture tree exists to be refused, and a runner that refused this
// repository for carrying the trees it is proved with would be a runner nobody
// could run here.
const FixturesDir = "testdata"

// gitDir is the checkout's own machinery rather than part of the tree record
// 0002 describes. That record fixes the layout against what git tracks, which
// it shows with `git ls-tree -d --name-only HEAD`, and that command never
// prints this directory. Refusing it would refuse every checkout, so it is
// walked past at the root and never descended into.
const gitDir = ".git"

// rootDirectories is the set record 0002 names at the root of the tree,
// written here as well as there because a check that reads a list out of a
// document is a check that goes quiet when the document is reworded. Adding a
// directory at the root is a change to that record and to this set, which is
// the cost the record takes on deliberately.
//
// Files at the root are not in it and are not judged. They arrive with the
// scaffolding, the legal and the release work, and a rule over them would
// redden most of those changes for a reason that has nothing to do with where
// experiments live.
var rootDirectories = map[string]bool{
	".github":      true,
	"cmd":          true,
	"docs":         true,
	ExperimentsDir: true,
	"internal":     true,
	FixturesDir:    true,
}

// The properties this package can refuse. A property is the rule, named once
// here and named nowhere else, so a fixture declaring what it expects and a
// refusal the runner produced are the same string or they are not equal.
const (
	// RecordBeginsWithAByteOrderMark refuses a record whose first bytes are
	// a byte-order mark. An editor adds one on save, the diff shows nothing
	// unusual, and the first header field then begins with bytes nobody
	// typed and nobody can see. A header field the runner does not
	// recognise is a field whose rule does not apply, and a rule an
	// invisible byte disables is not a rule.
	RecordBeginsWithAByteOrderMark = "record-begins-with-a-byte-order-mark"

	// RecordIsNotText refuses a record that is not text at all: an invalid
	// encoding sequence, or a null byte. Such a file was not written and
	// read by somebody, it arrived another way, and refusing it here costs
	// less than deciding what every later check does with a string the rest
	// of the runner cannot print.
	RecordIsNotText = "record-is-not-text"

	// RecordNamesAPathThatDoesNotResolve refuses a record naming a path
	// inside this repository that is not there at the commit being checked.
	// A record pointing at a file which is not there has quietly stopped
	// being true, and it happens most often exactly when it matters: an
	// experiment is answered, its code is removed under the decision that
	// allows that, and the record still says the measurement is in a script
	// that no longer exists.
	RecordNamesAPathThatDoesNotResolve = "record-names-a-path-that-does-not-resolve"

	// RecordStateIsNotOneOfTheThree refuses a record whose state is not one
	// of the three record 0003 names, which includes a state that is
	// misspelled, a state written with no value, and no state field at all.
	// Every other rule about a record's state is conditional on the state,
	// so a record with no readable state is a record none of them applies
	// to, and a rule that a typo disables is not a rule.
	RecordStateIsNotOneOfTheThree = "record-state-is-not-one-of-the-three"

	// RecordClaimsAnAnswerItDoesNotCarry refuses a record in state
	// answered whose answer section is empty. It does not make anybody
	// answer. What it makes impossible is claiming to have answered while
	// having written nothing, which turns a silent half-finished directory
	// into either a real answer or an honest abandoned.
	RecordClaimsAnAnswerItDoesNotCarry = "record-claims-an-answer-it-does-not-carry"

	// AbandonedRecordDoesNotSayWhatStoppedTheWork refuses a record in state
	// abandoned that says nothing about what stopped the work. Abandoning
	// is allowed here and that is the point of the state existing, but
	// abandoning without a sentence is the same silence this board is built
	// to make visible.
	AbandonedRecordDoesNotSayWhatStoppedTheWork = "abandoned-record-does-not-say-what-stopped-the-work"

	// ExperimentHasNoRecord refuses a directory under experiments/ that
	// carries no record a reader can open. It catches somebody dropping code
	// in and meaning to write the record later, and it fires on the change
	// that creates the directory, which is the only moment when writing the
	// question is still cheap.
	ExperimentHasNoRecord = "experiment-has-no-record"

	// RecordOutsideThePlaceRecordsLive refuses a record that is not directly
	// inside a directory under experiments/. Every other rule about a record
	// reads one the walk of that directory found, so a record anywhere else
	// is a record no rule applies to. The failure is ordinary rather than
	// malicious: a directory copied aside before being deleted, or a second
	// experiment nested inside the first while one question is being split
	// into two. In each of those the run is green and the work is invisible
	// to the only mechanism that would have asked for its question.
	RecordOutsideThePlaceRecordsLive = "record-outside-the-place-records-live"

	// RootHoldsADirectoryTheLayoutDoesNotName refuses a directory at the root
	// of the tree that record 0002 does not name. That record fixes the
	// layout and says a directory it does not name is refused rather than
	// tolerated, and this is the half a machine can see. Adding a top-level
	// directory is a change to the layout, so whoever adds one meets a
	// refusal naming this check and the record behind it, which is the right
	// order.
	RootHoldsADirectoryTheLayoutDoesNotName = "root-holds-a-directory-the-layout-does-not-name"
)

// pathInProse matches a repository-relative path written anywhere in a
// record, in a sentence or inside a link, because the paths that rot are the
// ones written in a sentence rather than in a link.
//
// The first segment has to be a directory record 0002 names. That is what
// keeps the pattern from reading a URL, a fraction or a word pair as a path:
// the leading segment of https://example.com/x is not one of these, and
// neither is the and of and/or. It also means a path this repository could
// never hold is not this check's business, which is the right boundary,
// because nothing here can say whether a file on somebody else's machine
// exists.
var pathInProse = regexp.MustCompile(
	`(^|[^A-Za-z0-9._/-])((?:\./)?(?:experiments|cmd|internal|docs|testdata|\.github)(?:/[A-Za-z0-9._-]+)+)`)

// pathsNamedInProse returns the repository-relative paths a text names, in
// the order it names them, each once.
//
// The leading group is what keeps a URL out. The path segments of
// https://example.invalid/docs/thing.md end in something this pattern would
// otherwise read as a path in this repository, and a record refused for a
// document on somebody else's server is the annoyance that gets a check
// switched off.
func pathsNamedInProse(text string) []string {
	var named []string
	seen := make(map[string]bool)

	for _, match := range pathInProse.FindAllStringSubmatch(text, -1) {
		path := strings.TrimRight(match[2], ".,;:")
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		named = append(named, path)
	}
	return named
}

// byteOrderMark is the UTF-8 encoding of U+FEFF.
var byteOrderMark = []byte{0xEF, 0xBB, 0xBF}

// A Refusal is one rule refusing one subject. The property is what a fixture
// declares and what the harness compares as a set. The subject is what the
// author has to go and look at, and it is carried separately from the detail
// so that a message can never be written without it.
type Refusal struct {
	// Property is the rule that refused, one of the constants above.
	Property string

	// Subject is the path that was refused, as the walk reached it.
	Subject string

	// Detail says what about the subject was wrong, in the words the author
	// needs to make the repair. Two records can trip one property for
	// different reasons and the repairs are different.
	Detail string
}

// String is what a reader sees. It leads with the subject, because somebody
// reading a red run is looking for the file to open first.
func (r Refusal) String() string {
	return fmt.Sprintf("%s: %s (%s)", r.Subject, r.Detail, r.Property)
}

// Result is what one walk examined. The counts are reported whatever they
// are, including zero, because zero found and zero examined are different
// statements and only one of them is evidence.
type Result struct {
	// Root is the directory the walk started from, as it was given.
	Root string

	// ExperimentsPresent says whether Root held an experiments directory at
	// all. A tree with none is not an error and it is not the same as a tree
	// whose experiments directory is empty, so the two are not collapsed
	// into a count of zero that reads the same either way.
	ExperimentsPresent bool

	// Directories is the number of experiment directories walked.
	Directories int

	// Records is the number of experiment records read.
	Records int

	// Refusals is what the walk refused, in the order the walk reached
	// them. Each one names its property and its subject.
	Refusals []Refusal
}

// Properties returns the set of properties this result refused, which is what
// a case declares and what the harness compares. A property refused twice is
// one entry: a verdict is a set, so a fixture tripping one rule at two places
// still refuses exactly that rule.
func (r Result) Properties() []string {
	seen := make(map[string]bool, len(r.Refusals))
	var props []string
	for _, refusal := range r.Refusals {
		if !seen[refusal.Property] {
			seen[refusal.Property] = true
			props = append(props, refusal.Property)
		}
	}
	return props
}

// Walk examines the tree rooted at root and returns what it found. The error
// it returns means the walk could not be made at all, which is a different
// outcome from a walk that completed and refused something.
func Walk(root string) (Result, error) {
	res := Result{Root: root}

	info, err := os.Stat(root)
	if err != nil {
		return res, fmt.Errorf("cannot read %s: %w", root, err)
	}
	if !info.IsDir() {
		return res, fmt.Errorf("%s is not a directory", root)
	}

	rootRefusals, err := refuseRootDirectories(root)
	if err != nil {
		return res, err
	}
	res.Refusals = append(res.Refusals, rootRefusals...)

	if err := walkExperiments(root, &res); err != nil {
		return res, err
	}

	strayRefusals, err := refuseStrayRecords(root)
	if err != nil {
		return res, err
	}
	res.Refusals = append(res.Refusals, strayRefusals...)

	return res, nil
}

// walkExperiments reads the one directory an experiment may live in and holds
// every record it finds to the rules about a record. It is the walk the rest
// of this package was built around; the two refusals either side of it in Walk
// are about where a record is rather than about what one says.
func walkExperiments(root string, res *Result) error {
	dir := filepath.Join(root, ExperimentsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		// No experiments directory is an ordinary state for this tree and
		// the caller is told about it rather than shown a zero that looks
		// like an empty one. Anything else is a tree the walk cannot read,
		// and reporting that as zero would be the failure this package
		// exists to avoid.
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot read %s: %w", dir, err)
	}
	res.ExperimentsPresent = true

	for _, entry := range entries {
		// os.ReadDir does not follow symbolic links, so a link pointing at a
		// directory reports itself as a link and is not walked. The tree the
		// runner reads is untrusted input, and following a link out of it is
		// how a checker reads something nobody put in the repository.
		if !entry.IsDir() {
			continue
		}
		res.Directories++

		experiment := filepath.Join(dir, entry.Name())
		record := filepath.Join(experiment, RecordName)
		recordInfo, err := os.Stat(record)
		if err != nil {
			if os.IsNotExist(err) {
				res.Refusals = append(res.Refusals, Refusal{
					Property: ExperimentHasNoRecord,
					Subject:  experiment,
					Detail:   fmt.Sprintf("there is no %s in it, so it states no question", RecordName),
				})
				continue
			}
			return fmt.Errorf("cannot read %s: %w", record, err)
		}
		if !recordInfo.Mode().Type().IsRegular() {
			res.Refusals = append(res.Refusals, Refusal{
				Property: ExperimentHasNoRecord,
				Subject:  experiment,
				Detail:   fmt.Sprintf("its %s is not a file a reader can open", RecordName),
			})
			continue
		}
		data, err := readRecord(record)
		if err != nil {
			return err
		}
		res.Records++
		res.Refusals = append(res.Refusals, refuseBytes(record, data)...)
		res.Refusals = append(res.Refusals, refusePaths(root, record, data)...)
		res.Refusals = append(res.Refusals, refuseState(record, data)...)
	}

	return nil
}

// refuseRootDirectories holds the root of the tree to the set record 0002
// names. It reads the root and nothing below it, because that is where the
// record applies strictly and where a directory nobody decided on appears.
//
// WHAT IT CANNOT SAY. It reads a filesystem rather than what git tracks, so a
// directory somebody keeps beside the tree without committing it is refused
// exactly like one that was committed. That is the intended answer rather than
// a defect: the refusal is about a directory nobody argued for, and whether it
// reached a commit yet does not change that. The one exception is the
// checkout's own machinery, which gitDir names and which is never part of the
// tree the record describes.
func refuseRootDirectories(root string) ([]Refusal, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", root, err)
	}

	var refusals []Refusal
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == gitDir || rootDirectories[entry.Name()] {
			continue
		}
		refusals = append(refusals, Refusal{
			Property: RootHoldsADirectoryTheLayoutDoesNotName,
			Subject:  filepath.Join(root, entry.Name()),
			Detail:   fmt.Sprintf("record 0002 names %s at the root and this is not one of them, so adding it is a change to that record", rootDirectoriesInWords()),
		})
	}
	return refusals, nil
}

// rootDirectoriesInWords is the permitted set in one string, sorted, so that a
// message naming it is derived from the set the check reads rather than from a
// list somebody typed into a format string and has to keep true.
func rootDirectoriesInWords() string {
	names := make([]string, 0, len(rootDirectories))
	for name := range rootDirectories {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// refuseStrayRecords finds every record in the tree that is not directly
// inside a directory under experiments/. It is the boundary every other record
// rule rests on: those read a record the walk of experiments/ found, so a
// record the walk never reaches is one none of them applies to.
//
// TWO PLACES IT DOES NOT LOOK, both of them deliberate. It does not descend
// into the checkout's own machinery, which holds no record anybody wrote. It
// does not descend into the fixtures either, which record 0002 says are
// fixtures and are not experiments: a fixture tree exists to be refused, and a
// runner that refused this repository for carrying the trees that prove it
// could not be run here at all. A record that is genuinely misplaced under
// testdata/ is therefore not refused, and a green run does not say otherwise.
func refuseStrayRecords(root string) ([]Refusal, error) {
	var refusals []Refusal

	// filepath.WalkDir does not follow symbolic links, so a link pointing out
	// of the tree is reported as a link and never descended into.
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && (entry.Name() == gitDir || entry.Name() == FixturesDir) {
				return fs.SkipDir
			}
			return nil
		}
		if entry.Name() != RecordName {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("cannot place %s inside %s: %w", path, root, err)
		}
		segments := strings.Split(filepath.ToSlash(relative), "/")
		if len(segments) == 3 && segments[0] == ExperimentsDir {
			return nil
		}
		refusals = append(refusals, Refusal{
			Property: RecordOutsideThePlaceRecordsLive,
			Subject:  path,
			Detail: fmt.Sprintf("a record lives at %s/<slug>/%s and this one is at %s, so no rule about a record reaches it",
				ExperimentsDir, RecordName, filepath.ToSlash(relative)),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s: %w", root, err)
	}
	return refusals, nil
}

// refusePaths holds a record to the paths it names. A path that was removed
// on purpose is the case this exists to catch rather than an exception to it:
// the repair is to update the record to name the commit that removed the file,
// which the code-removal decision already requires, and not to weaken this.
//
// Name no path you do not intend to resolve. A record naming an example path
// that was never meant to exist is refused, and that is expected rather than a
// defect in the check.
func refusePaths(root, path string, data []byte) []Refusal {
	var refusals []Refusal

	for _, named := range pathsNamedInProse(string(data)) {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(named))); err == nil {
			continue
		}
		refusals = append(refusals, Refusal{
			Property: RecordNamesAPathThatDoesNotResolve,
			Subject:  path,
			Detail:   fmt.Sprintf("it names %s, which is not in this tree", named),
		})
	}

	return refusals
}

// refuseState holds a record to what its own state says about it. Record 0003
// fixes the three states and what each one requires; this is the half of that
// a machine can see.
//
// What it buys is smaller than it looks and the difference is worth keeping
// straight. It does not make anybody finish an experiment or write an honest
// answer. It makes a record that contradicts itself refusable, which converts
// a silent half-finished directory into either a real answer or an honest
// abandoned. A record saying "no, this does not work" and a record saying it
// while its author knows otherwise are the same bytes, and review is the only
// thing that catches the second.
//
// WHERE THIS DOES NOT REACH. A record whose bytes do not parse as a record is
// not judged here at all. Nothing can read the state of something that has no
// header, and inventing a state refusal for it would name the wrong repair:
// whoever hits it has a file that is not a record rather than a record with a
// bad state. So a record with no header passes every rule below, which is a
// hole, and it is issue #16's rather than this one's. Read a green run
// accordingly.
func refuseState(path string, data []byte) []Refusal {
	rec, err := ParseRecord(data)
	if err != nil {
		return nil
	}

	state, present := rec.Field(FieldState)
	switch {
	case !present:
		return []Refusal{{
			Property: RecordStateIsNotOneOfTheThree,
			Subject:  path,
			Detail:   fmt.Sprintf("it declares no %s field, so no rule about a state applies to it. record 0003 names %s", FieldState, statesInWords()),
		}}
	case state == "":
		return []Refusal{{
			Property: RecordStateIsNotOneOfTheThree,
			Subject:  path,
			Detail:   fmt.Sprintf("its %s field is there and empty, which is not one of %s", FieldState, statesInWords()),
		}}
	case state != StateAsking && state != StateAnswered && state != StateAbandoned:
		return []Refusal{{
			Property: RecordStateIsNotOneOfTheThree,
			Subject:  path,
			Detail:   fmt.Sprintf("its %s is %q, and record 0003 names %s", FieldState, state, statesInWords()),
		}}
	}

	// The answer section carries both. For an answered record it is the
	// answer; for an abandoned one it is the sentence saying what stopped
	// the work, which record 0003 makes a lower bar than an answer on
	// purpose. They are two properties rather than one because the repairs
	// are different: an answered record with nothing under the heading is
	// either finished or mislabelled, and an abandoned one is one sentence
	// away from being honest.
	answer, _ := rec.Section(SectionAnswer)
	if strings.TrimSpace(answer) != "" {
		return nil
	}

	switch state {
	case StateAnswered:
		return []Refusal{{
			Property: RecordClaimsAnAnswerItDoesNotCarry,
			Subject:  path,
			Detail:   fmt.Sprintf("it says %s and its %s section is empty, so it claims an answer it does not carry", StateAnswered, SectionAnswer),
		}}
	case StateAbandoned:
		return []Refusal{{
			Property: AbandonedRecordDoesNotSayWhatStoppedTheWork,
			Subject:  path,
			Detail:   fmt.Sprintf("it says %s and its %s section is empty, so it does not say what stopped the work", StateAbandoned, SectionAnswer),
		}}
	}

	// asking, with an empty answer section, is the ordinary state of an
	// experiment that is running. Record 0008 asks for that heading to be
	// there and empty from the day the record is created.
	return nil
}

// statesInWords is the three states in one string, so that a message naming
// them is derived from the constants rather than from a list somebody typed
// into a format string and has to keep true.
func statesInWords() string {
	return strings.Join([]string{StateAsking, StateAnswered, StateAbandoned}, ", ")
}

// refuseBytes holds a record to being the text every later check assumes it
// is. It judges nothing about what the text says: that the bytes are text,
// and that the header a later check reads is the header the author wrote, is
// the whole of it.
func refuseBytes(path string, data []byte) []Refusal {
	var refusals []Refusal

	if bytes.HasPrefix(data, byteOrderMark) {
		refusals = append(refusals, Refusal{
			Property: RecordBeginsWithAByteOrderMark,
			Subject:  path,
			Detail:   "the file begins with a byte-order mark, so the first header field starts with three bytes nobody typed",
		})
	}

	if i := bytes.IndexByte(data, 0); i >= 0 {
		refusals = append(refusals, Refusal{
			Property: RecordIsNotText,
			Subject:  path,
			Detail:   fmt.Sprintf("a null byte at offset %d", i),
		})
	} else if !utf8.Valid(data) {
		// Checked second and only when there is no null byte, so a file
		// carrying both is named by the more specific of the two rather
		// than by whichever the code happened to reach first.
		refusals = append(refusals, Refusal{
			Property: RecordIsNotText,
			Subject:  path,
			Detail:   "the bytes are not a valid encoding sequence",
		})
	}

	return refusals
}

// readRecord opens a record and returns its bytes. Nothing reads the contents
// yet. It exists so that a record counted as read is one the walk actually
// opened, rather than one it saw the name of, since a file that cannot be
// opened would otherwise be counted as examined.
func readRecord(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	return data, nil
}

// Report writes what the walk examined, in the order a reader needs it: where
// it looked, what it covered, and only then what it refused. A run that
// covered less than the whole tree says so on its own line, so it cannot be
// read as one that covered everything and found nothing.
func (r Result) Report() string {
	out := fmt.Sprintf("examined %s\n", r.Root)
	if !r.ExperimentsPresent {
		out += fmt.Sprintf("no %s directory in this tree\n", ExperimentsDir)
	}
	out += fmt.Sprintf("%d experiment %s walked, %d %s read\n",
		r.Directories, plural(r.Directories, "directory", "directories"),
		r.Records, plural(r.Records, "record", "records"))
	out += fmt.Sprintf("%d refused\n", len(r.Refusals))
	for _, refusal := range r.Refusals {
		out += fmt.Sprintf("  %s\n", refusal)
	}
	return out
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
