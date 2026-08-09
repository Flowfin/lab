# 0001. The runner is written in Go

## What was decided

The runner is written in Go, and the version this repository pins is the 1.26
line. The module declares it and every job that sets up a toolchain names the
same line, so a contributor running the checker in a checkout and a job running
it on a build machine are running the same language.

Nothing in the tree pins anything yet, because the module does not exist. At the
commit this record lands on there is no `go.mod`:

    git ls-files go.mod
    (no output)

The record fixes the choice and issue #12 creates the module that carries it.
Which patch release of the 1.26 line a given run uses is not written here. A
version restated in a document drifts against the file that decides it, and the
module is the file that decides it.

Four reasons, in the order they matter to this repository.

An operator can run the checker without installing anything. Go builds a single
static binary per platform from one build machine, so somebody who wants to
check this board's claims downloads a file and runs it. That matters more here
than it would elsewhere, because the whole argument for this repository is that
its claims can be checked by somebody who does not already trust it, and a
person who has to install a toolchain belonging to the project they are
checking is in a weaker position than one who does not.

The runner's job is a job Go does with nothing outside its standard library.
Reading a checkout, walking a directory, parsing text records and refusing the
malformed ones is filesystem and string work. A checker that decides whether
this repository's rules were kept is security relevant in the small way that
matters, which is that somebody will eventually rely on its verdict, so the
number of third-party packages that could change what that verdict says stays
near zero.

The test harness is part of the toolchain. This repository requires a test for
every refusal it ships, so the harness is load-bearing rather than convenient,
and a harness that arrives with the language is one nobody has to choose,
version or replace later.

It keeps the runner independent of the experiments. An experiment here is
written in whatever the question needs, and the tool that judges the records is
not allowed to acquire a dependency on any of that. Record `0002` already stops
the runner importing anything under `experiments/`, and a language the
experiments have no reason to share keeps that separation from eroding by
habit.

The cost being paid knowingly is that this organisation gains a second
toolchain alongside .NET. It is accepted because the runner is not a plugin,
never ships inside one, and is downloaded rather than built by most of the
people who will use it, so the second toolchain is a cost carried by whoever
works on the runner and by nobody else.

## What it applies to

The runner and everything under `cmd/lab/` and `internal/`. It applies to the
checks that read records, since those are part of the runner, and to the
runner's own fixtures under `testdata/`.

It applies to nothing under `experiments/`. An experiment is written in whatever
answers its question, and this record is not a preference for Go anywhere except
in the tool that judges records. An experiment written in C#, Python or shell is
in no tension with it.

It does not settle what a released binary is or whether one is published at all.
Which platforms the runner builds and runs its suite on is record `0012`.
Whether this board publishes downloadable artefacts is entry four of issue #46
and is not decided here.

## What else was considered

C# and .NET, which is what the plugin work in this organisation is built in.

Python.

Shell.

Rust.

## What each rejected option would have cost

C# and .NET is the option with the strongest argument behind it, and the
argument is about the wrong artefact. An experiment that graduates from this
board is likely to be C#, because the boards it graduates to are, and that is a
statement about the experiments rather than about the tool that reads their
records. Choosing it would mean somebody who only wants to check this
repository installs an SDK first, which costs exactly the property the first
reason above was chosen for. It would also put the runner in the same toolchain
as the code it is meant to stay independent of, so the separation record `0002`
builds would rest on discipline rather than on the two having nothing in common.

Python is good at the text work this runner does, and shipping it to an operator
means shipping an interpreter or a bundler with it. That packaging cost is paid
on every platform, every release, and it lands on the person who wanted to check
a repository rather than on the person who chose the language. The writing
convenience is real and it is smaller than what it costs at the far end.

Shell costs the thing this repository's central claim rests on. A board whose
argument is that a machine refuses violations cannot rest that machine on a
language with no test harness and no type checking, so the proof that a refusal
bites would have to be built from scratch and maintained by hand. Word splitting
has broken gates of exactly this kind before, and the failure is quiet: a gate
that word-splits its input refuses valid work and passes the rest, which reads
as a strict check rather than as a broken one.

Rust would meet every reason above and costs more than it buys here. The runner
has no performance requirement worth the compile times and no memory-safety
requirement that Go does not already meet for a program that reads text and
exits, so the additional cost is paid in how long a contributor waits and in how
many people can read the checker without learning a language first. It is not
rejected as unsuitable, and if this tool ever grows a reason that Go cannot
meet, superseding this record with one that takes it is the route.
