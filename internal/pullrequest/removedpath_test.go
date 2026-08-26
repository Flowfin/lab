package pullrequest

import (
	"strings"
	"testing"
)

// The cases over a path the range takes out of the tree. They are returned into
// the one table in pullrequest_test.go rather than run from a second harness, so
// the proof that every property has a case reads them too.
//
// Every case here is one change away from a case that refuses nothing, and the
// one change is a sentence in the body rather than a file put back. That is the
// whole shape of the rule: removing stays allowed, removing in silence does not.
func removedPathJudgeCases() []judgeCase {
	const gone = "docs/quality-parity.md"
	const alsoGone = "internal/contexts/contexts_test.go"

	return []judgeCase{
		{
			name: "a change that removes a tracked path and never names it",
			change: func() Change {
				c := clean()
				c.Files = []File{
					{Path: ".github/workflows/zizmor.yml"},
					{Path: gone, Gone: true},
				}
				return c
			}(),
			// The incident this rule comes from, reduced to one path. The body
			// described something else, which is what its author believed the
			// change did, and a reader had no line to disagree with.
			want: []string{PathRemovedWithoutBeingNamed},
		},
		{
			name: "a change that removes a tracked path and names it in the body",
			change: func() Change {
				c := clean()
				c.Body = "This closes #24. It removes " + gone + ", whose two sections moved into the operator guide."
				c.Files = []File{
					{Path: ".github/workflows/zizmor.yml"},
					{Path: gone, Gone: true},
				}
				return c
			}(),
			// The near miss, and the one field that separates it from the case
			// above is the body. Nothing about the diff changed.
		},
		{
			name: "a change that removes two paths and names only one",
			change: func() Change {
				c := clean()
				c.Body = "This closes #24. It removes " + gone + " because the document moved."
				c.Files = []File{
					{Path: gone, Gone: true},
					{Path: alsoGone, Gone: true},
				}
				return c
			}(),
			// One refusal per path rather than one for the range, because the
			// repair is a sentence about a particular file and a reader given
			// one line saying something was removed has to go and find which.
			want: []string{PathRemovedWithoutBeingNamed},
		},
		{
			name: "a path that git reported as moved",
			change: func() Change {
				c := clean()
				c.Files = []File{
					{Path: "docs/parity.md", Gone: true},
					{Path: gone, From: "docs/parity.md"},
				}
				return c
			}(),
			// The boundary. The entry carries both paths, so the change says
			// where the content went and the diff is not silent about it, which
			// is the thing this rule is against.
		},
		{
			name: "a rename nothing reported as one",
			change: func() Change {
				c := clean()
				c.Files = []File{
					{Path: "docs/parity.md", Gone: true},
					{Path: gone},
				}
				return c
			}(),
			// The residual named at the rule, held in a case rather than only
			// in a comment. Where a move is not reported as one, and a rename
			// made together with a rewrite is where it will not be, the old
			// path is an ordinary removal here and has to be named.
			want: []string{PathRemovedWithoutBeingNamed},
		},
		{
			name: "a change that reads no body and removes a path",
			change: func() Change {
				c := clean()
				c.Body = ""
				c.BodyRead = false
				c.Files = []File{{Path: gone, Gone: true}}
				return c
			}(),
			// A run that read no body has not seen a removal go unnamed, and
			// saying so is different from passing. The issue rule skips on the
			// same input for the same reason.
			skips: []string{BodyNamesNoIssue, PathRemovedWithoutBeingNamed},
		},
	}
}

// TestARemovalRefusalNamesThePathAndTheRepair holds the message to the two
// things whoever hit it needs. Which file went, because a body has to name it
// and the author is about to type it; and that the repair is a sentence rather
// than putting the file back, because a rule that reads as forbidding removals
// is a rule somebody argues with instead of satisfying.
func TestARemovalRefusalNamesThePathAndTheRepair(t *testing.T) {
	const gone = "docs/quality-parity.md"
	change := clean()
	change.Files = []File{{Path: gone, Gone: true}}

	verdict := Judge(change)
	if len(verdict.Refusals) != 1 {
		t.Fatalf("expected one refusal, got %d", len(verdict.Refusals))
	}
	refusal := verdict.Refusals[0]
	if refusal.Property != PathRemovedWithoutBeingNamed {
		t.Fatalf("expected %s, got %s", PathRemovedWithoutBeingNamed, refusal.Property)
	}
	if refusal.Subject != gone {
		t.Errorf("the refusal names %q rather than the path that went", refusal.Subject)
	}
	if !strings.Contains(refusal.Detail, gone) {
		t.Errorf("the detail never names the path the author has to type: %q", refusal.Detail)
	}
	if !strings.Contains(refusal.Detail, "Removing stays allowed") {
		t.Errorf("the detail does not say the repair is a sentence rather than putting the file back: %q", refusal.Detail)
	}
}

// TestNamingOneRemovedPathDoesNotCoverAnother is the near miss that a rule
// written as "the body mentions a removal" would pass. A body saying one file
// went is not a body saying two did, and the failure this rule is against is
// exactly a change that removed more than its author knew about.
func TestNamingOneRemovedPathDoesNotCoverAnother(t *testing.T) {
	const named = "docs/quality-parity.md"
	const unnamed = "LICENSE"

	change := clean()
	change.Body = "This closes #24 and removes " + named + "."
	change.Files = []File{{Path: named, Gone: true}, {Path: unnamed, Gone: true}}

	verdict := Judge(change)
	if len(verdict.Refusals) != 1 {
		t.Fatalf("expected exactly one refusal, got %d", len(verdict.Refusals))
	}
	if verdict.Refusals[0].Subject != unnamed {
		t.Errorf("the refusal names %q rather than the path the body left out", verdict.Refusals[0].Subject)
	}
}
