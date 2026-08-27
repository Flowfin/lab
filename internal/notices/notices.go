// Package notices renders the third-party notices a binary owes whoever
// downloads it, from the module set the binary itself carries.
//
// WHY IT READS A BUILD RATHER THAN A FILE. A list of dependencies somebody
// maintains is correct on the day it is written and wrong shortly afterwards,
// and it is wrong in the direction that matters: a module present in the binary
// and absent from the list. The toolchain already records every module that went
// into a binary, inside the binary, so that record is the input here and there is
// no second list for anything to drift against.
//
// WHAT IT WILL NOT DO IS NAME A LICENCE WITHOUT CARRYING ITS TEXT. The obligation
// most licences impose is to supply the text, and a line saying which licence a
// module is under does not discharge it. A module whose text cannot be read is
// refused rather than listed with a gap, because a notices file that quietly
// omits one is worse than one that is obviously incomplete: it looks finished.
//
// WHAT IT CANNOT DO. It does not identify a licence. It reproduces whatever text
// the module shipped under a licence filename and takes no view on which licence
// that is, whether the module was entitled to offer it, or whether reproducing it
// is sufficient for that licence. Reading a name out of the text and reporting it
// as the module's licence would be a claim this package cannot support, and the
// document it renders says so where a reader will see it rather than only here.
package notices

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// The properties this package can refuse. A property is the rule, named once
// here and nowhere else, so a case declaring what it expects and a refusal the
// render produced are the same string or they are not equal. It is the shape
// internal/invariants and internal/check already use rather than a third one.
const (
	// DependencyHasNoLicenceText refuses a module the binary contains whose
	// licence text could not be read. Absent, unreadable and empty are one
	// property because the repair is one repair: supply the text or do not
	// ship the module. Splitting them would put three rows in front of a
	// reader whose next action is the same in all three cases.
	DependencyHasNoLicenceText = "dependency-has-no-licence-text"
)

// A Module is one entry in the module set a binary carries.
type Module struct {
	// Path is the module path as the toolchain records it.
	Path string

	// Version is the version that went in.
	Version string

	// ReplacedPath and ReplacedVersion are what this module stands in for,
	// and they are empty for a module that replaced nothing.
	//
	// A replacement is carried rather than flattened because the two are
	// different statements. The code in the binary is the replacement's and
	// its licence is the one that travels with the binary; the module the
	// build asked for is what a reader comparing this file against go.mod is
	// holding. A document showing only one of them makes one of those two
	// readers wrong.
	ReplacedPath    string
	ReplacedVersion string
}

// A Build is the module set of one binary, as that binary records it.
type Build struct {
	// Main is the module the binary was built from.
	Main Module

	// Revision is the commit the build was made at, empty where the build
	// carried none. It is reported rather than required: a binary built from
	// an archive rather than a checkout has no revision, and refusing that
	// would refuse a legitimate build for a fact about how it was fetched.
	Revision string

	// Deps is every module the binary contains beyond the main one.
	Deps []Module

	// Modified says the build was made from a working tree carrying changes
	// the version control system does not hold. The toolchain records it as a
	// build setting, so it is a fact about the binary in exactly the way the
	// module set is, and it is carried here rather than derived from the
	// version string: a version says the tree was dirty by appending a
	// suffix, which is a spelling, and this is the answer.
	Modified bool
}

// A Licence is the text one module shipped, and the filename it shipped it
// under. The filename is carried because it is evidence: a reader asking where
// a paragraph in this document came from is asking for a path, and a document
// naming only the module answers a different question.
type Licence struct {
	// File is the path the text was read from, relative to the module root.
	File string

	// Text is what that file held, verbatim.
	Text string
}

// A Source resolves a module to the licence it shipped. It is an interface so
// that the render can be proved against a directory a case wrote out in full,
// rather than against whichever module cache the machine running the suite
// happens to have populated.
type Source interface {
	// Licence returns the text the module shipped. The error says why it
	// could not be read, and it reaches the reader through the refusal rather
	// than stopping the render: one unreadable module must not cost the
	// document the other twenty entries.
	Licence(Module) (Licence, error)
}

// A Refusal is one rule refusing one subject. It carries the subject separately
// from the detail so that a message can never be written without it, which is
// the shape internal/invariants already carries.
type Refusal struct {
	Property string
	Subject  string
	Detail   string
}

// String leads with the subject, because somebody reading a red run is looking
// for the module to open first.
func (r Refusal) String() string {
	return fmt.Sprintf("%s: %s (%s)", r.Subject, r.Detail, r.Property)
}

// An Entry is one module as the document reports it.
type Entry struct {
	Module  Module
	Licence Licence
}

// A Document is the rendered notices and what producing it refused.
type Document struct {
	// Build is what the document is about.
	Build Build

	// Entries is one per dependency whose licence text was read, in module
	// path order.
	Entries []Entry

	// Refusals is every module whose text was not, in the same order.
	Refusals []Refusal
}

// Properties returns the set of properties this document refused, which is what
// a case declares and what the harness compares. A property refused twice is one
// entry: a verdict is a set.
func (d Document) Properties() []string {
	seen := make(map[string]bool, len(d.Refusals))
	var props []string
	for _, refusal := range d.Refusals {
		if !seen[refusal.Property] {
			seen[refusal.Property] = true
			props = append(props, refusal.Property)
		}
	}
	return props
}

// Render reads every dependency's licence out of the source and returns the
// document, with a refusal for each module whose text could not be read.
//
// IT SORTS BY MODULE PATH AND CARRIES NO CLOCK. The bytes this produces are a
// function of the build and of the texts the modules shipped, and of nothing
// else, so two runs from one tag produce one file. A document carrying the time
// it was made would differ between two such runs, and then a checksum over it
// could no longer separate a release built from different source from one built
// twice.
func Render(build Build, source Source) Document {
	document := Document{Build: build}

	deps := append([]Module(nil), build.Deps...)
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].Path != deps[j].Path {
			return deps[i].Path < deps[j].Path
		}
		return deps[i].Version < deps[j].Version
	})

	for _, module := range deps {
		licence, err := source.Licence(module)
		if err != nil {
			document.Refusals = append(document.Refusals, Refusal{
				Property: DependencyHasNoLicenceText,
				Subject:  module.Describe(),
				Detail: fmt.Sprintf("the binary contains it and its licence text could not be read (%v), so this document cannot supply the text that licence asks to be supplied",
					err),
			})
			continue
		}
		document.Entries = append(document.Entries, Entry{Module: module, Licence: licence})
	}
	return document
}

// Describe is how a module is named in a refusal and in the document. A
// replacement says both halves, because a reader holding go.mod is looking for
// the module that was asked for and a reader asking what is in the binary is
// looking for the one that arrived.
func (m Module) Describe() string {
	if m.ReplacedPath == "" {
		return m.Path + "@" + m.Version
	}
	return fmt.Sprintf("%s@%s, which replaces %s@%s", m.Path, m.Version, m.ReplacedPath, m.ReplacedVersion)
}

// Text renders the notices document.
//
// THE ZERO-DEPENDENCY CASE IS A SENTENCE AND NOT AN EMPTY FILE. A binary
// containing no third-party module still owes its recipient the statement that
// there is nothing to disclose, and an empty file cannot be told from a run that
// failed to write one.
func (d Document) Text() string {
	var out strings.Builder

	out.WriteString("# Third-party notices\n\n")
	fmt.Fprintf(&out, "These notices are for %s.\n", d.Build.Main.Describe())
	if d.Build.Revision != "" {
		fmt.Fprintf(&out, "Built at revision %s.\n", d.Build.Revision)
	} else {
		out.WriteString("The build carried no revision, so this document names none.\n")
	}
	out.WriteString("\n")
	out.WriteString(preamble)
	out.WriteString("\n")

	if len(d.Entries) == 0 && len(d.Refusals) == 0 {
		out.WriteString("## Nothing to disclose\n\n")
		out.WriteString("This binary contains no third-party module. That is a result this run\nproduced rather than a section nobody filled in.\n")
		return out.String()
	}

	if len(d.Entries) > 0 {
		out.WriteString("## What the binary contains\n\n")
		fmt.Fprintf(&out, "%s, in module path order.\n\n", plural(len(d.Entries), "module", "modules"))
	}
	for _, entry := range d.Entries {
		fmt.Fprintf(&out, "### %s\n\n", entry.Module.Describe())
		fmt.Fprintf(&out, "The text below is %s as that module shipped it.\n\n", entry.Licence.File)
		out.WriteString(fence(entry.Licence.Text))
		out.WriteString("\n")
	}

	if len(d.Refusals) > 0 {
		out.WriteString("## What this document could not supply\n\n")
		out.WriteString("Each line below is a module the binary contains whose licence text was not\nread. This document is incomplete by exactly these entries.\n\n")
		for _, refusal := range d.Refusals {
			fmt.Fprintf(&out, "- %s\n", refusal.String())
		}
	}
	return out.String()
}

// preamble is the part of the document that is the same whatever the build. It
// is here rather than inline so that the sentence bounding what this package
// claims sits next to the code that could not make a larger claim.
const preamble = `This file is generated from the module set the binary records, and not from a
list anybody maintains. What it lists is every module the binary contains,
the version that went in, and the licence text that module shipped, reproduced
in full because supplying the text is what most licences ask for and a link
is not a copy.

It does not say which licence a module is under. It reproduces the file the
module shipped and takes no view on what that file is, which is a judgement no
reading of the text makes reliably. Anybody who needs that answer reads the
text below rather than a label this document would have guessed.
`

// fence wraps a licence text so that it survives being read as a document. The
// text is somebody else's and is reproduced verbatim, so the fence is chosen to
// be longer than any run of backticks inside it rather than assumed to be
// three: a licence carrying a fenced block would otherwise end the block early
// and the rest of it would be read as prose.
func fence(text string) string {
	longest, current := 0, 0
	for _, r := range text {
		if r != '`' {
			current = 0
			continue
		}
		current++
		longest = max(longest, current)
	}
	marker := strings.Repeat("`", max(3, longest+1))

	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return marker + "text\n" + text + marker + "\n"
}

// plural writes a count with the word that agrees with it. A heading reading
// "1 modules" is the kind of thing a reader stops on, and stopping is attention
// spent on the document instead of on what it says.
func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// EscapePath is the module cache's spelling of a module path. The cache cannot
// use the path as written, because two module paths differing only in case are
// different modules and a case-insensitive filesystem would collapse them, so
// every upper-case letter is written as an exclamation mark and its lower-case
// form.
//
// It is here rather than in the cache reader because a case proves it, and the
// module path of this very repository carries a capital letter, so the spelling
// is not an edge somebody invented for a test.
func EscapePath(modulePath string) string {
	var out strings.Builder
	for _, r := range modulePath {
		if r >= 'A' && r <= 'Z' {
			out.WriteByte('!')
			out.WriteRune(r + ('a' - 'A'))
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

// LicenceFilenames is what a module's licence is looked for under, in the order
// it is looked for. The list is short and conventional on purpose: a reader
// asking why a module was refused wants to know which names were tried, and a
// longer list would make the answer to that question longer without making the
// search meaningfully better.
var LicenceFilenames = []string{
	"LICENSE",
	"LICENSE.md",
	"LICENSE.txt",
	"LICENCE",
	"LICENCE.md",
	"LICENCE.txt",
	"COPYING",
	"COPYING.md",
}

// ModuleDir is where the cache holds one module, under the cache root.
func ModuleDir(module Module) string {
	return EscapePath(module.Path) + "@" + module.Version
}

// LicencePath joins a module directory and one of the filenames above, in the
// slash-separated form the cache uses, so that a message naming it reads the
// same on every platform the runner is built for.
func LicencePath(module Module, filename string) string {
	return path.Join(ModuleDir(module), filename)
}
