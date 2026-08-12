package pullrequest

// The other half of this package: turning what git and the platform print into
// the values Judge reads. Every function here takes bytes and returns a value,
// so a test can hand it the exact output a real command produced without the
// command being run, and the judgement can be proved against a pull request
// nobody has to open.
//
// WHY THE PARSING IS SEPARATED FROM THE JUDGEMENT AT ALL. A rule proved against
// a value somebody typed into a test is proved against a shape somebody
// imagined. These functions are where a real diff enters, and the fixtures
// below them are copies of what git actually prints, including the two forms
// that are easy to get wrong and that no honest change produces often enough to
// find by accident: a rename, which prints two paths for one entry, and a
// binary file, which prints no line count at all.

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// fieldSeparator is what the commit format below puts between the hash and the
// subject. It is a control character rather than a space or a tab because a
// subject may hold either of those and may not hold this, so the split cannot
// be made at the wrong place by an ordinary message.
const fieldSeparator = "\x1f"

// CommitFormat is the format string the command passes to git log. It lives
// here, next to the function that reads its output, because a format changed in
// one place and read in another is a parser that goes wrong quietly on the
// first commit whose message contains whatever the two now disagree about.
//
// It asks for the whole message rather than the subject, which is what the rule
// over commits reads and why. A message holds newlines, and that is the reason
// the entries are separated by a null byte rather than by one.
const CommitFormat = "%H" + fieldSeparator + "%B"

// An Event is what the platform said about the run, reduced to what this
// package reads.
type Event struct {
	// IsPullRequest says the payload described a pull request at all. A run
	// on anything else is not a violation of anything and produces skips
	// rather than refusals.
	IsPullRequest bool

	// Body is the pull-request body. An author who wrote none gives an empty
	// string here, which is a body that references no issue rather than a body
	// that was not read.
	Body string

	// Base and Head are the two ends of the range, as the platform reported
	// them. The range is read three-dot from base to head, so what is judged
	// is what the head added since the two last agreed rather than everything
	// that happened on the base meanwhile.
	Base string
	Head string
}

// ParseEvent reads a webhook payload.
//
// It reads four fields out of a document with hundreds and ignores the rest,
// which is deliberate: every field this does not read is a field the platform
// may rename without breaking the check, and a parser that insisted on the
// whole shape would be a parser that fails on a payload it does not need to
// understand.
func ParseEvent(data []byte) (Event, error) {
	var payload struct {
		PullRequest *struct {
			Body *string `json:"body"`
			Base struct {
				SHA string `json:"sha"`
			} `json:"base"`
			Head struct {
				SHA string `json:"sha"`
			} `json:"head"`
		} `json:"pull_request"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return Event{}, fmt.Errorf("cannot read the event payload: %w", err)
	}
	if payload.PullRequest == nil {
		return Event{}, nil
	}

	event := Event{
		IsPullRequest: true,
		Base:          payload.PullRequest.Base.SHA,
		Head:          payload.PullRequest.Head.SHA,
	}
	// A body nobody wrote arrives as null rather than as an empty string, and
	// the two mean the same thing here: nothing was written, so nothing
	// references an issue.
	if payload.PullRequest.Body != nil {
		event.Body = *payload.PullRequest.Body
	}
	return event, nil
}

// ParseFiles reads the output of git diff --name-status -z.
//
// The separator is a null byte rather than a newline because git quotes a path
// holding an unusual character when it prints one per line, and a check that
// unquoted it would have to be right about git's quoting rules on every
// platform. Under -z nothing is quoted and a path is exactly its bytes.
//
// A rename produces two paths for one entry, and both are returned. The old one
// is gone at the head and the new one is not, which is what the record rule
// needs: moving a file out of an experiment is a change to that experiment
// whichever end of the move a reader looks at. The new one also carries where
// it came from, because a removal beside an addition and a move are the same
// two paths otherwise, and the rule over a renamed experiment has to name both
// ends of the move.
func ParseFiles(data []byte) ([]File, error) {
	fields := splitNul(data)
	var files []File

	for i := 0; i < len(fields); i++ {
		status := fields[i]
		if status == "" {
			continue
		}
		paths := 1
		if status[0] == 'R' || status[0] == 'C' {
			paths = 2
		}
		if i+paths > len(fields)-1 {
			return nil, fmt.Errorf("the diff ends after the status %q with no path after it", status)
		}
		switch paths {
		case 1:
			files = append(files, File{Path: fields[i+1], Gone: status[0] == 'D'})
		case 2:
			files = append(files,
				File{Path: fields[i+1], Gone: true},
				File{Path: fields[i+2], Gone: false, From: fields[i+1]})
		}
		i += paths
	}
	return files, nil
}

// ParseCommits reads the output of git log -z with CommitFormat.
//
// Merges are dropped by the command that runs git rather than here, because
// which commits are in the range is a question about the range and this
// function's whole job is reading what it was given.
func ParseCommits(data []byte) ([]Commit, error) {
	var commits []Commit

	for _, entry := range splitNul(data) {
		if strings.TrimSpace(entry) == "" {
			continue
		}
		// git puts a newline after each message under this format, so the
		// entry is trimmed at its ends rather than assumed clean. Nothing
		// inside the message is touched: what a rule reads has to be what the
		// author wrote.
		entry = strings.Trim(entry, "\n")
		hash, message, found := strings.Cut(entry, fieldSeparator)
		if !found {
			return nil, fmt.Errorf("a commit entry carries no separator: %q", entry)
		}
		commits = append(commits, Commit{Hash: hash, Message: message})
	}
	return commits, nil
}

// ParseLines reads the output of git diff --numstat -z and returns how many
// lines the range moved, together with the paths git counted no lines in.
//
// A binary file prints a dash for each count. It is returned rather than
// treated as zero, because a change of one binary blob and nothing else would
// otherwise report a size of zero, which is the smallest possible change
// printed against the one input whose size is unknown.
func ParseLines(data []byte) (int, []string, error) {
	fields := splitNul(data)
	total := 0
	var uncounted []string

	for i := 0; i < len(fields); i++ {
		entry := fields[i]
		if strings.TrimSpace(entry) == "" {
			continue
		}
		added, rest, found := strings.Cut(entry, "\t")
		if !found {
			return 0, nil, fmt.Errorf("a numstat entry carries no tab: %q", entry)
		}
		removed, path, found := strings.Cut(rest, "\t")
		if !found {
			return 0, nil, fmt.Errorf("a numstat entry carries one tab and needs two: %q", entry)
		}
		// A renamed file prints its two paths as the two fields after the
		// counts, so the entry itself ends at the second tab.
		if path == "" {
			if i+2 > len(fields)-1 {
				return 0, nil, fmt.Errorf("a renamed numstat entry ends before its paths: %q", entry)
			}
			path = fields[i+2]
			i += 2
		}

		if added == "-" || removed == "-" {
			uncounted = append(uncounted, path)
			continue
		}
		a, err := strconv.Atoi(added)
		if err != nil {
			return 0, nil, fmt.Errorf("cannot read the added count in %q: %w", entry, err)
		}
		r, err := strconv.Atoi(removed)
		if err != nil {
			return 0, nil, fmt.Errorf("cannot read the removed count in %q: %w", entry, err)
		}
		total += a + r
	}
	return total, uncounted, nil
}

// splitNul splits on the null byte and drops a trailing empty field, which is
// what a stream ending in a separator leaves behind.
func splitNul(data []byte) []string {
	fields := strings.Split(string(data), "\x00")
	for len(fields) > 0 && fields[len(fields)-1] == "" {
		fields = fields[:len(fields)-1]
	}
	return fields
}
