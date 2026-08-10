package check

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
