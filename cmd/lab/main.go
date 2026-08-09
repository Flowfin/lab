// Command lab reads a checkout of this repository and reports what it
// examined. It has two verbs that do work, one that judges and one that only
// reports, it takes no flags, and it writes nothing to the tree it reads.
package main

import (
	"fmt"
	"io"
	"os"
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
    lab list [path]    list the experiments at path, default ".", oldest
                       unanswered first
    lab help           print this text

lab writes nothing to the tree it reads.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, edges{
		walk: check.Walk,
		list: check.List,
		now:  time.Now(),
	}))
}

// edges is everything this command reaches for that is not its arguments: the
// walk, the listing, and the clock.
//
// The walk is a parameter for one reason: the mapping from a walk that refused
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
	walk func(string) (check.Result, error)
	list func(string, time.Time) (check.Listing, error)
	now  time.Time
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

	case "check":
		root, ok := pathArgument(args, errOut)
		if !ok {
			return exitCannot
		}

		result, err := e.walk(root)
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
