package check

import (
	"strings"
	"testing"
	"time"
)

// The header date refusals are proved by cases under testdata/cases, and the
// harness compares sets of property identifiers. That is the whole of what a
// case proves, and it leaves two things open that these tests close.
//
// The date-shape rule reads two fields in one loop, so a case with a bad
// question date satisfies the standing requirement for a bad answer date as
// well, and a change that stopped reading one of the two would stay green.
//
// The answer-date rule is refused at two sites for two different repairs. A
// record still asking that carries an answer date has been edited towards a
// state it is not in. A record whose answer is dated before its question has a
// timeline that could not have happened. The message is the only thing that
// separates them, and it is what these hold.

// TestBothHeaderDateFieldsAreHeldToTheShape holds the loop to reading both
// fields rather than only the one a case happens to spend its fixture on.
func TestBothHeaderDateFieldsAreHeldToTheShape(t *testing.T) {
	for _, c := range []struct {
		name   string
		record string
		wants  string
	}{
		{
			name:   "a question date that is a month",
			record: "Slug: one\nState: asking\nQuestion-Written: March 2026\n\n## Question\n\nWhy?\n",
			wants:  `its Question-Written is "March 2026"`,
		},
		{
			name:   "an answer date that is a sentence",
			record: "Slug: one\nState: answered\nQuestion-Written: 2026-01-01\nAnswer-Written: some time last spring\n\n## Question\n\nWhy?\n\n## Answer\n\nNo.\n",
			wants:  `its Answer-Written is "some time last spring"`,
		},
		{
			name:   "a date field declared with nothing after the colon",
			record: "Slug: one\nState: asking\nQuestion-Written:\n\n## Question\n\nWhy?\n",
			wants:  `its Question-Written is ""`,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			refusals := refuseHeaderDates("experiments/one/EXPERIMENT.md", []byte(c.record))
			if len(refusals) != 1 {
				t.Fatalf("refused %d times, want exactly one: %v", len(refusals), refusals)
			}
			if refusals[0].Property != RecordHeaderDateIsNotADate {
				t.Fatalf("refused %s, want %s", refusals[0].Property, RecordHeaderDateIsNotADate)
			}
			if !strings.Contains(refusals[0].Detail, c.wants) {
				t.Errorf("the message is %q and does not say %q", refusals[0].Detail, c.wants)
			}
		})
	}
}

// TestEachAnswerDateSiteNamesItsOwnRepair holds the two apart.
func TestEachAnswerDateSiteNamesItsOwnRepair(t *testing.T) {
	for _, c := range []struct {
		name   string
		record string
		wants  string
	}{
		{
			name:   "an answer date on a record still asking",
			record: "Slug: one\nState: asking\nQuestion-Written: 2026-01-01\nAnswer-Written: 2026-02-02\n\n## Question\n\nWhy?\n",
			wants:  "edited towards a state it is not in",
		},
		{
			name:   "an answer dated before the question",
			record: "Slug: one\nState: answered\nQuestion-Written: 2026-03-01\nAnswer-Written: 2026-02-02\n\n## Question\n\nWhy?\n\n## Answer\n\nNo.\n",
			wants:  "the answer is dated before the question",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			refusals := refuseHeaderDates("experiments/one/EXPERIMENT.md", []byte(c.record))
			if len(refusals) != 1 {
				t.Fatalf("refused %d times, want exactly one: %v", len(refusals), refusals)
			}
			if refusals[0].Property != RecordAnswerDateDisagreesWithTheRecord {
				t.Fatalf("refused %s, want %s", refusals[0].Property, RecordAnswerDateDisagreesWithTheRecord)
			}
			if !strings.Contains(refusals[0].Detail, c.wants) {
				t.Errorf("the message is %q and does not say %q, so it names the wrong repair", refusals[0].Detail, c.wants)
			}
		})
	}
}

// TestTheHeaderDatesARecordInOrderCarriesAreLeftAlone holds the boundaries the
// rules are written with. An absent answer date is the ordinary state of every
// record that has not been answered, and record 0013 makes absence legal for
// every field, so neither of these may refuse.
func TestTheHeaderDatesARecordInOrderCarriesAreLeftAlone(t *testing.T) {
	for _, c := range []struct {
		name   string
		record string
	}{
		{
			name:   "asking, with no answer date",
			record: "Slug: one\nState: asking\nQuestion-Written: 2026-01-01\n\n## Question\n\nWhy?\n",
		},
		{
			name:   "answered on the day the question was written",
			record: "Slug: one\nState: answered\nQuestion-Written: 2026-01-01\nAnswer-Written: 2026-01-01\n\n## Question\n\nWhy?\n\n## Answer\n\nNo.\n",
		},
		{
			name:   "no dates in the header at all",
			record: "Slug: one\nState: asking\n\n## Question\n\nWhy?\n",
		},
		{
			name:   "abandoned, carrying the date the work stopped",
			record: "Slug: one\nState: abandoned\nQuestion-Written: 2026-01-01\nAnswer-Written: 2026-02-02\n\n## Question\n\nWhy?\n\n## Answer\n\nThe machine went back.\n",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			if refusals := refuseHeaderDates("experiments/one/EXPERIMENT.md", []byte(c.record)); len(refusals) != 0 {
				t.Errorf("refused %v, and this record is in order", refusals)
			}
		})
	}
}

// TestTheExampleInAMessageIsADate ties the human spelling a refusal prints to
// the layout the parser uses. Without it the two are separate strings that
// happen to agree today, and the one that goes wrong is the one nothing reads
// back.
func TestTheExampleInAMessageIsADate(t *testing.T) {
	if _, err := time.Parse(DateFormat, dateExample); err != nil {
		t.Fatalf("a refusal offers %q as an example of %s and the parser rejects it: %v", dateExample, dateShape, err)
	}
	if len(dateExample) != len(dateShape) {
		t.Errorf("the example %q and the shape %s are different lengths, so one of them is not the other written out", dateExample, dateShape)
	}
}
