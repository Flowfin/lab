package pullrequest

// The rule this board's central claim rests on: a record already on the default
// branch is added to and never rewritten.
//
// WHY IT IS HERE AND NOT IN THE RUNNER. The runner reads a checkout, and
// telling a rewritten answer from a new one needs the version already on the
// default branch, which is history rather than a tree. Giving the runner a
// history reader would cost it the near-zero dependency surface record 0001
// chose and the claim that it opens no connection leans on. This check already
// holds both ends of the range, so the comparison belongs here.

import (
	"fmt"
	"strings"

	"github.com/Flowfin/lab/internal/check"
)

// AnswerAlreadyLandedWasRewritten refuses a change that removes or alters a
// line of an answer section in a record that was already answered at the base
// of the range.
//
// The failure has a shape and it is not malicious. An experiment answered no.
// Months later the same work is being offered somewhere, or the answer simply
// reads worse than it felt at the time, and one line makes it read better.
// Every other check here stays green through that edit, and the diff arrives in
// a pull request nobody reads closely because the record is four paragraphs
// long. A record whose answer can be quietly improved later is a record a
// reader cannot use as evidence, because the version in front of them is the
// version somebody was last comfortable with.
//
// Adding is the whole of what is allowed, and it covers every honest case
// including the one where the first answer turned out to be wrong: the
// correction is added underneath, saying what was wrong and how it was found.
const AnswerAlreadyLandedWasRewritten = "answer-already-landed-was-rewritten"

// QuestionAlreadyAskedWasRewritten refuses a change that removes or alters a
// line of the question section of a record that was already on the branch this
// change lands on.
//
// An experiment states its question before it starts, and the whole value of
// that sentence is the ordering: a question written before the work is a
// question the result cannot have been chosen to fit. Nothing held the question
// to that ordering after the commit that wrote it. It was checked for being
// non-empty on the day it landed and for nothing at all afterwards.
//
// The failure needs no bad faith and it is the most likely dishonest edit here,
// because it is the one that makes a record read better rather than worse. An
// experiment asked whether one approach was faster than another. The
// measurement came back saying something adjacent and more interesting. Editing
// the question to name what was measured turns a result nobody predicted into a
// result the record claims was the point, and every other check stays green
// through it. The edit arrives in the same pull request as the answer, which is
// the change a reader is reading for something else.
//
// Where the question turned out to be the wrong question, that is itself an
// answer and the lifecycle already names it: the answer section says so, the
// question stays as it was asked, and the record is finished rather than tidied.
const QuestionAlreadyAskedWasRewritten = "question-already-asked-was-rewritten"

// A RecordChange is one experiment record at both ends of the range. It carries
// bytes rather than a parsed record, because what parses at one end may not
// parse at the other and that difference is itself something a rule reads.
type RecordChange struct {
	// Path is where the record is, relative to the root of the repository.
	Path string

	// Before is the record as it was at the base of the range, and
	// BeforePresent says there was one. A record this change creates has no
	// before, which is not a violation of anything and is the case the board
	// wants most.
	Before        []byte
	BeforePresent bool

	// After is the record at the head of the range, and AfterPresent says it
	// is still there. A record this change removes has no after. That a landed
	// record may not be removed at all is a different rule about a different
	// failure, and it is record-already-landed-was-removed in removal.go
	// rather than either of the two here.
	After        []byte
	AfterPresent bool
}

// judgeRecords holds every record in the change to the two rules above.
//
// THE BOUNDARIES, WRITTEN HERE RATHER THAN DISCOVERED.
//
// A record that did not exist at the base of the range is not covered, and
// neither is a section that was not there. Writing the question for the first
// time is the thing this board wants most, and writing an answer for the first
// time is the second; a check that made a first draft permanent would teach
// people to commit nothing until they were sure, which is how an experiment
// ends up with nothing written down at all. The answer rule adds one more
// condition, at the rule itself, because a draft answer under a record that is
// still asking is the same case one step further on.
//
// Nothing here judges whether what was added is honest. A correction that
// contradicts the original without saying so passes, and so does a clarification
// that is a second question wearing a hat. What the rules buy is that the
// original words are still on the page next to whatever arrived later, which is
// the difference between a record and a current opinion, and review is where the
// rest is caught.
//
// NOTHING REQUIRES AN ADDITION TO SAY THAT IT WAS ADDED LATER. Issue #70
// describes a clarification as sitting under a line saying it arrived after the
// work started, and record 0008 fixes no such line, so there is no marker in
// the format for a check to read. A rule invented here would be a format
// decision taken inside a checker, which is the shape the record for the
// format would then have to be written around.
func judgeRecords(change Change) Verdict {
	if !change.RecordsRead {
		return Verdict{Skips: []Skip{
			{
				Rule: AnswerAlreadyLandedWasRewritten,
				Why:  "this run was given no records, so no answer was compared against the one already landed",
			},
			{
				Rule: QuestionAlreadyAskedWasRewritten,
				Why:  "this run was given no records, so no question was compared against the one the work began from",
			},
		}}
	}

	var verdict Verdict
	for _, record := range change.Records {
		verdict.add(judgeOneRecord(record))
	}
	return verdict
}

func judgeOneRecord(record RecordChange) Verdict {
	if !record.BeforePresent || !record.AfterPresent {
		return Verdict{}
	}

	before, err := check.ParseRecord(record.Before)
	if err != nil {
		// Nothing can read a section out of bytes that are not a record. What
		// the record said at the base decides whether either rule applies at
		// all, so a base that cannot be read is a base they have nothing to
		// say about, and the checker is what refuses a record that is not one.
		return Verdict{}
	}
	after, afterErr := check.ParseRecord(record.After)

	var verdict Verdict
	for _, rule := range sectionRules {
		verdict.add(rule.judge(record, before, after, afterErr))
	}
	return verdict
}

// A sectionRule is one section of a record held to growing rather than
// changing. The two rules differ in which section they read, in what makes a
// record covered at the base, and in what the message tells the author to do
// instead. Everything else is the same comparison, which is why it is one
// function: two copies of it would drift, and the drift would be a rule that is
// stricter about one section than about the other for no reason anybody wrote
// down.
type sectionRule struct {
	property string
	heading  string

	// covered says whether the record at the base is one this rule applies
	// to. It is the boundary each issue asks to be written at the check.
	covered func(before check.Record) bool

	// instead is what the refusal tells the author to do with the words they
	// wanted to change.
	instead string
}

var sectionRules = []sectionRule{
	{
		property: AnswerAlreadyLandedWasRewritten,
		heading:  check.SectionAnswer,
		// A record whose state at the base was not answered is not covered.
		// Writing an answer for the first time is the thing this board wants,
		// and a check that made the first draft permanent would teach people
		// to commit nothing until they were certain, which is how an
		// experiment ends up with no written answer at all.
		covered: func(before check.Record) bool {
			state, present := before.Field(check.FieldState)
			return present && state == check.StateAnswered
		},
		instead: "an answer already landed may grow and may not change, so a correction goes underneath saying what was wrong and how it was found",
	},
	{
		property: QuestionAlreadyAskedWasRewritten,
		heading:  check.SectionQuestion,
		// Every record that carried a question at the base is covered,
		// whatever state it is in. The question is what the work began from,
		// and it does not become editable because the work is unfinished.
		covered: func(check.Record) bool { return true },
		instead: "the question is what the work began from, so a clarification goes underneath rather than over the words the work started with",
	},
}

// judge holds one section of one record to the rule.
func (r sectionRule) judge(record RecordChange, before, after check.Record, afterErr error) Verdict {
	landed, present := before.Section(r.heading)
	if !present || !r.covered(before) {
		return Verdict{}
	}
	// A section that was there and empty carries nothing to lose, and the
	// rules that refuse an empty question and an answer claimed but not
	// carried are the checker's. Reading it here would put a second judgement
	// about emptiness in a rule about editing.
	if len(meaningfulLines(landed)) == 0 {
		return Verdict{}
	}

	if afterErr != nil {
		// The head no longer parses as a record, so nothing can be shown to
		// still be there. It is refused here rather than left to the checker,
		// because the checker walks the tree and would report a record it
		// cannot read without ever saying that words already on the branch
		// went missing inside it.
		return Verdict{Refusals: []Refusal{{
			Property: r.property,
			Subject:  record.Path,
			Detail:   fmt.Sprintf("it no longer parses as a record, so its %s section cannot be shown to still carry what it carried at the base of this range", r.heading),
		}}}
	}
	current, present := after.Section(r.heading)
	if !present {
		return Verdict{Refusals: []Refusal{{
			Property: r.property,
			Subject:  record.Path,
			Detail:   fmt.Sprintf("its %s section is gone and it carried one at the base of this range", r.heading),
		}}}
	}

	missing, found := firstLineLost(landed, current)
	if !found {
		return Verdict{}
	}
	return Verdict{Refusals: []Refusal{{
		Property: r.property,
		Subject:  record.Path,
		Detail:   fmt.Sprintf("its %s section no longer carries %q, and %s", r.heading, shorten(missing), r.instead),
	}}}
}

// firstLineLost says whether every line the landed section carried is still in
// the current one, in the order it was written, and returns the first that is
// not.
//
// It is a subsequence rather than a prefix, so a line added between two that
// were already there is allowed. Adding is the whole of what the rule permits
// and where the addition sits is the author's judgement.
//
// Blank lines are not compared. A blank line carries no words, and a paragraph
// separated differently is not an answer that was rewritten. Every other
// difference is: a line reflowed, a word changed, a sentence softened.
func firstLineLost(landed, current string) (string, bool) {
	want := meaningfulLines(landed)
	have := meaningfulLines(current)

	at := 0
	for _, line := range want {
		found := false
		for at < len(have) {
			if have[at] == line {
				at++
				found = true
				break
			}
			at++
		}
		if !found {
			return line, true
		}
	}
	return "", false
}

// meaningfulLines is the lines of a section that carry words, with the
// whitespace at their ends removed so that a trailing space is not the whole of
// what a check refuses. The formatting rules refuse such a space anyway, and
// this rule is about what an answer says.
func meaningfulLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, " \t\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

// shorten cuts a quoted line to something a message can carry. A refusal naming
// a whole paragraph is a refusal nobody reads to the end of.
func shorten(line string) string {
	const bound = 72
	if len([]rune(line)) <= bound {
		return line
	}
	return string([]rune(line)[:bound]) + "..."
}
