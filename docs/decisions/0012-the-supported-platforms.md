# 0012. The supported platforms and the release targets

## What was decided

Three issues in this plan depend on a list of platforms and none of them creates
it. The build workflow covers the platforms the release will later target, the
suite runs on every platform the release ships a binary for, and the release
workflow produces a binary for each platform the build already covers. All three
point at a set nothing fixes. This record fixes it, before the first workflow is
written against it.

The list is decided from what the runner does. The runner reads a checkout, and
the parts of that job which differ between machines are what earn a place:
whether the filesystem folds case, what separates a path, and what a line ending
arrives as. A platform differing in any of those is a platform the suite has to
run on. A platform differing only in processor architecture is a platform the
release builds for, and running the suite there buys much less.

Six platforms. Three get a build and a suite run, three get a build only.

`linux/amd64`, build and suite. Case-sensitive filesystem, `/` as the separator,
LF endings. This is where the checks themselves run and where most build
machines are, so it is the baseline every other entry is read against.

`windows/amd64`, build and suite. The filesystem folds case, `\` separates a
path, and text arrives with CRLF unless something stops it. It differs from the
baseline in all three of the properties the runner's job depends on, which makes
it the entry that would find the most defects if it were left off, and it is
also where a large share of this organisation's contributors work.

`darwin/arm64`, build and suite. `/` separates a path and endings are LF, so it
agrees with the baseline on two of the three, and its default filesystem
preserves case while comparing without it. That third property is a real
difference and it is the one that behaves unlike either of the other two
entries: a path that is refused on Linux and accepted on Windows may be accepted
here while the bytes on disk keep the case that was written. A checker that gets
case handling wrong stays green on the baseline, so the entry earns its run.

`linux/arm64`, build only. Same filesystem behaviour, same separator, same
endings as the baseline. What differs is the processor, and nothing the runner
does is sensitive to it.

`darwin/arm64` is covered above, so `darwin/amd64` is build only for the same
reason: it differs from an entry that already runs the suite in processor
architecture alone.

`windows/arm64`, build only. It differs from `windows/amd64` in processor
architecture alone, and `windows/amd64` runs the suite.

So this repository builds for six platforms and produces evidence on three. The
three it builds for and does not test on are `linux/arm64`, `darwin/amd64` and
`windows/arm64`. That sentence is here rather than left to be inferred, because
a green board across a matrix implies coverage, and on those three entries the
green means the compiler accepted the source and nothing more.

Whether a hosted runner is offered for each of those three is a separate
question from this decision and does not move it. Where one is not offered, the
entry is build-only by necessity rather than by choice, and the list is the
same. What that question decides is whether the gap could be closed by paying
for it, and the first workflow written against this record is where it gets
settled, since a run either arrives on the label or does not. Until such a run
exists, any statement here about which labels are available would be a claim
rather than a measurement, so none is made.

What each entry costs. Every platform on this list can hold a merge on its own,
and that is the point of having it rather than a drawback of it. The suite
entries cost a run each on every pull request, and they cost the slowest of the
three in wall time, since a merge waits for the last one. The build-only entries
cost a compile each and produce no evidence, so their whole value is that a
release does not discover on the day it is cut that the source does not build
somewhere. The shared cost across all six is that a platform nobody uses
produces failures nobody acts on, which teaches readers that red is survivable,
and that is the reason the list is six rather than everything the toolchain can
target.

This record fixes which platforms a release targets. It does not decide whether
this board publishes downloadable release artefacts at all, which is entry four
of issue #46 and is open. If the answer there is source only, the three
build-only entries lose their reason and this record is superseded by one that
says so. The three suite entries do not depend on that answer, because a
contributor runs the checker in a checkout whether or not anything is ever
published.

## What it applies to

The build and test workflow in issue #13, the platform suite in issue #57, and
the release workflow in issue #41. Each of them names exactly the platforms this
record names and no others, in the role this record gives them.

It applies to the platform any later work assumes. A check written to a path
separator or to a case-folding rule is written against this list rather than
against whatever machine its author is sitting at.

It does not apply to experiments. What an experiment runs on is a property of
the question it asks, and record `0009` keeps every automatic run out of
`experiments/` in any case.

## What else was considered

Building everything the toolchain can target.

Building for one platform only.

Running the suite on all six entries.

Running the suite on the baseline only, and treating the other entries as build
targets.

## What each rejected option would have cost

Building everything the toolchain can target costs the list its status as a
decision. It would become a property of the compiler, changing when the compiler
changes, and most of the entries would have no contributor, no runner and no way
to reproduce a failure. A failing build on a platform nobody can check out is a
merge held by something nobody can fix.

Building for one platform only costs the tool its ability to read its own
repository. Case folding and line endings are exactly the defects a checker of
this kind has, and they are defects on the platforms this organisation's
contributors actually use. Shipping a single-platform release means the people
most likely to trip the checker are the people who cannot run it.

Running the suite on all six costs twice the runs for evidence that repeats
itself. An arm64 run of a program that reads text and exits exercises the same
paths as the amd64 run beside it, so the second run of each pair would be a
merge-holding job whose failures would almost always be infrastructure rather
than the runner.

Running the suite on the baseline only is the cheapest option and it costs the
reason the list was written from the runner's job rather than from convention.
The baseline is the one platform where case folding, path separators and line
endings all behave the way the code was most likely written to assume, so it is
the platform least able to catch the mistakes this checker will make.
