package pullrequest

import (
	"strings"
	"testing"
)

// The cases over a record that was on the branch a change lands on and is not
// at the head of it. They are returned into the one table in pullrequest_test.go
// rather than run from a second harness, so the proof that every property has a
// case reads them too.
//
// Every case here is one change away from a case that refuses nothing, and the
// one change is always the same field. Whether the record had a version at the
// base is the whole of what separates a removal this rule is about from a
// directory somebody added and took out again while working.
func removalJudgeCases() []judgeCase {
	const record = "experiments/one/EXPERIMENT.md"
	const moved = "experiments/reading-a-tree-of-records/EXPERIMENT.md"
	landed := recordAt("answered", theLandedAnswer)

	return []judgeCase{
		{
			name: "a record already on the branch that this change removes",
			change: func() Change {
				c := clean()
				c.Body = "This closes #24. It removes " + record + " and experiments/one/measure.go."
				c.Files = []File{
					{Path: record, Gone: true},
					{Path: "experiments/one/measure.go", Gone: true},
				}
				c.Records = []RecordChange{{
					Path:          record,
					Before:        landed,
					BeforePresent: true,
				}}
				return c
			}(),
			// The ordinary shape rather than the malicious one. The prototype
			// went under record 0004, what was left looked like an empty
			// directory, and tidying it took the record with it.
			want: []string{RecordAlreadyLandedWasRemoved},
		},
		{
			name: "an experiment directory already on the branch that this change renames",
			change: func() Change {
				c := clean()
				c.Files = []File{
					{Path: record, Gone: true},
					{Path: moved, From: record},
				}
				c.Records = []RecordChange{
					{Path: record, Before: landed, BeforePresent: true},
					{Path: moved, After: landed, AfterPresent: true},
				}
				return c
			}(),
			// The record is still in the tree and every pointer at where it
			// was has stopped resolving, including a promotion section a
			// reader is following from another board.
			want: []string{ExperimentAlreadyLandedWasRenamed},
		},
		{
			name: "a record this change removes that was never on the branch it lands on",
			change: func() Change {
				c := clean()
				c.Body = "This closes #24. It removes " + record + ", which it also added."
				c.Files = []File{{Path: record, Gone: true}}
				c.Records = []RecordChange{{Path: record}}
				return c
			}(),
			// The boundary the issue asks for at the check, held in the field
			// the rule actually reads. Over a range read from base to head a
			// record added and taken out again inside one branch is not in the
			// diff at all and never reaches this rule; this case says the rule
			// answers the same way where it does reach it, so the boundary does
			// not rest on how the range was read.
		},
		{
			name: "an experiment's code removed with the record kept",
			change: func() Change {
				c := clean()
				c.Body = "This closes #24. It removes experiments/one/measure.go under record 0004."
				c.Files = []File{
					{Path: "experiments/one/measure.go", Gone: true},
					{Path: record},
				}
				c.Records = []RecordChange{{
					Path:          record,
					Before:        landed,
					BeforePresent: true,
					After: recordAt("answered", theLandedAnswer+"\n\n"+
						"The prototype was removed in 1111111111111111111111111111111111111111."),
					AfterPresent: true,
				}}
				return c
			}(),
			// What record 0004 permits, and the case that says this rule did
			// not quietly forbid it. The code goes, the record stays, and it
			// gains the line naming the commit that removed it.
		},
	}
}

// TestARemovedRecordNamesWhatMayBeRemovedInstead holds the removal refusal to
// carrying the repair. Somebody meeting this has already decided the directory
// is finished with, and a refusal that only says no sends them looking for
// which of the two rules about records they have hit.
func TestARemovedRecordNamesWhatMayBeRemovedInstead(t *testing.T) {
	const record = "experiments/one/EXPERIMENT.md"

	change := clean()
	change.Body = "This closes #24. It removes " + record + "."
	change.Files = []File{{Path: record, Gone: true}}
	change.Records = []RecordChange{{
		Path:          record,
		Before:        recordAt("answered", theLandedAnswer),
		BeforePresent: true,
	}}

	refusals := Judge(change).Refusals
	if len(refusals) != 1 {
		t.Fatalf("the verdict carries %d refusals, want exactly one", len(refusals))
	}
	if refusals[0].Subject != record {
		t.Errorf("the refusal names %q, want %q", refusals[0].Subject, record)
	}
	if !strings.Contains(refusals[0].Detail, "0004") {
		t.Errorf("the refusal does not say what may be removed instead: %q", refusals[0].Detail)
	}
}

// TestARenameNamesBothEndsOfTheMove is the half of the issue a property
// identifier cannot carry. The repair is a choice between keeping the slug and
// writing a new experiment, and neither can be made by somebody who has been
// told only where the record used to be.
func TestARenameNamesBothEndsOfTheMove(t *testing.T) {
	const from = "experiments/one/EXPERIMENT.md"
	const to = "experiments/two/EXPERIMENT.md"
	landed := recordAt("answered", theLandedAnswer)

	change := clean()
	change.Files = []File{{Path: from, Gone: true}, {Path: to, From: from}}
	change.Records = []RecordChange{
		{Path: from, Before: landed, BeforePresent: true},
		{Path: to, After: landed, AfterPresent: true},
	}

	refusals := Judge(change).Refusals
	if len(refusals) != 1 {
		t.Fatalf("the verdict carries %d refusals, want exactly one", len(refusals))
	}
	if refusals[0].Property != ExperimentAlreadyLandedWasRenamed {
		t.Fatalf("the verdict refuses %s, want %s", refusals[0].Property, ExperimentAlreadyLandedWasRenamed)
	}
	if refusals[0].Subject != from {
		t.Errorf("the refusal names %q as its subject, want the path the record was at, %q", refusals[0].Subject, from)
	}
	if !strings.Contains(refusals[0].Detail, to) {
		t.Errorf("the refusal never names where the record went: %q", refusals[0].Detail)
	}
}

// TestARenamedRecordSurvivesTheJoinFromTheDiff is the join between the two
// halves of this package, and it was added because deleting the pairing in the
// parser left every case above green. Each half is proved on its own: the cases
// hand the rule a value somebody typed, and the parser is held to what git
// prints. Neither of them notices the pairing being dropped on the way from one
// to the other, which is one field in one branch of ParseFiles and exactly the
// line somebody tidying it would take out.
//
// The diff below is what git prints for a renamed experiment, in the shape
// TestParseFiles already carries.
func TestARenamedRecordSurvivesTheJoinFromTheDiff(t *testing.T) {
	const from = "experiments/one/EXPERIMENT.md"
	const to = "experiments/two/EXPERIMENT.md"
	landed := recordAt("answered", theLandedAnswer)

	files, err := ParseFiles([]byte("R100\x00" + from + "\x00" + to + "\x00"))
	if err != nil {
		t.Fatalf("cannot read the diff: %v", err)
	}

	change := clean()
	change.Files = files
	change.Records = []RecordChange{
		{Path: from, Before: landed, BeforePresent: true},
		{Path: to, After: landed, AfterPresent: true},
	}

	properties := Judge(change).Properties()
	if len(properties) != 1 || properties[0] != ExperimentAlreadyLandedWasRenamed {
		t.Fatalf("the verdict is %v, want exactly %s", properties, ExperimentAlreadyLandedWasRenamed)
	}
}

// TestAMoveNothingReportedAsOneIsRefusedAsARemoval holds the residual the rule
// declares. Where git did not call the move a rename, the same change is
// refused for the same reason under the other property, and the repair the
// message names is the wrong one of the two. It is written down here so that a
// reader meeting it knows it was chosen rather than missed.
func TestAMoveNothingReportedAsOneIsRefusedAsARemoval(t *testing.T) {
	const from = "experiments/one/EXPERIMENT.md"
	const to = "experiments/two/EXPERIMENT.md"
	landed := recordAt("answered", theLandedAnswer)

	change := clean()
	change.Body = "This closes #24. It removes " + from + " and adds " + to + "."
	change.Files = []File{{Path: from, Gone: true}, {Path: to}}
	change.Records = []RecordChange{
		{Path: from, Before: landed, BeforePresent: true},
		{Path: to, After: landed, AfterPresent: true},
	}

	properties := Judge(change).Properties()
	if len(properties) != 1 || properties[0] != RecordAlreadyLandedWasRemoved {
		t.Fatalf("the verdict is %v, want exactly %s", properties, RecordAlreadyLandedWasRemoved)
	}
}
