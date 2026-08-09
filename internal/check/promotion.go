package check

import (
	"fmt"
	"regexp"
	"strings"
)

// The properties a promotion section can be refused for.
const (
	// PromotionSectionIsMissingSomething refuses a promotion section that does
	// not name all four of the things record 0005 fixes. A promotion that is
	// described but not linked rots: six months later the record says the work
	// continued on another board and nobody can find where, which is the same
	// as the work not having continued.
	PromotionSectionIsMissingSomething = "promotion-section-is-missing-something"

	// PromotionNamesABranchRatherThanARange refuses a commit range written as
	// something that moves. The whole purpose of the record is to say what was
	// handed over rather than what is there now, and a branch name answers the
	// second question on whatever day it is read.
	PromotionNamesABranchRatherThanARange = "promotion-names-a-branch-rather-than-a-range"

	// PromotionOnARecordThatIsNotAnswered refuses a promotion section on a
	// record in any other state. Record 0005 says the state does not change on
	// promotion, because an experiment that was offered elsewhere is still
	// answered, and that sentence only holds if the reverse is refused. A
	// record still asking that already names where its result went is either a
	// promotion of something with no written answer or a state nobody updated,
	// and both are the half-finished shape this board exists to make visible.
	PromotionOnARecordThatIsNotAnswered = "promotion-on-a-record-that-is-not-answered"
)

// The four things record 0005 says a promotion section names, and nothing about
// it is optional. They are read as a name, a colon and a value, which is the
// shape the record header already uses, so a reader writing one is not learning
// a second syntax for the same job.
const (
	PromotionRepository = "Repository"
	PromotionIssue      = "Issue"
	PromotionCommits    = "Commits"
	PromotionLicence    = "Licence"
)

// promotionFields are those four in the order record 0005 names them, so a
// refusal lists what is missing in the order somebody would fill it in.
var promotionFields = []string{
	PromotionRepository,
	PromotionIssue,
	PromotionCommits,
	PromotionLicence,
}

// commitRange is what the commits line has to look like: two revisions with two
// dots between them, each written as a hexadecimal object name long enough to
// be worth quoting. A name that is not one of those is something that moves,
// which is what this rule exists to refuse.
var commitRange = regexp.MustCompile(`^[0-9a-f]{7,40}\.\.\.?[0-9a-f]{7,40}$`)

// promotionLine reads one line of the section as a name, a colon and a value.
var promotionLine = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9-]*):\s*(.*)$`)

// refusePromotion holds a record that carries a promotion section to what
// record 0005 says such a section is.
//
// WHAT THIS CANNOT DO, and it is here rather than only in the issue that argued
// it. Nothing verifies that the destination repository exists, that the issue
// was ever opened, that anybody agreed to anything, or that the commit range is
// the one the receiving board took. The runner reads a checkout and makes no
// network call, which is a decision taken elsewhere and not one to trade away
// for this check. So the refusal is about shape, and the truth of the link is
// what a reader sees when they follow it. A green run is not confirmation that
// the hand-over happened.
func refusePromotion(path string, data []byte) []Refusal {
	record, err := ParseRecord(data)
	if err != nil {
		return nil
	}

	body, present := record.Section(SectionPromotion)
	if !present {
		// Absent from every record that has not been handed over, which record
		// 0005 fixes and which is the ordinary case rather than an omission.
		return nil
	}

	state, _ := record.Field(FieldState)
	if state != StateAnswered {
		return []Refusal{{
			Property: PromotionOnARecordThatIsNotAnswered,
			Subject:  path,
			Detail: fmt.Sprintf("it carries a %s section and its %s is %q, and record 0005 hands over an experiment that is %s",
				SectionPromotion, FieldState, state, StateAnswered),
		}}
	}

	named := make(map[string]string)
	for _, line := range strings.Split(body, "\n") {
		if match := promotionLine.FindStringSubmatch(strings.TrimSpace(line)); match != nil {
			named[match[1]] = strings.TrimSpace(match[2])
		}
	}

	var missing []string
	for _, field := range promotionFields {
		if named[field] == "" {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return []Refusal{{
			Property: PromotionSectionIsMissingSomething,
			Subject:  path,
			Detail: fmt.Sprintf("its %s section names no %s, and record 0005 says it names %s and that nothing about it is optional",
				SectionPromotion, strings.Join(missing, " and no "), strings.Join(promotionFields, ", ")),
		}}
	}

	if !commitRange.MatchString(named[PromotionCommits]) {
		return []Refusal{{
			Property: PromotionNamesABranchRatherThanARange,
			Subject:  path,
			Detail: fmt.Sprintf("its %s is %q, which is not two object names with dots between them, so it points at something that moves rather than at what was handed over",
				PromotionCommits, named[PromotionCommits]),
		}}
	}

	return nil
}
