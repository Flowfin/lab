# 0009. The checks do not build or run experiment code

## What was decided

Nothing that runs on its own builds or executes anything under `experiments/`.

That covers the build and test workflow, the record checks, and any other job a
pull request, a push or a timer starts. None of them compile an experiment, none
of them run an experiment's tests, and none of them install a toolchain that
exists only because some experiment wanted it.

What runs on its own is the runner's own suite and the checks that read records.
Those are this repository's own code and this repository's own text, and they run
against `testdata/` fixtures rather than against the real `experiments/` tree.

What runs only when asked is everything belonging to an experiment. A person can
run an experiment's tests explicitly, from a checkout, at the moment they want
the answer. The hardware harness is the named route for the experiments that
cannot be answered without a real device. No automatic run ever takes either
route.

The first reason is the one this board was opened for. An experiment is allowed
to fail, and the code left behind by an experiment that answered no is often
exactly how it failed. A gate that goes red on that code leaves two ways out,
deleting the evidence or rewriting the answer, and both destroy the thing the
record exists to preserve. Record `0004` permits deleting a prototype and
deliberately does not require it, which only holds if keeping a broken one costs
nothing.

The second is the dependency surface. Experiments are written in whatever
language the question needs, which is the point of keeping the runner independent
of them. Building all of them on every pull request means this repository's
checks carry every toolchain any experiment ever wanted, on a board whose runner
was chosen partly for carrying almost no dependencies at all.

The third is smaller and harder to undo. Record `0002` keeps the runner from
importing anything under `experiments/`, so that no experiment can break the tool
that judges it. Executing experiment code inside the job that judges records
would hand that back by a different route, with the tool and the thing it judges
in one process again.

Three consequences, written down here rather than left to be inferred.

The headless requirement binds what runs on its own. That is the runner's suite
and the record checks, and it means those complete with no display and no
elevation. It is not a claim that every experiment's tests are headless, because
nothing automatic runs them. That is narrower than the requirement looks at first
reading, and it is written narrow here rather than broad in a document and narrow
in practice.

The declaration an experiment makes about which harness it needs is a statement
to a reader deciding whether they can reproduce the work. It is checked for
agreement with where that experiment's tests are registered, and the refusal over
a mismatch is about the record being consistent with itself. It is not a
scheduling instruction. No run reads it and acts on it, and an experiment
declaring that it needs a device does not thereby cause anything to look for one.

An experiment's measurements are produced by whoever runs the experiment. The
record carries the command and what it produced. Nothing verifies that the
command was run, that the machine it ran on was what the record says, or that the
number is real. Review is what catches an unreproducible measurement, and the
promotion checklist asks for it again at the point where the number starts
mattering to somebody other than its author.

## What it applies to

Every automatic run this repository defines, now and later, and every path under
`experiments/`.

It does not apply to the record checks reading an experiment's `EXPERIMENT.md`.
Reading a record as text is not executing the experiment, and those checks are
the reason the tree is walkable at all.

It does not constrain what a person does in their own checkout.

## What else was considered

Building and testing every experiment on every pull request.

Building them in a separate automatic job that is not required to be green.

Letting each experiment opt in to being built automatically.

## What each rejected option would have cost

Building everything on every pull request costs all three reasons above at once,
and it costs one more thing that arrives sooner than any of them. The first
abandoned prototype would hold the board red until somebody deleted it, so the
board would be forced to choose between a permanently failing gate and destroying
the evidence, on the day the first experiment answered no.

A separate job nobody is allowed to act on is worse than no check. It teaches
every reader that red is survivable on this board, and the next red one they walk
past will be a real one. A check whose result changes nothing is not a weaker
check, it is a check that trains people out of reading checks.

Per-experiment opt-in is the strongest of the three and it is deferred rather
than dismissed. It needs a contract between an experiment and the checks, saying
what an opting-in experiment promises and what happens when it stops keeping that
promise, and there is no experiment in the tree yet to write such a contract
against. Writing it now would mean guessing at the shape of the first experiment,
and the guess would be load-bearing for everything after it.

## The condition that reopens this

If experiments start carrying code that is meant to keep working, rather than
code that answered a question and stopped, the opt-in option is the one to take.
At that point this record is superseded by one that takes it, rather than argued
with.
