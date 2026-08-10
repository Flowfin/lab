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
["deletion","non_fast_forward","pull_request"]

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
| `CodeQL` | Kept, retargeted | Static analysis of the runner's own source, in the language record `0001` chose rather than in C#. |
| `Analyze (csharp)` | Dropped as a name | The language-specific analysis job is replaced by the equivalent for this language rather than carried across under a name that describes nothing here. |
| `DCO sign-off` | Kept unchanged | Already in the tree, asserting the text at `DCO` on every non-merge commit. |
| `Deterministic PR-hygiene checks` | Kept, adapted | The class the other checks miss is the pull request itself, and one of the three refusals is this board's own invariant about a record moving with the code it describes. Issue #24 builds it and nothing in this tree reasons about a pull request today. |
| `Enforce greppable invariants` | Kept, different invariants | The invariants are properties of this repository's own tracked text, which is a different set from the target's, and they are in `internal/invariants/`. |
| `Reject Trojan Source Unicode` | Kept unchanged | Already in the tree, and the attack it refuses is a property of source rather than of a language. |
| `Audit workflows (zizmor)` | Kept unchanged | Already in the tree. The workflow YAML is the other executable thing here and it runs with write scopes. |
| `prettier` | Kept, split in two | Records here are Markdown, and a whitespace diff on a record hides the sentence that changed. The runner's own source is held to `gofmt` by a job already in the tree; the prose half is issue #50 and nothing in this tree formats Markdown today. |
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

The supply-chain self-audit stays outside the required set, for the same reason
it is outside the target's. It publishes from the default branch and cannot
gate a pull request, so requiring it would require a context that never
arrives on the thing being gated.

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
target.

| Rule type | Here | Target | Verdict |
| --- | --- | --- | --- |
| `deletion` | Present | Present | Kept. The branch this board's history lives on cannot be deleted on either side. |
| `non_fast_forward` | Present | Present | Kept, and it carries more weight here than a row suggests. It is the only thing refusing a rewrite of the default branch, and no check in this tree can see one: the runner reads a checkout, so a history replaced under it reads as the tree it was given. |
| `pull_request` | Present | Present | Kept. Its parameters are the table below. |
| `required_status_checks` | Absent | Present | The deviation the gap section above is already about, and the largest one in either walk. Issue #26 assembles the set. |

Enforcement and the bypass list are properties of the ruleset rather than rule
types, so neither command above prints them. Both boards answer the same way,
and this is the answer for this one:

```
gh api "repos/Flowfin/lab/rulesets/$(gh api repos/Flowfin/lab/rulesets \
  --jq '.[] | select(.name=="gate") | .id')" \
  --jq '{enforcement, bypass: .bypass_actors}'
{"bypass":[],"enforcement":"active"}
```

An empty bypass list is what makes the three rules above mean anything, because
an actor on that list is an actor none of them applies to. Issue #26 already
requires it to stay empty, so nothing here adds a second rule about it.

### The pull-request rule, parameter by parameter

Every parameter the third command prints, with the target's value beside it. On
2026-08-10 all but the first held the same value on both boards. A setting left
at its default and a setting chosen deliberately look identical afterwards,
which is why every row carries a reason and not only a verdict.

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

### The rule neither board carries

A verified signature on every commit is required by neither ruleset. This walk
does not add it and does not argue against it. Whether to require one is a
question about key custody rather than an engineering judgement, and it is an
entry on issue #46, which is open. The walk points there rather than deciding
it, and parity settles nothing in either direction, because the target does not
require one either.

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
