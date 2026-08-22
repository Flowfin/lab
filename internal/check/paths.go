package check

import (
	"regexp"
	"strings"
)

// PathsNamedInProse returns the repository-relative paths a text names, in the
// order it names them, each once. It is the same reading the record checks do,
// exported so that a second reader of this repository's prose is one caller of
// one rule rather than a second copy of the pattern.
//
// A second copy is the failure this exists to prevent. The pattern has to know
// which leading segments are directories in this tree and which shapes are not
// paths at all, and a copy of it goes quiet on the day one of those changes in
// the original. What either caller may do with a path is its own business; what
// counts as one is decided here.
func PathsNamedInProse(text string) []string {
	return pathsNamedInProse(text)
}

// linkTarget matches the target of a markdown inline link, which is the part
// between the parentheses. It stops at whitespace and at a parenthesis, so a
// sentence carrying an ordinary parenthetical is not read as a link.
var linkTarget = regexp.MustCompile(`\]\(([^()\s]+)\)`)

// LinkTargetsWithoutADirectory returns the markdown link targets a text carries
// that name a file beside the document rather than under a directory, in the
// order it names them, each once. A caller joins each one to the directory its
// document sits in and gets a repository-relative path.
//
// This is the reading PathsNamedInProse cannot make and is not widened to make.
// That pattern requires a leading segment naming a directory of this tree,
// which is what stops it reading a capitalised word, a version string or the
// last segment of a URL as a path, and a file at the root has no such segment
// in front of it. Widening the pattern to reach one would refuse honest prose,
// and a check that refuses honest prose is a check somebody switches off.
//
// The parentheses are what carry the intent instead. A name inside a link is
// somebody saying they expect a file to be there; the same name in a sentence
// is a word. So the target of a link is resolved and a bare word is not, and
// the two readings sit side by side rather than one replacing the other.
//
// What it deliberately does not return. A target carrying a directory is left
// to the pattern above, which reads it wherever the document writes it out. A
// target with a scheme in it is somebody else's server and nothing here can say
// whether a file on one exists. A target beginning with a fragment is a place
// inside the same page and names no file at all.
func LinkTargetsWithoutADirectory(text string) []string {
	var named []string
	seen := make(map[string]bool)

	for _, match := range linkTarget.FindAllStringSubmatch(text, -1) {
		target := match[1]
		if cut := strings.IndexAny(target, "#?"); cut >= 0 {
			target = target[:cut]
		}
		if target == "" || strings.ContainsAny(target, "/:\\") || seen[target] {
			continue
		}
		seen[target] = true
		named = append(named, target)
	}
	return named
}

// LinkTargetsAboveTheirDocument returns the markdown link targets a text
// carries that step out of the document's own directory, in the order it names
// them, each once. A caller joins each one to the directory its document sits
// in and gets a repository-relative path.
//
// This is a third reading rather than either of the two above growing, and the
// reason is that both of them are already exact about a different population.
// PathsNamedInProse requires a leading segment naming a directory of this tree,
// and `..` is not one; widening it would make it read a pair of full stops in a
// sentence. LinkTargetsWithoutADirectory skips a target carrying a slash, which
// is what its name says it returns and what its own comment hands to the prose
// pattern. A target spelled `../NAME` falls between the two: it is inside
// parentheses, so the intent is there, and it carries a slash, so the reading
// that resolves intent will not touch it.
//
// The shape it is for is a document under docs/ naming a file at the root,
// which is the only way one can write that link. `docs/operator-guide.md`
// pointing at `../LICENSE` is a reader being told a file is there, and until
// this existed nothing in this tree read that link at all.
//
// WHAT THE CALLER STILL OWES. The target is relative to the document and this
// returns it as written, so a caller that resolves it against the root instead
// resolves the wrong thing. It also has to decide what to do with a target that
// climbs past the root of the tree, because this returns that one too: `../..`
// from a document one directory down names something outside the checkout, and
// whether such a file exists is not a question any reading of this tree
// answers.
func LinkTargetsAboveTheirDocument(text string) []string {
	var named []string
	seen := make(map[string]bool)

	for _, match := range linkTarget.FindAllStringSubmatch(text, -1) {
		target := match[1]
		if cut := strings.IndexAny(target, "#?"); cut >= 0 {
			target = target[:cut]
		}
		if !strings.HasPrefix(target, "../") {
			continue
		}
		// A scheme or a backslash is left alone for the same reasons the
		// reading above leaves them alone: one is somebody else's server
		// and the other is not how a path is written here.
		if strings.ContainsAny(target, ":\\") || seen[target] {
			continue
		}
		seen[target] = true
		named = append(named, target)
	}
	return named
}
