# Supply chain

The supply-chain self-audit runs in this repository and publishes a score. A
score that is generated and never read is a badge, and a badge is not a control.
This document is the triage. Every check the audit reports has an outcome here,
so the next person to look at the score can tell a check that was considered from
one nobody has looked at.

The score itself moves for reasons outside this repository and is not worth
chasing. What is worth keeping is the reasoning below.

## The run this triage reads

```
gh api "repos/Flowfin/lab/actions/workflows/scorecard.yml/runs?per_page=5" \
  --jq '.workflow_runs[] | "\(.id) \(.event) \(.conclusion) \(.created_at)"'
31224359958 push success 2026-08-07T22:35:39Z
```

The full result is a JSON line in that run's log. This is the command that pulls
it out, and every score below comes from it:

```
gh run view 31224359958 --repo Flowfin/lab --log \
  | sed -n 's/.*\({"date":.*\)/\1/p'
```

It reports version `v5.5.0`, commit `b23e20b681d627bfc55cdf9aec820e765bff9423`,
an aggregate of `4.7`, and eighteen checks. A score of `-1` means the check did
not apply and produced no result.

The artifact the workflow uploads carries only the findings, not the eighteen
scores:

```
gh run download 31224359958 --repo Flowfin/lab
```

That file names one check. So the command above, and not the artifact, is what a
later reader should use.

## Where the audit sits

It is outside the required set and stays there. It runs on the default branch and
on its own cadence because publishing a score is only valid from there, so it
cannot report on a pull request and cannot gate one. That is a property of the
check rather than a judgement about its value.

## The eighteen checks

### Binary-Artifacts, 10

`no binaries found in the repo`

Accepted. The tree holds text and nothing else, and the layout decision gives
binaries no place to live. What would move this is a checked-in build output,
which is what `.gitignore` exists to keep out.

### Branch-Protection, 0

`branch protection not enabled on development/release branches`, with
`Warn: branch protection not enabled for branch 'main'`

Accepted for now, and this one disagrees with the tree. There is an active
ruleset on `main`:

```
gh api repos/Flowfin/lab/rulesets --jq '.[] | {id, name}'
{"id":20573693,"name":"gate"}
gh api repos/Flowfin/lab/rulesets/20573693 \
  --jq '{enforcement, bypass: .bypass_actors, rules: [.rules[].type]}'
{"enforcement":"active","bypass":[],"rules":["deletion","non_fast_forward","pull_request"]}
```

Why the audit does not see it is not established here. Two candidates, and
neither is confirmed. The run's token carried `Contents: read`,
`Metadata: read` and `SecurityEvents: write`, which is printed in the same log,
and reading protection settings through the API needs more than that. The other
is that this check reads classic branch protection rather than a ruleset, which
would make the score correct about what it looked at and wrong about what stands
behind a merge.

Either way the answer is the same: re-score after issue #26 lands the required
set, and compare then. Arguing about the number before that would be arguing
about a check whose input is about to change.

### CI-Tests, -1

`no pull request found`

Did not apply. The run was made against a repository whose default branch had one
commit and no merged pull request. It will produce a real result once there is
merge history, and issue #13 is what gives it tests to see.

### CII-Best-Practices, 0

`no effort to earn an OpenSSF best practices badge detected`

Accepted, and not planned. The badge is a separate self-certification programme
with its own questionnaire. Pursuing it would mean maintaining a second account of
this repository's practices next to the one in `docs/`, and a second account
drifts against the first. Nothing here is blocked by not having it.

### Code-Review, 0

`Found 0/1 approved changesets -- score normalized to 0`

Accepted. The approving review count on this board is zero, which the ruleset
output above shows, and issue #55 is where that parameter is walked against the
target gate and where the reason for the value is written. This check reads merge
history, so it will report differently as history accumulates whatever the
parameter says.

The honest form of what this check is asking for is in the contributing rule that
nothing reaches the mainline that only its own author has read. That is a rule a
person follows here and no machine enforces it.

### Contributors, 0

`project has 0 contributing companies or organizations -- score normalized to 0`

Accepted, and outside this repository's control. The check counts the
organisations that the accounts contributing here belong to. Nothing a change in
this tree can do moves it, and whether this board takes contributions from
outside at all is an open question on issue #46 rather than something to score
against.

### Dangerous-Workflow, 10

`no dangerous workflow patterns detected`

Accepted. What holds it there is that no workflow uses `pull_request_target` or
`workflow_run`:

```
git grep -nE 'pull_request_target|workflow_run' -- .github/workflows/ ; echo "exit=$?"
exit=1
```

and that no `run:` script interpolates a workflow expression at all. The DCO job
is the one that needs event data, and it passes the two commit hashes through
`env` so an attacker-crafted ref cannot be evaluated as shell. The zizmor gate
refuses a new workflow that departs from that, so this check has a second
reader.

### Dependency-Update-Tool, 0

`no update tool detected`

Owed rather than accepted, and issue #56 holds it. The pinned actions in this tree
are pinned by commit, which is the right shape and also the shape that goes stale
silently: a pin never moves, so a pinned action with a published advisory stays in
place until somebody looks. #56 either adds the tool that watches them or records
that nothing does. Until then this score is correct about this repository.

### Fuzzing, 0

`project is not fuzzed`

Accepted. There is no runner in this tree yet, so there is nothing to fuzz. When
there is, the runner's input is a checkout somebody wrote, which is exactly the
shape fuzzing is for, and issue #61 is where that input is first treated as
untrusted. Reopen this line when #61 lands rather than before.

### License, 0

`license file not detected`, with `Warn: project does not have a license file`

Correct, and it is the most consequential zero here. Without a licence file,
default copyright applies and nobody has permission to reuse anything in this
repository. Which licence is an open question on issue #46, and issue #47 lands
the file once that is answered. This score should go to 10 in the same change.

### Maintained, 0

`project was created within the last 90 days. Please review its contents
carefully`

Did not measure anything. The check states its own bound: it cannot assess a
repository younger than ninety days, so this is the absence of a measurement
rather than a finding. It will start reporting on its own.

### Packaging, -1

`packaging workflow not detected`

Did not apply, and whether it ever should is an open question. Entry four on
issue #46 asks whether this board publishes downloadable artefacts at all. If the
answer is no, this check stays at `-1` permanently and that is the correct
outcome rather than a gap.

### Pinned-Dependencies, 10

`all dependencies are pinned`, with `9 out of 9 GitHub-owned GitHubAction
dependencies pinned` and `2 out of 2 third-party GitHubAction dependencies
pinned`

Accepted, with one thing the score does not cover, written down here because a 10
invites a reader to stop looking. The check counted GitHub Actions. The zizmor
workflow also fetches a Python package at run time, pinned by version rather than
by hash:

```
git grep -n 'zizmor@' -- .github/workflows/zizmor.yml
.github/workflows/zizmor.yml:63:        run: uvx --no-build "zizmor@${ZIZMOR_VERSION}" --strict-collection --min-severity=low --format=sarif . > results.sarif
.github/workflows/zizmor.yml:82:        run: uvx --no-build "zizmor@${ZIZMOR_VERSION}" --strict-collection --min-severity=low --format=plain .
```

A version is not a hash. `--no-build` means the prebuilt wheel is used and no
source-dist build script runs, which is the larger half of the risk, and the
remaining half is that the wheel for a given version is trusted to be the wheel it
was. That is accepted here rather than fixed, because pinning it by hash is a
change to a gate this triage is not the place to make. It belongs with #56, which
is where keeping pins current is decided.

### SAST, 0

`no SAST tool detected`, with `Warn: no pull requests merged into dev branch`

Partly wrong and partly owed. Static analysis of the workflow YAML runs on every
pull request, in `zizmor.yml`, and it fails the build on any actionable finding.
The check did not see it, and the warning says why: it reads merged pull requests
and there were none.

What is genuinely absent is static analysis of the runner, because there is no
runner. Issue #23 adds it. Re-read this score once merge history exists rather
than treating it as a statement about the tree today.

### Security-Policy, 9

`security policy file detected`, with `Warn: One or no descriptive hints of
disclosure, vulnerability, and/or timelines in security policy`

Accepted, and the detail matters more than the nine. The policy the audit found
is the organisation-level default, not a file in this repository:

```
git ls-files SECURITY.md | wc -l
0
```

So this repository inherits a policy rather than carrying one, and the audit
cannot tell the difference. Issue #10 adds the file here. The point deducted is
about the policy naming a disclosure route in enough detail, which is the same
thing issue #51's decision needs the policy to carry, so the two land together or
the deduction stays.

### Signed-Releases, -1

`no releases found`

Did not apply, and it cannot pass until there are releases. Accepted now, reopened
by the release milestone. Whether artefacts are published and whether they are
signed are both entry four on issue #46.

### Token-Permissions, 10

`GitHub workflow tokens follow principle of least privilege`, with
`no jobLevel write permissions found`

Accepted. Every workflow declares a read-only or empty scope at the top level and
grants write only where a job needs it. The zizmor gate refuses a new workflow
that does otherwise, so this one is held by a check rather than by memory.

### Vulnerabilities, 10

`0 existing vulnerabilities detected`

Accepted, and it currently costs nothing to hold. The check queries the advisory
database for this repository's dependencies, and the tree declares none. It will
become a real result on the day the runner has a dependency, which is also the day
the `dependency-review` gate starts having something to review.

## What this repository depends on

The runner has no direct dependencies, because there is no runner in this tree
yet. That is an empty list produced deliberately rather than a section nobody
wrote, and it is the state a later reader should check against:

```
git ls-files cmd internal | wc -l
0
```

Issue #12 creates the runner and issue #2 decides what it is built in. When either
lands, this section gets the list and the reason each entry is present.

What the repository does depend on today is what its workflows fetch, which is
what the claim about the runner's dependency surface will eventually be measured
against. All of these are pinned by commit, with the version in a comment:

- `actions/checkout`, in all five workflows, to get the tree the check reads.
- `actions/dependency-review-action`, to compare a change against the advisory
  database.
- `ossf/scorecard-action`, which is the audit this document triages.
- `actions/upload-artifact`, to keep the raw result of that audit readable after
  the run.
- `github/codeql-action/upload-sarif`, to put findings in the code-scanning tab.
- `astral-sh/setup-uv`, to provide the runner that fetches the workflow auditor.

And one thing fetched at run time rather than pinned by commit, named in the
Pinned-Dependencies section above: `zizmor`, by version.

## What this document does not do

Nothing here is enforced. No check reads this file, no route refuses a score that
regressed, and no route notices when an outcome above stops being true. What
stands behind it is a person re-running the command at the top and comparing.

The outcomes are judgements about this repository at the commit named above. A
later run against a later commit is a different measurement, and the honest thing
to do with a disagreement between this document and a newer run is to trust the
run.
