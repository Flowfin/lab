package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// fixtures is a case directory from the package's own harness, reused here so
// the command is proved against the same trees the rules are.
const fixtures = "../../testdata/contexts"

// TestTheCommandReturnsTheCodeItsVerdictEarns pins the contract record 0011
// sets. Anything keyed on one of these codes is reading that record whether or
// not anybody said so, and a gate that returned the same number for a
// disagreement and for a broken fetch would report one as the other.
func TestTheCommandReturnsTheCodeItsVerdictEarns(t *testing.T) {
	for _, one := range []struct {
		name     string
		required string
		dir      string
		want     int
	}{
		{
			name:     "the two lists agree",
			required: "the first check\nthe second check\n",
			dir:      filepath.Join(fixtures, "everything-agrees", "workflows"),
			want:     exitClean,
		},
		{
			name:     "a required context nothing reports",
			required: "the first check\nthe second check\na check nobody wrote\n",
			dir:      filepath.Join(fixtures, "a-required-context-nothing-reports", "workflows"),
			want:     exitRefused,
		},
		{
			name:     "there is no workflow directory to read",
			required: "the first check\n",
			dir:      filepath.Join(fixtures, "there-is-no-such-case", "workflows"),
			want:     exitCannot,
		},
		{
			name:     "what arrived is not a list of check names",
			required: "the first check\tand a column of something else\n",
			dir:      filepath.Join(fixtures, "everything-agrees", "workflows"),
			want:     exitCannot,
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			got := run(strings.NewReader(one.required), &out, &errOut, one.dir, nil)
			if got != one.want {
				t.Errorf("returned %d, want %d\nstdout: %s\nstderr: %s", got, one.want, out.String(), errOut.String())
			}
		})
	}
}

// TestTheReportSaysWhatItCompared refuses a run whose whole output is a verdict.
// A comparison of two empty lists and a comparison of two full ones both produce
// no refusals, and the only thing that tells them apart is the count the run
// printed.
func TestTheReportSaysWhatItCompared(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run(
		strings.NewReader("the first check\nthe second check\n"),
		&out, &errOut,
		filepath.Join(fixtures, "everything-agrees", "workflows"),
		nil,
	)
	if code != exitClean {
		t.Fatalf("returned %d\n%s", code, errOut.String())
	}
	for _, want := range []string{
		"required by the ruleset: 2",
		"declared by the workflows: 2",
		"written down as deliberately absent:",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the report does not say %q\n%s", want, out.String())
		}
	}
}

// TestAnEmptyRequiredSetIsCompared refuses a command that treats no required
// contexts as nothing to do. That is the state of this board today, and a check
// that switched itself off there would be a check that has never run.
func TestAnEmptyRequiredSetIsCompared(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run(
		strings.NewReader(""),
		&out, &errOut,
		filepath.Join(fixtures, "everything-agrees", "workflows"),
		nil,
	)
	if code != exitRefused {
		t.Fatalf("returned %d, and with an empty required set every declared name is outside it\n%s%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "required by the ruleset: 0") {
		t.Errorf("the report does not say the required set was empty\n%s", out.String())
	}
}
