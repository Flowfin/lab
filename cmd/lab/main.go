// Command lab reads a checkout of this repository and reports what it
// examined. It has one verb that does work, it takes no flags, and it writes
// nothing to the tree it reads.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Flowfin/lab/internal/check"
)

// The exit codes this command can return today. The full contract, including
// the codes the later verbs need, is decided in issue #52 and recorded there;
// these two are chosen to fit that contract rather than to be moved by it.
const (
	// exitClean means the run completed and refused nothing. It does not
	// mean the tree is good, only that nothing this runner judges was found
	// wrong.
	exitClean = 0

	// exitCannot means the runner could not do its job: an argument it does
	// not understand, a path that is not a directory, a file it cannot read
	// at all. It is deliberately not the code a refusal returns, because a
	// gate that treats a broken invocation like a violation reports one as
	// the other and nobody investigates either.
	exitCannot = 2
)

const usage = `lab reads this repository and reports what it examined.

    lab check [path]   walk the tree at path, default ".", and report
    lab help           print this text

lab writes nothing to the tree it reads.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main with its edges passed in, so that what the command prints and
// what it returns can both be read by a test without starting a process.
func run(args []string, out, errOut io.Writer) int {
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

	case "check":
		root := "."
		switch len(args) {
		case 1:
		case 2:
			root = args[1]
		default:
			fmt.Fprintf(errOut, "lab check takes at most one path\n")
			return exitCannot
		}

		result, err := check.Walk(root)
		if err != nil {
			fmt.Fprintf(errOut, "lab check: %v\n", err)
			return exitCannot
		}
		fmt.Fprint(out, result.Report())
		// Nothing here refuses anything yet, so a completed walk is always
		// clean. The code a refusal returns is not written in advance: an
		// untested branch that nothing can reach is a branch nobody has
		// proved, and it would ship as the one deciding whether a gate goes
		// red. It arrives with the first refusal and with the test that
		// reaches it, under the contract issue #52 records.
		return exitClean

	default:
		fmt.Fprintf(errOut, "lab: unknown verb %q\n\n", args[0])
		fmt.Fprint(errOut, usage)
		return exitCannot
	}
}
