# 0004. What happens to an experiment's code once it has an answer

## What was decided

The record is permanent and the code is not.

An `EXPERIMENT.md` record is not deleted, and it is not rewritten to hide what it
said. An experiment that answered no keeps its record exactly as it was written,
and so does an experiment that was abandoned. What a record gains over time is
lines added to it, never lines replaced.

Once an experiment has an answer, its code can be removed in a later commit if
keeping it serves nobody. That is a choice and not an obligation. Code that still
teaches somebody something stays.

When the code is removed, the record gains one line naming the commit that
removed it. The line carries the full commit hash and one clause saying what was
removed, so a reader who wants the code can check it out from history and a
reader who does not is not made to walk past it. The line is added under the
answer, and nothing already in the record is touched.

Removing code is not removing the experiment. An experiment whose code is gone
still has its question, still has its answer, and still appears in the listing
with both. Nothing about it is marked as lesser, because a prototype that was
deleted after it answered its question is a finished piece of work rather than an
incomplete one.

What this refuses is silent removal. Code that disappears with no line in the
record is indistinguishable from work that never happened, and the difference
between those two is the thing this board exists to keep.

## What it applies to

Everything under `experiments/<slug>/` other than the record itself. The record
is covered by the first sentence above and by the checks that read it.

It applies once an answer is written and not before. An experiment still asking
its question has no answer to leave behind, so removing its code leaves nothing
readable, and that is deleting the experiment rather than removing its code.

It says nothing about what happens to a result somebody wants to keep using. Code
that is meant to keep working is not an experiment any more, and moving it
somewhere it will be maintained is a separate path with its own decision.

## What else was considered

Keeping every line of every experiment forever.

Deleting the code automatically once an answer is written.

Deleting the whole experiment directory, record included, when the code goes.

Moving removed code to an archive directory inside this repository instead of
leaving it in history.

## What each rejected option would have cost

Keeping everything forever costs the reader. The tree fills with prototypes that
answered their question years ago, and a person looking for the two experiments
that matter walks past forty that do not. It also puts a quiet tax on anyone
running a check across the tree, since every prototype is more surface for a
scanner to trip over, and it makes an experiment expensive to start, which is the
one thing this board is trying to make cheap.

Automatic deletion costs the evidence. Code left behind by an experiment that
answered no is often exactly how it failed, and a rule that removes it takes away
the most useful part of a negative result. It also makes the answer dangerous to
write, because writing it would trigger the deletion, and an author who wants to
keep the code would then leave the experiment unfinished on purpose.

Deleting the whole directory costs the record, which is the one thing this
decision holds permanent. It would also make an abandoned experiment and a
completed one look the same from outside, since both would be an absence.

An archive directory costs the layout and buys nothing git does not already
provide. The root list in record `0002` does not name one, history already holds
every removed file, and a second copy inside the tree is a second thing to keep
in step with the first. The commit hash in the record is a pointer to what git
holds anyway.
