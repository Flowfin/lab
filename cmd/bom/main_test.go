// What this command's own suite is for, and it is not a second copy of
// internal/bom's cases.
//
// The render is proved there, against module sets a case wrote out in full.
// What nothing there can prove is that the module table inside a real binary
// reaches that render at all: a build read wrongly, a field taken from the
// wrong place or a replacement flattened would leave every case green and
// produce a document describing nothing. So this builds a binary from this
// repository and reads that.
//
// WHERE THE COLLECTOR IS PROVED, because it is not proved here and a reader
// should not have to find that out by looking. The one place a dependency can
// be dropped between a binary and either document is notices.BuildOf, which is
// the single reader of the module table both commands call, and
// cmd/notices/tree_test.go builds a tree with a dependency in it and holds the
// document to naming it. That proof covers this command because it covers that
// function, and it stopped being two proofs when the second reader was removed.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Flowfin/lab/internal/bom"
)

// TestTheModuleTableOfARealBinaryReachesTheDocument builds the runner and
// describes it.
//
// THE MAIN MODULE IS WHAT IT ASSERTS ON, not the dependency list. This module
// has no third-party dependencies, so a list read out of the binary and a list
// that was never read look identical in that half. The main module path does
// not: it is in the binary and nowhere else this command looks, so a document
// naming it is a document that read the table. The revision is the second such
// field, and it comes from a build setting rather than from the module record,
// so the two together cover both halves of what BuildOf reads.
//
// IT DOES NOT REQUIRE ONE EXIT CODE. Whether this build carries a release
// version is a property of the checkout the suite is running in: a run at a tag
// carries one and a run on a branch carries a version derived from the commit.
// Both are legitimate here, so the test requires the code and the standard
// error to agree with each other rather than requiring a number the checkout
// decides.
func TestTheModuleTableOfARealBinaryReachesTheDocument(t *testing.T) {
	binary := buildTheRunner(t)

	var out, errOut bytes.Buffer
	code := run([]string{binary}, &out, &errOut)

	document := decode(t, out.Bytes())
	if document.Metadata.Component.Name != "github.com/Flowfin/lab" {
		t.Errorf("the document names %q as the module the binary was built from", document.Metadata.Component.Name)
	}
	if len(document.Components) != 0 {
		t.Errorf("this module has no third-party dependency and the document lists %d component(s)", len(document.Components))
	}
	if !strings.Contains(out.String(), `"components": []`) {
		t.Errorf("the empty component list is not written as an empty list, so a reader cannot tell it from a field nobody wrote:\n%s", out.String())
	}

	revision := ""
	for _, property := range document.Metadata.Properties {
		if property.Name == "lab:vcs.revision" {
			revision = property.Value
		}
	}
	if revision == "" {
		t.Errorf("the document carries no revision property at all, so it says neither what the build was made at nor that it was made without one")
	}

	switch code {
	case exitClean:
		if errOut.Len() != 0 {
			t.Errorf("returned %d and said %q, so a clean run wrote a complaint", exitClean, errOut.String())
		}
	case exitRefused:
		// This module has no third-party dependency, so the only two things
		// a build of it can be refused for are facts about the checkout the
		// suite is running in: a working tree carrying changes, and a commit
		// that is not a tag. Both are ordinary here and neither is a defect
		// in the command, so what is required is that the run named one of
		// them rather than returning a number with nothing behind it.
		named := strings.Contains(errOut.String(), bom.MainComponentHasNoReleaseVersion) ||
			strings.Contains(errOut.String(), bom.BuildIsFromAModifiedTree)
		if !named {
			t.Errorf("returned %d and named no property this build can be refused for: %q", exitRefused, errOut.String())
		}
	default:
		t.Fatalf("returned %d, and stderr said %q", code, errOut.String())
	}
	t.Logf("described %s at version %q and revision %q, returning %d",
		filepath.Base(binary), document.Metadata.Component.Version, revision, code)
}

// TestTheDocumentIsTheOnlyThingOnStandardOutput holds the command to writing a
// document a program can read.
//
// A refusal is written to standard error and never into the document, and this
// is where that is held rather than only stated. A run that mixed a complaint
// into the output would produce a file that is no longer JSON, and the reader
// that discovers it is somebody's scanner rather than this suite.
func TestTheDocumentIsTheOnlyThingOnStandardOutput(t *testing.T) {
	binary := buildTheRunner(t)

	var out, errOut bytes.Buffer
	run([]string{binary}, &out, &errOut)

	decode(t, out.Bytes())
	if !strings.HasSuffix(out.String(), "}\n") {
		t.Errorf("the output does not end with the document, so something else was written after it")
	}
}

// TestABrokenInvocationIsNotARefusal holds the two codes apart.
//
// A gate that returns the same number for "this build names no release" and
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
		{"two arguments", []string{binary, t.TempDir()}},
		{"a file with no module table", []string{notABinary}},
		{"a binary that is not there", []string{filepath.Join(t.TempDir(), "absent")}},
		{"a directory rather than a binary", []string{t.TempDir()}},
	} {
		t.Run(c.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if code := run(c.args, &out, &errOut); code != exitCannot {
				t.Errorf("returned %d rather than %d", code, exitCannot)
			}
			if errOut.Len() == 0 {
				t.Errorf("returned %d and said nothing about why", exitCannot)
			}
			if out.Len() != 0 {
				t.Errorf("could not do its job and wrote %d byte(s) of document anyway", out.Len())
			}
		})
	}
}

// decode reads the document back, and fails rather than skipping when it
// cannot. A document this suite could not parse is not a document that passed.
func decode(t *testing.T, data []byte) bom.Document {
	t.Helper()

	var document bom.Document
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("the output is not a document anything can read: %v\n%s", err, data)
	}
	if document.BOMFormat != "CycloneDX" || document.SpecVersion != bom.SpecVersion {
		t.Errorf("the document declares %q %q rather than CycloneDX %s", document.BOMFormat, document.SpecVersion, bom.SpecVersion)
	}
	return document
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
