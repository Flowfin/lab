# The operator guide

This page is for somebody who has never seen this board and wants to check it
for themselves. It walks getting the runner, running it once, and reading what
comes back, including the runs that end in a refusal.

## What you get, before you decide to run anything

`lab` reads a checkout of this repository and reports whether its records are
in order and what it examined. Three things somebody deciding in half a minute
whether to run an unfamiliar program wants to know, and each of the three is a
limit rather than a promise.

It changes nothing in the tree it reads. Every verb it has reads and none of
them writes, and the walk is held to that by a test rather than by this
sentence:

    go test ./internal/check -run TestWalkWritesNothing -count=1

That test fingerprints every fixture tree by path and by content, walks all of
them, and fingerprints again, so a file created, changed or removed anywhere
under them is a difference it reports. What it covers is the walk over those
trees. It is not a statement about a tree it has never been pointed at.

It sends nothing anywhere. The claim, the command that tests it, and the bound
on what that test proves are in [docs/privacy.md](privacy.md).

It needs no privileges and no graphical session. Both halves bind from the
first test on this board rather than from a later mechanism, which is
[docs/decisions/0007-headless-by-birth.md](decisions/0007-headless-by-birth.md).

## Getting it

There is nothing to download. This board publishes no binary, no checksum file
and no release, so nothing on this page tells you how to verify a download, and
whether anything is ever published is not decided. The route that exists is a
checkout and a Go toolchain, and the toolchain version comes from `go.mod`:

    git clone https://github.com/Flowfin/lab
    cd lab
    go build -o lab ./cmd/lab

Building the tool from the repository it is about is a weaker position than
downloading one somebody else built, and it is the only position available
today. The source of the checks is in the same checkout, which is the part that
makes it worth anything: what each rule refuses is readable next to the rule.

## The first run

    ./lab check .

Against a fresh clone of this repository at commit `bbfab50`, that printed:

    examined .
    1 experiment directory walked, 1 record read
    16 decision records read
    the time this run read is 2026-08-11T20:42:12Z
    0 refused

Four things in those lines. What was examined, which is the path you gave it.
What the walk found and what it managed to read, counted apart from each other,
because a directory walked with no record read is the shape of an experiment
that states no question. What the run read the clock as, so a verdict about a
date can be placed against the moment it was made in rather than the moment you
are reading in. And what was refused, printed as a number even when the number
is zero.

The counts move as the board grows, so a later run over a later tree prints
different numbers in the same shape. The commit is named above because that is
what the numbers are reproducible against.

## A run over a tree with no experiments

This is what a run against a directory holding nothing produces, and it is a
result rather than a broken run:

    examined .
    no experiments directory in this tree
    0 experiment directories walked, 0 records read
    no docs/decisions directory in this tree
    0 decision records read
    the time this run read is 2026-08-11T20:43:16Z
    0 refused

The two lines saying a directory is not there are what make that readable. A
run that examined nothing and a run that examined everything and found nothing
otherwise print the same zero, and telling them apart is the difference between
a clean board and a run that never happened.

## The listing

`lab list` reports the experiments instead of judging them. Over the same fresh
clone:

    ./lab list .

    examined .
    1 experiment
    the time this run read is 2026-08-11T20:42:15Z
      slug                       state     question written  waiting  needs
      reading-a-tree-of-records  answered  2026-08-11        -        none

Oldest unanswered first, so whatever has been asking longest is at the top.
Nothing here fails because an experiment is old, and the listing exists to make
the choice between answering one and abandoning it a visible one.

## What a refusal looks like

A refusal names the thing to open, says what is wrong in a sentence, and ends
with the rule that refused it. This run was made over a directory holding one
experiment directory with no record in it:

    examined .
    1 experiment directory walked, 0 records read
    no docs/decisions directory in this tree
    0 decision records read
    the time this run read is 2026-08-11T20:43:40Z
    1 refused
      experiments\a-question-with-no-record: there is no EXPERIMENT.md in it, so it states no question (experiment-has-no-record)

The separator in that path is the one the host uses, and that run was made on
Windows.

The name in brackets is the rule. It is the string to search for in the
checkout when the sentence is not enough, and it takes you to the place where
what the rule refuses and what it deliberately does not judge are written
together. The report still says what was examined, so a run that refused
something also tells you how much of the tree it got through.

## The exit codes

What each code means is
[docs/decisions/0011-the-exit-codes.md](decisions/0011-the-exit-codes.md), and
it is not restated here. A second copy drifts against the record, and the copy
is the one a reader finds first.

What produces each of them is the half that record cannot show you.

The first run above returned `0` and the refusing run returned `1`. Both are
completed walks, and the difference between them is in the output rather than
in whether the run worked.

`2` comes from an invocation the runner cannot act on, rather than from
anything in the tree. Three ways to reach it, each printing to standard error:

    ./lab check README.md
    lab check: README.md is not a directory

    ./lab frobnicate
    lab: unknown verb "frobnicate"
    (and then the help text, because the verb is not one it has)

    ./lab
    (the help text, because no verb was given)

The help text is described rather than pasted in both of those, because a copy
of it here drifts against the runner that prints it, and the runner is the
thing a reader is checking.

`3` is not a code this command returns. It belongs to the integration-hardware
harness under `internal/hardware`, which is asked for separately, and it is
declared where its only producer is:

    git grep 'ExitAskedAndDeliveredNothing = ' -- internal/hardware
    internal/hardware/hardware.go:const ExitAskedAndDeliveredNothing = 3

That command carries no line number on purpose. With `-n` the paste names a
line rather than a declaration, and an edit anywhere above the constant moves
the line without changing anything the sentence claims, so the quotation goes
stale while the claim it supports stays true. The file and the declaration are
what the claim rests on and neither of them moves.

So the record fixes four codes, `lab` returns three of them, and a caller
keyed on any of the four is reading that record whether or not anybody said so.

## What a green run does not say

That it refused nothing. That is the whole of it, and it is narrower than it
looks. It is not a statement that the tree is good, that an experiment worked,
or that anything a record claims is true. An experiment that answered no passes
exactly like one that answered yes, because a no is finished work here.

Some rules on this board have no mechanism behind them and can have none. The
largest of those is that real data stays on the host it is already on: nothing
in a checkout can tell a number somebody measured from a number they made up.
That is written plainly at the rule in [docs/privacy.md](privacy.md) rather
than left for a green tick to imply otherwise.

## The notice, the privacy document and the licence

[NOTICE.md](../NOTICE.md) at the root of the checkout is the intended-use
notice. It places responsibility for lawful use on whoever deploys and runs
this, and it is a notice rather than a control: printing it or shipping it
prevents nothing.

[docs/privacy.md](privacy.md) is the operator-facing half of what happens to
real data, and this page says the same thing it does deliberately, because a
reader who downloads a tool is not certain to open the other document. Where an
experiment needs real data to answer its question, the data stays on the host it
is already on. Only the measurement is written down. Nothing here uploads,
phones home, or reports usage, and there is no telemetry to turn off because
there is none to turn on.

[LICENSE](../LICENSE) at the root of the checkout is the GNU General Public
License version 3, and those are the terms this tree carries. The licence this
repository declares to its own checks is the same one, so a run here reports
what it compared the file against rather than reporting that it was not asked.
[docs/decisions/0018-the-licence-of-this-board.md](decisions/0018-the-licence-of-this-board.md)
is where the answer and what it costs are written down.

## The record format may change

This is not a promise of stability. It is a statement of what it is not.

The experiment record format may change. A change to it is decided in a record
under docs/decisions/ and is not made silently. What happens to the records
already on the default branch on the day it changes is
[docs/decisions/0013-how-the-record-format-changes.md](decisions/0013-how-the-record-format-changes.md),
which is where that was decided, rather than restated here.

There is no changelog in this repository and no release for one to carry an
entry for, so today a change to the format is announced by the record that
decides it and by nothing else.
