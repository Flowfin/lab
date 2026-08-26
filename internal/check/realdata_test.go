package check

import (
	"strings"
	"testing"
)

// The real-data refusal is proved by cases under testdata/cases, and the
// harness compares sets of property identifiers. These tests hold the two
// things a case cannot: what the message sends a reader to repair, and the
// three shapes that are deliberately not refused.

// TestAnEmptyRealDataDeclarationNamesWhatIsMissing holds the message to naming
// all three parts of the declaration. A record whose author left the value for
// later has to be told what the value is for, and the field name alone does
// not say it.
func TestAnEmptyRealDataDeclarationNamesWhatIsMissing(t *testing.T) {
	for _, c := range []struct {
		name   string
		record string
	}{
		{
			name:   "nothing after the colon",
			record: "Slug: one\nState: asking\nReal-Data:\n\n## Answer\n",
		},
		{
			name:   "a value that is only spaces",
			record: "Slug: one\nState: asking\nReal-Data:   \n\n## Answer\n",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			refusals := refuseRealData("experiments/one/EXPERIMENT.md", []byte(c.record))
			if len(refusals) != 1 {
				t.Fatalf("got %d refusals, want exactly 1: %v", len(refusals), refusals)
			}
			if refusals[0].Property != RecordRealDataDeclarationIsEmpty {
				t.Fatalf("refused %s, want %s", refusals[0].Property, RecordRealDataDeclarationIsEmpty)
			}
			for _, part := range []string{"no data", "no host", "written down", RealDataNone} {
				if !strings.Contains(refusals[0].Detail, part) {
					t.Errorf("the message is %q and does not say %q, so it names less than the declaration owes", refusals[0].Detail, part)
				}
			}
		})
	}
}

// TestARealDataDeclarationThatSaysSomethingRefusesNothing is the near
// neighbour at this level. Without it a check that refused every declaration
// would pass the test above.
func TestARealDataDeclarationThatSaysSomethingRefusesNothing(t *testing.T) {
	for _, declared := range []string{
		RealDataNone,
		"my own media library, on my own machine, and the scan time is what gets written down",
	} {
		t.Run(declared, func(t *testing.T) {
			record := "Slug: one\nState: asking\n" + FieldRealData + ": " + declared + "\n\n## Answer\n"
			if refusals := refuseRealData("experiments/one/EXPERIMENT.md", []byte(record)); len(refusals) != 0 {
				t.Errorf("a record declaring %q refused %v", declared, refusals)
			}
		})
	}
}

// TestAnAbsentRealDataFieldIsNotJudged holds open the hole record 0025 accepts
// by name. Record 0013 makes absence legal for every field added after it, so
// an experiment that touches real data and declares nothing passes here. This
// test is what stops that being closed by accident rather than by a record
// that argues for closing it.
func TestAnAbsentRealDataFieldIsNotJudged(t *testing.T) {
	record := "Slug: one\nState: asking\nQuestion-Written: 2026-01-01\n\n## Answer\n"
	if refusals := refuseRealData("experiments/one/EXPERIMENT.md", []byte(record)); len(refusals) != 0 {
		t.Errorf("a record declaring no %s refused %v", FieldRealData, refusals)
	}
}

// TestARecordWithNoHeaderIsNotJudgedOnItsRealData is the same hole every other
// header rule holds open. Nothing can read a field out of a file that has no
// header, and refusing one here would name the wrong repair.
func TestARecordWithNoHeaderIsNotJudgedOnItsRealData(t *testing.T) {
	if refusals := refuseRealData("experiments/one/EXPERIMENT.md", []byte("# One\n\nThe question.\n")); len(refusals) != 0 {
		t.Errorf("a file with no header was judged on its real-data declaration and refused %v", refusals)
	}
}
