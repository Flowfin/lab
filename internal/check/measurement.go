package check

import "fmt"

// FieldMeasurementCommit is the object name of the commit a measurement in the
// answer was produced at. Record 0016 adds it and record 0013 makes it
// optional, as it makes every field added after it.
//
// It is declared here rather than beside the four record 0008 fixes, and that
// is the shape headerDateFields already argues for at its own list: a field a
// later record adds arrives with the check that reads it, so a checker built
// before the field is unaware of it and a field with no check has nowhere to
// hide.
const FieldMeasurementCommit = "Measurement-Commit"

// RecordMeasurementCommitIsNotACommit refuses a measurement commit that is not
// the shape of an object name.
//
// An answer quotes a command and its output and says nothing about which
// version of the code the command ran against. Record 0004 lets that code be
// removed afterwards, and when it is, the answer keeps its numbers and loses
// the thing that produced them without reading as any less complete.
//
// The abbreviation is the value worth the fixture, because it is the mistake
// somebody actually makes: the short name is what git prints in a log, it is
// what a hand copies, it resolves on the machine it was copied on, and it stops
// resolving on a repository that has grown into a collision. It also cannot be
// told from a typo, since a short name that resolves to nothing and a mistyped
// one are the same string.
//
// A declared field with nothing after the colon is refused here too, for the
// reason the date refusal gives: record 0013 makes absence legal and an empty
// declaration a different statement, and a commit field with no value claims
// there is a commit rather than claiming there is none.
const RecordMeasurementCommitIsNotACommit = "record-measurement-commit-is-not-a-commit"

// The object name lengths this accepts, and the reason there are two.
//
// Every object in this repository is named in the first of them today. The
// second is what the same repository would name its objects in under the other
// hash git can be told to use, and refusing it would make this check a decision
// about the object format, which record 0016 does not take and which is not a
// question about a measurement.
const (
	shortObjectName = 40
	longObjectName  = 64
)

// refuseMeasurementCommit holds a measurement commit to the shape of an object
// name.
//
// WHAT IT CANNOT SAY, and this is the whole of the residual. Whether the object
// is in this repository, whether it is a commit rather than a blob, and whether
// the command quoted in the answer was ever run against it. The runner reads a
// checkout and opens no connection, and asking git any of those would cost it
// the dependency surface record 0001 chose. A value of forty hexadecimal
// characters that resolve to nothing passes here, so a green run says the name
// is resolvable in principle rather than that it resolves.
//
// A record whose bytes do not parse as a record is not judged here, for the
// reason every other header rule gives: nothing can read a field out of a file
// that has no header.
func refuseMeasurementCommit(path string, data []byte) []Refusal {
	record, err := ParseRecord(data)
	if err != nil {
		return nil
	}

	named, present := record.Field(FieldMeasurementCommit)
	if !present {
		// Absence is never a refusal. Record 0013 fixes that, and most
		// records carry no measurement at all.
		return nil
	}
	if isObjectName(named) {
		return nil
	}

	return []Refusal{{
		Property: RecordMeasurementCommitIsNotACommit,
		Subject:  path,
		Detail: fmt.Sprintf("its %s is %q, and record 0016 writes an object name in full, which is %d or %d characters of lowercase hexadecimal",
			FieldMeasurementCommit, named, shortObjectName, longObjectName),
	}}
}

// isObjectName says whether a value is the shape git names an object in.
func isObjectName(value string) bool {
	if len(value) != shortObjectName && len(value) != longObjectName {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
