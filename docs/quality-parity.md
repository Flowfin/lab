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

## What this document is not

It is not a claim that this board is as safe as the target. It records which of
the target's entries survive here and why, and a row saying kept is a statement
about a rule rather than about a mechanism: several of the rows above name an
issue that has not been built, and those say so.

It is also not a list of the checks in this tree. Two of the rows are already
about work that does not exist yet, so a reader who wants to know what actually
ran should read what a run printed. The run says what it examined.
