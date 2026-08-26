package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// listingsDir holds the trees the listing is read against. They are separate
// from the cases under testdata/cases because a case declares which refusals a
// tree produces and a listing declares nothing of the kind: it is a report, and
// the thing worth asserting about it is what it printed and in what order.
const listingsDir = "../../testdata/listings"

// listingNow is the day every assertion here is made from. It is the same value
// the rest of the suite walks with, because the runner has one notion of now and
// a suite with two would prove nothing about that. The clock is a parameter for
// exactly this reason: a test that computed the expected number the same way the
// listing does would assert nothing, and one that hard-coded a number against
// the real clock would be red tomorrow.
var listingNow = fixedNow

// TestTheListingSortsTheOldestUnansweredFirst is the ordering the verb exists
// for. A record sitting in asking with a real question breaks no rule, so
// nothing refuses it and the only thing that surfaces it is where it appears in
// this list.
func TestTheListingSortsTheOldestUnansweredFirst(t *testing.T) {
	listing := read(t, "several-experiments")

	want := []string{
		// Asking, oldest question first. These two are the whole point.
		"oldest-question",
		"newer-question",
		// Then everything else, oldest first.
		"given-up",
		"already-answered",
		// Then whatever carries no date to sort by, ordered by slug so the
		// answer does not depend on the filesystem.
		"no-header",
		"no-record-at-all",
	}

	if len(listing.Entries) != len(want) {
		t.Fatalf("listed %d experiments, want %d", len(listing.Entries), len(want))
	}
	for i, slug := range want {
		if listing.Entries[i].Slug != slug {
			t.Errorf("position %d is %s, want %s", i, listing.Entries[i].Slug, slug)
		}
	}
}

// TestTheWaitingColumnIsCountedFromTheClockItWasGiven asserts the column that
// carries the whole point of the listing against a fixed value, which is only
// possible because the time arrives as a parameter.
func TestTheWaitingColumnIsCountedFromTheClockItWasGiven(t *testing.T) {
	tests := []struct {
		slug    string
		days    int
		counted bool
	}{
		{slug: "oldest-question", days: 137, counted: true},
		{slug: "newer-question", days: 62, counted: true},
		// Nothing is waiting once it has stopped, whichever way it stopped.
		{slug: "already-answered", counted: false},
		{slug: "given-up", counted: false},
		// Nothing to count from, and the column says so rather than guessing.
		{slug: "no-header", counted: false},
		{slug: "no-record-at-all", counted: false},
	}

	entries := make(map[string]Entry)
	for _, entry := range read(t, "several-experiments").Entries {
		entries[entry.Slug] = entry
	}

	for _, tc := range tests {
		t.Run(tc.slug, func(t *testing.T) {
			entry, listed := entries[tc.slug]
			if !listed {
				t.Fatalf("%s is not in the listing", tc.slug)
			}
			days, counted := entry.Waiting(listingNow)
			if counted != tc.counted {
				t.Fatalf("waiting is counted %v, want %v", counted, tc.counted)
			}
			if counted && days != tc.days {
				t.Fatalf("waiting is %d days, want %d", days, tc.days)
			}
		})
	}
}

// TestTheListingReadsWhatItCanAndSaysWhatItCannot holds the half that decides
// whether this verb is useful on a tree that is not in order. A listing that
// stopped at the first unreadable record, or dropped the directory from the
// output, would hide exactly the experiment somebody is looking for.
func TestTheListingReadsWhatItCanAndSaysWhatItCannot(t *testing.T) {
	states := map[string]string{
		"oldest-question":  StateAsking,
		"already-answered": StateAnswered,
		"given-up":         StateAbandoned,
		"no-header":        stateUnreadable,
		"no-record-at-all": stateNoRecord,
	}

	for _, entry := range read(t, "several-experiments").Entries {
		want, named := states[entry.Slug]
		if !named {
			continue
		}
		if entry.State != want {
			t.Errorf("%s is listed as %q, want %q", entry.Slug, entry.State, want)
		}
	}
}

// TestTheHardwareColumnKeepsSilenceApartFromAClaim is the half of the column
// that is easy to lose. A record written before the field existed declares
// nothing, and a record saying none declares something a tree can contradict.
// Printing both as an empty cell, or both as none, would report a silence as a
// claim, which is the one thing this column must not do.
func TestTheHardwareColumnKeepsSilenceApartFromAClaim(t *testing.T) {
	want := map[string]string{
		"oldest-question":  "a spinning disk",
		"newer-question":   HardwareNone,
		"already-answered": hardwareNotDeclared,
		"given-up":         hardwareNotDeclared,
		"no-header":        stateUnreadable,
		"no-record-at-all": stateNoRecord,
	}

	for _, entry := range read(t, "several-experiments").Entries {
		if entry.NeedsHardware != want[entry.Slug] {
			t.Errorf("%s declares %q, want %q", entry.Slug, entry.NeedsHardware, want[entry.Slug])
		}
	}
}

// TestTheReportPrintsEveryColumn compares the whole report against the
// output stored beside the tree. Comparing the whole thing is what makes it
// mean something: an assertion that a slug appears somewhere would pass on a
// listing whose columns had silently swapped.
func TestTheReportPrintsEveryColumn(t *testing.T) {
	name := "several-experiments"
	got := read(t, name).Report(listingNow)

	data, err := os.ReadFile(filepath.Join(listingsDir, name, "expected-report"))
	if err != nil {
		t.Fatalf("cannot read the expected report: %v", err)
	}
	if got != string(data) {
		t.Fatalf("the report is\n%s\nand the case expects\n%s", got, data)
	}
}

// TestAListingOfATreeWithNoExperimentsSaysSo keeps the two statements apart. A
// tree with no experiments directory and one whose directory is empty both list
// nothing, and only one of them says the tree has no such directory.
func TestAListingOfATreeWithNoExperimentsSaysSo(t *testing.T) {
	listing, err := List(filepath.Join(casesDir, "no-experiments-directory", "tree"), listingNow)
	if err != nil {
		t.Fatalf("the listing failed: %v", err)
	}
	if listing.ExperimentsPresent {
		t.Error("the listing says the tree holds an experiments directory")
	}
	if len(listing.Entries) != 0 {
		t.Errorf("the listing holds %d entries, want none", len(listing.Entries))
	}
}

// read reads one listing tree, failing rather than skipping when it is not
// there, because a test that quietly ran against nothing is a green test that
// proves nothing.
func read(t *testing.T, name string) Listing {
	t.Helper()

	// Forward slashes rather than filepath.Join, because the report prints the
	// root exactly as it was given and the expected report is one file read on
	// every platform the suite runs on.
	listing, err := List(listingsDir+"/"+name+"/tree", listingNow)
	if err != nil {
		t.Fatalf("listing %s failed: %v", name, err)
	}
	return listing
}

// TestTheListingKeepsAHoldApartFromASilence is the half of this field that is
// easy to lose. Record 0013 makes a field added after it optional, so a record
// declaring nothing is the ordinary case and a record declaring a hold is the
// exception, and reading a missing field as an empty one would report every
// record in the tree as held back since nothing.
func TestTheListingKeepsAHoldApartFromASilence(t *testing.T) {
	holding := map[string]bool{
		"still-asking":            true,
		"unreadable-hold":         true,
		"finished-but-still-held": true,
		"no-hold":                 false,
	}

	for _, entry := range read(t, "held-back-experiments").Entries {
		want, named := holding[entry.Slug]
		if !named {
			t.Fatalf("%s is in the listing and this test does not name it", entry.Slug)
		}
		if entry.Holding != want {
			t.Errorf("%s is held back %v, want %v", entry.Slug, entry.Holding, want)
		}
	}
}

// TestAHoldIsDatedOnlyWhereTheRecordWroteADate keeps the parsed date apart from
// the value as it was written. The checker refuses a held-back field that is
// not a date, and the listing prints one, so a listing that silently read an
// unparseable value as the zero time would print the first day of year one as
// the moment the clock started.
func TestAHoldIsDatedOnlyWhereTheRecordWroteADate(t *testing.T) {
	tests := []struct {
		slug    string
		written string
		dated   bool
		started string
	}{
		{slug: "still-asking", written: "2026-06-01", dated: true, started: "2026-06-01"},
		{slug: "finished-but-still-held", written: "2026-04-05", dated: true, started: "2026-04-05"},
		{slug: "unreadable-hold", written: "some time in June", dated: false},
		{slug: "no-hold", written: "", dated: false},
	}

	entries := make(map[string]Entry)
	for _, entry := range read(t, "held-back-experiments").Entries {
		entries[entry.Slug] = entry
	}

	for _, tc := range tests {
		t.Run(tc.slug, func(t *testing.T) {
			entry, listed := entries[tc.slug]
			if !listed {
				t.Fatalf("%s is not in the listing", tc.slug)
			}
			if entry.HeldBack != tc.written {
				t.Errorf("its %s reads %q, want %q", FieldHeldBack, entry.HeldBack, tc.written)
			}
			if entry.HeldBackDated != tc.dated {
				t.Fatalf("the hold is dated %v, want %v", entry.HeldBackDated, tc.dated)
			}
			if tc.dated && entry.Started.Format(DateFormat) != tc.started {
				t.Errorf("the clock started %s, want %s", entry.Started.Format(DateFormat), tc.started)
			}
		})
	}
}

// TestTheReportPrintsTheHoldsAndTheirDates compares the whole report against
// the one stored beside the tree, for the reason TestTheReportPrintsEveryColumn
// gives: an assertion that a slug appears somewhere passes on a report whose
// lines have silently swapped their dates.
func TestTheReportPrintsTheHoldsAndTheirDates(t *testing.T) {
	name := "held-back-experiments"
	got := read(t, name).Report(listingNow)

	data, err := os.ReadFile(filepath.Join(listingsDir, name, "expected-report"))
	if err != nil {
		t.Fatalf("cannot read the expected report: %v", err)
	}
	if got != string(data) {
		t.Fatalf("the report is\n%s\nand the case expects\n%s", got, data)
	}
}

// TestTheReportCountsHoldsWhenThereAreNone is the line a reader needs to tell a
// tree with no hold from a listing that never looked for one. It is asserted on
// its own because it is the case almost every run of this verb produces, and a
// count printed only when it is not zero teaches nobody that it is printed.
func TestTheReportCountsHoldsWhenThereAreNone(t *testing.T) {
	report := read(t, "several-experiments").Report(listingNow)
	if !strings.Contains(report, "0 records are held back\n") {
		t.Errorf("the report does not count the holds it did not find:\n%s", report)
	}
}

// TestAListingOfATreeWithNoExperimentsStillCountsTheHolds holds the same line
// on the path that returns before the table is built. A count that disappears
// with the table is a count a reader cannot rely on being there.
func TestAListingOfATreeWithNoExperimentsStillCountsTheHolds(t *testing.T) {
	listing, err := List(filepath.Join(casesDir, "no-experiments-directory", "tree"), listingNow)
	if err != nil {
		t.Fatalf("the listing failed: %v", err)
	}
	if !strings.Contains(listing.Report(listingNow), "0 records are held back\n") {
		t.Errorf("the report does not count the holds it did not find:\n%s", listing.Report(listingNow))
	}
}
