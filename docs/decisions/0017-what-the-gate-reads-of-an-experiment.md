# 0017. What the gate reads of an experiment

## What was decided

Every command this repository is gated by names the runner's packages rather
than the whole module. Where a command takes a package pattern that is
`./cmd/... ./internal/...`, and where it takes directories it is `cmd` and
`internal`. It is one set of commands rather than two, so what CONTRIBUTING.md
asks somebody to run before they push and what a job runs are the same strings.

Record `0009` decided the property this implements. It says nothing that runs on
its own builds or executes anything under `experiments/`, and it names no
mechanism, so the sentence was true of the intention and not of the tree. This
record supplies the mechanism and takes nothing away from that one.

The state that made the gap visible. An experiment lives inside this module, so
its Go package was in every pattern the commands used, and the first experiment
to land put it there:

    git rev-parse HEAD
    9807dc783b9b981b652bdf00ac1e1467171a1051
    go list ./...
    github.com/Flowfin/lab/cmd/lab
    github.com/Flowfin/lab/cmd/pullrequest
    github.com/Flowfin/lab/experiments/reading-a-tree-of-records
    github.com/Flowfin/lab/internal/check
    github.com/Flowfin/lab/internal/hardware
    github.com/Flowfin/lab/internal/invariants
    github.com/Flowfin/lab/internal/prose
    github.com/Flowfin/lab/internal/pullrequest

Six build entries, three suite entries, the vet job and the static analysis all
compiled that package on every pull request, and the formatting job read its
bytes. A prototype that does not compile for one of the six platforms would have
held the board red, and the failure would have named a platform the experiment
was never about.

The two roots are record `0002`'s list read a second way. That record puts the
entry point in `cmd/lab/` and everything the runner is built from in
`internal/`, and says the runner imports nothing from `experiments/`. So the
question of what the gate covers is a question about the layout rather than
about a pattern somebody picked.

One thing nothing refuses, written here rather than left to be found. A Go file
at the root of the tree, or under a root directory the layout names that is
neither `cmd/` nor `internal/`, is outside these patterns, and no command
reaches it. A new root directory is refused by
`root-holds-a-directory-the-layout-does-not-name`, so the shape that escapes is
a package added under a directory already named, and what stands behind that is
a reader. That is a smaller hole than the one this record closes and it is not
zero.

## What it applies to

Every command in CONTRIBUTING.md and every job that runs one, now and later. A
job added afterwards that walks this module with a Go toolchain takes the same
two patterns rather than the whole module.

It does not apply to what a person runs in their own checkout. `go run
./experiments/reading-a-tree-of-records` is written in that experiment's record
and still runs from the root of a checkout, which is one of the reasons the
mechanism is this one and not another.

It does not apply to the checks that read an experiment's record as text.
Record `0009` already separates reading a record from executing an experiment,
and nothing here narrows what the record checks walk.

## What else was considered

Giving each experiment its own Go module.

Holding an experiment's Go files behind a build constraint.

Holding experiments to the gate and writing that down as deliberate.

Leaving the patterns alone and letting the first abandoned prototype decide.

## What each rejected option would have cost

A module per experiment costs a landed record its command. A directory holding
its own `go.mod` leaves the root module, which is the attraction, because every
pattern stops at it with nothing written down anywhere. The cost arrives at the
other end: `go run ./experiments/reading-a-tree-of-records` stops resolving from
the root of a checkout, and that command is in
`experiments/reading-a-tree-of-records/EXPERIMENT.md` twice, once in the method
and once inside the quoted answer. A record here gains lines and does not have
them replaced, so the repair would be an answered record whose method names a
command that does not run. It also puts a toolchain version in the tree per
experiment, on a board whose whole argument is that writing one should be cheap.

A build constraint costs the same command for the same reason one level down. A
file every tag excludes is not run by `go run` either, so the record's command
would need a flag it does not carry. It costs one thing the module option does
not: the constraint is a line at the top of a file that somebody has to
remember, nothing refuses its absence, and the day it is forgotten is the day an
abandoned prototype reddens the board, which is the failure being prevented
here.

Holding experiments to the gate is the option that would replace `0009` rather
than sit beside it, and nothing has happened to weaken the argument that record
makes. Its first reason is still the strongest: the code left behind by an
experiment that answered no is often exactly how it failed, and a gate that goes
red on that code leaves two ways out, deleting the evidence or rewriting the
answer. It would also make every toolchain any experiment ever wants a
dependency of this repository's checks.

Leaving the patterns alone costs `0009` its meaning. A decision record that is
not true of the tree is worse than an absent one, because it is quoted by
somebody who has no reason to go and check.

## The condition that reopens this

Record `0009` carries the condition that reopens the property, and this record
follows it rather than holding one of its own. If experiments start carrying
code that is meant to keep working, the mechanism here is not what is wrong
first.
