// Package check walks a checkout of this repository and reports what it
// examined. It opens files for reading and creates, modifies or removes
// nothing: a checker that repairs the tree it is judging cannot be trusted
// about what it found, because the reader can no longer tell a tree that was
// right from one that was corrected on the way past.
package check

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	dir := filepath.Join(root, ExperimentsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		// No experiments directory is an ordinary state for this tree and
		// the caller is told about it rather than shown a zero that looks
		// like an empty one. Anything else is a tree the walk cannot read,
		// and reporting that as zero would be the failure this package
		// exists to avoid.
		if os.IsNotExist(err) {
			return res, nil
		}
		return res, fmt.Errorf("cannot read %s: %w", dir, err)
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

		record := filepath.Join(dir, entry.Name(), RecordName)
		recordInfo, err := os.Stat(record)
		if err != nil {
			if os.IsNotExist(err) {
				// An experiment with no record is a thing this repository
				// intends to refuse, and the refusal is built in its own
				// change with the fixture that proves it. Counting it as a
				// record that was not read is all this walk does today.
				continue
			}
			return res, fmt.Errorf("cannot read %s: %w", record, err)
		}
		if !recordInfo.Mode().Type().IsRegular() {
			continue
		}
		data, err := readRecord(record)
		if err != nil {
			return res, err
		}
		res.Records++
		res.Refusals = append(res.Refusals, refuseBytes(record, data)...)
		res.Refusals = append(res.Refusals, refusePaths(root, record, data)...)
	}

	return res, nil
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
