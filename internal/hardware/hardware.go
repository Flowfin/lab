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
	"testing"
)

// Name is what this harness is called wherever it is referred to.
const Name = "integration-hardware"

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
// What happens when the harness was asked for and the hardware is not there
// is not decided here. That is issue #31, which is about what a skip says,
// and this function is where it will go.
func Needs(t *testing.T, hardware string) {
	t.Helper()

	if hardware == "" {
		t.Fatal("this test names no hardware, so a reader cannot tell what it would take to run it")
	}
	if !Asked() {
		t.Skipf("%s was not asked for. this test needs %s", Name, hardware)
	}
	t.Logf("this test needs %s, and it ran on whatever machine that is", hardware)
}
