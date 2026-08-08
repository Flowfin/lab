# 0002. Repository layout

## What was decided

Where a thing lives decides what can be checked about it. This record fixes the
layout, and the list below is the whole of it.

`experiments/<slug>/` holds one experiment and everything it needs. One
directory, one question. Nothing outside this tree is an experiment.

`experiments/<slug>/EXPERIMENT.md` is the record of that experiment. Its shape is
decided separately. Its location is fixed here, so a checker finds every record
by walking one directory instead of searching the tree.

`cmd/lab/` and `internal/` hold the runner. `cmd/lab/` is the entry point and
`internal/` is everything it is built from. The runner imports nothing from
`experiments/`.

`docs/decisions/` holds the decision records, in the shape record `0000` sets
out. `docs/` holds everything else a reader needs and that is not a decision.

`testdata/` holds the runner's own fixtures. These are fixtures and they are not
experiments. A check exercised against the real `experiments/` tree proves the
state of that tree on the day it ran and proves nothing about the check, so the
fixtures a check is judged by live here.

`.github/` holds the workflows and the templates. It is named because it is
already in the tree, and a record that listed every directory except the one that
was here first would be a record this repository disagreed with on the day it
landed.

At the commit this record lands on, the root of the tree holds two directories,
`.github` and `docs`:

    git ls-tree -d --name-only HEAD
    .github
    docs

The rest of the list is where things go as they arrive, not a claim that they are
here now.

A directory at the root of the tree that this record does not name is refused
rather than tolerated. Adding one is a change to this decision, argued in an
issue and landed as a later record, rather than something that happens quietly
while somebody is doing something else. The refusal is built in issue #65 and it
reads this list.

The rule that makes the layout worth having is that the runner and the
experiments never mix. The runner imports nothing from `experiments/`, and
nothing under `experiments/` is on the path any part of the runner is built from.
The reason is that an experiment is allowed to be abandoned, deleted or rewritten
in another language, and none of that may have any consequence for the tool that
judges it. The same separation lets the tool be released on its own schedule,
because its sources do not move when an experiment does.

## What it applies to

The whole tree, and every change that adds a file to it. It applies at the root
strictly, because that is where the refusal reads, and inside the named
directories loosely, because what a runner needs under `internal/` is an ordinary
judgement rather than a decision.

It applies to the checks too. A check that has to guess where a record lives is a
check that goes wrong quietly, so anything the runner walks is named here.

## What else was considered

A flat tree, with experiments at the root and the runner mixed in among them.

Keeping the runner in a separate repository from the experiments.

Leaving the root open, so a new directory needs no decision.

Putting the experiment records in one directory of their own, separate from the
experiment code they describe.

## What each rejected option would have cost

A flat tree costs the walk. There would be no directory a checker could walk to
find every experiment record, so finding them means matching filenames across the
whole tree, and a stray file with the right name becomes an experiment nobody
meant to declare. It also removes the separation rule entirely, since an
experiment and the runner would sit side by side with nothing between them.

A separate repository buys the strongest possible separation and costs the thing
this board is for. An experiment would then be a change in one repository judged
by a tool in another, which is two checkouts, two issue trackers and a version
skew between the record format and the checker that reads it. The import rule
above buys most of the separation for none of that.

An open root costs the property that makes the layout checkable. A directory
nobody decided on appears during unrelated work, the refusal has nothing to read,
and within a few months the tree is the one nobody can walk, which is the failure
this board was opened to avoid.

Splitting the records from the code they describe costs the one-directory rule.
An experiment would live in two places, which means two things to move when it is
renamed and two to delete when it is abandoned, and the code-removal decision
becomes much harder to state, because removing an experiment's code would leave
an empty directory in one tree and a record in the other.
