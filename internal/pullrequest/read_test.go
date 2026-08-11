package pullrequest

import (
	"strings"
	"testing"
)

// The fixtures below are what git actually prints rather than what a reader of
// its documentation would expect it to print. The first two were copied out of
// a run against this repository, and the command and the range are written next
// to each so that anybody can produce them again:
//
//	git diff --name-status -z e9a798d...c0a1f23
//	git diff --numstat -z e9a798d...c0a1f23
//
// The commit fixtures are written out here rather than copied, because a real
// message in this repository is forty lines and what these prove is the
// separator rather than the prose. The format string they are written against
// is CommitFormat, which the command passes to git, so a change to one is a
// change to the other.
//
// Nothing in this file runs git. A test that shelled out would prove the parser
// against whatever version of git the machine happened to carry, and would need
// a repository to exist before it could run at all.

func TestParseFiles(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want []File
	}{
		{
			name: "an empty diff",
			out:  "",
		},
		{
			name: "four modified files, copied from a real range",
			out:  "M\x00.github/workflows/build.yml\x00M\x00.github/workflows/headless.yml\x00M\x00CONTRIBUTING.md\x00M\x00README.md\x00",
			want: []File{
				{Path: ".github/workflows/build.yml"},
				{Path: ".github/workflows/headless.yml"},
				{Path: "CONTRIBUTING.md"},
				{Path: "README.md"},
			},
		},
		{
			name: "an addition and a removal",
			out:  "A\x00experiments/one/EXPERIMENT.md\x00D\x00experiments/one/measure.go\x00",
			want: []File{
				{Path: "experiments/one/EXPERIMENT.md"},
				{Path: "experiments/one/measure.go", Gone: true},
			},
		},
		{
			// The shape a reader of the flag would get wrong. One entry, two
			// paths, and the file is gone at one end of the move and present at
			// the other. A parser reading one path per status would take the
			// new path as the next status and produce nonsense from there on.
			name: "a rename, which prints two paths for one entry",
			out:  "R100\x00experiments/one/measure.go\x00experiments/two/measure.go\x00",
			want: []File{
				{Path: "experiments/one/measure.go", Gone: true},
				{Path: "experiments/two/measure.go"},
			},
		},
		{
			name: "a path holding a space, which is not quoted under -z",
			out:  "M\x00docs/a document.md\x00",
			want: []File{{Path: "docs/a document.md"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseFiles([]byte(tc.out))
			if err != nil {
				t.Fatalf("cannot read the diff: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("read %d files %v, want %d", len(got), got, len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("file %d is %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestParseFilesRefusesATruncatedDiff holds the parser to saying it could not
// read rather than to returning what it managed. A status with no path after it
// is a stream that was cut, and returning three of four files from it would be
// a run that judged less than the change and reported nothing about the
// difference.
func TestParseFilesRefusesATruncatedDiff(t *testing.T) {
	if _, err := ParseFiles([]byte("M\x00CONTRIBUTING.md\x00M\x00")); err == nil {
		t.Fatal("a diff ending after a status with no path was read without complaint")
	}
	if _, err := ParseFiles([]byte("R100\x00experiments/one/measure.go\x00")); err == nil {
		t.Fatal("a rename missing its second path was read without complaint")
	}
}

func TestParseLines(t *testing.T) {
	tests := []struct {
		name          string
		out           string
		want          int
		wantUncounted []string
	}{
		{
			name: "an empty diff",
		},
		{
			name: "four modified files, copied from a real range",
			out:  "14\t0\t.github/workflows/build.yml\x009\t0\t.github/workflows/headless.yml\x0011\t1\tCONTRIBUTING.md\x0054\t30\tREADME.md\x00",
			want: 14 + 0 + 9 + 0 + 11 + 1 + 54 + 30,
		},
		{
			// A binary file prints a dash for each count. Treating that as zero
			// would report the smallest possible change against the one input
			// whose size is unknown.
			name:          "a binary file, which git counts no lines in",
			out:           "-\t-\tdocs/diagram.png\x0012\t3\tREADME.md\x00",
			want:          15,
			wantUncounted: []string{"docs/diagram.png"},
		},
		{
			// Under -z a rename prints its counts, then an empty field, then
			// the two paths as separate fields.
			name: "a rename, whose paths follow the counts",
			out:  "2\t1\t\x00experiments/one/measure.go\x00experiments/two/measure.go\x00",
			want: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, uncounted, err := ParseLines([]byte(tc.out))
			if err != nil {
				t.Fatalf("cannot read the counts: %v", err)
			}
			if got != tc.want {
				t.Errorf("counted %d lines, want %d", got, tc.want)
			}
			if strings.Join(uncounted, ",") != strings.Join(tc.wantUncounted, ",") {
				t.Errorf("uncounted is %v, want %v", uncounted, tc.wantUncounted)
			}
		})
	}
}

func TestParseCommits(t *testing.T) {
	out := "1111111111111111111111111111111111111111\x1fRefuse a header whose dates disagree with themselves\n" +
		"\nThe listing sorts on a field nothing parsed. Closes #54.\n\n" +
		"Signed-off-by: A Contributor <nobody@example.invalid>\n\x00" +
		"2222222222222222222222222222222222222222\x1fFix the message\n\x00"

	commits, err := ParseCommits([]byte(out))
	if err != nil {
		t.Fatalf("cannot read the commits: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("read %d commits, want 2", len(commits))
	}
	if commits[0].Hash != "1111111111111111111111111111111111111111" {
		t.Errorf("the first hash is %q", commits[0].Hash)
	}
	// The subject stops at the first newline and the message does not, which is
	// the whole reason both exist.
	if commits[0].Subject() != "Refuse a header whose dates disagree with themselves" {
		t.Errorf("the first subject is %q", commits[0].Subject())
	}
	if !strings.Contains(commits[0].Message, "Closes #54") {
		t.Error("the first message lost its body, which is where the reference is")
	}
	if commits[1].Subject() != "Fix the message" {
		t.Errorf("the second subject is %q", commits[1].Subject())
	}
}

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name          string
		payload       string
		isPullRequest bool
		body          string
		base, head    string
	}{
		{
			name:    "an event that is not a pull request",
			payload: `{"ref":"refs/heads/main"}`,
		},
		{
			name:          "a pull request with a body",
			payload:       `{"pull_request":{"body":"Closes #24.","base":{"sha":"aaa"},"head":{"sha":"bbb"}},"number":7}`,
			isPullRequest: true,
			body:          "Closes #24.",
			base:          "aaa",
			head:          "bbb",
		},
		{
			// An author who wrote no body at all gives null rather than an
			// empty string. It is a body that references no issue rather than a
			// body that was not read, and reading it as the second would let
			// the emptiest pull request pass.
			name:          "a pull request whose body is null",
			payload:       `{"pull_request":{"body":null,"base":{"sha":"aaa"},"head":{"sha":"bbb"}}}`,
			isPullRequest: true,
			base:          "aaa",
			head:          "bbb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event, err := ParseEvent([]byte(tc.payload))
			if err != nil {
				t.Fatalf("cannot read the payload: %v", err)
			}
			if event.IsPullRequest != tc.isPullRequest {
				t.Errorf("is a pull request is %v, want %v", event.IsPullRequest, tc.isPullRequest)
			}
			if event.Body != tc.body {
				t.Errorf("the body is %q, want %q", event.Body, tc.body)
			}
			if event.Base != tc.base || event.Head != tc.head {
				t.Errorf("the range is %q...%q, want %q...%q", event.Base, event.Head, tc.base, tc.head)
			}
		})
	}
}

// TestParseEventRefusesSomethingThatIsNotAPayload holds the reader to saying it
// could not read rather than to returning an empty event. An empty event is a
// run that judges nothing and reports success, which is the one outcome a gate
// may never produce by accident.
func TestParseEventRefusesSomethingThatIsNotAPayload(t *testing.T) {
	if _, err := ParseEvent([]byte("this is not a payload")); err == nil {
		t.Fatal("a payload that is not JSON was read without complaint")
	}
}

// TestABodyThatNamesNoIssueIsNotRescuedByTheEvent is the join between the two
// halves of this package: what the platform said, judged by the rule. It exists
// because the two are proved separately everywhere else, and a reader is
// entitled to see them meet once.
func TestABodyThatNamesNoIssueIsNotRescuedByTheEvent(t *testing.T) {
	event, err := ParseEvent([]byte(`{"pull_request":{"body":"Tidies the check up.","base":{"sha":"aaa"},"head":{"sha":"bbb"}}}`))
	if err != nil {
		t.Fatalf("cannot read the payload: %v", err)
	}

	verdict := Judge(Change{Body: event.Body, BodyRead: event.IsPullRequest})
	properties := verdict.Properties()
	if len(properties) != 1 || properties[0] != BodyNamesNoIssue {
		t.Fatalf("the verdict is %v, want exactly %s", properties, BodyNamesNoIssue)
	}
}
