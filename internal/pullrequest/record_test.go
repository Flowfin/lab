package pullrequest

import (
	"strings"
	"testing"
)

// theQuestion is what the work in every case below began from.
const theQuestion = "Does the walk cost more than a second on a tree of a thousand records?"

// recordAt builds a record in the shape docs/experiment-template.md carries,
// with the state, the answer and where a case needs it the question. Building
// it from one function is what keeps every case a near miss: two records in a
// case differ in one section and in nothing else, so a refusal is about that
// section.
func recordAt(state, answer string, options ...func(*recordFixture)) []byte {
	fixture := recordFixture{question: theQuestion}
	for _, option := range options {
		option(&fixture)
	}
	return []byte("Slug: one\n" +
		"State: " + state + "\n" +
		"Question-Written: 2026-08-01\n" +
		"Needs-Hardware: none\n" +
		"\n" +
		"## Question\n" +
		"\n" +
		fixture.question + "\n" +
		"\n" +
		"## Method\n" +
		"\n" +
		"Built the tree and timed the walk.\n" +
		"\n" +
		"## Answer\n" +
		"\n" +
		answer + "\n")
}

// A recordFixture is what a case changed about the record it built.
type recordFixture struct {
	question string
}

// withQuestion replaces what the record asks.
func withQuestion(question string) func(*recordFixture) {
	return func(fixture *recordFixture) { fixture.question = question }
}

// theLandedAnswer is what every case below starts from at the base of the
// range.
const theLandedAnswer = "No. The walk took eleven seconds, and the cost is in\n" +
	"reading the records rather than in walking the tree."

// recordJudgeCases are the cases over the rule that an answer already landed
// may grow and may not change. They are returned into the one table so that the
// proof that every property has a case reads them too.
func recordJudgeCases() []judgeCase {
	landed := recordAt("answered", theLandedAnswer)

	withRecords := func(records ...RecordChange) Change {
		c := clean()
		c.Files = []File{{Path: "experiments/one/EXPERIMENT.md"}}
		c.Records = records
		return c
	}

	return []judgeCase{
		{
			name: "an answer already landed, with a correction added underneath",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        landed,
				BeforePresent: true,
				After: recordAt("answered", theLandedAnswer+"\n\n"+
					"Corrected on 2026-09-01. The eleven seconds were a cold cache,\n"+
					"and a warm one takes two. Found by running it twice."),
				AfterPresent: true,
			}),
		},
		{
			name: "an answer already landed, with one line altered",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        landed,
				BeforePresent: true,
				After: recordAt("answered", "No. The walk took two seconds, and the cost is in\n"+
					"reading the records rather than in walking the tree."),
				AfterPresent: true,
			}),
			// The whole failure in one line. The record now says something
			// nobody measured on the day it was written, and every other check
			// here stays green through it.
			want: []string{AnswerAlreadyLandedWasRewritten},
		},
		{
			name: "an answer already landed, with one line removed",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        landed,
				BeforePresent: true,
				After:         recordAt("answered", "No. The walk took eleven seconds, and the cost is in"),
				AfterPresent:  true,
			}),
			want: []string{AnswerAlreadyLandedWasRewritten},
		},
		{
			name: "an answer already landed, with its blank lines moved",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        landed,
				BeforePresent: true,
				After:         recordAt("answered", "\n"+theLandedAnswer+"\n"),
				AfterPresent:  true,
			}),
			// A blank line carries no words. Refusing this would be a rule
			// about layout wearing the name of a rule about honesty.
		},
		{
			name: "an answer written for the first time",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        recordAt("asking", ""),
				BeforePresent: true,
				After:         recordAt("answered", theLandedAnswer),
				AfterPresent:  true,
			}),
			// The boundary the issue asks for at the check. A rule that made
			// the first draft permanent would teach people to commit nothing
			// until they were certain, which is how an experiment ends up with
			// no written answer at all.
		},
		{
			name: "a draft answer redrafted while the record is still asking",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        recordAt("asking", "So far it looks like eleven seconds, and that is not measured yet."),
				BeforePresent: true,
				After:         recordAt("answered", theLandedAnswer),
				AfterPresent:  true,
			}),
			// The boundary, in the shape that proves it. A record still asking
			// carries whatever its author is working on, and holding a draft
			// to the rule for a landed answer would teach people to write
			// nothing down until they were certain. Without the state at the
			// base being read, this is refused.
		},
		{
			name: "a record this change creates",
			change: withRecords(RecordChange{
				Path:         "experiments/one/EXPERIMENT.md",
				After:        recordAt("asking", ""),
				AfterPresent: true,
			}),
		},
		{
			name: "an answered record whose answer heading is gone",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        landed,
				BeforePresent: true,
				After: []byte("Slug: one\nState: answered\nQuestion-Written: 2026-08-01\n\n" +
					"## Question\n\n" +
					"Does the walk cost more than a second on a tree of a thousand records?\n"),
				AfterPresent: true,
			}),
			// The question is carried across unchanged, so this case refuses
			// the one rule it is about. A fixture that tripped both would
			// prove neither.
			want: []string{AnswerAlreadyLandedWasRewritten},
		},
		{
			name: "an answered record that no longer parses as a record",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        landed,
				BeforePresent: true,
				After:         []byte("Just prose now, with no header at all.\n"),
				AfterPresent:  true,
			}),
			// The evasion worth closing. Removing the header takes every
			// section out of reach of every rule that reads one, and the
			// checker walking the tree would report a record it cannot read
			// without ever saying that words went missing inside it. Both
			// rules refuse it, because both of their sections went out of
			// reach in the same edit.
			want: []string{AnswerAlreadyLandedWasRewritten, QuestionAlreadyAskedWasRewritten},
		},
		{
			name: "a question altered after the work started",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        landed,
				BeforePresent: true,
				After: recordAt("answered", theLandedAnswer, withQuestion(
					"Where does the cost of the walk sit on a tree of a thousand records?")),
				AfterPresent: true,
			}),
			// The edit this rule exists for. The measurement came back saying
			// something adjacent and more interesting, and rewriting the
			// question to name it turns a result nobody predicted into a
			// result the record claims was the point.
			want: []string{QuestionAlreadyAskedWasRewritten},
		},
		{
			name: "a question clarified underneath, while it is still asking",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        recordAt("asking", ""),
				BeforePresent: true,
				After: recordAt("asking", "", withQuestion(
					theQuestion+"\n\nAdded after the work started: a record here means an\n"+
						"EXPERIMENT.md and not a directory.")),
				AfterPresent: true,
			}),
		},
		{
			name: "a question written for the first time",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        []byte("Slug: one\nState: asking\nQuestion-Written: 2026-08-01\n\n## Answer\n\n"),
				BeforePresent: true,
				After:         recordAt("asking", ""),
				AfterPresent:  true,
			}),
			// The boundary this rule asks for. A record whose question section
			// was not there at the base is not covered, because making a first
			// draft permanent would teach people to commit nothing until they
			// were sure, which is how an experiment ends up with no written
			// question at all.
		},
		{
			name: "a question section that was there and empty at the base",
			change: withRecords(RecordChange{
				Path:          "experiments/one/EXPERIMENT.md",
				Before:        recordAt("asking", "", withQuestion("")),
				BeforePresent: true,
				After:         recordAt("asking", ""),
				AfterPresent:  true,
			}),
			// A heading with nothing under it carries nothing to lose. That it
			// states no question at all is a refusal the checker already makes
			// when it walks the tree, and repeating the judgement here would
			// send whoever hit it to the wrong repair.
		},
	}
}

// TestARewrittenAnswerNamesTheRecordAndTheSection holds the refusal to carrying
// both things whoever hit it needs. A message saying an answer changed, without
// saying in which record or which section, sends a reader to search a diff.
func TestARewrittenAnswerNamesTheRecordAndTheSection(t *testing.T) {
	change := clean()
	change.Files = []File{{Path: "experiments/one/EXPERIMENT.md"}}
	change.Records = []RecordChange{{
		Path:          "experiments/one/EXPERIMENT.md",
		Before:        recordAt("answered", theLandedAnswer),
		BeforePresent: true,
		After:         recordAt("answered", "Yes, and it was quick."),
		AfterPresent:  true,
	}}

	verdict := Judge(change)
	if len(verdict.Refusals) != 1 {
		t.Fatalf("%d refusals, want 1: %v", len(verdict.Refusals), verdict.Refusals)
	}
	refusal := verdict.Refusals[0]
	if refusal.Subject != "experiments/one/EXPERIMENT.md" {
		t.Errorf("the refusal names %q as its subject", refusal.Subject)
	}
	if !strings.Contains(refusal.Detail, "Answer") {
		t.Errorf("the refusal does not name the section: %q", refusal.Detail)
	}
	if !strings.Contains(refusal.Detail, "No. The walk took eleven seconds") {
		t.Errorf("the refusal does not quote the line that went missing: %q", refusal.Detail)
	}
}

// TestARewrittenQuestionNamesTheRecordAndTheSection is the same obligation for
// the other rule. The two refusals are produced by one function, and the fixture
// here is what says so rather than a comment claiming it.
func TestARewrittenQuestionNamesTheRecordAndTheSection(t *testing.T) {
	change := clean()
	change.Files = []File{{Path: "experiments/one/EXPERIMENT.md"}}
	change.Records = []RecordChange{{
		Path:          "experiments/one/EXPERIMENT.md",
		Before:        recordAt("asking", ""),
		BeforePresent: true,
		After:         recordAt("asking", "", withQuestion("Where does the cost of the walk sit?")),
		AfterPresent:  true,
	}}

	verdict := Judge(change)
	if len(verdict.Refusals) != 1 {
		t.Fatalf("%d refusals, want 1: %v", len(verdict.Refusals), verdict.Refusals)
	}
	refusal := verdict.Refusals[0]
	if refusal.Property != QuestionAlreadyAskedWasRewritten {
		t.Errorf("the refusal is %s", refusal.Property)
	}
	if refusal.Subject != "experiments/one/EXPERIMENT.md" {
		t.Errorf("the refusal names %q as its subject", refusal.Subject)
	}
	if !strings.Contains(refusal.Detail, "Question") {
		t.Errorf("the refusal does not name the section: %q", refusal.Detail)
	}
	if !strings.Contains(refusal.Detail, "Does the walk cost more than a second") {
		t.Errorf("the refusal does not quote the words the work began from: %q", refusal.Detail)
	}
}

// TestALineIsFoundLostWhereverItWas proves the comparison the rule rests on,
// over the shapes the cases do not reach. It is a subsequence rather than a
// prefix, because where an addition sits is the author's judgement and only
// what was removed or altered is this rule's business.
func TestALineIsFoundLostWhereverItWas(t *testing.T) {
	tests := []struct {
		name            string
		landed, current string
		lost            bool
	}{
		{name: "nothing landed", current: "anything"},
		{name: "unchanged", landed: "one\ntwo", current: "one\ntwo"},
		{name: "added underneath", landed: "one\ntwo", current: "one\ntwo\nthree"},
		{name: "added in the middle", landed: "one\ntwo", current: "one\nand a half\ntwo"},
		{name: "added above", landed: "one\ntwo", current: "nought\none\ntwo"},
		{name: "reordered", landed: "one\ntwo", current: "two\none", lost: true},
		{name: "one word changed", landed: "one\ntwo", current: "one\ntoo", lost: true},
		{name: "one line removed", landed: "one\ntwo", current: "one", lost: true},
		{name: "everything removed", landed: "one\ntwo", current: "", lost: true},
		{name: "reflowed onto one line", landed: "one\ntwo", current: "one two", lost: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, lost := firstLineLost(tc.landed, tc.current)
			if lost != tc.lost {
				t.Fatalf("a line lost is %v, want %v", lost, tc.lost)
			}
		})
	}
}
