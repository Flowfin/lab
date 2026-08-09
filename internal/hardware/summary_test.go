package hardware

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// These run in the default suite, where no hardware test is compiled at all.
// That is deliberate rather than awkward: what they hold is the reporting and
// the exit code of a harness run, and both have to be provable on a machine
// with none of the hardware, which is every machine the default run happens
// on.

// withTally runs f against a tally this test controls, and puts the harness's
// own back afterwards. The counters are package state because TestMain reads
// them after the last test has finished, and there is nowhere else for them to
// live.
func withTally(t *testing.T, ran int, missing map[string]string, f func()) {
	t.Helper()

	saved := runs.ran
	savedMissing := runs.missing
	t.Cleanup(func() {
		runs.ran = saved
		runs.missing = savedMissing
	})

	runs.ran = ran
	runs.missing = missing
	f()
}

// TestVerdictReachesEveryCodeThisHarnessCanReturn is what record 0011 asks of
// every code: a test in the default run that reaches it. No default run can
// produce a harness run that skipped everything, so the mapping is a function
// and this is what proves it before the day a job should go red.
func TestVerdictReachesEveryCodeThisHarnessCanReturn(t *testing.T) {
	for _, c := range []struct {
		name  string
		code  int
		asked bool
		ran   int
		want  int
	}{
		{"asked for, tests ran, nothing failed", 0, true, 2, 0},
		{"asked for, nothing ran", 0, true, 0, ExitAskedAndDeliveredNothing},
		{"asked for, nothing ran, and something failed as well", 1, true, 0, 1},
		{"asked for, tests ran, something failed", 1, true, 2, 1},
		{"not asked for, so nothing ran and that is ordinary", 0, false, 0, 0},
		{"not asked for, and the meta tests failed", 1, false, 0, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := Verdict(c.code, c.asked, c.ran); got != c.want {
				t.Errorf("Verdict(%d, %v, %d) is %d, want %d", c.code, c.asked, c.ran, got, c.want)
			}
		})
	}
}

// TestSummaryNamesEveryTestThatDidNotRunAndWhatItWasMissing is the whole point
// of the summary. A skip is reported in the same colour as a pass, so a run
// where everything skipped and a run where everything passed look the same
// from outside unless the missing thing is named.
func TestSummaryNamesEveryTestThatDidNotRunAndWhatItWasMissing(t *testing.T) {
	withTally(t, 1, map[string]string{
		"TestReadsFromTheDevice":  "a device on the serial port",
		"TestWritesToRealStorage": "the storage device the temporary directory is on",
	}, func() {
		summary := Summary()

		for _, wants := range []string{
			"1 of its tests ran and 2 did not",
			"TestReadsFromTheDevice, missing a device on the serial port",
			"TestWritesToRealStorage, missing the storage device the temporary directory is on",
			"was not verified on this machine",
		} {
			if !strings.Contains(summary, wants) {
				t.Errorf("the summary does not say %q. it says:\n%s", wants, summary)
			}
		}
	})
}

// TestSummarySaysWhenNothingRanAtAll holds the sharpest case. A harness that
// executed nothing is evidence about nothing, and the line that says so is the
// one somebody quoting a green run needs to have read.
func TestSummarySaysWhenNothingRanAtAll(t *testing.T) {
	withTally(t, 0, nil, func() {
		summary := Summary()
		if !strings.Contains(summary, "0 of its tests ran and 0 did not") {
			t.Errorf("the summary does not carry the counts. it says:\n%s", summary)
		}
		if !strings.Contains(summary, "evidence about nothing") {
			t.Errorf("the summary does not say that the run proves nothing. it says:\n%s", summary)
		}
	})
}

// TestSummaryOfARunThatCoveredEverythingClaimsNothingMore is the near
// neighbour. Without it, a summary that always printed the admission would
// pass the two tests above.
func TestSummaryOfARunThatCoveredEverythingClaimsNothingMore(t *testing.T) {
	withTally(t, 3, nil, func() {
		summary := Summary()
		if !strings.Contains(summary, "3 of its tests ran and 0 did not") {
			t.Errorf("the summary does not carry the counts. it says:\n%s", summary)
		}
		if strings.Contains(summary, "not verified") || strings.Contains(summary, "evidence about nothing") {
			t.Errorf("a run that covered everything printed an admission. it says:\n%s", summary)
		}
	})
}

// TestAbsentRefusesASkipThatNamesNothing holds the other half of a useful
// skip. Naming the hardware without saying how its absence was established
// tells a reader nothing they can act on, and naming neither is the state this
// whole issue is against.
func TestAbsentRefusesASkipThatNamesNothing(t *testing.T) {
	for _, c := range []struct{ name, hardware, how string }{
		{"neither", "", ""},
		{"no hardware", "", "the port is not there"},
		{"no reason", "a device on the serial port", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			fake := &testing.T{}
			done := make(chan struct{})
			go func() {
				defer close(done)
				Absent(fake, c.hardware, c.how)
			}()
			<-done

			if !fake.Failed() {
				t.Errorf("Absent accepted a skip naming %q and %q", c.hardware, c.how)
			}
		})
	}
}

// TestEveryHardwareSkipNamesWhatWasMissing is what makes the rule reach a test
// nobody has written yet. Needs and Absent both name the missing thing and
// both are counted by the tally; a bare t.Skip in a harness file names nothing
// and is invisible to the summary, which is the exact failure this is against.
//
// It reads the harness source rather than running it, because the harness is
// not compiled into the default run and this test is.
func TestEveryHardwareSkipNamesWhatWasMissing(t *testing.T) {
	files := hardwareTestFiles(t)
	if len(files) == 0 {
		t.Fatalf("no file in this package is behind the %s constraint, which cannot be right while the harness exists", BuildTag)
	}

	fset := token.NewFileSet()
	for _, name := range files {
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("cannot parse %s: %v", name, err)
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if sel.Sel.Name == "Skip" || sel.Sel.Name == "Skipf" || sel.Sel.Name == "SkipNow" {
				t.Errorf("%s skips directly at %s. a skip here has to go through Needs or Absent, which name what was missing and are counted in the summary",
					name, fset.Position(call.Pos()))
			}
			return true
		})
	}
}

// TestADefaultRunReportsTheDisclosureAndNoSummary holds the two routes apart.
// The disclosure is what a run that was not asked for owes; a summary there
// would report zero of zero tests as though the harness had been run.
func TestADefaultRunReportsTheDisclosureAndNoSummary(t *testing.T) {
	if Asked() {
		t.Skipf("%s was asked for on this run, so there is no default-run report to check", Name)
	}

	report, code := Report(0)
	if !strings.Contains(report, "was not asked for") {
		t.Errorf("the default-run report does not disclose that the harness was not asked for. it says:\n%s", report)
	}
	if strings.Contains(report, "of its tests ran") {
		t.Errorf("the default-run report carries a summary of a run that did not happen. it says:\n%s", report)
	}
	if code != 0 {
		t.Errorf("a default run reports %d, want 0. not asking for the harness is ordinary", code)
	}
}
