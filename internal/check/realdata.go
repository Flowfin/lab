package check

import (
	"fmt"
	"strings"
)

// FieldRealData is what an experiment declares before it touches real personal
// data. Record 0025 adds it and record 0013 makes it optional, as it makes
// every field added after it.
//
// It is declared here rather than beside the fields record 0008 fixes, and
// that is the shape headerDateFields argues for at its own list: a field a
// later record adds arrives with the check that reads it, so a checker built
// before the field is unaware of it and a field with no check has nowhere to
// hide.
//
// What the field is for is not a property of this package and is written in
// record 0025 and in docs/privacy.md. What this package does with it is read
// it and hold it to saying something.
const FieldRealData = "Real-Data"

// RealDataNone is how a record says it touches no real personal data. It is
// the word that makes the declaration mean something in both directions: an
// absent field says nothing at all, and this says something a reader can hold
// the record to. It is the same shape as HardwareNone and for the same reason.
const RealDataNone = "none"

// RecordRealDataDeclarationIsEmpty refuses a record that declares Real-Data
// and writes nothing after the colon.
//
// Record 0013 makes an absent field legal and an empty declaration a different
// statement, and this is the second one. The field carries what category of
// data, on whose host, and what will be written down about it. A field with
// nothing after the colon carries none of the three while reading, to anybody
// scanning the header, as a record that declared its data and was cleared.
// That is the mistake somebody actually makes: the field is typed in the
// commit that writes the question, the value is left for the moment the work
// starts, and the moment the work starts is exactly when nobody goes back to
// the header.
//
// It says none of the three rather than only the third, and the property is
// named for the shape it can see rather than for the sentence in record 0025
// it serves. Whether words after the colon name a category, a host and a
// measurement is a judgement about meaning that no reading of a checkout
// makes, and naming this property after that sentence would claim a reading it
// does not perform.
const RecordRealDataDeclarationIsEmpty = "record-real-data-declaration-is-empty"

// refuseRealData holds a real-data declaration to saying something.
//
// WHERE IT DOES NOT REACH, and this is the whole of the residual.
//
// An absent field is never refused. Record 0013 fixes that, and record 0025
// says so of this field in its own words: an experiment that touches real
// personal data and declares nothing is refused by nothing here, and what
// stands behind that half is the template, the review and whoever writes the
// record. It is the same hole record 0015 accepted for Needs-Hardware and it
// is accepted here for the same price.
//
// A declaration this reads is not a declaration it believes. Nothing in a
// checkout says what an experiment did on somebody else's machine, so the
// value is a claim its author made and the only thing judged is that a claim
// was made at all. A declaration naming a category, a host and a measurement,
// all three invented, passes.
//
// Whether the declaration was written before the work started is outside it
// too. Record 0025 asks for it in the commit that writes the question, and a
// runner that reads a checkout reads neither commits nor the order they
// arrived in. What the field says about that is a claim, and the diff is where
// it is read.
//
// Nothing here reads what the record carries. The rule that real data never
// enters the tree has no mechanism, which docs/privacy.md already says of
// itself, and this check does not narrow that sentence by one word: a record
// declaring a category and carrying a sample of it passes every rule in this
// package.
//
// A record whose bytes do not parse as a record is not judged here, for the
// reason every other header rule gives: nothing can read a field out of a file
// that has no header.
func refuseRealData(path string, data []byte) []Refusal {
	record, err := ParseRecord(data)
	if err != nil {
		return nil
	}

	declared, present := record.Field(FieldRealData)
	if !present {
		return nil
	}
	if strings.TrimSpace(declared) != "" {
		return nil
	}

	return []Refusal{{
		Property: RecordRealDataDeclarationIsEmpty,
		Subject:  path,
		Detail: fmt.Sprintf("it declares %s and writes nothing after the colon, so it names no data, no host and nothing that will be written down, and it does not say %s",
			FieldRealData, RealDataNone),
	}}
}
