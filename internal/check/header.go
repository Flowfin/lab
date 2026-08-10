package check

import (
	"fmt"
	"time"
)

// The properties a header can be refused for where its own fields disagree.
const (
	// RecordHeaderDateIsNotADate refuses a date field carrying something that
	// is not a date written the one way record 0008 names. A field holding a
	// month, a season or a sentence parses as nothing, and the listing sorts
	// on the question's date: a value nothing can read sorts wherever the
	// comparison happens to put it, and the order is the one thing a reader
	// takes from that listing.
	//
	// A field written with nothing after the colon is refused here too. Record
	// 0013 makes absence legal for every field and an empty declaration a
	// different statement, and a declared date with no value is a claim that
	// there is a date rather than a claim that there is none.
	RecordHeaderDateIsNotADate = "record-header-date-is-not-a-date"

	// RecordAnswerDateDisagreesWithTheRecord refuses an answer date the rest
	// of the record contradicts, at two sites.
	//
	// An answer dated before the question is a timeline that cannot have
	// happened, whichever of the two fields is wrong.
	//
	// An answer date on a record still saying asking is the arm worth the
	// fixture. It passes every other rule in this tree, because a record in
	// asking is allowed an empty answer section and nothing else reads the
	// field, and what it means is a record edited towards a state it is not
	// in. Record 0008 says the answer date is added in the same change as the
	// answer, and this is that sentence made refusable.
	RecordAnswerDateDisagreesWithTheRecord = "record-answer-date-disagrees-with-the-record"
)

// headerDateFields are the fields record 0008 fixes as dates, in the order it
// names them, so a record carrying two bad ones is refused in the order
// somebody reads the header.
//
// A field a later record adds is not here. Record 0013 makes a new field
// optional and a checker built before it unaware of it, so a date field added
// later is refused by the change that adds it rather than by this list
// silently widening.
var headerDateFields = []string{FieldQuestionWritten, FieldAnswerWritten}

// dateShape is how a refusal spells the format to somebody reading it, and
// dateExample is a date in it. DateFormat is Go's own spelling of the same
// shape and it reads as a date rather than as a rule, so a message quoting it
// tells a reader that the second of January 2006 was expected.
//
// The two cannot drift apart in silence: TestTheExampleInAMessageIsADate
// parses dateExample under DateFormat, so a change to either that leaves them
// disagreeing reddens the suite.
const (
	dateShape   = "YYYY-MM-DD"
	dateExample = "2026-01-02"
)

// refuseHeaderDates holds the header's dates to being dates and to agreeing
// with each other and with the state.
//
// WHAT IT DOES NOT JUDGE, and this is the whole of the residual. Nothing in a
// checkout says when somebody started thinking about a question or when they
// finished. A date here is a claim its author made, and every rule below is
// about the shape of that claim and the order of two of them. A record whose
// dates are both invented and consistent passes, so a green run is evidence
// that the timeline is readable rather than that it is real.
//
// A record whose bytes do not parse as a record is not judged here, for the
// reason refuseState and refuseDates both give: nothing can read a field out of
// a file that has no header.
func refuseHeaderDates(path string, data []byte) []Refusal {
	record, err := ParseRecord(data)
	if err != nil {
		return nil
	}

	var refusals []Refusal
	dates := make(map[string]time.Time)

	for _, name := range headerDateFields {
		written, present := record.Field(name)
		if !present {
			// Absence is never a refusal. Record 0013 fixes that, and
			// Answer-Written is absent from every record that has not been
			// answered.
			continue
		}
		date, err := time.Parse(DateFormat, written)
		if err != nil {
			refusals = append(refusals, Refusal{
				Property: RecordHeaderDateIsNotADate,
				Subject:  path,
				Detail: fmt.Sprintf("its %s is %q, and record 0008 writes a date as %s, for example %s",
					name, written, dateShape, dateExample),
			})
			continue
		}
		dates[name] = date
	}

	answered, hasAnswerDate := dates[FieldAnswerWritten]
	if !hasAnswerDate {
		return refusals
	}

	if state, _ := record.Field(FieldState); state == StateAsking {
		refusals = append(refusals, Refusal{
			Property: RecordAnswerDateDisagreesWithTheRecord,
			Subject:  path,
			Detail: fmt.Sprintf("it says %s and carries an %s of %s, so it has been edited towards a state it is not in",
				StateAsking, FieldAnswerWritten, answered.Format(DateFormat)),
		})
		return refusals
	}

	if asked, hasQuestionDate := dates[FieldQuestionWritten]; hasQuestionDate && answered.Before(asked) {
		refusals = append(refusals, Refusal{
			Property: RecordAnswerDateDisagreesWithTheRecord,
			Subject:  path,
			Detail: fmt.Sprintf("its %s is %s and its %s is %s, so the answer is dated before the question",
				FieldAnswerWritten, answered.Format(DateFormat),
				FieldQuestionWritten, asked.Format(DateFormat)),
		})
	}

	return refusals
}
