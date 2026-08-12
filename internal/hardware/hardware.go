// Package hardware is the integration-hardware harness. Its name says out
// loud what it is: it needs hardware, and it is not part of the default run.
// A harness called something neutral gets mistaken for the main suite by
// whoever reads the output next, and a green board then means less than it
// appears to.
//
// Record 0007 keeps the default run headless and unelevated. This harness is
// where a test that genuinely cannot meet those halves goes instead, so the
// default run stays clean and the things it cannot cover are somewhere else
// with their own name rather than skipped inside it.
//
// A measurement this harness produces is a measurement about the machine it
// ran on. It is never reported as the default suite's result, on the board or
// in any document, because that is what it would have to be to mean anything
// there.
//
// Nothing runs from here unless it is asked for. The tests are behind a build
// constraint, so the default run does not compile them, and Needs refuses to
// let one run even when the constraint is satisfied and the harness was not
// asked for.
package hardware

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
)

// Name is what this harness is called wherever it is referred to.
const Name = "integration-hardware"

// ExitAskedAndDeliveredNothing is what a run of this harness reports when it
// was asked for and executed nothing. Decision record 0011 fixes the meaning:
// being asked for coverage and producing none is a failure of the request even
// though no assertion broke, and reporting it as zero would let a run that
// exercised nothing pass for a run that found nothing.
//
// Record 0011 is the contract for this number and for the three the runner
// returns. It lives here rather than beside those three because this is where
// its producer is, and that record refuses a code written into the runner as a
// branch nothing can reach.
//
// WHAT HOLDS THIS DECLARATION TO THE OTHERS. The exit-code leg in
// internal/invariants reads every exit-code constant in the tree and refuses a
// number declared under two codes, which is the direction that reaches this
// one: it is declared exactly once, so there is no second copy of it for
// anything to compare against, and what can go wrong is this number colliding
// with a meaning one of the runner's three already has. It finds a declaration
// by the shape of its name, so keep the convention this one and those three
// follow.
const ExitAskedAndDeliveredNothing = 3

// BuildTag is the constraint its tests are behind. The default run does not
// compile them, which is a stronger exclusion than skipping at run time: a
// test that is not compiled cannot be enabled by accident, and it cannot fail
// the default run by failing to build against hardware that is not there.
const BuildTag = "integration_hardware"

// EnvVar is how the harness is asked for. It is deliberately a second gate
// behind the build constraint, so that a person who builds with the tag out
// of curiosity still gets nothing until they say they meant it.
const EnvVar = "LAB_INTEGRATION_HARDWARE"

// Asked reports whether this harness was asked for.
func Asked() bool {
	return os.Getenv(EnvVar) == "1"
}

// Disclosure is the line a run prints when this harness was not asked for. It
// says that it was not asked for and what asking would cost, so a run that
// covered less than the whole set cannot be read as one that covered it and
// found nothing.
func Disclosure() string {
	return fmt.Sprintf(
		"the %s harness was not asked for and nothing in it ran.\n"+
			"asking costs a machine with the hardware each test names and an explicit request:\n"+
			"    go test -tags %s ./internal/hardware\n"+
			"with %s=1 in the environment. its results are about that machine and are not this suite's results.",
		Name, BuildTag, EnvVar)
}

// Needs names the hardware a test requires, in words a person can act on, and
// stops the test where the harness was not asked for.
//
// Every test in this harness calls it, and the default run holds them to that
// rather than trusting it: a failure saying a device was unavailable is
// useful, and one saying an assertion failed is not.
//
// It is also where a test is counted. The tally is taken at the end of the
// test rather than here, because a test that reaches this line and skips two
// lines later has not run, and a skip is the one result that is reported in
// the same colour as a pass.
func Needs(t *testing.T, hardware string) {
	t.Helper()

	if hardware == "" {
		t.Fatal("this test names no hardware, so a reader cannot tell what it would take to run it")
	}

	name := t.Name()
	t.Cleanup(func() {
		if t.Skipped() {
			runs.skipped(name, hardware)
			return
		}
		runs.executed()
	})

	if !Asked() {
		t.Skipf("%s was not asked for. this test needs %s", Name, hardware)
	}
	t.Logf("this test needs %s, and it ran on whatever machine that is", hardware)
}

// Absent stops a test because the hardware it named is not on this machine. It
// takes the hardware again and how that was established, because a skip saying
// only that something was missing sends the reader to the test to find out
// what, and the reader is usually somebody deciding whether their machine
// could run it.
//
// Every skip in this harness goes through Needs or through here, and both name
// what was missing. The tally counts either as a test that did not run.
func Absent(t *testing.T, hardware, how string) {
	t.Helper()

	if hardware == "" || how == "" {
		t.Fatal("a skip has to name the hardware that was missing and how that was established, or it says nothing a reader can act on")
	}
	t.Skipf("this test needs %s, which is not on this machine: %s", hardware, how)
}

// runs is what the harness counted. A run where everything skipped and a run
// where everything passed look the same from outside, and the count is the
// only thing that separates them.
var runs tally

// tally holds what a harness run executed and what it did not.
type tally struct {
	mu      sync.Mutex
	ran     int
	missing map[string]string
}

func (t *tally) executed() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ran++
}

func (t *tally) skipped(test, hardware string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.missing == nil {
		t.missing = make(map[string]string)
	}
	t.missing[test] = hardware
}

func (t *tally) counts() (ran, skipped int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ran, len(t.missing)
}

// Summary is the line a harness run prints when it was asked for. It says how
// many of its tests ran and how many did not, and it names what each one that
// did not was missing.
//
// A LINE ADMITTING SOMETHING WAS NOT VERIFIED SURVIVES EVERY LATER EDIT AND
// GETS SHARPER IF IT CHANGES AT ALL. Rewriting a skip notice into a statement
// that the case passed is worse than deleting the test, and this is where a
// reader looks for it, so this is where that is written down.
func Summary() string {
	runs.mu.Lock()
	defer runs.mu.Unlock()

	var out strings.Builder
	fmt.Fprintf(&out, "%s: %d of its tests ran and %d did not.\n", Name, runs.ran, len(runs.missing))

	if len(runs.missing) == 0 {
		if runs.ran == 0 {
			out.WriteString("nothing in this harness was executed, so this run is evidence about nothing.\n")
		}
		return out.String()
	}

	names := make([]string, 0, len(runs.missing))
	for name := range runs.missing {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&out, "  did not run: %s, missing %s\n", name, runs.missing[name])
	}
	out.WriteString("what those tests would have covered was not verified on this machine.\n")
	return out.String()
}

// Verdict is the exit code a finished harness run reports, given the code the
// test binary produced, whether the harness was asked for, and how many of its
// tests executed.
//
// It is a function rather than four lines inside TestMain because record 0011
// requires every code the harness can return to be reached by a test in the
// default run, and no default run can produce a harness run that skipped
// everything. The alternative is landing the mapping untested and finding out
// on the day a job should have gone red.
//
// WHAT A CALLER ACTUALLY SEES. This is the exit status of the test binary. go
// test collapses every non-zero status from a package to 1 in its own exit
// code, so a caller that needs the three apart runs the binary rather than the
// go command:
//
//	go test -tags integration_hardware -c -o harness ./internal/hardware
//	LAB_INTEGRATION_HARDWARE=1 ./harness
//
// Through go test, a harness that was asked for and skipped everything is red
// rather than distinctly red, which is the honest half of the claim and is
// written here rather than left to be discovered.
func Verdict(code int, asked bool, ran int) int {
	if code != 0 || !asked || ran > 0 {
		return code
	}
	return ExitAskedAndDeliveredNothing
}

// Report is what a harness run prints and returns when it finishes: the
// disclosure where it was not asked for, the summary where it was, and the
// exit code either way.
func Report(code int) (string, int) {
	if !Asked() {
		return Disclosure() + "\n", code
	}
	ran, _ := runs.counts()
	return Summary(), Verdict(code, true, ran)
}
