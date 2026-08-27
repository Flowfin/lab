// Command notices renders the third-party notices for a binary this repository
// built, from the module table that binary carries.
//
// WHY THIS IS NOT A VERB OF THE RUNNER. lab reads a checkout. This reads a
// compiled artefact and a module cache, which are neither of them a checkout,
// and the release build is the only place it is ever run. Putting it inside lab
// would widen what an operator has to believe about a binary they downloaded in
// exchange for a verb that is useless outside a release. It is the same
// judgement cmd/contexts and cmd/pullrequest already made, and record 0002 names
// none of the three, which is the thing in this file a reader should check
// rather than accept.
//
// IT READS TWO PATHS AND NOTHING ELSE. The binary to describe and the module
// cache to read licence texts out of, both given as arguments. It asks no
// environment variable and runs no other program, so what it produced can be
// reproduced by hand from the two paths in the command line. Asking the
// toolchain where the cache is would be one more thing to have installed at the
// moment a release is being built, and the workflow already knows the answer.
//
// IT OPENS NO CONNECTION. Every licence text comes out of the cache the build
// already populated, so a notices file can be produced from an archive of a
// build with the network unplugged. That is the same claim the runner makes and
// it is made here for the same reason: a document that can only be produced
// online is a document that stops being producible.
package main

import (
	"debug/buildinfo"
	"fmt"
	"io"
	"os"

	"github.com/Flowfin/lab/internal/notices"
)

// The exit codes. Decision record 0011 is the contract, and this command returns
// the same three the runner does, for the same reasons, which is why they are
// written here with the record named rather than invented. The exit-code leg in
// internal/invariants reads every exit-code constant in the tree and refuses a
// code declared with two numbers, or a number declared under two codes, so keep
// the convention these three follow.
const (
	// exitClean means the run completed and refused nothing. It does not mean
	// the notices are complete in the eyes of any licence, only that every
	// module the binary carries was listed with the text it shipped.
	exitClean = 0

	// exitRefused means the run completed and refused something. The document
	// is still written, because a document that is incomplete by named
	// entries is more useful than none, and it says which entries.
	exitRefused = 1

	// exitCannot means the command could not do its job: no arguments, a
	// binary it cannot read a module table out of, a cache path that is not
	// there. It is deliberately not the code a refusal returns, because a
	// gate that treats a broken invocation like a violation reports one as the
	// other and nobody investigates either.
	exitCannot = 2
)

const usage = "usage: notices <binary> <module-cache-root>"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main with its edges passed in.
//
// The document goes to standard output rather than to a path this command
// chooses. Where the file lands is the release build's business, and a command
// that writes where it likes is one more thing to read before you can tell what
// a release contains.
func run(args []string, out, errOut io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintf(errOut, "notices: %s\n", usage)
		return exitCannot
	}
	binaryPath, cacheRoot := args[0], args[1]

	info, err := buildinfo.ReadFile(binaryPath)
	if err != nil {
		fmt.Fprintf(errOut, "notices: cannot read the module table of %s: %v\n", binaryPath, err)
		return exitCannot
	}
	if stat, err := os.Stat(cacheRoot); err != nil || !stat.IsDir() {
		// A cache that is not there is a broken invocation and not a module
		// with no licence. Without this the run would refuse every dependency
		// under a property that names the module, and the reader would go
		// looking at the modules.
		fmt.Fprintf(errOut, "notices: %s is not a module cache this run can read\n", cacheRoot)
		return exitCannot
	}

	document := notices.Render(notices.BuildOf(info), notices.Cache{Root: cacheRoot})
	fmt.Fprint(out, document.Text())

	if len(document.Refusals) > 0 {
		for _, refusal := range document.Refusals {
			fmt.Fprintf(errOut, "notices: %s\n", refusal)
		}
		return exitRefused
	}
	return exitClean
}
