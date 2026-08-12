// Command pullrequest judges a pull request and is the only thing in this
// repository that talks to git. It reads the event the platform wrote, asks git
// what the range contains, hands both to internal/pullrequest, prints what it
// examined and returns what record 0011 says it should return.
//
// WHY THIS IS NOT A VERB OF THE RUNNER. lab reads a checkout and writes
// nothing, which is a claim a downloader is asked to take on trust, and every
// document here repeats it. Telling a change from the tree it lands needs both
// ends of a range, which is history rather than a checkout, and the only thing
// that can answer about history is git. Putting that inside lab would ship a
// verb that is useless everywhere except inside a workflow and would widen what
// an operator has to believe about a binary they downloaded. Keeping it here
// costs a second entry point under cmd/, which record 0002 does not name, and
// that is the one judgement in this file a reader should check rather than
// accept.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Flowfin/lab/internal/pullrequest"
)

// The exit codes. Decision record 0011 is the contract, and this command
// returns the same three the runner does, for the same reasons, which is why
// they are written here with the record named rather than invented. The
// integration-hardware harness declares its own code where its producer is, and
// this follows that shape: a code lives next to the thing that can return it.
//
// WHAT HOLDS THIS DECLARATION TO THE OTHERS. The exit-code leg in
// internal/invariants reads every exit-code constant in the tree and refuses a
// code declared with two numbers, or a number declared under two codes. Without
// it these three numbers being the same as the runner's is a thing a reader
// checked once, and a number that moved here would compile, pass this package's
// own tests and change the meaning of anything keyed on it. It finds a
// declaration by the shape of its name, so keep the convention these three
// follow.
const (
	// exitClean means the run completed and refused nothing. It does not mean
	// the pull request is good, only that nothing this check judges was found
	// wrong, and the report says what it examined so the two are not confused.
	exitClean = 0

	// exitRefused means the run completed and refused something. It is the
	// only code that carries refusals.
	exitRefused = 1

	// exitCannot means the check could not do its job: no event to read, a
	// payload it cannot parse, a range git will not answer about. It is
	// deliberately not the code a refusal returns, because a gate that treats
	// a broken invocation like a violation reports one as the other and
	// nobody investigates either.
	exitCannot = 2
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr, edges{
		event: eventFromTheEnvironment,
		git:   gitInThisCheckout,
	}))
}

// edges is everything this command reaches for that is not its own logic: the
// event the platform wrote and git. Both are parameters so that what the
// command prints and what it returns can be read by a test without a pull
// request, a platform or a repository existing.
type edges struct {
	event func() ([]byte, error)
	git   func(args ...string) ([]byte, error)
}

// run is main with its edges passed in.
func run(out, errOut io.Writer, e edges) int {
	data, err := e.event()
	if err != nil {
		fmt.Fprintf(errOut, "pullrequest: %v\n", err)
		return exitCannot
	}

	event, err := pullrequest.ParseEvent(data)
	if err != nil {
		fmt.Fprintf(errOut, "pullrequest: %v\n", err)
		return exitCannot
	}
	if !event.IsPullRequest {
		// Asked to judge a pull request where there is none. This is a broken
		// invocation rather than a clean run, and the difference matters more
		// here than anywhere else in this repository: a check that answered
		// zero refusals to this would be a green tick on a gate that had been
		// pointed at the wrong event, which is the failure issue #62 is about.
		fmt.Fprintf(errOut, "pullrequest: this event is not a pull request, so there was nothing to judge\n")
		return exitCannot
	}
	for _, end := range []string{event.Base, event.Head} {
		if !looksLikeAnObjectName(end) {
			// A range end is read out of a payload and handed to git as an
			// argument. It arrives from the platform rather than from an
			// author, so this is a bound rather than a defence, and it is
			// cheap: a value beginning with a dash would be read by git as a
			// flag, and a check that passed one on would be a check whose
			// behaviour a payload decides.
			fmt.Fprintf(errOut, "pullrequest: %q is not an object name, so the range was not read\n", end)
			return exitCannot
		}
	}

	change, err := readChange(event, e.git)
	if err != nil {
		fmt.Fprintf(errOut, "pullrequest: %v\n", err)
		return exitCannot
	}

	verdict := pullrequest.Judge(change)
	fmt.Fprint(out, verdict.Report(change))
	if len(verdict.Refusals) > 0 {
		return exitRefused
	}
	return exitClean
}

// readChange asks git the three questions this check has, and stops at the
// first one git will not answer.
//
// It stops rather than judging what it managed to read. A range git cannot
// answer about is usually a checkout without enough history for the merge base,
// and a run that quietly judged the two questions it could answer would report
// a clean pull request having read a third of it.
func readChange(event pullrequest.Event, git func(args ...string) ([]byte, error)) (pullrequest.Change, error) {
	change := pullrequest.Change{Body: event.Body, BodyRead: true}

	// Three dots rather than two for the files and the counts, so what is
	// judged is what this branch added since the two ends last agreed rather
	// than everything that happened on the base meanwhile. Somebody else's
	// commit landing on the base is not this pull request's change.
	span := event.Base + "..." + event.Head

	// Rename detection is asked for rather than relied on. git turns it on by
	// default, so leaving it out reads the same on almost every machine and
	// differently on one that set diff.renames off, and the rule over a
	// renamed experiment would then refuse a move as a removal for a reason
	// living in somebody's git config.
	out, err := git("diff", "--name-status", "--find-renames", "-z", span)
	if err != nil {
		return change, err
	}
	files, err := pullrequest.ParseFiles(out)
	if err != nil {
		return change, err
	}
	change.Files = files
	change.FilesRead = true

	out, err = git("log", "--no-merges", "-z", "--format="+pullrequest.CommitFormat, event.Base+".."+event.Head)
	if err != nil {
		return change, err
	}
	commits, err := pullrequest.ParseCommits(out)
	if err != nil {
		return change, err
	}
	change.Commits = commits
	change.CommitsRead = true

	records, err := readRecords(event, git, files)
	if err != nil {
		return change, err
	}
	change.Records = records
	change.RecordsRead = true

	out, err = git("diff", "--numstat", "-z", span)
	if err != nil {
		return change, err
	}
	lines, uncounted, err := pullrequest.ParseLines(out)
	if err != nil {
		return change, err
	}
	change.ChangedLines = lines
	change.Uncounted = uncounted
	change.LinesCounted = true

	return change, nil
}

// readRecords reads every experiment record the change touched, at both ends of
// the range.
//
// It reads them out of git rather than off the disk. What is checked out on a
// pull request is a merge of the two ends rather than either of them, so a file
// read from the working tree is a third thing, and a rule about what a change
// did to a record cannot be made against it.
//
// Presence is asked before the bytes are. A record this change creates has no
// version at the base, and a command that read a missing path as a failure
// would report the change this board wants most as a run that could not be
// made.
func readRecords(event pullrequest.Event, git func(args ...string) ([]byte, error), files []pullrequest.File) ([]pullrequest.RecordChange, error) {
	seen := make(map[string]bool)
	var records []pullrequest.RecordChange

	for _, file := range files {
		if !pullrequest.IsRecordPath(file.Path) || seen[file.Path] {
			continue
		}
		seen[file.Path] = true

		record := pullrequest.RecordChange{Path: file.Path}
		for _, end := range []struct {
			rev     string
			bytes   *[]byte
			present *bool
		}{
			{event.Base, &record.Before, &record.BeforePresent},
			{event.Head, &record.After, &record.AfterPresent},
		} {
			present, err := pathIsAt(git, end.rev, file.Path)
			if err != nil {
				return nil, err
			}
			*end.present = present
			if !present {
				continue
			}
			data, err := git("show", end.rev+":"+file.Path)
			if err != nil {
				return nil, err
			}
			*end.bytes = data
		}
		records = append(records, record)
	}
	return records, nil
}

// pathIsAt says whether a path is in the tree at a revision. It asks a question
// git answers with an empty list rather than with a failure, so a missing path
// and a broken invocation stay different outcomes.
func pathIsAt(git func(args ...string) ([]byte, error), rev, path string) (bool, error) {
	out, err := git("ls-tree", "-z", "--name-only", rev, "--", path)
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// looksLikeAnObjectName says whether a string is the shape git prints for a
// commit. Nothing here asks whether the object exists, which is git's answer
// and arrives as a failure from the command that needed it.
func looksLikeAnObjectName(name string) bool {
	if len(name) < 7 || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// eventFromTheEnvironment reads the payload the platform wrote for this run.
func eventFromTheEnvironment() ([]byte, error) {
	path := os.Getenv("GITHUB_EVENT_PATH")
	if path == "" {
		return nil, fmt.Errorf("there is no GITHUB_EVENT_PATH in this environment, and this check reads a pull request out of the event the platform writes")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the event at %s: %w", path, err)
	}
	return data, nil
}

// gitInThisCheckout runs git with the arguments it is given and nothing else.
//
// The arguments are a list rather than a string, so nothing here is parsed by a
// shell and a value holding a space is one argument whatever it holds. That is
// the whole of why this is a function and not a line in a workflow: the scope
// gate that word-split its input and refused valid work is the failure this
// shape does not have.
func gitInThisCheckout(args ...string) ([]byte, error) {
	command := exec.Command("git", args...)
	out, err := command.Output()
	if err != nil {
		var stderr string
		if exit, ok := err.(*exec.ExitError); ok {
			stderr = strings.TrimSpace(string(exit.Stderr))
		}
		return nil, fmt.Errorf("git %s failed: %v: %s", strings.Join(args, " "), err, stderr)
	}
	return out, nil
}
