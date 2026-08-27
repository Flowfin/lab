# Contributing

This board takes experiments from anybody, and it is for questions that will
probably fail. Everything here is public from the first commit, a failed
experiment included, and the one rule that keeps that from producing a
graveyard is that every experiment states its question before it starts and its
answer when it stops. The answer may be no. An experiment with no written answer
is not finished, it is abandoned, and the difference is meant to be visible.
Three rules have to be read before you start rather than after, because nothing
here refuses a violation of any of them: what may never be committed, in
[0006](docs/decisions/0006-everything-here-is-public.md) and again below; what
happens when an experiment finds a flaw in shipped software, in
[0010](docs/decisions/0010-a-flaw-in-shipped-software.md); and what an
experiment may do with real data, in [docs/privacy.md](docs/privacy.md), whose
conditions are
[0025](docs/decisions/0025-real-data-in-an-experiment.md). Why the door is open
at that price is
[0024](docs/decisions/0024-who-may-run-an-experiment-here.md).

## Before you push

Run these, in this order, from the root of a checkout:

```
go build ./cmd/... ./internal/...
go vet ./cmd/... ./internal/...
gofmt -l cmd internal
go test -count=1 -v ./cmd/... ./internal/...
```

Every one of them names the runner rather than the whole module, and an
experiment's code is outside all four. That is
[docs/decisions/0017-what-the-gate-reads-of-an-experiment.md](docs/decisions/0017-what-the-gate-reads-of-an-experiment.md),
which supplies the mechanism record
[docs/decisions/0009-the-checks-do-not-run-experiment-code.md](docs/decisions/0009-the-checks-do-not-run-experiment-code.md)
decided the property for. A prototype you abandon half-written is a thing this
board takes, and it holds nothing red. What follows from that is that nothing
compiles your experiment for you: run it yourself, from a checkout, at the
moment you want the answer.

`gofmt -l` prints the files it would change and exits zero either way, so read
its output rather than its exit code. No output is the passing result.

The two flags on the suite are not decoration and the command said `go test
./...` until they were added. `-count=1` runs the tests rather than replaying a
cached result, so a green run is one that happened on this machine. `-v` is
what makes the run say what it did not cover: `go test` prints nothing at all
for a package that passes, and the line saying the integration-hardware harness
was not asked for and what asking would cost is printed by a package that
passes. Without `-v` a run that covered less than the whole set reads exactly
like one that covered it and found nothing, which is the reading the harness
under `internal/hardware/` exists to prevent.

That is what the gate runs on the server, so running it here means running the
same thing rather than something that resembles it. What the checks are and what
each one covers is not listed in this document. The run says what it examined,
and a list written down here drifts against the thing it describes:

```
go run ./cmd/lab check .
```

The exit codes that command can return are a contract, in
[docs/decisions/0011-the-exit-codes.md](docs/decisions/0011-the-exit-codes.md),
and anything keyed on one of them is reading that record whether or not anybody
said so.

Nothing runs any of this for you on the way out. A fresh clone installs no git
hook, so the commands above are yours to run. If you skip them, the first thing
that notices is the pull request, and by then the failure is on the board rather
than in your terminal.

## Signing your work

Every commit carries a `Signed-off-by` trailer matching its author. That trailer
is you asserting the Developer Certificate of Origin, whose text is in
[DCO](DCO) at the root of this repository. Read it once. It is short, and it is
what you are asserting.

`git commit` takes `-s` and writes the trailer for you from your configured name
and address. A branch whose commits predate the habit takes
`git rebase --signoff` against the base you branched from. A commit whose
trailer does not match its author does not count, so if you change your git
identity mid-branch, check the trailers before you push.

## Starting an experiment

The question is written first. Not sketched, not held in your head while you get
something working, written.

Create `experiments/<slug>/` and `experiments/<slug>/EXPERIMENT.md` in one
commit, with the record in state `asking` and the question in it. One directory
holds one experiment and one question. A directory holding three questions has
no state that is true of all of them, and a record whose question is a topic
rather than a question can never be shown to have missed its answer.

The layout is fixed in
[docs/decisions/0002-repository-layout.md](docs/decisions/0002-repository-layout.md)
and the states are fixed in
[docs/decisions/0003-the-experiment-lifecycle.md](docs/decisions/0003-the-experiment-lifecycle.md).
There are three states and no others: `asking`, `answered`, `abandoned`.

## Starting from somebody else's code

An experiment may start from code somebody else wrote, and there is exactly one
place that code may live. The argument is in
[docs/decisions/0019-code-under-another-licence.md](docs/decisions/0019-code-under-another-licence.md);
what follows is what the rule asks of you.

The code goes in `experiments/<slug>/borrowed/`, and that directory carries its
own `LICENSE` naming the terms the code arrives under. One such directory per
experiment. Everything under it is under the licence that file names and
everything outside it is under this board's, so the boundary is something a
person walking the tree can see rather than something they have to be told.
That is the whole point of it: a directory named `borrowed` with a licence file
in it is visible to a walk, and a line in a header is visible only to somebody
who opened the header.

The record declares it too. `Borrowed:` in the header of
`experiments/<slug>/EXPERIMENT.md` names where the code came from and the
licence it arrives under, and it sits with the rest of the header in
[docs/experiment-template.md](docs/experiment-template.md). An experiment that
borrows nothing writes no such line.

Two of those a run refuses and the rest of them it does not, and the difference
is worth knowing before you lean on any of it. A `borrowed/` directory with no
`LICENSE` beside it is refused, and so is a record declaring `Borrowed:` in an
experiment that holds no such directory. A `borrowed/` directory in an
experiment whose record declares nothing passes, because a field added to the
format after
[docs/decisions/0013-how-the-record-format-changes.md](docs/decisions/0013-how-the-record-format-changes.md)
is never refused for being absent. A second quarantine deeper inside an
experiment passes as well, since the check reads `experiments/<slug>/borrowed`
and no other name, so the one-directory limit above is yours to keep rather
than the gate's.

Nothing opens the licence file. A green run says the layout and the declaration
do not contradict each other, and it says nothing about which licence the code
is actually under, or about whether the result may be promoted into a board
under other terms, which record `0019` leaves undecided on purpose.

## Stopping one

Two honest ways to stop, and both of them are finished work.

Write the answer and move the record to `answered`. No is an answer. So is
finding out that the question was the wrong question, which is often the most
useful thing an experiment produces. An answered experiment is finished
whichever way it went, and nothing on this board treats a no as a lesser
outcome. Anyone who ever feels a check here pushing them towards a more
flattering answer should report that as a defect in the check.

Or move the record to `abandoned` and write one sentence saying what stopped the
work. That is a lower bar than an answer on purpose. Moving here is always
available and always cheaper than the alternative, which is leaving a record in
`asking` while nothing happens. Anybody reading the tree may do it, not only
whoever started the work, and an abandoned experiment somebody picks up again
goes back to `asking`.

The record is permanent either way. It is not deleted and it is not rewritten to
hide what it said; what it gains over time is lines added to it, never lines
replaced. The code is not permanent: once there is an answer, the prototype can
be removed in a later commit, and the record gains one line naming the commit
that removed it, with the full hash. That is
[docs/decisions/0004-what-happens-to-the-code.md](docs/decisions/0004-what-happens-to-the-code.md).

## What may never be committed

Git history does not forget, so this is about what is committed at all rather
than about what is in the current checkout. The reasoning is in
[docs/decisions/0006-everything-here-is-public.md](docs/decisions/0006-everything-here-is-public.md).

- No credential, token, key or secret of any kind, including expired and rotated
  ones. An expired credential still reveals its format, its issuer and often the
  account it belonged to.
- No personal data of any kind, whether it is yours or somebody else's. Names,
  addresses, e-mail addresses, device identifiers and account identifiers are
  all covered.
- No copy of a real media library, real account data, or real logs from a
  running server, whoever runs it.
- No file whose licence forbids it being here.

Where an experiment needs real data to answer its question, the data is the
runner's own and the measurement is named in the record before the work starts,
which is [0025](docs/decisions/0025-real-data-in-an-experiment.md), and the data
stays on the host it is already on. It does not enter the tree as a fixture, as
a sample, as an attachment, redacted, or in a screenshot. Redaction is named because it is
the failure that feels safe: a partially masked identifier is still an
identifier, and a screenshot of a list is a copy of the list. What may be
written down is the measurement and the command that produced it.

Nothing in this repository refuses a violation of that list today. What stands
behind it is you reading before you commit and somebody reading the change, and
that is written here plainly so that nobody mistakes the list for a gate. A
mistake here is handled as an exposure rather than as a revert, which means
rotating what leaked and saying so.

## Headless and unelevated

Every test in the default run completes with no graphical session and as an
ordinary user. Both halves bind from the first test, not from the milestone that
builds the mechanism that enforces them, and the reasoning is in
[docs/decisions/0007-headless-by-birth.md](docs/decisions/0007-headless-by-birth.md).

Learn this before you write a test rather than from a red board afterwards. A
test that needs a display runs on one machine, and a suite that runs on one
machine stops being evidence the first time that machine is unavailable. A test
that needs administrator rights is worse, because the operating system answers
by asking a question only an administrator can answer, on somebody else's
machine, where nobody is watching, and that is indistinguishable from a hang.

A test that genuinely cannot meet both halves does not get an exception and it
does not skip itself inside the default run. It goes to the integration-hardware
harness under `internal/hardware/`, which is a separate thing with its own name.
Its tests are behind a build constraint, so the default run does not compile
them, and they are asked for explicitly:

```
LAB_INTEGRATION_HARDWARE=1 go test -tags integration_hardware ./internal/hardware
```

Every test there names the hardware it requires, in words somebody can act on,
and the default run holds them to that by reading the harness source. A
measurement that harness produces is a measurement about the machine it ran on.
It is never reported as the default suite's result.

## What this board does not take

Three things, and an idea that needs any of them is a product idea that belongs
on a product board rather than here.

Nothing a user is asked to install. This board ships a runner for people working
in this repository and nothing that ends up on anybody else's machine as a
product.

Nothing the site advertises. An experiment is allowed to fail, and something
that has been announced cannot fail quietly, which removes the only thing this
board offers.

Nothing another board is allowed to depend on. Work that leaves here does so by
promotion, which is a hand-over recorded on both sides and never a dependency
edge pointing back at an experiment. That route is
[docs/decisions/0005-how-a-result-leaves.md](docs/decisions/0005-how-a-result-leaves.md).
Nothing in this repository ever writes to another one.

## Everything else

Planning happens on the issue tracker first, and every decision that shapes the
architecture is argued there before the code that depends on it exists. What
survives the argument is written down in
[docs/decisions/](docs/decisions/), in the shape
[docs/decisions/0000-how-decisions-are-recorded.md](docs/decisions/0000-how-decisions-are-recorded.md)
sets out. A record is replaced by a later record rather than edited, so the
directory stays a history of what was argued instead of a description of what is
currently there.
