# The promotion checklist

A product board takes work by its own rules, and it is entitled to. What this
board can do is make sure that what it hands over is takeable, so the
conversation on the other side is about the result rather than about
archaeology.

This is the list somebody works through before the hand-over. What the record
then says is fixed by
[docs/decisions/0005-how-a-result-leaves.md](docs/decisions/0005-how-a-result-leaves.md),
which this document does not restate. Work through the list, then write the
section that record describes.

The list is short on purpose. A checklist nobody finishes reading is a
checklist whose last item is never met.

## What an experiment carries into a hand-over

The question and the answer, as they are in the record. That is what the
receiving board is being asked to believe, and it is the one thing on this list
that already exists before the hand-over is considered, because a record with
no written answer is not finished.

The measurement and the command that produced it, at the commit being handed
over rather than in somebody's working tree. A number produced on a machine
nobody else has is a claim, and it is worth writing as one instead of as a
measurement. The commit is named because the tree moves afterwards and the
result did not.

What the experiment did not test. This is usually the more useful half and it
is always the half that gets left out. A reader on the other side is deciding
whether the result holds for their case, and the fastest way to answer that is
the list of cases it was never put to.

The licence the code carries out. This board declares no licence today, so
there is nothing to inherit and the terms have to be settled before anything
leaves. Entry one of issue #46 is where that is answered, and entry three asks
who may place a contributor's work under another board's terms. Until both
carry an answer, the honest state of this line is that a hand-over of somebody
else's work cannot be completed, and writing anything else into it would be
inventing permission nobody gave.

What would have to change for this to be production code, written by whoever
did the work. They know and nobody else does. An experiment is allowed to cut
every corner that is not the question, and the value of this line is that it
names which corners those were while the reason is still remembered.

## What promotion does not create

It is not a promise of support from this board. Nothing here is maintained
against anybody else's use of it, and an experiment does not gain a maintainer
by being taken somewhere.

It is not a claim that the code is finished. The line above about what would
have to change is the measure of that, and it is part of the hand-over rather
than an admission bolted onto it.

It creates no obligation on the receiving board to take anything. That board
may take the idea and none of the code, rewrite what it takes, or decline it.
An experiment that is offered and declined is still an answered experiment, and
its record says so without gaining a section, because nothing left this board.

It is not a dependency edge. Nothing on another board is allowed to depend on
anything here, and nothing in this repository ever writes to another one. A
hand-over is a copy somebody else chose to take, recorded on both sides.

## What this document cannot do

Nothing checks that this list was worked through. The runner reads a checkout
and can see whether a promotion section names the four things record `0005`
fixes, which is shape rather than substance. Whether the measurement is real,
whether the untested half is honestly listed, and whether the licence line was
agreed with anybody are all things a reader decides by following the link and
asking.

So a green run on a promoted record says the section is complete, and it says
nothing about whether the hand-over was earned. That is the boundary, and it is
written here rather than discovered by somebody who trusted the tick.
