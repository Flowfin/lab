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

	// RecordSlugDisagreesWithItsDirectory refuses a record whose Slug names a
	// directory other than the one the record sits in. The slug is how a
	// reader gets from a quoted result back to the experiment that produced
	// it, and the promotion section makes that walk matter to somebody on
	// another board. A disagreement breaks the walk silently, because both
	// halves exist and only their agreement is missing.
	//
	// The message names both strings. Which of the two is wrong decides the
	// repair, and nothing in the tree can tell whether the field was mistyped
	// or the directory was misnamed.
	RecordSlugDisagreesWithItsDirectory = "record-slug-disagrees-with-its-directory"
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

// refuseSlugs holds every experiment to the shape and holds a record's Slug to
// naming the directory the record sits in.
//
// WHAT THE THIRD RULE REPLACED. Until this function was rewritten, the third
// rule here was two-experiments-share-a-slug, which refused two experiments
// declaring one slug once case was ignored. It is retired rather than kept
// beside the agreement rule, because the agreement rule subsumes it: two
// directories under experiments/ cannot share a name, so two experiments can
// only answer to one slug when at least one of their records declares a slug
// that is not its own directory's name, which is what this refuses. Keeping
// both would have meant every tree that tripped the sharing rule tripping the
// agreement rule as well, and a fixture tripping two rules proves neither
// cleanly. The retirement is deliberate and by name; issue #54 is where it was
// argued and decided.
//
// WHERE THE SUBSUMPTION STOPS, because a green run is otherwise read as more
// than it is. The comparison below is between two slugs, so it is made only
// where the directory name and the declared field are both legal slugs. Where
// either is not, the tree is already refused for that shape and this rule says
// nothing, so a pair answering to one slug from a directory that is not a slug
// is refused for the directory's name and never for the sharing. The tree is
// red either way and the message points somewhere else, which is the whole of
// what the retirement cost.
//
// Case is not folded here and the retired rule folded it. Under the shape
// above a legal slug carries no upper case, so two legal slugs can never differ
// by case alone, and an exact comparison between two strings that have both
// passed refuseSlug is the same comparison a folded one would make. Folding
// would only reach a pair where one side is already refused for its shape.
func refuseSlugs(experiments []experiment) []Refusal {
	var refusals []Refusal

	for _, exp := range experiments {
		directoryReason := refuseSlug(exp.directory)
		if directoryReason != "" {
			refusals = append(refusals, Refusal{
				Property: ExperimentDirectoryIsNotALegalSlug,
				Subject:  exp.path,
				Detail:   fmt.Sprintf("its directory is named %q and %s", exp.directory, directoryReason),
			})
		}

		// An absent field is never a refusal, which record 0013 fixes, and a
		// record declaring none answers to its directory name and cannot
		// disagree with it.
		if !exp.declaresSlug {
			continue
		}

		slugReason := refuseSlug(exp.slug)
		if slugReason != "" {
			refusals = append(refusals, Refusal{
				Property: RecordSlugIsNotALegalSlug,
				Subject:  exp.record,
				Detail:   fmt.Sprintf("its %s is %q and %s", FieldSlug, exp.slug, slugReason),
			})
		}

		if directoryReason != "" || slugReason != "" {
			continue
		}

		if exp.slug != exp.directory {
			refusals = append(refusals, Refusal{
				Property: RecordSlugDisagreesWithItsDirectory,
				Subject:  exp.record,
				Detail: fmt.Sprintf("its %s is %q and it sits in a directory named %q, so a reader walking back from the slug reaches another experiment or nothing",
					FieldSlug, exp.slug, exp.directory),
			})
		}
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
