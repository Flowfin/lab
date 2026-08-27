// Command lab reads a checkout of this repository and reports what it
// examined. It has two verbs that do work, one that judges and one that only
// reports, two that answer for the program itself, it takes no flags, and it
// writes nothing to the tree it reads.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Flowfin/lab/internal/check"
)

// The exit codes. Decision record 0011 is the contract and this is the only
// place the three the runner returns appear, so a caller keyed on one of them
// is reading the record whether or not anybody told it so. A fifth code
// supersedes that record rather than being added here.
//
// The fourth code the record fixes is not here. `3`, for a run that was asked
// for something and delivered nothing, belongs to the integration-hardware
// harness and is declared where its producer is, in internal/hardware. Record
// 0011 refuses a code written into the runner as a branch nothing can reach,
// and that is what one here would be.
//
// WHAT HOLDS THIS DECLARATION TO THE OTHERS. Keeping a code beside its producer
// means the contract is written down in more than one place, and each place was
// right about itself while nothing was right about all of them together. The
// exit-code leg in internal/invariants reads every declaration in the tree and
// refuses a code declared with two numbers, or a number declared under two
// codes, so a number that moves here is red in the default suite rather than
// green in this package. It finds a declaration by the shape of its name, so a
// code named something that does not read like one is invisible to it: keep the
// convention these three follow.
const (
	// exitClean means the run completed and refused nothing. It does not
	// mean the tree is good, only that nothing this runner judges was found
	// wrong.
	exitClean = 0

	// exitRefused means the run completed and refused something. It is the
	// only code that carries refusals, and the output names them.
	exitRefused = 1

	// exitCannot means the runner could not do its job: an argument it does
	// not understand, a path that is not a directory, a file it cannot read
	// at all. It is deliberately not the code a refusal returns, because a
	// gate that treats a broken invocation like a violation reports one as
	// the other and nobody investigates either.
	exitCannot = 2
)

// The three documents the last paragraph of the usage text names. They are
// named rather than restated, because a paraphrase printed by a binary somebody
// downloaded months ago is a copy of a document that has since moved on, and the
// reader has no way to tell which of the two they are holding.
//
// TWO LIMITS THAT BELONG HERE RATHER THAN ONLY IN THE ISSUE. A notice is not a
// control. Printing it stops nothing, and it should not be counted as a thing
// that prevents misuse when somebody later asks what does. And a notice an
// operator has to run a verb to see is weaker than one sitting in the download
// beside the binary, because the operator who most needs it is the one who runs
// the thing without asking it for help first. Both routes exist for that reason
// rather than either alone, and the download half is in the tree now: the
// release workflow copies these three files in beside the binaries, so an
// operator who never runs a verb still holds them.
//
// Nothing outside this package holds these strings to the tree. The paths leg
// of the invariants scan reads this repository's own documents, which is the
// files at the root and everything under docs/, and a path named inside the
// runner is outside that subject by the leg's own reckoning. So the guard is
// the test beside this declaration and there is no second one.
var documentsAnOperatorIsOwed = []string{
	"NOTICE.md",
	"LICENSE",
	"docs/privacy.md",
}

// documentsParagraph is the pointer the operator is owed, and it is one string
// rather than one per place that prints it. Both the usage text and the version
// output end with it, and a second copy would drift against the first the day
// one of the three files moves - which is the same failure the paragraph itself
// is written against, one level up.
const documentsParagraph = `NOTICE.md says what this program is for, LICENSE carries the terms it is under,
and docs/privacy.md says what stays on the host. Reading them is on you; this
text only says where they are.
`

const usage = `lab reads this repository and reports what it examined.

    lab check [path]   walk the tree at path, default ".", and report
    lab list [path]    list the experiments at path, default ".", oldest
                       unanswered first
    lab version        print the version this binary was built from
    lab help           print this text

lab writes nothing to the tree it reads.

` + documentsParagraph

// THIS IS WHERE THE RUNNER READS THE TIME, and it is read once. Everything
// downstream is given the value rather than asking again, so a run has one
// notion of now instead of one per caller, and two parts of a single run cannot
// disagree about which day it is.
//
// WHAT READING IT HERE DOES NOT BUY. Nothing makes the host clock right. A
// machine set two years fast refuses honest records and a machine set two years
// slow accepts a question dated next year as an ordinary one, and no reading of
// a checkout separates either case from a correct one. What it buys is that the
// value is one value, and that every run prints it, so a reader can see which
// now a verdict was made against instead of assuming it was the day they are
// reading on.
func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, edges{
		walk:      check.Walk,
		list:      check.List,
		now:       time.Now(),
		buildInfo: debug.ReadBuildInfo,
	}))
}

// edges is everything this command reaches for that is not its arguments: the
// walk, the listing, and the clock.
//
// The walk is a parameter for two reasons. The mapping from a walk that refused
// something to the exit code a refusal returns had to be proved before the
// first refusal existed, and no real walk could produce one then. The
// alternative was landing that mapping untested and finding out on the day a
// gate should have gone red.
//
// The clock is here because the listing prints how long an experiment has been
// asking, and a test that computed the expected number the same way the command
// does would assert nothing. Where the time is read is issue #63's, and this is
// only the seam a test can set.
type edges struct {
	walk func(string, time.Time) (check.Result, error)
	list func(string, time.Time) (check.Listing, error)
	now  time.Time

	// buildInfo is what the toolchain stamped into this binary. It is an
	// edge for the same reason the clock is: a test asserting what the
	// version verb prints cannot build a tagged binary to assert it
	// against, and one that read the real build info would assert whatever
	// the machine running the suite happened to produce, which is a
	// different string on a checkout, on a tag and on a modified tree.
	buildInfo func() (*debug.BuildInfo, bool)
}

// run is main with its edges passed in, so that what the command prints and
// what it returns can both be read by a test without starting a process.
func run(args []string, out, errOut io.Writer, e edges) int {
	if len(args) == 0 {
		fmt.Fprint(errOut, usage)
		return exitCannot
	}

	switch args[0] {
	case "help":
		if len(args) > 1 {
			fmt.Fprintf(errOut, "lab help takes no arguments\n")
			return exitCannot
		}
		fmt.Fprint(out, usage)
		return exitClean

	case "version":
		if len(args) > 1 {
			fmt.Fprintf(errOut, "lab version takes no arguments\n")
			return exitCannot
		}

		info, ok := e.buildInfo()
		if !ok {
			// The toolchain stamps this into every binary it builds in
			// module mode, so reaching here means the caller is holding
			// something this repository's build did not produce. Saying
			// which of the two it is beats printing an empty version and
			// letting the reader take it for one.
			fmt.Fprintf(errOut, "lab version: this binary carries no build information, so there is no version in it to report\n")
			return exitCannot
		}
		fmt.Fprint(out, versionText(info))
		return exitClean

	case "check":
		root, ok := pathArgument(args, errOut)
		if !ok {
			return exitCannot
		}

		result, err := e.walk(root, e.now)
		if err != nil {
			fmt.Fprintf(errOut, "lab check: %v\n", err)
			return exitCannot
		}
		fmt.Fprint(out, result.Report())
		if len(result.Refusals) > 0 {
			return exitRefused
		}
		return exitClean

	case "list":
		root, ok := pathArgument(args, errOut)
		if !ok {
			return exitCannot
		}

		listing, err := e.list(root, e.now)
		if err != nil {
			fmt.Fprintf(errOut, "lab list: %v\n", err)
			return exitCannot
		}
		fmt.Fprint(out, listing.Report(e.now))
		// Clean whatever it found. Nothing here fails because an experiment
		// is old, and that is deliberate: some questions take months, and a
		// check that timed them out would push somebody towards a hasty
		// answer to keep a build green. What the listing does is make the
		// choice between answering and abandoning a visible one.
		return exitClean

	default:
		fmt.Fprintf(errOut, "lab: unknown verb %q\n\n", args[0])
		fmt.Fprint(errOut, usage)
		return exitCannot
	}
}

// pathArgument reads the one optional path both walking verbs take. It is one
// function rather than two copies so that the two verbs cannot drift into
// taking different numbers of arguments, which is the kind of difference
// nobody notices until a script written against one is pointed at the other.
func pathArgument(args []string, errOut io.Writer) (string, bool) {
	switch len(args) {
	case 1:
		return ".", true
	case 2:
		return args[1], true
	default:
		fmt.Fprintf(errOut, "lab %s takes at most one path\n", args[0])
		return "", false
	}
}

// versionText renders what the version verb prints.
//
// WHERE THE VERSION COMES FROM AND WHY IT IS NOT A CONSTANT. The toolchain
// stamps the main module's version into the binary from the tags version
// control holds, so a build at a tag reports that tag and a build from a
// checkout with no tag on it reports a version the toolchain derived from the
// commit instead. A constant in a source file would be a second answer to the
// same question, and it would disagree with the tag silently, which makes every
// report from that build misleading rather than merely wrong.
//
// WHAT IT DOES NOT DECIDE. Nothing here says whether the string is a release
// version. That judgement already exists in internal/bom, where it refuses a
// published artefact that cannot be resolved back to a release, and a second
// copy of the pattern here would be the same rule answered in two places. What
// this prints instead is the string itself with the commit beside it, so the
// reader can see which of the two they are holding rather than being told.
func versionText(info *debug.BuildInfo) string {
	var b strings.Builder

	version := strings.TrimSpace(info.Main.Version)
	if version == "" {
		// An empty stamp is a different statement from a stamp the reader
		// cannot resolve, and printing "lab " with nothing after it would
		// read as the second.
		version = "(the toolchain stamped no version)"
	}
	fmt.Fprintf(&b, "lab %s\n", version)

	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}

	if revision := settings["vcs.revision"]; revision != "" {
		if built := settings["vcs.time"]; built != "" {
			fmt.Fprintf(&b, "built from commit %s, %s\n", revision, built)
		} else {
			fmt.Fprintf(&b, "built from commit %s\n", revision)
		}
	}

	// The one line that changes what the two above are worth. A build from a
	// tree carrying changes version control does not hold is described by
	// neither the tag nor the commit, and the stamp says so rather than
	// leaving a reader to infer it from a suffix on the version string.
	if settings["vcs.modified"] == "true" {
		fmt.Fprint(&b, "The tree this was built from carried changes version control did not hold,\nso neither the version nor the commit above describes every byte in this binary.\n")
	}

	// Three strings, one question, and the reader is told which of the three
	// they may be holding rather than being left to work it out from a shape.
	fmt.Fprint(&b, "\nThe version is what the toolchain stamped from version control when this binary\nwas built rather than a constant written into the source, so it is one of three\nthings. A tag, which is what a release carries. A version of the shape\nv0.0.0-<timestamp>-<commit>, which the toolchain derives from a commit no tag\nnames, and which is what building from an ordinary checkout produces. Or\n(devel), the placeholder it writes when it was told nothing at all, which is\nwhat running the source without building it produces.\n")

	fmt.Fprint(&b, "\n"+documentsParagraph)
	return b.String()
}
