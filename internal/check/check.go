// Package check walks a checkout of this repository and reports what it
// examined. It opens files for reading and creates, modifies or removes
// nothing: a checker that repairs the tree it is judging cannot be trusted
// about what it found, because the reader can no longer tell a tree that was
// right from one that was corrected on the way past.
package check

import (
	"fmt"
	"os"
	"path/filepath"
)

// ExperimentsDir is the one directory an experiment may live in. Decision
// record 0002 fixes the layout, and the walk reads that single directory
// rather than searching the tree, so a stray file with the right name
// elsewhere is not an experiment nobody meant to declare.
const ExperimentsDir = "experiments"

// RecordName is the file that holds an experiment's record. Decision record
// 0002 fixes its location; its shape is decided separately.
const RecordName = "EXPERIMENT.md"

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

	// Refusals is what the walk refused. There is nothing to refuse yet, so
	// it is always empty; the refusals arrive in their own changes and each
	// one ships with a fixture that proves it bites.
	Refusals []string
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
		if _, err := readRecord(record); err != nil {
			return res, err
		}
		res.Records++
	}

	return res, nil
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
