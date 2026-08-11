package pullrequest

import (
	"strings"
	"testing"
)

// recordAt builds a record in the shape docs/experiment-template.md carries,
// with the state and the answer a case needs. Building it from one function is
// what keeps every case below a near miss: two records in a case differ in the
// answer and in nothing else, so a refusal is about the answer.
func recordAt(state, answer string) []byte {
	return []byte("Slug: one\n" +
		"State: " + state + "\n" +
		"Question-Written: 2026-08-01\n" +
		"Needs-Hardware: none\n" +
		"\n" +
		"## Question\n" +
		"\n" +
		"Does the walk cost more than a second on a tree of a thousand records?\n" +
		"\n" +
		"## Method\n" +
		"\n" +
		"Built the tree and timed the walk.\n" +
		"\n" +
		"## Answer\n" +
		"\n" +
		answer + "\n")
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
					"## Question\n\nDoes it?\n"),
				AfterPresent: true,
			}),
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
			// The evasion worth closing. Removing the header takes the answer
			// out of reach of every rule that reads a section, and the checker
			// walking the tree would report a record it cannot read without
			// ever saying an answer went missing inside it.
			want: []string{AnswerAlreadyLandedWasRewritten},
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
