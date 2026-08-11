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
	// failure, and it is issue #69's rather than this one's.
	After        []byte
	AfterPresent bool
}

// judgeRecords holds every record in the change to the rule above.
//
// TWO BOUNDARIES, WRITTEN HERE RATHER THAN DISCOVERED.
//
// A record whose state at the base of the range was not answered is not
// covered. Writing an answer for the first time is the thing this board wants,
// and a check that made the first draft permanent would teach people to commit
// nothing until they were certain, which is how an experiment ends up with no
// written answer at all. The same reasoning covers a record that did not exist
// at the base.
//
// Nothing here judges whether an added correction is honest. A correction that
// contradicts the original without saying so passes this. What the check buys is
// that the original is still there to be read next to it, which is the
// difference between a record and a current opinion, and review is where the
// rest is caught.
func judgeRecords(change Change) Verdict {
	if !change.RecordsRead {
		return Verdict{Skips: []Skip{{
			Rule: AnswerAlreadyLandedWasRewritten,
			Why:  "this run was given no records, so no answer was compared against the one already landed",
		}}}
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
		// Nothing can read a section out of bytes that are not a record. The
		// state at the base decides whether this rule applies at all, so a
		// base that cannot be read is a base this rule has nothing to say
		// about, and the checker is what refuses a record that is not one.
		return Verdict{}
	}
	if state, present := before.Field(check.FieldState); !present || state != check.StateAnswered {
		return Verdict{}
	}
	landed, present := before.Section(check.SectionAnswer)
	if !present {
		return Verdict{}
	}

	after, err := check.ParseRecord(record.After)
	if err != nil {
		// The head no longer parses as a record, so the answer cannot be shown
		// to be intact. It is refused here rather than passed to the checker,
		// because the checker walks the tree and would report a record it
		// cannot read without ever saying that an answer already landed went
		// missing inside it.
		return Verdict{Refusals: []Refusal{{
			Property: AnswerAlreadyLandedWasRewritten,
			Subject:  record.Path,
			Detail:   fmt.Sprintf("it was answered at the base of this range and no longer parses as a record, so its %s section cannot be shown to still carry what it carried", check.SectionAnswer),
		}}}
	}
	current, present := after.Section(check.SectionAnswer)
	if !present {
		return Verdict{Refusals: []Refusal{{
			Property: AnswerAlreadyLandedWasRewritten,
			Subject:  record.Path,
			Detail:   fmt.Sprintf("its %s section is gone and it carried an answer at the base of this range", check.SectionAnswer),
		}}}
	}

	missing, found := firstLineLost(landed, current)
	if !found {
		return Verdict{}
	}
	return Verdict{Refusals: []Refusal{{
		Property: AnswerAlreadyLandedWasRewritten,
		Subject:  record.Path,
		Detail: fmt.Sprintf("its %s section no longer carries %q, and an answer already landed may grow and may not change, so a correction goes underneath saying what was wrong and how it was found",
			check.SectionAnswer, shorten(missing)),
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
