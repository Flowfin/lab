package check

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DateFormat is how record 0008 writes a date in a header field.
const DateFormat = "2006-01-02"

// What the state column says when there is no state to print. Each of these is
// a thing a refusal already names, and the listing repeats none of that
// judgement: it is a report, and a report that silently dropped a directory it
// could not read would hide exactly the experiment somebody is looking for.
const (
	stateNoRecord   = "no record"
	stateUnreadable = "unreadable"
	stateNotWritten = "not written"
)

// An Entry is one experiment as the listing reads it. Nothing here is refused
// and nothing is judged. Where a field cannot be read, the column says so.
type Entry struct {
	// Slug is the directory name, which is what the tree calls the
	// experiment whatever its header says.
	Slug string

	// State is the record's State field as it was written, or one of the
	// three strings above where there is none to print.
	State string

	// QuestionWritten is the Question-Written field as it was written, so a
	// value that is not a date is shown rather than swallowed.
	QuestionWritten string

	// Written is that field parsed as a date. It is the zero time where the
	// field is absent or is not a date, and Dated says which.
	Written time.Time

	// Dated says whether Written was read from a date.
	Dated bool
}

// Waiting is how long an experiment has been asking, in whole days, and
// whether the question has a date to count from. It counts calendar days in
// UTC, so a run at one hour of the day and a run at another give the same
// answer for the same record.
func (e Entry) Waiting(now time.Time) (int, bool) {
	if !e.Dated || e.State != StateAsking {
		return 0, false
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return int(today.Sub(e.Written).Hours() / 24), true
}

// A Listing is what one listing walk found.
type Listing struct {
	// Root is the directory the walk started from, as it was given.
	Root string

	// ExperimentsPresent says whether Root held an experiments directory at
	// all, for the same reason Result carries it: a tree with none and a tree
	// whose directory is empty are different statements.
	ExperimentsPresent bool

	// Entries are the experiments, in the order the listing prints them.
	Entries []Entry
}

// List reads every experiment and returns them in the order the listing prints
// them. It refuses nothing and it is not a gate.
//
// The ordering is the whole point of the verb. Everything still asking comes
// first, oldest question at the top, because a record sitting in asking with a
// real question breaks no rule and is the graveyard arriving by the front door.
// Everything else follows in the same order, so a reader scrolling past the
// unanswered work reaches the finished work oldest first rather than at random.
// A record whose question carries no readable date sorts to the end of its
// group by slug, because there is nothing to order it by and the alternative is
// an order that changes between filesystems.
//
// WHERE THIS DOES NOT REACH TODAY. A question dated later than now sorts to the
// bottom of the asking group and its waiting column counts down rather than up,
// which is the stalled experiment this verb exists to surface, concealed by the
// ordering meant to reveal it. Nothing here refuses such a date; issue #63 is
// where that refusal and the one place the time is read are argued. The column
// prints the negative count rather than hiding it, so the case is visible in
// the meantime.
func List(root string, now time.Time) (Listing, error) {
	listing := Listing{Root: root}

	info, err := os.Stat(root)
	if err != nil {
		return listing, fmt.Errorf("cannot read %s: %w", root, err)
	}
	if !info.IsDir() {
		return listing, fmt.Errorf("%s is not a directory", root)
	}

	dir := filepath.Join(root, ExperimentsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return listing, nil
		}
		return listing, fmt.Errorf("cannot read %s: %w", dir, err)
	}
	listing.ExperimentsPresent = true

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		listing.Entries = append(listing.Entries, readEntry(filepath.Join(dir, entry.Name()), entry.Name()))
	}

	sort.SliceStable(listing.Entries, func(i, j int) bool {
		return earlier(listing.Entries[i], listing.Entries[j])
	})
	return listing, nil
}

// readEntry reads one experiment directory into an entry. Every failure to read
// something becomes a value in a column rather than an error, because a listing
// that stops at the first unreadable record tells a reader nothing about the
// rest of the tree.
func readEntry(dir, slug string) Entry {
	entry := Entry{Slug: slug, State: stateNoRecord, QuestionWritten: stateNoRecord}

	data, err := os.ReadFile(filepath.Join(dir, RecordName))
	if err != nil {
		return entry
	}
	record, err := ParseRecord(data)
	if err != nil {
		entry.State = stateUnreadable
		entry.QuestionWritten = stateUnreadable
		return entry
	}

	entry.State = stateNotWritten
	if state, present := record.Field(FieldState); present {
		entry.State = state
	}

	entry.QuestionWritten = stateNotWritten
	if written, present := record.Field(FieldQuestionWritten); present {
		entry.QuestionWritten = written
		if date, err := time.Parse(DateFormat, written); err == nil {
			entry.Written = date
			entry.Dated = true
		}
	}
	return entry
}

// earlier decides which of two entries the listing prints first.
func earlier(a, b Entry) bool {
	if asking(a) != asking(b) {
		return asking(a)
	}
	if a.Dated != b.Dated {
		return a.Dated
	}
	if a.Dated && !a.Written.Equal(b.Written) {
		return a.Written.Before(b.Written)
	}
	return a.Slug < b.Slug
}

func asking(e Entry) bool { return e.State == StateAsking }

// Report writes the listing. It says where it looked and how many experiments
// it found before printing any of them, so a run that found none is a sentence
// rather than an empty screen a reader has to interpret.
//
// The columns are the four the verb exists for: the slug, the state, the date
// the question was written, and how long anything still asking has been there.
// Nothing else, because a listing wide enough to wrap stops being scannable and
// scanning it is the whole point.
func (l Listing) Report(now time.Time) string {
	out := fmt.Sprintf("examined %s\n", l.Root)
	if !l.ExperimentsPresent {
		out += fmt.Sprintf("no %s directory in this tree\n", ExperimentsDir)
	}
	out += fmt.Sprintf("%d %s\n", len(l.Entries), plural(len(l.Entries), "experiment", "experiments"))
	if len(l.Entries) == 0 {
		return out
	}

	rows := [][4]string{{"slug", "state", "question written", "waiting"}}
	for _, entry := range l.Entries {
		waiting := "-"
		if days, counted := entry.Waiting(now); counted {
			waiting = fmt.Sprintf("%d %s", days, plural(days, "day", "days"))
		}
		rows = append(rows, [4]string{entry.Slug, entry.State, entry.QuestionWritten, waiting})
	}

	var width [4]int
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > width[i] {
				width[i] = len(cell)
			}
		}
	}
	for _, row := range rows {
		line := make([]string, 0, len(row))
		for i, cell := range row {
			if i == len(row)-1 {
				line = append(line, cell)
				continue
			}
			line = append(line, cell+strings.Repeat(" ", width[i]-len(cell)))
		}
		out += "  " + strings.Join(line, "  ") + "\n"
	}
	return out
}
