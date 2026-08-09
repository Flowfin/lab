# 0003. The experiment lifecycle

## What was decided

An experiment has three states and no others: `asking`, `answered`, `abandoned`.
A record carries exactly one of them. A fourth state is a change to this record
rather than a value somebody adds while writing an experiment up.

`asking` means the question is written and the work is running. A record enters
this state in the same commit that creates the experiment directory, so no
experiment exists in this tree without a question. What the state requires is a
question that a reader who was not there can tell has been answered or not. A
question written as a topic rather than as a question fails that, and it fails
it quietly, because a record whose question is "media transcoding" can never be
shown to have missed its answer.

`answered` means the work stopped and the answer is written. What the state
requires is an answer that says what was learned. No is an answer. So is the
finding that the question turned out to be the wrong question, and that one is
often the most useful thing an experiment produces. An answered experiment is
finished whichever way it went, and nothing here treats a no as a lesser
outcome, because a board that did would collect experiments that were never
allowed to fail.

`abandoned` means the work stopped and nobody wrote an answer. It is a
legitimate state and it is meant to be uncomfortable to sit in. What it requires
is a sentence saying what stopped the work, which is a lower bar than an answer
on purpose: the point is that moving here is always available and always cheaper
than the alternative, which is leaving a record in `asking` while nothing
happens. A record can be moved here honestly at any time, by the person who
started it or by anybody else reading the tree.

The transitions are `asking` to `answered` and `asking` to `abandoned`. An
abandoned experiment that somebody picks up again returns to `asking`, because
the work is running again and the record should say so. Nothing goes directly to
`answered` without having been `asking`, since the question has to exist before
the answer does, and the commit that creates the directory is where it starts.

What a machine can do here is small and it is worth writing down next to the
rule, so that nobody reads the three states as an enforcement mechanism.

A machine can refuse a record that claims `answered` and carries no answer. It
can refuse a record whose state is not one of the three. It can refuse a record
that never wrote a question. It can make a record that has been in `asking`
without movement visible in a listing, so that a stalled experiment is a line
somebody reads rather than a directory nobody opens.

What no machine here can do is make a person finish an experiment, write an
honest answer, or move a dead record to `abandoned`. There is no measurement of
whether an answer is true, whether it follows from the work, or whether the
person who wrote it believed it. A record saying "no, this does not work" and a
record saying it while the author knows otherwise are the same bytes. Review is
the only thing that catches the second, and review is a person.

The listing is therefore the load-bearing half rather than the refusals. A
refusal stops a record that contradicts itself, which is a small class of
mistake. Visibility is what stops the failure this board was opened against,
which is a directory of half-finished things nobody can read, and it works by
being read rather than by refusing anything.

## What it applies to

Every experiment under `experiments/`, from the commit that creates its
directory onwards, and the record checks that read those records.

It applies to the listing the runner prints, which reports the state of every
record it finds and is the route by which a stalled experiment becomes visible.

It does not apply to the decision records in `docs/decisions/`. Those are not
experiments, they carry no state, and they change by superseding, which record
`0000` sets out.

It does not decide the shape of the record that carries the state. Which fields
an `EXPERIMENT.md` has and how the state is written down is fixed separately, in
issue #15, and this record fixes only what the states are and what each one
requires.

## What else was considered

Two states, running and finished, with the answer optional.

Four states, adding one for an experiment that is paused rather than abandoned.

A state for an experiment whose answer was yes and which has been promoted
elsewhere.

No states at all, with a record read as finished once an answer section is
present.

## What each rejected option would have cost

Two states cost the distinction this whole record exists to make. An experiment
that stopped with an answer and one that stopped because somebody lost interest
would read identically from outside, which is precisely the state this board was
opened to make impossible. It also removes the only pressure that exists here:
if there is no name for stopping without an answer, nobody has to admit to it.

A paused state costs the discomfort. Paused is what abandoned is called by
somebody who does not want to write the sentence saying what stopped the work,
and it would absorb every record that should have moved to `abandoned` while
looking like an active tree. The honest version of paused is `abandoned` with a
sentence saying the work may resume, which costs one sentence and stays true if
it never does.

A promoted state costs the meaning of `answered`. An experiment that was
promoted is still answered, and its answer did not change when somebody else
picked the work up. Record `0005` makes promotion a section the record gains
rather than a state it moves to, exactly so that the state keeps saying what
happened to the question rather than what happened to the code afterwards. A
promoted state would also have no honest value for an experiment that answered
no, which is most of them.

No states at all costs the refusal. A record read as finished because it has an
answer section can be finished by adding an empty answer section, and there
would be nothing for a check to compare against, since the claim and the
evidence would be the same text. Naming the state separately from the answer is
what makes "claims answered, carries no answer" a thing a machine can see.
