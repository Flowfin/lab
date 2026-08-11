package pullrequest

import (
	"sort"
	"strings"
	"testing"
)

// A case is one pull request and the whole set of properties judging it must
// refuse. The set is compared rather than searched, for the reason the record
// checks already compare theirs: asking only whether the expected refusal is
// present passes a fixture that trips its own rule and one other, and a fixture
// that trips two rules proves neither of them.
type judgeCase struct {
	name   string
	change Change
	want   []string
	notes  []string
	skips  []string
}

// clean is a pull request that breaks nothing, which every refusing case below
// is one change away from. Building the cases from it is what makes each of
// them a near miss rather than a tree that trips several rules at once and
// proves nothing about which.
func clean() Change {
	return Change{
		Body:        "This closes #24 and nothing else.",
		BodyRead:    true,
		CommitsRead: true,
		Commits: []Commit{
			{
				Hash:    "1111111111111111111111111111111111111111",
				Message: "Add the deterministic pull-request check\n\nThe class the other checks miss is the pull request itself. Closes #24.\n",
			},
		},
		FilesRead:    true,
		Files:        []File{{Path: "internal/pullrequest/pullrequest.go"}},
		RecordsRead:  true,
		ChangedLines: 12,
		LinesCounted: true,
	}
}

func TestJudge(t *testing.T) {
	for _, tc := range judgeCases() {
		t.Run(tc.name, func(t *testing.T) {
			verdict := Judge(tc.change)

			for _, diff := range diffSets("refusal", tc.want, verdict.Properties()) {
				t.Error(diff)
			}

			var notes []string
			for _, note := range verdict.Notes {
				notes = append(notes, note.Property)
			}
			for _, diff := range diffSets("note", tc.notes, notes) {
				t.Error(diff)
			}

			var skips []string
			for _, skip := range verdict.Skips {
				skips = append(skips, skip.Rule)
			}
			for _, diff := range diffSets("skip", tc.skips, skips) {
				t.Error(diff)
			}
		})
	}
}

// judgeCases is the table both tests over it read, so a case removed from here
// is removed from the coverage proof below as well. A list of cases kept beside
// the cases is the shape where a fixture is deleted and the test that counted
// it goes on counting.
//
// The cases over a record at both ends of the range are in record_test.go, next
// to the rule they are about, and they are returned into this table rather than
// run from a second harness.
func judgeCases() []judgeCase {
	return append(recordJudgeCases(), []judgeCase{
		{
			name:   "a change that names its issue and touches no experiment",
			change: clean(),
		},
		{
			name: "a body that references no issue",
			change: func() Change {
				c := clean()
				c.Body = "This tidies up the check and makes the message clearer."
				return c
			}(),
			want: []string{BodyNamesNoIssue},
		},
		{
			name: "a body that names a number without the reference",
			change: func() Change {
				c := clean()
				c.Body = "This closes issue 24."
				return c
			}(),
			// The near miss worth the fixture. It reads as a reference to
			// somebody skimming and there is nothing in it a reader can
			// follow, which is the whole of what the rule is for.
			want: []string{BodyNamesNoIssue},
		},
		{
			name: "a body whose reference is a link",
			change: func() Change {
				c := clean()
				c.Body = "Closes https://github.com/Flowfin/lab/issues/24."
				return c
			}(),
		},
		{
			name: "a body whose link carries this host inside somebody else's path",
			change: func() Change {
				c := clean()
				c.Body = "See https://example.invalid/https://github.com/Flowfin/lab/issues/24 for the details."
				return c
			}(),
			// The near miss the link rule exists in this shape for. A reader
			// following it arrives somewhere nobody meant, and a pattern
			// carrying the host rather than comparing it would pass this.
			want: []string{BodyNamesNoIssue},
		},
		{
			name: "a body whose link is wrapped in brackets and ends in a full stop",
			change: func() Change {
				c := clean()
				c.Body = "The argument is there (https://github.com/Flowfin/lab/issues/24)."
				return c
			}(),
		},
		{
			name: "a body whose reference names the repository",
			change: func() Change {
				c := clean()
				c.Body = "Closes Flowfin/lab#24."
				return c
			}(),
		},
		{
			name: "a body carrying a colour rather than a reference",
			change: func() Change {
				c := clean()
				c.Body = "The swatch in the screenshot is #0f0f0f and nothing else changed."
				return c
			}(),
			// A number beginning with a zero is not an issue number, and this
			// is the string that would otherwise let a body satisfy the rule
			// by accident.
			want: []string{BodyNamesNoIssue},
		},
		{
			name: "a commit message that references no issue",
			change: func() Change {
				c := clean()
				c.Commits = append(c.Commits, Commit{
					Hash:    "2222222222222222222222222222222222222222",
					Message: "Fix the message\n\nIt read badly and now it does not.\n",
				})
				return c
			}(),
			want: []string{CommitMessageNamesNoIssue},
		},
		{
			name: "a commit naming its issue in the message and not in the subject",
			change: func() Change {
				c := clean()
				c.Commits = append(c.Commits, Commit{
					Hash:    "4444444444444444444444444444444444444444",
					Message: "Refuse a header whose dates disagree with themselves\n\nThe listing sorts on a field nothing parsed. Closes #54.\n\nSigned-off-by: A Contributor <nobody@example.invalid>\n",
				})
				return c
			}(),
			// The shape forty-nine landed commits are mostly written in, and
			// the reason the rule reads the whole message. A subject reading
			// would refuse this and the change explains itself perfectly well.
		},
		{
			name: "an experiment changed without its record",
			change: func() Change {
				c := clean()
				c.Files = []File{{Path: "experiments/one/measure.go"}}
				return c
			}(),
			want: []string{ExperimentChangedWithoutItsRecord},
		},
		{
			name: "an experiment changed with its record",
			change: func() Change {
				c := clean()
				c.Files = []File{
					{Path: "experiments/one/measure.go"},
					{Path: "experiments/one/EXPERIMENT.md"},
				}
				return c
			}(),
		},
		{
			name: "one experiment moving with its record and another without",
			change: func() Change {
				c := clean()
				c.Files = []File{
					{Path: "experiments/one/measure.go"},
					{Path: "experiments/one/EXPERIMENT.md"},
					{Path: "experiments/two/measure.go"},
				}
				return c
			}(),
			want: []string{ExperimentChangedWithoutItsRecord},
		},
		{
			name: "an experiment's file removed without its record",
			change: func() Change {
				c := clean()
				c.Files = []File{{Path: "experiments/one/measure.go", Gone: true}}
				return c
			}(),
			// Record 0004 lets an answered experiment's code be removed and
			// requires the record to gain a line naming the commit that
			// removed it. A rule reading only additions and edits would be
			// silent on exactly that change.
			want: []string{ExperimentChangedWithoutItsRecord},
		},
		{
			name: "a record's own change, with nothing else in the experiment",
			change: func() Change {
				c := clean()
				c.Files = []File{{Path: "experiments/one/EXPERIMENT.md"}}
				return c
			}(),
		},
		{
			name: "a file directly under experiments and in no experiment",
			change: func() Change {
				c := clean()
				c.Files = []File{{Path: "experiments/README.md"}}
				return c
			}(),
			// Whether such a file may be there at all is a rule about a tree,
			// and the checker holds it. This package sees paths and never the
			// directory they came from.
		},
		{
			name: "a file deep inside an experiment",
			change: func() Change {
				c := clean()
				c.Files = []File{{Path: "experiments/one/cmd/measure/main.go"}}
				return c
			}(),
			want: []string{ExperimentChangedWithoutItsRecord},
		},
		{
			name: "a change one line above the reading bound",
			change: func() Change {
				c := clean()
				c.ChangedLines = ReadingBound + 1
				return c
			}(),
			notes: []string{ChangeIsLargerThanOneReading},
		},
		{
			name: "a change exactly at the reading bound",
			change: func() Change {
				c := clean()
				c.ChangedLines = ReadingBound
				return c
			}(),
		},
		{
			name: "a change whose lines git could not count",
			change: func() Change {
				c := clean()
				c.Uncounted = []string{"docs/diagram.png"}
				return c
			}(),
			skips: []string{ChangeIsLargerThanOneReading},
		},
		{
			name:   "a run given nothing at all",
			change: Change{},
			// Every rule says it was not applied. Nothing is refused, because
			// a run that read no pull request has found no violation in one.
			skips: []string{
				BodyNamesNoIssue,
				CommitMessageNamesNoIssue,
				ExperimentChangedWithoutItsRecord,
				AnswerAlreadyLandedWasRewritten,
				QuestionAlreadyAskedWasRewritten,
				ChangeIsLargerThanOneReading,
			},
		},
	}...)
}

// TestEveryPropertyHasACaseThatRefusesIt refuses a property no case in the
// table trips. A rule with no fixture is a rule nobody has seen bite, and it is
// the quietest way this package could grow a branch that never runs while the
// suite stays green.
//
// It reads the same table the cases run from, and it asks the judgement rather
// than the declarations in it, so a case whose want list drifted from what it
// actually refuses is counted by what it refuses.
func TestEveryPropertyHasACaseThatRefusesIt(t *testing.T) {
	properties := []string{
		BodyNamesNoIssue,
		CommitMessageNamesNoIssue,
		ExperimentChangedWithoutItsRecord,
		AnswerAlreadyLandedWasRewritten,
		QuestionAlreadyAskedWasRewritten,
	}

	refused := make(map[string]bool)
	for _, tc := range judgeCases() {
		for _, property := range Judge(tc.change).Properties() {
			refused[property] = true
		}
	}
	for _, property := range properties {
		if !refused[property] {
			t.Errorf("no case refuses %s, so nothing in this suite has seen it bite", property)
		}
	}
}

// TestARefusalNamesItsSubject holds every refusal to carrying the thing whoever
// hit it has to go and open. A message saying a rule was broken and not saying
// where sends a reader to search the diff, which is what the check was supposed
// to save them.
func TestARefusalNamesItsSubject(t *testing.T) {
	change := Change{
		BodyRead:    true,
		Body:        "nothing here",
		CommitsRead: true,
		Commits:     []Commit{{Hash: "3333333333333333333333333333333333333333", Message: "no reference anywhere in this one"}},
		FilesRead:   true,
		Files:       []File{{Path: "experiments/one/measure.go"}},
	}

	verdict := Judge(change)
	if len(verdict.Refusals) != 3 {
		t.Fatalf("this change breaks three rules and %d were refused", len(verdict.Refusals))
	}
	for _, refusal := range verdict.Refusals {
		if strings.TrimSpace(refusal.Subject) == "" {
			t.Errorf("%s refused with no subject", refusal.Property)
		}
		if strings.TrimSpace(refusal.Detail) == "" {
			t.Errorf("%s refused with no detail", refusal.Property)
		}
	}
}

// TestTheReportSaysWhatItExamined is the proof that a run which judged nothing
// does not print what a run that judged everything prints. The two verdicts
// below are both clean, and a reader has to be able to tell them apart.
func TestTheReportSaysWhatItExamined(t *testing.T) {
	nothing := Change{}
	everything := clean()

	quiet := Judge(nothing).Report(nothing)
	full := Judge(everything).Report(everything)

	if quiet == full {
		t.Fatal("a run that read no pull request printed the same report as one that read a whole pull request")
	}
	for _, want := range []string{"not judged", "none was read", "none were read", "not counted"} {
		if !strings.Contains(quiet, want) {
			t.Errorf("the report of a run that read nothing does not say %q", want)
		}
	}
	if strings.Contains(full, "not judged:") {
		t.Error("the report of a run that read a whole pull request says a rule was not judged")
	}
	if !strings.Contains(full, "0 rule(s) not judged") {
		t.Error("the report of a run that read a whole pull request does not count the rules it did not judge")
	}
}

// TestTheSizeAnnotationNeverRefuses is the property the bound exists for. A
// size cap that reddened a build would be met by splitting a change into pieces
// chosen to satisfy a counter, so this holds the note to being a note however
// large the change gets.
func TestTheSizeAnnotationNeverRefuses(t *testing.T) {
	change := clean()
	change.ChangedLines = 100 * ReadingBound

	verdict := Judge(change)
	if len(verdict.Refusals) != 0 {
		t.Fatalf("a change of %d lines was refused: %v", change.ChangedLines, verdict.Refusals)
	}
	if len(verdict.Notes) != 1 {
		t.Fatalf("a change of %d lines produced %d notes and one was expected", change.ChangedLines, len(verdict.Notes))
	}
}

// TestSetsAreComparedAsSets proves the comparison every case above rests on
// bites, over the combinations the cases themselves do not reach.
func TestSetsAreComparedAsSets(t *testing.T) {
	tests := []struct {
		name      string
		want, got []string
		diffs     int
	}{
		{name: "nothing expected and nothing produced", diffs: 0},
		{name: "the expected one and one more", want: []string{"a"}, got: []string{"a", "b"}, diffs: 1},
		{name: "the expected one is missing", want: []string{"a"}, diffs: 1},
		{name: "one nobody expected", got: []string{"a"}, diffs: 1},
		{name: "the same set in another order", want: []string{"a", "b"}, got: []string{"b", "a"}, diffs: 0},
		{name: "one expected and a different one produced", want: []string{"a"}, got: []string{"b"}, diffs: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diffs := diffSets("refusal", tc.want, tc.got)
			if len(diffs) != tc.diffs {
				t.Fatalf("got %d differences %v, want %d", len(diffs), diffs, tc.diffs)
			}
		})
	}
}

// diffSets compares two sets of names and returns one line per difference, in
// both directions. A name expected and not produced and a name produced and not
// expected are different failures and are reported as different lines.
func diffSets(kind string, want, got []string) []string {
	expected := make(map[string]bool, len(want))
	for _, name := range want {
		expected[name] = true
	}
	produced := make(map[string]bool, len(got))
	for _, name := range got {
		produced[name] = true
	}

	var diffs []string
	for name := range expected {
		if !produced[name] {
			diffs = append(diffs, kind+" expected and not produced: "+name)
		}
	}
	for name := range produced {
		if !expected[name] {
			diffs = append(diffs, kind+" produced that no case expected: "+name)
		}
	}
	sort.Strings(diffs)
	return diffs
}
