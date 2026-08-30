# Quality parity with the sso gate

Being goal-free is not being rule-free. The bar this board holds itself to is
the merge gate of another repository in this organisation rather than one
invented here, because a bar invented here would be set by the same people it
constrains.

The target is `Flowfin/jellyfin-plugin-sso`. Its required set is printed rather
than copied, since a list written into a document drifts against the ruleset
that decides it:

```
gh api repos/Flowfin/jellyfin-plugin-sso/rules/branches/main \
  --jq '.[] | select(.type=="required_status_checks")
        | .parameters.required_status_checks[].context'
```

On 2026-08-10 that printed thirteen contexts, and the table below carries one
row for each of them. Re-run the command before trusting the table. A row that
no longer has an entry behind it, or an entry with no row, is the drift this
document is most likely to develop, and finding it costs one command.

## The gap this rests on

The ruleset on the default branch of this board requires no status check at
all:

```
gh api repos/Flowfin/lab/rules/branches/main --jq '[.[].type]'
["deletion","non_fast_forward","pull_request","required_signatures"]

gh api repos/Flowfin/lab/rules/branches/main \
  --jq '.[] | select(.type=="required_status_checks")
        | .parameters.required_status_checks[].context'
(no output)
```

Every workflow in this tree runs, and none of them holds a merge. A red tick
and a green tick have the same effect on whether a change lands, so the checks
here are documentation until that changes. Parity is therefore mostly a
question of a required set that does not exist yet rather than of checks that
have not been written. Issue #26 is where the set is assembled.

Read the table with that in front of it. "Kept" below means kept as a rule this
board holds itself to, and it does not mean the tick holds a merge, because
today no tick does.

## The table

| Target context | Verdict here | Reason |
| --- | --- | --- |
| `build` | Kept, renamed | This repository builds a runner rather than a plugin assembly, and the entries are per platform, which `docs/decisions/0012-the-supported-platforms.md` fixes. |
| `ABI floor build` | Dropped | It exists so a plugin loads against the oldest server it claims to support, and nothing here loads into a host. |
| `Package (JPRM) / Build package` | Dropped | This repository ships no plugin, so there is no package to build. |
| `Package (JPRM) / Generate SBOM` | Kept in substance, moved out of the gate | A bill of materials is owed for anything downloadable, and this one is produced by the release build rather than by a check on a pull request. Issue #37 builds it and nothing in this tree produces one today. |
| `CodeQL` | Kept, retargeted | Static analysis of the runner's own source, in the language record `0001` chose rather than in C#. Two strings arrive here for it, the analysis job's `CodeQL (go)` and the code-scanning upload's `CodeQL`. Which of the two a required set can hold is open rather than decided below: the section on which contexts arrive separates the strings, and says of the upload name that the route which would decide it has not been walked. Issue #62 holds that walk. |
| `Analyze (csharp)` | Dropped as a name | The language-specific analysis job is replaced by the equivalent for this language rather than carried across under a name that describes nothing here. |
| `DCO sign-off` | Kept unchanged | Already in the tree, asserting the text at `DCO` on every non-merge commit. |
| `Deterministic PR-hygiene checks` | Kept, adapted | The class the other checks miss is the pull request itself, and one of the three refusals is this board's own invariant about a record moving with the code it describes. It is in the tree as the `pull request` job, judged in `internal/pullrequest/` and run from `.github/workflows/pull-request.yml`. |
| `Enforce greppable invariants` | Kept, different invariants | The invariants are properties of this repository's own tracked text, which is a different set from the target's, and they are in `internal/invariants/`. |
| `Reject Trojan Source Unicode` | Kept unchanged | Already in the tree, and the attack it refuses is a property of source rather than of a language. |
| `Audit workflows (zizmor)` | Kept unchanged | Already in the tree. The workflow YAML is the other executable thing here and it runs with write scopes. The job reports under this name and the code-scanning upload reports under `zizmor`, which is a different context and does not arrive on every pull request. |
| `prettier` | Kept, split in two, and both halves are in the tree | Records here are Markdown, and a whitespace diff on a record hides the sentence that changed. The runner's own source is held to `gofmt` by the `format` job and the prose half is the `prose format` job. Both are kept and both belong in the required set. Neither rewrites anything, which is where the split departs from the target: `prettier` formats and these two refuse, so a departure is a red tick here and a diff there. |
| `dependency-review` | Kept unchanged | Already in the tree, refusing a newly introduced dependency carrying a known advisory. |

## What this board adds that the target does not have

The record checks. The question and the answer are this repository's primary
artefact, so a change that leaves a record wrong is the failure this board most
needs to refuse, and the target has no equivalent because it has no records.
They belong in the required set here.

The headless and unelevated run. Record `0007` makes both halves a birth
requirement, and the job proves the default suite completes with no graphical
session and as an ordinary user. The target does not carry it because a plugin
assembly is not run by a contributor on their own machine in the same way.

Three more names this tree declares belong in this section, and until this
paragraph none of them appeared anywhere in this document, in a row or in a
sentence. They are declared here:

```
git grep -h -E '\{Name: "(test \(|vet"|required contexts")' origin/main \
  -- internal/contexts/contexts.go | sed 's/^[[:space:]]*//'
{Name: "test (linux/amd64)", Why: theSetIsEmpty, Until: "#26"},
{Name: "test (windows/amd64)", Why: theSetIsEmpty, Until: "#26"},
{Name: "test (darwin/arm64)", Why: theSetIsEmpty, Until: "#26"},
{Name: "vet", Why: theSetIsEmpty, Until: "#26"},
{Name: "required contexts", Why: theSetIsEmpty, Until: "#26"},
```

The leading indentation is stripped so that this document carries no tab, which
`prose-carries-a-tab` refuses in tracked Markdown. Nothing else about the five
lines is changed.

A name with no verdict here is the shape that costs issue #26 an answer rather
than a line of prose. The set it assembles is taken from this document, and a
declared name the document gives no verdict to can be read as kept or as
dropped with equal justice, so the two readings produce two different gates.
Each entry below therefore says kept or dropped and why, in the same terms the
table above uses.

The platform suite. Record `docs/decisions/0012-the-supported-platforms.md`
gives three of its six platforms a suite run as well as a build, and each of
the three reports under its own name, so the entries are `test (linux/amd64)`,
`test (windows/amd64)` and `test (darwin/arm64)`. The target requires a build
and requires nothing that runs a suite:

```
gh api repos/Flowfin/jellyfin-plugin-sso/rules/branches/main \
  --jq '.[] | select(.type=="required_status_checks")
        | .parameters.required_status_checks[].context' \
  | grep -E '^(test|vet)' ; echo "exit=$?"
exit=1
```

All three are kept and they belong in the required set. The differences record
`0012` picks those three platforms for, whether the filesystem folds case, what
separates a path and what a line ending arrives as, are the defects a runner
that reads a checkout actually has, and a build entry that compiled on a
platform says nothing about whether the runner reads that platform's tree
correctly. The target has no equivalent because the artefact it gates is loaded
by a server rather than run against a working copy.

`vet` is kept, and it belongs in the set at a smaller cost than the suite: one
job over the whole module rather than three entries on three machines, since
what it reads is the source rather than the filesystem underneath it. It has no
counterpart in the command above for the same reason the language-specific
analysis row does not carry across, which is that the target is written in
another language and holds its own source to that language's tools.

`required contexts` is kept, and it is the entry a reader is likeliest to find
absent here rather than wrong, because it is younger than the walk this
document is built on. The table above reads a ruleset as it answered on
2026-08-10 and this check landed two days later:

```
git log -1 --format='%h %ad %s' --date=short --diff-filter=A \
  -- .github/workflows/contexts.yml
2b954f5 2026-08-12 Refuse a required context and a check name that disagree (#71)
```

What it compares is the required set against the check names this tree
declares, so it is the one entry whose subject is the gate rather than the
tree. Requiring it is what refuses a set edited into disagreement with the
workflows, in either direction, and that is the direction no document can cover
because nothing reads a document. The target has no equivalent and the absence
is not an argument against carrying one here.

The supply-chain self-audit stays outside the required set, for the same reason
it is outside the target's. It publishes from the default branch and cannot
gate a pull request, so requiring it would require a context that never
arrives on the thing being gated.

Five more names this tree declares carried no verdict here until this
paragraph, and they are a different case from the five above. Those are waiting
for a required set to join. These can never join one.

The sweep, read at `9208ceb599a294328bde3d8b660c65a5fd3c5fb5` rather than at
`origin/main`, because the paragraphs below are what change its output and a
paste that stops reproducing the moment it lands is the drift this document
warns about at the top:

```
git show 9208ceb599a294328bde3d8b660c65a5fd3c5fb5:internal/contexts/contexts.go \
  | grep -oE 'Name: +"[^"]+"' | sed 's/Name: *"//; s/"$//' | sort -u \
  | while read -r n; do
      git grep -q -F "$n" 9208ceb599a294328bde3d8b660c65a5fd3c5fb5 \
        -- docs/quality-parity.md || echo "no literal: $n"
    done | grep -v '^no literal: build ('
no literal: smoke (darwin/arm64)
no literal: smoke (linux/amd64)
no literal: smoke (windows/amd64)
no literal: verify the published artefacts
```

The last stage is why this paste is four lines rather than ten, and it is not
tidying. The sweep also returns the six per-platform `build` entries, whose
verdict is the `build` row of the table above and reaches them through
`docs/decisions/0012-the-supported-platforms.md`. Those six are the one place a
literal is deliberately absent here, so pasting the unfiltered output would
write the six strings into this document and end the absence in the act of
describing it. The stage drops them by the row name the table already carries
rather than by naming any of the six.

That the paste needs the stage at all is the shape to take from this section: a
sweep for names this document does not carry cannot have its full output pasted
into this document, because the paste is what makes the answer wrong. What is
left after the stage is the gap, and it is four names.

A fifth name is one too and this sweep cannot show it at all. `release`
matches as a word inside the reason on the SBOM row rather than as a
verdict of its own, which is the false pass a fixed-string comparison gives and
the reason the tree holds the list that decides this in
`internal/contexts/contexts.go` instead.

All five are dropped from the required set, permanently rather than until #26.
What their jobs read is a tag or a published release, and a pull request has
neither, so these contexts arrive on nothing a merge is waiting for and
requiring one would hold every merge open for a tick that is not coming. That
is what `required-context-nothing-reports` exists to refuse, so the two
directions of the comparison would contradict each other. Four of the five
carry that reason as a shared constant:

```
git show origin/main:internal/contexts/contexts.go \
  | awk '/^var Absences/,/^}$/' | tr '\n' ' ' \
  | grep -oE '\{[^{}]*neverOnAPullRequest[^{}]*\}' | grep -oE 'Name: +"[^"]+"'
Name:  "verify the published artefacts"
Name: "smoke (linux/amd64)"
Name: "smoke (windows/amd64)"
Name: "smoke (darwin/arm64)"
```

`release` carries a reason of its own in the same list because its job runs on a
tag push and on nothing else, which is narrower than what the shared sentence
says. The workflow files agree with the list rather than being described by it:
`.github/workflows/release.yml` triggers on `push` with a tag pattern and
declares no pull-request trigger, and `.github/workflows/smoke.yml` says at its
own trigger block that the names its jobs report under arrive on no pull
request at all and that `internal/contexts` carries them as permanent
deliberate absences.

Dropped from the gate is not dropped as a rule, and the smoke jobs are the case
where the difference matters most. They are the only thing in this repository
that reads what an operator downloads rather than the source it was built from,
which is the distance every other check here is blind to. They hold that rule
on a published release and on a schedule instead of on a pull request, so what
is given up by leaving them out of the required set is the moment the failure is
caught rather than whether it is caught.

## What each side reports today

Neither list is written here. Both are printed, from a completed run rather
than from a workflow file, because a job name and a check-run name are not
always the same string:

```
gh api repos/Flowfin/lab/commits/$(git rev-parse origin/main)/check-runs \
  --jq '.check_runs[].name' | sort -u
```

That is the command issue #26 assembles the required set from, and it is the
one to run before quoting any context name back at this document.

## Which contexts arrive, and on which pull requests

A required context that does not arrive is not a red tick. It is a pull request
that cannot merge and says nothing about why, so the first response is to wait
and the second is to make the gate smaller. The names above are therefore
separated by what produces them and by the commit they were read from before
any of them goes into a required set. Issue #62 holds this walk.

### The command above reads a commit no pull request produced

`git rev-parse origin/main` resolves a commit on the default branch, which is
reached by a push. Four workflows in this tree carry no push trigger:

```
git grep -L 'push:' origin/main -- .github/workflows/
origin/main:.github/workflows/dco.yml
origin/main:.github/workflows/dependency-review.yml
origin/main:.github/workflows/pull-request.yml
origin/main:.github/workflows/smoke.yml
```

Three of those four run on a pull request and the fourth runs on neither kind of
commit, and only the first three are what this section is about.
`.github/workflows/smoke.yml` triggers on a published release, on a schedule and
on a manual dispatch, so it reports on no pull request and on no push, and it is
in this list for a reason the list is not about. That is the thing to read the
grep for: it gains a member every time this board adds a workflow that runs on
neither, and what would matter here is a file that does run on a pull request
appearing in it.

So the two kinds of commit report different sets, and the difference runs in
both directions. Between the head of the default branch and the head of a pull
request that landed on it, on 2026-08-16:

```
gh api repos/Flowfin/lab/commits/82c245f/check-runs?per_page=100 \
  --jq '.check_runs[].name' | sort -u > default-branch.txt
gh api repos/Flowfin/lab/commits/7374b6b/check-runs?per_page=100 \
  --jq '.check_runs[].name' | sort -u > pull-request.txt

comm -23 default-branch.txt pull-request.txt
Scorecard analysis

comm -13 default-branch.txt pull-request.txt
CodeQL
DCO sign-off
dependency-review
pull request
zizmor
```

Three of those five are the pull-request-only workflows above and two are the
subject of the next section. `Scorecard analysis` is the opposite case and is
already written down as a permanent absence in `internal/contexts/contexts.go`,
for the reason the supply-chain paragraph above gives.

A set assembled from a default-branch commit therefore leaves out three
contexts that every pull request reports, and it offers one that no pull
request can report. Which commit that command is run against decides what the
gate gets.

### One name is reported twice on the same commit

Every command in this document that lists check names deduplicates, and one of
the names they print once was produced by two runs. On `bd861cf`, the head of
pull request #145:

```
gh api repos/Flowfin/lab/commits/bd861cf/check-runs?per_page=100 \
  --jq '.check_runs[].name' | sort | uniq -d
Reject Trojan Source Unicode

gh api repos/Flowfin/lab/commits/bd861cf/check-runs?per_page=100 \
  --jq '.check_runs[] | select(.name=="Reject Trojan Source Unicode")
        | "suite \(.check_suite.id), \(.conclusion)"'
suite 86798670354, success
suite 86798547856, success
```

The cause is one trigger. Every push trigger in these files that names a branch
names the default branch except one, and one names no branch at all:

```
for f in $(git ls-tree --name-only origin/main .github/workflows/); do
  s=$(git show "origin/main:$f" | sed -n '/^  push:/{n;s/^ *//;p;}')
  [ -n "$s" ] && printf '%s %s\n' "$f" "$s"
done
.github/workflows/build.yml branches: [main]
.github/workflows/codeql.yml branches: [main]
.github/workflows/contexts.yml branches: [main]
.github/workflows/headless.yml branches: [main]
.github/workflows/invariants.yml branches: [main]
.github/workflows/prose.yml branches: [main]
.github/workflows/records.yml branches: [main]
.github/workflows/release.yml tags:
.github/workflows/scorecard.yml branches: [main]
.github/workflows/unicode-guard.yml branches: ["**"]
.github/workflows/zizmor.yml branches: [ main ]
```

The `tags:` line is the second kind and it doubles nothing. A tag push is not a
branch push, so `.github/workflows/release.yml` starts on neither a pull request
nor the branch one is opened from, and the file says that at its own trigger.
What produces the repeated name is the entry above carrying every branch.

A branch pushed for a pull request therefore starts that one workflow twice,
and both runs report under its job name. The file gives the reason its trigger
is wide, and the reason is about which branches the guard covers rather than
about the gate:

```
git show origin/main:.github/workflows/unicode-guard.yml | sed -n '3,11p'
on:
  # Every branch and every PR: the guard is a cheap read-only scan, so there is no
  # reason to narrow it to main, and it covers whatever branches this repository
  # grows later without being edited again.
  push:
    branches: ["**"]
  pull_request:
    branches: ["**"]
```

The two runs do not read the same tree, and that is what makes this more than a
repeated line. Each checkout says what it took, and the answers differ:

```
gh run view 32017539721 --log \
  | sed -n 's/.*\(git checkout --progress --force .*\)/\1/p'
git checkout --progress --force -B parity/one-job-asks-the-platform-and-the-fork-half-is-unwalked refs/remotes/origin/parity/one-job-asks-the-platform-and-the-fork-half-is-unwalked

gh run view 32017585271 --log \
  | sed -n 's/.*\(git checkout --progress --force .*\)/\1/p'
git checkout --progress --force refs/remotes/pull/145/merge
```

The first read the branch and the second read the branch merged into the base.
A tree that carries none of these characters on the branch and carries one once
merged separates the two runs, which is the case this particular guard exists
for one merge earlier.

`Reject Trojan Source Unicode` is a row the table above keeps, so it is a
candidate for the required set. What a merge does when two check runs answer to
one name is platform behaviour that nothing in this tree states, and the set
here is empty, so it cannot be measured on this board today either. Issue #26
assembles the set and its first clause meets this, because the command it
assembles from prints one line for the two.

### Two of the names come from an upload rather than from a job

The names on that same pull request head that no job in this tree produced:

```
gh api repos/Flowfin/lab/commits/7374b6b/check-runs?per_page=100 \
  --jq '.check_runs[] | select(.app.slug != "github-actions")
        | "\(.name) is created by \(.app.slug)"' | sort -u
CodeQL is created by github-advanced-security
zizmor is created by github-advanced-security
```

The two rows in the table above carry the target's strings. Here each of them
is two strings rather than one: the analysis job reports `CodeQL (go)` and the
code-scanning upload reports `CodeQL`, the audit job reports
`Audit workflows (zizmor)` and the upload reports `zizmor`. Two of the four are
written in a workflow file and two are written in no file in this tree, which
is how `internal/contexts/contexts.go` holds them.

### The upload is conditional and the step that fails on findings is not

```
git show origin/main:.github/workflows/zizmor.yml | sed -n '76,84p'
      - name: Upload SARIF
        # Only upload where the GITHUB_TOKEN can write security events: pushes to main
        # and same-repo human PRs. Fork and Dependabot pull requests run with a
        # read-only token, so the upload is skipped there - the gate step below still
        # runs and blocks on findings. continue-on-error keeps the security gate
        # independent of the upload: a transient code-scanning upload failure must not
        # skip the "Fail on actionable findings" step below.
        if: (github.event_name == 'push' && github.ref == 'refs/heads/main') || (github.event.pull_request.head.repo.full_name == github.repository && github.event.pull_request.user.login != 'dependabot[bot]')
        continue-on-error: true
```

That condition has a pull request on this board today that it excludes, so the
outcome is measured rather than read off the file. On the head of #135, whose
author the second arm names:

```
gh api repos/Flowfin/lab/actions/runs/31929377858/jobs \
  --jq '.jobs[].steps[] | select(.number >= 4 and .number <= 6)
        | "step \(.number), \(.name): \(.conclusion)"'
step 4, Audit workflows (SARIF for code scanning): success
step 5, Upload SARIF: skipped
step 6, Fail on actionable findings: failure
```

The upload did not run, and only one of the two names reached that commit:

```
gh api repos/Flowfin/lab/commits/d18d040/check-runs?per_page=100 \
  --jq '.check_runs[] | select(.name | test("zizmor"))
        | "\(.name): \(.conclusion)"' | sort -u
Audit workflows (zizmor): failure
```

So the job arrived and reported what the audit found while the upload-derived
context did not arrive at all, which is the separation the comment in the
workflow file argues for. It decides one name: `zizmor` cannot be in the
required set, because requiring it would hold open every pull request the
condition excludes, with nothing on the pull request saying why. That absence is
permanent rather than one issue #26 retires, and the job name beside it is
unaffected.

### None of these workflows narrows itself to some pull requests

A context that arrives on some pull requests and not on others is what this
section is written against, and three of the ways to build one are readable in
these files rather than walked. The readings below were made at
`9208ceb599a294328bde3d8b660c65a5fd3c5fb5` and cover every workflow file in the
tree at that commit. They replace a reading made at `1fc6961`, two of whose five
pastes had stopped reproducing by the time this one was taken: five files carry
no branch filter that reads every branch where three did, and two carry a type
filter where one did. Both differences are this board gaining the release and
smoke workflows, and neither of those runs on a pull request. The other three,
the path filters, the branch filter under `dependency-review.yml` and the `if:`
keys, reproduce unchanged at this commit.

A path filter is the ordinary way it happens, and no workflow here carries one:

```
git grep -n 'paths:\|paths-ignore:' origin/main -- .github/workflows/ ; echo "exit=$?"
exit=1
```

A trigger narrowed to some branches is the same failure by a second route, and
five of these files carry no branch filter that reads every branch:

```
git grep -L 'branches: \[ *"\*\*" *\]' origin/main -- .github/workflows/
origin/main:.github/workflows/dco.yml
origin/main:.github/workflows/dependency-review.yml
origin/main:.github/workflows/release.yml
origin/main:.github/workflows/scorecard.yml
origin/main:.github/workflows/smoke.yml
```

Three of the five declare no pull-request trigger at all and are outside the
required set already, permanently rather than until #26. The supply-chain
self-audit publishes from the default branch; the release and smoke workflows
read a tag or a published release, which is the verdict written for them above
under what this board adds. A file that runs on no pull request cannot narrow
one, so the membership of this command grows every time this board gains a
workflow of that shape, and the growth answers nothing the section asks. What
it is worth reading for is a file that does run on a pull request appearing in
it.

Neither of the two that do narrows anything. `dependency-review.yml` writes no
branch filter under any of its triggers, which is every branch:

```
git grep -n 'branches:' origin/main -- .github/workflows/dependency-review.yml ; echo "exit=$?"
exit=1
```

`dco.yml` carries the only type filter on a pull-request trigger in these
files. One other carries the key, on a trigger that is not a pull request at
all:

```
git grep -n 'types:' origin/main -- .github/workflows/
origin/main:.github/workflows/dco.yml:13:    types: [opened, synchronize, reopened]
origin/main:.github/workflows/smoke.yml:41:    types: [published]
```

The claim about the three types it names is that they are the three the
platform runs a pull-request workflow for when a file names none, so writing
them out removes no pull request. That is the platform's documented default
rather than something this tree says, and nothing here measures it. What sits
beside it is that `DCO sign-off` is among the contexts the pull-request head
compared above reported, which shows the workflow runs on an ordinary pull
request and shows nothing about one that arrives another way.

A job skipped by a condition is the third, and two of these files carry one at
all:

```
git grep -c '^\s*if:' origin/main -- .github/workflows/
origin/main:.github/workflows/scorecard.yml:1
origin/main:.github/workflows/zizmor.yml:1
```

Neither reaches a name the table above keeps. The first is a job-level
condition on the supply-chain self-audit, which the paragraph above already
places outside the required set for a reason of its own. The second is the
upload condition quoted earlier in this section, indented under a step rather
than a job, and the step that fails on findings sits after it carrying no
condition.

What these commands do not reach. They read the path filters, the branch and
type filters and the `if:` keys, which is what those three are written with in
these files today, and a fourth route written some other way is not something a
reader can take from them. None of them says anything about
what a job does once it has started, so a context that arrives having done
nothing is a different question and is not answered here.

### One job can start and still be unable to do its work

The readings above stop at the moment a job begins. One job here has a second
failure after that moment, because half of what it compares is a live setting
on the platform rather than a file in the checkout, so it has to ask. No other
workflow in this tree asks anything:

```
git grep -n 'gh api' origin/main -- .github/workflows/
origin/main:.github/workflows/contexts.yml:93:          if ! gh api "repos/${REPOSITORY}/rules/branches/${DEFAULT_BRANCH}" \
```

That reads the run blocks these files carry, so it says which workflow asks the
platform a question in a script this repository writes. What an action reaches
once it has started is not readable from here and the command says nothing
about it.

The job reports as `required contexts`, which the section above keeps and puts
in the required set. What it does when the answer does not arrive is written
into the step rather than left to the shell:

```
git show origin/main:.github/workflows/contexts.yml | sed -n '92,100p'
          set -euo pipefail
          if ! gh api "repos/${REPOSITORY}/rules/branches/${DEFAULT_BRANCH}" \
                --jq '.[] | select(.type=="required_status_checks")
                      | .parameters.required_status_checks[].context' > required.txt; then
            echo "::error::The ruleset on ${DEFAULT_BRANCH} could not be read, so this run could not judge whether the gate and the tree agree. That is not the same as them agreeing."
            exit 1
          fi
          echo "the ruleset on ${DEFAULT_BRANCH} requires $(wc -l < required.txt) context(s):"
          cat required.txt
```

So this route ends in a red context carrying its reason rather than in a
context that never arrives, which is the better of the two failures and is
still a pull request held open by something other than the change on it. An
empty answer is not that failure. The count is printed and the run carries on,
which is the state this board is in today.

One of the two things a fork run brings is measurable here already. A
Dependabot pull request runs with a read-only token, which is what the upload
condition quoted earlier in this section names it for, and #135 is one. On its
head the job ran and the fetch answered:

```
gh api repos/Flowfin/lab/commits/d18d040/check-runs?per_page=100 \
  --jq '.check_runs[] | select(.name=="required contexts")
        | "\(.name): \(.conclusion)"'
required contexts: success

gh run view 31929377870 --log \
  | sed -n 's/.*\(the ruleset on main requires .*\)/\1/p'
the ruleset on main requires 0 context(s):
```

A read-only token reads this board's ruleset, so that is not what would stop a
run from a fork. What is left is the other thing, which is that the token is
issued against a different repository. `github.repository` names this board on
a pull request from a fork, so the fetch asks about this board whichever side
the branch sits on, and whether a token issued that way may read this board's
ruleset is platform behaviour that nothing in this tree states and nothing
above measures.

`required contexts` therefore belongs in the fork clause of #62, and it is
there for a different reason from the two names already in it. Those two are
created by an upload and turn on a write scope. This one is a job that always
starts, and what is open is whether the answer it needs arrives.

### What this section does not settle

No pull request from a fork has been opened here. The condition above has two
arms and only the second was walked: the branch measured above is in this
repository, and a read-only token is what the two arms have in common rather
than what makes them one route. `CodeQL` did arrive on that pull request, so
nothing here says whether it arrives from a fork, and the fork clause of #62 is
open for it. It is open for `required contexts` as well, for the reason the
subsection above gives rather than for this one.

The required set is empty, so nothing above is a report of a required context
that failed to arrive. Every sentence here is about which names a set could
hold, and none of them is about a gate that bit.

## The rest of the ruleset

A required set is one rule out of the four that stand behind a merge at the
target. The rest decides who may push, what a merge does to the commits it
lands, and whether anybody has to read a change, and none of that is visible in
a table of contexts. The table above walks the required set and stops there, so
this is the other half of the same walk.

Print both rather than trusting what follows:

```
gh api repos/Flowfin/lab/rules/branches/main --jq '.[].type'
gh api repos/Flowfin/jellyfin-plugin-sso/rules/branches/main --jq '.[].type'
gh api repos/Flowfin/lab/rules/branches/main \
  --jq '.[] | select(.type=="pull_request") | .parameters'
```

### The rule types

On 2026-08-10 the first two commands printed three types here and four at the
target. Re-read on 2026-08-28 they print four here and five there, and the entry
that arrived on both sides between those two readings is the same one:

```
gh api repos/Flowfin/lab/rules/branches/main --jq '.[].type'
deletion
non_fast_forward
pull_request
required_signatures
gh api repos/Flowfin/jellyfin-plugin-sso/rules/branches/main --jq '.[].type'
deletion
non_fast_forward
required_status_checks
pull_request
required_signatures
```

| Rule type | Here | Target | Verdict |
| --- | --- | --- | --- |
| `deletion` | Present | Present | Kept. The branch this board's history lives on cannot be deleted on either side. |
| `non_fast_forward` | Present | Present | Kept, and it carries more weight here than a row suggests. It is the only thing refusing a rewrite of the default branch, and no check in this tree can see one: the runner reads a checkout, so a history replaced under it reads as the tree it was given. |
| `pull_request` | Present | Present | Kept. Its parameters are the table below. |
| `required_status_checks` | Absent | Present | The deviation the gap section above is already about, and the largest one in either walk. Issue #26 assembles the set. |
| `required_signatures` | Present | Present | Kept, and it is the row that moved. It was absent from both boards at every earlier reading of this walk and is configured on both now, so parity settles nothing in either direction. It arrived ahead of the condition `docs/decisions/0023-signed-commits-on-the-default-branch.md` makes it effective on, which is the subject of its own subsection below. |

Enforcement and the bypass list are properties of the ruleset rather than rule
types, so neither command above prints them. Both boards answer the same way,
and this is the answer for this one:

```
gh api "repos/Flowfin/lab/rulesets/$(gh api repos/Flowfin/lab/rulesets \
  --jq '.[] | select(.name=="gate") | .id')" \
  --jq '{enforcement, bypass: .bypass_actors}'
{"bypass":[],"enforcement":"active"}
```

An empty bypass list is what makes the four rules above mean anything, because
an actor on that list is an actor none of them applies to. Issue #26 already
requires it to stay empty, so nothing here adds a second rule about it.

### The pull-request rule, parameter by parameter

Every parameter the third command prints, with the target's value beside it. On
2026-08-10 all but the first held the same value on both boards, and that is
still the reading on 2026-08-21. A setting left at its default and a setting
chosen deliberately look identical afterwards, which is why every row carries a
reason and not only a verdict.

The table listed eight rows and the command prints nine names. Which of the two
that is, a parameter the platform started printing after the table was written
or a row the walk missed on the day, is not readable from here, and the last row
below is the one that was absent either way.

| Parameter | Here | Target | Verdict |
| --- | --- | --- | --- |
| `allowed_merge_methods` | `["merge","squash","rebase"]` | `["merge"]` | Change owed. This is the one deviation in this walk worth closing rather than reasoning away, and the reason is below. |
| `required_approving_review_count` | `0` | `0` | Kept. A count above zero on a board with one maintainer refuses every merge, and a rule nobody can satisfy is switched off in a hurry rather than met. |
| `dismiss_stale_reviews_on_push` | `false` | `false` | Kept. It only bites where a review is required, and none is required at a count of zero. |
| `require_last_push_approval` | `false` | `false` | Kept, for the reason in the row above. |
| `required_review_thread_resolution` | `false` | `false` | Kept, for the reason two rows above. |
| `require_code_owner_review` | `false` | `false` | Kept. There is no `CODEOWNERS` file in this tree, so requiring a code-owner review here would require an approval nothing can name. |
| `dismissal_restriction` | `{"allowed_actors":[],"enabled":false}` | The same | Kept. It restricts who may dismiss a review, and there is no required review to dismiss. |
| `required_reviewers` | `[]` | `[]` | Kept, for the reason in the row above. |
| `require_extra_approval_for_unattributed_changes` | `true` | `true` | Kept, and it is the row to read carefully. The four rows above turn on a review being required, and this one is written to ask for an approval where a change carries commits the pull request's author is not credited with, so it is the only parameter here that could ask for one at a count of zero. Whether it does is a statement about the platform and no command in this walk answers it, so nothing is claimed. It has never been observed to hold a merge on this board. |

### Why the merge methods are not a style preference

`docs/decisions/0004-what-happens-to-the-code.md` lets an answered experiment's
code be removed and requires the record to gain one line naming the commit that
removed it, carrying the full hash. That line is written in the same change as
the removal, so it names a commit that a squash replaces with a different one
and that a rebase rewrites. The record then points at nothing on the default
branch while still reading as complete, which is worse than a record that is
obviously wrong, because nothing about it looks unfinished.

`docs/decisions/0005-how-a-result-leaves.md` has the same shape one step further
out. A promotion names a commit range in this repository, and a range whose ends
were rewritten by the merge that landed them is a pointer the receiving board
cannot follow.

Restricting the methods to `["merge"]` is what keeps a named commit resolvable.
The restriction is not in place. The third command above still prints all three
methods on this board, so a squash or a rebase merge here can break a record of
either kind, and the only thing standing against it today is whoever picks the
button. Issue #55 holds the change.

### The rule both boards carry now, and the condition it arrived ahead of

A verified signature on every commit was required by neither ruleset at every
earlier reading of this walk. Both require one now. Re-read on 2026-08-28:

```
gh api repos/Flowfin/lab/rules/branches/main --jq '.[].type'
deletion
non_fast_forward
pull_request
required_signatures
gh api repos/Flowfin/jellyfin-plugin-sso/rules/branches/main --jq '.[].type'
deletion
non_fast_forward
required_status_checks
pull_request
required_signatures
```

WHAT STOOD HERE SAID NEITHER BOARD CARRIED IT, over a paste read on 2026-08-26,
under a heading that read as a settled absence rather than as a reading with a
date on it. The subsection below names the failure a document describing a live
setting always has, and this is an instance of it rather than an illustration:
the setting was changed on both boards by hand, nothing in this tree or at the
target announced it, and what found it was re-running the command before quoting
the row back.

This walk still does not decide whether to require one, and parity settles
nothing in either direction, because both sides answer the same way. The
question was an entry on issue #46, and that entry is answered:

```
gh issue view 46 --repo Flowfin/lab --json state,closedAt --jq '"\(.state) \(.closedAt)"'
CLOSED 2026-08-27T08:46:01Z
```

That paste carried `2026-08-24T19:11:13Z` until this reading, and the earlier
timestamp was correct when it was written rather than wrong. The issue was
reopened and closed again in between, so a document quoting a closing moment
carries a value that moves whenever an issue is reopened, which the timeline is
the authority for:

```
gh api repos/Flowfin/lab/issues/46/timeline --paginate \
  --jq '.[] | select(.event=="closed" or .event=="reopened")
        | "\(.event) \(.created_at)"'
closed 2026-08-24T19:11:13Z
reopened 2026-08-27T07:19:18Z
closed 2026-08-27T08:46:01Z
```

The answer is `docs/decisions/0023-signed-commits-on-the-default-branch.md`: a
commit on the default branch has to carry a verified signature, effective as the
account keys operations#1609 sets up for the working accounts land, and not
before. That record names this document among the things it applies to, so the
walk points at the record rather than at the issue that collected the question.

THE SETTING ARRIVED AHEAD OF THAT CONDITION AND THE CONDITION HAS SINCE BEEN
MET, which is two readings rather than one and only the first of them is a fact
about the ordering. Record 0023 gives its reason for that ordering in one
sentence: a rule that refuses every merge before anybody holds a key is a rule
that gets turned off rather than followed. The issue holding the keys is closed:

```
gh issue view 1609 --repo iderex/operations --json state --jq '.state'
CLOSED
```

WHAT STOOD HERE PASTED `OPEN` UNDER THAT COMMAND, and said a merge on this
board refuses an unsigned commit while the custody story the record conditions
that refusal on is unfinished. Both halves were correct when they were written
and the first stopped reproducing on 2026-08-28. What found it was running the
command before quoting the row back, which is the only way this class is found:
a claim about another artefact reads the same whether or not the artefact still
says it. A state moves in both directions and an issue can be reopened, so the
paste above is a reading with a date on it rather than the answer.

So the gap this row was about is closed rather than open, and what is left of it
is smaller and worth stating exactly. The requirement is configured on both
boards and the condition record 0023 makes it effective on has been met. That it
has not cost a landing here is a separate reading and it still holds, because
the commits reaching the default branch carry a signature the platform verifies.
Read at the three most recent non-merge commits:

```
for c in 45bfe62 2edacce 43b4fae; do
  gh api repos/Flowfin/lab/commits/$c --jq '.commit.verification | "\(.verified) \(.reason)"'
done
true valid
true valid
true valid
```

Three commits are three commits and not a property of every account that may
push here. This board takes experiments from anybody, an unsigned history
refuses the merge rather than the commit, and the repair is rebuilding the
branch at the end of the work, which record 0023 already writes down as the
moment it is most expensive. The keys landing does not make that smaller: a key
held by the account that cuts this board's changes says nothing about a
contributor who holds none, and record 0023 says in as many words that it
decides a signature is required and does not decide what happens to somebody
with no key. Whether the setting should stand at all is still not a question
this walk takes, and it is not one this document can answer: the ruleset is not
in this tree, and `allowed_merge_methods` remains the only ruleset edit this
document asks for.

A rule that is configured is not a rule this tree refuses, and that is the half
of this section a reader is most likely to collapse in the other direction now.
The record says of itself that nothing here refuses an unsigned commit, and that
is unchanged: the requirement is a setting on the branch protection, no check in
this repository reads a signature, and a green run says nothing about one.

### What this walk cannot do

It reads the API as it answered on one day. A rule or a parameter changed by
hand afterwards is not caught here, and the tables go wrong quietly rather than
loudly, which is the failure a document describing a live setting always has.
Re-run the commands before quoting any row back.

It also says nothing about whether a change was read. A ruleset carrying every
rule walked above still lands a change nobody looked at, because the approving
review count is zero and that is a deliberate row rather than an oversight.

## What this document is not

It is not a claim that this board is as safe as the target. It records which of
the target's entries survive here and why, and a row saying kept is a statement
about a rule rather than about a mechanism: several of the rows above name an
issue that has not been built, and those say so.

It is also not a list of the checks in this tree. Two of the rows are already
about work that does not exist yet, so a reader who wants to know what actually
ran should read what a run printed. The run says what it examined.
