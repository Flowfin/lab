// Command bom renders the bill of materials for a binary this repository built,
// from the module table that binary carries.
//
// WHY THIS IS NOT A VERB OF THE RUNNER, and why it is a second command beside
// notices. lab reads a checkout. This reads a compiled artefact, which is not a
// checkout, and the release build is the only place it is ever run, so putting
// it inside lab would widen what an operator has to believe about a binary they
// downloaded in exchange for a verb that is useless outside a release. That is
// the judgement cmd/notices, cmd/contexts and cmd/pullrequest already made.
//
// It is separate from cmd/notices because the two produce different documents
// for different readers, which internal/bom argues at length, and because a
// single command writing two files would have to choose where each one lands.
// Where the files land is the release build's business.
//
// IT READS ONE PATH AND NOTHING ELSE. The binary to describe, given as the one
// argument. Unlike the notices it needs no module cache, because a bill of
// materials names components and reproduces no licence text, so nothing outside
// the binary is read at all. It asks no environment variable and runs no other
// program, so what it produced can be reproduced by hand from the path in the
// command line.
//
// IT OPENS NO CONNECTION. Everything it writes comes out of the binary it was
// given, so the document can be produced from an archive of a build with the
// network unplugged. That is the same claim the runner makes and it is made
// here for the same reason: a document that can only be produced online is a
// document that stops being producible.
package main

import (
	"debug/buildinfo"
	"fmt"
	"io"
	"os"

	"github.com/Flowfin/lab/internal/bom"
	"github.com/Flowfin/lab/internal/notices"
)

// The exit codes. Decision record 0011 is the contract, and this command
// returns the same three the runner does, for the same reasons, which is why
// they are written here with the record named rather than invented. The
// exit-code leg in internal/invariants reads every exit-code constant in the
// tree and refuses a code declared with two numbers, or a number declared under
// two codes, so keep the convention these three follow.
const (
	// exitClean means the run completed and refused nothing. It does not mean
	// the document is complete in the eyes of anybody consuming it, only that
	// every module the binary carries reached it with a version, and that the
	// build named a release of its own.
	exitClean = 0

	// exitRefused means the run completed and refused something. The document
	// is still written, because a document that is incomplete by named
	// entries is more useful than none, and the entries are named on standard
	// error rather than inside it: a field this repository invented would be
	// dropped in silence by anything reading the document as CycloneDX.
	exitRefused = 1

	// exitCannot means the command could not do its job: no argument, or a
	// binary it cannot read a module table out of. It is deliberately not the
	// code a refusal returns, because a gate that treats a broken invocation
	// like a violation reports one as the other and nobody investigates
	// either.
	exitCannot = 2
)

const usage = "usage: bom <binary>"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main with its edges passed in.
//
// The document goes to standard output rather than to a path this command
// chooses, which is the shape cmd/notices already has. Where the file lands is
// the release build's business, and a command that writes where it likes is one
// more thing to read before you can tell what a release contains.
func run(args []string, out, errOut io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintf(errOut, "bom: %s\n", usage)
		return exitCannot
	}
	binaryPath := args[0]

	info, err := buildinfo.ReadFile(binaryPath)
	if err != nil {
		fmt.Fprintf(errOut, "bom: cannot read the module table of %s: %v\n", binaryPath, err)
		return exitCannot
	}

	document := bom.Render(notices.BuildOf(info))
	text, err := document.JSON()
	if err != nil {
		// Rendering is a pure function over values the toolchain wrote, so
		// this is not a case anybody expects to meet. It is reported as the
		// command being unable to do its job rather than as a refusal,
		// because a document that was never written is not a document with a
		// named gap in it.
		fmt.Fprintf(errOut, "bom: %v\n", err)
		return exitCannot
	}
	fmt.Fprint(out, text)

	if len(document.Refusals) > 0 {
		for _, refusal := range document.Refusals {
			fmt.Fprintf(errOut, "bom: %s\n", refusal)
		}
		return exitRefused
	}
	return exitClean
}
