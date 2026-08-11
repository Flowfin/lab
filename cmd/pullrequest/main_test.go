package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// The whole command, exercised without a pull request, a platform or a
// repository, because both of its edges are parameters. What a test cannot
// prove this way is that the real git prints what the fixtures below say it
// prints, and that is proved in internal/pullrequest against output copied out
// of a run, with the commands written next to it.

// event is a payload in the shape the platform writes, with the two ends of the
// range and a body.
func event(body string) []byte {
	return []byte(fmt.Sprintf(
		`{"pull_request":{"body":%q,"base":{"sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"head":{"sha":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}}`,
		body))
}

// gitSaying answers the questions the command asks, in whatever order it asks
// them, and records what it was asked so a test can hold the range to its shape.
func gitSaying(files, log, numstat string, asked *[][]string) func(...string) ([]byte, error) {
	return gitSayingWithBlobs(files, log, numstat, nil, asked)
}

// gitSayingWithBlobs is the same with a tree behind it. The keys are what git
// show is asked for, `<revision>:<path>`, and a key that is not there is a path
// that is not in the tree at that revision, which is the answer ls-tree gives
// as an empty list rather than as a failure.
func gitSayingWithBlobs(files, log, numstat string, blobs map[string]string, asked *[][]string) func(...string) ([]byte, error) {
	return func(args ...string) ([]byte, error) {
		if asked != nil {
			*asked = append(*asked, args)
		}
		switch {
		case args[0] == "log":
			return []byte(log), nil
		case args[0] == "show":
			return []byte(blobs[args[1]]), nil
		case args[0] == "ls-tree":
			// ls-tree -z --name-only <revision> -- <path>
			revision, path := args[3], args[5]
			if _, present := blobs[revision+":"+path]; !present {
				return nil, nil
			}
			return []byte(path + "\x00"), nil
		case contains(args, "--numstat"):
			return []byte(numstat), nil
		default:
			return []byte(files), nil
		}
	}
}

func contains(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

const cleanLog = "1111111111111111111111111111111111111111\x1fAdd the check\n\nCloses #24.\n\x00"

// TestTheCommandReachesEveryCodeItCanReturn is record 0011's third clause for
// this command. A code that exists only in the comment above it is a promise,
// and the one that would be discovered by an operator is the one nothing here
// produces.
func TestTheCommandReachesEveryCodeItCanReturn(t *testing.T) {
	tests := []struct {
		name string
		e    edges
		want int
	}{
		{
			name: "a pull request that breaks nothing",
			e: edges{
				event: func() ([]byte, error) { return event("Closes #24."), nil },
				git: gitSaying(
					"M\x00internal/pullrequest/pullrequest.go\x00",
					cleanLog,
					"12\t3\tinternal/pullrequest/pullrequest.go\x00", nil),
			},
			want: exitClean,
		},
		{
			name: "a pull request that changes an experiment and not its record",
			e: edges{
				event: func() ([]byte, error) { return event("Closes #24."), nil },
				git: gitSaying(
					"M\x00experiments/one/measure.go\x00",
					cleanLog,
					"12\t3\texperiments/one/measure.go\x00", nil),
			},
			want: exitRefused,
		},
		{
			name: "an environment with no event in it",
			e: edges{
				event: func() ([]byte, error) { return nil, fmt.Errorf("no event here") },
				git:   gitSaying("", "", "", nil),
			},
			want: exitCannot,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if got := run(&out, &errOut, tc.e); got != tc.want {
				t.Fatalf("returned %d, want %d, having printed %q and %q", got, tc.want, out.String(), errOut.String())
			}
		})
	}
}

// TestTheRecordRuleIsRedAndGreenWhereIssue24SaysItShouldBe is the done-condition
// of that issue at the level the command runs at. The two pull requests differ
// in one file and in nothing else, which is what makes the pair evidence rather
// than two runs that happen to have come out differently.
func TestTheRecordRuleIsRedAndGreenWhereIssue24SaysItShouldBe(t *testing.T) {
	withoutTheRecord := edges{
		event: func() ([]byte, error) { return event("Closes #24."), nil },
		git: gitSaying(
			"M\x00experiments/one/measure.go\x00",
			cleanLog,
			"12\t3\texperiments/one/measure.go\x00", nil),
	}
	withTheRecord := edges{
		event: func() ([]byte, error) { return event("Closes #24."), nil },
		git: gitSaying(
			"M\x00experiments/one/measure.go\x00M\x00experiments/one/EXPERIMENT.md\x00",
			cleanLog,
			"12\t3\texperiments/one/measure.go\x004\t0\texperiments/one/EXPERIMENT.md\x00", nil),
	}

	var red, redErr bytes.Buffer
	if got := run(&red, &redErr, withoutTheRecord); got != exitRefused {
		t.Fatalf("a change that edits an experiment without its record returned %d, want %d", got, exitRefused)
	}
	if !strings.Contains(red.String(), "experiments/one") {
		t.Errorf("the refusal does not name the experiment: %q", red.String())
	}

	var green, greenErr bytes.Buffer
	if got := run(&green, &greenErr, withTheRecord); got != exitClean {
		t.Fatalf("a change that edits an experiment and its record returned %d, want %d: %q %q", got, exitClean, green.String(), greenErr.String())
	}
}

// TestTheRangeIsAskedForThreeDots holds the command to asking about what this
// branch added rather than about everything that happened on the base
// meanwhile. Two dots for the files would refuse a pull request for somebody
// else's commit, which is the failure that gets a check switched off.
func TestTheRangeIsAskedForThreeDots(t *testing.T) {
	var asked [][]string
	e := edges{
		event: func() ([]byte, error) { return event("Closes #24."), nil },
		git:   gitSaying("", cleanLog, "", &asked),
	}

	var out, errOut bytes.Buffer
	if got := run(&out, &errOut, e); got != exitClean {
		t.Fatalf("returned %d: %q %q", got, out.String(), errOut.String())
	}
	if len(asked) != 3 {
		t.Fatalf("git was asked %d questions, want 3: %v", len(asked), asked)
	}

	base, head := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	for _, args := range asked {
		last := args[len(args)-1]
		want := base + "..." + head
		if args[0] == "log" {
			want = base + ".." + head
		}
		if last != want {
			t.Errorf("git %s was asked about %q, want %q", args[0], last, want)
		}
	}
}

// TestARangeEndThatIsNotAnObjectNameIsRefusedBeforeGitIsAsked is the near miss
// worth the fixture. A value beginning with a dash is read by git as a flag,
// and a check whose behaviour a payload decides is not a check.
func TestARangeEndThatIsNotAnObjectNameIsRefusedBeforeGitIsAsked(t *testing.T) {
	payload := []byte(`{"pull_request":{"body":"Closes #24.","base":{"sha":"--output=/tmp/x"},"head":{"sha":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}}`)

	asked := 0
	e := edges{
		event: func() ([]byte, error) { return payload, nil },
		git: func(args ...string) ([]byte, error) {
			asked++
			return nil, nil
		},
	}

	var out, errOut bytes.Buffer
	if got := run(&out, &errOut, e); got != exitCannot {
		t.Fatalf("returned %d, want %d", got, exitCannot)
	}
	if asked != 0 {
		t.Errorf("git was asked %d questions about a range end that is not an object name", asked)
	}
}

// TestAnEventThatIsNotAPullRequestCannotBeGreen holds the command to saying it
// could not do its job rather than to reporting a clean pull request it never
// saw. A green tick on a gate pointed at the wrong event is the shape a
// misconfiguration hides behind.
func TestAnEventThatIsNotAPullRequestCannotBeGreen(t *testing.T) {
	e := edges{
		event: func() ([]byte, error) { return []byte(`{"ref":"refs/heads/main"}`), nil },
		git:   gitSaying("", "", "", nil),
	}

	var out, errOut bytes.Buffer
	if got := run(&out, &errOut, e); got != exitCannot {
		t.Fatalf("returned %d, want %d", got, exitCannot)
	}
	if !strings.Contains(errOut.String(), "not a pull request") {
		t.Errorf("the run does not say why it stopped: %q", errOut.String())
	}
}

// TestGitFailingIsNotACleanRun holds the command to stopping at the first
// question git will not answer. A shallow checkout has no merge base, and a run
// that judged the two questions it could answer would report a clean pull
// request having read a third of it.
func TestGitFailingIsNotACleanRun(t *testing.T) {
	e := edges{
		event: func() ([]byte, error) { return event("Closes #24."), nil },
		git: func(args ...string) ([]byte, error) {
			return nil, fmt.Errorf("fatal: no merge base")
		},
	}

	var out, errOut bytes.Buffer
	if got := run(&out, &errOut, e); got != exitCannot {
		t.Fatalf("returned %d, want %d", got, exitCannot)
	}
	if strings.Contains(out.String(), "0 refusal(s)") {
		t.Error("a run that could not read the range printed a clean verdict")
	}
}

// TestTheReportSaysWhatItExamined holds the run that found nothing wrong to
// saying what it read. A green tick that printed nothing and a green tick over a
// whole pull request are the same tick, and only the report separates them.
func TestTheReportSaysWhatItExamined(t *testing.T) {
	e := edges{
		event: func() ([]byte, error) { return event("Closes #24."), nil },
		git: gitSaying(
			"M\x00internal/pullrequest/pullrequest.go\x00",
			cleanLog,
			"12\t3\tinternal/pullrequest/pullrequest.go\x00", nil),
	}

	var out, errOut bytes.Buffer
	if got := run(&out, &errOut, e); got != exitClean {
		t.Fatalf("returned %d: %q %q", got, out.String(), errOut.String())
	}
	for _, want := range []string{"examined the pull request", "commits: 1", "files: 1", "lines: 15"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the report does not say %q:\n%s", want, out.String())
		}
	}
}

// theLandedRecord is an experiment record in state answered, as it stands on
// the branch a change lands on.
const theLandedRecord = "Slug: one\nState: answered\nQuestion-Written: 2026-08-01\n\n" +
	"## Question\n\nDoes the walk cost more than a second?\n\n" +
	"## Method\n\nTimed it.\n\n" +
	"## Answer\n\nNo. It took eleven seconds.\n"

// TestAnAnswerAlreadyLandedIsReadFromBothEndsOfTheRange is the whole route for
// the rule that a landed answer may grow and may not change: the command asks
// git for the record at each end of the range, and the judgement compares them.
// The two pull requests below differ in the answer and in nothing else.
func TestAnAnswerAlreadyLandedIsReadFromBothEndsOfTheRange(t *testing.T) {
	const base = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const head = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	const path = "experiments/one/EXPERIMENT.md"

	rewritten := strings.Replace(theLandedRecord, "It took eleven seconds.", "It took two seconds.", 1)
	addedTo := theLandedRecord + "\nCorrected on 2026-09-01. Eleven seconds was a cold cache.\n"

	tests := []struct {
		name string
		at   string
		want int
	}{
		{name: "the answer is rewritten", at: rewritten, want: exitRefused},
		{name: "the answer is added to", at: addedTo, want: exitClean},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := edges{
				event: func() ([]byte, error) { return event("Closes #58."), nil },
				git: gitSayingWithBlobs(
					"M\x00"+path+"\x00",
					cleanLog,
					"2\t2\t"+path+"\x00",
					map[string]string{
						base + ":" + path: theLandedRecord,
						head + ":" + path: tc.at,
					}, nil),
			}

			var out, errOut bytes.Buffer
			if got := run(&out, &errOut, e); got != tc.want {
				t.Fatalf("returned %d, want %d: %s %s", got, tc.want, out.String(), errOut.String())
			}
			if !strings.Contains(out.String(), "records: 1, 1 of them already on the branch this lands on") {
				t.Errorf("the report does not say what it read:\n%s", out.String())
			}
		})
	}
}
