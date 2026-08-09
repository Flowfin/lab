package check

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// listingsDir holds the trees the listing is read against. They are separate
// from the cases under testdata/cases because a case declares which refusals a
// tree produces and a listing declares nothing of the kind: it is a report, and
// the thing worth asserting about it is what it printed and in what order.
const listingsDir = "../../testdata/listings"

// listingNow is the day every assertion here is made from. The clock is a
// parameter for exactly this reason: a test that computed the expected number
// the same way the listing does would assert nothing, and one that hard-coded a
// number against the real clock would be red tomorrow.
var listingNow = time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)

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
		{slug: "oldest-question", days: 100, counted: true},
		{slug: "newer-question", days: 25, counted: true},
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

// TestTheReportPrintsTheFourColumns compares the whole report against the
// output stored beside the tree. Comparing the whole thing is what makes it
// mean something: an assertion that a slug appears somewhere would pass on a
// listing whose columns had silently swapped.
func TestTheReportPrintsTheFourColumns(t *testing.T) {
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
