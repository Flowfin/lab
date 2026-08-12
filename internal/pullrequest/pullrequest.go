// Package pullrequest judges a pull request rather than the tree the pull
// request lands. Every other check here reads a checkout: the record checks
// walk experiments/, the invariants read tracked text, the prose rules read the
// bytes around the words. None of them can see the change itself, and the class
// they all miss is whether the change says which problem it solves and whether
// the record moved with the code it describes.
//
// THE JUDGEMENT IS A FUNCTION OVER VALUES AND IT READS NOTHING. Judge takes a
// Change and returns a verdict. Nothing here opens a file, runs git or reads an
// environment variable, so every rule below is proved against a value a test
// writes out in full rather than against whatever the machine the suite runs on
// happened to contain. Where the values come from is read.go's half, and the
// command under cmd/pullrequest is the only place that talks to git at all.
//
// WHAT A GREEN VERDICT HERE DOES NOT SAY. It never says the change is good, and
// it never says the change was read. Each rule below is a property of the shape
// of a pull request rather than of its content: that a number was written down,
// not that it is the right number; that two files moved together, not that what
// they say agrees. A reviewer is what stands behind the rest, and a run that
// found nothing prints what it examined so that a reader can see which of the
// two they are holding.
package pullrequest

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// ExperimentsDir and RecordName are where an experiment and its record live.
// Record 0002 fixes both. They are written here rather than imported from the
// checker because this package judges paths in a diff and never a tree, and an
// import for two strings would tie a judgement about a change to a package that
// walks a filesystem.
const (
	ExperimentsDir = "experiments"
	RecordName     = "EXPERIMENT.md"
)

// ReadingBound is the number of changed lines above which the verdict carries a
// note. It is an annotation and never a refusal, and the difference is the whole
// design of the rule: a prototype legitimately arrives in one piece, and a size
// cap that reddened a build would be satisfied by splitting a change into pieces
// chosen to please a counter rather than to be read.
//
// The number is the one the corpus this board's rules come from uses. Nothing
// here claims it is the right number for every change; what it buys is that a
// reviewer meeting a large diff meets a line saying so before they start rather
// than halfway down.
const ReadingBound = 400

// The properties this package can refuse. A property is the rule, named once
// here and named nowhere else, so a fixture declaring what it expects and a
// refusal the judgement produced are the same string or they are not equal.
const (
	// BodyNamesNoIssue refuses a pull request whose body references no issue.
	// Every change here starts as an issue, and a body that names none leaves
	// the reason for the change in whoever wrote it rather than in the tree.
	// The reference is what a reader follows to find what was wrong, what the
	// evidence was and what done meant, and none of that is reconstructable
	// from a diff six months later.
	BodyNamesNoIssue = "pull-request-body-names-no-issue"

	// CommitMessageNamesNoIssue refuses a commit in the range whose message
	// references no issue anywhere in it. A pull-request body is edited
	// afterwards and the commit message is what survives into the log, so a
	// body carrying the only reference leaves a reader running git log with a
	// change that explains itself nowhere.
	//
	// THE RULE READS THE WHOLE MESSAGE AND ISSUE #24 ASKS FOR THE SUBJECT, and
	// the departure is measured rather than a preference. On the default
	// branch at the commit this landed on, four of forty-nine non-merge
	// commits carry a reference in the subject and thirty-five carry one
	// somewhere in the message:
	//
	//	git log --no-merges --format=%s origin/main | wc -l
	//	49
	//	git log --no-merges --format=%s origin/main |
	//	  grep -cE '(^|[^A-Za-z0-9_])([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)?#[1-9][0-9]*'
	//	4
	//
	// So the subject reading would refuse forty-five landed commits, thirty-one
	// of which name their issue in the message body, and the same issue asks in
	// the same paragraph that the check not red-flag reasonable work. The
	// failure it names is a change that explains itself nowhere, and a message
	// whose second paragraph names the issue is not that change. Tightening
	// this to the subject is a decision about how messages are written here
	// rather than about what this check can see.
	//
	// Merge commits are not judged. Their message is written by the platform
	// rather than by an author, and refusing one would refuse text nobody in
	// the pull request can edit.
	CommitMessageNamesNoIssue = "commit-message-names-no-issue"

	// ExperimentChangedWithoutItsRecord refuses a change that touches a file
	// under an experiment without touching that experiment's record in the
	// same pull request. This is the one rule of the three that is this
	// board's own rather than ordinary hygiene, and it is why the check is
	// worth having here instead of being copied for symmetry.
	//
	// The failure it prevents is the record going quietly out of date. An
	// experiment gains a script and the record still describes the old
	// measurement. An experiment's code is removed under record 0004, which
	// requires the record to gain a line naming the commit that removed it,
	// and the removal lands without that line. In both the tree is green, the
	// record reads as complete, and what it says is no longer true.
	//
	// REMOVALS COUNT AS TOUCHING, and the reason is record 0004 rather than a
	// wish to be strict. A change that deletes an experiment's prototype is
	// exactly the change that owes the record a line, so a rule that read only
	// additions and edits would be silent on the case the record was written
	// for.
	ExperimentChangedWithoutItsRecord = "experiment-changed-without-its-record"
)

// ChangeIsLargerThanOneReading is the note, and it is deliberately not one of
// the constants above. Nothing in this package can turn it into a refusal by
// accident: refusals and notes are different fields of the verdict and only one
// of them decides the exit code.
const ChangeIsLargerThanOneReading = "change-is-larger-than-one-reading"

// issueReference matches a reference to an issue in this repository or in
// another one on the same host, in either of the two spellings a person
// actually writes.
//
// The leading group is what keeps a fragment out. Without it the pattern reads
// the tail of a word ending in a hash and a number as a reference, and a body
// that mentioned a colour or an anchor would satisfy the rule by accident. A
// reference in the owner/repository spelling still matches, because the group
// is asked for before the owner rather than before the hash.
//
// The number may not begin with a zero. An issue is numbered from one, so a
// hash followed by a zero is not a reference to anything and the shape it
// usually is instead is a colour.
var issueReference = regexp.MustCompile(`(^|[^A-Za-z0-9_])(?:[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)?#[1-9][0-9]*`)

// The same reference written as a link, which is what a body composed in a
// browser usually carries.
//
// THE HOST IS COMPARED AND NOT MATCHED, and the pattern below holds the path
// only. A regular expression carrying a host name matches anywhere in a string
// unless it is anchored, so a link to somewhere else that happens to carry this
// host inside its own path would satisfy the rule, and the reader who followed
// it would arrive somewhere nobody meant. Comparing the beginning of a field
// against a fixed string cannot do that, and the pattern that is left is
// anchored at the front and has no host in it at all.
const issueLinkHost = "https://github.com/"

var issueLinkPath = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/(?:issues|pull)/[1-9][0-9]*`)

// namesAnIssueLink says whether a text carries a link to an issue.
//
// The text is cut into fields at whitespace and at the characters prose wraps a
// link in, because a link inside brackets or followed by a full stop is a link
// somebody wrote and a check that missed it would be met by a body that
// references its issue perfectly well.
func namesAnIssueLink(text string) bool {
	for _, field := range strings.FieldsFunc(text, isLinkDelimiter) {
		rest, found := strings.CutPrefix(field, issueLinkHost)
		if !found {
			continue
		}
		if issueLinkPath.MatchString(rest) {
			return true
		}
	}
	return false
}

// isLinkDelimiter says whether a rune ends a link written inside prose.
func isLinkDelimiter(r rune) bool {
	return unicode.IsSpace(r) || strings.ContainsRune("()<>[]{}\"'`,", r)
}

// NamesAnIssue says whether a text references an issue.
//
// WHAT IT DOES NOT JUDGE, and this is the whole of the residual. Whether the
// issue exists, whether it is open, whether it is the issue this change is
// about, and whether the reference was added to satisfy this check. Nothing
// that reads a string can answer any of those, and a reviewer is what stands
// behind them. What the rule converts is the case where no issue was named at
// all, which is the case it exists for and the only one this can separate from
// the rest.
func NamesAnIssue(text string) bool {
	return issueReference.MatchString(text) || namesAnIssueLink(text)
}

// A File is one path the change touched. The path is repository-relative and
// written with forward slashes, which is what git prints on every platform this
// board builds for, so nothing here has to know which platform read the diff.
type File struct {
	// Path is where the file is, relative to the root of the repository.
	Path string

	// Gone says the path is not present at the head of the range. It is
	// carried because a reader of a refusal wants to know which repair they
	// are making, and never because a removal is judged more leniently: the
	// record rule treats a removal as touching the experiment, for the reason
	// written at that property.
	Gone bool

	// From is where this file was moved from, and it is empty on every file
	// that was not moved. git reports a rename as one entry carrying two
	// paths, and the pairing between them is the whole of what the entry adds
	// over a removal beside an addition. Dropping it costs the rule over a
	// renamed experiment the second path its refusal has to name, and a
	// refusal that names only where a record used to be sends the reader
	// looking for where it went.
	From string
}

// A Commit is one commit in the range, as much of it as any rule here reads.
type Commit struct {
	// Hash is the commit, in full. It is what a refusal names, because a
	// subject is not unique and the repair is made against one commit.
	Hash string

	// Message is the whole message, subject and body together, as git printed
	// it.
	Message string
}

// Subject is the first line of the message, which is what a refusal quotes so
// that a reader recognises the commit without going and looking it up.
func (c Commit) Subject() string {
	subject, _, _ := strings.Cut(strings.TrimLeft(c.Message, "\n"), "\n")
	return strings.TrimSpace(subject)
}

// A Change is one pull request, reduced to what the rules read. Every field is
// data, so a test can write out a pull request that has never existed.
//
// The two present flags are why this is not a struct of plain values. A body
// that is empty and a body that could not be read are different statements, and
// collapsing them would let a run that read nothing refuse a pull request for
// carrying nothing. Where a flag is false the verdict carries a skip rather
// than a refusal.
type Change struct {
	// Body is the pull-request body as it was written.
	Body string

	// BodyRead says a body was read at all. It is false where the run was
	// given no pull request, which is not a violation of anything.
	BodyRead bool

	// Commits are the commits in the range, oldest first, merges already
	// dropped by whoever assembled this.
	Commits []Commit

	// CommitsRead says the range's commits were read at all.
	CommitsRead bool

	// Files are the paths the range touched.
	Files []File

	// FilesRead says the range's files were read at all. Where it is false
	// the record rule is skipped rather than passed, because a rule that
	// found no experiment in a diff it never read is a rule that says nothing.
	FilesRead bool

	// Records are the experiment records the change touched, at both ends of
	// the range. They are separate from Files because the rules over them read
	// bytes rather than paths, and reading the bytes at the base costs a
	// question per record that the path rules do not need.
	Records []RecordChange

	// RecordsRead says the records were read at all.
	RecordsRead bool

	// ChangedLines is how many lines the range added and removed together.
	ChangedLines int

	// LinesCounted says the count above was read. A diff whose files are all
	// binary produces no count, and reporting that as zero would print the
	// smallest possible change against the one input where the number is
	// unknown.
	LinesCounted bool

	// Uncounted names the files whose lines git could not count, which are
	// the binary ones. The count above is a lower bound whenever this is not
	// empty, and the note says so rather than quoting a total that is short by
	// an unknown amount.
	Uncounted []string
}

// A Refusal is one rule refusing one subject. The subject is carried separately
// from the detail so that a message can never be written without it, which is
// the shape the rest of this repository's checks already use.
type Refusal struct {
	Property string
	Subject  string
	Detail   string
}

// String leads with the subject, because somebody reading a red run is looking
// for the thing to open first.
func (r Refusal) String() string {
	return fmt.Sprintf("%s: %s (%s)", r.Subject, r.Detail, r.Property)
}

// A Note is something worth saying that is not a refusal. It never decides an
// exit code.
type Note struct {
	Property string
	Detail   string
}

// A Skip is a rule that was not applied and why. It exists so that a run which
// verified nothing cannot print the same thing as a run that verified
// everything and found it clean, which is the failure every green tick on a
// misconfigured gate is made of.
type Skip struct {
	// Rule is what was not applied, in the words the rule is named by.
	Rule string

	// Why is what was missing. It names the input rather than the mechanism,
	// because whoever reads it is deciding whether the gap matters to them.
	Why string
}

// A Verdict is what one judgement found. The three lists are separate because
// they are read by different people for different reasons, and only the first
// of them is a failure.
type Verdict struct {
	Refusals []Refusal
	Notes    []Note
	Skips    []Skip
}

// Properties returns the set of properties this verdict refused, which is what
// a case declares and what the tests compare. A property refused twice is one
// entry, so a fixture tripping one rule at two places still refuses exactly
// that rule.
func (v Verdict) Properties() []string {
	seen := make(map[string]bool, len(v.Refusals))
	var props []string
	for _, refusal := range v.Refusals {
		if !seen[refusal.Property] {
			seen[refusal.Property] = true
			props = append(props, refusal.Property)
		}
	}
	return props
}

// Judge holds a change to the rules above and returns what it found.
//
// The order the rules run in is the order a reader meets the change: the body
// first, because it is what a reviewer opens, then the commits, then the tree
// the change touched. Nothing depends on that order and no rule reads another
// rule's verdict, which is why a fixture can trip exactly one of them.
func Judge(change Change) Verdict {
	var verdict Verdict

	verdict.add(judgeBody(change))
	verdict.add(judgeCommits(change))
	verdict.add(judgeExperiments(change))
	verdict.add(judgeRecords(change))
	verdict.add(judgeRemovals(change))
	verdict.add(judgeSize(change))

	return verdict
}

// add folds one rule's verdict into the whole.
func (v *Verdict) add(other Verdict) {
	v.Refusals = append(v.Refusals, other.Refusals...)
	v.Notes = append(v.Notes, other.Notes...)
	v.Skips = append(v.Skips, other.Skips...)
}

// judgeBody holds the pull-request body to naming an issue.
func judgeBody(change Change) Verdict {
	if !change.BodyRead {
		return Verdict{Skips: []Skip{{
			Rule: BodyNamesNoIssue,
			Why:  "this run was given no pull-request body, so nothing was read to look for a reference in",
		}}}
	}
	if NamesAnIssue(change.Body) {
		return Verdict{}
	}
	return Verdict{Refusals: []Refusal{{
		Property: BodyNamesNoIssue,
		Subject:  "the pull-request body",
		Detail:   "it references no issue, so the reason for this change is not written anywhere a reader can follow",
	}}}
}

// judgeCommits holds every commit message in the range to naming an issue. It
// refuses once per commit rather than once for the range, because the repair is
// made against a particular message and a reader given one line saying some
// commit is wrong has to go and find which.
func judgeCommits(change Change) Verdict {
	if !change.CommitsRead {
		return Verdict{Skips: []Skip{{
			Rule: CommitMessageNamesNoIssue,
			Why:  "this run was given no commits, so no message was read",
		}}}
	}

	var verdict Verdict
	for _, commit := range change.Commits {
		if NamesAnIssue(commit.Message) {
			continue
		}
		verdict.Refusals = append(verdict.Refusals, Refusal{
			Property: CommitMessageNamesNoIssue,
			Subject:  commit.Hash,
			Detail:   fmt.Sprintf("its message references no issue, and a message is what survives into the log after a pull-request body has been edited, so this change explains itself nowhere afterwards. Its subject is %q", commit.Subject()),
		})
	}
	return verdict
}

// judgeExperiments holds a change that touches an experiment to touching that
// experiment's record.
//
// TWO BOUNDARIES, WRITTEN HERE RATHER THAN DISCOVERED.
//
// A file directly under experiments/ belongs to no experiment, so nothing here
// reads it. Whether such a file may exist at all is a rule about a tree and the
// checker holds it, which is the right place: this package sees a list of paths
// and never the directory they came from.
//
// A change that removes an experiment's record along with its files satisfies
// this rule, because the record is in the set of paths the change touched. That
// a landed record may not be removed at all is a different rule about a
// different failure, and it is record-already-landed-was-removed rather than
// this one. Nothing here should be read as permitting it.
func judgeExperiments(change Change) Verdict {
	if !change.FilesRead {
		return Verdict{Skips: []Skip{{
			Rule: ExperimentChangedWithoutItsRecord,
			Why:  "this run was given no changed paths, so no experiment was looked for",
		}}}
	}

	touched := make(map[string]bool)
	records := make(map[string]bool)
	for _, file := range change.Files {
		slug, isRecord, ok := experimentOf(file.Path)
		if !ok {
			continue
		}
		if isRecord {
			records[slug] = true
			continue
		}
		touched[slug] = true
	}

	slugs := make([]string, 0, len(touched))
	for slug := range touched {
		if !records[slug] {
			slugs = append(slugs, slug)
		}
	}
	// Sorted so that two runs over the same change print the same thing. A map
	// is walked in whatever order it is walked in, and a report that reordered
	// itself between runs would show as a change nobody made.
	sort.Strings(slugs)

	var verdict Verdict
	for _, slug := range slugs {
		record := strings.Join([]string{ExperimentsDir, slug, RecordName}, "/")
		verdict.Refusals = append(verdict.Refusals, Refusal{
			Property: ExperimentChangedWithoutItsRecord,
			Subject:  strings.Join([]string{ExperimentsDir, slug}, "/"),
			Detail:   fmt.Sprintf("this change touches files in it and does not touch %s, so the record still describes the experiment as it was before", record),
		})
	}
	return verdict
}

// IsRecordPath says whether a path is an experiment's record. It is exported
// because whoever assembles a Change has to know which of the paths in a diff
// are worth asking git about twice, and a second copy of that judgement in the
// command is a second place for it to be wrong.
func IsRecordPath(path string) bool {
	_, isRecord, ok := experimentOf(path)
	return ok && isRecord
}

// experimentOf places a path inside an experiment. It returns the slug, whether
// the path is that experiment's record, and whether the path is inside an
// experiment at all.
//
// It reads the path as git prints it, which is forward slashes on every
// platform, so a run on Windows places a path exactly as a run on Linux does.
func experimentOf(path string) (slug string, isRecord bool, ok bool) {
	segments := strings.Split(path, "/")
	if len(segments) < 3 || segments[0] != ExperimentsDir {
		return "", false, false
	}
	return segments[1], len(segments) == 3 && segments[2] == RecordName, true
}

// judgeSize annotates a change that is larger than one reading. It never
// refuses, for the reason written at ReadingBound.
func judgeSize(change Change) Verdict {
	if !change.LinesCounted {
		return Verdict{Skips: []Skip{{
			Rule: ChangeIsLargerThanOneReading,
			Why:  "this run was given no line counts, so the size of the change is unknown rather than small",
		}}}
	}

	var verdict Verdict
	if len(change.Uncounted) > 0 {
		verdict.Skips = append(verdict.Skips, Skip{
			Rule: ChangeIsLargerThanOneReading,
			Why: fmt.Sprintf("git counts no lines in %s, so %d is a lower bound rather than the size of this change",
				strings.Join(change.Uncounted, ", "), change.ChangedLines),
		})
	}
	if change.ChangedLines > ReadingBound {
		verdict.Notes = append(verdict.Notes, Note{
			Property: ChangeIsLargerThanOneReading,
			Detail: fmt.Sprintf("this change moves %d lines and %d is where one reading stops being one reading, which is worth knowing before the review starts rather than halfway down",
				change.ChangedLines, ReadingBound),
		})
	}
	return verdict
}

// Report is what the run prints. It says what it examined before it says what
// it found, whatever it found, because zero refusals over a change nobody read
// and zero refusals over a change that was read are the same line otherwise.
func (v Verdict) Report(change Change) string {
	out := "examined the pull request\n"
	out += fmt.Sprintf("  body: %s\n", describeBody(change))
	out += fmt.Sprintf("  commits: %s\n", describeCommits(change))
	out += fmt.Sprintf("  files: %s\n", describeFiles(change))
	out += fmt.Sprintf("  records: %s\n", describeRecords(change))
	out += fmt.Sprintf("  lines: %s\n", describeLines(change))

	for _, skip := range v.Skips {
		out += fmt.Sprintf("not judged: %s, because %s\n", skip.Rule, skip.Why)
	}
	for _, note := range v.Notes {
		out += fmt.Sprintf("note: %s (%s)\n", note.Detail, note.Property)
	}
	for _, refusal := range v.Refusals {
		out += "refused: " + refusal.String() + "\n"
	}

	out += fmt.Sprintf("%d refusal(s), %d note(s), %d rule(s) not judged\n",
		len(v.Refusals), len(v.Notes), len(v.Skips))
	return out
}

func describeBody(change Change) string {
	if !change.BodyRead {
		return "none was read"
	}
	return fmt.Sprintf("%d character(s)", len([]rune(change.Body)))
}

func describeCommits(change Change) string {
	if !change.CommitsRead {
		return "none were read"
	}
	return fmt.Sprintf("%d, merges excluded", len(change.Commits))
}

func describeFiles(change Change) string {
	if !change.FilesRead {
		return "none were read"
	}
	return fmt.Sprintf("%d", len(change.Files))
}

func describeRecords(change Change) string {
	if !change.RecordsRead {
		return "none were read"
	}
	landed := 0
	for _, record := range change.Records {
		if record.BeforePresent {
			landed++
		}
	}
	return fmt.Sprintf("%d, %d of them already on the branch this lands on", len(change.Records), landed)
}

func describeLines(change Change) string {
	if !change.LinesCounted {
		return "not counted"
	}
	if len(change.Uncounted) > 0 {
		return fmt.Sprintf("%d or more, because git counts no lines in %d of them", change.ChangedLines, len(change.Uncounted))
	}
	return fmt.Sprintf("%d", change.ChangedLines)
}
