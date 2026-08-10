package main

import (
	"bytes"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Flowfin/lab/internal/check"
)

// The runner opens no network connection. It reads a checkout and writes to its
// own output, and that is the whole of its contact with the world. There is no
// telemetry, no update check, no schema fetched from anywhere and no version
// ping.
//
// The two tests below prove it from different directions, and neither is the
// whole proof on its own.
//
// The first reads the runner's transitive dependency set and refuses any
// package that can open a socket. It reaches a network client arriving inside a
// dependency nobody read, which is the shape a hand audit of this package's own
// imports would miss.
//
// The second runs the runner over the largest trees this repository holds, by
// every verb it has, and asserts each one completed with a verdict. It reaches
// a route that is never exercised, because a dependency proof says what the
// binary could do and says nothing about which of it ran.
//
// WHAT THE PAIR PROVES AND WHAT IT DOES NOT. Together they say that a full run
// over the largest input available executed in a binary that links nothing able
// to open a connection, so no outbound connection was attempted because none
// was reachable. That is a structural argument, not an observation: nothing
// here watches syscalls, and there is no portable way to do so across the three
// platforms this suite runs on that would not itself be the thing most likely
// to break. A run under a syscall tracer would say more, and it would say it on
// one platform.
//
// The first test is the one that catches the change this claim will actually
// face, which is a future contributor adding a helpful update check. That
// arrives with good intentions and a reasonable justification, and it cannot
// arrive without a network package appearing below.

// networkImport says whether an import path is a package that can open an
// outbound connection.
//
// It is the whole of net rather than a list of the parts of it that dial. A
// list would have to be right about which parts those are, forever, and the
// runner needs none of net today, so the strict rule costs nothing and the
// permissive one would cost a judgement every time the standard library grows.
// A change that genuinely needs one of these fails here and is argued in the
// issue that wants it, which is the right place for it.
func networkImport(path string) bool {
	return path == "net" || strings.HasPrefix(path, "net/")
}

// TestTheRunnerLinksNoNetworkPackage is the dependency proof.
//
// It asks the toolchain rather than reading this package's own import block,
// because the import block says what this file reaches for and the question is
// what the binary carries. The list is transitive, so a network client three
// dependencies down is in it.
func TestTheRunnerLinksNoNetworkPackage(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Fatalf("cannot read the runner's dependencies: %v", err)
	}

	deps := strings.Fields(string(out))
	if len(deps) == 0 {
		t.Fatal("the dependency list is empty, which cannot be right, and an empty list would pass every check below")
	}

	// Printed before the verdict and phrased as a count rather than as a
	// result. A line saying none of them is a network package, written next to
	// an error saying nine of them are, is the report contradicting itself in
	// the one run somebody is reading closely.
	t.Logf("%d packages read", len(deps))

	var found []string
	for _, dep := range deps {
		if networkImport(dep) {
			found = append(found, dep)
		}
	}
	if len(found) > 0 {
		t.Errorf("the runner links %s, so it can open an outbound connection",
			strings.Join(found, ", "))
	}
}

// TestTheNetworkRuleWouldRefuseANetworkPackage is the near miss. Without it,
// a rule that refused nothing at all would pass the test above on a tree that
// happens to be clean, and the two are indistinguishable from a green run.
func TestTheNetworkRuleWouldRefuseANetworkPackage(t *testing.T) {
	for _, path := range []string{"net", "net/http", "net/http/httptest", "net/smtp"} {
		if !networkImport(path) {
			t.Errorf("%s is not read as a network package, and the rule above would let it into the runner", path)
		}
	}
	for _, path := range []string{"os", "fmt", "internal/check", "strings"} {
		if networkImport(path) {
			t.Errorf("%s is read as a network package, and the rule above refuses work it has no business refusing", path)
		}
	}
}

// TestAFullRunOverTheLargestTreeCompletes is the behavioural half.
//
// Every verb, over the two largest trees this repository holds, asserting a
// verdict rather than a broken invocation. The whole fixture corpus is the
// largest tree the fixtures provide, larger than any single case in it, and the
// checkout itself is the tree the gate actually runs over. A run that returned
// exitCannot would prove nothing about a run, and it is the outcome a change to
// the walk is most likely to produce by accident.
//
// A verdict here is a verdict of either colour. Pointed at the fixture corpus
// the runner refuses a great deal, because that corpus is built out of trees
// that exist to be refused, and that is a run rather than a failure.
func TestAFullRunOverTheLargestTreeCompletes(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)

	for _, root := range []string{
		filepath.Join("..", "..", check.FixturesDir),
		filepath.Join("..", ".."),
	} {
		t.Logf("%s holds %d files", root, filesUnder(t, root))

		for _, verb := range []string{"check", "list"} {
			var out, errOut bytes.Buffer
			code := run([]string{verb, root}, &out, &errOut, edges{
				walk: check.Walk,
				list: check.List,
				now:  now,
			})
			if code == exitCannot {
				t.Fatalf("lab %s over %s could not do its job: %s", verb, root, errOut.String())
			}
			if out.Len() == 0 {
				t.Errorf("lab %s over %s printed nothing, so nothing was examined", verb, root)
			}
		}
	}
}

// filesUnder counts the files below a directory, so the test says how large the
// tree it ran over actually was rather than asserting that it was large.
func filesUnder(t *testing.T, root string) int {
	t.Helper()

	files := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}
		if !d.IsDir() {
			files++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("cannot walk %s: %v", root, err)
	}
	if files == 0 {
		t.Fatalf("%s holds no files, so a run over it examines nothing", root)
	}
	return files
}
