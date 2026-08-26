package check

import (
	"fmt"
	"time"
)

// FieldHeldBack is the date of the report a held-back record's window is
// counted from. Record 0022 adds it and record 0013 makes it optional, as it
// makes every field added after it.
//
// It is declared here rather than beside the fields record 0008 fixes, and
// that is the shape headerDateFields argues for at its own list: a field a
// later record adds arrives with the check that reads it, so a checker built
// before the field is unaware of it and a field with no check has nowhere to
// hide.
//
// What the field is for is not a property of this package and is written in
// record 0022 and in SECURITY.md. What this package does with it is read it,
// hold it to being a date, and give the listing something to print.
const FieldHeldBack = "Held-back"

// RecordHeldBackIsNotADate refuses a held-back field carrying something that
// is not a date written the one way record 0008 names.
//
// The value is the start of a window somebody outside this board is entitled
// to plan against, and it is the only thing in the tree that says when that
// window started. A value nothing can read leaves the record saying it is held
// back and saying nothing about since when, which is the half of the
// disclosure that costs something: the fact of a hold with no date is what
// record 0022 rejected as waiting indefinitely, arriving through a typo
// instead of through a decision.
//
// A field written with nothing after the colon is refused here too, for the
// reason RecordHeaderDateIsNotADate gives. Record 0013 makes absence legal for
// every field and an empty declaration a different statement, and a declared
// date with no value claims there is a date rather than claiming there is
// none.
const RecordHeldBackIsNotADate = "record-held-back-is-not-a-date"

// refuseHeldBack holds a held-back date to being a date.
//
// WHERE IT DOES NOT REACH, and this is the whole of the residual.
//
// An absent field is never refused. Record 0013 fixes that and record 0022
// says so of this field in its own words: a record that is being held back and
// carries no field sits in asking with a question that says nothing, and
// nothing here can refuse it. What stands behind that half is the template,
// the review and whoever writes the record.
//
// A date this reads is not a date it believes. Nothing in a checkout says when
// a report arrived, so the value is a claim its author made and the only thing
// judged is the shape of that claim. A held-back date invented and well formed
// passes.
//
// Two shapes a reader might expect here are deliberately outside it. A date
// later than the time the run read is not refused, so a window that has not
// started yet passes, and a held-back field on a record that is not asking is
// not refused either, though record 0022 fixes the state a hold is recorded
// in. Both are visible in the listing rather than refused, and neither has a
// property in this tree.
//
// A record whose bytes do not parse as a record is not judged here, for the
// reason every other header rule gives: nothing can read a field out of a file
// that has no header.
func refuseHeldBack(path string, data []byte) []Refusal {
	record, err := ParseRecord(data)
	if err != nil {
		return nil
	}

	written, present := record.Field(FieldHeldBack)
	if !present {
		return nil
	}
	if _, err := time.Parse(DateFormat, written); err == nil {
		return nil
	}

	return []Refusal{{
		Property: RecordHeldBackIsNotADate,
		Subject:  path,
		Detail: fmt.Sprintf("its %s is %q, and record 0022 counts the window from a date written as %s, for example %s",
			FieldHeldBack, written, dateShape, dateExample),
	}}
}
