package check

import (
	"strings"
	"testing"
)

// The three state refusals are proved by cases under testdata/cases, and the
// harness compares sets of property identifiers. That is the whole of what a
// case can prove, and it is not enough here, because the state property is
// refused at three sites and a fixture for one of them satisfies the standing
// requirement for all three.
//
// The three sites are not interchangeable. A record with no state field, a
// record whose state field is there and empty, and a record whose state is
// misspelled are three different repairs, and the message is the only thing
// that separates them. So the message is what these tests hold, at the
// function rather than through the walk, and the case harness holds that the
// property is refused at all.
//
// This is the bound in harness_test.go acted on rather than restated. Deleting
// the empty-state arm leaves every case green, because an empty value falls
// through to the arm below and refuses the same property under a different
// message. Measured, not supposed.

// TestEachStateSiteNamesItsOwnRepair holds the three apart.
func TestEachStateSiteNamesItsOwnRepair(t *testing.T) {
	for _, c := range []struct {
		name   string
		record string
		wants  string
	}{
		{
			name:   "no state field at all",
			record: "Slug: one\nQuestion-Written: 2026-01-01\n\n## Answer\n",
			wants:  "declares no State field",
		},
		{
			name:   "a state field with nothing after the colon",
			record: "Slug: one\nState:\n\n## Answer\n",
			wants:  "is there and empty",
		},
		{
			name:   "a state that is misspelled",
			record: "Slug: one\nState: answerd\n\n## Answer\n",
			wants:  `is "answerd"`,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			refusals := refuseState("experiments/one/EXPERIMENT.md", []byte(c.record))
			if len(refusals) != 1 {
				t.Fatalf("got %d refusals, want exactly 1: %v", len(refusals), refusals)
			}
			if refusals[0].Property != RecordStateIsNotOneOfTheThree {
				t.Fatalf("refused %s, want %s", refusals[0].Property, RecordStateIsNotOneOfTheThree)
			}
			if !strings.Contains(refusals[0].Detail, c.wants) {
				t.Errorf("the message is %q and does not say %q, so it names the wrong repair", refusals[0].Detail, c.wants)
			}
			// Every one of the three names the three states, because a
			// reader hitting this needs to know what the legal values are
			// without opening a decision record.
			if !strings.Contains(refusals[0].Detail, statesInWords()) {
				t.Errorf("the message is %q and does not name the three states", refusals[0].Detail)
			}
		})
	}
}

// TestAStateThatIsLegalRefusesNothingOnItsOwn is the near neighbour of the
// three above, at the same level. Without it, a site that refused every state
// would pass the test above.
func TestAStateThatIsLegalRefusesNothingOnItsOwn(t *testing.T) {
	for _, state := range []string{StateAsking, StateAnswered, StateAbandoned} {
		t.Run(state, func(t *testing.T) {
			record := "Slug: one\nState: " + state + "\n\n## Answer\n\nSomething is written here.\n"
			if refusals := refuseState("experiments/one/EXPERIMENT.md", []byte(record)); len(refusals) != 0 {
				t.Errorf("a record in state %s with an answer refused %v", state, refusals)
			}
		})
	}
}

// TestARecordWithNoHeaderIsNotJudgedHere holds the hole open deliberately.
// Nothing can read the state of something that has no header, and inventing a
// state refusal for it would name the wrong repair: whoever hits it has a file
// that is not a record rather than a record with a bad state. Issue #16 is
// where that gap closes, and this test is what stops it being closed here by
// accident.
func TestARecordWithNoHeaderIsNotJudgedHere(t *testing.T) {
	if refusals := refuseState("experiments/one/EXPERIMENT.md", []byte("# One\n\nThe question.\n")); len(refusals) != 0 {
		t.Errorf("a file with no header was judged on its state and refused %v", refusals)
	}
}
