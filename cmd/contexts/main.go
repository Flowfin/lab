// Command contexts compares the contexts the ruleset requires on the default
// branch against the check names the workflows in this repository declare, and
// refuses when the two disagree.
//
// WHY THIS IS NOT A VERB OF THE RUNNER. lab reads a checkout and opens no
// network connection, which is a claim a downloader is asked to take on trust
// and which cmd/lab's own suite holds it to. Half of this comparison is a live
// setting on the hosting platform, so it can only be answered by asking the API,
// and putting that inside lab would widen what an operator has to believe about
// a binary they downloaded in exchange for a verb that is useless outside a
// workflow. Keeping it here costs a third entry point under cmd/, which record
// 0002 does not name, and that is the judgement in this file a reader should
// check rather than accept. It is the same one cmd/pullrequest already made for
// the same reason.
//
// THE API CALL IS NOT IN HERE EITHER. The required set arrives on standard
// input, one context per line, so this command reads a checkout and a list and
// nothing else. The workflow that runs it is where the token is, which keeps the
// credential in one visible place and lets every rule below be proved against a
// list a test wrote out in full.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Flowfin/lab/internal/contexts"
)

// The exit codes. Decision record 0011 is the contract, and this command returns
// the same three the runner does, for the same reasons, which is why they are
// written here with the record named rather than invented. The exit-code leg in
// internal/invariants reads every exit-code constant in the tree and refuses a
// code declared with two numbers, or a number declared under two codes, so keep
// the convention these three follow.
const (
	// exitClean means the run completed and refused nothing. It does not mean
	// the gate is intact, only that the two lists agree, and the report says
	// what it compared so the two are not confused.
	exitClean = 0

	// exitRefused means the run completed and refused something. It is the
	// only code that carries refusals.
	exitRefused = 1

	// exitCannot means the check could not do its job: no required set on
	// standard input, no workflow directory to read, a file in a shape the
	// reader was not built for. It is deliberately not the code a refusal
	// returns, because a gate that treats a broken invocation like a violation
	// reports one as the other and nobody investigates either.
	exitCannot = 2
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, contexts.WorkflowsDir, contexts.Absences))
}

// run is main with its edges passed in: where the workflows are read from, and
// which deliberate-absence list the comparison is made against. Both are
// parameters so that what the command prints and what it returns can be read by
// a test against a tree written out in full, rather than against whichever
// repository the suite happens to be running inside.
func run(in io.Reader, out, errOut io.Writer, workflowsDir string, absences []contexts.Absence) int {
	required, err := readRequired(in)
	if err != nil {
		fmt.Fprintf(errOut, "contexts: %v\n", err)
		return exitCannot
	}

	declared, err := contexts.ReadWorkflows(workflowsDir)
	if err != nil {
		fmt.Fprintf(errOut, "contexts: %v\n", err)
		return exitCannot
	}

	verdict := contexts.Judge(declared, required, absences)
	fmt.Fprint(out, verdict.Report(declared, required, absences))
	if len(verdict.Refusals) > 0 {
		return exitRefused
	}
	return exitClean
}

// readRequired reads the contexts the ruleset requires, one per line.
//
// AN EMPTY LIST IS A LIST AND NOT A FAILURE. The ruleset on this board requires
// no status check today, so a run that read nothing is the ordinary case rather
// than a broken invocation, and the comparison still has work to do: every check
// name the tree declares has to be written down as a deliberate absence, and an
// absence naming nothing is still refused. A command that treated no input as a
// reason to stop would switch the whole check off on exactly the board it was
// written for.
//
// What it will not accept is a line that is not a context name. A required
// context is a check-run name, and a line carrying a tab or a control character
// is a fetch that returned something other than the list this expects, which is
// a broken invocation rather than a gate that disagrees.
func readRequired(in io.Reader) ([]string, error) {
	var required []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.ContainsAny(line, "\t") || strings.ContainsFunc(line, isControl) {
			return nil, fmt.Errorf("the required set carries the line %q, which is not a check-run name, so what arrived on standard input is not the list this expects", line)
		}
		required = append(required, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot read the required set: %w", err)
	}
	return required, nil
}

func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f
}
