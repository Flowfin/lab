package hardware

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestMain says what the run covered, and it says it at the end rather than at
// the start, because that is where a reader looks after the last line of
// output. On a default run that is the disclosure: this harness was not asked
// for and nothing in it ran. On a run that asked for it, it is the count of
// how many of its tests executed and how many did not, and what each one that
// did not was missing.
//
// The exit code comes from the same place. A run asked for and delivering
// nothing reports its own code, because no assertion breaking is not the same
// as coverage having been produced.
func TestMain(m *testing.M) {
	report, code := Report(m.Run())
	fmt.Print(report)
	os.Exit(code)
}

// TestTheHarnessIsNotInTheDefaultRun holds the exclusion rather than trusting
// it. Every test of this harness is behind the build constraint, so the
// default run does not compile them; this file has no constraint, so it runs
// either way and can say which of the two it is in.
func TestTheHarnessIsNotInTheDefaultRun(t *testing.T) {
	if Asked() {
		t.Skip("this run asked for the harness, so there is no default-run exclusion to check")
	}

	for _, name := range hardwareTestFiles(t) {
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("cannot read %s: %v", name, err)
		}
		// The first line, whatever ends it. A file that arrived through a
		// translating checkout is a separate problem with its own message,
		// and reporting it here as a missing constraint would send the
		// reader to the wrong repair.
		first, _, _ := strings.Cut(string(source), "\n")
		if strings.TrimSpace(first) != "//go:build "+BuildTag {
			t.Errorf("%s is a harness file and does not open with the %s build constraint, so the default run would compile it", name, BuildTag)
		}
	}
}

// TestEveryHardwareTestNamesItsHardware refuses a test in this harness that
// does not say what it needs. A failure that says a device was unavailable is
// useful and one that says an assertion failed is not, and the difference is
// whether the test named the hardware.
//
// It reads the harness source rather than running it, because the harness is
// not compiled into the default run and this test is. That is what makes the
// requirement reach a hardware test nobody has written yet.
func TestEveryHardwareTestNamesItsHardware(t *testing.T) {
	files := hardwareTestFiles(t)
	if len(files) == 0 {
		t.Fatalf("no file in this package is behind the %s constraint, which cannot be right while the harness exists", BuildTag)
	}

	fset := token.NewFileSet()
	found := 0

	for _, name := range files {
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("cannot parse %s: %v", name, err)
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Name.Name == "TestMain" {
				continue
			}
			found++

			if !callsNeeds(fn) {
				t.Errorf("%s in %s does not name the hardware it requires, so a reader who cannot run it is told nothing", fn.Name.Name, name)
			}
		}
	}

	if found == 0 {
		t.Error("the harness holds no test, so nothing here proves the requirement reaches one")
	}
}

// callsNeeds reports whether a test calls Needs, which is the one place a
// test names its hardware.
func callsNeeds(fn *ast.FuncDecl) bool {
	calls := false
	ast.Inspect(fn, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			if fun.Name == "Needs" {
				calls = true
			}
		case *ast.SelectorExpr:
			if fun.Sel.Name == "Needs" {
				calls = true
			}
		}
		return true
	})
	return calls
}

// hardwareTestFiles returns the files this harness's tests live in, found by
// their name rather than by a list. The naming rule matters more than the
// exact string: a file whose name says integration hardware is a file nobody
// mistakes for the default suite while reading a directory listing.
func hardwareTestFiles(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("cannot read this package: %v", err)
	}

	var files []string
	for _, entry := range entries {
		if name := entry.Name(); strings.HasSuffix(name, "_"+BuildTag+"_test.go") {
			files = append(files, name)
		}
	}
	return files
}

// TestNeedsRefusesATestThatNamesNothing proves the one thing Needs refuses on
// its own. The requirement above is read out of the source, so it catches a
// test with no call at all; this catches a call that names nothing, which the
// source reader cannot tell from a good one.
func TestNeedsRefusesATestThatNamesNothing(t *testing.T) {
	fake := &testing.T{}
	done := make(chan struct{})

	go func() {
		defer close(done)
		Needs(fake, "")
	}()
	<-done

	if !fake.Failed() {
		t.Error("Needs accepted a test that names no hardware")
	}
}
