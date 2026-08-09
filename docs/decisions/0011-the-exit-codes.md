# 0011. The runner's exit codes

## What was decided

Four codes and no others. A caller reading only the number learns which of four
things happened, and a caller that wants to know what was refused reads the
output, because an exit code is a summary and the run is the evidence.

`0` means the run completed and refused nothing. It does not mean the tree is
good. It means nothing this runner judges was found wrong, which is a narrower
statement, and anything reading this code is entitled to no more than that.

`1` means the run completed and refused something. This is the ordinary red a
contributor sees, and it is the only code that carries refusals. A caller that
sees it can rely on the output naming what was refused.

`2` means the runner could not do its job. A path that is not a directory, a
record it cannot read at all, a verb or an argument it does not understand.
Separating this from `1` is the whole reason for having more than one non-zero
code. A gate that treats them alike reports a broken invocation as either a
clean tree or a violation, and both readings are wrong in the way nobody
investigates: the first is silent and the second sends somebody looking for a
refusal that was never made.

`3` means the run was asked for something and delivered nothing. This is what
the hardware harness needs when it was asked to run and every test skipped.
Being asked for coverage and producing none is a failure of the request even
though no assertion broke, and reporting it as `0` would let a run that
exercised nothing pass for a run that found nothing.

Two properties come with the set rather than after it.

Every code the runner can return is reached by a test in the default suite. A
code that exists only in this document is caught here rather than by an
operator, and the test is what keeps the document and the runner from drifting
apart.

No code is added later without superseding this record. A workflow keyed on a
number is a reader of this contract whether or not anybody told it so, and the
day a fifth code appears is the day some job's meaning changes without its file
being edited.

What the runner returns at the commit this record lands on is narrower than the
set, and the gap is written down rather than left for somebody to discover.
`0` and `2` have producers and tests. `1` has its mapping in the runner and a
test that reaches it through the walk's result, and no walk can produce a
refusal yet, because nothing refuses anything yet. `3` has no producer at all,
because the hardware harness it belongs to does not exist. Both arrive with the
work that produces them, and neither is written into the runner as a branch
nothing can reach, since an unreachable branch is one nobody has proved and it
would be the branch deciding whether a gate goes red.

## What it applies to

Every verb the runner has and every verb it grows, including the ones that do
not exist yet. It applies to the hardware harness when that is built, which is
where `3` comes from.

It applies to anything that reads the number: a workflow step, a pre-push hook,
a person at a shell. Those readers are the reason the set is small and fixed.

It does not apply to what the runner prints. The output says what was examined
and what was refused whatever the code is, and a caller that needs the detail
reads the output rather than a wider code space.

## What else was considered

One non-zero code for everything.

A larger space modelled on a well-known tool, with a distinct code per class of
problem.

A code for a run that completed with refusals somebody had waived.

Returning the number of refusals as the exit code.

## What each rejected option would have cost

One non-zero code collapses a broken invocation into a refusal, and the
collapse is invisible because both look like a red check. Somebody investigating
goes looking in the tree for the violation, finds none, and the actual fault is
a mistyped path. The cost is paid every time it happens and it is paid by
whoever did not write the gate.

A larger space costs promises. A code nothing returns is a promise, and the
operator guide would then explain outcomes that cannot happen, which teaches a
reader that the document describes something other than this tool. It also
grows without argument, because there is always one more distinction worth
making, and each one is a number some workflow may already be keyed on.

A waiver code costs the meaning of `1`. A refusal that was waived is still a
refusal, and a separate code would let a job treat waived and clean alike while
the output says otherwise. Whether a refusal can be waived at all is a question
for whatever builds waivers, and it is not answered by adding a number here.

Returning the count costs everything above 255 and everything below 4. A count
of 256 refusals is zero on most systems, which reports the worst tree this
runner could meet as a clean one, and a count of 2 is indistinguishable from
the runner failing to start.
