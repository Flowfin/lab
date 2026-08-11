package check

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// The header field names record 0008 fixes. They are named once here and
// nowhere else, so a check reading a field and a template writing one are the
// same string or they are not equal.
const (
	// FieldSlug repeats the name of the directory the record sits in, so a
	// reader holding a quoted slug can walk back to the experiment.
	FieldSlug = "Slug"

	// FieldState is the record's state, one of the three below.
	FieldState = "State"

	// FieldQuestionWritten is the date the question was written. It is what
	// the listing sorts by, which is why record 0008 says it does not move
	// after the commit that creates the directory.
	FieldQuestionWritten = "Question-Written"

	// FieldAnswerWritten is the date the answer was written. It is absent
	// until there is an answer.
	FieldAnswerWritten = "Answer-Written"

	// FieldNeedsHardware is what an experiment needs beyond the runner, in
	// words somebody can act on, or HardwareNone. Record 0015 adds it and
	// record 0013 makes it optional, so it is absent from every record
	// written before it and that is never a refusal.
	FieldNeedsHardware = "Needs-Hardware"
)

// HardwareNone is how a record says it needs nothing beyond the runner. It is
// the word that makes the declaration checkable in both directions: an absent
// field says nothing a tree can contradict, and this says something it can.
const HardwareNone = "none"

// The three states record 0003 fixes, written exactly as that record writes
// them.
const (
	StateAsking    = "asking"
	StateAnswered  = "answered"
	StateAbandoned = "abandoned"
)

// The level-two headings record 0008 fixes. Promotion is absent from every
// record that has not been handed over, which record 0005 sets out.
const (
	SectionQuestion  = "Question"
	SectionMethod    = "Method"
	SectionAnswer    = "Answer"
	SectionPromotion = "Promotion"
)

// fieldName is what a header field name may be. A name at column zero, a
// letter first, then letters, digits and hyphens. The pattern is here rather
// than in a sentence somewhere because it is what decides whether a line is a
// malformed field or something that is not a field at all, and those produce
// different messages.
var fieldName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*$`)

// A Field is one header line, as it was written.
type Field struct {
	// Name is the field name, without the colon.
	Name string

	// Value is what followed the colon, with the surrounding spaces
	// removed. A field written with nothing after the colon has an empty
	// value and is still present, because declaring a field and leaving it
	// blank is a different statement from not declaring it.
	Value string

	// Line is the one-based line the field was written on, so a message
	// about it can send a reader to the right place.
	Line int
}

// A Section is one level-two heading and the text under it, up to the next
// level-two heading or the end of the file.
type Section struct {
	// Heading is the text of the heading, without the two hashes.
	Heading string

	// Body is the text under it, with the surrounding blank lines removed.
	// An empty body is what an unanswered record's answer section has, and
	// that is the shape rather than an omission.
	Body string

	// Line is the one-based line the heading was written on.
	Line int
}

// A Record is an EXPERIMENT.md read into the two halves record 0008 divides it
// into: the header, which the runner reads, and the sections, which a person
// reads. Nothing here judges what any of it says.
type Record struct {
	// Fields are the header fields, in the order they were written.
	Fields []Field

	// Sections are the level-two sections, in the order they were written.
	Sections []Section

	// HeaderLines is how many lines the header occupied, so a caller can
	// tell a record whose header is the whole file from one that carries
	// prose as well.
	HeaderLines int
}

// Field returns the value of a header field and whether it was present. The
// two are separate answers: a field written with an empty value and a field
// nobody wrote are the same string and different statements, and record 0013
// makes the difference load-bearing, because absence is never a refusal and an
// empty declaration may be.
func (r Record) Field(name string) (string, bool) {
	for _, field := range r.Fields {
		if field.Name == name {
			return field.Value, true
		}
	}
	return "", false
}

// Section returns the body of a section and whether the heading was there. The
// same split as Field, and for the same reason: a record with no answer
// section and a record whose answer section is empty are different things, and
// record 0008 asks for the second one from the day a record is created.
func (r Record) Section(heading string) (string, bool) {
	for _, section := range r.Sections {
		if section.Heading == heading {
			return section.Body, true
		}
	}
	return "", false
}

// ParseRecord reads the bytes of an EXPERIMENT.md into a Record.
//
// It returns an error only where the bytes are not a record at all: no header,
// a header line that is not a field, or one name written twice. Everything a
// field or a section says is somebody else's judgement, and this returns it
// unread. In particular an absent field is never an error here, because record
// 0013 makes absence legal for every field added after it and a parser that
// refused one would be the same rule from the other side.
//
// WHAT THIS DOES NOT DO. Nothing in the runner calls this yet. It is reached
// by this package's own tests and by the test that parses the shipped
// template, and no walk refuses a record over anything it reports. The
// refusals arrive with the issues that argue them, and a green run today says
// the records were read rather than that they were judged.
func ParseRecord(data []byte) (Record, error) {
	var rec Record

	// A record that arrived through a translating checkout ends its lines
	// with a carriage return, and this reads it as the same record. The
	// tree stores text with LF and .gitattributes holds every checkout to
	// it, so such a record arrived some other way. Reading it as a state
	// whose value ends in an invisible byte would send whoever hits it to
	// the wrong repair, and refusing it is a decision about the bytes of a
	// record, which refuseBytes owns rather than this.
	//
	// There is no separate step for it. Every value and every heading below
	// is trimmed of surrounding whitespace, and a carriage return is
	// whitespace, so the tolerance falls out of the trimming rather than
	// out of a line somebody could delete without the suite noticing.
	// TestParseRecordToleratesCarriageReturns is what holds it.
	lines := strings.Split(string(data), "\n")

	if len(bytes.TrimSpace(data)) == 0 {
		return rec, fmt.Errorf("the file is empty, so it carries no header")
	}
	if strings.TrimSpace(lines[0]) == "" {
		return rec, fmt.Errorf("line 1 is blank, so the file opens with no header")
	}

	seen := make(map[string]int)
	end := len(lines)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			end = i
			break
		}
		name, value, found := strings.Cut(line, ":")
		if !found || !fieldName.MatchString(name) {
			return rec, fmt.Errorf("line %d is %q, which is not a header field written as a name, a colon and a value", i+1, line)
		}
		if first, repeated := seen[name]; repeated {
			return rec, fmt.Errorf("line %d writes %s a second time, first written on line %d, so a check reading that field would have to choose", i+1, name, first)
		}
		seen[name] = i + 1
		rec.Fields = append(rec.Fields, Field{Name: name, Value: strings.TrimSpace(value), Line: i + 1})
	}
	rec.HeaderLines = end

	headings := make(map[string]int)
	for i := end; i < len(lines); i++ {
		heading, isHeading := strings.CutPrefix(lines[i], "## ")
		if !isHeading {
			continue
		}
		heading = strings.TrimSpace(heading)
		if first, repeated := headings[heading]; repeated {
			return rec, fmt.Errorf("line %d writes the %s section a second time, first written on line %d, so a check reading that section would have to choose", i+1, heading, first)
		}
		headings[heading] = i + 1

		body := i + 1
		for ; body < len(lines); body++ {
			if strings.HasPrefix(lines[body], "## ") {
				break
			}
		}
		rec.Sections = append(rec.Sections, Section{
			Heading: heading,
			Body:    strings.TrimSpace(strings.Join(lines[i+1:body], "\n")),
			Line:    i + 1,
		})
	}

	return rec, nil
}
