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

// The exit codes. Decision record 0011 is the contract and this is the only
// place the numbers appear, so a caller keyed on one of them is reading the
// record whether or not anybody told it so. A fifth code supersedes that
// record rather than being added here.
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

const usage = `lab reads this repository and reports what it examined.

    lab check [path]   walk the tree at path, default ".", and report
    lab help           print this text

lab writes nothing to the tree it reads.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, check.Walk))
}

// run is main with its edges passed in, so that what the command prints and
// what it returns can both be read by a test without starting a process.
//
// The walk is a parameter for one reason: the mapping from a walk that refused
// something to the exit code a refusal returns has to be proved before the
// first refusal exists, and no real walk can produce one yet. The alternative
// is landing that mapping untested and finding out on the day a gate should
// have gone red.
func run(args []string, out, errOut io.Writer, walk func(string) (check.Result, error)) int {
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

		result, err := walk(root)
		if err != nil {
			fmt.Fprintf(errOut, "lab check: %v\n", err)
			return exitCannot
		}
		fmt.Fprint(out, result.Report())
		if len(result.Refusals) > 0 {
			return exitRefused
		}
		return exitClean

	default:
		fmt.Fprintf(errOut, "lab: unknown verb %q\n\n", args[0])
		fmt.Fprint(errOut, usage)
		return exitCannot
	}
}
