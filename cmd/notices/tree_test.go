// What this file proves that neither of the other two suites can.
//
// internal/notices proves the render, against module sets a case wrote out in
// full. main_test.go proves that the module table inside a real binary reaches
// that render, against a binary built from this repository. Neither of them
// starts from a tree: this repository has no third-party dependency, so a build
// of it exercises the empty half only, and a module set constructed in a test is
// a statement about the test rather than about a build.
//
// Issue #37 asks for the other half in its own words - a run against a tree with
// a dependency added produces a notices file that lists it - so this builds two
// binaries from one tree, the second differing from the first by a require line
// and an import, and holds the difference between the two documents.
//
// IT OPENS NO CONNECTION, which is what made this leg look unreachable. The
// module is served out of a directory laid out as a module proxy and written by
// this test, the checksum database is off, and the toolchain is pinned to the
// one already installed, so nothing here resolves a name or fetches a version.
// The module cache the build populates is a temporary directory, so the licence
// text the document carries came from this run rather than from whatever the
// machine happened to have downloaded before it.
package main

import (
	"archive/zip"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The module the tree under test depends on.
//
// It is invented rather than borrowed from the real ecosystem. A test that
// depends on something real depends on that thing staying fetchable and staying
// licensed the way it was on the day the test was written, and both of those are
// facts about somebody else's repository.
const (
	dependencyPath    = "example.com/dep"
	dependencyVersion = "v1.0.0"
)

// The go directive is well below the toolchain this repository builds with on
// purpose. The tree here is built by whatever toolchain is running the suite,
// and a directive naming a version newer than that one asks the toolchain to
// fetch a different one, which is the network this file exists without.
const dependencyGoMod = "module example.com/dep\n\ngo 1.21\n"

const dependencySource = `package dep

// Answer is here so that the tree under test imports this module and uses it.
// A module that is required and never imported does not reach the binary, and
// then the document would be correct to leave it out.
func Answer() int { return 37 }
`

// The licence text is written to be unmistakable in a document. What the
// assertion below has to separate is a document naming a module from a document
// carrying the text that module shipped, and a conventional licence header
// appears in enough places to make that a weak test.
const dependencyLicenceText = `Copyright 2026 the author of example.com/dep

Permission is granted to reproduce this text in a notices document, which is
the whole of what this file exists to be reproduced by.
`

// TestATreeWithADependencyAddedProducesADocumentThatListsIt builds one tree
// twice and compares the two documents.
//
// The comparison is the test. A tree that had the dependency from the start
// would pass against a command that lists whatever it finds and also against one
// that lists a module somebody hard-coded, and the difference between the two
// runs is what separates them.
func TestATreeWithADependencyAddedProducesADocumentThatListsIt(t *testing.T) {
	workspace := t.TempDir()
	proxy := writeTheModuleProxy(t, filepath.Join(workspace, "proxy"))
	cache := filepath.Join(workspace, "modcache")
	tree := filepath.Join(workspace, "tree")

	// The cache is created here rather than left to the build. The first of the
	// two builds needs nothing, so it populates no cache, and a command told to
	// read a directory that is not there refuses the invocation instead of
	// describing the binary.
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}

	environment := offlineEnvironment(proxy, cache)
	cleanTheModuleCache(t, environment)

	writeTheTree(t, tree, false)
	before, code := describe(t, buildTheTree(t, tree, environment, "before"), cache)
	if code != exitClean {
		t.Fatalf("describing the tree without the dependency returned %d", code)
	}
	if strings.Contains(before, dependencyPath) {
		t.Fatalf("the tree without the dependency already names %s, so the comparison below proves nothing:\n%s", dependencyPath, before)
	}
	if !strings.Contains(before, "This binary contains no third-party module.") {
		t.Errorf("the tree without the dependency does not say that there is nothing to disclose:\n%s", before)
	}

	writeTheTree(t, tree, true)
	after, code := describe(t, buildTheTree(t, tree, environment, "after"), cache)
	if code != exitClean {
		t.Fatalf("describing the tree with the dependency returned %d rather than %d, so the licence text was not read out of the cache the build populated", code, exitClean)
	}

	named := dependencyPath + "@" + dependencyVersion
	if !strings.Contains(after, named) {
		t.Errorf("the document does not name %s:\n%s", named, after)
	}
	if !strings.Contains(after, strings.TrimSpace(dependencyLicenceText)) {
		t.Errorf("the document names %s and does not carry the text it shipped, which is the half a link would also have failed to supply:\n%s", named, after)
	}
	t.Logf("adding %s moved the document from %d to %d byte(s) and carried its licence text", named, len(before), len(after))
}

// describe runs the command over one binary and returns what it wrote and the
// code it returned.
func describe(t *testing.T, binary, cache string) (string, int) {
	t.Helper()

	var out, errOut bytes.Buffer
	code := run([]string{binary, cache}, &out, &errOut)
	if errOut.Len() > 0 {
		t.Logf("the run said: %s", strings.TrimSpace(errOut.String()))
	}
	return out.String(), code
}

// writeTheTree writes a module that either imports the dependency or does not.
//
// Both halves are written rather than the second being an edit of the first, so
// that what the two builds differ by is stated in one place a reader can see.
func writeTheTree(t *testing.T, dir string, withDependency bool) {
	t.Helper()

	goMod := "module example.com/tree\n\ngo 1.21\n"
	source := "package main\n\nfunc main() { println(\"the tree ran\") }\n"
	if withDependency {
		goMod = "module example.com/tree\n\ngo 1.21\n\nrequire " + dependencyPath + " " + dependencyVersion + "\n"
		source = "package main\n\nimport \"" + dependencyPath + "\"\n\nfunc main() { println(dep.Answer()) }\n"
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "go.mod"), goMod)
	write(t, filepath.Join(dir, "main.go"), source)
}

// buildTheTree compiles the tree and returns the path of the binary.
func buildTheTree(t *testing.T, dir string, environment []string, name string) string {
	t.Helper()

	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)

	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = dir
	build.Env = environment
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("cannot build the tree under test: %v\n%s", err, output)
	}
	return binary
}

// writeTheModuleProxy lays out the dependency the way a module proxy serves it
// and returns the directory to point the toolchain at.
func writeTheModuleProxy(t *testing.T, root string) string {
	t.Helper()

	dir := filepath.Join(root, filepath.FromSlash(dependencyPath), "@v")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "list"), dependencyVersion+"\n")
	write(t, filepath.Join(dir, dependencyVersion+".info"), `{"Version":"`+dependencyVersion+`"}`)
	write(t, filepath.Join(dir, dependencyVersion+".mod"), dependencyGoMod)

	if err := os.WriteFile(filepath.Join(dir, dependencyVersion+".zip"), moduleArchive(t), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// moduleArchive builds the zip a module proxy serves, which is every file of the
// module under one directory named for the module and its version.
func moduleArchive(t *testing.T) []byte {
	t.Helper()

	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	prefix := dependencyPath + "@" + dependencyVersion + "/"

	for _, file := range []struct{ name, content string }{
		{"go.mod", dependencyGoMod},
		{"dep.go", dependencySource},
		{"LICENSE", dependencyLicenceText},
	} {
		entry, err := archive.Create(prefix + file.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(file.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

// offlineEnvironment is what the two builds run under.
//
// Every entry is here to remove one route to the network or to the machine's own
// state. The proxy is a directory. The checksum database is off, because it is a
// service and the module it would be asked about does not exist outside this
// test. The toolchain is the installed one, so a go directive is never a
// download. The workspace file is off, so a go.work above the temporary
// directory cannot pull this build into it. The module cache is a temporary
// directory, so the licence text in the document came from this run.
//
// THE THREE EMPTIED VARIABLES ARE THE ONES THAT SEND A BUILD PAST THE PROXY.
// They are inherited from whoever is running the suite, and any of them matching
// this module makes the toolchain resolve the path as a domain and fetch over
// https, which is a failure that only appears on a machine configured a
// particular way. Emptying them is how this file stops depending on that
// configuration.
func offlineEnvironment(proxy, cache string) []string {
	return append(os.Environ(),
		"GOPROXY="+proxyURL(proxy),
		"GOSUMDB=off",
		"GOPRIVATE=",
		"GONOPROXY=",
		"GONOSUMDB=",
		"GOFLAGS=-mod=mod",
		"GOMODCACHE="+cache,
		"GOWORK=off",
		"GOTOOLCHAIN=local",
	)
}

// proxyURL spells a directory as the file URL the toolchain reads a proxy from.
// The leading slash is added where the path does not have one, which is every
// path on Windows and no path anywhere else.
func proxyURL(dir string) string {
	slashed := filepath.ToSlash(dir)
	if !strings.HasPrefix(slashed, "/") {
		slashed = "/" + slashed
	}
	return "file://" + slashed
}

// cleanTheModuleCache empties the temporary cache before the directory holding
// it is removed.
//
// The toolchain writes the cache read-only, and removing a read-only file is
// refused on one of the platforms this repository supports, so the cleanup the
// temporary directory does on its own fails there. This is registered after that
// cleanup and therefore runs before it.
func cleanTheModuleCache(t *testing.T, environment []string) {
	t.Helper()

	t.Cleanup(func() {
		clean := exec.Command("go", "clean", "-modcache")
		clean.Env = environment
		if output, err := clean.CombinedOutput(); err != nil {
			t.Logf("cannot empty the temporary module cache: %v\n%s", err, output)
		}
	})
}

// write is os.WriteFile with the test failed rather than the error returned,
// because every caller here would do the same three lines.
func write(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
