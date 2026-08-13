// What this command's own suite is for, and it is not a second copy of
// internal/notices' cases.
//
// The render is proved there, against module sets a case wrote out in full. What
// nothing there can prove is that the module table inside a real binary reaches
// that render at all: a build read wrongly, a field taken from the wrong place
// or a replacement flattened would leave every case green and produce a notices
// file describing nothing. So this builds a binary from this repository and
// reads that.
package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestTheModuleTableOfARealBinaryReachesTheDocument builds the runner and
// describes it.
//
// THE MAIN MODULE IS WHAT IT ASSERTS ON, not the dependency list. This module
// has no third-party dependencies, so a list read out of the binary and a list
// that was never read look identical in that half. The main module path does
// not: it is in the binary and nowhere else this command looks, so a document
// naming it is a document that read the table.
//
// It also asserts the disclosure sentence, because "no third-party module" and
// "the run produced nothing" are the two outcomes this whole package exists to
// keep apart.
func TestTheModuleTableOfARealBinaryReachesTheDocument(t *testing.T) {
	binary := buildTheRunner(t)
	cache := t.TempDir()

	var out, errOut bytes.Buffer
	if code := run([]string{binary, cache}, &out, &errOut); code != exitClean {
		t.Fatalf("returned %d, and stderr said %q", code, errOut.String())
	}

	document := out.String()
	if !strings.Contains(document, "github.com/Flowfin/lab") {
		t.Errorf("the document does not name the module the binary was built from:\n%s", document)
	}
	if !strings.Contains(document, "This binary contains no third-party module.") {
		t.Errorf("the document does not say that there is nothing to disclose:\n%s", document)
	}
	t.Logf("described %s in %d byte(s)", filepath.Base(binary), len(document))
}

// TestABrokenInvocationIsNotARefusal holds the two codes apart.
//
// A gate that returns the same number for "this module ships no licence" and
// "you pointed me at a directory" reports one as the other, and record 0011 is
// the contract that says it may not.
func TestABrokenInvocationIsNotARefusal(t *testing.T) {
	notABinary := filepath.Join(t.TempDir(), "not-a-binary")
	if err := os.WriteFile(notABinary, []byte("this is text\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := buildTheRunner(t)

	for _, c := range []struct {
		name string
		args []string
	}{
		{"no arguments at all", nil},
		{"one argument", []string{binary}},
		{"three arguments", []string{binary, t.TempDir(), "and one more"}},
		{"a file with no module table", []string{notABinary, t.TempDir()}},
		{"a binary that is not there", []string{filepath.Join(t.TempDir(), "absent"), t.TempDir()}},
		{"a cache root that is not there", []string{binary, filepath.Join(t.TempDir(), "absent")}},
		{"a cache root that is a file", []string{binary, notABinary}},
	} {
		t.Run(c.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if code := run(c.args, &out, &errOut); code != exitCannot {
				t.Errorf("returned %d rather than %d", code, exitCannot)
			}
			if errOut.Len() == 0 {
				t.Errorf("returned %d and said nothing about why", exitCannot)
			}
		})
	}
}

// buildTheRunner compiles this repository's runner into a temporary directory
// and returns the path.
//
// It builds rather than reading the test binary this suite is running inside.
// The test binary carries the testing packages and is not what a release ships,
// and the whole point of this file is to read the thing that is shipped.
func buildTheRunner(t *testing.T) string {
	t.Helper()

	name := "lab"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)

	build := exec.Command("go", "build", "-o", binary, "../lab")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("cannot build the runner: %v\n%s", err, output)
	}
	return binary
}
