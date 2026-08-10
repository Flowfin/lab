// Package invariants holds the rules that are properties of this repository's
// tracked text and of nothing else. They are cheap to check exactly and
// expensive to leave to a reader, and until this package existed every one of
// them was a paragraph somebody had to remember.
//
// It is a separate package from the record checks on purpose. Those judge a
// tree of experiments and are pointed at a fixture tree that exists to be
// refused; these judge a repository, and the two would fight over what the root
// of the tree means. A run of the record checks against a fixture that carries
// no licence file must stay green, and a run of these against this repository
// must not.
//
// WHAT IT READS. A filesystem, from the root it is given downwards, rather than
// what git tracks. The difference is real and is the same answer the layout
// refusal already gives: a file somebody keeps beside the tree without
// committing it is scanned exactly like a committed one, and a tracked file
// deleted from the working tree is not scanned at all. For a scan whose purpose
// is to catch a string that must never be committed, reading the working tree
// is the earlier and therefore the better moment.
//
// WHAT IT DOES NOT READ. The fixtures. Record 0002 says testdata/ holds the
// runner's own fixtures and that they are fixtures rather than content, and the
// cases below have to carry a credential-shaped string and an attribution
// marker in order to prove that the legs bite. Scanning them would refuse this
// repository for holding the evidence that its checks work. The cost is stated
// rather than hidden: a real credential committed under testdata/ is not caught
// here, and a green run of this package says nothing about that directory.
package invariants

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/Flowfin/lab/internal/check"
)

// The files these legs are about, named once here so a message and a check read
// the same string.
const (
	// NoticeName is the intended-use notice. Every board in this
	// organisation carries one, and a guard that can be deleted without
	// anything noticing is a paragraph rather than a guard.
	NoticeName = "NOTICE.md"

	// ReadmeName is the first thing a visitor reads, which is why the notice
	// has to be reachable from it rather than only present in the tree.
	ReadmeName = "README.md"

	// LicenceName is the file the licence question's answer lands in. It is
	// spelled the way the hosting platform reads it, which is not this
	// repository's spelling of the word elsewhere.
	LicenceName = "LICENSE"

	// DocsDir is one half of what the path leg calls this repository's own
	// documents. The other half is the files at the root.
	DocsDir = "docs"

	// FixturesDir is the directory this package does not scan, for the reason
	// the package comment gives.
	FixturesDir = "testdata"

	// gitDir is the checkout's own machinery. It is not tracked text and
	// walking it would read every blob in the history.
	gitDir = ".git"
)

// DeclaredLicence is the licence this repository has decided on, and it is
// empty because that question is open. Entry one of the maintainer question
// issue holds it, issue #47 lands the file and the decision record once it is
// answered, and setting this string is the last line of that change.
//
// While it is empty the licence leg reports that it was not asked rather than
// passing, so a run that could not check the licence is never read as one that
// checked and was satisfied.
const DeclaredLicence = ""

// The properties this package can refuse. A property is the rule, named once
// here and nowhere else, so a case declaring what it expects and a refusal the
// scan produced are the same string or they are not equal.
const (
	// IntendedUseNoticeIsMissing refuses a tree with no intended-use notice
	// at its root. The notice is one of the guards every board here carries
	// and the only one that is a single file, which makes it the one that can
	// be removed by a tidy-up nobody reviews closely.
	IntendedUseNoticeIsMissing = "intended-use-notice-is-missing"

	// ReadmeDoesNotLinkTheIntendedUseNotice refuses a tree whose readme does
	// not point at the notice. A notice present and unreachable is the failure
	// this half exists for: the file passes a check that only asks whether it
	// is there, and no visitor ever arrives at it.
	ReadmeDoesNotLinkTheIntendedUseNotice = "readme-does-not-link-the-intended-use-notice"

	// LicenceFileDoesNotMatchTheDeclaredLicence refuses a tree that has
	// decided on a licence and whose licence file is absent or does not name
	// it. Absent and wrong are one property because the repair is one repair:
	// put the canonical text of the declared licence at the root.
	LicenceFileDoesNotMatchTheDeclaredLicence = "licence-file-does-not-match-the-declared-licence"

	// CredentialShapedStringInTrackedText refuses a credential shape in text
	// this repository tracks. What the vocabulary holds and what it cannot
	// hold is written at the vocabulary rather than here.
	CredentialShapedStringInTrackedText = "credential-shaped-string-in-tracked-text"

	// AttributionMarkerInTrackedText refuses a tool name or an attribution
	// line in tracked text. Where the origin of a passage genuinely matters it
	// is stated in the passage rather than hidden.
	AttributionMarkerInTrackedText = "attribution-marker-in-tracked-text"

	// DocumentNamesAPathThatDoesNotResolve refuses a path this repository's
	// own documents name that is not in the tree. The record checks already do
	// this for experiment records; this is the same rule over the documents a
	// visitor reads first, which is where a dead pointer does the most damage.
	DocumentNamesAPathThatDoesNotResolve = "document-names-a-path-that-does-not-resolve"
)

// A Refusal is one rule refusing one subject. It carries the subject separately
// from the detail so that a message can never be written without it.
type Refusal struct {
	Property string
	Subject  string
	Detail   string
}

// String leads with the subject, because somebody reading a red run is looking
// for the file to open first.
func (r Refusal) String() string {
	return fmt.Sprintf("%s: %s (%s)", r.Subject, r.Detail, r.Property)
}

// A Leg is one of the checks in this package and what it did. A leg that was
// not asked says so and says what asking would cost, so a run that covered less
// than the whole set cannot be read as one that covered it and found nothing.
type Leg struct {
	// Name is what the leg is called in the report.
	Name string

	// Examined is how many files or documents the leg read.
	Examined int

	// Asked is false where the leg could not run at all.
	Asked bool

	// NotAsked says why, and what asking would cost. It is empty exactly when
	// Asked is true.
	NotAsked string
}

// Report is what one scan examined.
type Report struct {
	// Root is the directory the scan started from, as it was given.
	Root string

	// TextFiles is the number of text files the scan read. Zero is reported
	// rather than hidden: a scan that read nothing and a scan that read
	// everything and found nothing look identical without it.
	TextFiles int

	// Legs is every leg, in the order they ran, asked or not.
	Legs []Leg

	// Refusals is what the scan refused, in the order it reached them.
	Refusals []Refusal
}

// Properties returns the set of properties this report refused, which is what a
// case declares and what the harness compares. A property refused twice is one
// entry: a verdict is a set.
func (r Report) Properties() []string {
	seen := make(map[string]bool, len(r.Refusals))
	var props []string
	for _, refusal := range r.Refusals {
		if !seen[refusal.Property] {
			seen[refusal.Property] = true
			props = append(props, refusal.Property)
		}
	}
	return props
}

// Scan reads the tree at root and returns what it found. The declared licence
// is a parameter rather than the constant above so that the licence leg can be
// proved by a case: the constant is what this repository passes, and a fixture
// passes whatever it needs to trip the rule.
//
// The error it returns means the scan could not be made at all, which is a
// different outcome from a scan that completed and refused something.
func Scan(root, declaredLicence string) (Report, error) {
	rep := Report{Root: root}

	info, err := os.Stat(root)
	if err != nil {
		return rep, fmt.Errorf("cannot read %s: %w", root, err)
	}
	if !info.IsDir() {
		return rep, fmt.Errorf("%s is not a directory", root)
	}

	texts, err := readTextFiles(root)
	if err != nil {
		return rep, err
	}
	rep.TextFiles = len(texts)

	leg, refusals := noticeLeg(root, texts)
	rep.Legs = append(rep.Legs, leg)
	rep.Refusals = append(rep.Refusals, refusals...)

	leg, refusals = licenceLeg(root, declaredLicence)
	rep.Legs = append(rep.Legs, leg)
	rep.Refusals = append(rep.Refusals, refusals...)

	leg, refusals = credentialLeg(texts)
	rep.Legs = append(rep.Legs, leg)
	rep.Refusals = append(rep.Refusals, refusals...)

	leg, refusals = attributionLeg(texts)
	rep.Legs = append(rep.Legs, leg)
	rep.Refusals = append(rep.Refusals, refusals...)

	leg, refusals = pathsLeg(root, texts)
	rep.Legs = append(rep.Legs, leg)
	rep.Refusals = append(rep.Refusals, refusals...)

	return rep, nil
}

// textFile is one file the scan read, with the path a refusal names and the
// path the leg reasons about.
type textFile struct {
	// path is the file as the walk reached it, which is what a reader opens.
	path string

	// relative is the path from the root, in slashes, which is what the
	// document leg matches against.
	relative string

	// text is the whole file.
	text string
}

// readTextFiles walks the tree and returns every file that is text. A file that
// is not valid UTF-8, or that holds a null byte, is not text and is not
// scanned: a vocabulary match inside a compiled object or an image is not a
// string anybody wrote, and reporting one would train a reader to ignore this
// check.
func readTextFiles(root string) ([]textFile, error) {
	var texts []textFile

	// filepath.WalkDir does not follow symbolic links, so a link pointing out
	// of the tree is reported as a link and never descended into.
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && (entry.Name() == gitDir || entry.Name() == FixturesDir) {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", path, err)
		}
		if strings.IndexByte(string(data), 0) >= 0 || !utf8.Valid(data) {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("cannot place %s inside %s: %w", path, root, err)
		}
		texts = append(texts, textFile{
			path:     path,
			relative: filepath.ToSlash(relative),
			text:     string(data),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s: %w", root, err)
	}

	sort.Slice(texts, func(i, j int) bool { return texts[i].relative < texts[j].relative })
	return texts, nil
}

// noticeLeg holds the intended-use notice to being present and reachable. The
// two halves are two properties because the repairs are different: one writes a
// file and the other adds a line to the readme.
func noticeLeg(root string, texts []textFile) (Leg, []Refusal) {
	leg := Leg{Name: "the intended-use notice", Asked: true, Examined: 2}
	var refusals []Refusal

	if _, err := os.Stat(filepath.Join(root, NoticeName)); err != nil {
		refusals = append(refusals, Refusal{
			Property: IntendedUseNoticeIsMissing,
			Subject:  filepath.Join(root, NoticeName),
			Detail:   fmt.Sprintf("there is no %s at the root of this tree, so the notice every board here carries is not in it", NoticeName),
		})
	}

	readme, ok := fileAt(texts, ReadmeName)
	if !ok {
		// A tree with no readme is refused by the half that can act on it.
		// Saying the readme does not link the notice when there is no readme
		// names the wrong repair.
		refusals = append(refusals, Refusal{
			Property: ReadmeDoesNotLinkTheIntendedUseNotice,
			Subject:  filepath.Join(root, ReadmeName),
			Detail:   fmt.Sprintf("there is no %s, so nothing points a visitor at %s", ReadmeName, NoticeName),
		})
		return leg, refusals
	}
	if !strings.Contains(readme.text, NoticeName) {
		refusals = append(refusals, Refusal{
			Property: ReadmeDoesNotLinkTheIntendedUseNotice,
			Subject:  readme.path,
			Detail:   fmt.Sprintf("it never names %s, so the notice is in the tree and no visitor arrives at it", NoticeName),
		})
	}
	return leg, refusals
}

// licenceLeg holds the licence file to the licence this repository has decided
// on. It is the leg that reports it was not asked, and that is its ordinary
// state today rather than a failure.
//
// WHAT IT CANNOT SAY once it is asked. It reads whether the file names the
// declared licence, not whether the text is the canonical text of it. A licence
// retyped from memory or trimmed for length names the same licence and is a
// different licence, and no reading of this tree separates the two. Whoever
// lands the file takes the text from its canonical source, which is issue #47's
// sentence rather than this check's.
func licenceLeg(root, declaredLicence string) (Leg, []Refusal) {
	if declaredLicence == "" {
		return Leg{
			Name: "the licence",
			NotAsked: fmt.Sprintf("no licence is declared, so there is nothing to compare %s against. "+
				"asking costs answering entry one of the maintainer question issue and landing the file, "+
				"the repository metadata and the decision record that names it, which is issue #47", LicenceName),
		}, nil
	}

	leg := Leg{Name: "the licence", Asked: true, Examined: 1}
	path := filepath.Join(root, LicenceName)
	data, err := os.ReadFile(path)
	if err != nil {
		return leg, []Refusal{{
			Property: LicenceFileDoesNotMatchTheDeclaredLicence,
			Subject:  path,
			Detail:   fmt.Sprintf("this repository declares %s and there is no %s to read", declaredLicence, LicenceName),
		}}
	}
	if !strings.Contains(string(data), declaredLicence) {
		return leg, []Refusal{{
			Property: LicenceFileDoesNotMatchTheDeclaredLicence,
			Subject:  path,
			Detail:   fmt.Sprintf("this repository declares %s and the file never names it", declaredLicence),
		}}
	}
	return leg, nil
}

// hit is one vocabulary entry found in one file: what matched, where, and in
// which file. It carries no copy of the matched text, because a report that
// pastes the credential it found has copied it somewhere new.
type hit struct {
	path string
	name string
	line int
}

// findVocabulary reads every text file for one vocabulary. It is one function
// for both rather than two that have to be kept saying the same thing.
//
// It returns findings rather than refusals so that the property each leg names
// is written at the leg, as a constant, where the requirement that every
// refusable property owns a case can read it out of the source. A property
// passed in as a parameter would be invisible there, and a leg nothing can see
// is a leg nothing can require a fixture for.
func findVocabulary(vocabulary []shape, texts []textFile) []hit {
	var hits []hit
	for _, file := range texts {
		for _, entry := range vocabulary {
			match := entry.pattern.FindStringIndex(file.text)
			if match == nil {
				continue
			}
			hits = append(hits, hit{
				path: file.path,
				name: entry.name,
				line: lineOf(file.text, match[0]),
			})
		}
	}
	return hits
}

// credentialLeg refuses a credential shape in tracked text.
func credentialLeg(texts []textFile) (Leg, []Refusal) {
	leg := Leg{Name: "credential shapes in tracked text", Asked: true, Examined: len(texts)}
	var refusals []Refusal
	for _, found := range findVocabulary(credentialShapes, texts) {
		refusals = append(refusals, Refusal{
			Property: CredentialShapedStringInTrackedText,
			Subject:  found.path,
			Detail:   fmt.Sprintf("%s at line %d, and this text is tracked", found.name, found.line),
		})
	}
	return leg, refusals
}

// attributionLeg refuses a tool name or an attribution line in tracked text.
func attributionLeg(texts []textFile) (Leg, []Refusal) {
	leg := Leg{Name: "attribution markers in tracked text", Asked: true, Examined: len(texts)}
	var refusals []Refusal
	for _, found := range findVocabulary(attributionMarkers, texts) {
		refusals = append(refusals, Refusal{
			Property: AttributionMarkerInTrackedText,
			Subject:  found.path,
			Detail:   fmt.Sprintf("%s at line %d, and this text is tracked", found.name, found.line),
		})
	}
	return leg, refusals
}

// lineOf turns a byte offset into a line number, so a refusal sends the reader
// to a line rather than to a file. The matched text is deliberately not quoted:
// a report that pastes the credential it found has copied it somewhere new.
func lineOf(text string, offset int) int {
	return 1 + strings.Count(text[:offset], "\n")
}

// pathsLeg holds this repository's own documents to the paths they name. What
// counts as one of those documents is the files at the root and everything
// under docs/, which is what a visitor reads before anything else and where a
// dead pointer costs the most.
//
// It reads paths the same way the record checks do, through one exported
// function in the record package, rather than through a second copy of the
// pattern. Two copies of a rule drift, and the one that drifts is the copy
// nobody is looking at.
func pathsLeg(root string, texts []textFile) (Leg, []Refusal) {
	var documents []textFile
	for _, file := range texts {
		if isOwnDocument(file.relative) {
			documents = append(documents, file)
		}
	}

	leg := Leg{Name: "paths this repository's own documents name", Asked: true, Examined: len(documents)}
	var refusals []Refusal

	for _, document := range documents {
		for _, named := range check.PathsNamedInProse(document.text) {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(named))); err == nil {
				continue
			}
			refusals = append(refusals, Refusal{
				Property: DocumentNamesAPathThatDoesNotResolve,
				Subject:  document.path,
				Detail:   fmt.Sprintf("it names %s, which is not in this tree", named),
			})
		}
	}
	return leg, refusals
}

// isOwnDocument says whether a path is one of this repository's own documents.
// A file at the root is one; anything under docs/ is one; nothing else is,
// because a path named in a comment inside the runner is held by the compiler
// and by the record checks rather than by this leg.
func isOwnDocument(relative string) bool {
	if strings.HasPrefix(relative, DocsDir+"/") {
		return true
	}
	return !strings.Contains(relative, "/")
}

// fileAt returns the text file at a repository-relative path.
func fileAt(texts []textFile, relative string) (textFile, bool) {
	for _, file := range texts {
		if file.relative == relative {
			return file, true
		}
	}
	return textFile{}, false
}

// String writes what the scan examined, in the order a reader needs it: where
// it looked, what each leg covered or why it did not run, and only then what it
// refused.
func (r Report) String() string {
	out := fmt.Sprintf("examined %s\n", r.Root)
	out += fmt.Sprintf("%d text %s read\n", r.TextFiles, plural(r.TextFiles, "file", "files"))
	for _, leg := range r.Legs {
		if !leg.Asked {
			out += fmt.Sprintf("  %s: NOT ASKED, %s\n", leg.Name, leg.NotAsked)
			continue
		}
		out += fmt.Sprintf("  %s: %d examined\n", leg.Name, leg.Examined)
	}
	out += fmt.Sprintf("%d refused\n", len(r.Refusals))
	for _, refusal := range r.Refusals {
		out += fmt.Sprintf("  %s\n", refusal)
	}
	return out
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
