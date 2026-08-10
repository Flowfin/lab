package check

import (
	"strings"
	"testing"
)

// The question refusal is proved by two cases under testdata/cases, and the
// harness compares sets of property identifiers. That is not enough here, for
// the reason state_test.go already sets out about the state refusal: the
// property is refused at two sites, and a fixture for either satisfies the
// standing requirement for both.
//
// The two are not interchangeable to the person reading the refusal. A record
// with no question heading is somebody who has not started writing it. A
// record with the heading and nothing under it is somebody who copied the
// template and stopped, and the repair is a sentence rather than a section.
// The message is the only thing that separates them.
//
// Measured rather than supposed. Deleting the absent arm leaves both cases
// green, because a section that is not there reads back as an empty body and
// falls through to the arm below, which refuses the same property under the
// wrong message. That is the bound in harness_test.go acted on, and this test
// is what acts on it.

// TestEachQuestionSiteNamesItsOwnRepair holds the two apart.
func TestEachQuestionSiteNamesItsOwnRepair(t *testing.T) {
	for _, c := range []struct {
		name   string
		record string
		wants  string
	}{
		{
			name:   "no question section at all",
			record: "Slug: one\nState: asking\nQuestion-Written: 2026-01-01\n\n## Method\n\nIt ran.\n\n## Answer\n",
			wants:  "carries no Question section at all",
		},
		{
			name:   "a question heading with nothing under it",
			record: "Slug: one\nState: asking\nQuestion-Written: 2026-01-01\n\n## Question\n\n## Method\n\nIt ran.\n\n## Answer\n",
			wants:  "Question heading is there with nothing under it",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			refusals := refuseQuestion("experiments/one/EXPERIMENT.md", []byte(c.record))
			if len(refusals) != 1 {
				t.Fatalf("refused %d times, want exactly one: %v", len(refusals), refusals)
			}
			if refusals[0].Property != RecordStatesNoQuestion {
				t.Fatalf("refused %s, want %s", refusals[0].Property, RecordStatesNoQuestion)
			}
			if !strings.Contains(refusals[0].Detail, c.wants) {
				t.Errorf("the message is %q and does not say %q, so it names the wrong repair", refusals[0].Detail, c.wants)
			}
		})
	}
}

// TestAQuestionOfOneCharacterIsEnough is the boundary written at the check,
// held to rather than described. Nothing here judges what the question says,
// so the smallest thing somebody can write under the heading passes, and a
// later change that started reading the text would fail this rather than
// arriving as a surprise refusal on somebody's honest record.
func TestAQuestionOfOneCharacterIsEnough(t *testing.T) {
	record := "Slug: one\nState: asking\nQuestion-Written: 2026-01-01\n\n## Question\n\n?\n\n## Answer\n"
	if refusals := refuseQuestion("experiments/one/EXPERIMENT.md", []byte(record)); len(refusals) != 0 {
		t.Errorf("refused %v, and this check judges whether something was written rather than what it says", refusals)
	}
}

// TestARecordThatIsNotARecordIsNotJudgedHere holds the disclosure at
// refuseQuestion to being true. A file with no header states no question and
// is not refused for it, because nothing can read a section out of a file that
// has no header and the repair a question refusal names would be the wrong
// one. If a later change closes that hole, this test is what says so out loud
// instead of the gap quietly changing shape.
func TestARecordThatIsNotARecordIsNotJudgedHere(t *testing.T) {
	record := "# Timing test\n\nSomething somebody meant to turn into a record.\n"
	if _, err := ParseRecord([]byte(record)); err == nil {
		t.Fatalf("this record parses, so it no longer stands for the case being described")
	}
	if refusals := refuseQuestion("experiments/one/EXPERIMENT.md", []byte(record)); len(refusals) != 0 {
		t.Errorf("refused %v, and the comment at refuseQuestion says this case is not reached", refusals)
	}
}
