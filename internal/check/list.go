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

// What the hardware column says about a record that carries no such field.
// Record 0013 makes the field optional and record 0015 adds it, so a record
// written before it declares nothing, and "not declared" is a different
// statement from "none". Collapsing the two would let the listing report a
// silence as a claim.
const hardwareNotDeclared = "not declared"

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

	// NeedsHardware is the Needs-Hardware field as it was written, or one of
	// the strings above where there is none to print. It is shown rather than
	// judged: a declaration the directory contradicts is a refusal the
	// checker produces, and repeating that judgement here would put a verdict
	// in a report.
	NeedsHardware string
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

// refuseDates holds a record to the one notion of now the run was given. It
// reads the question's date and nothing else about time.
//
// WHAT THIS DOES NOT BUY, and it belongs here rather than only in the issue
// that argued it. Nothing makes the host clock right. A machine set two years
// fast refuses honest records and a machine set two years slow accepts a future
// date as an ordinary one, and no reading of a checkout separates either case
// from a correct one. What the rule buys is that the runner has one notion of
// now instead of several, that a reader can see which value a run used because
// every run prints it, and that the listing's ordering is a property of the
// records rather than of when somebody happened to look.
//
// A record whose bytes do not parse is not judged here, for the same reason the
// state rules do not judge one: nothing can read a date out of something that
// has no header.
func refuseDates(path string, data []byte, now time.Time) []Refusal {
	record, err := ParseRecord(data)
	if err != nil {
		return nil
	}

	written, present := record.Field(FieldQuestionWritten)
	if !present {
		return nil
	}
	date, err := time.Parse(DateFormat, written)
	if err != nil {
		// A value that is not a date is a different repair and a different
		// rule. Refusing it here would name the wrong one.
		return nil
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if !date.After(today) {
		return nil
	}
	return []Refusal{{
		Property: QuestionDatedLaterThanTheRun,
		Subject:  path,
		Detail: fmt.Sprintf("its %s is %s and this run read %s, so the listing sorts it below every honest question and leaves it there",
			FieldQuestionWritten, written, today.Format(DateFormat)),
	}}
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
// A question dated later than the time the run read would sort to the bottom of
// the asking group and its waiting column would count down rather than up,
// which is the stalled experiment this verb exists to surface concealed by the
// ordering meant to reveal it. The listing does not repair that and does not
// hide it: it prints the negative count, and refuseDates is what refuses the
// record, so the case reaches a reader through the checker rather than through
// a position in this list.
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
	entry := Entry{Slug: slug, State: stateNoRecord, QuestionWritten: stateNoRecord, NeedsHardware: stateNoRecord}

	data, err := os.ReadFile(filepath.Join(dir, RecordName))
	if err != nil {
		return entry
	}
	record, err := ParseRecord(data)
	if err != nil {
		entry.State = stateUnreadable
		entry.QuestionWritten = stateUnreadable
		entry.NeedsHardware = stateUnreadable
		return entry
	}

	entry.NeedsHardware = hardwareNotDeclared
	if needs, present := record.Field(FieldNeedsHardware); present {
		entry.NeedsHardware = needs
		if strings.TrimSpace(needs) == "" {
			entry.NeedsHardware = stateNotWritten
		}
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
// The columns are the ones the verb exists for: the slug, the state, the date
// the question was written, how long anything still asking has been there, and
// what the record says it needs beyond the runner. Nothing else, because a
// listing wide enough to wrap stops being scannable and scanning it is the
// whole point.
//
// The last column is here because the reader this verb is written for is
// deciding what to do about an experiment, and whether they could reproduce it
// on the machine in front of them is part of that. Record 0015 adds the field
// and this prints what it says without judging it.
func (l Listing) Report(now time.Time) string {
	out := fmt.Sprintf("examined %s\n", l.Root)
	if !l.ExperimentsPresent {
		out += fmt.Sprintf("no %s directory in this tree\n", ExperimentsDir)
	}
	out += fmt.Sprintf("%d %s\n", len(l.Entries), plural(len(l.Entries), "experiment", "experiments"))
	out += fmt.Sprintf("the time this run read is %s\n", now.UTC().Format(time.RFC3339))
	if len(l.Entries) == 0 {
		return out
	}

	rows := [][5]string{{"slug", "state", "question written", "waiting", "needs"}}
	for _, entry := range l.Entries {
		waiting := "-"
		if days, counted := entry.Waiting(now); counted {
			waiting = fmt.Sprintf("%d %s", days, plural(days, "day", "days"))
		}
		rows = append(rows, [5]string{entry.Slug, entry.State, entry.QuestionWritten, waiting, entry.NeedsHardware})
	}

	var width [5]int
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
