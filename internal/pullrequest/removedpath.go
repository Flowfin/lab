package pullrequest

// The rule the incident in issue #155 asked for, which is about any tracked
// path rather than about a record.
//
// One merge on the default branch removed landed work on seven paths under a
// message describing a change to how one workflow pin is commented. A licence,
// a guard and four hundred lines of a document went, nothing was red, and the
// suite stayed green afterwards because almost everything removed was prose and
// the one test that went with it was the only thing reading the code path it
// covered.
//
// WHAT EVERY OTHER RULE HERE READS IS WHAT THE TREE BECAME. The record checks
// walk experiments/, the invariants read tracked text, the formatting rules read
// the bytes around the words, and a merge whose result compiles and passes says
// nothing about what it replaced. This package holds both ends of the range, so
// it is the only place a removal is visible at all.
//
// REMOVING STAYS ALLOWED. What ends is removing in silence. The refusal is
// against a body that does not mention the path, and the repair is one sentence
// in the body rather than putting the file back.

import (
	"fmt"
	"strings"
)

// PathRemovedWithoutBeingNamed refuses a change that takes a tracked path out
// of the tree without naming that path in the pull-request body.
//
// The failure is what the incident actually looked like rather than a
// hypothetical. Nobody decided to delete a licence. A branch was cut from an
// older state of the default branch and pushed on top of a newer one, the merge
// replaced seven paths with what they had said earlier, and the body of the
// change described something else entirely because that is what its author
// believed it did. A reader had no line to disagree with.
//
// What the refusal converts is exactly that: a removal nobody wrote down
// becomes a removal somebody had to write down. It cannot ask whether the
// sentence is true, and the residual is written at judgeRemovedPaths below.
const PathRemovedWithoutBeingNamed = "path-removed-without-being-named"

// judgeRemovedPaths holds every path the range takes out of the tree to being
// named in the pull-request body.
//
// THE BOUNDARIES, WRITTEN HERE RATHER THAN DISCOVERED.
//
// It reads whether the body carries the path, and never whether what the body
// says about it is true or even that it is about the removal at all. A body
// that names the path while claiming the opposite passes, and so does one that
// names it for an unrelated reason. That bound is the same one the issue rule
// above carries: what a rule reading a string can separate is the case where
// nothing was written from the case where something was, and a reviewer is what
// stands behind the rest. It is worth having anyway, because the case where
// nothing was written is the case that happened.
//
// The source path of a move is not judged. git reports a rename as one entry
// carrying both paths, so the change says where the content went and a reader
// of the diff is not left guessing, which is the thing this rule is against.
// Where a move is not reported as one, and a rename made together with a
// rewrite is where it will not be, the old path is an ordinary removal here and
// has to be named. That is red for a reason a reader can act on rather than
// silently narrower than it looks.
//
// A path added and removed inside one branch never reaches this rule, because
// the range is read from base to head and the file is not in that diff at all.
//
// WHAT IT CANNOT SEE is a removal that arrives without passing this gate: a
// direct push, an edit made through the web interface, or history rewritten on
// the branch. The first is refused by the ruleset on the default branch, which
// requires a pull request, and the others are named here so that a green run is
// not read as covering them.
func judgeRemovedPaths(change Change) Verdict {
	if !change.FilesRead {
		return Verdict{Skips: []Skip{{
			Rule: PathRemovedWithoutBeingNamed,
			Why:  "this run was given no changed paths, so nothing was read that could have been removed",
		}}}
	}
	if !change.BodyRead {
		return Verdict{Skips: []Skip{{
			Rule: PathRemovedWithoutBeingNamed,
			Why:  "this run was given no pull-request body, so nothing was read to look for a removed path in",
		}}}
	}

	moved := renames(change)
	var verdict Verdict
	for _, file := range change.Files {
		if !file.Gone {
			continue
		}
		if _, isMove := moved[file.Path]; isMove {
			continue
		}
		if namesPath(change.Body, file.Path) {
			continue
		}
		verdict.Refusals = append(verdict.Refusals, Refusal{
			Property: PathRemovedWithoutBeingNamed,
			Subject:  file.Path,
			Detail:   fmt.Sprintf("this change takes it out of the tree and the pull-request body never names it, so a reader has nothing to disagree with. Removing stays allowed: name %s in the body and say why it goes", file.Path),
		})
	}
	return verdict
}

// namesPath says whether a body carries a path.
//
// The comparison is a plain substring and that is a decision rather than the
// easy thing. A path written inside a link, inside backticks, at the end of a
// sentence or inside a longer path all carry the same bytes, and a rule that
// tried to be cleverer about the surroundings would refuse a body that names
// the file perfectly well. What it costs is that a body naming a longer path
// which happens to end in this one satisfies the rule for both, which is a
// direction that lets honest work through rather than one that hides a removal:
// the longer path is the file beside it, and a reader who reaches either
// sentence is reading about the directory that went.
func namesPath(body, path string) bool {
	return path != "" && strings.Contains(body, path)
}
