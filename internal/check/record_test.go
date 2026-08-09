package check

import (
	"os"
	"strings"
	"testing"
)

// templatePath is the template record 0008 says carries the shape. Decision
// record 0002 puts it under docs/, so the path climbs out of internal/check.
const templatePath = "../../docs/experiment-template.md"

// TestTheTemplateParses is the standing agreement between the format and the
// file a contributor copies. A template is a convenience and the record is the
// authority, so the way the two are kept from disagreeing is that the runner
// reads the template in its own suite rather than a person reading both.
func TestTheTemplateParses(t *testing.T) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("cannot read %s: %v", templatePath, err)
	}

	rec, err := ParseRecord(data)
	if err != nil {
		t.Fatalf("the template does not parse: %v", err)
	}

	for _, name := range []string{FieldSlug, FieldState, FieldQuestionWritten} {
		if _, present := rec.Field(name); !present {
			t.Errorf("the template carries no %s field", name)
		}
	}

	// Answer-Written is absent on purpose. A template that carried it would
	// teach every new experiment to declare an answer date on the day the
	// question was written.
	if _, present := rec.Field(FieldAnswerWritten); present {
		t.Errorf("the template carries %s, which belongs in the change that writes the answer", FieldAnswerWritten)
	}

	if state, _ := rec.Field(FieldState); state != StateAsking {
		t.Errorf("the template opens in state %q, want %q", state, StateAsking)
	}

	for _, heading := range []string{SectionQuestion, SectionMethod, SectionAnswer} {
		if _, present := rec.Section(heading); !present {
			t.Errorf("the template carries no %s section", heading)
		}
	}

	// The empty answer section is the shape rather than an oversight. Record
	// 0008 asks for it from the day a record is created, so that writing an
	// answer is filling in something visibly missing.
	if body, present := rec.Section(SectionAnswer); present && body != "" {
		t.Errorf("the template's %s section carries %q, and it is meant to be empty", SectionAnswer, body)
	}

	if _, present := rec.Section(SectionPromotion); present {
		t.Errorf("the template carries a %s section, which record 0005 adds when work is handed over", SectionPromotion)
	}
}

// TestParseRecordReadsAHeaderAndItsSections is the ordinary case, written out
// in full rather than assembled, so that what the parser was given is the thing
// a reader of this test sees.
func TestParseRecordReadsAHeaderAndItsSections(t *testing.T) {
	rec, err := ParseRecord([]byte(`Slug: one
State: answered
Question-Written: 2026-01-01
Answer-Written: 2026-02-02

A title nobody parses.

## Question

Does it?

## Method

It was measured.

## Answer

No, and that is a finished piece of work.
`))
	if err != nil {
		t.Fatalf("this record does not parse: %v", err)
	}

	for _, want := range []struct{ name, value string }{
		{FieldSlug, "one"},
		{FieldState, StateAnswered},
		{FieldQuestionWritten, "2026-01-01"},
		{FieldAnswerWritten, "2026-02-02"},
	} {
		got, present := rec.Field(want.name)
		if !present {
			t.Errorf("%s is absent", want.name)
			continue
		}
		if got != want.value {
			t.Errorf("%s is %q, want %q", want.name, got, want.value)
		}
	}

	if rec.HeaderLines != 4 {
		t.Errorf("the header is %d lines, want 4", rec.HeaderLines)
	}

	// The text between the header and the first heading is prose and is not a
	// section. Reading it as one is how a title becomes a malformed section.
	if len(rec.Sections) != 3 {
		t.Fatalf("the record has %d sections, want 3", len(rec.Sections))
	}
	for i, want := range []string{SectionQuestion, SectionMethod, SectionAnswer} {
		if rec.Sections[i].Heading != want {
			t.Errorf("section %d is %q, want %q", i, rec.Sections[i].Heading, want)
		}
	}

	if body, _ := rec.Section(SectionAnswer); body != "No, and that is a finished piece of work." {
		t.Errorf("the answer section is %q", body)
	}
}

// TestParseRecordSeparatesAnAbsentFieldFromAnEmptyOne holds the split the rest
// of this format rests on. Record 0013 makes an absent field legal for ever and
// leaves a field declared empty refusable, so a parser that reported the two
// alike would remove the only thing a later check has to go on.
func TestParseRecordSeparatesAnAbsentFieldFromAnEmptyOne(t *testing.T) {
	rec, err := ParseRecord([]byte("Slug: one\nState:\n\n## Answer\n"))
	if err != nil {
		t.Fatalf("this record does not parse: %v", err)
	}

	value, present := rec.Field(FieldState)
	if !present {
		t.Errorf("%s was written with an empty value and reads as absent", FieldState)
	}
	if value != "" {
		t.Errorf("%s is %q, want the empty string", FieldState, value)
	}

	if _, present := rec.Field(FieldQuestionWritten); present {
		t.Errorf("%s was never written and reads as present", FieldQuestionWritten)
	}

	body, present := rec.Section(SectionAnswer)
	if !present {
		t.Errorf("the %s heading is there and reads as absent", SectionAnswer)
	}
	if body != "" {
		t.Errorf("the %s section is %q, want the empty string", SectionAnswer, body)
	}
}

// TestParseRecordToleratesCarriageReturns fixes what a record that arrived
// through a translating checkout parses to. The tree stores text with LF and
// .gitattributes holds every checkout to it, so this is a record that arrived
// some other way. Reading it as a state whose value ends in an invisible byte
// would send whoever hits it to the wrong repair.
//
// The bytes are escape sequences in this source rather than literal carriage
// returns, so nothing between here and git can remove the byte the test exists
// for.
func TestParseRecordToleratesCarriageReturns(t *testing.T) {
	rec, err := ParseRecord([]byte("Slug: one\r\nState: asking\r\n\r\n## Answer\r\n"))
	if err != nil {
		t.Fatalf("this record does not parse: %v", err)
	}
	if state, _ := rec.Field(FieldState); state != StateAsking {
		t.Errorf("state is %q, want %q", state, StateAsking)
	}
	if _, present := rec.Section(SectionAnswer); !present {
		t.Errorf("the %s heading was not read", SectionAnswer)
	}
}

// TestParseRecordRefusesBytesThatAreNotARecord covers the whole of what this
// parser returns an error for. Each case is a structural statement about the
// bytes and none of them is a field being absent, which record 0013 makes legal
// for ever.
func TestParseRecordRefusesBytesThatAreNotARecord(t *testing.T) {
	for _, c := range []struct {
		name, input, wants string
	}{
		{"an empty file", "", "empty"},
		{"only whitespace", "   \n\n", "empty"},
		{"opening with a blank line", "\nSlug: one\n", "line 1 is blank"},
		{"a header line that is not a field", "Slug: one\nnot a field at all\n\n## Answer\n", "line 2"},
		{"a field name with a space in it", "Slug: one\nAnswer Written: 2026-01-01\n", "line 2"},
		{"one field name written twice", "Slug: one\nSlug: two\n", "a second time"},
		{"one heading written twice", "Slug: one\n\n## Answer\n\n## Answer\n", "a second time"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseRecord([]byte(c.input))
			if err == nil {
				t.Fatalf("this parsed and should not have")
			}
			if !strings.Contains(err.Error(), c.wants) {
				t.Errorf("the message is %q and does not say %q, so a reader is not sent to the repair", err, c.wants)
			}
		})
	}
}

// TestParseRecordAcceptsAFieldNameTheFormatDoesNotFix holds the direction
// record 0013 is read from here. A field added by a later record is unknown to
// every checker built before it, so refusing an unrecognised name would break
// the same rule from the other side. The cost is that a misspelled name is read
// as a new field, and it is paid rather than overlooked.
func TestParseRecordAcceptsAFieldNameTheFormatDoesNotFix(t *testing.T) {
	rec, err := ParseRecord([]byte("Slug: one\nTouches-Real-Data: no\n\n## Answer\n"))
	if err != nil {
		t.Fatalf("a record carrying a field this format does not fix was refused: %v", err)
	}
	if value, present := rec.Field("Touches-Real-Data"); !present || value != "no" {
		t.Errorf("the unfixed field reads as %q, present=%v", value, present)
	}
}
