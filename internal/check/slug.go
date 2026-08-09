package check

import (
	"fmt"
	"regexp"
	"strings"
)

// The properties a slug can be refused for.
const (
	// ExperimentDirectoryIsNotALegalSlug refuses a directory under
	// experiments/ whose name is not a slug. A directory named
	// "Timing test (final, v2)" satisfies every other rule in this tree, and
	// the string it produces survives none of the places a slug is quoted: a
	// listing, a promotion section read on another board, a sentence somebody
	// writes about the result.
	ExperimentDirectoryIsNotALegalSlug = "experiment-directory-is-not-a-legal-slug"

	// RecordSlugIsNotALegalSlug refuses a Slug field that is not a slug. It is
	// the same rule read from the other side, and it catches a record copied
	// out of another experiment and half edited.
	RecordSlugIsNotALegalSlug = "record-slug-is-not-a-legal-slug"

	// TwoExperimentsShareASlug refuses two experiments declaring one slug once
	// case is ignored. A slug is what a reader holds when they walk back from
	// a quoted result to the experiment that produced it, and two experiments
	// answering to it means that walk lands in the wrong place or nowhere.
	TwoExperimentsShareASlug = "two-experiments-share-a-slug"
)

// legalSlug is the shape record 0014 fixes: lower case letters, digits and
// single hyphens, beginning and ending with a letter or a digit. It is written
// here as well as in the record, because a check that reads its rule out of a
// document goes quiet the day the document is reworded.
var legalSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// slugLimit is the longest a slug may be. Record 0014 chose the number: longer
// than any question anybody has needed to name, short enough that a path built
// from it survives every filesystem the release targets.
const slugLimit = 64

// refuseSlug holds one string to the shape. It returns the reason it is not a
// slug, or the empty string where it is one, so that a refusal can say which
// part of the rule was missed rather than repeating the rule and leaving the
// reader to spot the difference.
func refuseSlug(slug string) string {
	switch {
	case slug == "":
		return "it is empty"
	case len(slug) > slugLimit:
		return fmt.Sprintf("it is %d characters and record 0014 allows at most %d", len(slug), slugLimit)
	case strings.ToLower(slug) != slug:
		return "it carries upper case, and two slugs differing only in case are one directory on a filesystem that folds case and two on one that does not"
	case !legalSlug.MatchString(slug):
		return "record 0014 allows lower case letters, digits and single hyphens, beginning and ending with a letter or a digit"
	}
	return ""
}

// refuseSlugs holds every experiment to the shape and to being the only one
// answering to its slug.
//
// WHERE THE THIRD RULE CAN AND CANNOT BITE, because a green run is otherwise
// read as more than it is. Slugs are compared with case ignored, and under the
// shape above two legal slugs can never differ by case alone, so the comparison
// bites today only on two experiments declaring the same slug exactly. The
// case-folding half is a guard for a pair where at least one side is already
// refused for its shape, and for the day the shape is argued again. That is why
// the fixture proving this rule is a pair of records declaring one slug rather
// than a pair differing in case: a pair differing in case trips two rules, and
// a fixture tripping two proves neither cleanly.
func refuseSlugs(experiments []experiment) []Refusal {
	var refusals []Refusal
	declared := make(map[string]experiment)

	for _, exp := range experiments {
		if reason := refuseSlug(exp.directory); reason != "" {
			refusals = append(refusals, Refusal{
				Property: ExperimentDirectoryIsNotALegalSlug,
				Subject:  exp.path,
				Detail:   fmt.Sprintf("its directory is named %q and %s", exp.directory, reason),
			})
		}

		if exp.declaresSlug {
			if reason := refuseSlug(exp.slug); reason != "" {
				refusals = append(refusals, Refusal{
					Property: RecordSlugIsNotALegalSlug,
					Subject:  exp.record,
					Detail:   fmt.Sprintf("its %s is %q and %s", FieldSlug, exp.slug, reason),
				})
			}
		}

		// The slug an experiment answers to is the one its record declares,
		// and the directory name where it declares none. Both are the same
		// string in a record that is in order, and the comparison has to work
		// on one that is not.
		answered := exp.directory
		if exp.declaresSlug {
			answered = exp.slug
		}
		folded := strings.ToLower(answered)
		if first, taken := declared[folded]; taken {
			refusals = append(refusals, Refusal{
				Property: TwoExperimentsShareASlug,
				Subject:  exp.path,
				Detail: fmt.Sprintf("it answers to the slug %q and so does %s, so a reader holding that slug cannot tell which of them produced a result",
					answered, first.path),
			})
			continue
		}
		declared[folded] = exp
	}

	return refusals
}

// An experiment is what the walk read about one directory under experiments/,
// reduced to what the slug rules need.
type experiment struct {
	// path is the directory, as the walk reached it, which is what a refusal
	// sends a reader to.
	path string

	// record is the path of its record, named separately because a refusal
	// about a header field sends the reader to the file rather than to the
	// directory.
	record string

	// directory is the directory's own name.
	directory string

	// slug is the Slug field the record declares, and declaresSlug says
	// whether it declared one. An absent field is never a refusal, which
	// record 0013 fixes, so the two are separate answers.
	slug         string
	declaresSlug bool
}
