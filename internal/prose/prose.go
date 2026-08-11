// Package prose holds the formatting rules over this repository's tracked
// prose. The records are the primary artefact here rather than a supporting
// detail, so the diff a reviewer reads is the product, and a whitespace change
// mixed into a content change is the one thing that reliably hides a sentence
// nobody meant to change.
//
// It is a separate package from the invariants for the same reason those are
// separate from the record checks. Those say a guard is intact; these say the
// bytes around the words are the bytes everybody agreed on. A reader looking at
// a red tick should know which of the two broke without opening it.
//
// WHAT IT READS. Every file whose name ends in .md, from the root it is given
// downwards. Markdown is what carries prose here: the experiment records, the
// decision records, the documents under docs/, the files at the root and the
// templates under .github/. A file that is not Markdown is not read, and the
// one at the root a reader might expect to be is named at isProse, with the
// reason, rather than left to be discovered.
//
// WHAT IT DOES NOT READ. The fixtures under testdata/. The harness stores them
// so that no translation can touch them, because a fixture that exists to prove
// a byte is refused loses its whole value the moment something normalises it. A
// formatter or a formatting check pointed at that directory would refuse the
// evidence that the checks work, and this is the near miss worth naming: it is
// one path entry away, and the failure it produces leaves no trace, because the
// fixture is still there and the test is still green.
package prose

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The extension that decides what this package reads, and the directories it
// never descends into.
const (
	// Extension is what a file carrying prose here is called.
	Extension = ".md"

	// FixturesDir is the directory this package does not scan, for the reason
	// the package comment gives.
	FixturesDir = "testdata"

	// gitDir is the checkout's own machinery. It is not tracked text and
	// walking it would read every blob in the history.
	gitDir = ".git"
)

// The properties this package can refuse. A property is the rule, named once
// here and nowhere else, so a case declaring what it expects and a refusal the
// scan produced are the same string or they are not equal.
//
// Each of the three is a rule .editorconfig already declares for every file in
// this tree. That file asks an editor to produce these bytes and nothing reads
// whether it did, which is the gap: an editor that ignores it, a paste from a
// browser and a generated block all arrive looking exactly like typed text.
const (
	// ProseCarriesTrailingWhitespace refuses a line ending in a space or a
	// tab. It is invisible in every editor and in the rendered page, it moves
	// under a reformat, and the diff it produces is a line marked changed
	// where nothing a reader can see has changed. That is the shape that
	// hides the sentence next to it.
	ProseCarriesTrailingWhitespace = "prose-carries-trailing-whitespace"

	// ProseCarriesATab refuses a tab anywhere in a prose file. A tab is two
	// columns wide in one reader's view and eight in another's, so an indented
	// block that lines up for whoever wrote it does not line up for anybody
	// else, and the alignment a record uses to set a command apart from the
	// sentence around it stops being a property of the file.
	ProseCarriesATab = "prose-carries-a-tab"

	// ProseDoesNotEndWithExactlyOneNewline refuses a file that ends without a
	// newline or with more than one, at two sites. The first makes the last
	// line of a record collide with the first line of whatever a tool prints
	// next and shows in a diff as a change to a line nobody touched. The
	// second is invisible and grows, one blank line per careless save.
	ProseDoesNotEndWithExactlyOneNewline = "prose-does-not-end-with-exactly-one-newline"
)

// A Refusal is one rule refusing one subject. It carries the subject separately
// from the detail so that a message can never be written without it.
type Refusal struct {
	Property string
	Subject  string
	Detail   string
}

// String leads with the subject, because somebody reading a red run is looking
// for the file to open first.
func (r Refusal) String() string {
	return fmt.Sprintf("%s: %s (%s)", r.Subject, r.Detail, r.Property)
}

// A Report is what one scan examined. The count is reported whatever it is,
// including zero, because a scan that read nothing and a scan that read
// everything and found nothing look identical without it.
type Report struct {
	// Root is the directory the scan started from, as it was given.
	Root string

	// Files is the number of prose files the scan read.
	Files int

	// Refusals is what the scan refused, in the order it reached them.
	Refusals []Refusal
}

// Properties returns the set of properties this report refused, which is what a
// case declares and what the harness compares. A property refused twice is one
// entry: a verdict is a set, so a file departing from one rule on four lines
// still refuses exactly that rule.
func (r Report) Properties() []string {
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

// Scan reads the prose under root and returns what it found. The error it
// returns means the scan could not be made at all, which is a different outcome
// from a scan that completed and refused something.
func Scan(root string) (Report, error) {
	rep := Report{Root: root}

	info, err := os.Stat(root)
	if err != nil {
		return rep, fmt.Errorf("cannot read %s: %w", root, err)
	}
	if !info.IsDir() {
		return rep, fmt.Errorf("%s is not a directory", root)
	}

	files, err := proseFiles(root)
	if err != nil {
		return rep, err
	}
	rep.Files = len(files)

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return rep, fmt.Errorf("cannot read %s: %w", path, err)
		}
		rep.Refusals = append(rep.Refusals, refuseFormat(path, string(data))...)
	}
	return rep, nil
}

// proseFiles walks the tree and returns every prose file, sorted, so that a
// report reads the same way on every filesystem the suite runs on.
func proseFiles(root string) ([]string, error) {
	var files []string

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
		if entry.Type().IsRegular() && isProse(entry.Name()) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s: %w", root, err)
	}

	sort.Strings(files)
	return files, nil
}

// isProse says whether a file carries prose these rules apply to.
//
// The extension is the whole test, and the one file at the root a reader might
// expect to be here and is not is DCO. That is the Developer Certificate of
// Origin verbatim, taken from its canonical source, and reformatting somebody
// else's legal text to satisfy a rule of ours would change what a contributor
// asserts when they sign off. It is left exactly as it arrived.
func isProse(name string) bool {
	return strings.HasSuffix(name, Extension)
}

// refuseFormat holds one prose file to the three rules.
//
// WHAT IT DOES NOT JUDGE, and it is worth saying because the check sits next to
// the ones that judge records. Formatting is not correctness and it is not
// honesty. A record can be perfectly formatted and say nothing, and the
// refusals over the question and the answer are what catch that. What this buys
// is that the diff a reviewer reads carries only the changes somebody meant,
// and claiming more for it would be claiming that a tidy record is a true one.
//
// Line length is deliberately not among the rules. Measured before the choice
// rather than after: the documents in this tree carry fenced blocks holding
// command output, which cannot be wrapped without changing what the command
// printed, so a length rule would either refuse honest evidence or need an
// exception for the fences, and a rule with an exception is one people stop
// trusting. .editorconfig still asks an editor for 80 columns in Markdown,
// which is the right place for a preference that nothing refuses.
func refuseFormat(path, text string) []Refusal {
	var refusals []Refusal

	for i, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimRight(line, " \t"); trimmed != line {
			refusals = append(refusals, Refusal{
				Property: ProseCarriesTrailingWhitespace,
				Subject:  path,
				Detail:   fmt.Sprintf("line %d ends in whitespace nobody can see", i+1),
			})
		}
		if strings.Contains(line, "\t") {
			refusals = append(refusals, Refusal{
				Property: ProseCarriesATab,
				Subject:  path,
				Detail:   fmt.Sprintf("line %d carries a tab, which is a different width for every reader", i+1),
			})
		}
	}

	// An empty file has no last byte to judge and is not refused here. It is
	// an odd file rather than a badly formatted one, and the rules that care
	// whether a record says anything are the record checks.
	if text == "" {
		return refusals
	}

	switch {
	case !strings.HasSuffix(text, "\n"):
		refusals = append(refusals, Refusal{
			Property: ProseDoesNotEndWithExactlyOneNewline,
			Subject:  path,
			Detail:   "its last line has no newline after it, so a diff marks it changed the next time anybody adds a line",
		})
	case strings.HasSuffix(text, "\n\n"):
		refusals = append(refusals, Refusal{
			Property: ProseDoesNotEndWithExactlyOneNewline,
			Subject:  path,
			Detail:   "it ends with a blank line, which is invisible and grows one save at a time",
		})
	}

	return refusals
}

// String writes what the scan examined, in the order a reader needs it: where
// it looked, what it covered, and only then what it refused.
func (r Report) String() string {
	out := fmt.Sprintf("examined %s\n", r.Root)
	out += fmt.Sprintf("%d prose %s read, and %s is not read\n", r.Files, plural(r.Files, "file", "files"), FixturesDir)
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
