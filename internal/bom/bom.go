// Package bom renders the bill of materials a release owes whoever downloads
// it, from the module set the binary itself carries.
//
// WHY IT IS A SECOND DOCUMENT AND NOT A SECTION OF THE NOTICES. The two answer
// different questions for different readers. The notices discharge a legal
// obligation and are read by a person, so they reproduce licence text in full.
// A bill of materials answers what an operator is running and is read by a
// program, so it names every component in a form something else can match
// against an advisory feed. A single document doing both would be worse at each
// of them, and the reader who needs one of the two would have to carry the
// other.
//
// WHY IT READS A BUILD RATHER THAN A FILE. The same reason internal/notices
// does, and from the same input: the toolchain records every module that went
// into a binary, inside the binary, so there is no second list for anything to
// drift against. The build type is that package's rather than a copy declared
// here. Two declarations of one input is exactly the drift that would show up
// as a notices file and a bill of materials disagreeing about what is inside
// the same binary, which is the failure both documents exist to prevent.
//
// WHAT IT WILL NOT DO IS CARRY A CLOCK OR A RANDOM IDENTIFIER. CycloneDX allows
// both a timestamp and a serial number and both are optional, so neither is
// written. Two runs from one tag have to produce one file: a release that
// publishes a checksum over this document is claiming the document is a
// function of the build, and a field that moves between two runs of the same
// build destroys that claim while looking like metadata.
//
// WHAT IT CANNOT DO. It declares a specVersion and nothing in this repository
// validates the document against the published CycloneDX schema, so what the
// suite holds is the fields this renderer writes and not conformance to that
// specification. It also does not identify a licence, for the reason
// internal/notices gives at greater length: the licence a module is under is a
// judgement no reading of its text makes reliably, and a bill of materials
// carrying a guessed identifier is worse than one carrying none, because the
// guess is machine-readable and will be believed.
package bom

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Flowfin/lab/internal/notices"
)

// SpecVersion is the CycloneDX version the document declares. It is a constant
// here rather than a literal in the render so that the string a case asserts
// and the string the document carries cannot become two strings.
const SpecVersion = "1.6"

// The properties this package can refuse. A property is the rule, named once
// here and nowhere else, so a case declaring what it expects and a refusal the
// render produced are the same string or they are not equal. It is the shape
// internal/notices, internal/contexts and internal/check already use rather
// than a fourth one.
const (
	// MainComponentHasNoReleaseVersion refuses a build whose own module
	// version is not one a reader can resolve back to a release.
	//
	// This is the refusal that earns the package. The toolchain stamps the
	// main module's version from version control, so a build made at a tag
	// carries the tag and a build made anywhere else carries a pseudo-version
	// or nothing at all. A bill of materials is published beside an artefact
	// and read later by somebody asking which version they are running, and a
	// document answering that question with a commit timestamp names no
	// release. Refusing it at the moment the release is built is the only
	// moment the answer can still be corrected.
	MainComponentHasNoReleaseVersion = "main-component-has-no-release-version"

	// DependencyHasNoVersion refuses a module the binary contains that is
	// recorded with no version.
	//
	// A component with no version cannot be matched against an advisory, and
	// matching against advisories is the whole operational reason this
	// document exists. Listing it anyway would put a row in front of a
	// scanner that the scanner has to either ignore or guess at, and both of
	// those are worse than a document that says which entry it could not
	// complete.
	DependencyHasNoVersion = "dependency-has-no-version"

	// BuildIsFromAModifiedTree refuses a build made from a working tree
	// carrying changes version control does not hold.
	//
	// It is a second property rather than a second reason under the one
	// above, because the repair is a different repair: a build that names no
	// release is repaired by tagging, and this one is repaired by committing
	// or discarding what is in the tree. A document that collapsed them would
	// send a reader to the wrong one.
	//
	// IT IS READ FROM THE BUILD SETTING AND NOT FROM THE VERSION STRING, and
	// that is the whole reason this property exists rather than being folded
	// into the version rule. The toolchain says a tree was modified twice: as
	// a setting, and by appending a suffix to the version. The suffix comes
	// after the commit, so a rule matching a version derived from a commit
	// and anchored at its end accepts a dirty one, which is what this
	// repository's own build did until a case read the answer out of a real
	// binary and printed it.
	BuildIsFromAModifiedTree = "build-is-from-a-modified-tree"
)

// pseudoVersion matches the version-control-derived version the toolchain
// writes for a build that is not at a tag: a base version, the commit time and
// the short commit.
//
// It is written out rather than inferred from the presence of a hyphen, because
// an ordinary pre-release version carries one too. Refusing v1.0.0-rc.1 as
// unresolvable would refuse a real release for looking like a build that is not
// one.
//
// THE TRAILING BUILD METADATA IS PART OF THE PATTERN AND WAS ADDED AFTER IT WAS
// MEASURED. A build from a modified tree carries a suffix after the commit, so
// a pattern anchored at the commit does not match it, and this repository built
// itself and returned a clean run over exactly that string. The modified tree
// is refused under its own property; this half is the pattern being right about
// what a version derived from a commit looks like.
var pseudoVersion = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?-[0-9]{14}-[0-9a-f]{12}(\+[0-9A-Za-z.-]+)?$`)

// A Refusal is one rule refusing one subject. It carries the subject separately
// from the detail so that a message can never be written without it.
type Refusal struct {
	Property string
	Subject  string
	Detail   string
}

// String leads with the subject, because somebody reading a red run is looking
// for the thing to open first.
func (r Refusal) String() string {
	return fmt.Sprintf("%s: %s (%s)", r.Subject, r.Detail, r.Property)
}

// A Component is one entry in the document, in the shape CycloneDX names them.
//
// The field names carry their JSON spelling rather than being renamed, so a
// reader comparing this struct against the specification is comparing the same
// words. A hyphen is not legal in a Go identifier, which is why the tag on the
// reference field is there at all and not only for case.
type Component struct {
	Type    string `json:"type"`
	BOMRef  string `json:"bom-ref"`
	Name    string `json:"name"`
	Version string `json:"version"`
	PURL    string `json:"purl"`

	// Description is written only for a component standing in for another
	// module, and says which one. A replacement is what is in the binary and
	// the module the build asked for is what a reader holding go.mod is
	// looking for, so a document showing only one of them makes one of those
	// two readers wrong.
	Description string `json:"description,omitempty"`
}

// A Property is one name and value in the metadata block. The document uses it
// for the facts about the build that CycloneDX has no field of its own for.
type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// A Metadata block says what the document is about.
type Metadata struct {
	Component  Component  `json:"component"`
	Properties []Property `json:"properties,omitempty"`
}

// A Document is the rendered bill of materials and what producing it refused.
//
// The JSON tags are the document's wire shape and the field order below is the
// order they are written in, because encoding/json writes a struct in
// declaration order and a document whose field order moved between two runs
// would break the checksum this file is published under.
type Document struct {
	BOMFormat   string      `json:"bomFormat"`
	SpecVersion string      `json:"specVersion"`
	Version     int         `json:"version"`
	Metadata    Metadata    `json:"metadata"`
	Components  []Component `json:"components"`

	// Refusals is every entry the render could not complete. It is not part
	// of the document's wire shape: a bill of materials is read by a program
	// that knows the specification, and a field this repository invented
	// would be dropped by it in silence. The refusals reach a person through
	// the command's standard error and its exit code instead.
	Refusals []Refusal `json:"-"`
}

// Properties returns the set of properties this document refused, which is what
// a case declares and what the harness compares. A property refused twice is
// one entry: a verdict is a set.
func (d Document) Properties() []string {
	seen := make(map[string]bool, len(d.Refusals))
	var props []string
	for _, refusal := range d.Refusals {
		if !seen[refusal.Property] {
			seen[refusal.Property] = true
			props = append(props, refusal.Property)
		}
	}
	sort.Strings(props)
	return props
}

// Render builds the document from the module set a binary carries.
//
// IT SORTS BY MODULE PATH AND CARRIES NO CLOCK, for the reason the package
// comment gives: the bytes this produces are a function of the build and of
// nothing else, so two runs from one tag produce one file.
func Render(build notices.Build) Document {
	document := Document{
		BOMFormat:   "CycloneDX",
		SpecVersion: SpecVersion,
		// The document's own revision, which is 1 for a document that has
		// never been superseded. It is not the version of the thing being
		// described, and the two are next to each other in the output, so the
		// distinction is written here where somebody changing one of them
		// will read it.
		Version: 1,
		// Never nil. A document with no dependencies has to serialise as an
		// empty list rather than as null, because a reader distinguishing
		// nothing in it from the field was not written gets a different answer
		// from each, and a Go nil slice writes the second one.
		Components: []Component{},
	}

	document.Metadata = Metadata{
		Component:  componentOf(build.Main, "application"),
		Properties: metadataProperties(build),
	}
	if !isReleaseVersion(build.Main.Version) {
		document.Refusals = append(document.Refusals, Refusal{
			Property: MainComponentHasNoReleaseVersion,
			Subject:  build.Main.Describe(),
			Detail: fmt.Sprintf("the toolchain stamped %q as this build's own version, which names no release, so a reader holding the artefact cannot resolve this document back to one",
				build.Main.Version),
		})
	}

	if build.Modified {
		document.Refusals = append(document.Refusals, Refusal{
			Property: BuildIsFromAModifiedTree,
			Subject:  build.Main.Describe(),
			Detail:   "the toolchain recorded that the tree this was built from carried changes version control does not hold, so the revision beside it names a commit that is not what was compiled",
		})
	}

	deps := append([]notices.Module(nil), build.Deps...)
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].Path != deps[j].Path {
			return deps[i].Path < deps[j].Path
		}
		return deps[i].Version < deps[j].Version
	})

	for _, module := range deps {
		if strings.TrimSpace(module.Version) == "" {
			document.Refusals = append(document.Refusals, Refusal{
				Property: DependencyHasNoVersion,
				Subject:  module.Describe(),
				Detail:   "the binary contains it and the module table records no version for it, so nothing reading this document can match it against an advisory",
			})
			continue
		}
		document.Components = append(document.Components, componentOf(module, "library"))
	}
	return document
}

// metadataProperties carries the facts about the build that CycloneDX has no
// field of its own for. The names are namespaced to this repository, because an
// unprefixed property name is a claim on a word the specification may give to
// something else later.
func metadataProperties(build notices.Build) []Property {
	if build.Revision == "" {
		// Said rather than left out. A document with no revision property and
		// one whose revision was dropped look identical, and the first of
		// those is an ordinary build from an archive while the second is a
		// defect.
		return []Property{{
			Name:  "lab:vcs.revision",
			Value: "the build carried no revision",
		}}
	}
	return []Property{{Name: "lab:vcs.revision", Value: build.Revision}}
}

// componentOf turns one module into one component.
func componentOf(module notices.Module, kind string) Component {
	component := Component{
		Type:    kind,
		BOMRef:  PURL(module),
		Name:    module.Path,
		Version: module.Version,
		PURL:    PURL(module),
	}
	if module.ReplacedPath != "" {
		component.Description = fmt.Sprintf("stands in for %s@%s, which is what the build asked for", module.ReplacedPath, module.ReplacedVersion)
	}
	return component
}

// PURL is the package URL for one module.
//
// WHAT IT IS AND WHAT IT IS NOT. It is built from the module path and the
// version the toolchain recorded, with the characters that are not safe in a
// URL path percent-encoded. Nothing here validates it against the package-url
// specification, and a reader who needs that guarantee has to check it rather
// than take this function's word.
//
// THE CASE OF THE MODULE PATH IS NOT FOLDED. Two module paths differing only in
// case are different modules, which is why the module cache escapes upper case
// rather than lower-casing it, and a bill of materials that folded them would
// name a module that was never in the binary. That is a deliberate departure
// from any rule asking for a lower-cased name, and it is written here rather
// than discovered by somebody comparing this output against another tool's.
func PURL(module notices.Module) string {
	return "pkg:golang/" + encodePath(module.Path) + "@" + encodeSegment(module.Version)
}

// encodePath percent-encodes a module path, leaving the separators alone so the
// result still reads as a path.
func encodePath(modulePath string) string {
	segments := strings.Split(modulePath, "/")
	for i, segment := range segments {
		segments[i] = encodeSegment(segment)
	}
	return strings.Join(segments, "/")
}

// encodeSegment percent-encodes everything outside the unreserved set. It is
// written here rather than taken from net/url, because url.PathEscape leaves
// several sub-delimiters in place and this is one place where being stricter
// than necessary costs nothing and being looser costs a document a reader has
// to guess at.
func encodeSegment(segment string) string {
	var out strings.Builder
	for _, b := range []byte(segment) {
		if unreserved(b) {
			out.WriteByte(b)
			continue
		}
		fmt.Fprintf(&out, "%%%02X", b)
	}
	return out.String()
}

// unreserved is the set of bytes that need no encoding. It is the unreserved
// set of the URI syntax plus the plus sign, which a module version carries for
// an incompatible major version and which reads wrong when encoded.
func unreserved(b byte) bool {
	switch {
	case b >= 'A' && b <= 'Z', b >= 'a' && b <= 'z', b >= '0' && b <= '9':
		return true
	case b == '-', b == '.', b == '_', b == '~', b == '+':
		return true
	}
	return false
}

// isReleaseVersion says whether the toolchain stamped a version a reader can
// resolve back to a release.
//
// Three things are not one: an empty version, the placeholder the toolchain
// writes when it knows nothing, and a version derived from a commit. All three
// leave the reader unable to say which release they are holding, which is why
// one property covers them, and the refusal quotes the string so the reader can
// tell which of the three they met.
func isReleaseVersion(version string) bool {
	version = strings.TrimSpace(version)
	if version == "" || version == "(devel)" {
		return false
	}
	return !pseudoVersion.MatchString(version)
}

// JSON renders the document.
//
// Two spaces of indentation and a trailing newline, so the published file is
// one a person can open as well as one a program can read, and so that a diff
// between two releases is a diff of the entries rather than of one long line.
func (d Document) JSON() (string, error) {
	encoded, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", fmt.Errorf("cannot render the bill of materials: %w", err)
	}
	return string(encoded) + "\n", nil
}
